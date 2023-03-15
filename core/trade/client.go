package trade

import (
	"fmt"
	"math/big"
	"time"

	"github.com/magicpool-co/pool/core/bank"
	"github.com/magicpool-co/pool/core/trade/binance"
	"github.com/magicpool-co/pool/core/trade/bittrex"
	"github.com/magicpool-co/pool/core/trade/kucoin"
	"github.com/magicpool-co/pool/core/trade/mexc"
	"github.com/magicpool-co/pool/internal/accounting"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

var (
	inputThresholds = map[string]*big.Int{
		"CFX":  common.MustParseBigInt("2000000000000000000000000"), // 2,000,000 CFX
		"CTXC": common.MustParseBigInt("5000000000000000000000000"), // 5,000,000 CTXC
		"ERGO": new(big.Int).SetUint64(100_000_000_000),             // 100 ERGO
		"ETC":  common.MustParseBigInt("25000000000000000000"),      // 25 ETC
		"KAS":  new(big.Int).SetUint64(1_000_000_000_000),           // 10,000 KAS
		"FIRO": new(big.Int).SetUint64(10_000_000_000),              // 100 FIRO
		"FLUX": new(big.Int).SetUint64(30_000_000_000),              // 300 FLUX
		"RVN":  new(big.Int).SetUint64(500_000_000_000),             // 5,000 RVN
	}
)

/* exchange */

func NewExchange(exchangeID types.ExchangeID, apiKey, secretKey, secretPassphrase string) (types.Exchange, error) {
	switch exchangeID {
	case types.BinanceID:
		return binance.New(apiKey, secretKey), nil
	case types.KucoinID:
		return kucoin.New(apiKey, secretKey, secretPassphrase), nil
	case types.BittrexID:
		return bittrex.New(apiKey, secretKey), nil
	case types.MEXCGlobalID:
		return mexc.New(apiKey, secretKey), nil
	default:
		return nil, fmt.Errorf("unsupported exchange %d", exchangeID)
	}
}

/* client */

type Status int

const (
	BatchInactive Status = iota
	DepositsActive
	DepositsRegistered
	DepositsComplete
	TradesInactive
	TradesActiveStageOne
	TradesCompleteStageOne
	TradesActiveStageTwo
	TradesCompleteStageTwo
	TradesActiveStageThree
	TradesCompleteStageThree
	WithdrawalsActive
	WithdrawalsComplete
	BatchComplete
)

type Client struct {
	exchangeID types.ExchangeID
	exchange   types.Exchange
	pooldb     *dbcl.Client
	redis      *redis.Client
	nodes      map[string]types.PayoutNode
	telegram   *telegram.Client
	bank       *bank.Client
}

func New(pooldbClient *dbcl.Client, redisClient *redis.Client, nodes []types.PayoutNode, exchange types.Exchange, telegramClient *telegram.Client) *Client {
	nodeIdx := make(map[string]types.PayoutNode)
	for _, node := range nodes {
		nodeIdx[node.Chain()] = node
	}

	client := &Client{
		exchangeID: exchange.ID(),
		exchange:   exchange,
		pooldb:     pooldbClient,
		redis:      redisClient,
		nodes:      nodeIdx,
		telegram:   telegramClient,
		bank:       bank.New(pooldbClient, redisClient, telegramClient),
	}

	return client
}

func (c *Client) updateBatchStatus(batchID uint64, status Status) error {
	var completedAt *time.Time
	if status == BatchComplete {
		completedAt = types.TimePtr(time.Now())
	}

	batch := &pooldb.ExchangeBatch{
		ID:          batchID,
		Status:      int(status),
		CompletedAt: completedAt,
	}
	cols := []string{"status", "completed_at"}

	return pooldb.UpdateExchangeBatch(c.pooldb.Writer(), batch, cols)
}

/* core methods */

func (c *Client) CheckForNewBatch() error {
	activeBatches, err := pooldb.GetActiveExchangeBatches(c.pooldb.Reader())
	if err != nil {
		return err
	} else if len(activeBatches) > 0 {
		return nil
	}

	balanceInputs, err := pooldb.GetPendingBalanceInputsWithoutBatch(c.pooldb.Reader())
	if err != nil {
		return err
	}

	// calculate the input paths from balance inputs, fetch the
	// exchange's output thresholds and current prices. though input
	// thresholds are global (since exchanges don't charge any deposit
	// fees, the only deposit fee is the network tx fee), output thresholds
	// are exchange specific since withdrawal fees can vary across exchanges.
	inputPaths, err := balanceInputsToInputPaths(balanceInputs)
	if err != nil {
		return err
	}

	// check all input and output paths, remove the inputs/outputs that
	// do not have deposits or withdrawals currently enabled
	for inChainID, outputIdx := range inputPaths {
		if inChainID == "KAS" {
			delete(inputPaths, inChainID)
			continue
		}

		// check for deposits being enabled on the input chains
		depositsEnabled, _, err := c.exchange.GetWalletStatus(inChainID)
		if err != nil {
			return err
		} else if !depositsEnabled {
			delete(inputPaths, inChainID)
			continue
		}

		for outChainID := range outputIdx {
			// check for withdrawals being enabled on the output chains
			_, withdrawalsEnabled, err := c.exchange.GetWalletStatus(outChainID)
			if err != nil {
				return err
			} else if !withdrawalsEnabled {
				delete(inputPaths[inChainID], outChainID)
			}
		}
	}

	outputThresholds := c.exchange.GetOutputThresholds()
	prices, err := c.exchange.GetPrices(inputPaths)
	if err != nil {
		return err
	}

	// calculate the paths that meet the given thresholds, based off
	// of the input paths, output thresholds, and current prices (since
	// the final price of each trade is only know at runtime, cumulative output
	// values are estimated through the current prices - see the exchange accountant
	// for more details on this process).
	outputPaths, err := accounting.CalculateExchangePaths(inputPaths, inputThresholds,
		outputThresholds, prices)
	if err != nil {
		return err
	} else if len(outputPaths) == 0 {
		return nil
	}

	// create a db tx to make sure the batch, all of the trade paths,
	// and all of the balance inputs are inserted (or updated). if
	// anything fails, rollback the tx
	tx, err := c.pooldb.Begin()
	if err != nil {
		return err
	}
	defer tx.SafeRollback()

	batch := &pooldb.ExchangeBatch{
		ExchangeID: int(c.exchangeID),
		Status:     int(BatchInactive),
	}

	batchID, err := pooldb.InsertExchangeBatch(tx, batch)
	if err != nil {
		return err
	}

	for _, balanceInput := range balanceInputs {
		// verify that the path is actually in the batch, to avoid
		// balance inputs that are not included as inputs having
		// a batch ID set
		if _, ok := outputPaths[balanceInput.ChainID]; !ok {
			continue
		} else if _, ok := outputPaths[balanceInput.ChainID][balanceInput.OutChainID]; !ok {
			continue
		}

		balanceInput.BatchID = types.Uint64Ptr(batchID)
		err = pooldb.UpdateBalanceInput(tx, balanceInput, []string{"batch_id"})
		if err != nil {
			return err
		}
	}

	// calculate the exchange inputs based off of the output paths
	// that meet both the input thresholds and output thresholds (these
	// are held in the db to avoid having to recalculate the output paths,
	// since it is an expensive calculation and prices could change).
	exchangeInputs := make([]*pooldb.ExchangeInput, 0)
	for inChainID, outputIdx := range outputPaths {
		for outChainID, value := range outputIdx {
			exchangeInput := &pooldb.ExchangeInput{
				BatchID:    batchID,
				InChainID:  inChainID,
				OutChainID: outChainID,

				Value: dbcl.NullBigInt{Valid: true, BigInt: value},
			}
			exchangeInputs = append(exchangeInputs, exchangeInput)
		}
	}

	err = pooldb.InsertExchangeInputs(tx, exchangeInputs...)
	if err != nil {
		return err
	}

	c.telegram.NotifyInitiateExchangeBatch(batchID)

	return tx.SafeCommit()
}

func (c *Client) ProcessBatch(batchID uint64) error {
	batch, err := pooldb.GetExchangeBatch(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	} else if batch.ID != batchID {
		return fmt.Errorf("batch not found")
	}

	// switch on the batch status and execute the proper action
	switch Status(batch.Status) {
	case BatchInactive:
		return c.InitiateDeposits(batchID)
	case DepositsActive:
		return c.RegisterDeposits(batchID)
	case DepositsRegistered:
		return c.ConfirmDeposits(batchID)
	case DepositsComplete:
		return c.InitiateTrades(batchID)
	case TradesInactive:
		return c.InitiateTradeStage(batchID, 1)
	case TradesActiveStageOne:
		return c.ConfirmTradeStage(batchID, 1)
	case TradesCompleteStageOne:
		return c.InitiateTradeStage(batchID, 2)
	case TradesActiveStageTwo:
		return c.ConfirmTradeStage(batchID, 2)
	case TradesCompleteStageTwo:
		return c.InitiateTradeStage(batchID, 3)
	case TradesActiveStageThree:
		return c.ConfirmTradeStage(batchID, 3)
	case TradesCompleteStageThree:
		return c.InitiateWithdrawals(batchID)
	case WithdrawalsActive:
		return c.ConfirmWithdrawals(batchID)
	case WithdrawalsComplete:
		return c.CreditWithdrawals(batchID)
	case BatchComplete:
		return nil
	default:
		return fmt.Errorf("unknown batch status %d", batch.Status)
	}
}

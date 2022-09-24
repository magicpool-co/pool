package trade

import (
	"fmt"

	"github.com/magicpool-co/pool/core/trade/binance"
	"github.com/magicpool-co/pool/core/trade/bittrex"
	"github.com/magicpool-co/pool/core/trade/kucoin"
	"github.com/magicpool-co/pool/internal/accounting"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
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
	WithdrawalsActive
	WithdrawalsComplete
	BatchComplete
)

type Client struct {
	exchangeID types.ExchangeID
	exchange   types.Exchange
	pooldb     *dbcl.Client
	nodes      map[string]types.PayoutNode
}

func New(pooldbClient *dbcl.Client, nodes map[string]types.PayoutNode, exchangeID types.ExchangeID, apiKey, secretKey, secretPassphrase string) (*Client, error) {
	exchange, err := NewExchange(exchangeID, apiKey, secretKey, secretPassphrase)
	if err != nil {
		return nil, err
	}

	client := &Client{
		exchangeID: exchangeID,
		exchange:   exchange,
		pooldb:     pooldbClient,
		nodes:      nodes,
	}

	return client, nil
}

func (c *Client) updateBatchStatus(batchID uint64, status Status) error {
	batch := &pooldb.ExchangeBatch{
		ID:     batchID,
		Status: int(status),
	}

	return pooldb.UpdateExchangeBatch(c.pooldb.Writer(), batch, []string{"status"})
}

/* core methods */

func (c *Client) CheckForNewBatch() error {
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
	outputPaths, err := accounting.CalculateExchangePaths(inputPaths, outputThresholds, prices)
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

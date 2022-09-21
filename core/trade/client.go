package trade

import (
	"fmt"
	"math/big"
	"time"

	"github.com/magicpool-co/pool/core/trade/binance"
	"github.com/magicpool-co/pool/core/trade/bittrex"
	"github.com/magicpool-co/pool/core/trade/kucoin"
	"github.com/magicpool-co/pool/internal/accounting"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

/* exchange */

type ExchangeID int

const (
	BinanceID ExchangeID = iota
	KucoinID
	BittrexID
)

type Exchange interface {
	GetAccountStatus() error
	GetRate(base, quote string) (float64, error)
	GetHistoricalRate(base, quote string, timestamp time.Time) (float64, error)
	GetWalletStatus(chain string) (bool, error)
	GetWalletAddress(chain string) (string, error)
	GetWalletBalance(chain string) (float64, error)
	GetDepositStatus(chain, txid string) (bool, error)
	TransferToTradeAccount(chain string, quantity float64) error
	TransferToMainAccount(chain string, quantity float64) error
	GenerateTradePath(fromChain, toChain string, quantity *big.Int) ([]*pooldb.ExchangeTrade, error)
	CreateOrder(market, direction string, quantity float64) (string, error)
	GetOrderStatus(market, orderID string) (bool, error)
	CreateWithdrawal(chain, address string, quantity float64) (string, error)
	GetWithdrawalStatus(chain, withdrawalID string) (bool, error)
}

func NewExchange(exchangeID ExchangeID, apiKey, secretKey, secretPassphrase string) (Exchange, error) {
	switch exchangeID {
	case BinanceID:
		return binance.New(apiKey, secretKey), nil
	case KucoinID:
		return kucoin.New(apiKey, secretKey, secretPassphrase), nil
	case BittrexID:
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
	exchangeID ExchangeID
	exchange   Exchange
	pooldb     *dbcl.Client
	nodes      map[string]types.PayoutNode
}

func New(nodes map[string]types.PayoutNode, exchangeID ExchangeID, apiKey, secretKey, secretPassphrase string) (*Client, error) {
	exchange, err := NewExchange(exchangeID, apiKey, secretKey, secretPassphrase)
	if err != nil {
		return nil, err
	}

	client := &Client{
		exchangeID: exchangeID,
		exchange:   exchange,
		nodes:      nodes,
	}

	return client, nil
}

/* helpers */

func (c *Client) balanceInputsToInputPaths(balanceInputs []*pooldb.BalanceInput) (map[string]map[string]*big.Int, error) {
	inputPaths := make(map[string]map[string]*big.Int)
	for _, balanceInput := range balanceInputs {
		if !balanceInput.Value.Valid {
			return nil, fmt.Errorf("no value for balance input %d", balanceInput.ID)
		}
		inChainID := balanceInput.ChainID
		outChainID := balanceInput.OutputChainID
		value := balanceInput.Value.BigInt

		if _, ok := inputPaths[inChainID]; !ok {
			inputPaths[inChainID] = make(map[string]*big.Int)
		}
		if _, ok := inputPaths[inChainID][outChainID]; !ok {
			inputPaths[inChainID][outChainID] = new(big.Int)
		}

		inputPaths[inChainID][outChainID].Add(inputPaths[inChainID][outChainID], value)
	}

	return inputPaths, nil
}

func (c *Client) exchangeInputsToOutputPaths(exchangeInputs []*pooldb.ExchangeInput) (map[string]map[string]*big.Int, error) {
	outputPaths := make(map[string]map[string]*big.Int)
	for _, exchangeInput := range exchangeInputs {
		if !exchangeInput.Value.Valid {
			return nil, fmt.Errorf("no value for exchange input %d", exchangeInput.ID)
		}
		inChainID := exchangeInput.InputChainID
		outChainID := exchangeInput.OutputChainID
		value := exchangeInput.Value.BigInt

		if _, ok := outputPaths[inChainID]; !ok {
			outputPaths[inChainID] = make(map[string]*big.Int)
		}
		if _, ok := outputPaths[inChainID][outChainID]; !ok {
			outputPaths[inChainID][outChainID] = new(big.Int)
		}

		outputPaths[inChainID][outChainID].Add(outputPaths[inChainID][outChainID], value)
	}

	return outputPaths, nil
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

	inputPaths, err := c.balanceInputsToInputPaths(balanceInputs)
	if err != nil {
		return err
	}

	outputThresholds := c.exchange.GetOutputThresholds()
	prices, err := c.exchange.GetPrices(inputPaths)
	if err != nil {
		return err
	}

	outputPaths, err := accounting.CalculateExchangePaths(inputPaths, outputThresholds, prices)
	if err != nil {
		return err
	} else if len(outputPaths) == 0 {
		return nil
	}

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

	exchangeInputs := make([]*pooldb.ExchangeInput, 0)
	for inChainID, outputIdx := range outputPaths {
		for outChainID, value := range outputIdx {
			exchangeInput := &pooldb.ExchangeInput{
				BatchID:       batchID,
				InputChainID:  inChainID,
				OutputChainID: outChainID,

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

package trade

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/core/trade/binance"
	"github.com/magicpool-co/pool/core/trade/bittrex"
	"github.com/magicpool-co/pool/core/trade/kucoin"
	"github.com/magicpool-co/pool/internal/accounting"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
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

func New(nodes map[string]types.PayoutNode, exchangeID types.ExchangeID, apiKey, secretKey, secretPassphrase string) (*Client, error) {
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
	// iterate through each balance input, creating a trade path from the
	// given input chain to the desired, summing any overlapping paths (since
	// paths are batched and redistributed during withdrawal crediting)

	// note that "output path" just refers to the path being an "unconfirmed"
	// path, meaning it is still unknown whether or not it will actually be executed
	inputPaths := make(map[string]map[string]*big.Int)
	for _, balanceInput := range balanceInputs {
		if !balanceInput.Value.Valid {
			return nil, fmt.Errorf("no value for balance input %d", balanceInput.ID)
		}
		inChainID := balanceInput.ChainID
		outChainID := balanceInput.OutChainID
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

func (c *Client) balanceInputsToInitialProportions(balanceInputs []*pooldb.BalanceInput) (map[string]map[string]map[uint64]*big.Int, error) {
	// iterate through each balance input, summing the owed amount for each
	// miner participating in the exchange batch

	// note that the in chain is the primary key and the
	// out chain is the secondary key
	proportions := make(map[string]map[string]map[uint64]*big.Int)
	for _, balanceInput := range balanceInputs {
		if !balanceInput.Value.Valid {
			return nil, fmt.Errorf("no value for balance input %d", balanceInput.ID)
		}
		inChainID := balanceInput.ChainID
		outChainID := balanceInput.OutChainID
		minerID := balanceInput.MinerID
		value := balanceInput.Value.BigInt

		if _, ok := proportions[outChainID]; !ok {
			proportions[outChainID] = make(map[string]map[uint64]*big.Int)
		}
		if _, ok := proportions[outChainID][inChainID]; !ok {
			proportions[outChainID][inChainID] = make(map[uint64]*big.Int)
		}
		if _, ok := proportions[outChainID][inChainID][minerID]; !ok {
			proportions[outChainID][inChainID][minerID] = new(big.Int)
		}

		proportions[outChainID][inChainID][minerID].Add(proportions[outChainID][inChainID][minerID], value)
	}

	return proportions, nil
}

func (c *Client) exchangeInputsToOutputPaths(exchangeInputs []*pooldb.ExchangeInput) (map[string]map[string]*big.Int, error) {
	// iterate through each exchange input, creating a trade path from the
	// given input chain to the desired (exchange inputs should already be summed)

	// note that "output path" just refers to the path being
	// a "confirmed" path that will be executed
	outputPaths := make(map[string]map[string]*big.Int)
	for _, exchangeInput := range exchangeInputs {
		if !exchangeInput.Value.Valid {
			return nil, fmt.Errorf("no value for exchange input %d", exchangeInput.ID)
		}
		inChainID := exchangeInput.InChainID
		outChainID := exchangeInput.OutChainID
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

func (c *Client) finalTradesToFinalProportions(finalTrades []*pooldb.ExchangeTrade) (map[string]map[string]*big.Int, error) {
	// iterate through each final trade, creating an index of the value each
	// trade path ended up receiving after all trades were executed

	// note that the final chain is the primary key and the
	// initial chain is the secondary key
	proportions := make(map[string]map[string]*big.Int)
	for _, finalTrade := range finalTrades {
		if !finalTrade.Value.Valid {
			return nil, fmt.Errorf("no value for trade %d", finalTrade.ID)
		}
		initialChainID := finalTrade.InitialChainID
		finalChainID := finalTrade.ToChainID
		value := finalTrade.Value.BigInt

		if _, ok := proportions[finalChainID]; !ok {
			proportions[finalChainID] = make(map[string]*big.Int)
		}
		if _, ok := proportions[finalChainID][initialChainID]; !ok {
			proportions[finalChainID][initialChainID] = new(big.Int)
		}

		proportions[finalChainID][initialChainID].Add(proportions[finalChainID][initialChainID], value)
	}

	return proportions, nil
}

func (c *Client) finalTradesToAvgWeightedPrice(finalTrades []*pooldb.ExchangeTrade) (map[string]map[string]float64, error) {
	// iterate through each final trade, calculating the averaged weighted
	// cumulative fill price for each path

	// calculate the sum prices and weights for each path
	prices := make(map[string]map[string]float64)
	weights := make(map[string]map[string]float64)
	for _, finalTrade := range finalTrades {
		if !finalTrade.Value.Valid {
			return nil, fmt.Errorf("no value for trade %d", finalTrade.ID)
		} else if finalTrade.CumulativeFillPrice == nil {
			return nil, fmt.Errorf("no cumulative fill price for trade %d", finalTrade.ID)
		}
		initialChainID := finalTrade.InitialChainID
		finalChainID := finalTrade.ToChainID

		if _, ok := prices[initialChainID]; !ok {
			prices[initialChainID] = make(map[string]float64)
			weights[initialChainID] = make(map[string]float64)
		}

		// grab the units to convert the weight into a more reasonable value
		// (instead of massive numbers that could easily be >1e20)
		units, err := common.GetDefaultUnits(finalChainID)
		if err != nil {
			return nil, err
		}

		prices[initialChainID][finalChainID] += types.Float64Value(finalTrade.CumulativeFillPrice)
		weights[initialChainID][finalChainID] += common.BigIntToFloat64(finalTrade.Value.BigInt, units)
	}

	// divide the sum prices by the sum weights to calculate the average weighted price
	for initialChainID, weightIdx := range prices {
		for finalChainID, weight := range weightIdx {
			if weight > 0 {
				prices[initialChainID][finalChainID] /= weight
			} else {
				prices[initialChainID][finalChainID] = 0
			}
		}
	}

	return prices, nil
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
	inputPaths, err := c.balanceInputsToInputPaths(balanceInputs)
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

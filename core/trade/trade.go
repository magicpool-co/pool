package trade

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) InitiateTrades(batchID uint64) error {
	deposits, err := pooldb.GetExchangeDeposits(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	}

	// create a map of deposit fees and values by chain
	depositValues := make(map[string]*big.Int)
	depositFees := make(map[string]*big.Int)
	for _, deposit := range deposits {
		if !deposit.Value.Valid {
			return fmt.Errorf("no value for deposit %d", deposit.ID)
		} else if !deposit.Fees.Valid {
			return fmt.Errorf("no fees for deposit %d", deposit.ID)
		} else if _, ok := depositValues[deposit.ChainID]; !ok {
			depositValues[deposit.ChainID] = new(big.Int)
			depositFees[deposit.ChainID] = new(big.Int)
		}

		depositValues[deposit.ChainID].Add(depositValues[deposit.ChainID], deposit.Value.BigInt)
		depositFees[deposit.ChainID].Add(depositFees[deposit.ChainID], deposit.Fees.BigInt)
	}

	// fetch the balance inputs and convert them into the output paths
	balanceInputs, err := pooldb.GetExchangeInputs(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	}

	outputPaths, err := exchangeInputsToOutputPaths(balanceInputs)
	if err != nil {
		return err
	}

	// estimate (ignore rounding errors) the proportional
	// deposit fees for each trade path
	depositFeeIdx := make(map[string]map[string]*big.Int)
	for fromChainID, outputIdx := range outputPaths {
		depositFeeIdx[fromChainID] = make(map[string]*big.Int)
		for toChainID, value := range outputIdx {
			if _, ok := depositValues[fromChainID]; !ok {
				return fmt.Errorf("no deposit value for %s", fromChainID)
			} else if _, ok := depositFees[fromChainID]; !ok {
				return fmt.Errorf("no deposit fee for %s", fromChainID)
			}

			depositFee := new(big.Int).Mul(value, depositFees[fromChainID])
			depositFee.Div(depositFee, depositFees[fromChainID])
			depositFeeIdx[fromChainID][toChainID] = depositFee
		}
	}

	// iterate through the trade paths and create the initial trades
	var pathID int
	trades := make([]*pooldb.ExchangeTrade, 0)
	for fromChainID, outputIdx := range outputPaths {
		for toChainID, value := range outputIdx {
			// generate the respective exchange's trades for a given trade path
			parsedTrades, err := c.exchange.GenerateTradePath(fromChainID, toChainID)
			if err != nil {
				return err
			}

			pathID++
			for i, parsedTrade := range parsedTrades {
				// for the first trade, set the initial trade value and deposit fees
				var initialTradeValue, initialDepositFees dbcl.NullBigInt
				if i == 0 {
					initialTradeValue = dbcl.NullBigInt{Valid: true, BigInt: value}
					initialDepositFees = dbcl.NullBigInt{Valid: true, BigInt: depositFeeIdx[fromChainID][toChainID]}
				}

				trade := &pooldb.ExchangeTrade{
					BatchID: batchID,
					PathID:  pathID,
					StageID: i + 1,

					InitialChainID: fromChainID,
					FromChainID:    parsedTrade.FromChain,
					ToChainID:      parsedTrade.ToChain,
					Market:         parsedTrade.Market,
					Direction:      int(parsedTrade.Direction),

					Value:                 initialTradeValue,
					CumulativeDepositFees: initialDepositFees,
				}
				trades = append(trades, trade)
			}
		}
	}

	err = pooldb.InsertExchangeTrades(c.pooldb.Writer(), trades...)
	if err != nil {
		return err
	}

	return c.updateBatchStatus(batchID, TradesInactive)
}

func (c *Client) InitiateTradeStage(batchID uint64, stage int) error {
	var tradeStageStatus Status
	switch stage {
	case 1:
		tradeStageStatus = TradesActiveStageOne
	case 2:
		tradeStageStatus = TradesActiveStageTwo
	default:
		return fmt.Errorf("unsupported trade stage %d", stage)
	}

	trades, err := pooldb.GetExchangeTradesByStage(c.pooldb.Reader(), batchID, stage)
	if err != nil {
		return err
	}

	for _, trade := range trades {
		if trade.Initiated {
			continue
		} else if !trade.Value.Valid {
			return fmt.Errorf("no value for trade %d", trade.ID)
		}

		// fetch the from chain's units
		units, err := common.GetDefaultUnits(trade.FromChainID)
		if err != nil {
			return err
		}

		orderPrice, err := c.exchange.GetRate(trade.Market)
		if err != nil {
			return err
		}

		// process the trade value as a float and execute the trade
		floatValue := common.BigIntToFloat64(trade.Value.BigInt, units)
		direction := types.TradeDirection(trade.Direction)
		tradeID, err := c.exchange.CreateTrade(trade.Market, direction, floatValue)
		if err != nil {
			return err
		}

		trade.ExchangeTradeID = types.StringPtr(tradeID)
		trade.OrderPrice = types.Float64Ptr(orderPrice)
		trade.Initiated = true
		trade.Confirmed = false

		cols := []string{"exchange_trade_id", "initiated", "confirmed"}
		err = pooldb.UpdateExchangeTrade(c.pooldb.Writer(), trade, cols)
		if err != nil {
			return err
		}

		c.telegram.NotifyInitiateTrade(trade.ID, trade.PathID, trade.StageID, trade.Market, direction.String(), floatValue)
	}

	return c.updateBatchStatus(batchID, tradeStageStatus)
}

func (c *Client) ConfirmTradeStage(batchID uint64, stage int) error {
	var tradeStageStatus Status
	switch stage {
	case 1:
		tradeStageStatus = TradesCompleteStageOne
	case 2:
		tradeStageStatus = TradesCompleteStageTwo
	default:
		return fmt.Errorf("unsupported trade stage %d", stage)
	}

	trades, err := pooldb.GetExchangeTradesByStage(c.pooldb.Reader(), batchID, stage)
	if err != nil {
		return err
	}

	completedAll := true
	for _, trade := range trades {
		if trade.Initiated && trade.Confirmed {
			continue
		} else if trade.ExchangeTradeID == nil {
			return fmt.Errorf("trade %d has no exchange trade id", trade.ID)
		} else if !trade.Value.Valid {
			return fmt.Errorf("no value for trade %d", trade.ID)
		} else if trade.OrderPrice == nil {
			return fmt.Errorf("no order price for trade %d", trade.ID)
		}

		// fetch the units for the from and to chains
		fromUnits, err := common.GetDefaultUnits(trade.FromChainID)
		if err != nil {
			return err
		}

		toUnits, err := common.GetDefaultUnits(trade.ToChainID)
		if err != nil {
			return err
		}

		// process the trade value as a float and fetch the trade from the exchange
		tradeID := types.StringValue(trade.ExchangeTradeID)
		value := common.BigIntToFloat64(trade.Value.BigInt, fromUnits)
		parsedTrade, err := c.exchange.GetTradeByID(trade.Market, tradeID, value)
		if err != nil {
			return err
		} else if !parsedTrade.Completed {
			completedAll = false
			continue
		}

		// process the proceeds and fees as big ints in the current chain's units
		proceeds, err := common.StringDecimalToBigint(parsedTrade.Proceeds, toUnits)
		if err != nil {
			return err
		}

		fees, err := common.StringDecimalToBigint(parsedTrade.Fees, toUnits)
		if err != nil {
			return err
		}

		fillPrice, err := strconv.ParseFloat(parsedTrade.Price, 64)
		if err != nil {
			return err
		}

		// check for the previous trade to collect cumulative deposit and trade fees. if no
		// previous trade exists, collect the initial deposit fees from the current trade
		var cumulativeFillPrice, cumulativeDepositFees, cumulativeTradeFees float64
		prevTrade, err := pooldb.GetExchangeTradeByPathAndStage(c.pooldb.Reader(), batchID, trade.PathID, trade.StageID-1)
		if err != nil {
			return err
		} else if prevTrade != nil {
			if prevTrade.CumulativeFillPrice == nil {
				return fmt.Errorf("no cumulative fill price for trade %d", prevTrade.ID)
			} else if !prevTrade.CumulativeDepositFees.Valid {
				return fmt.Errorf("no cumulative deposit fees for trade %d", prevTrade.ID)
			} else if !prevTrade.CumulativeTradeFees.Valid {
				return fmt.Errorf("no cumulative trade fees for trade %d", prevTrade.ID)
			}

			// set cumulative fill price to previous cumulative fill price
			cumulativeFillPrice = types.Float64Value(prevTrade.CumulativeFillPrice)

			// adjust cumulative deposit and trade fees to the current chain's price
			cumulativeDepositFees = fillPrice * common.BigIntToFloat64(prevTrade.CumulativeDepositFees.BigInt, fromUnits)
			cumulativeTradeFees = fillPrice * common.BigIntToFloat64(prevTrade.CumulativeTradeFees.BigInt, fromUnits)
		} else {
			if !trade.CumulativeDepositFees.Valid {
				return fmt.Errorf("no cumulative deposit fees for trade %d", trade.ID)
			}

			// set cumulative fill price to 1
			cumulativeFillPrice = 1

			// adjust initial cumulative deposit fees to the current chain's price
			cumulativeDepositFees = fillPrice * common.BigIntToFloat64(trade.CumulativeDepositFees.BigInt, fromUnits)
		}

		// adjust cumulative fill price to account for previous
		if fillPrice > 0 {
			switch types.TradeDirection(trade.Direction) {
			case types.TradeBuy:
				cumulativeFillPrice /= fillPrice
			case types.TradeSell:
				cumulativeFillPrice *= fillPrice
			}
		}

		// sum cumulative deposit and trade fees
		cumulativeTradeFees += common.BigIntToFloat64(fees, toUnits)
		cumulativeDepositFeesBig := common.Float64ToBigInt(cumulativeDepositFees, toUnits)
		cumulativeTradeFeesBig := common.Float64ToBigInt(cumulativeTradeFees, toUnits)

		// check for the next trade to update the trade value, since it is only known
		// after the previous (current) trade is filled. if no next trade exists, transfer
		// the balance from the trade account to the main account (kucoin only, empty method otherwise)
		nextTrade, err := pooldb.GetExchangeTradeByPathAndStage(c.pooldb.Reader(), batchID, trade.PathID, trade.StageID+1)
		if err != nil {
			return err
		} else if nextTrade == nil {
			proceedsFloat := common.BigIntToFloat64(proceeds, toUnits)
			err = c.exchange.TransferToMainAccount(trade.ToChainID, proceedsFloat)
			if err != nil {
				return err
			}
		} else {
			nextTrade.Value = dbcl.NullBigInt{Valid: true, BigInt: proceeds}
			cols := []string{"value"}
			err = pooldb.UpdateExchangeTrade(c.pooldb.Writer(), trade, cols)
			if err != nil {
				return err
			}
		}

		// calculate the slippage from the initial order price and the fill price
		var slippage float64
		orderPrice := types.Float64Value(trade.OrderPrice)
		if types.TradeDirection(trade.Direction) == types.TradeSell {
			slippage = (fillPrice - orderPrice) / orderPrice
		} else {
			slippage = (orderPrice - fillPrice) / fillPrice
		}

		// finalize the trade in the database
		trade.Proceeds = dbcl.NullBigInt{Valid: true, BigInt: proceeds}
		trade.TradeFees = dbcl.NullBigInt{Valid: true, BigInt: fees}
		trade.CumulativeDepositFees = dbcl.NullBigInt{Valid: true, BigInt: cumulativeDepositFeesBig}
		trade.CumulativeTradeFees = dbcl.NullBigInt{Valid: true, BigInt: cumulativeTradeFeesBig}
		trade.FillPrice = types.Float64Ptr(fillPrice)
		trade.CumulativeFillPrice = types.Float64Ptr(cumulativeFillPrice)
		trade.Slippage = types.Float64Ptr(slippage)
		trade.Confirmed = true

		cols := []string{"proceeds", "trade_fees", "cumulative_deposit_fees",
			"cumulative_trade_fees", "fill_price", "cumulative_fill_price",
			"slippage", "confirmed"}
		err = pooldb.UpdateExchangeTrade(c.pooldb.Writer(), trade, cols)
		if err != nil {
			return err
		}

		c.telegram.NotifyFinalizeTrade(trade.ID)
	}

	if completedAll {
		return c.updateBatchStatus(batchID, tradeStageStatus)
	}

	return nil
}

package trade

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/magicpool-co/pool/internal/accounting"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) InitiateTrades(batchID uint64, exchange types.Exchange) error {
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
	depositValueIdx := make(map[string]map[string]*big.Int)
	depositFeeIdx := make(map[string]map[string]*big.Int)
	for fromChainID, outputIdx := range outputPaths {
		if _, ok := depositValues[fromChainID]; !ok {
			return fmt.Errorf("no deposit value for %s", fromChainID)
		} else if _, ok := depositFees[fromChainID]; !ok {
			return fmt.Errorf("no deposit fee for %s", fromChainID)
		}

		depositValueIdx[fromChainID], depositFeeIdx[fromChainID], err = accounting.CalculateProportionalValues(
			depositValues[fromChainID], depositFees[fromChainID], outputIdx)
		if err != nil {
			return err
		}
	}

	// iterate through the trade paths and create the initial trades
	var pathID int
	trades := make([]*pooldb.ExchangeTrade, 0)
	for fromChainID, outputIdx := range depositValueIdx {
		for toChainID, depositValue := range outputIdx {
			// generate the respective exchange's trades for a given trade path
			parsedTrades, err := exchange.GenerateTradePath(fromChainID, toChainID)
			if err != nil {
				return err
			}

			pathID++
			for i, parsedTrade := range parsedTrades {
				// for the first trade, set the initial trade value and deposit fees
				var initialTradeValue, initialDepositFees dbcl.NullBigInt
				if i == 0 {
					initialTradeValue = dbcl.NullBigInt{Valid: true, BigInt: depositValue}
					initialDepositFees = dbcl.NullBigInt{Valid: true, BigInt: depositFeeIdx[fromChainID][toChainID]}
				}

				trade := &pooldb.ExchangeTrade{
					BatchID: batchID,
					PathID:  pathID,
					StageID: i + 1,
					StepID:  0,

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

	return c.updateBatchStatus(c.pooldb.Writer(), batchID, TradesInactive)
}

func (c *Client) InitiateTradeStage(batchID uint64, exchange types.Exchange, stage int) error {
	var tradeStageStatus Status
	switch stage {
	case 1:
		tradeStageStatus = TradesActiveStageOne
	case 2:
		tradeStageStatus = TradesActiveStageTwo
	case 3:
		tradeStageStatus = TradesActiveStageThree
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

		orderPrice, err := exchange.GetRate(trade.Market)
		if err != nil {
			return err
		}

		// process the trade value as a float and execute the trade
		floatValue := common.BigIntToFloat64(trade.Value.BigInt, units)
		direction := types.TradeDirection(trade.Direction)
		tradeID, err := exchange.CreateTrade(trade.Market, direction, floatValue)
		if err != nil {
			return err
		}

		trade.ExchangeTradeID = types.StringPtr(tradeID)
		trade.OrderPrice = types.Float64Ptr(orderPrice)
		trade.Initiated = true
		trade.Confirmed = false

		cols := []string{"exchange_trade_id", "order_price", "initiated", "confirmed"}
		err = pooldb.UpdateExchangeTrade(c.pooldb.Writer(), trade, cols)
		if err != nil {
			return err
		}

		c.telegram.NotifyInitiateTrade(trade.ID, trade.PathID, trade.StageID, trade.Market, direction.String(), floatValue)
	}

	return c.updateBatchStatus(c.pooldb.Writer(), batchID, tradeStageStatus)
}

func (c *Client) confirmTrade(batchID uint64, exchange types.Exchange, stage int, trade *pooldb.ExchangeTrade) (bool, error) {
	var tradeStageIncompleteStatus Status
	switch stage {
	case 1:
		tradeStageIncompleteStatus = TradesInactive
	case 2:
		tradeStageIncompleteStatus = TradesCompleteStageOne
	case 3:
		tradeStageIncompleteStatus = TradesCompleteStageTwo
	default:
		return false, fmt.Errorf("unsupported trade stage %d", stage)
	}

	completedTrade := true
	if trade.Initiated && trade.Confirmed {
		return completedTrade, nil
	} else if trade.ExchangeTradeID == nil {
		return completedTrade, fmt.Errorf("trade %d has no exchange trade id", trade.ID)
	} else if !trade.Value.Valid {
		return completedTrade, fmt.Errorf("no value for trade %d", trade.ID)
	} else if trade.OrderPrice == nil {
		return completedTrade, fmt.Errorf("no order price for trade %d", trade.ID)
	}

	// fetch the units for the from and to chains
	fromUnits, err := common.GetDefaultUnits(trade.FromChainID)
	if err != nil {
		return completedTrade, err
	}

	toUnits, err := common.GetDefaultUnits(trade.ToChainID)
	if err != nil {
		return completedTrade, err
	}

	tx, err := c.pooldb.Begin()
	if err != nil {
		return false, err
	}
	defer tx.SafeRollback()

	// process the trade value as a float and fetch the trade from the exchange
	tradeID := types.StringValue(trade.ExchangeTradeID)
	value := common.BigIntToFloat64(trade.Value.BigInt, fromUnits)
	parsedTrade, err := exchange.GetTradeByID(trade.Market, tradeID, value)
	if err != nil {
		return completedTrade, err
	} else if parsedTrade == nil || !parsedTrade.Completed {
		completedTrade = false
		timeout := exchange.GetTradeTimeout()
		// if the exchange doesn't support trade timeouts, or the timeout
		// hasn't yet been exceeded, continue to wait.
		if timeout == 0 || time.Since(trade.UpdatedAt) < timeout {
			return completedTrade, nil
		}

		// otherwise cancel it, refetch the trade, then make a new trade
		// to finish out the rest of the order
		exchange.CancelTradeByID(trade.Market, tradeID)
		// if err != nil {
		// 	return completedTrade, err
		// }

		time.Sleep(time.Second)

		// refetch the trade in case more was filled
		parsedTrade, err = exchange.GetTradeByID(trade.Market, tradeID, value)
		if err != nil {
			return completedTrade, err
		}

		// set the old trade value to zero and the new trade value to
		// the full desired trade value. check if any of the old trade was filled
		// alter the old trade and new trade values accordingly.
		tradeValue := dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)}
		newTradeValue := trade.Value
		if parsedTrade != nil {
			filledValue, err := common.StringDecimalToBigint(parsedTrade.Value, fromUnits)
			if err != nil {
				return completedTrade, err
			}

			remainingValueBig := new(big.Int).Sub(trade.Value.BigInt, filledValue)
			tradeValue = dbcl.NullBigInt{Valid: true, BigInt: filledValue}
			newTradeValue = dbcl.NullBigInt{Valid: true, BigInt: remainingValueBig}
		}

		// modify the old trade value, make the new trade, and update/insert them.
		trade.Value = tradeValue
		nextTrade := &pooldb.ExchangeTrade{
			BatchID: trade.BatchID,
			PathID:  trade.PathID,
			StageID: trade.StageID,
			StepID:  trade.StepID + 1,

			InitialChainID: trade.InitialChainID,
			FromChainID:    trade.FromChainID,
			ToChainID:      trade.ToChainID,
			Market:         trade.Market,
			Direction:      trade.Direction,

			Value:                 newTradeValue,
			CumulativeDepositFees: dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
		}

		err = pooldb.InsertExchangeTrades(tx, nextTrade)
		if err != nil {
			return completedTrade, err
		}

		// flip back to the previous stage to kick off the newly created trade.
		err = c.updateBatchStatus(tx, batchID, tradeStageIncompleteStatus)
		if err != nil {
			return completedTrade, err
		}
	}

	// process the proceeds and fees as big ints in the current chain's units
	proceeds, err := common.StringDecimalToBigint(parsedTrade.Proceeds, toUnits)
	if err != nil {
		return completedTrade, err
	}

	fees, err := common.StringDecimalToBigint(parsedTrade.Fees, toUnits)
	if err != nil {
		return completedTrade, err
	}

	fillPrice, err := strconv.ParseFloat(parsedTrade.Price, 64)
	if err != nil {
		return completedTrade, err
	}

	// check for the previous trade to collect cumulative deposit and trade fees. if no
	// previous trade exists, collect the initial deposit fees from the current trade
	var cumulativeFillPrice, cumulativeDepositFees, cumulativeTradeFees float64
	prevTrade, err := pooldb.GetExchangeTradeByPathAndStage(tx, batchID, trade.PathID, trade.StageID-1)
	if err != nil {
		return completedTrade, err
	} else if prevTrade != nil && completedTrade {
		if prevTrade.CumulativeFillPrice == nil {
			return completedTrade, fmt.Errorf("no cumulative fill price for trade %d", prevTrade.ID)
		} else if !prevTrade.Proceeds.Valid || prevTrade.Proceeds.BigInt.Cmp(common.Big0) <= 0 {
			return completedTrade, fmt.Errorf("no proceeds for trade %d", prevTrade.ID)
		} else if !prevTrade.CumulativeDepositFees.Valid {
			return completedTrade, fmt.Errorf("no cumulative deposit fees for trade %d", prevTrade.ID)
		} else if !prevTrade.CumulativeTradeFees.Valid {
			return completedTrade, fmt.Errorf("no cumulative trade fees for trade %d", prevTrade.ID)
		}

		// set cumulative fill price to previous cumulative fill price
		cumulativeFillPrice = types.Float64Value(prevTrade.CumulativeFillPrice)

		// adjust previous deposit fees to be proprtional to the amount
		// the current trade used (against the total value of the previous trade)
		previousDepositFees := prevTrade.CumulativeDepositFees.BigInt
		previousDepositFees.Mul(previousDepositFees, trade.Value.BigInt)
		previousDepositFees.Div(previousDepositFees, prevTrade.Proceeds.BigInt)

		// adjust previous deposit fees to be proprtional to the amount
		// the current trade used (against the total value of the previous trade)
		previousTradeFees := prevTrade.CumulativeTradeFees.BigInt
		previousTradeFees.Mul(previousTradeFees, trade.Value.BigInt)
		previousTradeFees.Div(previousTradeFees, prevTrade.Proceeds.BigInt)

		// adjust cumulative deposit and trade fees to the current chain's price
		if fillPrice > 0 {
			switch types.TradeDirection(trade.Direction) {
			case types.TradeBuy:
				cumulativeFillPrice /= fillPrice
				cumulativeDepositFees = common.BigIntToFloat64(previousDepositFees, fromUnits) / fillPrice
				cumulativeTradeFees = common.BigIntToFloat64(previousTradeFees, fromUnits) / fillPrice
			case types.TradeSell:
				cumulativeFillPrice *= fillPrice
				cumulativeDepositFees = common.BigIntToFloat64(previousDepositFees, fromUnits) * fillPrice
				cumulativeTradeFees = common.BigIntToFloat64(previousTradeFees, fromUnits) * fillPrice
			}
		}
	} else {
		if !trade.CumulativeDepositFees.Valid {
			trade.CumulativeDepositFees = dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)}
		}

		if fillPrice > 0 {
			depositFees := trade.CumulativeDepositFees.BigInt
			switch types.TradeDirection(trade.Direction) {
			case types.TradeBuy:
				cumulativeFillPrice = 1 / fillPrice
				cumulativeDepositFees = common.BigIntToFloat64(depositFees, fromUnits) / fillPrice
			case types.TradeSell:
				cumulativeFillPrice = fillPrice
				cumulativeDepositFees = common.BigIntToFloat64(depositFees, fromUnits) * fillPrice
			}
		}
	}

	// sum cumulative deposit and trade fees
	cumulativeTradeFees += common.BigIntToFloat64(fees, toUnits)
	cumulativeDepositFeesBig := common.Float64ToBigInt(cumulativeDepositFees, toUnits)
	cumulativeTradeFeesBig := common.Float64ToBigInt(cumulativeTradeFees, toUnits)

	// check for the next trade to update the trade value, since it is only known
	// after the previous (current) trade is filled. if no next trade exists, transfer
	// the balance from the trade account to the main account (kucoin only, empty method otherwise).
	// if there are multiple trade steps, this will just partially update the next trade's value
	// sequentially.
	nextTrade, err := pooldb.GetExchangeTradeByPathAndStage(tx, batchID, trade.PathID, trade.StageID+1)
	if err != nil {
		return completedTrade, err
	} else if nextTrade == nil {
		proceedsFloat := common.BigIntToFloat64(proceeds, toUnits)
		err = exchange.TransferToMainAccount(trade.ToChainID, proceedsFloat)
		if err != nil {
			return completedTrade, err
		}
	} else {
		if nextTrade.Value.Valid {
			nextTrade.Value.BigInt.Add(nextTrade.Value.BigInt, proceeds)
		} else {
			nextTrade.Value = dbcl.NullBigInt{Valid: true, BigInt: proceeds}
		}

		cols := []string{"value"}
		err = pooldb.UpdateExchangeTrade(tx, nextTrade, cols)
		if err != nil {
			return completedTrade, err
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

	cols := []string{"value", "proceeds", "trade_fees", "cumulative_deposit_fees",
		"cumulative_trade_fees", "fill_price", "cumulative_fill_price",
		"slippage", "confirmed"}
	err = pooldb.UpdateExchangeTrade(tx, trade, cols)
	if err != nil {
		return completedTrade, err
	}

	c.telegram.NotifyFinalizeTrade(trade.ID)
	err = tx.SafeCommit()

	return completedTrade, err
}

func (c *Client) ConfirmTradeStage(batchID uint64, exchange types.Exchange, stage int) error {
	var tradeStageStatus Status
	switch stage {
	case 1:
		tradeStageStatus = TradesCompleteStageOne
	case 2:
		tradeStageStatus = TradesCompleteStageTwo
	case 3:
		tradeStageStatus = TradesCompleteStageThree
	default:
		return fmt.Errorf("unsupported trade stage %d", stage)
	}

	trades, err := pooldb.GetExchangeTradesByStage(c.pooldb.Reader(), batchID, stage)
	if err != nil {
		return err
	}

	completedAll := true
	for _, trade := range trades {
		completedTrade, err := c.confirmTrade(batchID, exchange, stage, trade)
		if err != nil {
			return err
		} else if !completedTrade {
			completedTrade = false
		}
	}

	if completedAll {
		return c.updateBatchStatus(c.pooldb.Writer(), batchID, tradeStageStatus)
	}

	return nil
}

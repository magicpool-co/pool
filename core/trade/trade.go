package trade

import (
	"fmt"
	"strconv"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) InitiateTrades(batchID uint64) error {
	balanceInputs, err := pooldb.GetExchangeInputs(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	} else if len(balanceInputs) == 0 {
		return nil
	}

	outputPaths, err := c.exchangeInputsToOutputPaths(balanceInputs)
	if err != nil {
		return err
	}

	var pathID int
	trades := make([]*pooldb.ExchangeTrade, 0)
	for fromChainID, outputIdx := range outputPaths {
		for toChainID, value := range outputIdx {
			parsedTrades, err := c.exchange.GenerateTradePath(fromChainID, toChainID)
			if err != nil {
				return err
			}

			pathID++
			for i, parsedTrade := range parsedTrades {
				var tradeValue dbcl.NullBigInt
				if i == 0 {
					tradeValue = dbcl.NullBigInt{Valid: true, BigInt: value}
				}

				trade := &pooldb.ExchangeTrade{
					BatchID: batchID,
					Path:    pathID,
					Stage:   i + 1,

					FromChainID: parsedTrade.FromChain,
					ToChainID:   parsedTrade.ToChain,
					Market:      parsedTrade.Market,
					Direction:   int(parsedTrade.Direction),

					Value: tradeValue,
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

		units, err := common.GetDefaultUnits(trade.FromChainID)
		if err != nil {
			return err
		}

		quantity := common.BigIntToFloat64(trade.Value.BigInt, units)
		tradeID, err := c.exchange.CreateTrade(trade.Market, types.TradeDirection(trade.Direction), quantity)
		if err != nil {
			return err
		}

		trade.ExchangeTradeID = types.StringPtr(tradeID)
		trade.Initiated = true
		trade.Open = true

		cols := []string{"exchange_trade_id", "initiated", "open"}
		err = pooldb.UpdateExchangeTrade(c.pooldb.Writer(), trade, cols)
		if err != nil {
			return err
		}
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
		if trade.Initiated && !trade.Open && trade.Filled {
			continue
		} else if trade.ExchangeTradeID == nil {
			return fmt.Errorf("trade %d has no exchange trade id", trade.ID)
		} else if !trade.Value.Valid {
			return fmt.Errorf("no value for trade %d", trade.ID)
		} else if trade.OrderPrice == nil {
			return fmt.Errorf("no order price for trade %d", trade.ID)
		}

		fromUnits, err := common.GetDefaultUnits(trade.FromChainID)
		if err != nil {
			return err
		}

		toUnits, err := common.GetDefaultUnits(trade.ToChainID)
		if err != nil {
			return err
		}

		initialQuantity := common.BigIntToFloat64(trade.Value.BigInt, fromUnits)
		parsedTrade, err := c.exchange.GetTradeByID(trade.Market, initialQuantity)
		if err != nil {
			return err
		} else if parsedTrade == nil || !parsedTrade.Completed {
			completedAll = false
			continue
		}

		proceeds, err := common.StringDecimalToBigint(parsedTrade.Proceeds, toUnits)
		if err != nil {
			return err
		}

		fees, err := common.StringDecimalToBigint(parsedTrade.Fees, toUnits)
		if err != nil {
			return err
		}

		fillPrice, err := strconv.ParseFloat(parsedTrade.Price, 10)
		if err != nil {
			return err
		}

		var cumulativeDepositFees, cumulativeTradeFees float64
		prevTrade, err := pooldb.GetExchangeTradeByPathAndStage(c.pooldb.Reader(), batchID, trade.Path, trade.Stage-1)
		if err != nil {
			return err
		} else if prevTrade != nil {
			if !prevTrade.CumulativeDepositFees.Valid {
				return fmt.Errorf("no cumulative deposit fees for trade %d", prevTrade.ID)
			} else if !prevTrade.CumulativeTradeFees.Valid {
				return fmt.Errorf("no cumulative trade fees for trade %d", prevTrade.ID)
			}

			cumulativeDepositFees = fillPrice * common.BigIntToFloat64(prevTrade.CumulativeDepositFees.BigInt, fromUnits)
			cumulativeTradeFees = fillPrice * common.BigIntToFloat64(prevTrade.CumulativeTradeFees.BigInt, fromUnits)
		} else {
			if !trade.CumulativeDepositFees.Valid {
				return fmt.Errorf("no cumulative deposit fees for trade %d", prevTrade.ID)
			}

			cumulativeDepositFees = fillPrice * common.BigIntToFloat64(trade.CumulativeDepositFees.BigInt, fromUnits)
		}
		cumulativeTradeFees += common.BigIntToFloat64(fees, toUnits)
		cumulativeDepositFeesBig := common.Float64ToBigInt(cumulativeDepositFees, toUnits)
		cumulativeTradeFeesBig := common.Float64ToBigInt(cumulativeTradeFees, toUnits)

		nextTrade, err := pooldb.GetExchangeTradeByPathAndStage(c.pooldb.Reader(), batchID, trade.Path, trade.Stage+1)
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

		var slippage float64
		orderPrice := types.Float64Value(trade.OrderPrice)
		if types.TradeDirection(trade.Direction) == types.TradeSell {
			slippage = (fillPrice - orderPrice) / orderPrice
		} else {
			slippage = (orderPrice - fillPrice) / fillPrice
		}

		trade.Proceeds = dbcl.NullBigInt{Valid: true, BigInt: proceeds}
		trade.TradeFees = dbcl.NullBigInt{Valid: true, BigInt: fees}
		trade.CumulativeDepositFees = dbcl.NullBigInt{Valid: true, BigInt: cumulativeDepositFeesBig}
		trade.CumulativeTradeFees = dbcl.NullBigInt{Valid: true, BigInt: cumulativeTradeFeesBig}
		trade.FillPrice = types.Float64Ptr(fillPrice)
		trade.Slippage = types.Float64Ptr(slippage)
		trade.Open = false
		trade.Filled = true

		cols := []string{"open", "filled"}
		err = pooldb.UpdateExchangeTrade(c.pooldb.Writer(), trade, cols)
		if err != nil {
			return err
		}
	}

	if completedAll {
		return c.updateBatchStatus(batchID, tradeStageStatus)
	}

	return nil
}

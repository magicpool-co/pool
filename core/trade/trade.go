package trade

import (
	"fmt"
	"math/big"

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
	for inChainID, outputIdx := range outputPaths {
		for outChainID, value := range outputIdx {
			localTrades, err := c.exchange.GenerateTradePath(inChainID, outChainID, value)
			if err != nil {
				return err
			}

			pathID++
			for i, trade := range localTrades {
				trade.BatchID = batchID
				trade.Path = pathID
				trade.Stage = i + 1
			}

			trades = append(trades, localTrades...)
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
		}

		var direction string
		switch trade.Direction {
		case 0:
			direction = "BUY"
		case 1:
			direction = "SELL"
		}

		units, err := common.GetDefaultUnits(trade.FromChainID)
		if err != nil {
			return err
		}

		quantity := common.BigIntToFloat64(trade.Value.BigInt, units)
		if trade.Increment > 0 {
			var remainder *big.Int
			quantity, remainder = common.FloorFloatByIncrement(quantity, trade.Increment, 1e8)
			trade.Remainder = dbcl.NullBigInt{Valid: true, BigInt: remainder}
		}

		tradeID, err := c.exchange.CreateOrder(trade.Market, direction, quantity)
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
		}

		tradeID := types.StringValue(trade.ExchangeTradeID)
		completed, err := c.exchange.GetOrderStatus("", tradeID)
		if err != nil {
			return err
		} else if !completed {
			completedAll = false
			continue
		}

		nextTrade, err := pooldb.GetExchangeTradeByPathAndStage(c.pooldb.Reader(), batchID, trade.Path, trade.Stage+1)
		if err != nil {
			return err
		} else if nextTrade == nil {
			// @TODO: transfer balance properly
			err = c.exchange.TransferToMainAccount(trade.ToChainID, 0)
			if err != nil {
				return err
			}
		} else {
			// @TODO: handle next trade properly
		}

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

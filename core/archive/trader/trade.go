package trader

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/magicpool-co/pool/pkg/accounting"
	"github.com/magicpool-co/pool/pkg/bittrex"
	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/types"
	"github.com/magicpool-co/pool/pkg/utils"
)

/* initiate trades */

func (c *client) initiateTrades(profitSwitch *db.Switch) error {
	accountant := accounting.NewSwitchAccountant()

	if balances, err := c.db.GetBalancesBySwitch(profitSwitch.ID); err != nil {
		return err
	} else {
		for _, balance := range balances {
			if err := accountant.AddBalance(balance, false); err != nil {
				return err
			}
		}
	}

	costBasisPrices := make(map[string]float64)
	if deposits, err := c.db.GetDepositsBySwitch(profitSwitch.ID); err != nil {
		return err
	} else {
		for _, deposit := range deposits {
			if err := accountant.AddDeposit(deposit); err != nil {
				return err
			} else if !deposit.CostBasisPrice.Valid {
				return fmt.Errorf("deposit %d has no valid CostBasisPrice", deposit.ID)
			}

			costBasisPrices[deposit.CoinID] = deposit.CostBasisPrice.Float64
		}
	}

	if fullTrades, err := accountant.GenerateTrades(profitSwitch.ID); err != nil {
		return err
	} else {
		for _, trades := range fullTrades {
			var nextTradeID sql.NullInt64
			// insert trades backwards to update NextTradeID
			for i := len(trades) - 1; i >= 0; i-- {
				// add initial cost basis price for first trade
				if i == 0 {
					trades[i].CostBasisPrice = sql.NullFloat64{Valid: true, Float64: costBasisPrices[trades[i].FromCoin]}
				}

				trades[i].NextTradeID = nextTradeID
				if tradeID, err := c.db.InsertTrade(trades[i]); err != nil {
					return err
				} else {
					nextTradeID = sql.NullInt64{int64(tradeID), true}
				}
			}
		}
	}

	profitSwitch.Status = int(types.TradesInactive)
	if err := c.db.UpdateSwitch(profitSwitch, []string{"status"}); err != nil {
		return err
	}

	return nil
}

/* execute trades */

func (c *client) executeTradeStage(profitSwitch *db.Switch, stage int) error {
	if trades, err := c.db.GetUninitiatedTradesBySwitch(profitSwitch.ID, stage); err != nil {
		return err
	} else {
		for _, trade := range trades {
			if err := c.executeSingleTrade(trade); err != nil {
				return err
			}

			// give it a minor sleep to see if it solves the unfilled problem
			time.Sleep(time.Second * 3)
		}
	}

	// if an order is never completely filled, it will write to the db but get
	// stuck at this stage, which is the ideal behavior initially
	if unfilled, err := c.db.GetUnfilledTradesBySwitch(profitSwitch.ID); err != nil {
		return err
	} else if len(unfilled) == 0 {
		var status int
		if stage == 1 {
			status = int(types.TradesCompleteStageOne)
		} else if stage == 2 {
			status = int(types.TradesCompleteStageTwo)
		} else {
			return fmt.Errorf("%d is an unsupported trading stage", stage)
		}

		profitSwitch.Status = status
		if err := c.db.UpdateSwitch(profitSwitch, []string{"status"}); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) executeSingleTrade(trade *db.SwitchTrade) error {
	var tradeQuantity float64
	if !trade.Value.Valid {
		return fmt.Errorf("invalid bigint quantity for trade %d", trade.ID)
	} else if node, ok := c.nodes[trade.FromCoin]; !ok {
		return fmt.Errorf("unable to find node for %s", trade.FromCoin)
	} else {
		units := node.GetUnits().Big()
		tradeQuantity = utils.BigintToFloat64(trade.Value.BigInt, units)
	}

	var direction string
	switch trade.Direction {
	case int(types.BuyDirection):
		direction = "BUY"
	case int(types.SellDirection):
		direction = "SELL"
	default:
		return fmt.Errorf("invalid direction type %d for trade %d", trade.Direction, trade.ID)
	}

	expectedRate, err := c.bittrex.GetRate(trade.FromCoin, trade.ToCoin)
	if err != nil {
		return err
	}

	usdFromPrice, err := c.bittrex.GetRate(trade.FromCoin, "USD")
	if err != nil {
		return err
	}

	usdToPrice, err := c.bittrex.GetRate(trade.ToCoin, "USD")
	if err != nil {
		return err
	}

	order, err := c.bittrex.CreateOrder(trade.Market, direction, tradeQuantity)
	if err != nil {
		return err
	}

	switch bittrex.OrderStatus(order.Status) {
	case bittrex.ORDER_OPEN:
		return fmt.Errorf("order %s is still open, bailing", order.ID)
	case bittrex.ORDER_CLOSED:
		var slippage float64
		if order.Rate != 0.0 {
			if direction == "SELL" {
				slippage = (order.Rate - expectedRate) / expectedRate
			} else {
				slippage = (expectedRate - order.Rate) / order.Rate
			}
		}

		trade.BittrexID = sql.NullString{order.ID, true}
		trade.Initiated = true
		trade.Open = false
		trade.Filled = order.Filled
		trade.Value = db.NullBigInt{order.InputAmount, true}
		trade.Proceeds = db.NullBigInt{order.OutputAmount, true}
		trade.Fees = db.NullBigInt{order.Fees, true}
		trade.Slippage = sql.NullFloat64{slippage, true}
		trade.FairMarketPrice = sql.NullFloat64{usdFromPrice, true}

		cols := []string{"bittrex_id", "initiated", "open", "filled", "value",
			"proceeds", "fees", "slippage", "fair_market_price"}
		if err = c.db.UpdateTrade(trade, cols); err != nil {
			return err
		} else if trade.NextTradeID.Valid {
			nextTrade := &db.SwitchTrade{
				ID:             uint64(trade.NextTradeID.Int64),
				Value:          db.NullBigInt{order.OutputAmount, true},
				CostBasisPrice: sql.NullFloat64{usdToPrice, true},
			}

			cols := []string{"value", "cost_basis_price"}
			if err := c.db.UpdateTrade(nextTrade, cols); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown status %s for order %s", order.Status, order.ID)
	}

	return nil
}

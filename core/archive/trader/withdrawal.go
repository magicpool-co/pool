package trader

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/magicpool-co/pool/core/banker"
	"github.com/magicpool-co/pool/pkg/accounting"
	"github.com/magicpool-co/pool/pkg/bittrex"
	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/types"
	"github.com/magicpool-co/pool/pkg/utils"
)

func (c *client) initiateWithdrawals(profitSwitch *db.Switch) error {
	accountant := accounting.NewSwitchAccountant()

	if paths, err := c.db.GetTradePathsBySwitch(profitSwitch.ID); err != nil {
		return err
	} else {
		for _, path := range paths {
			if err := accountant.AddInitialPath(profitSwitch.ID, path); err != nil {
				return err
			}
		}
	}

	for _, withdrawal := range accountant.Withdrawals() {
		coin := withdrawal.CoinID

		if currency, err := c.bittrex.GetCurrency(coin); err != nil {
			return err
		} else if currency.Status != "ONLINE" {
			return fmt.Errorf("bittrex wallet for %s is not online", coin)
		}

		node, ok := c.nodes[coin]
		if !ok {
			return fmt.Errorf("unable to find node for %s", coin)
		}

		units := node.GetUnits().Big()
		amount := utils.BigintToFloat64(withdrawal.Value.BigInt, units)
		address := node.GetWallet().Address

		withdrawalRes, err := c.bittrex.Withdraw(address, coin, amount)
		if err != nil {
			return err
		}

		quantity, err := utils.StringDecimalToBigint(withdrawalRes.Quantity, units)
		if err != nil {
			return err
		}

		fees, err := utils.StringDecimalToBigint(withdrawalRes.TxCost, units)
		if err != nil {
			return err
		}

		withdrawal.BittrexID = withdrawalRes.ID
		withdrawal.Value = db.NullBigInt{quantity, true}
		withdrawal.Fees = db.NullBigInt{fees, true}

		if _, err := c.db.InsertWithdrawal(withdrawal); err != nil {
			return err
		}
	}

	profitSwitch.Status = int(types.WithdrawalsActive)
	if err := c.db.UpdateSwitch(profitSwitch, []string{"status"}); err != nil {
		return err
	}

	return nil
}

func (c *client) confirmWithdrawals(profitSwitch *db.Switch) error {
	activeWithdrawals, err := c.db.GetActiveWithdrawalsBySwitch(profitSwitch.ID)
	if err != nil {
		return err
	}

	withdrawalsComplete := true
	for _, active := range activeWithdrawals {
		withdrawal, err := c.bittrex.GetWithdrawalByID(active.BittrexID)
		if err != nil {
			return err
		}

		switch bittrex.WithdrawalStatus(withdrawal.Status) {
		case bittrex.WITHDRAWAL_REQUESTED:
			withdrawalsComplete = false
		case bittrex.WITHDRAWAL_AUTHORIZED:
			withdrawalsComplete = false
		case bittrex.WITHDRAWAL_PENDING:
			withdrawalsComplete = false
		case bittrex.WITHDRAWAL_COMPLETED:
			node, ok := c.nodes[active.CoinID]
			if !ok {
				return fmt.Errorf("unable to find node for withdrawal %d", active.ID)
			}

			tx, err := node.GetTx(withdrawal.TxID)
			if err != nil {
				return err
			} else if !tx.Confirmed {
				return fmt.Errorf("tx %s not confirmed for withdrawal %d", withdrawal.TxID, active.ID)
			}

			active.TxID = sql.NullString{withdrawal.TxID, true}
			active.Height = sql.NullInt64{int64(tx.BlockNumber), true}
			active.Confirmed = tx.Confirmed
			if err := c.db.UpdateWithdrawal(active, []string{"txid", "height", "confirmed"}); err != nil {
				return err
			}

			if err := banker.AddWithdrawal(c.db, active, tx, node.GetWallet().Address); err != nil {
				return err
			}
		case bittrex.WITHDRAWAL_CANCELLED:
			return fmt.Errorf("withdrawal %d is cancelled", active.ID)
		case bittrex.WITHDRAWAL_ERROR_INVALID_ADDRESS:
			return fmt.Errorf("withdrawal %d has an invalid address", active.ID)
		default:
			return fmt.Errorf("unknown status %s for withdrawal %d",
				withdrawal.Status, active.ID)
		}
	}

	if withdrawalsComplete {
		profitSwitch.Status = int(types.WithdrawalsComplete)
		err = c.db.UpdateSwitch(profitSwitch, []string{"status"})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) creditWithdrawals(profitSwitch *db.Switch) error {
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

	for _, payoutID := range accountant.PayoutIDs() {
		if payout, err := c.db.GetPayout(payoutID); err != nil {
			return err
		} else if payout.ID != payoutID {
			return fmt.Errorf("unable to find payout %d in db", payoutID)
		} else {
			if err := accountant.AddPayout(payout); err != nil {
				return err
			}
		}
	}

	if deposits, err := c.db.GetDepositsBySwitch(profitSwitch.ID); err != nil {
		return err
	} else {
		for _, deposit := range deposits {
			if err := accountant.AddDeposit(deposit); err != nil {
				return err
			}
		}
	}

	if paths, err := c.db.GetTradePathsBySwitch(profitSwitch.ID); err != nil {
		return err
	} else {
		for _, path := range paths {
			if err := accountant.AddFinalPath(path); err != nil {
				return err
			}
		}
	}

	if withdrawals, err := c.db.GetWithdrawalsBySwitch(profitSwitch.ID); err != nil {
		return err
	} else {
		for _, withdrawal := range withdrawals {
			if err := accountant.AddWithdrawal(withdrawal); err != nil {
				return err
			}
		}

		if err := accountant.Distribute(); err != nil {
			return err
		}

		for _, withdrawal := range withdrawals {
			withdrawal.Spent = true
			if err := c.db.UpdateWithdrawal(withdrawal, []string{"spent"}); err != nil {
				return err
			}
		}
	}

	for _, balance := range accountant.Balances() {
		balance.Pending = false
		cols := []string{"pending", "pool_fees", "exchange_fees", "out_value"}
		if err := c.db.UpdateBalance(balance, cols); err != nil {
			return err
		}
	}

	for _, payout := range accountant.Payouts() {
		cols := []string{"value", "pool_fees", "exchange_fees",
			"fee_balance_pending", "in_fee_balance"}
		if err := c.db.UpdatePayout(payout, cols); err != nil {
			return err
		}
	}

	profitSwitch.Status = int(types.SwitchComplete)
	profitSwitch.CompletedAt = sql.NullTime{time.Now().UTC(), true}

	cols := []string{"status", "completed_at"}
	if err := c.db.UpdateSwitch(profitSwitch, cols); err != nil {
		return err
	}

	if c.telegram != nil {
		c.telegram.NotifyFinalizeSwitch(profitSwitch.ID)
	}

	return nil
}

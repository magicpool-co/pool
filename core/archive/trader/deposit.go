package trader

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/core/banker"
	"github.com/magicpool-co/pool/core/creditor"
	"github.com/magicpool-co/pool/pkg/accounting"
	"github.com/magicpool-co/pool/pkg/bittrex"
	"github.com/magicpool-co/pool/pkg/config"
	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/types"
	"github.com/magicpool-co/pool/pkg/utils"
)

func (c *client) initiateDeposits(profitSwitch *db.Switch) error {
	accountant := accounting.NewSwitchAccountant()
	balances, err := c.db.GetBalancesBySwitch(profitSwitch.ID)
	if err != nil {
		return err
	}

	rounds, err := c.db.GetRoundsBySwitch(profitSwitch.ID)
	if err != nil {
		return err
	}

	roundPrices := make(map[uint64]float64)
	for _, round := range rounds {
		if round.CostBasisPrice.Valid {
			roundPrices[round.ID] = round.CostBasisPrice.Float64
		}
	}

	sumWeightedPrice := make(map[string]*big.Float)
	for _, balance := range balances {
		if balance.InRoundID.Valid && balance.InValue.Valid {
			if _, ok := sumWeightedPrice[balance.InCoin]; !ok {
				sumWeightedPrice[balance.InCoin] = new(big.Float)
			} else if _, ok := roundPrices[uint64(balance.InRoundID.Int64)]; !ok {
				return fmt.Errorf("round %d has no cost basis price", balance.InRoundID.Int64)
			}

			weightedPrice := new(big.Float).SetFloat64(roundPrices[uint64(balance.InRoundID.Int64)])
			weightedPrice.Mul(weightedPrice, new(big.Float).SetInt(balance.InValue.BigInt))
			sumWeightedPrice[balance.InCoin].Add(sumWeightedPrice[balance.InCoin], weightedPrice)
		}

		if err := accountant.AddBalance(balance, false); err != nil {
			return err
		}
	}

	effectivePrice := make(map[string]float64)
	for coin, value := range accountant.InputValues() {
		if value.Cmp(new(big.Int)) != 1 {
			return fmt.Errorf("switch %d has nothing to trade for %s",
				profitSwitch.ID, coin)
		} else if _, ok := sumWeightedPrice[coin]; !ok {
			return fmt.Errorf("switch %d has no sum weighted price for %s",
				profitSwitch.ID, coin)
		}

		weightedPrice := new(big.Float).Quo(sumWeightedPrice[coin], new(big.Float).SetInt(value))
		effectivePrice[coin], _ = weightedPrice.Float64()
	}

	depositedAll := true
	for coin, value := range accountant.InputValues() {
		if deposit, err := c.db.GetDepositBySwitchCoin(profitSwitch.ID, coin); err != nil {
			return err
		} else if deposit.ID != 0 {
			// continue if already a registered deposit for this switch + coin
			continue
		}

		unregisteredDeposits, err := c.db.GetUnregisteredDepositsByCoin(coin)
		if err != nil {
			return err
		}

		// make sure no other active deposits are present
		if len(unregisteredDeposits) > 0 {
			depositedAll = false
			continue
		}

		depositAddress, err := c.bittrex.GetDepositAddress(coin)
		if err != nil {
			return err
		}

		currency, err := c.bittrex.GetCurrency(coin)
		if err != nil {
			return err
		} else if currency.Status != "ONLINE" {
			return fmt.Errorf("bittrex wallet for %s is not online", coin)
		}

		conf := &config.Config{
			Bittrex:     c.bittrex,
			PayoutNodes: c.nodes,
			DB:          c.db,
		}

		address := depositAddress.CryptoAddress
		tx, err := banker.SpendDeposit(conf, coin, value, address)
		if err != nil {
			return err
		}

		deposit := &db.SwitchDeposit{
			CoinID:         coin,
			SwitchID:       profitSwitch.ID,
			TxID:           tx.TxID,
			Value:          db.NullBigInt{value, true},
			CostBasisPrice: sql.NullFloat64{Valid: true, Float64: effectivePrice[coin]},
			Registered:     false,
			Pending:        true,
			Spent:          false,
		}

		depositID, err := c.db.InsertDeposit(deposit)
		if err != nil {
			return err
		}

		if utxoInputs, err := c.db.GetUTXOsByOutTxID(tx.TxID); err != nil {
			return err
		} else {
			for _, utxo := range utxoInputs {
				utxo.OutDepositID = sql.NullInt64{int64(depositID), true}
				err := c.db.UpdateUTXO(utxo, []string{"out_deposit_id"})
				if err != nil {
					return err
				}
			}
		}

		if utxoOutputs, err := c.db.GetUTXOsByInTxID(tx.TxID); err != nil {
			return err
		} else {
			for _, utxo := range utxoOutputs {
				utxo.InDepositID = sql.NullInt64{int64(depositID), true}
				err := c.db.UpdateUTXO(utxo, []string{"in_deposit_id"})
				if err != nil {
					return err
				}
			}
		}
	}

	if depositedAll {
		profitSwitch.Status = int(types.DepositUnregistered)
		err = c.db.UpdateSwitch(profitSwitch, []string{"status"})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) registerDeposits(profitSwitch *db.Switch) error {
	registeredAll := true
	deposits, err := c.db.GetUnregistedDepositsBySwitch(profitSwitch.ID)
	if err != nil {
		return err
	}

	for _, deposit := range deposits {
		if deposit.Registered {
			continue
		}

		openDeposits, err := c.bittrex.GetOpenDepositsForTicker(deposit.CoinID)
		if err != nil {
			return err
		}

		var depositID string
		if len(openDeposits) == 0 {
			closedDeposits, err := c.bittrex.GetClosedDepositsForTicker(deposit.CoinID)
			if err != nil {
				return err
			} else if len(closedDeposits) > 0 {
				depositID = closedDeposits[0].ID
			} else {
				registeredAll = false
				continue
			}
		} else if length := len(openDeposits); length > 1 {
			return fmt.Errorf("%d open deposits for %s, want one",
				length, deposit.CoinID)
		} else {
			depositID = openDeposits[0].ID
		}

		if len(depositID) == 0 {
			return fmt.Errorf("depositID for deposit %d is empty", deposit.ID)
		} else if existingDeposit, err := c.db.GetDepositByBittrexID(depositID); err != nil {
			return err
		} else if existingDeposit.ID != 0 {
			return fmt.Errorf("deposit %d already uses exchange ID %s", existingDeposit.ID, depositID)
		}

		deposit.Registered = true
		deposit.BittrexID = sql.NullString{depositID, true}
		err = c.db.UpdateDeposit(deposit, []string{"bittrex_id", "registered"})
		if err != nil {
			return err
		}
	}

	if registeredAll {
		profitSwitch.Status = int(types.DepositUnconfirmed)
		err = c.db.UpdateSwitch(profitSwitch, []string{"status"})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) confirmDeposits(profitSwitch *db.Switch) error {
	confirmedAll := true
	deposits, err := c.db.GetPendingDepositsBySwitch(profitSwitch.ID)
	if err != nil {
		return err
	}

	for _, deposit := range deposits {
		if !deposit.BittrexID.Valid {
			return fmt.Errorf("invalid value for BittrexID on deposit %d", deposit.ID)
		} else if !deposit.Value.Valid {
			return fmt.Errorf("invalid value for MinerValue on deposit %d", deposit.ID)
		}

		rawDeposit, err := c.bittrex.GetDepositByID(deposit.BittrexID.String)
		if err != nil {
			return err
		}

		switch bittrex.DepositStatus(rawDeposit.Status) {
		case bittrex.DEPOSIT_COMPLETED:
		case bittrex.DEPOSIT_PENDING:
			confirmedAll = false
			continue
		case bittrex.DEPOSIT_ORPHANED:
			return fmt.Errorf("deposit %d has been orphaned", deposit.ID)
		case bittrex.DEPOSIT_INVALIDATED:
			return fmt.Errorf("deposit %d has been invalidated", deposit.ID)
		default:
			return fmt.Errorf("unknown status %s for deposit %d", rawDeposit.Status, deposit.ID)
		}

		node, ok := c.nodes[deposit.CoinID]
		if !ok {
			return fmt.Errorf("unable to find node for %s", deposit.CoinID)
		}

		// fee balances on deposits only exist for ETH bc of EIP1559
		var tx *types.RawTx
		if deposit.CoinID == "ETH" {
			tx, err = node.GetTx(deposit.TxID)
			if err != nil {
				return err
			} else if tx == nil || !tx.Confirmed || tx.FeeBalance == nil {
				continue
			}
		}

		units := node.GetUnits().Big()
		rawDepositValue, err := utils.StringDecimalToBigint(rawDeposit.Quantity, units)
		if err != nil {
			return err
		}

		// calculate the fees fees, spread them out for everyone
		fees := new(big.Int).Sub(deposit.Value.BigInt, rawDepositValue)
		if fees.Cmp(new(big.Int)) <= 0 {
			return fmt.Errorf("no value in deposit %d: have %s",
				deposit.ID, rawDepositValue)
		}

		deposit.Pending = false
		deposit.BittrexTxID = sql.NullString{rawDeposit.TxID, true}
		deposit.Value = db.NullBigInt{rawDepositValue, true}
		deposit.Fees = db.NullBigInt{fees, true}

		cols := []string{"bittrex_txid", "value", "fees", "pending"}
		err = c.db.UpdateDeposit(deposit, cols)
		if err != nil {
			return err
		}

		if tx != nil && tx.FeeBalance.Cmp(big.NewInt(0)) > 0 {
			// "randomly" choose a recipient to issue the deposit fee balance
			recipientID := deposit.ID%2 + 1
			recipient, err := c.db.GetRecipient(recipientID)
			if err != nil {
				return err
			} else if recipient == nil || recipient.ID != recipientID {
				return fmt.Errorf("unable to find recipient to distribute ETH fee balance of %s to", tx.FeeBalance)
			}

			payout, err := creditor.GetOrCreatePayout(c.db, "recipient", recipient.CoinID, recipient.Address, recipient.ID)
			if err != nil {
				return err
			}

			feeBalance := &db.Balance{
				RecipientID: sql.NullInt64{int64(recipientID), true},

				InCoin:      deposit.CoinID,
				InValue:     db.NullBigInt{tx.FeeBalance, true},
				InDepositID: sql.NullInt64{int64(deposit.ID), true},

				Pending:      false,
				PoolFees:     db.NullBigInt{new(big.Int), true},
				ExchangeFees: db.NullBigInt{new(big.Int), true},

				OutCoin:     deposit.CoinID,
				OutValue:    db.NullBigInt{tx.FeeBalance, true},
				OutPayoutID: payout.ID,
			}

			if _, err := c.db.InsertBalance(feeBalance); err != nil {
				return err
			}

			if !payout.Value.Valid {
				return fmt.Errorf("invalid value for payout %d", payout.ID)
			} else {
				payout.Value.BigInt.Add(payout.Value.BigInt, tx.FeeBalance)
				if err := c.db.UpdatePayout(payout, []string{"value"}); err != nil {
					return err
				}
			}

			if err := banker.AddDepositChange(c.db, deposit, tx.FeeBalance); err != nil {
				return err
			}
		}
	}

	if confirmedAll {
		profitSwitch.Status = int(types.DepositComplete)
		err = c.db.UpdateSwitch(profitSwitch, []string{"status"})
		if err != nil {
			return err
		}
	}

	return nil
}

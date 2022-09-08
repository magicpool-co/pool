package payout

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/magicpool-co/pool/core/banker"
	"github.com/magicpool-co/pool/pkg/config"
	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/types"
)

func PayoutBalances(conf *config.Config) error {
	for _, chain := range []string{"BTC", "ETH", "USDC"} {
		threshold, err := types.DefaultPayoutThreshold(chain)
		if err != nil {
			return err
		}

		payouts, err := conf.DB.GetUnspentPayoutsAboveThresholdByChain(chain, threshold.String())
		if err != nil {
			return err
		} else if len(payouts) == 0 {
			continue
		}

		for _, payout := range payouts {
			if payout.FeeBalancePending {
				return fmt.Errorf("pending fee balance for payout %d", payout.ID)
			} else if !payout.Value.Valid {
				return fmt.Errorf("invalid Value for payout %d", payout.ID)
			} else if !payout.PoolFees.Valid {
				return fmt.Errorf("invalid PoolFees for payout %d", payout.ID)
			} else if !payout.ExchangeFees.Valid {
				return fmt.Errorf("invalid ExchangeFees for payout %d", payout.ID)
			} else if !payout.InFeeBalance.Valid {
				return fmt.Errorf("invalid InFeeBalance for payout %d", payout.ID)
			} else if payout.Spent {
				return fmt.Errorf("already spent payout %d", payout.ID)
			} else if len(payout.Address) <= 2 {
				return fmt.Errorf("invalid address for payout %d", payout.ID)
			}
		}

		switch chain {
		case "BTC":
			if len(payouts) > 100 {
				payouts = payouts[:100]
			}

			tx, err := sendPayout(conf, payouts)
			if err != nil {
				return err
			}

			err = updatePayouts(conf, chain, tx, payouts)
			if err != nil {
				return err
			}
		case "ETH", "USDC":
			for _, payout := range payouts {
				tx, err := sendPayout(conf, []*db.Payout{payout})
				if err != nil {
					return err
				}

				err = updatePayouts(conf, chain, tx, []*db.Payout{payout})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func sendPayout(conf *config.Config, payouts []*db.Payout) (*types.TxResponse, error) {
	if len(payouts) == 0 {
		return nil, fmt.Errorf("no payouts to send")
	}

	tx, err := banker.SpendPayouts(conf, payouts)
	if err != nil {
		return tx, err
	}

	if utxoInputs, err := conf.DB.GetUTXOsByOutTxID(tx.TxID); err != nil {
		return tx, err
	} else {
		for _, utxo := range utxoInputs {
			utxo.OutPayoutID = sql.NullInt64{int64(payouts[0].ID), true}
			err := conf.DB.UpdateUTXO(utxo, []string{"out_payout_id"})
			if err != nil {
				return tx, err
			}
		}
	}

	if utxoOutputs, err := conf.DB.GetUTXOsByInTxID(tx.TxID); err != nil {
		return tx, err
	} else {
		for _, utxo := range utxoOutputs {
			utxo.InPayoutID = sql.NullInt64{int64(payouts[0].ID), true}
			err := conf.DB.UpdateUTXO(utxo, []string{"in_payout_id"})
			if err != nil {
				return tx, err
			}
		}
	}

	return tx, nil
}

func updatePayouts(conf *config.Config, chain string, tx *types.TxResponse, payouts []*db.Payout) error {
	if conf.Telegram != nil {
		for _, payout := range payouts {
			var name string
			if payout.MinerID.Valid {
				client, err := conf.DB.GetClientByMinerID(uint64(payout.MinerID.Int64))
				if err == nil && client != nil {
					name = client.Name
				}
			} else if payout.RecipientID.Valid {
				name = fmt.Sprintf("fee recipient %d", payout.RecipientID.Int64)
			}

			conf.Telegram.NotifyPayoutSent(chain, tx.TxID, name)
		}
	}

	for _, payout := range payouts {
		switch chain {
		case "ETH":
			payout.Value = db.NullBigInt{tx.Value, true}
			payout.TxFees = db.NullBigInt{tx.Fees, true}
			payout.OutFeeBalance = db.NullBigInt{tx.FeeBalance, true}
		case "USDC":
			// don't change payout value for USDC since it's fixed
			payout.TxFees = db.NullBigInt{tx.Fees, true}
			payout.OutFeeBalance = db.NullBigInt{tx.FeeBalance, true}
		case "BTC":
			payout.TxFees = db.NullBigInt{new(big.Int), true}
			payout.OutFeeBalance = db.NullBigInt{new(big.Int), true}
		}

		payout.Spent = true
		payout.Confirmed = false
		payout.TxID = sql.NullString{tx.TxID, true}
		payout.SpentAt = sql.NullTime{time.Now(), true}

		cols := []string{"value", "spent", "confirmed", "txid", "tx_fees",
			"out_fee_balance", "spent_at"}
		if err := conf.DB.UpdatePayout(payout, cols); err != nil {
			return err
		}

		// create a new payout entry and distribute pending balances
		if err := createOrUpdatePayout(conf.DB, payout); err != nil {
			return err
		}
	}

	return nil
}

func ConfirmPayouts(conf *config.Config) error {
	for chain, node := range conf.PayoutNodes {
		switch chain {
		case "ETC", "RVN":
			continue
		}

		payouts, err := conf.DB.GetUnconfirmedPayoutsByChain(chain)
		if err != nil {
			return err
		}

		txs := make(map[string]*types.RawTx)
		for _, payout := range payouts {
			if !payout.TxID.Valid {
				continue
			} else if _, ok := txs[payout.TxID.String]; ok {
				continue
			} else {
				txs[payout.TxID.String], err = node.GetTx(payout.TxID.String)
				if err != nil {
					return err
				}
			}
		}

		for _, payout := range payouts {
			if !payout.TxID.Valid || !payout.Value.Valid {
				continue
			} else if tx, ok := txs[payout.TxID.String]; ok && tx != nil && tx.Confirmed {
				switch chain {
				case "ETH", "USDC":
					payout.TxFees = db.NullBigInt{tx.Fee, true}
					if chain != "USDC" {
						payout.Value = db.NullBigInt{tx.Value, true}
					}

					if tx.FeeBalance.Cmp(new(big.Int)) > 0 {
						if payout.OutFeeBalance.Valid {
							payout.OutFeeBalance.BigInt.Add(payout.OutFeeBalance.BigInt, tx.FeeBalance)
						} else {
							payout.OutFeeBalance = db.NullBigInt{tx.FeeBalance, true}
						}

						err := banker.AddPayoutChange(conf.DB, payout, payout.OutFeeBalance.BigInt)
						if err != nil {
							return err
						}
					}
				case "BTC":
					outputValue := new(big.Int)
					for _, utxo := range tx.Outputs {
						if utxo.Address == payout.Address {
							outputValue.Add(outputValue, new(big.Int).SetUint64(utxo.Value))
						}
					}

					if outputValue.Cmp(new(big.Int)) == 0 {
						continue
					}

					fees := new(big.Int).Sub(payout.Value.BigInt, outputValue)
					if fees.Cmp(new(big.Int)) < 0 {
						return fmt.Errorf("negative fees for tx %s and address %s", payout.TxID.String, payout.Address)
					}

					payout.Value = db.NullBigInt{outputValue, true}
					payout.TxFees = db.NullBigInt{fees, true}
				}

				payout.Height = sql.NullInt64{int64(tx.BlockNumber), true}
				payout.Confirmed = tx.Confirmed
				payout.ConfirmedAt = sql.NullTime{time.Now(), true}

				cols := []string{"confirmed", "value", "tx_fees",
					"height", "out_fee_balance", "confirmed_at"}
				if err := conf.DB.UpdatePayout(payout, cols); err != nil {
					return err
				}

				if err := createOrUpdatePayout(conf.DB, payout); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func createOrUpdatePayout(dbClient *db.DBClient, oldPayout *db.Payout) error {
	if !oldPayout.Spent {
		return fmt.Errorf("unspent old payout %d", oldPayout.ID)
	}

	inFeeBalance := new(big.Int)
	if oldPayout.Confirmed && oldPayout.OutFeeBalance.Valid {
		inFeeBalance.Add(inFeeBalance, oldPayout.OutFeeBalance.BigInt)
	}

	var nextPayout *db.Payout
	if oldPayout.MinerID.Valid {
		id := uint64(oldPayout.MinerID.Int64)
		unspentPayouts, err := dbClient.GetUnspentPayoutsByMiner(id, oldPayout.CoinID)
		if err != nil {
			return err
		} else if len(unspentPayouts) > 1 {
			return fmt.Errorf("multiple unspent payouts for miner %d %s", id, oldPayout.CoinID)
		} else if len(unspentPayouts) == 1 {
			nextPayout = unspentPayouts[0]
		}
	} else if oldPayout.RecipientID.Valid {
		id := uint64(oldPayout.RecipientID.Int64)
		unspentPayouts, err := dbClient.GetUnspentPayoutsByRecipient(id, oldPayout.CoinID)
		if err != nil {
			return err
		} else if len(unspentPayouts) > 1 {
			return fmt.Errorf("multiple unspent payouts for recipient %d %s", id, oldPayout.CoinID)
		} else if len(unspentPayouts) == 1 {
			nextPayout = unspentPayouts[0]
		}
	} else {
		return fmt.Errorf("invalid old payout %d", oldPayout.ID)
	}

	if nextPayout == nil {
		nextPayout = &db.Payout{
			MinerID:           oldPayout.MinerID,
			RecipientID:       oldPayout.RecipientID,
			CoinID:            oldPayout.CoinID,
			Address:           oldPayout.Address,
			Value:             db.NullBigInt{new(big.Int), true},
			PoolFees:          db.NullBigInt{new(big.Int), true},
			ExchangeFees:      db.NullBigInt{new(big.Int), true},
			FeeBalanceCoin:    oldPayout.FeeBalanceCoin,
			FeeBalancePending: false,
			InFeeBalance:      db.NullBigInt{inFeeBalance, true},
		}

		var err error
		nextPayout.ID, err = dbClient.InsertPayout(nextPayout)
		if err != nil {
			return err
		}
	} else {
		if nextPayout.InFeeBalance.Valid {
			nextPayout.InFeeBalance.BigInt.Add(nextPayout.InFeeBalance.BigInt, inFeeBalance)
		} else {
			nextPayout.InFeeBalance = db.NullBigInt{inFeeBalance, true}
		}

		cols := []string{"in_fee_balance"}
		if err := dbClient.UpdatePayout(nextPayout, cols); err != nil {
			return err
		}
	}

	if err := dbClient.UpdatePendingBalancesSwapOutPayoutID(oldPayout.ID, nextPayout.ID); err != nil {
		return err
	}

	return nil
}

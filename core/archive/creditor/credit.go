package creditor

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/accounting"
	"github.com/magicpool-co/pool/pkg/bittrex"
	"github.com/magicpool-co/pool/pkg/db"
)

func GetOrCreatePayout(dbClient *db.DBClient, payoutType, coin, address string, id uint64) (*db.Payout, error) {
	if len(address) <= 2 {
		return nil, fmt.Errorf("invalid address %s", address)
	}

	var err error
	var unspentPayouts []*db.Payout
	var minerID, recipientID sql.NullInt64

	switch payoutType {
	case "miner":
		minerID = sql.NullInt64{int64(id), true}
		unspentPayouts, err = dbClient.GetUnspentPayoutsByMiner(id, coin)
	case "recipient":
		recipientID = sql.NullInt64{int64(id), true}
		unspentPayouts, err = dbClient.GetUnspentPayoutsByRecipient(id, coin)
	default:
		err = fmt.Errorf("unknown payout type %s", payoutType)
	}

	var unspentPayout *db.Payout
	if err != nil {
		return nil, err
	} else if count := len(unspentPayouts); count > 1 {
		return nil, fmt.Errorf("%d unspent payouts for %s %d", count, payoutType, id)
	} else if len(unspentPayouts) == 1 {
		unspentPayout = unspentPayouts[0]
	} else if len(unspentPayouts) == 0 {
		feeBalanceCoin := coin
		if feeBalanceCoin == "USDC" {
			feeBalanceCoin = "ETH"
		}

		unspentPayout = &db.Payout{
			MinerID:           minerID,
			RecipientID:       recipientID,
			CoinID:            coin,
			Address:           address,
			Value:             db.NullBigInt{new(big.Int), true},
			PoolFees:          db.NullBigInt{new(big.Int), true},
			ExchangeFees:      db.NullBigInt{new(big.Int), true},
			FeeBalanceCoin:    feeBalanceCoin,
			FeeBalancePending: false,
			InFeeBalance:      db.NullBigInt{new(big.Int), true},
		}

		payoutID, err := dbClient.InsertPayout(unspentPayout)
		if err != nil {
			return nil, err
		}

		unspentPayout.ID = payoutID
	}

	return unspentPayout, nil
}

func CreditRounds(dbClient *db.DBClient, bittrexClient bittrex.BittrexClient) error {
	for _, ticker := range []string{"ETH", "ETC", "RVN"} {
		/* get unspent rounds */

		rounds, err := dbClient.GetMatureUnspentRounds(ticker)
		if err != nil {
			return err
		} else if len(rounds) == 0 {
			continue
		}

		/* initialize accountant */

		ethRate, err := bittrexClient.GetRate(ticker, "ETH")
		if err != nil {
			return err
		}

		accountant := accounting.NewRoundAccountant(ethRate)

		/* add rounds to accountant */

		for _, round := range rounds {
			if err := accountant.AddRound(round); err != nil {
				return err
			}
		}

		if err := accountant.FinalizeSetRounds(); err != nil {
			return err
		}

		/* credit rounds */

		for _, round := range rounds {
			if shares, err := dbClient.GetSharesByRound(round.ID); err != nil {
				return err
			} else {
				if err := accountant.CreditRound(round, shares); err != nil {
					return err
				}
			}
		}

		if err := accountant.FinalizeCreditRounds(); err != nil {
			return err
		}

		/* credit miners */

		for _, minerID := range accountant.GetMinerIDs() {
			m, err := dbClient.GetMiner(minerID)
			if err != nil {
				return err
			} else if minerID != m.ID {
				return fmt.Errorf("unable to find miner %d", minerID)
			}

			if payout, err := GetOrCreatePayout(dbClient, "miner", m.CoinID, m.Address, m.ID); err != nil {
				return err
			} else {
				if err := accountant.CreditMiner(m, payout); err != nil {
					return err
				}
			}
		}

		if err := accountant.FinalizeCreditMiners(); err != nil {
			return err
		}

		/* credit recipients */

		if recipients, err := dbClient.GetRecipients(); err != nil {
			return err
		} else {
			for _, r := range recipients {
				if payout, err := GetOrCreatePayout(dbClient, "recipient", r.CoinID, r.Address, r.ID); err != nil {
					return err
				} else {
					if err := accountant.CreditRecipient(r, payout); err != nil {
						return err
					}
				}
			}
		}

		if err := accountant.FinalizeCreditRecipients(); err != nil {
			return err
		}

		/* validate books */

		if err := accountant.ValidateBooks(); err != nil {
			return err
		}

		/* insert balances */

		if balances, err := accountant.GetBalances(); err != nil {
			return err
		} else {
			for _, balance := range balances {
				if _, err := dbClient.InsertBalance(balance); err != nil {
					return err
				}
			}
		}

		/* update payouts */

		if payouts, err := accountant.GetPayouts(); err != nil {
			return err
		} else {
			for _, payout := range payouts {
				cols := []string{"value", "pool_fees", "in_fee_balance", "fee_balance_pending"}
				if err := dbClient.UpdatePayout(payout, cols); err != nil {
					return err
				}
			}
		}

		/* spend rounds */

		for _, round := range rounds {
			round.Spent = true
			if err := dbClient.UpdateRound(round, []string{"spent"}); err != nil {
				return err
			}
		}
	}

	return nil
}

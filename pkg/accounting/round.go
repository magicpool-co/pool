package accounting

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/types"
)

// @TODO: be careful with setting values to existing bigints,
// because it is just a pointer, meaning the value will persist. in
// reality, all situations like that should be handled with a new big int

type roundStage int

var (
	initialized        roundStage = 0
	roundsSet          roundStage = 1
	roundsCredited     roundStage = 2
	minersCredited     roundStage = 3
	recipientsCredited roundStage = 4
	booksValidated     roundStage = 5
)

var (
	usdcFeeBalance = new(big.Int).SetUint64(25000000000000000)
)

type RoundAccountant struct {
	stage     roundStage
	ethRate   float64
	spendable *big.Int
	rounds    []*db.Round
	balances  []*db.Balance
	payouts   []*db.Payout
	miners    map[uint64]bool
	fees      map[uint64]*big.Int
	usedFees  map[uint64]*big.Int
}

func NewRoundAccountant(ethRate float64) *RoundAccountant {
	return &RoundAccountant{
		stage:     initialized,
		ethRate:   ethRate,
		spendable: new(big.Int),
		rounds:    make([]*db.Round, 0),
		balances:  make([]*db.Balance, 0),
		miners:    make(map[uint64]bool),
		fees:      make(map[uint64]*big.Int),
		usedFees:  make(map[uint64]*big.Int),
	}
}

func (a *RoundAccountant) addFees(round uint64, value *big.Int) {
	if _, ok := a.fees[round]; !ok {
		a.fees[round] = new(big.Int)
	}

	a.fees[round].Add(a.fees[round], value)
}

func (a *RoundAccountant) addUsedFees(round uint64, value *big.Int) {
	if _, ok := a.usedFees[round]; !ok {
		a.usedFees[round] = new(big.Int)
	}

	a.usedFees[round].Add(a.usedFees[round], value)
}

func (a *RoundAccountant) GetSpendableAmount() *big.Int {
	return a.spendable
}

func (a *RoundAccountant) GetRounds() ([]*db.Round, error) {
	if a.stage < roundsSet {
		return nil, fmt.Errorf("rounds are not yet set")
	}

	return a.rounds, nil
}

func (a *RoundAccountant) GetBalances() ([]*db.Balance, error) {
	if a.stage != booksValidated {
		return nil, fmt.Errorf("books are not yet validated")
	}

	return a.balances, nil
}

func (a *RoundAccountant) GetPayouts() ([]*db.Payout, error) {
	if a.stage != booksValidated {
		return nil, fmt.Errorf("books are not yet validated")
	}

	return a.payouts, nil
}

func (a *RoundAccountant) AddRound(round *db.Round) error {
	if a.stage != initialized {
		return fmt.Errorf("rounds are already finalized")
	}

	if !round.Value.Valid {
		return fmt.Errorf("invalid value for round %d", round.ID)
	}

	a.spendable.Add(a.spendable, round.Value.BigInt)
	if round.MEVValue.Valid {
		a.spendable.Add(a.spendable, round.MEVValue.BigInt)
	}

	a.rounds = append(a.rounds, round)

	return nil
}

func (a *RoundAccountant) FinalizeSetRounds() error {
	if a.stage != initialized {
		return fmt.Errorf("stage is not initialized")
	}

	a.stage = roundsSet

	return nil
}

func (a *RoundAccountant) FinalizeCreditRounds() error {
	if a.stage != roundsSet {
		return fmt.Errorf("stage is not rounds set")
	}

	totalValue := new(big.Int)
	for _, balance := range a.balances {
		totalValue.Add(totalValue, balance.InValue.BigInt)
	}

	if totalValue.Cmp(a.spendable) != 0 {
		return fmt.Errorf("credit round mismatch: have %s spendable, used %s",
			a.spendable, totalValue)
	}

	a.stage = roundsCredited

	return nil
}

func (a *RoundAccountant) FinalizeCreditMiners() error {
	if a.stage != roundsCredited {
		return fmt.Errorf("stage is not rounds credited")
	}

	totalValue := new(big.Int)
	for _, balance := range a.balances {
		totalValue.Add(totalValue, balance.InValue.BigInt)
	}

	for _, fees := range a.fees {
		totalValue.Add(totalValue, fees)
	}

	if totalValue.Cmp(a.spendable) != 0 {
		return fmt.Errorf("credit miner mismatch: have %s spendable, used %s",
			a.spendable, totalValue)
	}

	a.stage = minersCredited

	return nil
}

func (a *RoundAccountant) FinalizeCreditRecipients() error {
	if a.stage != minersCredited {
		return fmt.Errorf("stage is not miners credited")
	}

	// credit remainders to recipients
	for roundID, value := range a.fees {
		usedValue, ok := a.usedFees[roundID]
		if !ok {
			return fmt.Errorf("unable to find usedFees for round %d", roundID)
		} else if value.Cmp(usedValue) < 0 {
			return fmt.Errorf("used more fees than were available: have %s, want %s",
				value, usedValue)
		}

		remainder := new(big.Int).Sub(value, usedValue)
		for _, balance := range a.balances {
			if uint64(balance.InRoundID.Int64) == roundID && balance.RecipientID.Valid {
				// @TODO: InValue and OutValue share the same pointer so this goes to both
				balance.InValue.BigInt.Add(balance.InValue.BigInt, remainder)

				if balance.InCoin == balance.OutCoin {
					for _, payout := range a.payouts {
						if payout.ID == balance.OutPayoutID {
							if !payout.Value.Valid {
								payout.Value = db.NullBigInt{Valid: true, BigInt: new(big.Int)}
							}

							payout.Value.BigInt.Add(payout.Value.BigInt, remainder)
							break
						}
					}
				}

				a.addUsedFees(roundID, remainder)
				break
			}
		}
	}

	totalValue := new(big.Int)
	for _, balance := range a.balances {
		totalValue.Add(totalValue, balance.InValue.BigInt)
	}

	if totalValue.Cmp(a.spendable) != 0 {
		return fmt.Errorf("credit recipient mismatch: have %s spendable, used %s",
			a.spendable, totalValue)
	}

	a.stage = recipientsCredited

	return nil
}

func (a *RoundAccountant) CreditRound(round *db.Round, shares []*db.Share) error {
	if a.stage != roundsSet {
		return fmt.Errorf("rounds are already credited")
	}

	totalShares := new(big.Int)
	for _, share := range shares {
		totalShares.Add(totalShares, new(big.Int).SetUint64(share.Count))
	}

	totalValue := new(big.Int).Set(round.Value.BigInt)
	if round.MEVValue.Valid {
		totalValue.Add(totalValue, round.MEVValue.BigInt)
	}

	usedValue := new(big.Int)

	for _, share := range shares {
		if share.RoundID != round.ID {
			return fmt.Errorf("share round ID %d and round ID %d mismatch",
				share.RoundID, round.ID)
		}

		numShares := new(big.Int).SetUint64(share.Count)
		minerValue := new(big.Int).Mul(totalValue, numShares)
		minerValue.Div(minerValue, totalShares)
		usedValue.Add(usedValue, minerValue)

		balance := &db.Balance{
			MinerID:   sql.NullInt64{int64(share.MinerID), true},
			InCoin:    round.CoinID,
			InValue:   db.NullBigInt{minerValue, true},
			InRoundID: sql.NullInt64{int64(round.ID), true},
		}

		if !a.miners[share.MinerID] {
			a.miners[share.MinerID] = true
		}

		a.balances = append(a.balances, balance)
	}

	remainder := new(big.Int).Sub(totalValue, usedValue)
	for _, balance := range a.balances {
		if uint64(balance.InRoundID.Int64) == round.ID && balance.MinerID.Valid {
			balance.InValue.BigInt.Add(balance.InValue.BigInt, remainder)
			break
		}
	}

	return nil
}

func (a *RoundAccountant) GetMinerIDs() []uint64 {
	ids := make([]uint64, len(a.miners))

	i := 0
	for miner := range a.miners {
		ids[i] = miner
		i++
	}

	return ids
}

func (a *RoundAccountant) payoutNeedsFeeBalance(payout *db.Payout) *big.Int {
	if payout.CoinID != "USDC" {
		return new(big.Int)
	} else if payout.FeeBalancePending {
		return new(big.Int)
	} else if payout.InFeeBalance.BigInt.Cmp(usdcFeeBalance) >= 0 {
		return new(big.Int)
	}

	var units float64
	if payout.CoinID == "ETH" {
		return usdcFeeBalance
	} else if payout.CoinID == "RVN" {
		units = 1e8
	} else {
		units = 1e18
	}

	// want 0.02 ETH, add some extra for fees
	rawQuantity := 0.025 / a.ethRate
	adjustedQuantity := uint64(rawQuantity * units)
	// @TODO: could overflow if adjustedQuantity is > 18.44
	// and units are 1e18, should do this another way
	finalQuantity := new(big.Int).SetUint64(adjustedQuantity)

	return finalQuantity
}

func (a *RoundAccountant) CreditMiner(miner *db.Miner, payout *db.Payout) error {
	if a.stage != roundsCredited {
		return fmt.Errorf("miners are already credited")
	}

	feeBalanceQty := a.payoutNeedsFeeBalance(payout)

	for _, balance := range a.balances {
		if !balance.MinerID.Valid {
			continue
		} else if uint64(balance.MinerID.Int64) != miner.ID {
			continue
		} else if !balance.InRoundID.Valid {
			continue
		} else if !balance.InValue.Valid {
			continue
		}

		pending := (miner.CoinID != balance.InCoin)
		inValue := balance.InValue.BigInt

		// give Telnyx 0.45% fee, everyone else a 1% fee
		var fee *big.Int
		if miner.Address == "0x7dD8E752F5e606Aca3A40DD1dEaDF363dbbCa100" {
			fee = splitBig(inValue, int64(45))
		} else {
			fee = splitBig(inValue, int64(100))
		}

		inValue.Sub(inValue, fee)

		var outValue, exchangeFees db.NullBigInt
		if !pending {
			exchangeFees = db.NullBigInt{new(big.Int), true}
			outValue = db.NullBigInt{inValue, true}
		}

		balance.PoolFees = db.NullBigInt{fee, true}
		balance.ExchangeFees = exchangeFees
		balance.InValue = db.NullBigInt{inValue, true}
		balance.Pending = pending
		balance.OutCoin = miner.CoinID
		balance.OutType = int(types.StandardBalance)
		balance.OutValue = outValue
		balance.OutPayoutID = payout.ID

		if feeBalanceQty.Cmp(new(big.Int)) > 0 {
			feeBalanceQty = a.addFeeBalance(payout, balance, feeBalanceQty)
		}

		a.addFees(uint64(balance.InRoundID.Int64), balance.PoolFees.BigInt)

		if !pending {
			payout.Value.BigInt.Add(payout.Value.BigInt, balance.OutValue.BigInt)
			payout.PoolFees.BigInt.Add(payout.PoolFees.BigInt, balance.PoolFees.BigInt)
		}
	}

	a.payouts = append(a.payouts, payout)

	return nil
}

func (a *RoundAccountant) CreditRecipient(recipient *db.Recipient, payout *db.Payout) error {
	if a.stage != minersCredited {
		return fmt.Errorf("recipients are already credited")
	}

	feeBalanceQty := a.payoutNeedsFeeBalance(payout)

	for _, round := range a.rounds {
		fees := a.fees[round.ID]
		if fees == nil || fees.Cmp(big.NewInt(0)) != 1 {
			continue
		}

		recipientValue := splitBig(fees, int64(recipient.Fraction)*100)
		a.addUsedFees(round.ID, recipientValue)

		pending := (recipient.CoinID != round.CoinID)
		outValue := db.NullBigInt{}
		if !pending {
			outValue = db.NullBigInt{recipientValue, true}
		}

		balance := &db.Balance{
			RecipientID: sql.NullInt64{int64(recipient.ID), true},
			InCoin:      round.CoinID,
			InValue:     db.NullBigInt{recipientValue, true},
			InRoundID:   sql.NullInt64{int64(round.ID), true},

			Pending: pending,

			PoolFees:     db.NullBigInt{new(big.Int), true},
			ExchangeFees: db.NullBigInt{new(big.Int), true},

			OutCoin:     recipient.CoinID,
			OutType:     int(types.StandardBalance),
			OutValue:    outValue,
			OutPayoutID: payout.ID,
		}

		a.balances = append(a.balances, balance)

		if feeBalanceQty.Cmp(new(big.Int)) > 0 {
			feeBalanceQty = a.addFeeBalance(payout, balance, feeBalanceQty)
		}

		if !pending {
			payout.Value.BigInt.Add(payout.Value.BigInt, balance.OutValue.BigInt)
			payout.PoolFees.BigInt.Add(payout.PoolFees.BigInt, balance.PoolFees.BigInt)
		}
	}

	a.payouts = append(a.payouts, payout)

	return nil
}

func (a *RoundAccountant) addFeeBalance(payout *db.Payout, balance *db.Balance, qty *big.Int) *big.Int {
	remainder := new(big.Int).Sub(balance.InValue.BigInt, qty)

	var feeValue, poolFees *big.Int
	if remainder.Cmp(big.NewInt(0)) != 1 {
		feeValue = balance.InValue.BigInt
		poolFees = balance.PoolFees.BigInt
		// set the balance/fee value to zero
		balance.InValue = db.NullBigInt{new(big.Int), true}
		balance.PoolFees = db.NullBigInt{new(big.Int), true}
	} else {
		feeValue = qty
		poolFees = new(big.Int).Mul(balance.PoolFees.BigInt, qty)
		poolFees.Div(poolFees, balance.InValue.BigInt)
		// adjust balance InValue to account for fee balance
		balance.InValue = db.NullBigInt{remainder, true}
		balance.PoolFees.BigInt = new(big.Int).Sub(balance.PoolFees.BigInt, poolFees)
		if balance.OutValue.Valid {
			balance.OutValue = balance.InValue
		}
	}

	pending := balance.InCoin != "ETH"
	outValue := db.NullBigInt{}
	if !pending {
		outValue = db.NullBigInt{feeValue, true}
	}

	feeBalance := &db.Balance{
		MinerID:     balance.MinerID,
		RecipientID: balance.RecipientID,

		InCoin:    balance.InCoin,
		InValue:   db.NullBigInt{feeValue, true},
		InRoundID: balance.InRoundID,

		Pending: pending,

		PoolFees:     db.NullBigInt{poolFees, true},
		ExchangeFees: db.NullBigInt{new(big.Int), true},

		OutCoin:     "ETH",
		OutType:     int(types.FeeBalance),
		OutValue:    outValue,
		OutPayoutID: balance.OutPayoutID,
	}

	a.addFees(uint64(balance.InRoundID.Int64), poolFees)
	a.balances = append(a.balances, feeBalance)

	if !pending {
		payout.InFeeBalance.BigInt.Add(payout.InFeeBalance.BigInt, feeValue)

		// normalize USDC pool fees to be in USDC, not in ETH
		bigEthRate := new(big.Float).SetFloat64(a.ethRate)
		bigPoolFees := new(big.Float).SetInt(feeBalance.PoolFees.BigInt)
		bigPoolFees.Mul(bigPoolFees, bigEthRate).Quo(bigPoolFees, new(big.Float).SetFloat64(1e12))

		usdcPoolFees, _ := bigPoolFees.Int(new(big.Int))
		payout.PoolFees.BigInt.Add(payout.PoolFees.BigInt, usdcPoolFees)
	}

	return new(big.Int).Sub(qty, feeValue)
}

// validate that the spendable amount is equal to the amount used
// and that the sum recipient value is identical to the sum miner pool fees
func (a *RoundAccountant) ValidateBooks() error {
	if a.stage != recipientsCredited {
		return fmt.Errorf("books are already validated")
	}

	totalValue := new(big.Int)
	minerFeeValue := new(big.Int)
	poolFeeValue := new(big.Int)

	for _, balance := range a.balances {
		totalValue.Add(totalValue, balance.InValue.BigInt)

		if balance.MinerID.Valid {
			minerFeeValue.Add(minerFeeValue, balance.PoolFees.BigInt)
		} else {
			poolFeeValue.Add(poolFeeValue, balance.InValue.BigInt)
		}
	}

	for roundID, value := range a.fees {
		usedValue, ok := a.usedFees[roundID]
		if !ok {
			return fmt.Errorf("unable to find usedFees for round %d", roundID)
		}

		if usedValue.Cmp(value) != 0 {
			return fmt.Errorf("fee mismatch on round %d: have %s, used %s",
				roundID, value, usedValue)
		}
	}

	if minerFeeValue.Cmp(poolFeeValue) != 0 {
		return fmt.Errorf("miner and pool fees mismatch: have %s and %s",
			minerFeeValue, poolFeeValue)
	}

	if a.spendable.Cmp(totalValue) != 0 {
		return fmt.Errorf("spendable and used mismatch: have %s, want %s",
			totalValue, a.spendable)
	}

	for _, round := range a.rounds {
		expectedRoundValue := new(big.Int).Set(round.Value.BigInt)
		if round.MEVValue.Valid {
			expectedRoundValue.Add(expectedRoundValue, round.MEVValue.BigInt)
		}

		actualRoundValue := new(big.Int)
		for _, balance := range a.balances {
			if uint64(balance.InRoundID.Int64) == round.ID {
				actualRoundValue.Add(actualRoundValue, balance.InValue.BigInt)
			}
		}

		if expectedRoundValue.Cmp(actualRoundValue) != 0 {
			return fmt.Errorf("actual round %d value and expected mismatch: have %s, want %s",
				round.ID, actualRoundValue, expectedRoundValue)
		}
	}

	a.stage = booksValidated

	return nil
}

package credit

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/core/trade/kucoin"
	"github.com/magicpool-co/pool/internal/accounting"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func ProcessUSDCFeeBalance(pooldbClient *dbcl.Client, miner *pooldb.Miner, round *pooldb.Round, value, poolFee *big.Int) (*big.Int, *big.Int, *pooldb.BalanceInput, error) {
	var minFeeBalance = new(big.Int).SetUint64(10_000_000_000_000_000)

	if miner.ChainID != "USDC" {
		return value, poolFee, nil, nil
	}

	// fetch the miner's current ETH balance and calculate the necessary needed fee balance
	ethBalance, err := pooldb.GetSumBalanceOutputValueByMiner(pooldbClient.Reader(), miner.ID, "ETH")
	if err != nil {
		return nil, nil, nil, err
	} else if ethBalance.Cmp(minFeeBalance) >= 0 {
		return value, poolFee, nil, nil
	}
	neededFeeBalance := new(big.Int).Sub(minFeeBalance, ethBalance)

	usdcPath := map[string]map[string]*big.Int{
		round.ChainID: map[string]*big.Int{
			miner.ChainID: value,
		},
	}
	prices, err := kucoin.New("", "", "").GetPrices(usdcPath)
	if err != nil {
		return nil, nil, nil, err
	} else if _, ok := prices[round.ChainID]; !ok {
		return nil, nil, nil, fmt.Errorf("no prices found for %s", round.ChainID)
	} else if _, ok := prices[round.ChainID][miner.ChainID]; !ok {
		return nil, nil, nil, fmt.Errorf("no prices found for %s->%s", round.ChainID, miner.ChainID)
	}

	initialUnitsBig, err := common.GetDefaultUnits(round.ChainID)
	if err != nil {
		return nil, nil, nil, err
	}

	finalUnitsBig, err := common.GetDefaultUnits(miner.ChainID)
	if err != nil {
		return nil, nil, nil, err
	}

	// set all of the required values as big floats for the adjustment calculation
	valueFloat := new(big.Float).SetInt(value)
	priceFloat := new(big.Float).SetFloat64(prices[round.ChainID][miner.ChainID])
	initialUnitsFloat := new(big.Float).SetInt(initialUnitsBig)
	finalUnitsFloat := new(big.Float).SetInt(finalUnitsBig)

	// calculate the adjusted USDC proceeds
	estimatedValueFloat := new(big.Float).Quo(valueFloat, initialUnitsFloat)
	estimatedValueFloat.Mul(estimatedValueFloat, priceFloat)
	estimatedValueFloat.Mul(estimatedValueFloat, finalUnitsFloat)
	estimatedValue, _ := estimatedValueFloat.Int(nil)

	// if the estimated value is less than or equal to the remaining
	// needed fee balance, only create the fee balance input
	if neededFeeBalance.Cmp(estimatedValue) <= 0 {
		feeBalanceInput := &pooldb.BalanceInput{
			RoundID: round.ID,
			ChainID: "ETH",
			MinerID: miner.ID,

			Value:    dbcl.NullBigInt{Valid: true, BigInt: value},
			PoolFees: dbcl.NullBigInt{Valid: true, BigInt: poolFee},
			Pending:  true,
		}

		return new(big.Int), new(big.Int), feeBalanceInput, nil
	}

	// value * (neededFeeBalance / estimatedValue) is the formula
	// for calculating the proportional value that goes to USDC
	proportionalValue := new(big.Int).Mul(value, neededFeeBalance)
	proportionalValue.Div(proportionalValue, estimatedValue)

	// poolFee * (neededFeeBalance / estimatedValue) is the formula
	// for calculating the proportional pool fees that go to USDC
	proportionalFee := new(big.Int).Mul(poolFee, neededFeeBalance)
	proportionalFee.Div(proportionalFee, estimatedValue)

	// verify that value is non-negative
	value.Sub(value, proportionalValue)
	if value.Cmp(common.Big0) < 0 {
		return nil, nil, nil, fmt.Errorf("negative remainder value for usdc proportions")
	}

	// verify that fee is non-negative
	poolFee.Sub(poolFee, proportionalFee)
	if poolFee.Cmp(common.Big0) < 0 {
		return nil, nil, nil, fmt.Errorf("negative remainder fee for usdc proportions")
	}

	// create the final fee balance input, update value and pool fees to the remaining
	// values after the fee balance input has been taken out
	feeBalanceInput := &pooldb.BalanceInput{
		RoundID: round.ID,
		ChainID: "ETH",
		MinerID: miner.ID,

		Value:    dbcl.NullBigInt{Valid: true, BigInt: proportionalValue},
		PoolFees: dbcl.NullBigInt{Valid: true, BigInt: proportionalFee},
		Pending:  true,
	}

	return value, poolFee, feeBalanceInput, nil
}

func CreditRound(pooldbClient *dbcl.Client, round *pooldb.Round, shares []*pooldb.Share) error {
	// create a miner index of proportional share values
	minerIdx := make(map[uint64]uint64)
	for _, share := range shares {
		minerIdx[share.MinerID] += share.Count
	}

	// fetch the recipients and create a recipient index of proportional fee values
	recipients, err := pooldb.GetRecipients(pooldbClient.Reader())
	if err != nil {
		return err
	}

	recipientIdx := make(map[uint64]uint64)
	for _, recipient := range recipients {
		if recipient.RecipientFeePercent == nil {
			return fmt.Errorf("no recipient fee set for %d", recipient.ID)
		}
		recipientIdx[recipient.ID] += types.Uint64Value(recipient.RecipientFeePercent)
	}

	// distribute the proceeds to miners and recipients
	compoundValues, minerFees, err := accounting.CreditRound(round.Value.BigInt, minerIdx, recipientIdx)
	if err != nil {
		return err
	}

	// fetch miners and recipients to check their payout chain
	compoundIDs := make([]uint64, 0)
	for compoundID := range compoundValues {
		compoundIDs = append(compoundIDs, compoundID)
	}

	miners, err := pooldb.GetMiners(pooldbClient.Reader(), compoundIDs)
	if err != nil {
		return err
	}

	// create compound index for miners and recipients
	compoundIdx := make(map[uint64]*pooldb.Miner)
	for _, miner := range miners {
		compoundIdx[miner.ID] = miner
	}

	// process the balance inputs for all round distributions
	usedValue := new(big.Int)
	inputs := make([]*pooldb.BalanceInput, 0)
	for minerID, value := range compoundValues {
		miner, ok := compoundIdx[minerID]
		if !ok {
			return fmt.Errorf("no miner found for %d", minerID)
		} else if value == nil || value.Cmp(common.Big0) == 0 {
			continue
		}
		usedValue.Add(usedValue, value)

		poolFee, ok := minerFees[minerID]
		if !ok {
			poolFee = new(big.Int)
		}

		// calculate the USDC fee balance that needs to be switched,
		// based off of the miner's current fee balance
		value, poolFee, feeBalanceInput, err := ProcessUSDCFeeBalance(pooldbClient, miner, round, value, poolFee)
		if err != nil {
			return err
		}

		// add the balance input only of the new value is positive and non-zero
		if value.Cmp(common.Big0) > 0 {
			balanceInput := &pooldb.BalanceInput{
				RoundID: round.ID,
				ChainID: round.ChainID,
				MinerID: miner.ID,

				Value:    dbcl.NullBigInt{Valid: true, BigInt: value},
				PoolFees: dbcl.NullBigInt{Valid: true, BigInt: poolFee},
				Pending:  round.ChainID != miner.ChainID,
			}
			inputs = append(inputs, balanceInput)
		}

		// add the fee balance input if exists
		if feeBalanceInput != nil {
			inputs = append(inputs, feeBalanceInput)
		}

		delete(compoundIdx, miner.ID)
		delete(compoundValues, miner.ID)
	}

	if len(compoundIdx) > 0 {
		return fmt.Errorf("unable to find %d miners in idx", len(compoundIdx))
	} else if len(compoundValues) > 0 {
		return fmt.Errorf("unable to find %d miners in values", len(compoundValues))
	} else if usedValue.Cmp(round.Value.BigInt) != 0 {
		return fmt.Errorf("crediting mismatch: have %s, want %s", usedValue, round.Value.BigInt)
	}

	tx, err := pooldbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.SafeRollback()

	round.Spent = true
	if err := pooldb.InsertBalanceInputs(tx, inputs...); err != nil {
		return err
	} else if pooldb.UpdateRound(tx, round, []string{"spent"}); err != nil {
		return err
	}

	return tx.SafeCommit()
}

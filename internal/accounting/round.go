package accounting

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
)

// splits a value according to the proportional total for each value in the index,
// along with calculating the remainder and verifying it is positive
func splitValue(
	value *big.Int,
	idx map[uint64]uint64,
) (map[uint64]*big.Int, *big.Int, error) {
	if value.Cmp(common.Big0) == 0 {
		return nil, new(big.Int), nil
	}

	// calculate the sum of the index counts
	var totalCount uint64
	for _, count := range idx {
		totalCount += count
	}

	// split the value as a percentage of each individual count against
	// the total account, add used quantity to account for remainder
	used := new(big.Int)
	values := make(map[uint64]*big.Int)
	for id, count := range idx {
		values[id] = common.SplitBigPercentage(value, count, totalCount)
		used.Add(used, values[id])
	}

	// calculate the remainder and verify it is non-negative
	remainder := new(big.Int).Sub(value, used)
	if remainder.Cmp(common.Big0) < 0 {
		return nil, nil, fmt.Errorf("negative remainder")
	}

	return values, remainder, nil
}

// credits a round based off of the share index and the fee recipient distributions. the output is a merged
// map of miners and recipients since we can safely assume that miner ids and recipient ids are globally unique.
func CreditRound(
	roundValue *big.Int,
	minerIdx, recipientIdx map[uint64]uint64,
) (map[uint64]*big.Int, map[uint64]*big.Int, error) {
	if roundValue == nil {
		return nil, nil, fmt.Errorf("empty round value")
	} else if len(minerIdx) == 0 {
		return nil, nil, fmt.Errorf("empty miner index")
	} else if len(recipientIdx) == 0 {
		return nil, nil, fmt.Errorf("empty recipient index")
	}

	// copy value to avoid overwriting it elsewhere
	roundValue = new(big.Int).Set(roundValue)

	// takes a 0.01% fee from the value as the pool fees
	feeValue := common.SplitBigPercentage(roundValue, 1, 10000)
	adjustedRoundValue := new(big.Int).Sub(roundValue, feeValue)

	// calculate the miner distributions and remainder
	minerValues, remainder, err := splitValue(adjustedRoundValue, minerIdx)
	if err != nil {
		return nil, nil, err
	}
	feeValue.Add(feeValue, remainder)

	// calculate the miner fees by recalculating the distributions without
	// the fee and subtracting each miner's distribution with the fee
	initialMinerValues, _, err := splitValue(roundValue, minerIdx)
	if err != nil {
		return nil, nil, err
	}

	minerFees := make(map[uint64]*big.Int)
	for minerID, initialValue := range initialMinerValues {
		minerFees[minerID] = new(big.Int).Sub(initialValue, minerValues[minerID])
	}

	// calculate the fee recipients distributions and remainder
	recipientValues, remainder, err := splitValue(feeValue, recipientIdx)
	if err != nil {
		return nil, nil, err
	}

	// add the remainder to the fee recipient recieving the lowest quantity
	// (if the quantity is the same, give it to the lowest recipient ID).
	// this is required since recipientIdx is a map, meaning there is no
	// deterministic first item like in a slice.
	var lowestValue *big.Int
	var lowestRecipientID uint64
	for id, value := range recipientValues {
		if lowestValue == nil {
			lowestRecipientID = id
			lowestValue = value
		} else {
			switch lowestValue.Cmp(value) {
			case 0:
				if lowestRecipientID == 0 || lowestRecipientID > id {
					lowestRecipientID = id
					lowestValue = value
				}
			case 1:
				lowestValue = value
			}
		}
	}
	lowestValue.Add(lowestValue, remainder)

	compoundValues := minerValues
	for recipientID, value := range recipientValues {
		if _, ok := compoundValues[recipientID]; ok {
			compoundValues[recipientID].Add(compoundValues[recipientID], value)
		} else {
			compoundValues[recipientID] = value
		}
	}

	return compoundValues, minerFees, nil
}

// given a miner's round distributions, check to see if any fee balance is needed and, if so,
// return the estimated needed fee balance
func ProcessFeeBalance(
	roundChain, minerChain string,
	value, poolFee, feeBalance *big.Int,
	price float64,
) (*big.Int, *big.Int, error) {
	// specify the minumum fee balance required for the given chain,
	// throw an error if the chain doesn't exist
	var minFeeBalance *big.Int
	switch minerChain {
	case "USDC":
		minFeeBalance = new(big.Int).SetUint64(10_000_000_000_000_000)
	default:
		return nil, nil, fmt.Errorf("unsupported fee balance chain")
	}

	// if the mienr already has enough fee balance, return with empty values
	if feeBalance.Cmp(minFeeBalance) >= 0 {
		return new(big.Int), new(big.Int), nil
	}
	neededFeeBalance := new(big.Int).Sub(minFeeBalance, feeBalance)

	// fetch the units for both chains
	initialUnitsBig, err := common.GetDefaultUnits(roundChain)
	if err != nil {
		return nil, nil, err
	}

	finalUnitsBig, err := common.GetDefaultUnits("ETH")
	if err != nil {
		return nil, nil, err
	}

	// set all of the required values as big floats for the adjustment calculation
	valueFloat := new(big.Float).SetInt(value)
	priceFloat := new(big.Float).SetFloat64(price)
	initialUnitsFloat := new(big.Float).SetInt(initialUnitsBig)
	finalUnitsFloat := new(big.Float).SetInt(finalUnitsBig)

	// calculate the adjusted ETH proceeds
	estimatedValueFloat := new(big.Float).Quo(valueFloat, initialUnitsFloat)
	estimatedValueFloat.Mul(estimatedValueFloat, priceFloat)
	estimatedValueFloat.Mul(estimatedValueFloat, finalUnitsFloat)
	estimatedValue, _ := estimatedValueFloat.Int(nil)

	// if the estimated value is less than or equal to the remaining
	// needed fee balance, only create the fee balance input
	if neededFeeBalance.Cmp(estimatedValue) > 0 {
		feeBalanceValue := new(big.Int).Set(value)
		feeBalancePoolFee := new(big.Int).Set(poolFee)

		return feeBalanceValue, feeBalancePoolFee, nil
	}

	// protect against divide by zero
	if estimatedValue.Cmp(common.Big0) <= 0 {
		return nil, nil, fmt.Errorf("empty estimated value")
	}

	// value * (neededFeeBalance / estimatedValue) is the formula
	// for calculating the proportional value that goes to USDC
	proportionalValue := new(big.Int).Mul(value, neededFeeBalance)
	proportionalValue.Div(proportionalValue, estimatedValue)

	// poolFee * (neededFeeBalance / estimatedValue) is the formula
	// for calculating the proportional pool fees that go to USDC
	proportionalFee := new(big.Int).Mul(poolFee, neededFeeBalance)
	proportionalFee.Div(proportionalFee, estimatedValue)

	// verify that the proportional value and pool fee are not greater than the initial value
	if value.Cmp(proportionalValue) < 0 {
		return nil, nil, fmt.Errorf("negative remainder value for usdc proportions")
	} else if poolFee.Cmp(proportionalFee) < 0 {
		return nil, nil, fmt.Errorf("negative remainder fee for usdc proportions")
	}

	return proportionalValue, proportionalFee, nil
}

package accounting

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
)

// splits a value according to the proportional total for each value in the index,
// along with calculating the remainder and verifying it is positive
func splitValue(value *big.Int, idx map[uint64]uint64) (map[uint64]*big.Int, *big.Int, error) {
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

func CreditRound(roundValue *big.Int, minerIdx, recipientIdx map[uint64]uint64) (map[uint64]*big.Int, map[uint64]*big.Int, error) {
	if roundValue == nil {
		return nil, nil, fmt.Errorf("empty round value")
	} else if len(minerIdx) == 0 {
		return nil, nil, fmt.Errorf("empty miner index")
	} else if len(recipientIdx) == 0 {
		return nil, nil, fmt.Errorf("empty recipient index")
	}

	// copy value to avoid overwriting it elsewhere
	roundValue = new(big.Int).Set(roundValue)

	// takes a 1% fee from the value as the pool fees
	feeValue := common.SplitBigPercentage(roundValue, 100, 10000)
	roundValue.Sub(roundValue, feeValue)

	// calculate the miner distributions and remainder
	minerValues, remainder, err := splitValue(roundValue, minerIdx)
	if err != nil {
		return nil, nil, err
	}
	feeValue.Add(feeValue, remainder)

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

	return minerValues, recipientValues, nil
}

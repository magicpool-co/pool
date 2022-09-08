package ergo

import (
	"math/big"
)

const (
	initialEndHeight     = 655200
	reductionStartHeight = 719200
	reductionInterval    = 64800
)

var (
	initialReward        = new(big.Int).SetUint64(67_500_000_000)
	reductionStartReward = new(big.Int).SetUint64(66_000_000_000)
)

func getBlockReward(height uint64) *big.Int {
	if height < initialEndHeight {
		return initialReward
	} else if height < reductionStartHeight {
		return reductionStartReward
	}

	reduction := 3e9 * (((height - reductionStartHeight) / reductionInterval) + 1)
	if height >= 777217 {
		reduction += 12e9
	}

	reductionBig := new(big.Int).SetUint64(reduction)
	if reductionBig.Cmp(reductionStartReward) != -1 {
		return new(big.Int)
	}

	blockReward := new(big.Int).Sub(reductionStartReward, reductionBig)

	return blockReward
}

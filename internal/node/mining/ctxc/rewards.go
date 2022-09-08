package ctxc

import (
	"math/big"
)

var (
	// magic numbers
	big2  = new(big.Int).SetUint64(2)
	big8  = new(big.Int).SetUint64(8)
	big32 = new(big.Int).SetUint64(32)

	blockRewardPeriod = new(big.Int).SetUint64(8_409_600)
	frontierReward    = new(big.Int).SetUint64(7_000_000_000_000_000_000)
)

func getBlockReward(height, uncleCount uint64) *big.Int {
	baseReward := new(big.Int).Set(frontierReward)
	d := new(big.Int).Div(new(big.Int).SetUint64(height), blockRewardPeriod)
	e := new(big.Int).Exp(big2, d, nil)
	blockReward := new(big.Int).Div(baseReward, e)

	if uncleCount != 0 {
		uncleInclusionReward := new(big.Int).Div(baseReward, big32)
		uncleReward := new(big.Int).Mul(uncleInclusionReward, new(big.Int).SetUint64(uncleCount))
		blockReward.Add(blockReward, uncleReward)
	}

	return blockReward
}

func getUncleReward(height, uncleHeight uint64) *big.Int {
	baseReward := getBlockReward(height, 0)
	k := height - uncleHeight
	uncleReward := new(big.Int).Mul(baseReward, new(big.Int).SetUint64(8-k))
	uncleReward.Div(uncleReward, big8)

	return uncleReward
}

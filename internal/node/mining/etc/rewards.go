package etc

import (
	"math/big"
)

var (
	// magic numbers
	big0  = new(big.Int)
	big1  = new(big.Int).SetUint64(1)
	big4  = new(big.Int).SetUint64(4)
	big5  = new(big.Int).SetUint64(5)
	big32 = new(big.Int).SetUint64(32)

	homesteadReward   = new(big.Int).SetUint64(5_000_000_000_000_000_000)
	ecip1017EraLength = new(big.Int).SetUint64(5_000_000)
)

func getBlockEra(height uint64) *big.Int {
	blockNum := new(big.Int).SetUint64(height)
	if blockNum.Cmp(big0) < 0 {
		return new(big.Int)
	}

	remainder := new(big.Int).Mod(new(big.Int).Sub(blockNum, big1), ecip1017EraLength)
	base := new(big.Int).Sub(blockNum, remainder)
	d := new(big.Int).Div(base, ecip1017EraLength)
	dRemainder := new(big.Int).Mod(d, big1)
	era := new(big.Int).Sub(d, dRemainder)

	return era
}

func getBlockReward(height, uncleCount uint64) *big.Int {
	era := getBlockEra(height)
	blockReward := new(big.Int)
	blockReward.Mul(homesteadReward, new(big.Int).Exp(big4, era, nil))
	blockReward.Div(blockReward, new(big.Int).Exp(big5, era, nil))

	if uncleCount != 0 {
		uncleInclusionReward := new(big.Int).Div(blockReward, big32)
		uncleReward := new(big.Int).Mul(uncleInclusionReward, new(big.Int).SetUint64(uncleCount))
		blockReward.Add(blockReward, uncleReward)
	}

	return blockReward
}

func getUncleReward(height uint64) *big.Int {
	baseReward := getBlockReward(height, 0)
	uncleReward := new(big.Int).Div(baseReward, big32)

	return uncleReward
}

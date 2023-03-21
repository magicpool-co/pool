package erg

import (
	"math/big"
)

const (
	initialEndHeight     = 655200
	initialReward        = 67_500_000_000
	reductionStartHeight = 719200
	reductionStartReward = 66_000_000_000
	softForkHeight       = 777217

	epochLength     = 64800
	fixedRatePeriod = 525600
	fixedRate       = 75_000_000_000
	reductionRate   = 3_000_000_000
)

func emissionRateFromEpoch(epoch uint64) uint64 {
	rate := fixedRate - reductionRate*epoch
	if rate < 0 {
		rate = 0
	}

	return rate
}

func reserveRate(emissionRate uint64) uint64 {
	if emissionRate >= 15_000_000_000 {
		return 12_000_000_000
	} else if emissionRate > reductionRate {
		return emissionRate - reductionRate
	}

	return 0
}

func getBlockReward(height uint64) *big.Int {
	if height < initialEndHeight {
		return new(big.Int).SetUint64(initialReward)
	} else if height < reductionStartHeight {
		return new(big.Int).SetUint64(reductionStartReward)
	}

	var epoch uint64
	if height >= fixedRatePeriod {
		epoch = ((height - fixedRatePeriod) / epochLength) + 1
	}

	rate := fixedRate - epoch*reductionRate
	if rate < 0 {
		rate = 0
	} else if height >= softForkHeight {
		rate -= reserveRate(emissionRateFromEpoch(epoch))
	}

	return new(big.Int).SetUint64(rate)
}

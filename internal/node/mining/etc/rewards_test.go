package etc

import (
	"math/big"
	"testing"
)

func TestGetBlockReward(t *testing.T) {
	tests := []struct {
		height     uint64
		uncleCount uint64
		reward     *big.Int
	}{
		{
			height:     13676163,
			uncleCount: 0,
			reward:     new(big.Int).SetUint64(3_200_000_000_000_000_000),
		},
		{
			height:     15518349,
			uncleCount: 0,
			reward:     new(big.Int).SetUint64(2_560_000_000_000_000_000),
		},
		{
			height:     15518350,
			uncleCount: 1,
			reward:     new(big.Int).SetUint64(2_640_000_000_000_000_000),
		},
	}

	for i, tt := range tests {
		reward := getBlockReward(tt.height, tt.uncleCount)
		if reward.Cmp(tt.reward) != 0 {
			t.Errorf("failed on %d: have %s, want %s", i, reward, tt.reward)
		}
	}
}

func TestGetUncleReward(t *testing.T) {
	tests := []struct {
		blockHeight uint64
		uncleHeight uint64
		reward      *big.Int
	}{
		{
			blockHeight: 14075679,
			uncleHeight: 14075677,
			reward:      new(big.Int).SetUint64(100_000_000_000_000_000),
		},
		{
			blockHeight: 14075898,
			uncleHeight: 14075897,
			reward:      new(big.Int).SetUint64(100_000_000_000_000_000),
		},
		{
			blockHeight: 15518350,
			uncleHeight: 15518349,
			reward:      new(big.Int).SetUint64(80_000_000_000_000_000),
		},
	}

	for i, tt := range tests {
		reward := getUncleReward(tt.blockHeight)
		if reward.Cmp(tt.reward) != 0 {
			t.Errorf("failed on %d: have %s, want %s", i, reward, tt.reward)
		}
	}
}

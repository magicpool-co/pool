package etc

import (
	"math/big"
	"testing"
)

func TestGetBlockRewardETC(t *testing.T) {
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
		reward := getBlockRewardETC(tt.height, tt.uncleCount)
		if reward.Cmp(tt.reward) != 0 {
			t.Errorf("failed on %d: have %s, want %s", i, reward, tt.reward)
		}
	}
}

func TestGetUncleRewardETC(t *testing.T) {
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
		reward := getUncleRewardETC(tt.blockHeight, tt.uncleHeight)
		if reward.Cmp(tt.reward) != 0 {
			t.Errorf("failed on %d: have %s, want %s", i, reward, tt.reward)
		}
	}
}

func TestGetBlockRewardETHW(t *testing.T) {
	tests := []struct {
		height     uint64
		uncleCount uint64
		reward     *big.Int
	}{
		{
			height:     13367629,
			uncleCount: 0,
			reward:     new(big.Int).SetUint64(2_000_000_000_000_000_000),
		},
		{
			height:     13367629,
			uncleCount: 1,
			reward:     new(big.Int).SetUint64(2_062_500_000_000_000_000),
		},
		{
			height:     13367629,
			uncleCount: 2,
			reward:     new(big.Int).SetUint64(2_125_000_000_000_000_000),
		},
	}

	for i, tt := range tests {
		reward := getBlockRewardETHW(tt.height, tt.uncleCount)
		if reward.Cmp(tt.reward) != 0 {
			t.Errorf("failed on %d: have %s, want %s", i, reward, tt.reward)
		}
	}
}

func TestGetUncleRewardETHW(t *testing.T) {
	tests := []struct {
		blockHeight uint64
		uncleHeight uint64
		reward      *big.Int
	}{
		{
			blockHeight: 14075679,
			uncleHeight: 14075677,
			reward:      new(big.Int).SetUint64(1_500_000_000_000_000_000),
		},
		{
			blockHeight: 14075898,
			uncleHeight: 14075897,
			reward:      new(big.Int).SetUint64(1_750_000_000_000_000_000),
		},
	}

	for i, tt := range tests {
		reward := getUncleRewardETHW(tt.blockHeight, tt.uncleHeight)
		if reward.Cmp(tt.reward) != 0 {
			t.Errorf("failed on %d: have %s, want %s", i, reward, tt.reward)
		}
	}
}

package ctxc

import (
	"math/big"
	"testing"
)

func TestGetBlockReward(t *testing.T) {
	tests := []struct {
		height uint64
		uncles uint64
		reward *big.Int
	}{
		{
			height: 6622227,
			uncles: 0,
			reward: new(big.Int).SetUint64(7_000_000_000_000_000_000),
		},
		{
			height: 6622229,
			uncles: 1,
			reward: new(big.Int).SetUint64(7_218_750_000_000_000_000),
		},
	}

	for i, tt := range tests {
		reward := getBlockReward(tt.height, tt.uncles)
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
			blockHeight: 6621942,
			uncleHeight: 6621941,
			reward:      new(big.Int).SetUint64(6_125_000_000_000_000_000),
		},
		{
			blockHeight: 6622139,
			uncleHeight: 6622137,
			reward:      new(big.Int).SetUint64(5_250_000_000_000_000_000),
		},
	}

	for i, tt := range tests {
		reward := getUncleReward(tt.blockHeight, tt.uncleHeight)
		if reward.Cmp(tt.reward) != 0 {
			t.Errorf("failed on %d: have %s, want %s", i, reward, tt.reward)
		}
	}
}

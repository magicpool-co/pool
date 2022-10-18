package ergo

import (
	"math/big"
	"testing"
)

func TestGetBlockReward(t *testing.T) {
	tests := []struct {
		height uint64
		reward *big.Int
	}{
		{
			height: 1,
			reward: new(big.Int).SetUint64(67_500_000_000),
		},
		{
			height: 481280,
			reward: new(big.Int).SetUint64(67_500_000_000),
		},
		{
			height: 655200,
			reward: new(big.Int).SetUint64(66_000_000_000),
		},
		{
			height: 720000,
			reward: new(big.Int).SetUint64(63_000_000_000),
		},
		{
			height: 784800,
			reward: new(big.Int).SetUint64(48_000_000_000),
		},
		{
			height: 849484,
			reward: new(big.Int).SetUint64(48_000_000_000),
		},
		{
			height: 849600,
			reward: new(big.Int).SetUint64(45_000_000_000),
		},
		{
			height: 914400,
			reward: new(big.Int).SetUint64(42_000_000_000),
		},
		{
			height: 979200,
			reward: new(big.Int).SetUint64(39_000_000_000),
		},
	}

	for i, tt := range tests {
		reward := getBlockReward(tt.height)
		if reward.Cmp(tt.reward) != 0 {
			t.Errorf("failed on %d: have %s, want %s", i, reward, tt.reward)
		}
	}
}

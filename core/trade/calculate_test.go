package trade

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func TestBalanceInputsToInputPaths(t *testing.T) {
	tests := []struct {
		balanceInputs []*pooldb.BalanceInput
		inputPaths    map[string]map[string]*big.Int
	}{
		{
			balanceInputs: []*pooldb.BalanceInput{
				&pooldb.BalanceInput{
					ChainID:    "ETC",
					OutChainID: "ETH",
					Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(32)},
				},
			},
			inputPaths: map[string]map[string]*big.Int{
				"ETC": map[string]*big.Int{
					"ETH": new(big.Int).SetUint64(32),
				},
			},
		},
		{
			balanceInputs: []*pooldb.BalanceInput{
				&pooldb.BalanceInput{
					ChainID:    "ETC",
					OutChainID: "ETH",
					Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(5818283841)},
				},
				&pooldb.BalanceInput{
					ChainID:    "ETC",
					OutChainID: "ETH",
					Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(94124512312)},
				},
				&pooldb.BalanceInput{
					ChainID:    "ETC",
					OutChainID: "BTC",
					Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(5124885812)},
				},
			},
			inputPaths: map[string]map[string]*big.Int{
				"ETC": map[string]*big.Int{
					"ETH": new(big.Int).SetUint64(99942796153),
					"BTC": new(big.Int).SetUint64(5124885812),
				},
			},
		},
	}

	for i, tt := range tests {
		inputPaths, err := balanceInputsToInputPaths(tt.balanceInputs)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !common.DeepEqualMapBigInt2D(inputPaths, tt.inputPaths) {
			t.Errorf("failed on %d: input paths mismatch: have %v, want %v", i, inputPaths, tt.inputPaths)
		}
	}
}

func TestBalanceInputsToInitialProportions(t *testing.T) {
	tests := []struct {
		balanceInputs      []*pooldb.BalanceInput
		initialProportions map[string]map[string]map[uint64]*big.Int
	}{
		{
			balanceInputs: []*pooldb.BalanceInput{
				&pooldb.BalanceInput{
					MinerID:    1,
					ChainID:    "ETC",
					OutChainID: "ETH",
					Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(32)},
				},
				&pooldb.BalanceInput{
					MinerID:    2,
					ChainID:    "ETC",
					OutChainID: "ETH",
					Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(97)},
				},
			},
			initialProportions: map[string]map[string]map[uint64]*big.Int{
				"ETH": map[string]map[uint64]*big.Int{
					"ETC": map[uint64]*big.Int{
						1: new(big.Int).SetUint64(32),
						2: new(big.Int).SetUint64(97),
					},
				},
			},
		},
	}

	for i, tt := range tests {
		initialProportions, err := balanceInputsToInitialProportions(tt.balanceInputs)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !common.DeepEqualMapBigInt3D(initialProportions, tt.initialProportions) {
			t.Errorf("failed on %d: initial proportions mismatch: have %v, want %v", i,
				initialProportions, tt.initialProportions)
		}
	}
}

func TestExchangeInputsToOutputPaths(t *testing.T) {
	tests := []struct {
		exchangeInputs []*pooldb.ExchangeInput
		outputPaths    map[string]map[string]*big.Int
	}{
		{
			exchangeInputs: []*pooldb.ExchangeInput{
				&pooldb.ExchangeInput{
					InChainID:  "ETC",
					OutChainID: "ETH",
					Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(32)},
				},
			},
			outputPaths: map[string]map[string]*big.Int{
				"ETC": map[string]*big.Int{
					"ETH": new(big.Int).SetUint64(32),
				},
			},
		},
		{
			exchangeInputs: []*pooldb.ExchangeInput{
				&pooldb.ExchangeInput{
					InChainID:  "ETC",
					OutChainID: "ETH",
					Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(5818283841)},
				},
				&pooldb.ExchangeInput{
					InChainID:  "ETC",
					OutChainID: "ETH",
					Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(94124512312)},
				},
				&pooldb.ExchangeInput{
					InChainID:  "ETC",
					OutChainID: "BTC",
					Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(5124885812)},
				},
			},
			outputPaths: map[string]map[string]*big.Int{
				"ETC": map[string]*big.Int{
					"ETH": new(big.Int).SetUint64(99942796153),
					"BTC": new(big.Int).SetUint64(5124885812),
				},
			},
		},
	}

	for i, tt := range tests {
		outputPaths, err := exchangeInputsToOutputPaths(tt.exchangeInputs)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !common.DeepEqualMapBigInt2D(outputPaths, tt.outputPaths) {
			t.Errorf("failed on %d: output paths mismatch: have %v, want %v", i, outputPaths, tt.outputPaths)
		}
	}
}

func TestFinalTradesToFinalProportions(t *testing.T) {
	tests := []struct {
		finalTrades      []*pooldb.ExchangeTrade
		finalProportions map[string]map[string]*big.Int
	}{
		{
			finalTrades: []*pooldb.ExchangeTrade{
				&pooldb.ExchangeTrade{
					InitialChainID: "ETC",
					ToChainID:      "ETH",
					Proceeds:       dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(32)},
				},
			},
			finalProportions: map[string]map[string]*big.Int{
				"ETH": map[string]*big.Int{
					"ETC": new(big.Int).SetUint64(32),
				},
			},
		},
		{
			finalTrades: []*pooldb.ExchangeTrade{
				&pooldb.ExchangeTrade{
					InitialChainID: "ETC",
					ToChainID:      "ETH",
					Proceeds:       dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(5818283841)},
				},
				&pooldb.ExchangeTrade{
					InitialChainID: "ETC",
					ToChainID:      "ETH",
					Proceeds:       dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(94124512312)},
				},
				&pooldb.ExchangeTrade{
					InitialChainID: "ETC",
					ToChainID:      "BTC",
					Proceeds:       dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(5124885812)},
				},
			},
			finalProportions: map[string]map[string]*big.Int{
				"ETH": map[string]*big.Int{
					"ETC": new(big.Int).SetUint64(99942796153),
				},
				"BTC": map[string]*big.Int{
					"ETC": new(big.Int).SetUint64(5124885812),
				},
			},
		},
		{
			finalTrades: []*pooldb.ExchangeTrade{
				&pooldb.ExchangeTrade{
					InitialChainID: "ETC",
					ToChainID:      "BTC",
					Proceeds:       dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(6132400)},
				},
				&pooldb.ExchangeTrade{
					InitialChainID: "ETC",
					ToChainID:      "ETH",
					Proceeds:       dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(2997019000000000000)},
				},
				&pooldb.ExchangeTrade{
					InitialChainID: "FLUX",
					ToChainID:      "ETH",
					Proceeds:       dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(199449400000000000)},
				},
				&pooldb.ExchangeTrade{
					InitialChainID: "FLUX",
					ToChainID:      "BTC",
					Proceeds:       dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(1675000)},
				},
			},
			finalProportions: map[string]map[string]*big.Int{
				"BTC": map[string]*big.Int{
					"ETC":  new(big.Int).SetUint64(6132400),
					"FLUX": new(big.Int).SetUint64(1675000),
				},
				"ETH": map[string]*big.Int{
					"ETC":  new(big.Int).SetUint64(2997019000000000000),
					"FLUX": new(big.Int).SetUint64(199449400000000000),
				},
			},
		},
	}

	for i, tt := range tests {
		finalProportions, err := finalTradesToFinalProportions(tt.finalTrades)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !common.DeepEqualMapBigInt2D(finalProportions, tt.finalProportions) {
			t.Errorf("failed on %d: final proportions mismatch: have %v, want %v", i,
				finalProportions, tt.finalProportions)
		}
	}
}

func TestFinalTradesToAvgWeightedPrice(t *testing.T) {
	tests := []struct {
		finalTrades       []*pooldb.ExchangeTrade
		avgWeightedPrices map[string]map[string]float64
	}{
		{
			finalTrades: []*pooldb.ExchangeTrade{
				&pooldb.ExchangeTrade{
					InitialChainID:      "ETC",
					ToChainID:           "ETH",
					CumulativeFillPrice: types.Float64Ptr(0.05),
					Proceeds:            dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(5818283841)},
				},
				&pooldb.ExchangeTrade{
					InitialChainID:      "ETC",
					ToChainID:           "ETH",
					CumulativeFillPrice: types.Float64Ptr(0.04),
					Proceeds:            dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(9412451231)},
				},
				&pooldb.ExchangeTrade{
					InitialChainID:      "ETC",
					ToChainID:           "BTC",
					CumulativeFillPrice: types.Float64Ptr(0.05),
					Proceeds:            dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(5124885812)},
				},
			},
			avgWeightedPrices: map[string]map[string]float64{
				"ETC": map[string]float64{
					"ETH": 0.0438200939176575,
					"BTC": 0.05,
				},
			},
		},
	}

	for i, tt := range tests {
		avgWeightedPrices, err := finalTradesToAvgWeightedPrice(tt.finalTrades)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !reflect.DeepEqual(avgWeightedPrices, tt.avgWeightedPrices) {
			t.Errorf("failed on %d: avg weighted prices mismatch: have %v, want %v", i,
				avgWeightedPrices, tt.avgWeightedPrices)
		}
	}
}

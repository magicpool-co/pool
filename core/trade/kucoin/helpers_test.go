package kucoin

import (
	"math/big"
	"testing"

	"github.com/magicpool-co/pool/pkg/common"
)

func TestParseIncrement(t *testing.T) {
	tests := []struct {
		raw    string
		parsed int
	}{
		{
			raw:    "0.1",
			parsed: 1e1,
		},
		{
			raw:    "0.01",
			parsed: 1e2,
		},
		{
			raw:    "0.00001",
			parsed: 1e5,
		},
		{
			raw:    "0.000001",
			parsed: 1e6,
		},
	}

	for i, tt := range tests {
		parsed, err := parseIncrement(tt.raw)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if parsed != tt.parsed {
			t.Errorf("failed on %d: parsed mismatch: have %d, want %d", i, parsed, tt.parsed)
		}
	}
}

func TestCalcFeeAsBig(t *testing.T) {
	tests := []struct {
		value   string
		feeRate string
		units   *big.Int
		fee     *big.Int
	}{
		{
			value:   "96.04799",
			feeRate: "0.001",
			units:   new(big.Int).SetUint64(1e9),
			fee:     new(big.Int).SetUint64(96047990),
		},
		{
			value:   "96.047990000",
			feeRate: "0.001",
			units:   new(big.Int).SetUint64(1e9),
			fee:     new(big.Int).SetUint64(96047990),
		},
		{
			value:   "0.30479900",
			feeRate: "0.01",
			units:   new(big.Int).SetUint64(1e8),
			fee:     new(big.Int).SetUint64(304799),
		},
	}

	for i, tt := range tests {
		fee, err := calcFeeAsBig(tt.value, tt.feeRate, tt.units)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if fee.Cmp(tt.fee) != 0 {
			t.Errorf("failed on %d: fee mismatch: have %s, want %s", i, fee, tt.fee)
		}
	}
}

func TestSafeSubtractFee(t *testing.T) {
	tests := []struct {
		chain          string
		quantity       string
		feeRate        string
		parsedQuantity float64
		parsedFee      float64
	}{
		{
			chain:          "BTC",
			quantity:       "0.30479900",
			feeRate:        "0.001",
			parsedQuantity: 0.304494,
			parsedFee:      0.000305,
		},
		{
			chain:          "BTC",
			quantity:       "0.30479900",
			feeRate:        "0.0025",
			parsedQuantity: 0.304037,
			parsedFee:      0.000762,
		},
		{
			chain:          "USDT",
			quantity:       "3512.3512561235",
			feeRate:        "0.0025",
			parsedQuantity: 3503.570300,
			parsedFee:      8.780956,
		},
	}

	for i, tt := range tests {
		parsedQuantity, parsedFee, err := safeSubtractFee(tt.chain, tt.quantity, tt.feeRate)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !common.AlmostEqualFloat64(parsedQuantity, tt.parsedQuantity) {
			t.Errorf("failed on %d: quantity mismatch: have %f, want %f", i, parsedQuantity, tt.parsedQuantity)
		} else if !common.AlmostEqualFloat64(parsedFee, tt.parsedFee) {
			t.Errorf("failed on %d: fee mismatch: have %f, want %f", i, parsedFee, tt.parsedFee)
		}
	}
}

package common

import (
	"math/big"
	"testing"
)

func TestSafeRoundedFloat(t *testing.T) {
	tests := []struct {
		value       float64
		decimals    int
		outputValue float64
	}{
		{
			value:       32.51235123,
			decimals:    2,
			outputValue: 32.510000,
		},
		{
			value:       32.51235123,
			decimals:    4,
			outputValue: 32.512400,
		},
		{
			value:       32.51235123,
			decimals:    10,
			outputValue: 32.512351,
		},
		{
			value:       3200.523,
			decimals:    10,
			outputValue: 3200.523000,
		},
		{
			value:       320523510.523,
			decimals:    2,
			outputValue: 320523510.520000,
		},
	}

	for i, tt := range tests {
		outputValue := SafeRoundedFloat(tt.value, tt.decimals)
		if !AlmostEqualFloat64(outputValue, tt.outputValue) {
			t.Errorf("failed on %d: value mismatch: have %f, want %f", i, outputValue, tt.outputValue)
		}
	}
}

func TestFloorFloatByIncrement(t *testing.T) {
	tests := []struct {
		value       float64
		incr        int
		exp         int
		outputValue float64
	}{
		{
			value:       32.51235123,
			incr:        1e6,
			exp:         1e8,
			outputValue: 32.512351,
		},
		{
			value:       32.51235123,
			incr:        1e4,
			exp:         1e8,
			outputValue: 32.512300,
		},
		{
			value:       32.51235123,
			incr:        1e2,
			exp:         1e8,
			outputValue: 32.510000,
		},
		{
			value:       32.51235123,
			incr:        1e1,
			exp:         1e8,
			outputValue: 32.500000,
		},
		{
			value:       32.51235123,
			incr:        0,
			exp:         1e8,
			outputValue: 32.51235123,
		},
	}

	for i, tt := range tests {
		outputValue := FloorFloatByIncrement(tt.value, tt.incr, tt.exp)
		if !AlmostEqualFloat64(outputValue, tt.outputValue) {
			t.Errorf("failed on %d: value mismatch: have %f, want %f", i, outputValue, tt.outputValue)
		}
	}
}

func TestSplitBigPercentage(t *testing.T) {
	tests := []struct {
		input       *big.Int
		numerator   uint64
		denominator uint64
		output      *big.Int
	}{
		{
			input:       new(big.Int).SetUint64(51235123),
			numerator:   50,
			denominator: 100,
			output:      new(big.Int).SetUint64(25617561),
		},
		{
			input:       new(big.Int).SetUint64(51235123),
			numerator:   90,
			denominator: 100,
			output:      new(big.Int).SetUint64(46111610),
		},
		{
			input:       new(big.Int).SetUint64(51235123),
			numerator:   23451,
			denominator: 51299591239,
			output:      new(big.Int).SetUint64(23),
		},
	}

	for i, tt := range tests {
		output := SplitBigPercentage(tt.input, tt.numerator, tt.denominator)
		if output.Cmp(tt.output) != 0 {
			t.Errorf("failed on %d: output mismatch: have %s, want %s", i, output, tt.output)
		}
	}
}

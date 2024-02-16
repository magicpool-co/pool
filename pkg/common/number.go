package common

import (
	"math"
	"math/big"
)

func AlmostEqualFloat64(a, b float64) bool {
	const equalityThreshold = 1e-6
	return math.Abs(a-b) <= equalityThreshold
}

func SafeRoundedFloat(value float64, decimals int) float64 {
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return 0
	}

	exp := math.Pow(10, float64(decimals))
	value = math.Round(value*exp) / exp

	return value
}

func FloorFloatByIncrement(value float64, incr, exp int) float64 {
	if incr > exp || incr == 0 {
		return value
	}

	intIncr := new(big.Int).SetUint64(uint64(exp / incr))
	intExp := new(big.Int).SetUint64(uint64(exp))
	intVal := Float64ToBigInt(value, intExp)
	rem := new(big.Int).Mod(intVal, intIncr)
	intVal.Sub(intVal, rem)

	return BigIntToFloat64(intVal, intExp)
}

func SplitBigPercentage(input *big.Int, numerator, denominator uint64) *big.Int {
	numeratorBig := new(big.Int).SetUint64(numerator)
	output := new(big.Int).Mul(input, numeratorBig)
	if denominator > 0 {
		denominatorBig := new(big.Int).SetUint64(denominator)
		output.Div(output, denominatorBig)
	}

	return output
}

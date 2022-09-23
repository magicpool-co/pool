package common

import (
	"math"
	"math/big"
)

func SafeRoundedFloat(val float64) float64 {
	if math.IsInf(val, 0) || math.IsNaN(val) {
		return 0
	}

	return math.Round(val*1000) / 1000
}

func int64Pow(x, n int) int {
	val := 1
	for i := 0; i < n; i++ {
		val *= x
	}

	return val
}

func FloorFloatByIncrement(val float64, incr, exp int) float64 {
	intIncr := new(big.Int).SetUint64(uint64(int64Pow(10, incr)))
	intExp := new(big.Int).SetUint64(uint64(exp))

	intVal := Float64ToBigInt(val, intExp)
	rem := new(big.Int).Mod(intVal, intIncr)
	intVal.Sub(intVal, rem)

	return BigIntToFloat64(intVal, intExp)
}

func SplitBigPercentage(input *big.Int, numerator, denominator uint64) *big.Int {
	output := new(big.Int)
	output.Mul(input, new(big.Int).SetUint64(numerator))
	output.Div(output, new(big.Int).SetUint64(denominator))

	return output
}

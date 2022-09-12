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

func SplitBigPercentage(input *big.Int, numerator, denominator uint64) *big.Int {
	output := new(big.Int)
	output.Mul(input, new(big.Int).SetUint64(numerator))
	output.Div(output, new(big.Int).SetUint64(denominator))

	return output
}

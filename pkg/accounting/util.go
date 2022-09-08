package accounting

import (
	"math/big"
)

/* big int utils */

func adjustValue(input, numerator, denominator *big.Int) *big.Int {
	output := new(big.Int)
	output.Mul(input, numerator)
	output.Div(output, denominator)

	return output
}

func splitBig(value *big.Int, fraction int64) *big.Int {
	return adjustValue(value, big.NewInt(fraction), big.NewInt(10000))
}

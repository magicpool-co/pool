package common

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const (
	zeroHex1   = "0"
	zeroHex256 = "0000000000000000000000000000000000000000000000000000000000000000"
)

var (
	Big0  = new(big.Int).SetUint64(0)
	Big1  = new(big.Int).SetUint64(1)
	Big10 = new(big.Int).SetUint64(10)
)

func HexToBig(str string) (*big.Int, error) {
	if len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X') {
		str = str[2:]
	}

	val := new(big.Int)
	val.SetString(str, 16)

	if len(str) > 0 && str != zeroHex1 && str != zeroHex256 {
		if val.Cmp(big.NewInt(0)) <= 0 {
			return nil, fmt.Errorf("negative hex string")
		}
	}

	return val, nil
}

func HexToUint64(str string) (uint64, error) {
	if len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X') {
		str = str[2:]
	}

	val, err := strconv.ParseUint(str, 16, 64)
	if err != nil {
		return 0, err
	}

	return uint64(val), nil
}

func HexToBytes(str string) ([]byte, error) {
	if len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X') {
		str = str[2:]
	}

	if len(str)%2 == 1 {
		str = "0" + str
	}

	return hex.DecodeString(str)
}

func Uint64ToHex(inp uint64) string {
	return "0x" + strconv.FormatUint(inp, 16)
}

func Float64ToBigInt(num float64, exp *big.Int) *big.Int {
	numFloat := new(big.Float).SetFloat64(num)
	expFloat := new(big.Float).SetInt(exp)

	val, _ := new(big.Float).Mul(numFloat, expFloat).Int(new(big.Int))

	return val
}

func BigIntToFloat64(num, exp *big.Int) float64 {
	numFloat := new(big.Float).SetInt(num)
	expFloat := new(big.Float).SetInt(exp)

	val, _ := new(big.Float).Quo(numFloat, expFloat).Float64()

	return val
}

func StringDecimalToBigint(dec string, exp *big.Int) (*big.Int, error) {
	// separate the integer and fraction parts of the decimal
	raw := strings.Split(dec, ".")
	if len(raw) != 2 {
		if strings.ReplaceAll(dec, "0", "") == "" {
			return new(big.Int), nil
		} else if len(raw) != 1 {
			return nil, fmt.Errorf("invalid decimal, %s", dec)
		}

		raw = append(raw, "0")
	}

	intPart, fracPart := raw[0], raw[1]

	// verify that the fraction size is not greater than the exponent size
	fracDiff := (len(exp.String()) - 1) - len(fracPart)

	// set the integer part as big int
	intBig, ok := new(big.Int).SetString(intPart, 10)
	if !ok && len(intPart) > 0 {
		return nil, fmt.Errorf("unable to convert integer part to big int")
	} else if intBig == nil {
		intBig = new(big.Int)
	}

	// multiply the integer by the exponent size
	intBig.Mul(intBig, exp)

	// set the fraction part as big int
	fracBig, ok := new(big.Int).SetString(fracPart, 10)
	if !ok && len(fracPart) > 0 {
		return nil, fmt.Errorf("unable to convert fraction part to big int")
	} else if fracBig == nil {
		fracBig = new(big.Int)
	}

	// normalize the fraction part if less than the exponent size
	if fracDiff > 0 {
		fracBig.Mul(fracBig, new(big.Int).Exp(Big10, big.NewInt(int64(fracDiff)), nil))
	} else if fracDiff < 0 {
		fracBig.Div(fracBig, new(big.Int).Exp(Big10, big.NewInt(int64(-fracDiff)), nil))
	}

	// return the sum of the integer and fraction parts
	return new(big.Int).Add(intBig, fracBig), nil
}

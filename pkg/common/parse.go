package common

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

func MustParseHex(inp string) []byte {
	data, err := hex.DecodeString(inp)
	if err != nil {
		panic(fmt.Errorf("MustParseHex: %v", err))
	}

	return data
}

func MustParseBigHex(inp string) *big.Int {
	data, ok := new(big.Int).SetString(inp, 16)
	if !ok {
		panic("MustParseBigHex256: invalid input")
	}

	return data
}

func MustParseBigInt(inp string) *big.Int {
	data, ok := new(big.Int).SetString(inp, 10)
	if !ok {
		panic("MustParseBigInt: invalid input")
	}

	return data
}

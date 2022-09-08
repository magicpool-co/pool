package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

var (
	limit = mustParseBig256("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364140")
)

func Obscure(input string) (string, error) {
	rawInput, err := hex.DecodeString(input)
	if err != nil {
		return "", err
	}

	hashed := hashSHA256(rawInput)
	obscured := bytes.Join([][]byte{
		hashed[29:31],
		hashed[:10],
		hashed[31:],
		hashed[10:29],
	}, nil)

	final := hashSHA256(obscured)

	if err := checkValid(final); err != nil {
		return "", err
	}

	output := hex.EncodeToString(final)

	return output, nil
}

func hashSHA256(input []byte) []byte {
	val := sha256.Sum256(input)

	return val[:]
}

func mustParseBig256(inp string) *big.Int {
	val, ok := new(big.Int).SetString(inp, 16)
	if !ok {
		panic("invalid 256 bit integer: " + inp)
	} else if val.BitLen() > 256 {
		panic("invalid 256 bit integer: " + inp)
	}

	return val
}

func checkValid(input []byte) error {
	if len(input) != 32 {
		return fmt.Errorf("input has invalid length")
	}

	val := new(big.Int).SetBytes(input)
	if limit.Cmp(val) < 0 {
		return fmt.Errorf("input too big")
	} else if val.Cmp(new(big.Int)) <= 0 {
		return fmt.Errorf("input too small")
	}

	return nil
}

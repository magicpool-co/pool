package crypto

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

/* general utilities */

func ReverseBytes(b []byte) []byte {
	_b := make([]byte, len(b))
	copy(_b, b)

	for i, j := 0, len(_b)-1; i < j; i, j = i+1, j-1 {
		_b[i], _b[j] = _b[j], _b[i]
	}
	return _b
}

func ValidateSecp256k1PrivateKey(key []byte) error {
	limit, ok := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364140", 16)
	if !ok {
		return fmt.Errorf("unable to parse limit")
	}

	val := new(big.Int).SetBytes(key)
	if limit.Cmp(val) < 0 {
		return fmt.Errorf("input too big")
	} else if val.Cmp(new(big.Int)) <= 0 {
		return fmt.Errorf("input too small")
	}

	return nil
}

func DivideIntCeil(x, y int) int {
	return (x + y - 1) / y
}

func SerializeBlockHeight(blockHeight uint64) ([]byte, []byte, error) {
	blockHeightSerial := fmt.Sprintf("%x", blockHeight)
	if len(blockHeightSerial)%2 == 1 {
		blockHeightSerial = "0" + blockHeightSerial
	}

	height := DivideIntCeil(len(fmt.Sprintf("%b", (blockHeight<<1))), 8)
	lengthDiff := len(blockHeightSerial)/2 - height
	for i := 0; i < lengthDiff; i++ {
		blockHeightSerial += "00"
	}

	blockHeightSerialBytes, err := hex.DecodeString(blockHeightSerial)
	if err != nil {
		return nil, nil, err
	}

	length := fmt.Sprintf("0%d", height)
	lengthBytes, err := hex.DecodeString(length)
	if err != nil {
		return nil, nil, err
	}

	return blockHeightSerialBytes, lengthBytes, nil
}

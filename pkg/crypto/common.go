package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/dchest/blake2b"
	blake2bStd "golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"

	"github.com/magicpool-co/pool/pkg/common"
)

var (
	secp256k1Limit = common.MustParseBigHex("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364140")
)

/* hash functions */

func Sha256(b []byte) []byte {
	d := sha256.Sum256(b)
	return d[:]
}

func Sha256d(b []byte) []byte {
	return Sha256(Sha256(b))
}

func Sha512(b []byte) []byte {
	d := sha512.Sum512(b)
	return d[:]
}

func Keccak256(b []byte) []byte {
	d := sha3.NewLegacyKeccak256()
	d.Write(b)
	return d.Sum(nil)
}

func Ripemd160(b []byte) []byte {
	d := ripemd160.New()
	d.Reset()
	d.Write(b)
	return d.Sum(nil)
}

func Blake2b256(data []byte) []byte {
	out := blake2b.Sum256(data)
	return out[:]
}

func Blake2b256Personal(data, personal []byte) ([]byte, error) {
	d, err := blake2b.New(&blake2b.Config{Person: personal, Size: 32})
	if err != nil {
		return nil, err
	}
	d.Write(data)
	return d.Sum(nil), nil
}

func Blake2b256MAC(data, key []byte) ([]byte, error) {
	hasher, err := blake2bStd.New(32, key)
	if err != nil {
		return nil, err
	}
	hasher.Write(data)
	return hasher.Sum(nil), nil
}

func HmacSha256(key, data string) []byte {
	d := hmac.New(sha256.New, []byte(key))
	d.Write([]byte(data))
	return d.Sum(nil)
}

func HmacSha512(key, data string) []byte {
	d := hmac.New(sha512.New, []byte(key))
	d.Write([]byte(data))
	return d.Sum(nil)
}

/* byte helpers */

func ReverseBytes(a []byte) []byte {
	b := make([]byte, len(a))
	copy(b, a)

	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	return b
}

func EthashSeedHash(height, epochLength uint64) []byte {
	epoch := height / epochLength
	seedHash := make([]byte, 32)

	var i uint64
	for i = 0; i < epoch; i++ {
		seedHash = Keccak256(seedHash)
	}

	return seedHash
}

func ObscureHex(input string) ([]byte, error) {
	rawInput, err := hex.DecodeString(input)
	if err != nil {
		return nil, err
	}

	hashed := Sha256(rawInput)
	obscured := bytes.Join([][]byte{
		hashed[29:31],
		hashed[:10],
		hashed[31:],
		hashed[10:29],
	}, nil)

	final := Sha256(obscured)

	// validate that key is within secp256k1 range
	value := new(big.Int).SetBytes(final)
	if secp256k1Limit.Cmp(value) < 0 {
		return nil, fmt.Errorf("input too big")
	} else if value.Cmp(common.Big0) <= 0 {
		return nil, fmt.Errorf("input too small")
	}

	return final, nil
}

/* block helpers */

func divideIntCeil(x, y int) int {
	return (x + y - 1) / y
}

func SerializeBlockHeight(blockHeight uint64) ([]byte, []byte, error) {
	blockHeightSerial := fmt.Sprintf("%x", blockHeight)
	if len(blockHeightSerial)%2 == 1 {
		blockHeightSerial = "0" + blockHeightSerial
	}

	height := divideIntCeil(len(fmt.Sprintf("%b", (blockHeight<<1))), 8)
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

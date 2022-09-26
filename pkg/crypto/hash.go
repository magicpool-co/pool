package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"github.com/dchest/blake2b"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

// standard

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

// custom

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

	return final, nil
}

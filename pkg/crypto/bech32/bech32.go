package bech32

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil/bech32"
)

const (
	checksumLength = 8
)

func encodeToBase32(charset string, data []byte) string {
	// build body string from alphabet
	builder := strings.Builder{}
	for _, v := range data {
		builder.WriteRune(rune(charset[v]))
	}

	return builder.String()
}

func decodeFromBase32(charset, encoded string) ([]byte, error) {
	decoded := make([]byte, 0, len(encoded))
	for i := 0; i < len(encoded); i++ {
		index := strings.IndexByte(charset, encoded[i])
		if index < 0 {
			return nil, fmt.Errorf("invalid bech32 character")
		}
		decoded = append(decoded, byte(index))
	}

	return decoded, nil
}

func polymod(v []byte) uint64 {
	var c uint64 = 1
	for _, d := range v {
		c0 := byte(c >> 35)
		c = ((c & 0x07ffffffff) << 5) ^ uint64(d)
		if c0&0x01 != 0 {
			c ^= 0x98f2bc8e61
		}
		if c0&0x02 != 0 {
			c ^= 0x79b76d99e2
		}
		if c0&0x04 != 0 {
			c ^= 0xf33e5fb3c4
		}
		if c0&0x08 != 0 {
			c ^= 0xae2eabe2a8
		}
		if c0&0x10 != 0 {
			c ^= 0x1e4f43e470
		}
	}
	return c ^ 1
}

func calcChecksum(prefix string, body []byte) ([]byte, error) {
	var lower5bitsNettype []byte
	for _, v := range prefix {
		lower5bitsNettype = append(lower5bitsNettype, byte(v)&0x1f)
	}

	template := [8]byte{}

	checksumInput := append(lower5bitsNettype, 0x00)
	checksumInput = append(checksumInput, body[:]...)
	checksumInput = append(checksumInput, template[:]...)

	uint64Chc := polymod(checksumInput)

	low40BitsChc := make([]byte, 8)
	for i := 0; i < 8; i++ {
		low40BitsChc[7-i] = byte(uint64Chc >> uint(i*8))
	}
	low40BitsChc = low40BitsChc[3:]

	checksumIn5Bits, err := bech32.ConvertBits(low40BitsChc, 8, 5, true)
	if err != nil {
		return nil, err
	}

	return checksumIn5Bits, nil
}

func EncodeModified(charset, prefix string, version byte, body []byte) (string, error) {
	data := make([]byte, len(body)+1)
	data[0] = version
	copy(data[1:], body)

	converted, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", err
	}

	checksum, err := calcChecksum(prefix, converted)
	if err != nil {
		return "", err
	}

	address := encodeToBase32(charset, append(converted, checksum...))

	return prefix + ":" + address, nil
}

func DecodeModified(charset, encoded string) (string, byte, []byte, error) {
	if len(encoded) < checksumLength+2 {
		return "", 0, nil, fmt.Errorf("invalid bech32 string length")
	}

	for i := 0; i < len(encoded); i++ {
		if encoded[i] < 33 || encoded[i] > 126 {
			return "", 0, nil, fmt.Errorf("invalid bech32 character")
		}
	}

	lower := strings.ToLower(encoded)
	upper := strings.ToUpper(encoded)
	if encoded != lower && encoded != upper {
		return "", 0, nil, fmt.Errorf("string not all lowercase or all uppercase")
	}

	encoded = lower
	colonIndex := strings.LastIndexByte(encoded, ':')
	if colonIndex < 1 || colonIndex+checksumLength+1 > len(encoded) {
		return "", 0, nil, fmt.Errorf("invalid index of ':'")
	}

	prefix := encoded[:colonIndex]
	decoded, err := decodeFromBase32(charset, encoded[colonIndex+1:])
	if err != nil {
		return "", 0, nil, err
	}

	checksum := encoded[len(encoded)-checksumLength:]
	calculated, err := calcChecksum(prefix, decoded[:len(decoded)-checksumLength])
	if err != nil {
		return "", 0, nil, err
	} else if encodeToBase32(charset, calculated) != checksum {
		return "", 0, nil, fmt.Errorf("checksums do not match")
	}

	decoded = decoded[:len(decoded)-checksumLength]
	converted, err := bech32.ConvertBits(decoded, 5, 8, false)
	if err != nil {
		return "", 0, nil, err
	}

	return prefix, converted[0], converted[1:], nil
}

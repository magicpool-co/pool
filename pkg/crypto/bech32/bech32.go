// Copyright 2019 Conflux Foundation. All rights reserved.
// Conflux is free software and distributed under GNU General Public License.
// See http://www.gnu.org/licenses/

package bech32

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil/bech32"
)

const (
	checksumLength = 8
)

func encodeToBase32(charset string, data []byte) string {
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
		if index == -1 {
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
	var prefixAdjusted []byte
	for _, v := range prefix {
		prefixAdjusted = append(prefixAdjusted, byte(v)&0x1f)
	}

	checksumInput := bytes.Join([][]byte{
		prefixAdjusted,
		[]byte{0x00},
		body,
		make([]byte, 8),
	}, nil)

	checksumPolymod := polymod(checksumInput)
	checksumAdjusted := make([]byte, 8)
	for i := 0; i < 8; i++ {
		checksumAdjusted[7-i] = byte(checksumPolymod >> uint(i*8))
	}

	return bech32.ConvertBits(checksumAdjusted[3:], 8, 5, true)
}

func EncodeBCH(charset, prefix string, version byte, body []byte) (string, error) {
	data := bytes.Join([][]byte{
		[]byte{version},
		body,
	}, nil)

	decoded, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", err
	}

	checksum, err := calcChecksum(prefix, decoded)
	if err != nil {
		return "", err
	}

	decoded = append(decoded, checksum...)
	encoded := encodeToBase32(charset, decoded)

	return prefix + ":" + encoded, nil
}

func DecodeBCH(charset, encoded string) (string, byte, []byte, error) {
	if len(encoded) < checksumLength+2 {
		return "", 0, nil, fmt.Errorf("invalid bech32 string length")
	}

	for i := 0; i < len(encoded); i++ {
		if encoded[i] < 33 || encoded[i] > 126 {
			return "", 0, nil, fmt.Errorf("invalid bech32 character")
		}
	}

	encodedLower := strings.ToLower(encoded)
	encodedUpper := strings.ToUpper(encoded)
	if encoded != encodedLower && encoded != encodedUpper {
		return "", 0, nil, fmt.Errorf("string not all lowercase or all uppercase")
	}

	colonIdx := strings.LastIndexByte(encodedLower, ':')
	if colonIdx < 1 || colonIdx+checksumLength+1 > len(encodedLower) {
		return "", 0, nil, fmt.Errorf("invalid index of ':'")
	}
	prefix := encodedLower[:colonIdx]

	decoded, err := decodeFromBase32(charset, encodedLower[colonIdx+1:])
	if err != nil {
		return "", 0, nil, err
	}
	checksum := encodedLower[len(encodedLower)-checksumLength:]

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

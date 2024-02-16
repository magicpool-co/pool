/* btckeygenie v1.0.0
 * https://github.com/vsergeev/btckeygenie
 * License: MIT
 */

package base58

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/ripemd160"

	"github.com/magicpool-co/pool/pkg/crypto"
)

const (
	base58Table = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

var (
	big0  = new(big.Int)
	big58 = new(big.Int).SetUint64(58)
)

// Encode encodes a byte slice into a base-58 encoded string
func Encode(decoded []byte) string {
	value := new(big.Int).SetBytes(decoded)
	rem := new(big.Int)

	var encoded string
	for value.Cmp(big0) > 0 {
		value.QuoRem(value, big58, rem)
		encoded = string(base58Table[rem.Int64()]) + encoded
	}

	return encoded
}

// Decode decodes a base-58 encoded string into a byte slice
func Decode(encoded string) ([]byte, error) {
	value := big.NewInt(0)
	for i := 0; i < len(encoded); i++ {
		b58Idx := strings.IndexByte(base58Table, encoded[i])
		if b58Idx == -1 {
			return nil, fmt.Errorf("invalid base58 character")
		}

		value.Mul(value, big58)
		value.Add(value, new(big.Int).SetInt64(int64(b58Idx)))
	}

	return value.Bytes(), nil
}

// CheckEncode encodes a version and a byte slice into a base-58 check encoded string.
// Version is a slice, not a byte, to handle for multi-length prefixes (ZCash).
func CheckEncode(version, decoded []byte) string {
	checkDecoded := append(version, decoded...)
	checksum := crypto.Sha256d(checkDecoded)
	checkDecoded = append(checkDecoded, checksum[0:4]...)
	checkEncoded := Encode(checkDecoded)

	for _, v := range checkDecoded {
		if v != 0 {
			break
		}
		checkEncoded = "1" + checkEncoded
	}

	return checkEncoded
}

// CheckDecode decodes a base-58 check encoded string into a version and a byte slice.
// Version is a slice, not a byte, to handle for multi-length prefixes (ZCash).
func CheckDecode(checkEncoded string) ([]byte, []byte, error) {
	checkDecoded, err := Decode(checkEncoded)
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < len(checkEncoded); i++ {
		if checkEncoded[i] != '1' {
			break
		}
		checkDecoded = append([]byte{0x00}, checkDecoded...)
	}

	if len(checkDecoded) < 5 {
		return nil, nil, fmt.Errorf("missing checksum")
	}

	checksum := crypto.Sha256d(checkDecoded[:len(checkDecoded)-4])
	if bytes.Compare(checksum[0:4], checkDecoded[len(checkDecoded)-4:]) != 0 {
		return nil, nil, fmt.Errorf("invalid checksum")
	}
	checkDecoded = checkDecoded[:len(checkDecoded)-4]

	prefixLength := len(checkDecoded) - ripemd160.Size
	if prefixLength < 0 {
		return nil, nil, fmt.Errorf("invalid prefix length")
	}

	version := checkDecoded[:prefixLength]
	decoded := checkDecoded[prefixLength:]

	return version, decoded, nil
}

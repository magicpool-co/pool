/* btckeygenie v1.0.0
 * https://github.com/vsergeev/btckeygenie
 * License: MIT
 */

package util

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
)

// b58encode encodes a byte slice b into a base-58 encoded string.
func EncodeBase58(b []byte) (s string) {
	/* See https://en.bitcoin.it/wiki/Base58Check_encoding */

	const BITCOIN_BASE58_TABLE = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	/* Convert big endian bytes to big int */
	x := new(big.Int).SetBytes(b)

	/* Initialize */
	r := new(big.Int)
	m := big.NewInt(58)
	zero := big.NewInt(0)
	s = ""

	/* Convert big int to string */
	for x.Cmp(zero) > 0 {
		/* x, r = (x / 58, x % 58) */
		x.QuoRem(x, m, r)
		/* Prepend ASCII character */
		s = string(BITCOIN_BASE58_TABLE[r.Int64()]) + s
	}

	return s
}

// b58decode decodes a base-58 encoded string into a byte slice b.
func DecodeBase58(s string) (b []byte, err error) {
	/* See https://en.bitcoin.it/wiki/Base58Check_encoding */

	const BITCOIN_BASE58_TABLE = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	/* Initialize */
	x := big.NewInt(0)
	m := big.NewInt(58)

	/* Convert string to big int */
	for i := 0; i < len(s); i++ {
		b58index := strings.IndexByte(BITCOIN_BASE58_TABLE, s[i])
		if b58index == -1 {
			return nil, fmt.Errorf("Invalid base-58 character encountered: '%c', index %d.", s[i], i)
		}
		b58value := big.NewInt(int64(b58index))
		x.Mul(x, m)
		x.Add(x, b58value)
	}

	/* Convert big int to big endian bytes */
	b = x.Bytes()

	return b, nil
}

// b58checkencode encodes version ver and byte slice b into a base-58 check encoded string.
func CheckEncodeBase58(ver uint8, b []byte) (s string) {
	/* Prepend version */
	bcpy := append([]byte{ver}, b...)

	hash1 := Sha256(bcpy)
	hash2 := Sha256(hash1)

	/* Append first four bytes of hash */
	bcpy = append(bcpy, hash2[0:4]...)

	/* Encode base58 string */
	s = EncodeBase58(bcpy)

	/* For number of leading 0's in bytes, prepend 1 */
	for _, v := range bcpy {
		if v != 0 {
			break
		}
		s = "1" + s
	}

	return s
}

// b58checkdecode decodes base-58 check encoded string s into a version ver and byte slice b.
func CheckDecodeBase58(s string, verLen int) (ver []byte, b []byte, err error) {
	/* Decode base58 string */
	b, err = DecodeBase58(s)
	if err != nil {
		return nil, nil, err
	}

	/* Add leading zero bytes */
	for i := 0; i < len(s); i++ {
		if s[i] != '1' {
			break
		}
		b = append([]byte{0x00}, b...)
	}

	/* Verify checksum */
	if len(b) < 5 {
		return nil, nil, fmt.Errorf("Invalid base-58 check string: missing checksum.")
	}

	hash1 := Sha256(b[:len(b)-4])
	hash2 := Sha256(hash1)

	/* Compare checksum */
	if bytes.Compare(hash2[0:4], b[len(b)-4:]) != 0 {
		return nil, nil, fmt.Errorf("Invalid base-58 check string: invalid checksum.")
	}

	/* Strip checksum bytes */
	b = b[:len(b)-4]

	/* Extract and strip version */
	ver = b[:verLen]
	b = b[verLen:]

	return ver, b, nil
}

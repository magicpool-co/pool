package cfx

import (
	"fmt"
	"strings"
)

const (
	typeBits    = 0
	addressType = 0
	sizeBits    = 0
	alphabet    = "abcdefghjkmnprstuvwxyz0123456789"
)

var (
	versionByte byte = (typeBits & 0x80) | (addressType << 3) | sizeBits
)

func convert(data []byte, inBits uint, outBits uint) ([]byte, error) {
	if inBits > 8 || outBits > 8 {
		return nil, fmt.Errorf("only support bits length<=8")
	}

	var accBits uint
	var acc uint16
	var ret []byte
	for _, d := range data {
		acc = acc<<uint16(inBits) | uint16(d)
		accBits += inBits
		for accBits >= outBits {
			ret = append(ret, byte(acc>>uint16(accBits-outBits)))
			acc = acc & uint16(1<<(accBits-outBits)-1)
			accBits -= outBits
		}
	}

	if accBits > 0 && (inBits > outBits) {
		ret = append(ret, byte(acc<<uint16(outBits-accBits)))
	}

	return ret, nil
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

func calcChecksum(body []byte, networkPrefix string) ([]byte, error) {
	var lower5bitsNettype []byte
	for _, v := range networkPrefix {
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

	checksumIn5Bits, err := convert(low40BitsChc, 8, 5)
	if err != nil {
		return nil, err
	}

	return checksumIn5Bits, nil
}

func ETHAddressToCFX(ethAddress []byte, networkPrefix string) (string, error) {
	ethAddress = append([]byte{versionByte}, ethAddress...)
	body, err := convert(ethAddress, 8, 5)
	if err != nil {
		return "", err
	}

	checksum, err := calcChecksum(body, networkPrefix)
	if err != nil {
		return "", err
	}

	// build body string from alphabet
	builder := strings.Builder{}
	for _, v := range append(body, checksum...) {
		builder.WriteRune(rune(alphabet[v]))
	}

	address := fmt.Sprintf("%s:%s", networkPrefix, builder.String())

	return address, nil
}

package crypto

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

/* general utilities */

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

func HexToUint64(inp string) (uint64, error) {
	inp = strings.ReplaceAll(inp, "0x", "")
	val, err := strconv.ParseUint(inp, 16, 64)
	if err != nil {
		return 0, err
	}

	return uint64(val), nil
}

/* bignum utilities */

func StringToBig(h string) *big.Int {
	n := new(big.Int)
	n.SetString(h, 0)
	return n
}

func StringToFloat(h string) float64 {
	n := new(big.Int)
	n.SetString(h, 16)
	f := new(big.Float).SetInt(n)
	f64, _ := f.Float64()
	return f64
}

func BytesToBig(data []byte) *big.Int {
	n := new(big.Int)
	n.SetBytes(data)
	return n
}

/* crypto utilities */

func PadUint64(seq uint64) []byte {
	buf := make([]byte, 8)
	for i := len(buf) - 1; seq != 0; i-- {
		buf[i] = byte(seq & 0xff)
		seq >>= 8
	}
	return buf
}

func VarIntSerializeSize(val uint64) int {
	// The value is small enough to be represented by itself, so it's
	// just 1 byte.
	if val < 0xfd {
		return 1
	}

	// Discriminant 1 byte plus 2 bytes for the uint16.
	if val <= math.MaxUint16 {
		return 3
	}

	// Discriminant 1 byte plus 4 bytes for the uint32.
	if val <= math.MaxUint32 {
		return 5
	}

	// Discriminant 1 byte plus 8 bytes for the uint64.
	return 9
}

func SerializeNumber(num uint64) []byte {
	if num >= 1 && num <= 16 {
		return []byte{byte(0x50 + num)}
	} else if num <= 0x7F {
		return []byte{1, byte(num)}
	}

	var count int
	n := num
	buf := make([]byte, 0)
	for n > 0x7F {
		buf = append(buf, byte(n&0xFF))
		n = n >> 8
		count++
	}

	data := make([]byte, count+2)
	data[0] = byte(count + 1)
	copy(data[1:], buf) // 4
	data[count+1] = byte(n)

	return data
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

func VarIntToBytes(n uint64) []byte {
	if n < 0xFD {
		return []byte{byte(n)}
	}

	if n <= 0xFFFF {
		buff := make([]byte, 3)
		buff[0] = 0xFD
		binary.LittleEndian.PutUint16(buff[1:], uint16(n))
		return buff
	}

	if n <= 0xFFFFFFFF {
		buff := make([]byte, 5)
		buff[0] = 0xFE
		binary.LittleEndian.PutUint32(buff[1:], uint32(n))
		return buff
	}

	buff := make([]byte, 9)
	buff[0] = 0xFF
	binary.LittleEndian.PutUint64(buff[1:], uint64(n))
	return buff
}

func VarIntLength(prefix byte) int {
	switch prefix {
	case 0xFD:
		return 2
	case 0xFE:
		return 4
	case 0xFF:
		return 8
	default:
		return 1
	}
}

func BytesToVarInt(prefix int, raw []byte) (uint64, error) {
	switch prefix {
	case 1:
		return uint64(prefix), nil
	case 2:
		if len(raw) != prefix {
			return 0, fmt.Errorf("invalid var int length")
		}

		return uint64(binary.LittleEndian.Uint16(raw)), nil

	case 4:
		if len(raw) != prefix {
			return 0, fmt.Errorf("invalid var int length")
		}

		return uint64(binary.LittleEndian.Uint32(raw)), nil
	case 8:
		if len(raw) != prefix {
			return 0, fmt.Errorf("invalid var int length")
		}

		return binary.LittleEndian.Uint64(raw), nil

	default:
		return 0, fmt.Errorf("invalid var int prefix")
	}
}

/* binary utilites */

func PaddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}

func ReverseBytes(b []byte) []byte {
	_b := make([]byte, len(b))
	copy(_b, b)

	for i, j := 0, len(_b)-1; i < j; i, j = i+1, j-1 {
		_b[i], _b[j] = _b[j], _b[i]
	}
	return _b
}

func WriteUint8(d uint8) []byte {
	b := make([]byte, 1)
	b[0] = d

	return b
}

func WriteUint16Le(d uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, d)

	return b
}

func WriteUint16Be(d uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, d)

	return b
}

func WriteUint32Le(d uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, d)

	return b
}

func WriteUint32Be(d uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, d)

	return b
}

func WriteUint64Be(d uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, d)

	return b
}

func WriteUint64Le(d uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, d)

	return b
}

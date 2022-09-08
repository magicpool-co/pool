package types

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

var (
	ErrUnknownInputType = fmt.Errorf("unknown input type")
)

/* Hash type */

type Hash struct {
	hex   string
	bytes []byte
	big   *big.Int
}

func (h *Hash) UnmarshalJSON(data []byte) error {
	var value interface{}
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}

	switch v := value.(type) {
	case string:
		_, err = h.SetFromHex(v)
	default:
		err = ErrUnknownInputType
	}

	return err
}

func (h *Hash) Hex() string         { return h.hex }
func (h *Hash) PrefixedHex() string { return "0x" + h.hex }
func (h *Hash) Bytes() []byte       { return h.bytes }
func (h *Hash) Big() *big.Int       { return h.big }
func (h *Hash) MeetsDifficulty(diff *Difficulty) bool {
	return diff.TargetBig().Cmp(h.big) >= 0
}

func (h *Hash) SetFromBytes(value []byte) *Hash {
	h.hex = hex.EncodeToString(value)
	h.bytes = value
	h.big = new(big.Int).SetBytes(value)

	return h
}

func (h *Hash) SetFromHex(value string) (*Hash, error) {
	value = strings.ReplaceAll(value, "0x", "")
	b, err := hex.DecodeString(value)
	if err != nil {
		return h, err
	}

	h.hex = value
	h.bytes = b
	h.big = new(big.Int).SetBytes(b)

	return h, nil
}

/* Number type */

type Number struct {
	value   uint64
	hex     string
	big     *big.Int
	bytesLE []byte
	bytesBE []byte
}

func (n *Number) UnmarshalJSON(data []byte) error {
	var value interface{}
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}

	switch v := value.(type) {
	case string:
		_, err = n.SetFromHex(v)
	case float64:
		var preciseValue uint64
		err = json.Unmarshal(data, &preciseValue)
		if err != nil {
			return err
		}
		n.SetFromValue(preciseValue)
	default:
		err = ErrUnknownInputType
	}

	return err
}

func (n *Number) Value() uint64       { return n.value }
func (n *Number) Hex() string         { return n.hex }
func (n *Number) PrefixedHex() string { return "0x" + n.hex }
func (n *Number) Big() *big.Int       { return n.big }
func (n *Number) BytesLE() []byte     { return n.bytesLE }
func (n *Number) BytesBE() []byte     { return n.bytesBE }

func (n *Number) SetFromValue(value uint64) *Number {
	n.value = value
	n.hex = strconv.FormatUint(value, 16)
	n.big = new(big.Int).SetUint64(value)

	n.bytesLE = make([]byte, 8)
	binary.LittleEndian.PutUint64(n.bytesLE, value)
	n.bytesBE = make([]byte, 8)
	binary.BigEndian.PutUint64(n.bytesBE, value)

	return n
}

func (n *Number) SetFromHex(value string) (*Number, error) {
	value = strings.ReplaceAll(value, "0x", "")
	num, err := strconv.ParseUint(value, 16, 64)
	if err != nil {
		return n, err
	}

	n.value = num
	n.hex = value
	n.big = new(big.Int).SetUint64(num)

	n.bytesLE = make([]byte, 8)
	binary.LittleEndian.PutUint64(n.bytesLE, num)
	n.bytesBE = make([]byte, 8)
	binary.BigEndian.PutUint64(n.bytesBE, num)

	return n, nil
}

func (n *Number) SetFromString(value string) (*Number, error) {
	num, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return n, err
	}

	n.value = num
	n.hex = value
	n.big = new(big.Int).SetUint64(num)

	n.bytesLE = make([]byte, 8)
	binary.LittleEndian.PutUint64(n.bytesLE, num)
	n.bytesBE = make([]byte, 8)
	binary.BigEndian.PutUint64(n.bytesBE, num)

	return n, nil
}

// note that the uint64 will not fit if the value is greater than a uint64,
// even though we allow the full precision to be stored in the big int
func (n *Number) SetFromBytes(value []byte) *Number {
	n.hex = fmt.Sprintf("%0x", value)
	n.big = new(big.Int).SetBytes(value)
	n.value = n.big.Uint64()

	n.bytesLE = make([]byte, len(value))
	copy(n.bytesLE, value)

	n.bytesBE = make([]byte, len(value))
	copy(n.bytesBE, value)
	for i, j := 0, len(n.bytesBE)-1; i < j; i, j = i+1, j-1 {
		n.bytesBE[i], n.bytesBE[j] = n.bytesBE[j], n.bytesBE[i]
	}

	return n
}

/* Solution type */

type Solution struct {
	size int
	data []uint64
	hex  string
}

func (s *Solution) UnmarshalJSON(data []byte) error {
	var value interface{}
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}

	switch v := value.(type) {
	case string:
		_, err = s.SetFromHex(v)
	case []interface{}:
		var preciseValue []uint64
		if len(v) == 0 {
			return fmt.Errorf("empty interface slice")
		} else if _, ok := v[0].(string); ok {
			preciseValue = make([]uint64, len(v))
			for i, raw := range v {
				strValue, ok := raw.(string)
				if !ok {
					return fmt.Errorf("invalid string value")
				} else if len(strValue) > 2 && strValue[:2] == "0x" {
					strValue = strValue[2:]
				}

				preciseValue[i], err = strconv.ParseUint(strValue, 16, 64)
				if err != nil {
					return err
				}
			}
		} else if _, ok := v[0].(float64); ok {
			err = json.Unmarshal(data, &preciseValue)
			if err != nil {
				return err
			}
		}
		s.SetFromData(preciseValue)
	default:
		err = ErrUnknownInputType
	}

	return err
}

func (s *Solution) Size() int           { return s.size }
func (s *Solution) Data() []uint64      { return s.data }
func (s *Solution) Hex() string         { return s.hex }
func (s *Solution) PrefixedHex() string { return "0x" + s.hex }

func (s *Solution) SetFromData(data []uint64) *Solution {
	s.size = len(data)
	s.data = data

	for _, val := range s.data {
		s.hex += fmt.Sprintf("%08x", val)
	}

	return s
}

func (s *Solution) SetFromHex(raw string) (*Solution, error) {
	raw = strings.ReplaceAll(raw, "0x", "")
	if len(raw) < 8 || len(raw)%8 != 0 {
		return nil, fmt.Errorf("invalid solution %s", raw)
	}

	s.size = len(raw) / 8
	s.data = make([]uint64, s.size)
	for i := range s.data {
		var err error
		s.data[i], err = strconv.ParseUint(raw[i*8:(i+1)*8], 16, 64)
		if err != nil {
			return nil, err
		}
	}
	s.hex = raw

	return s, nil
}

/* Difficulty type */

type Difficulty struct {
	value     uint64
	targetHex string
	targetBig *big.Int
	bits      uint32
}

func (d *Difficulty) Value() uint64             { return d.value }
func (d *Difficulty) TargetHex() string         { return d.targetHex }
func (d *Difficulty) TargetPrefixedHex() string { return "0x" + d.targetHex }
func (d *Difficulty) TargetBig() *big.Int       { return d.targetBig }
func (d *Difficulty) Bits() uint32              { return d.bits }

func (d *Difficulty) SetFromBig(targetBig *big.Int, maxDiff *big.Int) *Difficulty {
	targetHex := fmt.Sprintf("%064x", targetBig)
	valueBig := new(big.Int).Div(maxDiff, targetBig)

	d.value = valueBig.Uint64()
	d.targetHex = targetHex
	d.targetBig = targetBig
	d.bits = bigToCompact(targetBig)

	return d
}

func (d *Difficulty) SetFromValue(value uint64, maxDiff *big.Int) *Difficulty {
	valueBig := new(big.Int).SetUint64(value)
	targetBig := new(big.Int).Div(maxDiff, valueBig)
	targetHex := fmt.Sprintf("%064x", targetBig)

	d.value = value
	d.targetHex = targetHex
	d.targetBig = targetBig
	d.bits = bigToCompact(targetBig)

	return d
}

func (d *Difficulty) SetFromBits(bits uint32, maxDiff *big.Int) *Difficulty {
	targetBig := compactToBig(bits)
	targetHex := fmt.Sprintf("%064x", targetBig)
	valueBig := new(big.Int).Div(maxDiff, targetBig)

	d.value = valueBig.Uint64()
	d.targetHex = targetHex
	d.targetBig = targetBig
	d.bits = bits

	return d
}

/* Difficulty utility functions */
/* taken from github.com/btcsuite/btcd */

func compactToBig(compact uint32) *big.Int {
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}

func bigToCompact(n *big.Int) uint32 {
	if n.Sign() == 0 {
		return 0
	}

	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}

	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}

	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}

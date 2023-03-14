// Copyright © 2022 Igor Kroitor

package common

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

const (
	defaultPrecision  = 18
	digitsToBitsRatio = math.Ln10 / math.Ln2
)

var (
	big0      = new(big.Int)
	big1      = new(big.Int).SetUint64(1)
	big2      = new(big.Int).SetUint64(2)
	big10     = new(big.Int).SetUint64(10)
	bigMinus1 = new(big.Int).SetInt64(-1)

	prec2 = &Precise{Value: new(big.Int).SetUint64(2), Decimals: 0}
)

type Precise struct {
	Value    *big.Int
	Decimals int
}

func NewPrecise(number string) (*Precise, error) {
	number = strings.ToLower(number)

	var decimals int
	var modifier int64
	var err error
	if strings.Index(number, "e") != -1 {
		parts := strings.Split(number, "e")
		number = parts[0]
		modifier, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, err
		}
	}

	idx := strings.Index(number, ".")
	if idx != -1 {
		decimals = len(number) - idx - 1
	}
	decimals -= int(modifier)

	number = strings.ReplaceAll(number, ".", "")
	value, ok := new(big.Int).SetString(number, 10)
	if !ok {
		return nil, fmt.Errorf("unable to set string as *big.Int: %s", number)
	}

	precise := &Precise{
		Value:    value,
		Decimals: decimals,
	}

	return precise, nil
}

func newPrecisePair(a, b string) (*Precise, *Precise, error) {
	aPrec, err := NewPrecise(a)
	if err != nil {
		return nil, nil, err
	}

	bPrec, err := NewPrecise(b)

	return aPrec, bPrec, err
}

func (a *Precise) reduce() {
	strValue := a.Value.String()
	start := len(strValue) - 1
	if start == 0 {
		if strValue == "0" {
			a.Decimals = 0
		}
		return
	}

	var i int
	for i = start; i >= 0; i-- {
		if strValue[i] != '0' {
			break
		}
	}

	diff := start - i
	if diff > 0 {
		a.Decimals -= diff
		a.Value, _ = new(big.Int).SetString(strValue[:i+1], 10)
	}
}

func (a *Precise) String() string {
	a.reduce()

	var sign string
	value := new(big.Int).Set(a.Value)
	if value.Cmp(big0) < 0 {
		sign = "-"
		value.Mul(value, bigMinus1)
	}

	strValue := rightJustify(value.String(), "0", a.Decimals)

	var item string
	idx := len(strValue) - a.Decimals
	if idx == 0 {
		return sign + "0." + strValue
	} else if a.Decimals < 0 {
		for i := 0; i < -a.Decimals; i++ {
			item += "0"
		}
	} else if a.Decimals != 0 {
		item = "."
	}

	if idx > len(strValue) {
		return sign + strValue + item
	}

	return sign + strValue[:idx] + item + strValue[idx:]
}

func (a *Precise) GreaterThan(b *Precise) bool {
	return a.Sub(b).Value.Cmp(big0) > 0
}

func (a *Precise) GreaterThanEqual(b *Precise) bool {
	return a.Sub(b).Value.Cmp(big0) >= 0
}

func (a *Precise) LessThan(b *Precise) bool {
	return b.GreaterThan(a)
}

func (a *Precise) LessThanEqual(b *Precise) bool {
	return b.GreaterThanEqual(a)
}

func (a *Precise) Abs() *Precise {
	if a.Value.Cmp(big0) < 0 {
		return a.Neg()
	}

	return &Precise{Value: new(big.Int).Set(a.Value), Decimals: a.Decimals}
}

func (a *Precise) Sign() int {
	return a.Value.Cmp(big0)
}

func (a *Precise) Neg() *Precise {
	return &Precise{Value: new(big.Int).Mul(a.Value, bigMinus1), Decimals: a.Decimals}
}

func (a *Precise) AdjustedDigits() int {
	switch a.Value.Cmp(big0) {
	case -1:
		return len(a.Value.String()) - a.Decimals - 2
	case 1:
		return len(a.Value.String()) - a.Decimals - 1
	default:
		return 0
	}
}

func (a *Precise) Add(b *Precise) *Precise {
	if a.Decimals == b.Decimals {
		return &Precise{Value: new(big.Int).Add(a.Value, b.Value), Decimals: a.Decimals}
	}

	var smallerValue, biggerValue *big.Int
	var biggerDecimals, diff int
	if a.Decimals > b.Decimals {
		smallerValue, biggerValue, biggerDecimals = b.Value, a.Value, a.Decimals
		diff = a.Decimals - b.Decimals
	} else {
		smallerValue, biggerValue, biggerDecimals = a.Value, b.Value, b.Decimals
		diff = b.Decimals - a.Decimals
	}

	exponent := powerOf10Big(diff)
	normalized := exponent.Mul(exponent, smallerValue)

	return &Precise{Value: normalized.Add(normalized, biggerValue), Decimals: biggerDecimals}
}

func (a *Precise) Sub(b *Precise) *Precise {
	return a.Add(b.Neg())
}

func (a *Precise) Mul(b *Precise) *Precise {
	return &Precise{Value: new(big.Int).Mul(a.Value, b.Value), Decimals: a.Decimals + b.Decimals}
}

func (a *Precise) div(b *Precise, precision int) *Precise {
	distance := precision - a.Decimals + b.Decimals
	var numerator *big.Int
	if distance == 0 {
		numerator = new(big.Int).Set(a.Value)
	} else if distance < 0 {
		numerator = new(big.Int).Quo(a.Value, powerOf10Big(-distance))
	} else {
		numerator = new(big.Int).Mul(a.Value, powerOf10Big(distance))
	}

	return &Precise{Value: new(big.Int).Quo(numerator, b.Value), Decimals: precision}
}

func (a *Precise) Div(b *Precise) *Precise {
	return a.div(b, defaultPrecision)
}

func (a *Precise) DivWithPrecision(b *Precise, precision int) *Precise {
	return a.div(b, precision)
}

func (a *Precise) Quantize(precision int, mode RoundingMode) *Precise {
	diff := a.Decimals - precision
	if diff < 0 {
		return &Precise{Value: new(big.Int).Set(a.Value), Decimals: a.Decimals}
	}

	exponent := powerOf10Big(diff)
	rem := new(big.Int).Mod(a.Value, exponent)
	adjusted := new(big.Int).Sub(a.Value, rem)
	adjusted.Div(adjusted, exponent)

	if mode == Round && rem.Cmp(big0) > 0 {
		roundingCutoff := new(big.Int).Quo(exponent, big2)
		if rem.Cmp(roundingCutoff) >= 0 {
			adjusted.Add(adjusted, big1)
		}
	}

	return &Precise{Value: adjusted, Decimals: precision}
}

func (a *Precise) Mod(b *Precise) *Precise {
	var ratNumerator int
	if diff := b.Decimals - a.Decimals; diff > 0 {
		ratNumerator = diff
	}
	numerator := new(big.Int).Mul(a.Value, powerOf10Big(ratNumerator))

	var ratDenominator int
	if diff := a.Decimals - b.Decimals; diff > 0 {
		ratDenominator = diff
	}
	denominator := new(big.Int).Mul(b.Value, powerOf10Big(ratDenominator))

	mod := new(big.Int).Mod(numerator, denominator)

	return &Precise{Value: mod, Decimals: int(ratDenominator) + b.Decimals}
}

func PreciseStringAdd(a, b string) (string, error) {
	precA, precB, err := newPrecisePair(a, b)
	if err != nil {
		return "", err
	}

	return precA.Add(precB).String(), nil
}

func PreciseStringSub(a, b string) (string, error) {
	precA, precB, err := newPrecisePair(a, b)
	if err != nil {
		return "", err
	}

	return precA.Sub(precB).String(), nil
}

func PreciseStringMul(a, b string) (string, error) {
	precA, precB, err := newPrecisePair(a, b)
	if err != nil {
		return "", err
	}

	return precA.Mul(precB).String(), nil
}

func PreciseStringDiv(a, b string) (string, error) {
	precA, precB, err := newPrecisePair(a, b)
	if err != nil {
		return "", err
	}

	return precA.Div(precB).String(), nil
}

func PreciseStringDivWithPrecision(a, b string, precision int) (string, error) {
	precA, precB, err := newPrecisePair(a, b)
	if err != nil {
		return "", err
	}

	return precA.DivWithPrecision(precB, precision).String(), nil
}

func PreciseStringMod(a, b string) (string, error) {
	precA, precB, err := newPrecisePair(a, b)
	if err != nil {
		return "", err
	}

	return precA.Mod(precB).String(), nil
}

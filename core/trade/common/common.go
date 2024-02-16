// Copyright © 2022 Igor Kroitor

package common

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

const (
	maxPrecision = 28
)

var (
	re = regexp.MustCompile(`0+$`)
)

func rightJustify(inp, fill string, n int) string {
	diff := n - len(inp)
	if diff <= 0 {
		return inp
	}

	return strings.Repeat(fill, diff/len(fill)) + inp
}

func leftJustify(inp, fill string, n int) string {
	diff := n - len(inp)
	if diff <= 0 {
		return inp
	}

	return inp + strings.Repeat(fill, diff/len(fill))
}

func powerOf10Big(n int) *big.Int {
	return new(big.Int).Exp(big10, new(big.Int).SetInt64(int64(n)), nil)
}

type RoundingMode int

const (
	Truncate RoundingMode = iota
	Round
)

type DigitsMode int

const (
	DecimalPlaces DigitsMode = iota
	SignificantDigits
	TickSize
)

type PaddingMode int

const (
	NoPadding PaddingMode = iota
	PadWithZero
)

func decimalToPrecision(
	number *Precise,
	precision int,
	rounding RoundingMode,
	digits DigitsMode,
) (*Precise, error) {
	if precision < 0 {
		exponent := &Precise{Value: powerOf10Big(-precision), Decimals: 0}
		switch rounding {
		case Round:
			shortened := number.Div(exponent).Abs()
			rounded, err := decimalToPrecision(shortened, 0, rounding, DecimalPlaces)
			if err != nil {
				return nil, err
			} else if number.Sign() < 0 {
				rounded = rounded.Neg()
			}

			return rounded.Mul(exponent), nil
		case Truncate:
			rem := number.Abs().Mod(exponent)
			if number.Sign() < 0 {
				rem = rem.Neg()
			}

			truncated := number.Sub(rem)

			return decimalToPrecision(truncated, 0, rounding, DecimalPlaces)
		}
	}

	strNumber := number.String()
	switch rounding {
	case Round:
		switch digits {
		case DecimalPlaces:
			return number.Quantize(precision, rounding), nil
		case SignificantDigits:
			q := precision - number.AdjustedDigits() - 1
			if q < 0 {
				stringToPrecision := strNumber[:precision]
				if stringToPrecision == "" {
					stringToPrecision = "0"
				}

				smaller, err := NewPrecise(stringToPrecision)
				if err != nil {
					return nil, err
				}

				sigfig, err := NewPrecise(powerOf10Big(-q).String())
				if err != nil {
					return nil, err
				}

				below := sigfig.Mul(smaller)
				belowDiff := below.Sub(number).Abs()
				above := below.Add(sigfig)
				aboveDiff := above.Sub(number).Abs()

				if belowDiff.LessThan(aboveDiff) {
					return below, nil
				}

				return above, nil
			}

			return number.Quantize(q, rounding), nil
		}
	}

	return number, nil
}

func formatPrecise(
	number *Precise,
	precision int,
	rounding RoundingMode,
	digits DigitsMode,
) string {
	precise := number.String()
	switch rounding {
	case Round:
		if precise == ("-0." + strings.Repeat("0", len(precise)))[:2] || precise == "-0" {
			precise = precise[1:]
		}
	case Truncate:
		switch digits {
		case DecimalPlaces:
			var before, after string
			if idx := strings.Index(precise, "."); idx != -1 {
				before, after = precise[:idx], precise[idx+1:]
			} else {
				before = precise
			}

			endIdx := precision
			afterLen := len(after)
			if afterLen == 0 {
				endIdx = 0
			} else if endIdx < 0 {
				endIdx = afterLen + 1 + endIdx
				if endIdx < 0 {
					endIdx = 0
				}
			} else if afterLen < endIdx {
				endIdx = afterLen
			}

			precise = before + "." + after[:endIdx]
		case SignificantDigits:
			if precision == 0 {
				return "0"
			} else if precise == "0" { // @TODO: need to fix this
				return "0"
			}

			dot := strings.Index(precise, ".")
			if dot == -1 {
				dot = len(precise)
			}

			start := dot - number.AdjustedDigits()
			end := start + precision
			if dot >= end {
				end--
			}

			if end < 0 {
				end += len(precise)
			}

			if precision < len(strings.ReplaceAll(precise, ".", "")) {
				if end >= 0 {
					precise = precise[:end]
				}

				precise = leftJustify(precise, "0", dot)
			}
		}

		if precise == ("-0." + strings.Repeat("0", len(precise)))[:3] || precise == "-0" {
			precise = precise[1:]
		}
		precise = strings.TrimRight(precise, ".")
	}

	return precise
}

func padPrecise(precise string, precision int, digits DigitsMode, padding PaddingMode) string {
	switch padding {
	case NoPadding:
		if strings.Index(precise, ".") != -1 {
			precise = strings.TrimRight(strings.TrimRight(precise, "0"), ".")
		}
	case PadWithZero:
		if idx := strings.Index(precise, "."); idx != -1 {
			switch digits {
			case DecimalPlaces:
				before, after := precise[:idx], precise[idx+1:]
				precise = before + "." + leftJustify(after, "0", precision)
			case SignificantDigits:
				var idx int
				for idx = 0; idx < len(precise); idx++ {
					switch precise[idx] {
					case '0', '.':
						continue
					}
					break
				}

				if strings.Index(precise[idx:], ".") != -1 {
					precision++
				}

				precise = precise[:idx] + leftJustify(strings.TrimRight(precise[idx:], "0"), "0", precision)
			}
		} else {
			switch digits {
			case DecimalPlaces:
				if precision > 0 {
					precise = precise + "." + strings.Repeat("0", precision)
				}
			case SignificantDigits:
				if diff := precision - len(precise); diff > 0 {
					precise = precise + "." + strings.Repeat("0", diff)
				}
			}
		}
	}

	return precise
}

func DecimalToPrecision(
	strNumber, strPrecision string,
	rounding RoundingMode,
	digits DigitsMode,
	padding PaddingMode,
) (string, error) {
	number, err := NewPrecise(strNumber)
	if err != nil {
		return "", err
	}

	var newPrecision int
	if digits == TickSize {
		precision, err := NewPrecise(strPrecision)
		if err != nil {
			return "", err
		} else if precision.Sign() != 1 {
			return "", fmt.Errorf("TickSize is not compatible with a negative precision")
		}

		missing := number.Abs().Mod(precision)
		if missing.Value.Cmp(big0) != 0 {
			switch rounding {
			case Round:
				if number.Sign() > 0 {
					if missing.GreaterThanEqual(precision.Div(prec2)) {
						number = number.Sub(missing).Add(precision)
					} else {
						number = number.Sub(missing)
					}
				} else {
					if missing.GreaterThanEqual(precision.Div(prec2)) {
						number = number.Add(missing).Sub(precision)
					} else {
						number = number.Add(missing)
					}
				}
			case Truncate:
				if number.Sign() < 0 {
					number = number.Add(missing)
				} else {
					number = number.Sub(missing)
				}
			}
		}

		rounding, digits = Round, DecimalPlaces
		parts := strings.Split(re.ReplaceAllLiteralString(precision.String(), ""), ".")
		if len(parts) > 1 {
			newPrecision = len(parts[1])
		} else {
			newPrecision = -len(re.FindString(parts[0]))
		}
	} else {
		var err error
		newPrecision, err = strconv.Atoi(strPrecision)
		if err != nil {
			return "", err
		} else if newPrecision < 0 {
			digits = DecimalPlaces
		}
	}

	precise, err := decimalToPrecision(number, newPrecision, rounding, digits)
	if err != nil {
		return "", err
	}

	formatted := formatPrecise(precise, newPrecision, rounding, digits)
	padded := padPrecise(formatted, newPrecision, digits, padding)

	return padded, nil
}

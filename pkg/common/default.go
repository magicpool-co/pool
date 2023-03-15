package common

import (
	"fmt"
	"math/big"
	"strings"
)

func GetDefaultUnitScale(val float64) (string, float64) {
	if val < 1_000.0 {
		return "", 1.0
	} else if val < 1_000_000.0 {
		return "K", 1_000.0
	} else if val < 1_000_000_000.0 {
		return "M", 1_000_000.0
	} else if val < 1_000_000_000_000.0 {
		return "G", 1_000_000_000.0
	} else if val < 1_000_000_000_000_000.0 {
		return "T", 1_000_000_000_000.0
	} else if val < 1_000_000_000_000_000_000.0 {
		return "P", 1_000_000_000_000_000.0
	} else if val < 1_000_000_000_000_000_000_000.0 {
		return "E", 1_000_000_000_000_000_000.0
	}

	return "", 1.0
}

func GetDefaultUnits(chain string) (*big.Int, error) {
	var units uint64
	switch strings.ToUpper(chain) {
	case "BTC", "FIRO", "FLUX", "KAS", "RVN":
		units = 1e8
	case "CFX", "CTXC", "ETC", "ETH":
		units = 1e18
	case "ERGO":
		units = 1e9
	case "USDC", "USDT", "BUSD":
		units = 1e6
	default:
		return nil, fmt.Errorf("unsupported chain %s for get units", chain)
	}

	return new(big.Int).SetUint64(units), nil
}

type PayoutBounds struct {
	Min       *big.Int
	Default   *big.Int
	Max       *big.Int
	Precision uint64
	Units     uint64
}

func (b *PayoutBounds) PrecisionMask() *big.Int {
	precision := new(big.Int).SetUint64(b.Units - b.Precision)
	mask := new(big.Int).Exp(Big10, precision, nil)

	return mask
}

func GetDefaultPayoutBounds(chain string) (*PayoutBounds, error) {
	var bounds *PayoutBounds
	switch strings.ToUpper(chain) {
	case "BTC":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("50000"),
			Default:   MustParseBigInt("10000000"),
			Max:       MustParseBigInt("200000000"),
			Precision: 4,
			Units:     8,
		}
	case "CFX":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("10000000000000000000"),
			Default:   MustParseBigInt("50000000000000000000"),
			Max:       MustParseBigInt("1000000000000000000000000"),
			Precision: 1,
			Units:     18,
		}
	case "CTXC":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("5000000000000000000"),
			Default:   MustParseBigInt("25000000000000000000"),
			Max:       MustParseBigInt("250000000000000000000000"),
			Precision: 1,
			Units:     18,
		}
	case "ERGO":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("100000000"),
			Default:   MustParseBigInt("5000000000"),
			Max:       MustParseBigInt("25000000000000"),
			Precision: 2,
			Units:     9,
		}
	case "ETC":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("50000000000000000"),
			Default:   MustParseBigInt("500000000000000000"),
			Max:       MustParseBigInt("2000000000000000000000"),
			Precision: 3,
			Units:     18,
		}
	case "ETH":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("5000000000000000"),
			Default:   MustParseBigInt("100000000000000000"),
			Max:       MustParseBigInt("25000000000000000000"),
			Precision: 4,
			Units:     18,
		}
	case "ETHW":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("1000000000000000000"),
			Default:   MustParseBigInt("5000000000000000000"),
			Max:       MustParseBigInt("20000000000000000000000"),
			Precision: 3,
			Units:     18,
		}
	case "FIRO":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("10000000"),
			Default:   MustParseBigInt("500000000"),
			Max:       MustParseBigInt("2500000000000"),
			Precision: 2,
			Units:     8,
		}
	case "FLUX":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("10000000"),
			Default:   MustParseBigInt("100000000"),
			Max:       MustParseBigInt("5000000000000"),
			Precision: 2,
			Units:     8,
		}
	case "KAS":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("10000000000"),
			Default:   MustParseBigInt("100000000000"),
			Max:       MustParseBigInt("1000000000000000"),
			Precision: 1,
			Units:     8,
		}
	case "RVN":
		bounds = &PayoutBounds{
			Min:       MustParseBigInt("1000000000"),
			Default:   MustParseBigInt("20000000000"),
			Max:       MustParseBigInt("200000000000000"),
			Precision: 1,
			Units:     8,
		}
	default:
		return nil, fmt.Errorf("unsupported chain %s for get payout bounds", chain)
	}

	return bounds, nil
}

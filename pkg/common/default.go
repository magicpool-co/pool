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
	case "BTC", "FIRO", "FLUX", "RVN":
		units = 1e8
	case "CFX", "CTXC", "ETC", "ETH":
		units = 1e18
	case "ERGO":
		units = 1e9
	case "USDC", "USDT":
		units = 1e6
	default:
		return nil, fmt.Errorf("unsupported chain %s for get units", chain)
	}

	return new(big.Int).SetUint64(units), nil
}

func GetDefaultPayoutThreshold(chain string) (*big.Int, error) {
	var threshold *big.Int
	switch strings.ToUpper(chain) {
	case "BTC":
		threshold = MustParseBigInt("10000000")
	case "CFX":
		threshold = MustParseBigInt("1000000000000000000")
	case "CTXC":
		threshold = MustParseBigInt("1000000000000000000000")
	case "ERGO":
		threshold = MustParseBigInt("5000000000")
	case "ETC":
		threshold = MustParseBigInt("500000000000000000")
	case "ETH":
		threshold = MustParseBigInt("100000000000000000")
	case "ETHW":
		threshold = MustParseBigInt("5000000000000000000")
	case "FIRO":
		threshold = MustParseBigInt("3000000000")
	case "FLUX":
		threshold = MustParseBigInt("1000000000")
	case "RVN":
		threshold = MustParseBigInt("250000000000")
	case "USDC":
		threshold = MustParseBigInt("100000000")
	default:
		return nil, fmt.Errorf("unsupported chain %s for get threshold", chain)
	}

	return threshold, nil
}

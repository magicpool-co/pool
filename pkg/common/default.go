package common

import (
	"fmt"
	"math/big"
	"strings"
)

func GetDefaultUnits(chain string) (*big.Int, error) {
	var units uint64
	switch strings.ToUpper(chain) {
	case "BTC", "FIRO", "FLUX", "RVN":
		units = 1e8
	case "CFX", "CTXC", "ETC", "ETH":
		units = 1e18
	case "ERGO":
		units = 1e9
	case "USDC":
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
		threshold = MustParseBigInt("2000000000000000000000")
	case "CTXC":
		threshold = MustParseBigInt("1000000000000000000000")
	case "ERGO":
		threshold = MustParseBigInt("100000000000")
	case "ETC":
		threshold = MustParseBigInt("5000000000000000000")
	case "ETH":
		threshold = MustParseBigInt("100000000000000000")
	case "FIRO":
		threshold = MustParseBigInt("3000000000")
	case "FLUX":
		threshold = MustParseBigInt("10000000000")
	case "RVN":
		threshold = MustParseBigInt("250000000000")
	case "USDC":
		threshold = MustParseBigInt("100000000")
	default:
		return nil, fmt.Errorf("unsupported chain %s for get threshold", chain)
	}

	return threshold, nil
}
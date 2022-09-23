package common

import (
	"fmt"
	"math/big"
	"strings"
)

func GetDefaultUnits(chain string) (*big.Int, error) {
	var units uint64
	switch strings.ToUpper(chain) {
	case "USDC":
		units = 1e6
	case "BTC", "FIRO", "FLUX", "RVN":
		units = 1e8
	case "ERGO":
		units = 1e9
	case "CFX", "CTXC", "ETC", "ETH":
		units = 1e18
	default:
		return nil, fmt.Errorf("unsupported chain %s for get units", chain)
	}

	return new(big.Int).SetUint64(units), nil
}

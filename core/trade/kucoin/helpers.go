package kucoin

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/magicpool-co/pool/pkg/common"
)

func formatChain(chain string) string {
	chain = strings.ToUpper(chain)
	switch chain {
	case "ERGO":
		return "ERG"
	default:
		return chain
	}
}

func unformatChain(chain string) string {
	chain = strings.ToUpper(chain)
	switch chain {
	case "ERG":
		return "ERGO"
	case "ERC20":
		return "ETH"
	default:
		return chain
	}
}

func chainToNetwork(chain string) string {
	network := strings.ToLower(chain)
	switch network {
	default:
		return network
	}
}

func parseIncrement(increment string) (int, error) {
	switch increment {
	case "0.1":
		return 1e1, nil
	case "0.01":
		return 1e2, nil
	case "0.001":
		return 1e3, nil
	case "0.0001":
		return 1e4, nil
	case "0.00001":
		return 1e5, nil
	case "0.000001":
		return 1e6, nil
	case "0.0000001":
		return 1e7, nil
	case "0.00000001":
		return 1e8, nil
	default:
		return 1, fmt.Errorf("unknown increment %s", increment)
	}
}

func calcFeeAsBig(value, feeRate string, units *big.Int) (*big.Int, error) {
	feeParts := strings.Split(feeRate, ".")
	if len(feeParts) != 2 {
		return nil, fmt.Errorf("invalid fee rate %s", feeRate)
	}

	feeDec := feeParts[1]
	num := len(feeDec)
	mult, err := strconv.ParseUint(strings.TrimLeft(feeDec, "0"), 10, 64)
	if err != nil {
		return nil, err
	}

	var moved int
	if num <= 0 || value == "0" {
		return new(big.Int), nil
	}

	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid float %s", value)
	}

	newInt := parts[0]
	newDec := ""
	if newInt != "0" {
		for i := len(newInt) - 1; i >= 0; i-- {
			newDec = newInt[i:] + newDec
			newInt = newInt[:i]

			moved++
			if moved == num {
				if newInt == "" {
					newInt = "0"
				}
				break
			}
		}
	}

	for i := 0; i < num-moved; i++ {
		newDec = "0" + newDec
	}

	feeStr := newInt + "." + newDec + parts[1]

	feeBig, err := common.StringDecimalToBigint(feeStr, units)
	if err != nil {
		return nil, err
	}

	feeBig.Mul(feeBig, new(big.Int).SetUint64(mult))

	return feeBig, nil
}

func safeSubtractFee(chain, quantity, feeRate string) (float64, float64, error) {
	var precision int
	switch chain {
	case "BTC", "ETH":
		precision = 1e6
	case "USDC", "USDT":
		precision = 1e4
	default:
		return 0, 0, fmt.Errorf("unsupported chain to split fees on %s", chain)
	}

	// max precision *should be 1e8, add 3 for 1e3 fee
	units := new(big.Int).SetUint64(1e11)

	// convert quantity into safe big int
	quantityBig, err := common.StringDecimalToBigint(quantity, units)
	if err != nil {
		return 0, 0, err
	}

	// calculate fees safely as big int
	feeBig, err := calcFeeAsBig(quantity, feeRate, units)
	if err != nil {
		return 0, 0, err
	}

	// subtract fees from quantity
	quantityBig.Sub(quantityBig, feeBig)

	// safely convert quantity back to float
	quantityFloat := common.BigIntToFloat64(quantityBig, units)

	// round quantity down to defined precision
	quantityFloat = common.FloorFloatByIncrement(quantityFloat, precision, 1e8)

	// calculate the fees (ignoring floating point errors)
	initialFloat, err := strconv.ParseFloat(quantity, 64)
	if err != nil {
		return 0, 0, err
	}
	feesFloat := initialFloat - quantityFloat

	return quantityFloat, feesFloat, nil
}

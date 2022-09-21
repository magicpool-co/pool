package accounting

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
)

/*
there are three steps to an exchange batch:
	- calculate the trades based off of the sum input amouts, (static) input thresholds,
	estimated output amounts (based off of market rate for each individual input->output pair),
	and exchange specific output thresholds
	- deposit, generate and execute the trades, and withdrawal
	- discount the cumulative exchange fees and credit the withdrawal to all miners
*/

var (
	units = map[string]*big.Int{
		"BTC":  new(big.Int).SetUint64(1e8),
		"CFX":  new(big.Int).SetUint64(1e18),
		"CTXC": new(big.Int).SetUint64(1e18),
		"ERGO": new(big.Int).SetUint64(1e9),
		"ETC":  new(big.Int).SetUint64(1e18),
		"ETH":  new(big.Int).SetUint64(1e18),
		"FIRO": new(big.Int).SetUint64(1e8),
		"FLUX": new(big.Int).SetUint64(1e8),
		"RVN":  new(big.Int).SetUint64(1e8),
		"USDC": new(big.Int).SetUint64(1e6),
	}

	inputThresholds = map[string]*big.Int{
		"CFX":  common.MustParseBigInt("2000000000000000000000"),  // 2,000 CFX
		"CTXC": common.MustParseBigInt("500000000000000000000"),   // 500 CTXC
		"ERGO": new(big.Int).SetUint64(10_000_000_000),            // 10 ERGO
		"ETC":  new(big.Int).SetUint64(3_000_000_000_000_000_000), // 3 ETC
		"FIRO": new(big.Int).SetUint64(10_000_000_000),            // 100 FIRO
		"FLUX": new(big.Int).SetUint64(30_000_000_000),            // 300 FLUX
		"RVN":  new(big.Int).SetUint64(20_000_000_000),            // 200 RVN
	}

	outputThresholds = map[string]*big.Int{
		"BTC":  new(big.Int).SetUint64(50_000_000),                // 0.5 BTC
		"ETH":  new(big.Int).SetUint64(5_000_000_000_000_000_000), // 5 ETH
		"USDC": new(big.Int).SetUint64(20_000_000_000),            // 20,000 USDC
	}
)

func reverseMap(input map[string]map[string]*big.Int, prices map[string]map[string]float64) (map[string]map[string]*big.Int, error) {
	output := make(map[string]map[string]*big.Int)
	for from, toIdx := range input {
		for to, value := range toIdx {
			if _, ok := output[to]; !ok {
				output[to] = make(map[string]*big.Int)
			}
			if _, ok := output[to][from]; !ok {
				output[to][from] = new(big.Int)
			}

			reverseValue := new(big.Int).Set(value)
			if _, ok := prices[from]; ok {
				price, ok := prices[from][to]
				if ok {
					fromUnits, ok := units[from]
					if !ok {
						return nil, fmt.Errorf("no units for from chain %s", from)
					}

					toUnits, ok := units[to]
					if !ok {
						return nil, fmt.Errorf("no units for to chain %s", to)
					}

					rate, err := common.StringDecimalToBigint(fmt.Sprintf("%.8f", price), toUnits)
					if err != nil {
						return nil, err
					}

					reverseValue.Mul(reverseValue, rate)
					reverseValue.Div(reverseValue, fromUnits)
				}
			}

			output[to][from].Add(output[to][from], reverseValue)
		}
	}

	return output, nil
}

func sumMap(input map[string]map[string]*big.Int) map[string]*big.Int {
	output := make(map[string]*big.Int)
	for from, toIdx := range input {
		for _, value := range toIdx {
			if _, ok := output[from]; !ok {
				output[from] = new(big.Int)
			}

			output[from].Add(output[from], value)
		}
	}

	return output
}

func CalculateExchangePaths(inputPaths map[string]map[string]*big.Int, outputThresholds map[string]*big.Int, prices map[string]map[string]float64) (map[string]map[string]*big.Int, error) {
	for {
		var hasChanges bool
		if len(inputPaths) == 0 {
			break
		}

		outputPaths, err := reverseMap(inputPaths, prices)
		if err != nil {
			return nil, err
		}

		inputSum := sumMap(inputPaths)
		outputSum := sumMap(outputPaths)

		for input, value := range inputSum {
			threshold, ok := inputThresholds[input]
			if !ok {
				return nil, fmt.Errorf("no input threshold for %s", input)
			} else if value.Cmp(threshold) < 0 {
				hasChanges = true
				delete(inputPaths, input)
			}
		}

		for output, value := range outputSum {
			threshold, ok := outputThresholds[output]
			if !ok {
				return nil, fmt.Errorf("no output threshold for %s", output)
			} else if value.Cmp(threshold) < 0 {
				hasChanges = true
				for _, outputIdx := range inputPaths {
					delete(outputIdx, output)
				}
			}
		}

		if !hasChanges {
			break
		}
	}

	for input, outputIdx := range inputPaths {
		if len(outputIdx) == 0 {
			delete(inputPaths, input)
		}
	}

	return inputPaths, nil
}

package accounting

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
)

var (
	DefaultInputThresholds = map[string]*big.Int{
		"CFX":  common.MustParseBigInt("2000000000000000000000"),  // 2,000 CFX
		"CTXC": common.MustParseBigInt("500000000000000000000"),   // 500 CTXC
		"ERGO": new(big.Int).SetUint64(10_000_000_000),            // 10 ERGO
		"ETC":  new(big.Int).SetUint64(3_000_000_000_000_000_000), // 3 ETC
		"FIRO": new(big.Int).SetUint64(10_000_000_000),            // 100 FIRO
		"FLUX": new(big.Int).SetUint64(30_000_000_000),            // 300 FLUX
		"RVN":  new(big.Int).SetUint64(20_000_000_000),            // 200 RVN
	}

	DefaultOutputThresholds = map[string]*big.Int{
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

			// set the reverse value as initial value (to avoid accidental overwriting)
			reverseValue := new(big.Int).Set(value)

			// if prices exists, adjust the value based off of the price
			if _, ok := prices[from]; ok {
				price, ok := prices[from][to]
				if ok {
					fromUnits, err := common.GetDefaultUnits(from)
					if err != nil {
						return nil, err
					}

					toUnits, err := common.GetDefaultUnits(to)
					if err != nil {
						return nil, err
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
		output[from] = new(big.Int)
		for _, value := range toIdx {
			output[from].Add(output[from], value)
		}
	}

	return output
}

func CalculateExchangePaths(inputPaths map[string]map[string]*big.Int, inputThresholds, outputThresholds map[string]*big.Int, prices map[string]map[string]float64) (map[string]map[string]*big.Int, error) {
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

func CalculateProportionalValues[K string | uint64](value, fee *big.Int, proportions map[K]*big.Int) (map[K]*big.Int, map[K]*big.Int, error) {
	if value.Cmp(common.Big0) <= 0 {
		return nil, nil, fmt.Errorf("input value is not positive")
	}

	// keep track of the used value and fees to make sure any remainders are distributed
	usedValue := new(big.Int)
	usedFee := new(big.Int)

	// calculate the sum input value
	sumInputValue := new(big.Int)
	for _, inputValue := range proportions {
		sumInputValue.Add(sumInputValue, inputValue)
	}

	// protect against divide by zero
	if sumInputValue.Cmp(common.Big0) <= 0 {
		return nil, nil, fmt.Errorf("sum input value is not positive")
	}

	// iterate through the proportions, calculating the proportional
	// value and fee for each item in the map
	proportionalValues := make(map[K]*big.Int)
	proportionalFees := make(map[K]*big.Int)
	for k, inputValue := range proportions {
		// value * (inputValue / sumInputValue) is the formula for calculating
		// the proportional value based on the given input value
		proportionalValues[k] = new(big.Int).Mul(value, inputValue)
		proportionalValues[k].Div(proportionalValues[k], sumInputValue)
		usedValue.Add(usedValue, proportionalValues[k])

		// fee * (tradeValue / sumTradeValue) is the formula for calculating
		// the proportional fee based on the given input value.
		// fee * (proportionalValues[k] / value) would also work
		proportionalFees[k] = new(big.Int).Mul(fee, inputValue)
		proportionalFees[k].Div(proportionalFees[k], sumInputValue)
		usedFee.Add(usedFee, proportionalFees[k])
	}

	// calculate the remainder values and verify they are non-negative (over-credit protection)
	remainderValue := usedValue.Sub(value, usedValue)
	if remainderValue.Cmp(common.Big0) < 0 {
		return nil, nil, fmt.Errorf("negative remainder value for trade proportions")
	}

	// calculate the remainder fees and verify they are non-negative (over-credit protection)
	remainderFee := usedFee.Sub(fee, usedFee)
	if remainderFee.Cmp(common.Big0) < 0 {
		return nil, nil, fmt.Errorf("negative remainder fee for trade proportions")
	}

	// find the key with the largest value for deterministic remainder distribution
	// (if multiple values are equivalent, choose the smallest key)
	var smallestKey K
	largestValue := new(big.Int)
	for k, v := range proportionalValues {
		diff := v.Cmp(largestValue)
		if diff > 0 || (diff == 0 && k < smallestKey) {
			smallestKey = k
			largestValue.Set(v)
		}
	}

	if _, ok := proportionalValues[smallestKey]; !ok {
		return nil, nil, fmt.Errorf("key not found in proportionalValues: %v", smallestKey)
	} else if _, ok := proportionalFees[smallestKey]; !ok {
		return nil, nil, fmt.Errorf("key not found in proportionalFees: %v", smallestKey)
	}

	// distribute the remainder fees
	proportionalValues[smallestKey].Add(proportionalValues[smallestKey], remainderValue)
	proportionalFees[smallestKey].Add(proportionalFees[smallestKey], remainderFee)

	return proportionalValues, proportionalFees, nil
}

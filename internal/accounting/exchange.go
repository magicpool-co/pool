package accounting

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
)

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

func CalculateProportionalTradeValues(withdrawalValue, feeValue *big.Int, tradeProportions map[string]*big.Int) (map[string]*big.Int, map[string]*big.Int, error) {
	// keep track of the used value and fees to make sure any remainders are distributed
	usedValue := new(big.Int)
	usedFee := new(big.Int)

	// calculate the sum trade value (to use as the denominator for each trades's proportion)
	sumTradeValue := new(big.Int)
	for _, tradeValue := range tradeProportions {
		sumTradeValue.Add(sumTradeValue, tradeValue)
	}

	// iterate through the final values of each trade path (by the final chain)
	// and calculate the proportional withdrawal value and cumulative fees for each
	proportionalValues := make(map[string]*big.Int)
	proportionalFees := make(map[string]*big.Int)
	for initialChainID, tradeValue := range tradeProportions {
		// withdrawalValue * (tradeValue / sumTradeValue) is the formula
		// for calculating the distributed withdrawal values for that trade path
		proportionalValues[initialChainID] = new(big.Int).Mul(withdrawalValue, tradeValue)
		proportionalValues[initialChainID].Div(proportionalValues[initialChainID], sumTradeValue)
		usedValue.Add(usedValue, proportionalValues[initialChainID])

		// feeValue * (tradeValue / sumTradeValue) is the formula
		// for calculating the distributed withdrawal values for that trade path
		// feeValue * (proportionalValue / withdrawalValue) would also work
		proportionalFees[initialChainID] = new(big.Int).Mul(feeValue, tradeValue)
		proportionalFees[initialChainID].Div(proportionalFees[initialChainID], sumTradeValue)
		usedFee.Add(usedFee, proportionalFees[initialChainID])
	}

	// calculate the remainder values and verify they are non-negative (over-credit protection)
	remainderValue := usedValue.Sub(withdrawalValue, usedValue)
	if remainderValue.Cmp(common.Big0) < 0 {
		return nil, nil, fmt.Errorf("negative remainder value for trade proportions")
	}

	// calculate the remainder fees and verify they are non-negative (over-credit protection)
	remainderFee := usedFee.Sub(feeValue, usedFee)
	if remainderFee.Cmp(common.Big0) < 0 {
		return nil, nil, fmt.Errorf("negative remainder fee for trade proportions")
	}

	// calculate the output chain with the largest value for determinstic remainder distribution
	// (if two values are the same, use the chain with the smallest lexographical order)
	var smallestChainID string
	largestValue := new(big.Int)
	for chainID, proportionalValue := range proportionalValues {
		diff := proportionalValue.Cmp(largestValue)
		if diff > 0 || (diff == 0 && chainID < smallestChainID) {
			smallestChainID = chainID
			largestValue.Set(proportionalValue)
		}
	}

	// distribute the remainder fees
	proportionalValues[smallestChainID].Add(proportionalValues[smallestChainID], remainderValue)
	proportionalFees[smallestChainID].Add(proportionalFees[smallestChainID], remainderFee)

	return proportionalValues, proportionalFees, nil
}

func CalculateProportionalMinerValues(withdrawalValue, feeValue *big.Int, minerProportions map[uint64]*big.Int) (map[uint64]*big.Int, map[uint64]*big.Int, error) {
	// calculate the sum initial value (to use as the denominator for each miner's proportion)
	sumInitialValue := new(big.Int)
	for _, initialValue := range minerProportions {
		sumInitialValue.Add(sumInitialValue, initialValue)
	}

	// keep track of the used value and fees to make sure any remainders are distributed
	usedValue := new(big.Int)
	usedFee := new(big.Int)

	// iterate through the initial values for each miner (by the final chain)
	// and calculate the proportional withdrawal value and cumulative fees for each
	proportionalValues := make(map[uint64]*big.Int)
	proportionalFees := make(map[uint64]*big.Int)
	for minerID, initialValue := range minerProportions {
		// finalValue * (initialValue / sumInitialValue) is the formula
		// for calculating the distributed withdrawal values for that miner
		proportionalValues[minerID] = new(big.Int).Mul(withdrawalValue, initialValue)
		proportionalValues[minerID].Div(proportionalValues[minerID], sumInitialValue)
		usedValue.Add(usedValue, proportionalValues[minerID])

		// feeValue * (initialValue / sumInitialValue) is the formula
		// for calculating the distributed withdrawal values for that trade path
		// feeValue * (proportionalValue / withdrawalValue) would also work
		proportionalFees[minerID] = new(big.Int).Mul(feeValue, initialValue)
		proportionalFees[minerID].Div(proportionalFees[minerID], sumInitialValue)
		usedFee.Add(usedFee, proportionalFees[minerID])
	}

	// calculate the remainder values and verify they are non-negative (over-credit protection)
	remainderValue := usedValue.Sub(withdrawalValue, usedValue)
	if remainderValue.Cmp(common.Big0) < 0 {
		return nil, nil, fmt.Errorf("negative remainder value for miner proportions")
	}

	// calculate the remainder fees and verify they are non-negative (over-credit protection)
	remainderFee := usedFee.Sub(feeValue, usedFee)
	if remainderFee.Cmp(common.Big0) < 0 {
		return nil, nil, fmt.Errorf("negative remainder fee for miner proportions")
	}

	// calculate the output chain with the largest value for determinstic remainder distribution
	// (if two values are the same, use the miner with the lowest ID)
	var smallestMinerID uint64
	largestValue := new(big.Int)
	for minerID, proportionalValue := range proportionalValues {
		diff := proportionalValue.Cmp(largestValue)
		if diff > 0 || (diff == 0 && minerID < smallestMinerID) {
			smallestMinerID = minerID
			largestValue.Set(proportionalValue)
		}
	}

	// distribute the remainder fees
	proportionalValues[smallestMinerID].Add(proportionalValues[smallestMinerID], remainderValue)
	proportionalFees[smallestMinerID].Add(proportionalFees[smallestMinerID], remainderFee)

	return proportionalValues, proportionalFees, nil
}

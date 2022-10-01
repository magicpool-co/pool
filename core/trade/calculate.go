package trade

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

func balanceInputsToInputPaths(balanceInputs []*pooldb.BalanceInput) (map[string]map[string]*big.Int, error) {
	// iterate through each balance input, creating a trade path from the
	// given input chain to the desired, summing any overlapping paths (since
	// paths are batched and redistributed during withdrawal crediting)

	// note that "output path" just refers to the path being an "unconfirmed"
	// path, meaning it is still unknown whether or not it will actually be executed
	inputPaths := make(map[string]map[string]*big.Int)
	for _, balanceInput := range balanceInputs {
		if !balanceInput.Value.Valid {
			return nil, fmt.Errorf("no value for balance input %d", balanceInput.ID)
		}
		inChainID := balanceInput.ChainID
		outChainID := balanceInput.OutChainID
		value := balanceInput.Value.BigInt

		if _, ok := inputPaths[inChainID]; !ok {
			inputPaths[inChainID] = make(map[string]*big.Int)
		}
		if _, ok := inputPaths[inChainID][outChainID]; !ok {
			inputPaths[inChainID][outChainID] = new(big.Int)
		}

		inputPaths[inChainID][outChainID].Add(inputPaths[inChainID][outChainID], value)
	}

	return inputPaths, nil
}

func balanceInputsToInitialProportions(balanceInputs []*pooldb.BalanceInput) (map[string]map[string]map[uint64]*big.Int, error) {
	// iterate through each balance input, summing the owed amount for each
	// miner participating in the exchange batch

	// note that the in chain is the primary key and the
	// out chain is the secondary key
	proportions := make(map[string]map[string]map[uint64]*big.Int)
	for _, balanceInput := range balanceInputs {
		if !balanceInput.Value.Valid {
			return nil, fmt.Errorf("no value for balance input %d", balanceInput.ID)
		}
		inChainID := balanceInput.ChainID
		outChainID := balanceInput.OutChainID
		minerID := balanceInput.MinerID
		value := balanceInput.Value.BigInt

		if _, ok := proportions[outChainID]; !ok {
			proportions[outChainID] = make(map[string]map[uint64]*big.Int)
		}
		if _, ok := proportions[outChainID][inChainID]; !ok {
			proportions[outChainID][inChainID] = make(map[uint64]*big.Int)
		}
		if _, ok := proportions[outChainID][inChainID][minerID]; !ok {
			proportions[outChainID][inChainID][minerID] = new(big.Int)
		}

		proportions[outChainID][inChainID][minerID].Add(proportions[outChainID][inChainID][minerID], value)
	}

	return proportions, nil
}

func exchangeInputsToOutputPaths(exchangeInputs []*pooldb.ExchangeInput) (map[string]map[string]*big.Int, error) {
	// iterate through each exchange input, creating a trade path from the
	// given input chain to the desired (exchange inputs should already be summed)

	// note that "output path" just refers to the path being
	// a "confirmed" path that will be executed
	outputPaths := make(map[string]map[string]*big.Int)
	for _, exchangeInput := range exchangeInputs {
		if !exchangeInput.Value.Valid {
			return nil, fmt.Errorf("no value for exchange input %d", exchangeInput.ID)
		}
		inChainID := exchangeInput.InChainID
		outChainID := exchangeInput.OutChainID
		value := exchangeInput.Value.BigInt

		if _, ok := outputPaths[inChainID]; !ok {
			outputPaths[inChainID] = make(map[string]*big.Int)
		}
		if _, ok := outputPaths[inChainID][outChainID]; !ok {
			outputPaths[inChainID][outChainID] = new(big.Int)
		}

		outputPaths[inChainID][outChainID].Add(outputPaths[inChainID][outChainID], value)
	}

	return outputPaths, nil
}

func finalTradesToFinalProportions(finalTrades []*pooldb.ExchangeTrade) (map[string]map[string]*big.Int, error) {
	// iterate through each final trade, creating an index of the value each
	// trade path ended up receiving after all trades were executed

	// note that the final chain is the primary key and the
	// initial chain is the secondary key
	proportions := make(map[string]map[string]*big.Int)
	for _, finalTrade := range finalTrades {
		if !finalTrade.Proceeds.Valid {
			return nil, fmt.Errorf("no proceeds for trade %d", finalTrade.ID)
		}
		initialChainID := finalTrade.InitialChainID
		finalChainID := finalTrade.ToChainID
		proceeds := finalTrade.Proceeds.BigInt

		if _, ok := proportions[finalChainID]; !ok {
			proportions[finalChainID] = make(map[string]*big.Int)
		}
		if _, ok := proportions[finalChainID][initialChainID]; !ok {
			proportions[finalChainID][initialChainID] = new(big.Int)
		}

		proportions[finalChainID][initialChainID].Add(proportions[finalChainID][initialChainID], proceeds)
	}

	return proportions, nil
}

func finalTradesToAvgWeightedPrice(finalTrades []*pooldb.ExchangeTrade) (map[string]map[string]float64, error) {
	// iterate through each final trade, calculating the averaged weighted
	// cumulative fill price for each path

	// calculate the sum prices and weights for each path
	prices := make(map[string]map[string]float64)
	weights := make(map[string]map[string]float64)
	for _, finalTrade := range finalTrades {
		if !finalTrade.Proceeds.Valid {
			return nil, fmt.Errorf("no proceeds for trade %d", finalTrade.ID)
		} else if finalTrade.CumulativeFillPrice == nil {
			return nil, fmt.Errorf("no cumulative fill price for trade %d", finalTrade.ID)
		}
		initialChainID := finalTrade.InitialChainID
		finalChainID := finalTrade.ToChainID

		if _, ok := prices[initialChainID]; !ok {
			prices[initialChainID] = make(map[string]float64)
			weights[initialChainID] = make(map[string]float64)
		}

		// grab the units to convert the weight into a more reasonable value
		// (instead of massive numbers that could easily be >1e20)
		units, err := common.GetDefaultUnits(finalChainID)
		if err != nil {
			return nil, err
		}

		weight := common.BigIntToFloat64(finalTrade.Proceeds.BigInt, units)
		prices[initialChainID][finalChainID] += types.Float64Value(finalTrade.CumulativeFillPrice) * weight
		weights[initialChainID][finalChainID] += weight
	}

	// divide the sum prices by the sum weights to calculate the average weighted price
	for initialChainID, weightIdx := range weights {
		for finalChainID, weight := range weightIdx {
			if weight > 0 {
				prices[initialChainID][finalChainID] /= weight
			} else {
				prices[initialChainID][finalChainID] = 0
			}
		}
	}

	return prices, nil
}

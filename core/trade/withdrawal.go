package trade

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/accounting"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) InitiateWithdrawals(batchID uint64) error {
	trades, err := pooldb.GetFinalExchangeTrades(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	} else if len(trades) == 0 {
		return nil
	}

	// sum the proceeds, cumulative deposit fees, and cumulative
	// trade fees for every trade (based off of the final output chain)
	values := make(map[string]*big.Int)
	depositFees := make(map[string]*big.Int)
	tradeFees := make(map[string]*big.Int)
	for _, trade := range trades {
		if !trade.Proceeds.Valid {
			return fmt.Errorf("no proceeds for trade %d", trade.ID)
		} else if !trade.CumulativeDepositFees.Valid {
			return fmt.Errorf("no cumulative deposit fees for trade %d", trade.ID)
		} else if !trade.CumulativeTradeFees.Valid {
			return fmt.Errorf("no cumulative trade fees for trade %d", trade.ID)
		}

		if _, ok := values[trade.ToChainID]; !ok {
			values[trade.ToChainID] = new(big.Int)
		}
		if _, ok := depositFees[trade.ToChainID]; !ok {
			depositFees[trade.ToChainID] = new(big.Int)
		}
		if _, ok := tradeFees[trade.ToChainID]; !ok {
			tradeFees[trade.ToChainID] = new(big.Int)
		}

		values[trade.ToChainID].Add(values[trade.ToChainID], trade.Proceeds.BigInt)
		values[trade.ToChainID].Add(values[trade.ToChainID], trade.CumulativeDepositFees.BigInt)
		values[trade.ToChainID].Add(values[trade.ToChainID], trade.CumulativeTradeFees.BigInt)
	}

	// validate each proposed withdrawal
	for chain, value := range values {
		if value.Cmp(common.Big0) <= 0 {
			return fmt.Errorf("no value for %s", chain)
		} else if _, ok := c.nodes[chain]; !ok {
			return fmt.Errorf("no node for %s", chain)
		}

		// @TODO: maybe ignore this and just let it error out if they're disabled?
		walletActive, err := c.exchange.GetWalletStatus(chain)
		if err != nil {
			return err
		} else if !walletActive {
			return fmt.Errorf("withdrawals not enabled for chain %s", chain)
		}
	}

	// execute the withdrawal for each chain
	for chain, value := range values {
		// @TODO: check if withdrawal has already been executed
		floatValue := common.BigIntToFloat64(value, c.nodes[chain].GetUnits().Big())
		withdrawalID, err := c.exchange.CreateWithdrawal(chain, c.nodes[chain].Address(), floatValue)
		if err != nil {
			return err
		}

		withdrawal := &pooldb.ExchangeWithdrawal{
			BatchID:   batchID,
			ChainID:   chain,
			NetworkID: chain,

			ExchangeWithdrawalID: withdrawalID,

			Value:       dbcl.NullBigInt{Valid: true, BigInt: value},
			DepositFees: dbcl.NullBigInt{Valid: true, BigInt: depositFees[chain]},
			TradeFees:   dbcl.NullBigInt{Valid: true, BigInt: tradeFees[chain]},
			Confirmed:   false,
			Spent:       false,
		}

		_, err = pooldb.InsertExchangeWithdrawal(c.pooldb.Writer(), withdrawal)
		if err != nil {
			return err
		}
	}

	return c.updateBatchStatus(batchID, WithdrawalsActive)
}

func (c *Client) ConfirmWithdrawals(batchID uint64) error {
	withdrawals, err := pooldb.GetExchangeWithdrawals(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	} else if len(withdrawals) == 0 {
		return nil
	}

	completedAll := true
	for _, withdrawal := range withdrawals {
		if withdrawal.Confirmed {
			continue
		} else if !withdrawal.DepositFees.Valid {
			return fmt.Errorf("no deposit fees for withdrawal %d", withdrawal.ID)
		} else if !withdrawal.TradeFees.Valid {
			return fmt.Errorf("no trade fees for withdrawal %d", withdrawal.ID)
		}

		// fetch the withdrawal from the exchange
		parsedWithdrawal, err := c.exchange.GetWithdrawalByID(withdrawal.ChainID, withdrawal.ExchangeWithdrawalID)
		if err != nil {
			return err
		} else if parsedWithdrawal == nil || !parsedWithdrawal.Completed {
			completedAll = false
			continue
		}

		// fetch the chain's units
		units, err := common.GetDefaultUnits(withdrawal.ChainID)
		if err != nil {
			return err
		}

		// process the withdrawal value as a big int in the chain's units, calculate fees
		valueBig, err := common.StringDecimalToBigint(parsedWithdrawal.Value, units)
		if err != nil {
			return err
		}

		withdrawalFeesBig, err := common.StringDecimalToBigint(parsedWithdrawal.Fee, units)
		if err != nil {
			return err
		}

		// sum all fees to get the final cumulative fee
		cumulativeFeesBig := new(big.Int)
		cumulativeFeesBig.Add(cumulativeFeesBig, withdrawal.DepositFees.BigInt)
		cumulativeFeesBig.Add(cumulativeFeesBig, withdrawal.TradeFees.BigInt)
		cumulativeFeesBig.Add(cumulativeFeesBig, withdrawalFeesBig)

		withdrawal.Confirmed = true
		withdrawal.Value = dbcl.NullBigInt{Valid: true, BigInt: valueBig}
		withdrawal.WithdrawalFees = dbcl.NullBigInt{Valid: true, BigInt: withdrawalFeesBig}
		withdrawal.CumulativeFees = dbcl.NullBigInt{Valid: true, BigInt: cumulativeFeesBig}

		cols := []string{"value", "withdrawal_fees", "cumulative_fees", "confirmed"}
		err = pooldb.UpdateExchangeWithdrawal(c.pooldb.Writer(), withdrawal, cols)
		if err != nil {
			return err
		}
	}

	if completedAll {
		return c.updateBatchStatus(batchID, WithdrawalsComplete)
	}

	return nil
}

func (c *Client) CreditWithdrawals(batchID uint64) error {
	// fetch all miner balance inputs for the batch
	balanceInputs, err := pooldb.GetBalanceInputsByBatch(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	}

	// calculate the initial proportional value for each miner (how much
	// each one contributed to each trade path prior to the batch)
	initialProportions, err := balanceInputsToInitialProportions(balanceInputs)
	if err != nil {
		return err
	}

	// fetch the final exchange trades across every path the batch
	finalTrades, err := pooldb.GetFinalExchangeTrades(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	}

	// calculate the final, global proportional value for each trade path
	// (how	much each path ended up with after all trades were executed)
	finalProportions, err := finalTradesToFinalProportions(finalTrades)
	if err != nil {
		return err
	}

	// calculate the average weighted fill prices for each trade path
	avgPrices, err := finalTradesToAvgWeightedPrice(finalTrades)
	if err != nil {
		return err
	}

	// fetch all withdrawals for the batch
	withdrawals, err := pooldb.GetExchangeWithdrawals(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	}

	// calculate the sum adjusted pool fees for each miner and pair combination.
	// since the pool fees are in the initial chain and they can only be summed
	// in a common chain, the cumulative fill price is used to adjust them into
	// an (estimated) final chain value.
	poolFeeIdx := make(map[uint64]map[string]*big.Int)
	for _, balanceInput := range balanceInputs {
		minerID := balanceInput.MinerID
		initialChainID := balanceInput.ChainID
		finalChainID := balanceInput.OutChainID

		if !balanceInput.PoolFees.Valid {
			return fmt.Errorf("no pool fees for balance input %d", balanceInput.ID)
		} else if _, ok := avgPrices[initialChainID]; !ok {
			return fmt.Errorf("no average price for balance input %d", balanceInput.ID)
		} else if _, ok := avgPrices[initialChainID][finalChainID]; !ok {
			return fmt.Errorf("no average price for balance input %d", balanceInput.ID)
		}

		if _, ok := poolFeeIdx[minerID]; !ok {
			poolFeeIdx[minerID] = make(map[string]*big.Int)
		}
		if _, ok := poolFeeIdx[minerID][finalChainID]; !ok {
			poolFeeIdx[minerID][finalChainID] = new(big.Int)
		}

		initialUnitsBig, err := common.GetDefaultUnits(initialChainID)
		if err != nil {
			return err
		}

		finalUnitsBig, err := common.GetDefaultUnits(finalChainID)
		if err != nil {
			return err
		}

		// set all of the required values as big floats for the adjustment calculation
		poolFeeFloat := new(big.Float).SetInt(balanceInput.PoolFees.BigInt)
		priceFloat := new(big.Float).SetFloat64(avgPrices[initialChainID][finalChainID])
		initialUnitsFloat := new(big.Float).SetInt(initialUnitsBig)
		finalUnitsFloat := new(big.Float).SetInt(finalUnitsBig)

		// calculate the adjusted pool fees
		adjustedPoolFeeFloat := new(big.Float).Quo(poolFeeFloat, initialUnitsFloat)
		adjustedPoolFeeFloat.Mul(adjustedPoolFeeFloat, priceFloat)
		adjustedPoolFeeFloat.Mul(adjustedPoolFeeFloat, finalUnitsFloat)
		adjustedPoolFeeBig, _ := adjustedPoolFeeFloat.Int(nil)

		poolFeeIdx[minerID][finalChainID].Add(poolFeeIdx[minerID][finalChainID], adjustedPoolFeeBig)
	}

	// iterate through all withdrawals to calculate the proportional values
	// for each miner and create the proper balance outputs for each
	balanceOutputs := make([]*pooldb.BalanceOutput, 0)
	for _, withdrawal := range withdrawals {
		if withdrawal.Spent {
			continue
		} else if !withdrawal.Value.Valid {
			return fmt.Errorf("no value for withdrawal %d", withdrawal.ID)
		} else if !withdrawal.WithdrawalFees.Valid {
			return fmt.Errorf("no withdrawal fees for withdrawal %d", withdrawal.ID)
		} else if !withdrawal.CumulativeFees.Valid {
			return fmt.Errorf("no cumulative fees for withdrawal %d", withdrawal.ID)
		}

		// calculate the exact proportional values and fees for each trade path
		proportionalValues, proportionalFees, err := accounting.CalculateProportionalValues(withdrawal.Value.BigInt,
			withdrawal.CumulativeFees.BigInt, finalProportions[withdrawal.ChainID])
		if err != nil {
			return err
		}

		// finally, iterate through the exact proportional values and fees for each
		// input trade path and calculate the exact proportional values for each miner
		for initialChainID, proportionalValue := range proportionalValues {
			proportionalFee := proportionalFees[initialChainID]

			// make sure the given (reversed) trade path is present in the initial proportions
			if _, ok := initialProportions[withdrawal.ChainID]; !ok {
				return fmt.Errorf("no initial proportions found for %s", withdrawal.ChainID)
			} else if _, ok := initialProportions[withdrawal.ChainID][initialChainID]; !ok {
				return fmt.Errorf("no initial proportions found for %s->%s", withdrawal.ChainID, initialChainID)
			}

			// calculate the exact proportional values and fees for each miner
			minerProportionalValues, minerProportionalFees, err := accounting.CalculateProportionalValues(proportionalValue,
				proportionalFee, initialProportions[withdrawal.ChainID][initialChainID])
			if err != nil {
				return err
			}

			// create balance outputs for each miner
			for minerID, value := range minerProportionalValues {
				if _, ok := poolFeeIdx[minerID]; !ok {
					return fmt.Errorf("no pool fees found for miner %d", minerID)
				} else if _, ok := poolFeeIdx[minerID][withdrawal.ChainID]; !ok {
					return fmt.Errorf("no pool fees found for miner %d and chain %s", minerID, withdrawal.ChainID)
				}

				poolFee := poolFeeIdx[minerID][withdrawal.ChainID]
				exchangeFee := minerProportionalFees[minerID]

				balanceOutput := &pooldb.BalanceOutput{
					ChainID: withdrawal.ChainID,
					MinerID: minerID,

					Value:        dbcl.NullBigInt{Valid: true, BigInt: value},
					PoolFees:     dbcl.NullBigInt{Valid: true, BigInt: poolFee},
					ExchangeFees: dbcl.NullBigInt{Valid: true, BigInt: exchangeFee},
				}
				balanceOutputs = append(balanceOutputs, balanceOutput)
			}
		}
	}

	// create a db tx to make sure all of the balance outputs are
	// inserted, all of the balance inputs are marked as not pending
	// with the corresponding balance output, and all of the withdrawals
	// are marked as spent
	tx, err := c.pooldb.Begin()
	if err != nil {
		return err
	}
	defer tx.SafeRollback()

	// insert all of the newly created balance outputs
	err = pooldb.InsertBalanceOutputs(tx, balanceOutputs...)
	if err != nil {
		return err
	}

	// re-fetch balance outputs with their new ids
	balanceOutputs, err = pooldb.GetBalanceOutputsByBatch(tx, batchID)
	if err != nil {
		return err
	}

	// create an index for all balance output ids by miner and chain
	balanceOutputIdx := make(map[uint64]map[string]uint64)
	for _, balanceOutput := range balanceOutputs {
		minerID := balanceOutput.MinerID
		chainID := balanceOutput.ChainID

		if _, ok := balanceOutputIdx[minerID]; !ok {
			balanceOutputIdx[minerID] = make(map[string]uint64)
		}

		balanceOutputIdx[minerID][chainID] = balanceOutput.ID
	}

	// update every balance input to be marked as not pending and set
	// with the corresponding balance output id
	for _, balanceInput := range balanceInputs {
		minerID := balanceInput.MinerID
		chainID := balanceInput.OutChainID
		if _, ok := balanceOutputIdx[minerID]; !ok {
			return fmt.Errorf("no balance output found for miner %d", minerID)
		} else if _, ok := balanceOutputIdx[minerID][chainID]; !ok {
			return fmt.Errorf("no balance output found for miner %d and chain %s", minerID, chainID)
		}

		balanceInput.BalanceOutputID = types.Uint64Ptr(balanceOutputIdx[minerID][chainID])
		balanceInput.Pending = false

		cols := []string{"balance_output_id", "pending"}
		err = pooldb.UpdateBalanceInput(tx, balanceInput, cols)
		if err != nil {
			return err
		}
	}

	// mark all of the withdrawals as spent
	for _, withdrawal := range withdrawals {
		withdrawal.Spent = true
		err = pooldb.UpdateExchangeWithdrawal(tx, withdrawal, []string{"spent"})
		if err != nil {
			return err
		}
	}

	err = tx.SafeCommit()
	if err != nil {
		return err
	}

	c.telegram.NotifyFinalizeExchangeBatch(batchID)

	return c.updateBatchStatus(batchID, WithdrawalsComplete)
}
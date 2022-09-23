package trade

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
)

func (c *Client) InitiateWithdrawals(batchID uint64) error {
	trades, err := pooldb.GetFinalExchangeTrades(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	} else if len(trades) == 0 {
		return nil
	}

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
			Pending:     true,
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
		if !withdrawal.Pending {
			continue
		} else if !withdrawal.DepositFees.Valid {
			return fmt.Errorf("no deposit fees for withdrawal %d", withdrawal.ID)
		} else if !withdrawal.TradeFees.Valid {
			return fmt.Errorf("no trade fees for withdrawal %d", withdrawal.ID)
		}

		parsedWithdrawal, err := c.exchange.GetWithdrawalByID(withdrawal.ChainID, withdrawal.ExchangeWithdrawalID)
		if err != nil {
			return err
		} else if parsedWithdrawal == nil || !parsedWithdrawal.Completed {
			completedAll = false
			continue
		}

		units, err := common.GetDefaultUnits(withdrawal.ChainID)
		if err != nil {
			return err
		}

		valueBig, err := common.StringDecimalToBigint(parsedWithdrawal.Value, units)
		if err != nil {
			return err
		}

		withdrawalFeesBig, err := common.StringDecimalToBigint(parsedWithdrawal.Fee, units)
		if err != nil {
			return err
		}

		cumulativeFeesBig := new(big.Int)
		cumulativeFeesBig.Add(cumulativeFeesBig, withdrawal.DepositFees.BigInt)
		cumulativeFeesBig.Add(cumulativeFeesBig, withdrawal.TradeFees.BigInt)
		cumulativeFeesBig.Add(cumulativeFeesBig, withdrawalFeesBig)

		withdrawal.Pending = false
		withdrawal.Value = dbcl.NullBigInt{Valid: true, BigInt: valueBig}
		withdrawal.WithdrawalFees = dbcl.NullBigInt{Valid: true, BigInt: withdrawalFeesBig}
		withdrawal.CumulativeFees = dbcl.NullBigInt{Valid: true, BigInt: cumulativeFeesBig}

		cols := []string{"value", "withdrawal_fees", "cumulative_fees", "pending"}
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
	withdrawals, err := pooldb.GetExchangeWithdrawals(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	} else if len(withdrawals) == 0 {
		return nil
	}

	// @TODO: set to true
	completedAll := false

	// @TODO: properly credit withdrawals

	if completedAll {
		return c.updateBatchStatus(batchID, WithdrawalsComplete)
	}

	return nil
}

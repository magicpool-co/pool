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
	for _, trade := range trades {
		if !trade.Proceeds.Valid {
			return fmt.Errorf("no proceeds for trade %d", trade.ID)
		}

		if _, ok := values[trade.ToChainID]; !ok {
			values[trade.ToChainID] = new(big.Int)
		}
		values[trade.ToChainID].Add(values[trade.ToChainID], trade.Proceeds.BigInt)
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
		// @TODO: properly convert amount to float, using precision defined by exchange
		var amount float64

		withdrawalID, err := c.exchange.CreateWithdrawal(chain, c.nodes[chain].Address(), amount)
		if err != nil {
			return err
		}

		// @TODO: handle trade fees and withdrawal fees (and calculate deposit fees too)
		withdrawal := &pooldb.ExchangeWithdrawal{
			BatchID:   batchID,
			ChainID:   chain,
			NetworkID: chain,

			ExchangeWithdrawalID: withdrawalID,

			Value:          dbcl.NullBigInt{Valid: true, BigInt: value},
			TradeFees:      dbcl.NullBigInt{},
			WithdrawalFees: dbcl.NullBigInt{},
			Pending:        true,
			Spent:          false,
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
		}

		completed, err := c.exchange.GetWithdrawalStatus(withdrawal.ChainID, withdrawal.ExchangeWithdrawalID)
		if err != nil {
			return err
		} else if !completed {
			completedAll = false
			continue
		}

		withdrawal.Pending = false

		cols := []string{"pending"}
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

	// @TODO: properly credit withdrawals

	return nil
}

package trade

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/banker"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) InitiateDeposits(batchID uint64) error {
	balanceInputs, err := pooldb.GetExchangeInputs(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	} else if len(balanceInputs) == 0 {
		return nil
	}

	// create summed values from the inputs
	values := make(map[string]*big.Int)
	for _, balanceInput := range balanceInputs {
		if !balanceInput.Value.Valid {
			return fmt.Errorf("no value for input %d", balanceInput.ID)
		} else if _, ok := values[balanceInput.InputChainID]; !ok {
			values[balanceInput.InputChainID] = new(big.Int)
		}

		chainID := balanceInput.InputChainID
		value := balanceInput.Value.BigInt
		values[chainID].Add(values[chainID], value)
	}

	// validate each proposed deposit, create the tx outputs
	txInputIdx := make(map[string][]*types.TxInput, len(values))
	txOutputIdx := make(map[string][]*types.TxOutput, len(values))
	for chain, value := range values {
		if value.Cmp(common.Big0) <= 0 {
			return fmt.Errorf("no value for %s", chain)
		} else if _, ok := c.nodes[chain]; !ok {
			return fmt.Errorf("no node for %s", chain)
		}

		walletActive, err := c.exchange.GetWalletStatus(chain)
		if err != nil {
			return err
		} else if !walletActive {
			return fmt.Errorf("deposits not enabled for chain %s", chain)
		}

		address, err := c.exchange.GetWalletAddress(chain)
		if err != nil {
			return err
		}

		// @TODO: fetch inputs from utxos (or have banker verify there is enough value)

		txOutputIdx[chain] = []*types.TxOutput{
			&types.TxOutput{
				Address:  address,
				Value:    values[chain],
				SplitFee: true,
			},
		}
	}

	// execute the deposit for each chain
	// @TODO: manage completed all for deposits
	for chain, value := range values {
		// @TODO: check if deposit has already been executed
		txid, err := banker.SendOutgoingTx(c.nodes[chain], txInputIdx[chain], txOutputIdx[chain])
		if err != nil {
			return err
		}

		deposit := &pooldb.ExchangeDeposit{
			BatchID:   batchID,
			ChainID:   chain,
			NetworkID: chain,

			DepositTxID: txid,

			Value: dbcl.NullBigInt{Valid: true, BigInt: value},
		}

		_, err = pooldb.InsertExchangeDeposit(c.pooldb.Writer(), deposit)
		if err != nil {
			return err
		}

		// @TODO: mark inputs as spent
	}

	return c.updateBatchStatus(batchID, DepositsActive)
}

func (c *Client) RegisterDeposits(batchID uint64) error {
	deposits, err := pooldb.GetExchangeDeposits(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	} else if len(deposits) == 0 {
		return nil
	}

	registeredAll := true
	for _, deposit := range deposits {
		if deposit.Registered {
			continue
		}

		var exchangeTxID, exchangeDepositID string
		// depositTxID, depositExchangeID, err := c.exchange.GetDepositIDs(deposit.ChainID, deposit.DepositTxID)
		if err != nil {
			return err
		} else if exchangeTxID == "" || exchangeDepositID == "" {
			registeredAll = false
			continue
		}

		deposit.Registered = true
		deposit.ExchangeTxID = types.StringPtr(exchangeTxID)
		deposit.ExchangeDepositID = types.StringPtr(exchangeDepositID)

		cols := []string{"exchange_txid", "exchange_deposit_id", "registered"}
		err := pooldb.UpdateExchangeDeposit(c.pooldb.Writer(), deposit, cols)
		if err != nil {
			return err
		}
	}

	if registeredAll {
		return c.updateBatchStatus(batchID, DepositsRegistered)
	}

	return nil
}

func (c *Client) ConfirmDeposits(batchID uint64) error {
	deposits, err := pooldb.GetExchangeDeposits(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	} else if len(deposits) == 0 {
		return nil
	}

	confirmedAll := true
	for _, deposit := range deposits {
		depositID := types.StringValue(deposit.ExchangeDepositID)
		if depositID == "" {
			return fmt.Errorf("no exchange deposit ID for %d", deposit.ID)
		} else if !deposit.Pending {
			continue
		}

		completed, err := c.exchange.GetDepositStatus(deposit.ChainID, depositID)
		if err != nil {
			return err
		} else if !completed {
			confirmedAll = false
			continue
		}

		// @TODO: update deposit amount
		// @TODO: transfer to proper account
		// @TODO: confirm tx with banker

		/*
			err = c.exchange.TransferToTradeAccount(deposit.ChainID, 7501.71076222)
			if err != nil {
				return err
			}

			node, ok := c.nodes[deposit.ChainID]
			if !ok {
				return fmt.Errorf("no node found for %s", deposit.ChainID)
			}

			banker.ConfirmOutgoingTx(node, deposit.DepositTxID)
		*/

		deposit.Pending = false
		err = pooldb.UpdateExchangeDeposit(c.pooldb.Writer(), deposit, []string{"pending"})
		if err != nil {
			return err
		}
	}

	if confirmedAll {
		return c.updateBatchStatus(batchID, DepositsComplete)
	}

	return nil
}

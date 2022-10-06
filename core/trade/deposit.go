package trade

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/core/bank"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) InitiateDeposits(batchID uint64) error {
	exchangeInputs, err := pooldb.GetExchangeInputs(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	}

	// create summed values from the exchange inputs
	values := make(map[string]*big.Int)
	for _, exchangeInput := range exchangeInputs {
		chainID := exchangeInput.InChainID
		if !exchangeInput.Value.Valid {
			return fmt.Errorf("no value for input %d", exchangeInput.ID)
		} else if _, ok := values[chainID]; !ok {
			values[chainID] = new(big.Int)
		}

		values[chainID].Add(values[chainID], exchangeInput.Value.BigInt)
	}

	// validate each proposed deposit, create the tx outputs
	txOutputIdx := make(map[string][]*types.TxOutput, len(values))
	for chain, value := range values {
		if value.Cmp(common.Big0) <= 0 {
			return fmt.Errorf("no value for %s", chain)
		} else if _, ok := c.nodes[chain]; !ok {
			return fmt.Errorf("no node for %s", chain)
		}

		// verify the exchange supports the chain for deposits and withdrawals
		depositsEnabled, _, err := c.exchange.GetWalletStatus(chain)
		if err != nil {
			return err
		} else if !depositsEnabled {
			return fmt.Errorf("deposits not enabled for chain %s", chain)
		}

		// fetch the deposit address from the exchange
		address, err := c.exchange.GetDepositAddress(chain)
		if err != nil {
			return err
		}

		txOutputIdx[chain] = []*types.TxOutput{
			&types.TxOutput{
				Address:  address,
				Value:    values[chain],
				SplitFee: true,
			},
		}
	}

	// make sure any deposits don't end up going through twice
	deposits, err := pooldb.GetExchangeDeposits(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	}

	for _, deposit := range deposits {
		delete(values, deposit.ChainID)
	}

	// execute the deposit for each chain
	for chain, value := range values {
		// @TODO: check if deposit has already been executed
		txid, err := bank.SendOutgoingTx(c.nodes[chain], c.pooldb, txOutputIdx[chain])
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

		depositID, err := pooldb.InsertExchangeDeposit(c.pooldb.Writer(), deposit)
		if err != nil {
			return err
		}

		floatValue := common.BigIntToFloat64(value, c.nodes[chain].GetUnits().Big())
		c.telegram.NotifyInitiateDeposit(depositID, chain, txid, c.nodes[chain].GetTxExplorerURL(txid), floatValue)
	}

	return c.updateBatchStatus(batchID, DepositsActive)
}

func (c *Client) RegisterDeposits(batchID uint64) error {
	deposits, err := pooldb.GetExchangeDeposits(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	}

	registeredAll := true
	for _, deposit := range deposits {
		if deposit.Registered {
			continue
		}

		// fetch the deposit from the exchange and register it in the db
		parsedDeposit, err := c.exchange.GetDepositByTxID(deposit.ChainID, deposit.DepositTxID)
		if err != nil {
			return err
		} else if parsedDeposit.ID == "" {
			registeredAll = false
			continue
		}

		deposit.Registered = true
		deposit.ExchangeTxID = types.StringPtr(parsedDeposit.TxID)
		deposit.ExchangeDepositID = types.StringPtr(parsedDeposit.ID)

		cols := []string{"exchange_txid", "exchange_deposit_id", "registered"}
		err = pooldb.UpdateExchangeDeposit(c.pooldb.Writer(), deposit, cols)
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
	}

	confirmedAll := true
	for _, deposit := range deposits {
		depositID := types.StringValue(deposit.ExchangeDepositID)
		if deposit.Confirmed {
			continue
		} else if depositID == "" {
			return fmt.Errorf("no exchange deposit ID for %d", deposit.ID)
		} else if !deposit.Value.Valid {
			return fmt.Errorf("no value for deposit %d", deposit.ID)
		}

		// fetch the deposit from the exchange
		parsedDeposit, err := c.exchange.GetDepositByID(deposit.ChainID, depositID)
		if err != nil {
			return err
		} else if !parsedDeposit.Completed {
			confirmedAll = false
			continue
		}

		// fetch the chain's units
		units, err := common.GetDefaultUnits(deposit.ChainID)
		if err != nil {
			return err
		}

		// process the deposit value as a big int in the chain's units, calculate fees
		valueBig, err := common.StringDecimalToBigint(parsedDeposit.Value, units)
		if err != nil {
			return err
		}
		feesBig := new(big.Int).Sub(deposit.Value.BigInt, valueBig)

		// transfer the balance from the main account to the
		// trade account (kucoin only, empty method otherwise)
		err = c.exchange.TransferToTradeAccount(deposit.ChainID, common.BigIntToFloat64(valueBig, units))
		if err != nil {
			return err
		}

		deposit.Confirmed = true
		deposit.Value = dbcl.NullBigInt{Valid: true, BigInt: valueBig}
		deposit.Fees = dbcl.NullBigInt{Valid: true, BigInt: feesBig}

		cols := []string{"value", "fees", "confirmed"}
		err = pooldb.UpdateExchangeDeposit(c.pooldb.Writer(), deposit, cols)
		if err != nil {
			return err
		}

		c.telegram.NotifyFinalizeDeposit(deposit.ID)
	}

	if confirmedAll {
		return c.updateBatchStatus(batchID, DepositsComplete)
	}

	return nil
}

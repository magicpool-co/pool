package trade

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	txCommon "github.com/magicpool-co/pool/pkg/crypto/tx"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) executeAndMaybeSplitDeposit(
	batchID uint64,
	chain string,
	output *types.TxOutput,
	split int,
) error {
	dbTx, err := c.pooldb.Begin()
	if err != nil {
		return err
	}
	defer dbTx.SafeRollback()

	var outputs []*types.TxOutput
	if split > 1 {
		splitCount := new(big.Int).SetUint64(uint64(split))
		splitValue := new(big.Int).Div(output.Value, splitCount)
		sum := new(big.Int).Mul(splitValue, splitCount)
		remainder := sum.Sub(output.Value, sum)

		outputs = make([]*types.TxOutput, split)
		for i := 0; i < split; i++ {
			outputs[i] = &types.TxOutput{
				Address:  output.Address,
				Value:    new(big.Int).Set(splitValue),
				SplitFee: true,
			}

			if i == 0 {
				outputs[i].Value.Add(outputs[i].Value, remainder)
			}
		}
	} else {
		outputs = []*types.TxOutput{output}
	}

	outputList := make([][]*types.TxOutput, len(outputs))
	for i, splitOutput := range outputs {
		outputList[i] = []*types.TxOutput{splitOutput}
	}

	txs, err := c.bank.PrepareOutgoingTxs(dbTx, c.nodes[chain], types.DepositTx, outputList...)
	if err != nil {
		if err == txCommon.ErrTxTooBig {
			if split > 10 {
				return fmt.Errorf("past split limit of 10")
			}

			return c.executeAndMaybeSplitDeposit(batchID, chain, output, split+1)
		}

		return err
	} else if len(txs) == 0 {
		return fmt.Errorf("no txs for deposit preparation")
	}

	for i, tx := range txs {
		splitValue := outputList[i][0].Value
		deposit := &pooldb.ExchangeDeposit{
			BatchID:   batchID,
			ChainID:   chain,
			NetworkID: chain,

			DepositTxID: tx.TxID,

			Value: dbcl.NullBigInt{Valid: true, BigInt: splitValue},
		}

		depositID, err := pooldb.InsertExchangeDeposit(c.pooldb.Writer(), deposit)
		if err != nil {
			return err
		}

		floatValue := common.BigIntToFloat64(splitValue, c.nodes[chain].GetUnits().Big())
		c.telegram.NotifyInitiateDeposit(depositID, chain, floatValue)
	}

	return dbTx.SafeCommit()
}

func (c *Client) InitiateDeposits(batchID uint64, exchange types.Exchange) error {
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
	txOutputIdx := make(map[string]*types.TxOutput, len(values))
	for chain, value := range values {
		if value.Cmp(common.Big0) <= 0 {
			return fmt.Errorf("no value for %s", chain)
		} else if _, ok := c.nodes[chain]; !ok {
			return fmt.Errorf("no node for %s", chain)
		}

		// verify the exchange supports the chain for deposits and withdrawals
		depositsEnabled, _, err := exchange.GetWalletStatus(chain)
		if err != nil {
			return err
		} else if !depositsEnabled {
			return fmt.Errorf("deposits not enabled for chain %s", chain)
		}

		// fetch the deposit address from the exchange
		address, err := exchange.GetDepositAddress(chain)
		if err != nil {
			return err
		}

		txOutputIdx[chain] = &types.TxOutput{
			Address:  address,
			Value:    value,
			SplitFee: true,
		}
	}

	deposits, err := pooldb.GetExchangeDeposits(c.pooldb.Reader(), batchID)
	if err != nil {
		return err
	}

	depositValueIdx := make(map[string]*big.Int)
	for _, deposit := range deposits {
		if !deposit.Value.Valid {
			continue
		} else if _, ok := depositValueIdx[deposit.ChainID]; !ok {
			depositValueIdx[deposit.ChainID] = new(big.Int)
		}

		tx, err := pooldb.GetTransactionByTxID(c.pooldb.Reader(), deposit.DepositTxID)
		if err != nil {
			return err
		} else if !tx.Fee.Valid {
			return fmt.Errorf("no fee for tx %d", tx.ID)
		}

		depositValueIdx[deposit.ChainID].Add(depositValueIdx[deposit.ChainID], deposit.Value.BigInt)
		depositValueIdx[deposit.ChainID].Add(depositValueIdx[deposit.ChainID], tx.Fee.BigInt)
	}

	initiatedAll := true
	for chain, value := range values {
		if depositedValue, ok := depositValueIdx[chain]; ok {
			if depositedValue.Cmp(value) >= 0 {
				continue
			}

			txOutputIdx[chain].Value.Sub(txOutputIdx[chain].Value, depositedValue)
		}

		err := c.executeAndMaybeSplitDeposit(batchID, chain, txOutputIdx[chain], 1)
		if err != nil {
			return err
		}
	}

	if initiatedAll {
		return c.updateBatchStatus(c.pooldb.Writer(), batchID, DepositsActive)
	}

	return nil
}

func (c *Client) RegisterDeposits(batchID uint64, exchange types.Exchange) error {
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
		parsedDeposit, err := exchange.GetDepositByTxID(deposit.ChainID, deposit.DepositTxID)
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
		return c.updateBatchStatus(c.pooldb.Writer(), batchID, DepositsRegistered)
	}

	return nil
}

func (c *Client) ConfirmDeposits(batchID uint64, exchange types.Exchange) error {
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
		parsedDeposit, err := exchange.GetDepositByID(deposit.ChainID, depositID)
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
		err = exchange.TransferToTradeAccount(deposit.ChainID, common.BigIntToFloat64(valueBig, units))
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
		return c.updateBatchStatus(c.pooldb.Writer(), batchID, DepositsComplete)
	}

	return nil
}

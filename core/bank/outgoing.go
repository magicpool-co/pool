package bank

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	txCommon "github.com/magicpool-co/pool/pkg/crypto/tx"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) PrepareOutgoingTxs(
	q dbcl.Querier,
	node types.PayoutNode,
	txType types.TransactionType,
	txOutputList ...[]*types.TxOutput,
) ([]*pooldb.Transaction, error) {
	txs := make([]*pooldb.Transaction, len(txOutputList))

	// verify that there are no other unspent transactions active
	count, err := pooldb.GetUnspentTransactionCount(q, node.Chain())
	if err != nil {
		return nil, err
	} else if count > 0 {
		return nil, nil
	}

	allInputUTXOs, err := pooldb.GetUnspentUTXOsByChain(q, node.Chain())
	if err != nil {
		return nil, err
	} else if len(allInputUTXOs) == 0 {
		return nil, nil
	}

	// calculate total spendable value
	totalInputUTXOSum := new(big.Int)
	for _, inputUTXO := range allInputUTXOs {
		if !inputUTXO.Value.Valid {
			return nil, fmt.Errorf("no value for utxo %d", inputUTXO.ID)
		}
		totalInputUTXOSum.Add(totalInputUTXOSum, inputUTXO.Value.BigInt)
	}

	totalTxOutputSum := new(big.Int)
	for _, txOutputs := range txOutputList {
		for _, txOutput := range txOutputs {
			totalTxOutputSum.Add(totalTxOutputSum, txOutput.Value)
		}
	}

	// check for empty, negative, and over spends
	remainder := new(big.Int).Sub(totalInputUTXOSum, totalTxOutputSum)
	if remainder.Cmp(common.Big0) < 0 {
		return txs, fmt.Errorf("%s overspend: %s < %s", node.Chain(), totalInputUTXOSum, totalTxOutputSum)
	}

	for i, txOutputs := range txOutputList {
		// copy utxos to new list
		inputUTXOs := make([]*pooldb.UTXO, len(allInputUTXOs))
		for j, utxo := range allInputUTXOs {
			inputUTXOs[j] = utxo
		}

		// calculate total spendable value
		inputUTXOSum := new(big.Int)
		for _, inputUTXO := range inputUTXOs {
			if !inputUTXO.Value.Valid {
				return nil, fmt.Errorf("no value for utxo %d", inputUTXO.ID)
			}
			inputUTXOSum.Add(inputUTXOSum, inputUTXO.Value.BigInt)
		}

		// calculate total tx output value
		txOutputSum := new(big.Int)
		for _, txOutput := range txOutputs {
			txOutputSum.Add(txOutputSum, txOutput.Value)
		}

		// check for empty, negative, and over spends
		remainder := new(big.Int).Sub(inputUTXOSum, txOutputSum)
		if txOutputSum.Cmp(common.Big0) <= 0 {
			return nil, fmt.Errorf("%s empty or negative spend: %s", node.Chain(), txOutputSum)
		} else if remainder.Cmp(common.Big0) < 0 {
			return nil, fmt.Errorf("%s overspend: %s < %s", node.Chain(), inputUTXOSum, txOutputSum)
		}

		// add inputs from UTXOs based off of chain accounting type (account or UTXO)
		var inputs []*types.TxInput
		switch node.GetAccountingType() {
		case types.AccountStructure:
			// spend all utxos since theyre artificial anyways
			allInputUTXOs = []*pooldb.UTXO{}

			count, err := pooldb.GetUnspentTransactionCount(q, node.Chain())
			if err != nil {
				return nil, err
			}

			// txOutputs count has to be non-zero since output
			// sum has already been verified as non-zero
			inputs = []*types.TxInput{
				&types.TxInput{
					Value:      txOutputs[0].Value,
					FeeBalance: txOutputs[0].FeeBalance,
					Index:      uint32(count),
				},
			}
		case types.UTXOStructure:
			// prune utxos to only use enough to cover the balance
			for j := len(inputUTXOs) - 1; j > 0; j-- {
				utxo := inputUTXOs[j]
				if remainder.Cmp(utxo.Value.BigInt) < 0 {
					break
				}

				inputUTXOs = inputUTXOs[:j]
				remainder.Sub(remainder, utxo.Value.BigInt)
			}

			allInputUTXOs = allInputUTXOs[len(inputUTXOs):]

			// convert pooldb.UTXO to types.TxInput as txInputs
			inputs = make([]*types.TxInput, len(inputUTXOs))
			for j, inputUTXO := range inputUTXOs {
				inputs[j] = &types.TxInput{
					Hash:  inputUTXO.TxID,
					Index: inputUTXO.Index,
					Value: inputUTXO.Value.BigInt,
				}
			}

			// if the remainder is non-zero, add a remainder output
			// (except for ERG, since wallet is managed by the node)
			if remainder.Cmp(common.Big0) > 0 && node.Chain() != "ERG" {
				remainderOutput := &types.TxOutput{
					Address:  node.Address(),
					Value:    remainder,
					SplitFee: false,
				}
				txOutputs = append(txOutputs, remainderOutput)
			}
		}

		// create tx and insert it into the db
		txid, txHex, err := node.CreateTx(inputs, txOutputs)
		if err != nil {
			return nil, err
		}

		feeSum := new(big.Int)
		for _, txOutput := range txOutputs {
			feeSum.Add(feeSum, txOutput.Fee)
		}

		txs[i] = &pooldb.Transaction{
			ChainID:      node.Chain(),
			Type:         int(txType),
			TxID:         txid,
			TxHex:        txHex,
			Value:        dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).Sub(txOutputSum, feeSum)},
			Fee:          dbcl.NullBigInt{Valid: true, BigInt: feeSum},
			Remainder:    dbcl.NullBigInt{Valid: true, BigInt: remainder},
			RemainderIdx: uint32(len(txOutputs) - 1),
		}

		txs[i].ID, err = pooldb.InsertTransaction(q, txs[i])
		if err != nil {
			return nil, err
		}

		// bind utxos to transaction
		for _, utxo := range inputUTXOs {
			utxo.TransactionID = types.Uint64Ptr(txs[i].ID)
			err = pooldb.UpdateUTXO(q, utxo, []string{"transaction_id"})
			if err != nil {
				return nil, err
			}
		}

		// iterate through all outputs and add to list of remainder values (and remainder indexes).
		// this only will work for UTXO-based chains since account-based chains don't have a
		// remainder txOutput (this is designed to handle multiple remainder outputs
		// to the wallet in the case of UTXO merging).
		remainderValues := make([]*big.Int, 0)
		remainderIndexes := make([]int, 0)
		for j, output := range txOutputs {
			if output.Address == node.Address() {
				remainderValues = append(remainderValues, new(big.Int).Set(output.Value))
				remainderIndexes = append(remainderIndexes, j)
			}
		}

		// since account-based chains have no txOutputs (along with ERG),
		// if the remainder is non-zero add it manually
		if len(remainderValues) == 0 && remainder.Cmp(common.Big0) > 0 {
			remainderValues = append(remainderValues, new(big.Int).Set(remainder))
			remainderIndexes = append(remainderIndexes, 0)
		}

		// create a UTXO for each remainder value and add it to the inputUTXOs
		// for use in the next loop
		for j, remainderValue := range remainderValues {
			remainderUTXO := &pooldb.UTXO{
				ChainID: node.Chain(),
				TxID:    txid,
				Index:   uint32(remainderIndexes[j]),
				Value:   dbcl.NullBigInt{Valid: true, BigInt: remainderValue},
				Active:  false,
			}

			allInputUTXOs = append(allInputUTXOs, remainderUTXO)
			remainderUTXO.ID, err = pooldb.InsertUTXO(q, remainderUTXO)
			if err != nil {
				return nil, err
			}
		}
	}

	return txs, nil
}

func (c *Client) MergeUTXOs(node types.PayoutNode, count int) error {
	if node.GetAccountingType() != types.UTXOStructure || !node.ShouldMergeUTXOs() {
		return nil
	}

	dbTx, err := c.pooldb.Begin()
	if err != nil {
		return err
	}
	defer dbTx.SafeRollback()

	inputUTXOs, err := pooldb.GetUnspentUTXOsByChain(dbTx, node.Chain())
	if err != nil {
		return err
	} else if len(inputUTXOs) == 0 {
		return nil
	}

	size := len(inputUTXOs) / count
	for {
		txOutputList := make([][]*types.TxOutput, count)
		for i := 0; i < count; i++ {
			outputSum := new(big.Int)
			for _, utxo := range inputUTXOs[i*size : (i+1)*size] {
				outputSum.Add(outputSum, utxo.Value.BigInt)
			}

			txOutputList[i] = []*types.TxOutput{
				&types.TxOutput{
					Address:  node.Address(),
					Value:    outputSum,
					SplitFee: true,
				},
			}
		}

		txs, err := c.PrepareOutgoingTxs(dbTx, node, types.MergeTx, txOutputList...)
		if err == txCommon.ErrTxTooBig {
			size /= 2
			if size < 2 {
				return nil
			}
			continue
		} else if err != nil {
			return err
		} else if len(txOutputList) != len(txs) {
			return fmt.Errorf("mismatch on tx output list and tx lengths")
		}

		for i, txOutputs := range txOutputList {
			tx := txs[i]
			if tx == nil {
				return fmt.Errorf("tx not found")
			}

			fee := txOutputs[0].Fee
			if fee == nil {
				return fmt.Errorf("empty tx fee")
			}

			// fetch a random balance output that is above the fee value
			balanceOutput, err := pooldb.GetRandomBalanceOutputAboveValue(dbTx, node.Chain(), fee.String())
			if err != nil {
				return err
			} else if balanceOutput == nil || balanceOutput.ID == 0 || !balanceOutput.Value.Valid {
				return fmt.Errorf("no balance output found")
			} else if balanceOutput.Value.BigInt.Cmp(fee) <= 0 {
				return fmt.Errorf("balance output less than or equal to merge tx fee")
			}

			// update the balance output value to subtract the fee
			balanceOutput.Value.BigInt.Sub(balanceOutput.Value.BigInt, fee)
			err = pooldb.UpdateBalanceOutput(dbTx, balanceOutput, []string{"value"})
			if err != nil {
				return err
			}

			// create a mature, spent balance output to maintain the record of
			// who was charged fro the merge fee
			subBalanceOutput := &pooldb.BalanceOutput{
				ChainID: balanceOutput.ChainID,
				MinerID: balanceOutput.MinerID,

				OutMergeTransactionID: types.Uint64Ptr(tx.ID),

				Value:        dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				PoolFees:     dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				ExchangeFees: dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				TxFees:       dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).Set(fee)},
				Mature:       true,
				Spent:        true,
			}
			err = pooldb.InsertBalanceOutputs(dbTx, subBalanceOutput)
			if err != nil {
				return err
			}

			// subtract sum value for balance outputs spent in the merge
			err = pooldb.InsertSubtractBalanceSums(dbTx, &pooldb.BalanceSum{
				MinerID: balanceOutput.MinerID,
				ChainID: balanceOutput.ChainID,

				MatureValue: dbcl.NullBigInt{Valid: true, BigInt: fee},
			})
			if err != nil {
				return err
			}
		}

		break
	}

	return dbTx.SafeCommit()
}

func (c *Client) spendTx(node types.PayoutNode, tx *pooldb.Transaction) error {
	lock, err := c.FetchLock(node.Chain())
	if err != nil {
		return err
	}
	defer lock.Release(context.Background())

	dbTx, err := c.pooldb.Begin()
	if err != nil {
		return err
	}
	defer dbTx.SafeRollback()

	// verify input utxo sum to make sure there are enough funds in the wallet
	inputUTXOs, err := pooldb.GetUTXOsByTransactionID(dbTx, tx.ID)
	if err != nil {
		return err
	}

	inputUTXOSum := new(big.Int)
	for _, inputUTXO := range inputUTXOs {
		if !inputUTXO.Value.Valid {
			return fmt.Errorf("no value for utxo %d", inputUTXO.ID)
		}
		inputUTXOSum.Add(inputUTXOSum, inputUTXO.Value.BigInt)
	}

	if !tx.Value.Valid {
		return fmt.Errorf("no value for tx %d", tx.ID)
	} else if !tx.Fee.Valid {
		return fmt.Errorf("no fee for tx %d", tx.ID)
	} else if inputUTXOSum.Cmp(tx.Value.BigInt) < 0 {
		return fmt.Errorf("overspend on tx %d: have %s, want %s", tx.ID, inputUTXOSum, tx.Value.BigInt)
	}

	switch tx.Type {
	case int(types.DepositTx): // verify the tx is the same as the value of all unregistered deposits
		deposits, err := pooldb.GetUnregisteredExchangeDepositsByChain(dbTx, node.Chain())
		if err != nil {
			return err
		}

		depositSum := new(big.Int)
		for _, deposit := range deposits {
			if !deposit.Value.Valid {
				return fmt.Errorf("no value for deposit %d", deposit.ID)
			} else if deposit.DepositTxID != tx.TxID {
				continue
			}

			depositSum.Add(depositSum, deposit.Value.BigInt)
		}

		if depositSum.Cmp(tx.Value.BigInt) != 0 {
			return fmt.Errorf("deposit sum and tx value mismatch: have %s, want %s", depositSum, tx.Value.BigInt)
		}
	case int(types.PayoutTx): // verify balance output sum to make sure the correct amount the miner is owed is being spent
		balanceOutputs, err := pooldb.GetBalanceOutputsByPayoutTransaction(dbTx, tx.ID)
		if err != nil {
			return err
		}

		balanceOutputSum := new(big.Int)
		for _, balanceOutput := range balanceOutputs {
			if !balanceOutput.Value.Valid {
				return fmt.Errorf("no value for balance output %d", balanceOutput.ID)
			} else if balanceOutput.Spent {
				continue
			}
			balanceOutputSum.Add(balanceOutputSum, balanceOutput.Value.BigInt)
		}

		balanceOutputSum.Sub(balanceOutputSum, tx.Fee.BigInt)
		if balanceOutputSum.Cmp(tx.Value.BigInt) != 0 {
			return fmt.Errorf("balance output sum and tx value mismatch: have %s, want %s", balanceOutputSum, tx.Value.BigInt)
		}
	}

	txid, err := node.BroadcastTx(tx.TxHex)
	if err != nil {
		return fmt.Errorf("broadcast: %v", err)
	}

	// spend input utxos
	for _, utxo := range inputUTXOs {
		utxo.Spent = true
		err = pooldb.UpdateUTXO(dbTx, utxo, []string{"spent"})
		if err != nil {
			return err
		}
	}

	err = pooldb.UpdateUTXOByTxID(dbTx, &pooldb.UTXO{TxID: tx.TxID, Active: true}, []string{"active"})
	if err != nil {
		return err
	}

	tx.Spent = true
	if txid != tx.TxID {
		tx.TxID = txid
	}

	err = pooldb.UpdateTransaction(dbTx, tx, []string{"txid", "spent"})
	if err != nil {
		return err
	}

	floatValue := common.BigIntToFloat64(tx.Value.BigInt, node.GetUnits().Big())
	c.telegram.NotifyTransactionSent(tx.ID, node.Chain(), tx.TxID, node.GetTxExplorerURL(tx.TxID), floatValue)

	return dbTx.SafeCommit()
}

func (c *Client) BroadcastOutgoingTxs(node types.PayoutNode) error {
	txs, err := pooldb.GetUnspentTransactions(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return err
	}

	for _, tx := range txs {
		err := c.spendTx(node, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) ConfirmOutgoingTxs(node types.PayoutNode) error {
	txs, err := pooldb.GetUnconfirmedTransactions(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return err
	}

	for _, tx := range txs {
		nodeTx, err := node.GetTx(tx.TxID)
		if err != nil {
			return err
		} else if nodeTx == nil {
			continue
		} else if !nodeTx.Confirmed {
			if time.Since(tx.CreatedAt) > time.Hour*24 {
				// @TODO: manage failed transactions
			}
			continue
		}

		tx.Height = types.Uint64Ptr(nodeTx.BlockNumber)
		tx.Confirmed = true
		if nodeTx.FeeBalance != nil && nodeTx.FeeBalance.Cmp(common.Big0) > 0 {
			if nodeTx.Fee == nil {
				return fmt.Errorf("no fee for tx %s", nodeTx.Hash)
			}

			tx.Fee = dbcl.NullBigInt{Valid: true, BigInt: nodeTx.Fee}
			tx.FeeBalance = dbcl.NullBigInt{Valid: true, BigInt: nodeTx.FeeBalance}

			utxo := &pooldb.UTXO{
				ChainID: node.Chain(),
				TxID:    tx.TxID,
				Index:   0,
				Value:   tx.FeeBalance,
				Active:  true,
				Spent:   false,
			}

			err = pooldb.InsertUTXOs(c.pooldb.Writer(), utxo)
			if err != nil {
				return err
			}
		}

		cols := []string{"height", "fee", "fee_balance", "confirmed"}
		err = pooldb.UpdateTransaction(c.pooldb.Writer(), tx, cols)
		if err != nil {
			return err
		}
	}

	return nil
}

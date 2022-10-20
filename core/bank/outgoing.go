package bank

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) PrepareOutgoingTxs(q dbcl.Querier, node types.PayoutNode, txOutputList ...[]*types.TxOutput) ([]*pooldb.Transaction, error) {
	txs := make([]*pooldb.Transaction, len(txOutputList))

	// distributed lock to avoid race conditions
	lock, err := c.fetchLock(node.Chain())
	if err != nil {
		return txs, err
	} else if lock == nil {
		return txs, nil
	}
	defer lock.Release(context.Background())

	// verify that there are no other unspent transactions active
	count, err := pooldb.GetUnspentTransactionCount(q, node.Chain())
	if err != nil {
		return txs, err
	} else if count > 0 {
		return txs, nil
	}

	inputUTXOs, err := pooldb.GetUnspentUTXOsByChain(q, node.Chain())
	if err != nil {
		return txs, err
	} else if len(inputUTXOs) == 0 {
		return txs, nil
	}

	// calculate total spendable value
	totalInputUTXOSum := new(big.Int)
	for _, inputUTXO := range inputUTXOs {
		if !inputUTXO.Value.Valid {
			return txs, fmt.Errorf("no value for utxo %d", inputUTXO.ID)
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
	if remainder.Cmp(common.Big0) <= 0 {
		return txs, fmt.Errorf("%s overspend: %s < %s", node.Chain(), totalInputUTXOSum, totalTxOutputSum)
	}

	for i, txOutputs := range txOutputList {
		// calculate total spendable value
		inputUTXOSum := new(big.Int)
		for _, inputUTXO := range inputUTXOs {
			if !inputUTXO.Value.Valid {
				return txs, fmt.Errorf("no value for utxo %d", inputUTXO.ID)
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
			return txs, fmt.Errorf("%s empty or negative spend: %s", node.Chain(), txOutputSum)
		} else if remainder.Cmp(common.Big0) <= 0 {
			return txs, fmt.Errorf("%s overspend: %s < %s", node.Chain(), inputUTXOSum, txOutputSum)
		}

		// add inputs from UTXOs based off of chain accounting type (account or UTXO)
		var inputs []*types.TxInput
		switch node.GetAccountingType() {
		case types.AccountStructure:
			count, err := pooldb.GetUnspentTransactionCount(q, node.Chain())
			if err != nil {
				return txs, err
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
			// convert pooldb.UTXO to types.TxInput as txInputs
			inputs = make([]*types.TxInput, len(inputUTXOs))
			for i, inputUTXO := range inputUTXOs {
				inputs[i] = &types.TxInput{
					Hash:  inputUTXO.TxID,
					Index: inputUTXO.Index,
					Value: inputUTXO.Value.BigInt,
				}
			}

			// if the remainder is non-zero, add a remainder output
			// (except for ERGO, since wallet is managed by the node)
			if remainder.Cmp(common.Big0) > 0 && node.Chain() != "ERGO" {
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
			return txs, err
		}

		feeSum := new(big.Int)
		for _, txOutput := range txOutputs {
			feeSum.Add(feeSum, txOutput.Fee)
		}

		txs[i] = &pooldb.Transaction{
			ChainID:      node.Chain(),
			TxID:         txid,
			TxHex:        txHex,
			Value:        dbcl.NullBigInt{Valid: true, BigInt: txOutputSum},
			Fee:          dbcl.NullBigInt{Valid: true, BigInt: feeSum},
			Remainder:    dbcl.NullBigInt{Valid: true, BigInt: remainder},
			RemainderIdx: uint32(len(txOutputs) - 1),
		}

		txs[i].ID, err = pooldb.InsertTransaction(q, txs[i])
		if err != nil {
			return txs, err
		}

		// bind utxos to transaction
		for _, utxo := range inputUTXOs {
			utxo.TransactionID = types.Uint64Ptr(txs[i].ID)
			err = pooldb.UpdateUTXO(q, utxo, []string{"transaction_id"})
			if err != nil {
				return txs, err
			}
		}

		inputUTXOs = []*pooldb.UTXO{}
		if remainder.Cmp(common.Big0) > 0 {
			inputUTXOs = []*pooldb.UTXO{
				&pooldb.UTXO{
					ChainID: node.Chain(),
					TxID:    txid,
					Index:   txs[i].RemainderIdx,
					Value:   txs[i].Remainder,
					Active:  false,
				},
			}

			err = pooldb.InsertUTXOs(q, inputUTXOs...)
			if err != nil {
				return txs, err
			}
		}
	}

	return txs, nil
}

func (c *Client) spendTx(node types.PayoutNode, tx, nextTx *pooldb.Transaction) error {
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
	} else if inputUTXOSum.Cmp(tx.Value.BigInt) < 0 {
		return fmt.Errorf("overspend on tx %d: have %s, want %s", tx.ID, inputUTXOSum, tx.Value.BigInt)
	}

	// verify balance output sum to make sure the correct amount the miner is owed is being spent
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

	if balanceOutputSum.Cmp(tx.Value.BigInt) != 0 {
		return fmt.Errorf("balance output sum and tx value mismatch: have %s, want %s", balanceOutputSum, tx.Value.BigInt)
	}

	txid, err := node.BroadcastTx(tx.TxHex)
	if err != nil {
		return err
	}

	// spend input utxos
	for _, utxo := range inputUTXOs {
		utxo.Spent = true
		err = pooldb.UpdateUTXO(dbTx, utxo, []string{"spent"})
		if err != nil {
			return err
		}
	}

	if nextTx != nil {
		utxos, err := pooldb.GetUTXOsByTransactionID(dbTx, nextTx.ID)
		if err != nil {
			return err
		}

		for _, utxo := range utxos {
			utxo.Active = true
			err = pooldb.UpdateUTXO(dbTx, utxo, []string{"active"})
			if err != nil {
				return err
			}
		}
	}

	tx.Spent = true
	if txid != tx.TxID {
		tx.TxID = txid
	}

	err = pooldb.UpdateTransaction(dbTx, tx, []string{"txid", "spent"})
	if err != nil {
		return err
	}

	return dbTx.SafeCommit()
}

func (c *Client) BroadcastOutgoingTxs(node types.PayoutNode) error {
	// distributed lock to avoid race conditions
	lock, err := c.fetchLock(node.Chain())
	if err != nil {
		return err
	} else if lock == nil {
		return nil
	}
	defer lock.Release(context.Background())

	const maxTxLimit = 5
	txs, err := pooldb.GetUnspentTransactions(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return err
	} else if len(txs) > maxTxLimit {
		txs = txs[:maxTxLimit]
	}

	for i, tx := range txs {
		var nextTx *pooldb.Transaction
		if i < len(txs)-1 {
			nextTx = txs[i+1]
		}

		err := c.spendTx(node, tx, nextTx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) ConfirmOutgoingTxs(node types.PayoutNode) error {
	// distributed lock to avoid race conditions
	lock, err := c.fetchLock(node.Chain())
	if err != nil {
		return err
	} else if lock == nil {
		return nil
	}
	defer lock.Release(context.Background())

	txs, err := pooldb.GetUnconfirmedTransactions(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return err
	}

	for _, tx := range txs {
		nodeTx, err := node.GetTx(tx.TxID)
		if err != nil {
			return err
		} else if !nodeTx.Confirmed {
			if time.Since(tx.CreatedAt) > time.Hour*24 {
				// @TODO: manage failed transactions
			}
			continue
		} else if nodeTx.Value == nil {
			return fmt.Errorf("no value for tx %s", nodeTx.Hash)
		} else if nodeTx.Fee == nil {
			return fmt.Errorf("no fee for tx %s", nodeTx.Hash)
		}

		tx.Height = types.Uint64Ptr(nodeTx.BlockNumber)
		tx.Confirmed = true
		if nodeTx.FeeBalance != nil && nodeTx.FeeBalance.Cmp(common.Big0) > 0 {
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

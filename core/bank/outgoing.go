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

func (c *Client) PrepareOutgoingTx(node types.PayoutNode, txOutputs []*types.TxOutput) (*pooldb.Transaction, error) {
	// distributed lock to avoid race conditions
	lock, err := c.fetchLock(node.Chain())
	if err != nil {
		return nil, err
	} else if lock == nil {
		return nil, nil
	}
	defer lock.Release(context.Background())

	inputUTXOs, err := pooldb.GetUnspentUTXOsByChain(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return nil, err
	} else if len(inputUTXOs) == 0 {
		return nil, nil
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
	} else if inputUTXOSum.Cmp(txOutputSum) < 0 {
		return nil, fmt.Errorf("%s overspend: %s < %s", node.Chain(), inputUTXOSum, txOutputSum)
	}

	// add inputs from UTXOs based off of chain accounting type (account or UTXO)
	var inputs []*types.TxInput
	switch node.GetAccountingType() {
	case types.AccountStructure:
		count, err := pooldb.GetUnspentTransactionCount(c.pooldb.Writer(), node.Chain())
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
		return nil, err
	}

	feeSum := new(big.Int)
	for _, txOutput := range txOutputs {
		feeSum.Add(feeSum, txOutput.Fee)
	}

	tx := &pooldb.Transaction{
		ChainID:      node.Chain(),
		TxID:         txid,
		TxHex:        txHex,
		Value:        dbcl.NullBigInt{Valid: true, BigInt: txOutputSum},
		Fee:          dbcl.NullBigInt{Valid: true, BigInt: feeSum},
		Remainder:    dbcl.NullBigInt{Valid: true, BigInt: remainder},
		RemainderIdx: uint32(len(txOutputs) - 1),
	}

	tx.ID, err = pooldb.InsertTransaction(c.pooldb.Writer(), tx)
	if err != nil {
		return nil, err
	}

	// spend input utxos
	for _, utxo := range inputUTXOs {
		utxo.TransactionID = types.Uint64Ptr(tx.ID)
		err = pooldb.UpdateUTXO(c.pooldb.Writer(), utxo, []string{"transaction_id"})
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
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
	txs, err := pooldb.GetUnspentTransactions(c.pooldb, node.Chain())
	if err != nil {
		return err
	} else if len(txs) > maxTxLimit {
		txs = txs[:maxTxLimit]
	}

	for _, tx := range txs {
		err := c.sendOutgoingTx(tx)
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

	txs, err := pooldb.GetUnconfirmedTransactions(c.pooldb, node.Chain())
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

func (c *Client) sendOutgoingTx(node types.PayoutNode, tx *pooldb.Transaction) error {
	txid, err := node.BroadcastTx(tx.TxHex)
	if err != nil {
		return err
	}

	// if the remainder is non-zero, add the final UTXO
	var outputUTXOs []*pooldb.UTXO
	var outputBalances []*pooldb.BalanceOutput
	if tx.Remainder.Valid && tx.Remainder.BigInt.Cmp(common.Big0) > 0 {
		outputUTXOs = []*pooldb.UTXO{
			&pooldb.UTXO{
				ChainID:       node.Chain(),
				TransactionID: tx.NextTransactionID,
				TxID:          txid,
				Index:         tx.RemainderIdx,
				Value:         tx.Remainder,
			},
		}
	}

	inputUTXOs, err := pooldb.GetUTXOsByTransactionID(c.pooldb.Reader(), tx.ID)
	if err != nil {
		return err
	}

	// spend input utxos
	for _, utxo := range inputUTXOs {
		utxo.Spent = true
		err = pooldb.UpdateUTXO(c.pooldb.Writer(), utxo, []string{"spent"})
		if err != nil {
			return err
		}
	}

	// insert output utxos
	err = pooldb.InsertUTXOs(c.pooldb.Writer(), outputUTXOs...)
	if err != nil {
		return err
	}

	err = pooldb.InsertBalanceOutputs(c.pooldb.Writer(), outputBalances...)
	if err != nil {
		return err
	}

	tx.Spent = true
	if txid != tx.TxID {
		tx.TxID = txid
	}

	err = pooldb.UpdateTransaction(c.pooldb.Writer(), tx, []string{"txid", "spent"})
	if err != nil {
		return err
	}

	return nil
}

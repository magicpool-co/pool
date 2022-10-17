package bank

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) PrepareOutgoingTx(node types.PayoutNode, txOutputs []*types.TxOutput) (*pooldb.Transaction, error) {
	// distributed lock to avoid race conditions
	ctx := context.Background()
	key := "payout:" + strings.ToLower(node.Chain()) + ":prep"
	lock, err := c.locker.Obtain(ctx, key, time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			return nil, err
		}
		return nil, nil
	}
	defer lock.Release(ctx)

	inputUTXOs, err := pooldb.GetUnspentUTXOsByChain(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return nil, err
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
		var count uint32
		/*count, err := pooldb.GetPendingTransactionCount(c.pooldb.Writer(), node.Chain())
		if err != nil {
			return nil, err
		}*/

		// txOutputs count has to be non-zero since output
		// sum has already been verified as non-zero
		inputs = []*types.TxInput{
			&types.TxInput{
				Value:      txOutputs[0].Value,
				FeeBalance: txOutputs[0].FeeBalance,
				Index:      count,
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

func (c *Client) SendOutgoingTx(node types.PayoutNode, tx *pooldb.Transaction) (string, error) {
	// @TODO: check redis to see if the txid has been spent
	txid, err := node.BroadcastTx(tx.TxHex)
	if err != nil {
		return "", err
	} else if txid != tx.TxID {
		// @TODO: do this after we mark the tx as spent
		return "", fmt.Errorf("txid mismatch: have %s, want %s", txid, tx.TxID)
	}

	// if the remainder is non-zero, add the final UTXO
	var outputUTXOs []*pooldb.UTXO
	var outputBalances []*pooldb.BalanceOutput
	if tx.Remainder.Valid && tx.Remainder.BigInt.Cmp(common.Big0) > 0 {
		outputUTXOs = []*pooldb.UTXO{
			&pooldb.UTXO{
				ChainID: node.Chain(),
				TxID:    txid,
				Index:   tx.RemainderIdx,
				Value:   tx.Remainder,
			},
		}
	}

	// @TODO: grab the utxos bound to the tx
	inputUTXOs, err := pooldb.GetUnspentUTXOsByChain(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return "", err
	}

	// spend input utxos
	for _, utxo := range inputUTXOs {
		utxo.Spent = true
		err = pooldb.UpdateUTXO(c.pooldb.Writer(), utxo, []string{"spent"})
		if err != nil {
			return txid, err
		}
	}

	// insert output utxos
	err = pooldb.InsertUTXOs(c.pooldb.Writer(), outputUTXOs...)
	if err != nil {
		return txid, err
	}

	err = pooldb.InsertBalanceOutputs(c.pooldb.Writer(), outputBalances...)
	if err != nil {
		return txid, err
	}

	return txid, nil
}

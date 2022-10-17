package bank

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func PrepareOutgoingTx(node types.PayoutNode, pooldbClient *dbcl.Client, txOutputs []*types.TxOutput) (*pooldb.Transaction, error) {
	inputUTXOs, err := pooldb.GetUnspentUTXOsByChain(pooldbClient.Reader(), node.Chain())
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
		// txOutputs count has to be non-zero since output
		// sum has already been verified as non-zero
		inputs = []*types.TxInput{
			&types.TxInput{
				Value:      txOutputs[0].Value,
				FeeBalance: txOutputs[0].FeeBalance,
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
		err = pooldb.UpdateUTXO(pooldbClient.Writer(), utxo, []string{"transaction_id"})
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func SendOutgoingTx(tx *pooldb.Transaction, node types.PayoutNode, pooldbClient *dbcl.Client) (string, error) {
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
	inputUTXOs, err := pooldb.GetUnspentUTXOsByChain(pooldbClient.Reader(), node.Chain())
	if err != nil {
		return "", err
	}

	// spend input utxos
	for _, utxo := range inputUTXOs {
		utxo.Spent = true
		err = pooldb.UpdateUTXO(pooldbClient.Writer(), utxo, []string{"spent"})
		if err != nil {
			return txid, err
		}
	}

	// insert output utxos
	err = pooldb.InsertUTXOs(pooldbClient.Writer(), outputUTXOs...)
	if err != nil {
		return txid, err
	}

	err = pooldb.InsertBalanceOutputs(pooldbClient.Writer(), outputBalances...)
	if err != nil {
		return txid, err
	}

	return txid, nil
}

func RegisterIncomingTx(node types.PayoutNode, pooldbClient *dbcl.Client, txid string) (bool, error) {
	tx, err := node.GetTx(txid)
	if err != nil {
		return false, err
	} else if tx == nil || !tx.Confirmed {
		return false, nil
	}

	var utxos []*pooldb.UTXO
	switch node.GetAccountingType() {
	case types.AccountStructure:
		utxos = []*pooldb.UTXO{
			&pooldb.UTXO{
				ChainID: node.Chain(),
				TxID:    txid,
				Value:   dbcl.NullBigInt{Valid: true, BigInt: tx.Value},
			},
		}
	case types.UTXOStructure:
		utxos = make([]*pooldb.UTXO, 0)
		for _, output := range tx.Outputs {
			if output.Address != node.Address() {
				continue
			}

			utxo := &pooldb.UTXO{
				ChainID: node.Chain(),
				TxID:    txid,
				Index:   output.Index,
				Value:   dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(output.Value)},
			}
			utxos = append(utxos, utxo)
		}
	}

	err = pooldb.InsertUTXOs(pooldbClient.Writer(), utxos...)
	if err != nil {
		return true, err
	}

	return true, nil
}

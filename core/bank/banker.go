package bank

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func SendOutgoingTx(node types.PayoutNode, pooldbClient *dbcl.Client, txOutputs []*types.TxOutput) (string, error) {
	inputUTXOs, err := pooldb.GetUnspentUTXOsByChain(pooldbClient.Reader(), node.Chain())
	if err != nil {
		return "", err
	}

	// calculate total spendable value
	inputUTXOSum := new(big.Int)
	for _, inputUTXO := range inputUTXOs {
		if !inputUTXO.Value.Valid {
			return "", fmt.Errorf("no value for utxo %d", inputUTXO.ID)
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
		return "", fmt.Errorf("%s empty or negative spend: %s", node.Chain(), txOutputSum)
	} else if inputUTXOSum.Cmp(txOutputSum) < 0 {
		return "", fmt.Errorf("%s overspend: %s < %s", node.Chain(), inputUTXOSum, txOutputSum)
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

	// create and broadcast tx
	tx, err := node.CreateTx(inputs, txOutputs)
	if err != nil {
		return "", err
	}

	txid, err := node.BroadcastTx(tx)
	if err != nil {
		return "", err
	}

	// if the remainder is non-zero, add the final UTXO
	var outputUTXOs []*pooldb.UTXO
	if remainder.Cmp(common.Big0) > 0 {
		outputUTXOs = []*pooldb.UTXO{
			&pooldb.UTXO{
				ChainID: node.Chain(),
				TxID:    txid,
				Index:   uint32(len(txOutputs) - 1),
				Value:   dbcl.NullBigInt{Valid: true, BigInt: remainder},
				Spent:   false,
			},
		}
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

	return txid, nil
}

func ConfirmOutgoingTx(node types.PayoutNode, txid string) {}

func RegisterIncomingTx(node types.PayoutNode, txid string) {}
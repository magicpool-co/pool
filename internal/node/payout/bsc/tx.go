package bsc

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/crypto/tx/ethtx"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetBalance() (*big.Int, error) {
	return node.getBalance(node.address)
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	return nil, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, error) {
	if len(inputs) != 1 || len(outputs) != 1 {
		return "", fmt.Errorf("must have exactly one input and output")
	} else if inputs[0].Value.Cmp(outputs[0].Value) != 0 {
		return "", fmt.Errorf("inputs and outputs must have same value")
	}
	input := inputs[0]
	output := outputs[0]

	nonce, err := node.getPendingNonce(node.address)
	if err != nil {
		return "", err
	}

	chainID, err := node.getChainID()
	if err != nil {
		return "", err
	}

	gasPrice, err := node.getGasPrice()
	if err != nil {
		return "", err
	}

	gasLimit, err := node.sendEstimateGas(node.address, output.Address,
		input.Data, output.Value, gasPrice, nonce)
	if err != nil {
		return "", err
	}

	tx, err := ethtx.NewLegacyTx(node.privKey.ToECDSA(), output.Address, input.Data,
		output.Value, gasPrice, gasLimit, nonce, chainID)

	return tx, err
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.sendRawTransaction(tx)
}

package ae

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/crypto/tx/aetx"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	return nil, nil
}

func (node Node) GetBalance(address string) (*big.Int, error) {
	return node.getBalance(address)
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, error) {
	if len(inputs) != 1 || len(outputs) != 1 {
		return "", fmt.Errorf("must have exactly one input and output")
	} else if inputs[0].Value.Cmp(outputs[0].Value) != 0 {
		return "", fmt.Errorf("inputs and outputs must have same value")
	}
	output := outputs[0]

	nextNonce, err := node.getNextNonce(node.address)
	if err != nil {
		return "", err
	}

	return aetx.NewTx(node.privKey, node.networkID, node.address, output.Address, output.Value, nextNonce)
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.postTransaction(tx)
}

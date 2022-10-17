package ae

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/crypto/tx/aetx"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://explorer.aeternity.io/transactions/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://explorer.aeternity.io/account/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	return node.getBalance(node.address)
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	return nil, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, string, error) {
	if len(inputs) != 1 || len(outputs) != 1 {
		return "", "", fmt.Errorf("must have exactly one input and output")
	} else if inputs[0].Value.Cmp(outputs[0].Value) != 0 {
		return "", "", fmt.Errorf("inputs and outputs must have same value")
	}
	input := inputs[0]
	output := outputs[0]

	nonce, err := node.getNextNonce(node.address)
	if err != nil {
		return "", "", err
	}
	// handle for future nonces
	nonce += uint64(input.Index)

	tx, fee, err := aetx.NewTx(node.privKey, node.networkID, node.address, output.Address, output.Value, nonce)
	if err != nil {
		return "", "", err
	}
	txid := aetx.CalculateTxID(tx)

	output.Value.Sub(output.Value, fee)
	output.Fee = fee

	return txid, tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.postTransaction(tx)
}

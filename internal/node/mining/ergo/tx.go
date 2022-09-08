package ergo

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/types"
)

func (node Node) GetBalance(address string) (*big.Int, error) {
	balance, err := node.getWalletBalances()
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetUint64(balance.Balance), nil
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
	output := outputs[0]

	return node.postWalletPaymentSend(output.Address, output.Value.Uint64())
}

func (node Node) BroadcastTx(txid string) (string, error) {
	return txid, nil
}

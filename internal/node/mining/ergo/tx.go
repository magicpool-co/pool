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
	if len(outputs) == 0 {
		return "", fmt.Errorf("need at least one output")
	}

	addresses := make([]string, len(outputs))
	amounts := make([]uint64, len(outputs))
	for i, output := range outputs {
		addresses[i] = output.Address
		amounts[i] = output.Value.Uint64()
	}

	return node.postWalletPaymentSend(addresses, amounts)
}

func (node Node) BroadcastTx(txid string) (string, error) {
	return txid, nil
}

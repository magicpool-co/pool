package kas

import (
	"math/big"

	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://explorer.kaspa.org/txs/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://explorer.kaspa.org/addresses/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	return new(big.Int), nil
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	return nil, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, string, error) {
	return "", "", nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return "", nil
}

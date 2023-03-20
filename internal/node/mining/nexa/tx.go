package nexa

import (
	"math/big"

	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://explorer.nexa.org/tx/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://explorer.nexa.org/address/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	return new(big.Int), nil
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	tx, err := node.getRawTransaction(txid)
	if err != nil {
		return nil, err
	}

	var height uint64
	var confirmed bool
	if tx.Height > 0 && tx.Confirmations > 0 {
		confirmed = true
		height = uint64(tx.Height)
	}

	res := &types.TxResponse{
		Hash:        txid,
		BlockNumber: height,
		Confirmed:   confirmed,
	}

	return res, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, string, error) {
	return "", "", nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.sendRawTransaction(tx)
}

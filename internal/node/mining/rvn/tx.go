package rvn

import (
	"encoding/hex"
	"math/big"

	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetBalance(address string) (*big.Int, error) {
	// @TODO: need to use an explorer
	return nil, nil
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	return nil, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, error) {
	const feeRate = 2000

	baseTx := btctx.NewTransaction(txVersion, 0, node.prefixP2PKH, node.prefixP2SH)
	rawTx, err := btctx.GenerateTx(node.privKey, baseTx, inputs, outputs, feeRate)
	if err != nil {
		return "", err
	}
	tx := hex.EncodeToString(rawTx)

	return tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.sendRawTransaction(tx)
}

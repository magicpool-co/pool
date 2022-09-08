package flux

import (
	"encoding/hex"
	"fmt"
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
	const feeRate = 0
	const expiryHeight = 21

	height, syncing, err := node.GetStatus()
	if err != nil {
		return "", err
	} else if syncing {
		return "", fmt.Errorf("node is syncing")
	}

	baseTx := btctx.NewTransaction(txVersion, 0, node.prefixP2PKH, node.prefixP2SH)
	baseTx.SetVersionMask(versionMask)
	baseTx.SetVersionGroupID(versionGroupID)
	baseTx.SetExpiryHeight(uint32(height + expiryHeight))

	rawTx, err := btctx.GenerateRawTx(baseTx, inputs, outputs, feeRate)
	if err != nil {
		return "", err
	}

	rawTxSerialized, err := rawTx.Serialize(nil)
	if err != nil {
		return "", err
	}

	return node.signRawTransaction(hex.EncodeToString(rawTxSerialized), node.wif)
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.sendRawTransaction(tx)
}

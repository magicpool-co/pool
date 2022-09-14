package banker

import (
	"github.com/magicpool-co/pool/types"
)

func SendOutgoingTx(node types.PayoutNode, inputs []*types.TxInput, outputs []*types.TxOutput) (string, error) {
	tx, err := node.CreateTx(inputs, outputs)
	if err != nil {
		return "", err
	}

	txid, err := node.BroadcastTx(tx)
	if err != nil {
		return "", err
	}

	return txid, nil
}

func ConfirmOutgoingTx(node types.PayoutNode, txid string) {}

func RegisterIncomingTx(node types.PayoutNode, txid string) {}

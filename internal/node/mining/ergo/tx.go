package ergo

import (
	"encoding/hex"
	"fmt"
	"math/big"

	txCommon "github.com/magicpool-co/pool/pkg/crypto/tx"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://explorer.ergoplatform.com/en/transactions/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://explorer.ergoplatform.com/en/addresses/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	balance, err := node.getWalletBalances()
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetUint64(balance.Balance), nil
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	tx, err := node.getWalletTransactionByID(txid)
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
	if len(outputs) == 0 {
		return "", "", fmt.Errorf("need at least one output")
	}

	const fee = 1000000
	err := txCommon.DistributeFees(inputs, outputs, fee, false)
	if err != nil {
		return "", "", err
	}

	addresses := make([]string, len(outputs))
	amounts := make([]uint64, len(outputs))
	for i, output := range outputs {
		addresses[i] = output.Address
		amounts[i] = output.Value.Uint64()
	}

	txBytes, err := node.postWalletTransactionGenerate(addresses, amounts, fee)
	if err != nil {
		return "", "", err
	}
	tx := hex.EncodeToString(txBytes)

	txid, err := node.postWalletTransactionCheck(txBytes)
	if err != nil {
		return "", "", err
	}

	return txid, tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	txBytes, err := hex.DecodeString(tx)
	if err != nil {
		return "", err
	}

	return node.postWalletTransactionSend(txBytes)
}

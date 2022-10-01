package etc

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/tx/ethtx"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://blockscout.com/etc/mainnet/tx/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://blockscout.com/etc/mainnet/address/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	return node.getBalance(node.address)
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	tx, err := node.getTransactionByHash(txid)
	if err != nil || tx == nil || len(tx.BlockNumber) <= 2 {
		return nil, err
	}

	txReceipt, err := node.getTransactionReceipt(txid)
	if err != nil {
		return nil, err
	}

	height, err := common.HexToUint64(tx.BlockNumber)
	if err != nil {
		return nil, err
	}

	value, err := common.HexToBig(tx.Value)
	if err != nil {
		return nil, err
	}

	gasUsed, err := common.HexToBig(txReceipt.GasUsed)
	if err != nil {
		return nil, err
	}

	gasTotal, err := common.HexToBig(tx.Gas)
	if err != nil {
		return nil, err
	}

	gasPrice, err := common.HexToBig(tx.GasPrice)
	if err != nil {
		return nil, err
	}

	confirmed := txReceipt.Status == "0x1"
	gasLeftover := new(big.Int).Sub(gasTotal, gasUsed)
	fees := new(big.Int).Mul(gasUsed, gasPrice)
	feeBalance := new(big.Int).Mul(gasLeftover, gasPrice)

	res := &types.TxResponse{
		Hash:        tx.Hash,
		BlockNumber: height,
		Value:       value,
		Fee:         fees,
		FeeBalance:  feeBalance,
		Confirmed:   confirmed,
	}

	return res, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, error) {
	if len(inputs) != 1 || len(outputs) != 1 {
		return "", fmt.Errorf("must have exactly one input and output")
	} else if inputs[0].Value.Cmp(outputs[0].Value) != 0 {
		return "", fmt.Errorf("inputs and outputs must have same value")
	}
	input := inputs[0]
	output := outputs[0]

	chainID, err := node.getChainID()
	if err != nil {
		return "", err
	}

	nonce, err := node.getPendingNonce(node.address)
	if err != nil {
		return "", err
	}

	gasLimit, err := node.sendEstimateGas(node.address, output.Address)
	if err != nil {
		return "", err
	} else if gasLimit != 21000 {
		gasLimit += 30000
	}

	gasPrice, err := node.getGasPrice()
	if err != nil {
		return "", err
	}

	tx, _, err := ethtx.NewLegacyTx(node.privKey.ToECDSA(), output.Address, nil,
		input.Value, gasPrice, gasLimit, nonce, chainID)

	return tx, err
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.sendRawTransaction(tx)
}

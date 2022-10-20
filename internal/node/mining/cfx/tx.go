package cfx

import (
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/tx/cfxtx"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://www.confluxscan.io/tx/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://www.confluxscan.io/address/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	return node.getBalance(node.address)
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	tx, err := node.getTransactionByHash(txid)
	if err != nil {
		return nil, err
	}

	var height uint64
	var confirmed bool
	if len(tx.EpochHeight) > 0 {
		confirmed = true
		height, err = common.HexToUint64(tx.EpochHeight)
		if err != nil {
			return nil, err
		}
	}

	res := &types.TxResponse{
		Hash:        tx.Hash,
		BlockNumber: height,
		Confirmed:   confirmed,
	}

	return res, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, string, error) {
	if len(inputs) != 1 || len(outputs) != 1 {
		return "", "", fmt.Errorf("must have exactly one input and output")
	} else if inputs[0].Value.Cmp(outputs[0].Value) != 0 {
		return "", "", fmt.Errorf("inputs and outputs must have same value")
	}
	input := inputs[0]
	output := outputs[0]

	if len(output.Address) > 2 && output.Address[:2] == "0x" {
		var err error
		rawAddress := ethCommon.HexToAddress(output.Address).Bytes()
		output.Address, err = ETHAddressToCFX(rawAddress, node.networkPrefix)
		if err != nil {
			return "", "", err
		}
	}

	nonce, err := node.getPendingNonce(node.address)
	if err != nil {
		return "", "", err
	}
	// handle for future nonces
	nonce += uint64(input.Index)

	epochNumber, err := node.getEpochNumber()
	if err != nil {
		return "", "", err
	}

	gasPrice, err := node.getGasPrice()
	if err != nil {
		return "", "", err
	}

	gasLimit, storageLimit, err := node.sendEstimateGas(node.address,
		output.Address, input.Data, output.Value, gasPrice, nonce)
	if err != nil {
		return "", "", err
	}

	tx, fee, err := cfxtx.NewTx(node.privKey.ToECDSA(), output.Address, input.Data,
		output.Value, gasPrice, gasLimit, storageLimit, nonce, node.networkID, epochNumber)
	if err != nil {
		return "", "", err
	}
	txid := cfxtx.CalculateTxID(tx)

	output.Value.Sub(output.Value, fee)
	output.Fee = fee

	return txid, tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.sendRawTransaction(tx)
}

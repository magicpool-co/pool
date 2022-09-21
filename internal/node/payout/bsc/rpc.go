package bsc

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/common"
)

func (node Node) getBalance(address string) (*big.Int, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_getBalance", address, "latest")
	if err != nil {
		return nil, err
	}

	var hexBalance string
	err = json.Unmarshal(res.Result, &hexBalance)
	if err != nil {
		return nil, err
	}

	return common.HexToBig(hexBalance)
}

func (node Node) getGasPrice() (*big.Int, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_gasPrice")
	if err != nil {
		return nil, err
	}

	var hexPrice string
	err = json.Unmarshal(res.Result, &hexPrice)
	if err != nil {
		return nil, err
	}

	return common.HexToBig(hexPrice)
}

func (node Node) getChainID() (uint64, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_chainId")
	if err != nil {
		return 0, err
	}

	var hexID string
	err = json.Unmarshal(res.Result, &hexID)
	if err != nil {
		return 0, err
	}

	return common.HexToUint64(hexID)
}

func (node Node) getPendingNonce(address string) (uint64, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_getTransactionCount", address, "pending")
	if err != nil {
		return 0, err
	}

	var hexNonce string
	err = json.Unmarshal(res.Result, &hexNonce)
	if err != nil {
		return 0, err
	}

	return common.HexToUint64(hexNonce)
}

func (node Node) sendEstimateGas(from, to string, data []byte, amount, gasPrice *big.Int, nonce uint64) (uint64, error) {
	tx := map[string]interface{}{
		"from":     from,
		"to":       to,
		"value":    "0x" + fmt.Sprintf("%x", amount),
		"data":     "0x" + hex.EncodeToString(data),
		"gasPrice": "0x" + fmt.Sprintf("%x", gasPrice),
		"nonce":    "0x" + fmt.Sprintf("%x", nonce),
	}
	res, err := node.rpcHost.ExecRPCFromArgs("eth_estimateGas", tx)
	if err != nil {
		return 0, err
	}

	var hexEstimate string
	err = json.Unmarshal(res.Result, &hexEstimate)
	if err != nil {
		return 0, err
	}

	return common.HexToUint64(hexEstimate)
}

func (node Node) sendRawTransaction(tx string) (string, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_sendRawTransaction", tx)
	if err != nil {
		return "", err
	}

	var txid string
	if err := json.Unmarshal(res.Result, &txid); err != nil {
		return "", err
	}

	return txid, nil
}

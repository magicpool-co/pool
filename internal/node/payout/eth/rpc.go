package eth

import (
	"fmt"
	"math/big"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func (node Node) getTransactionByHash(txid string) (*Transaction, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_getTransactionByHash", txid)
	if err != nil {
		return nil, err
	}

	tx := new(Transaction)
	if err := json.Unmarshal(res.Result, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (node Node) getTransactionReceipt(txid string) (*TransactionReceipt, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_getTransactionReceipt", txid)
	if err != nil {
		return nil, err
	}

	receipt := new(TransactionReceipt)
	if err := json.Unmarshal(res.Result, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}

func (node Node) getTransactionReceiptMany(txids []string) ([]*TransactionReceipt, error) {
	reqs := make([]*rpc.Request, len(txids))
	var err error
	for i, txid := range txids {
		reqs[i], err = rpc.NewRequestWithID(i, "eth_getTransactionReceipt", txid)
		if err != nil {
			return nil, err
		}
	}

	responses, err := node.rpcHost.ExecRPCBulk(reqs)
	if err != nil {
		return nil, err
	} else if len(responses) != len(reqs) {
		return nil, fmt.Errorf("request and response length mismatch: %d and %d", len(responses), len(reqs))
	}

	receipts := make([]*TransactionReceipt, len(responses))
	for i, res := range responses {
		err := json.Unmarshal(res.Result, &receipts[i])
		if err != nil {
			return nil, err
		}
	}

	return receipts, nil
}

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

func (node Node) getChainID() (uint64, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_chainId")
	if err != nil {
		return 0, err
	}

	var hexID string
	if err := json.Unmarshal(res.Result, &hexID); err != nil {
		return 0, err
	}

	return common.HexToUint64(hexID)
}

func (node Node) getGasPrice() (*big.Int, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_gasPrice")
	if err != nil {
		return nil, err
	}

	var hexPrice string
	if err := json.Unmarshal(res.Result, &hexPrice); err != nil {
		return nil, err
	}

	return common.HexToBig(hexPrice)
}

func (node Node) getPendingNonce(address string) (uint64, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_getTransactionCount", address, "pending")
	if err != nil {
		return 0, err
	}

	var hexNonce string
	if err := json.Unmarshal(res.Result, &hexNonce); err != nil {
		return 0, err
	}

	return common.HexToUint64(hexNonce)
}

func (node Node) getBlockNumber() (uint64, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_blockNumber")
	if err != nil {
		return 0, err
	}

	var rawHeight string
	if err := json.Unmarshal(res.Result, &rawHeight); err != nil {
		return 0, err
	}

	return common.HexToUint64(rawHeight)
}

func (node Node) getBlockByNumber(height uint64) (*Block, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_getBlockByNumber", common.Uint64ToHex(height), false)
	if err != nil {
		return nil, err
	}

	block := new(Block)
	if err := json.Unmarshal(res.Result, block); err != nil {
		return nil, err
	}

	return block, nil
}

func (node Node) sendEstimateGas(from, to string) (uint64, error) {
	tx := map[string]interface{}{"from": from, "to": to}
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

func (node Node) sendCall(params []interface{}) (*big.Int, error) {
	res, err := node.rpcHost.ExecRPCFromArgs("eth_call", params...)
	if err != nil {
		return nil, err
	}

	var result string
	if err := json.Unmarshal(res.Result, &result); err != nil {
		return nil, err
	}

	return common.HexToBig(result)
}

func (node Node) sendRawTransaction(tx string) (string, error) {
	res, err := node.rpcHost.ExecRPCFromArgsOnce("eth_sendRawTransaction", tx)
	if err != nil {
		return "", err
	}

	var txid string
	if err := json.Unmarshal(res.Result, &txid); err != nil {
		return "", err
	}

	return txid, nil
}

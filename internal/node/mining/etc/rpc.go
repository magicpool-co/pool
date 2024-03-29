package etc

import (
	"fmt"
	"math/big"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/node/mining/etc/mock"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func (node Node) getTransactionByHash(txid string) (*Transaction, error) {
	res, err := node.rpcHost.ExecRPCFromArgsSynced("eth_getTransactionByHash", txid)
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
	res, err := node.rpcHost.ExecRPCFromArgsSynced("eth_getTransactionReceipt", txid)
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
	var responses []*rpc.Response
	if node.mocked {
		responses = mock.GetTransactionReceiptMany(txids)
	} else {
		reqs := make([]*rpc.Request, len(txids))
		var err error
		for i, txid := range txids {
			reqs[i], err = rpc.NewRequestWithID(i, "eth_getTransactionReceipt", txid)
			if err != nil {
				return nil, err
			}
		}

		responses, err = node.rpcHost.ExecRPCBulk(reqs)
		if err != nil {
			return nil, err
		} else if len(responses) != len(reqs) {
			return nil, fmt.Errorf("request and response length mismatch: %d and %d", len(responses), len(reqs))
		}
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

func (node Node) getBlockByNumberMany(heights []uint64) ([]*Block, error) {
	var responses []*rpc.Response
	if node.mocked {
		responses = mock.GetBlockByNumberMany(heights)
	} else {
		reqs := make([]*rpc.Request, len(heights))
		var err error
		for i, height := range heights {
			reqs[i], err = rpc.NewRequestWithID(i, "eth_getBlockByNumber", common.Uint64ToHex(height), true)
			if err != nil {
				return nil, err
			}
		}

		responses, err = node.rpcHost.ExecRPCBulk(reqs)
		if err != nil {
			return nil, err
		} else if len(responses) != len(reqs) {
			return nil, fmt.Errorf("request and response length mismatch: %d and %d", len(responses), len(reqs))
		}
	}

	blocks := make([]*Block, len(responses))
	for i, res := range responses {
		err := json.Unmarshal(res.Result, &blocks[i])
		if err != nil {
			return nil, err
		}
	}

	return blocks, nil
}

func (node Node) getBalance(address string) (*big.Int, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetBalance(address)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("eth_getBalance", address, "latest")
		if err != nil {
			return nil, err
		}
	}

	var hexBalance string
	err = json.Unmarshal(res.Result, &hexBalance)
	if err != nil {
		return nil, err
	}

	return common.HexToBig(hexBalance)
}

func (node Node) getChainID() (uint64, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetChainID()
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("eth_chainId")
		if err != nil {
			return 0, err
		}
	}

	var hexID string
	if err := json.Unmarshal(res.Result, &hexID); err != nil {
		return 0, err
	}

	return common.HexToUint64(hexID)
}

func (node Node) getGasPrice() (*big.Int, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetGasPrice()
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("eth_gasPrice")
		if err != nil {
			return nil, err
		}
	}

	var hexPrice string
	if err := json.Unmarshal(res.Result, &hexPrice); err != nil {
		return nil, err
	}

	return common.HexToBig(hexPrice)
}

func (node Node) getPendingNonce(address string) (uint64, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetPendingNonce(address)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("eth_getTransactionCount", address, "pending")
		if err != nil {
			return 0, err
		}
	}

	var hexNonce string
	if err := json.Unmarshal(res.Result, &hexNonce); err != nil {
		return 0, err
	}

	return common.HexToUint64(hexNonce)
}

func (node Node) getBlockNumber(hostID string) (uint64, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetBlockNumber()
	} else {
		if hostID == "" {
			res, err = node.rpcHost.ExecRPCFromArgsSynced("eth_blockNumber")
		} else {
			req, err := rpc.NewRequestWithHostID(hostID, "eth_blockNumber")
			if err != nil {
				return 0, err
			}

			res, err = node.rpcHost.ExecRPC(req)
		}

		if err != nil {
			return 0, err
		}
	}

	var rawHeight string
	if err := json.Unmarshal(res.Result, &rawHeight); err != nil {
		return 0, err
	}

	return common.HexToUint64(rawHeight)
}

func (node Node) getSyncing(hostID string) (bool, error) {
	var res *rpc.Response
	if node.mocked {
		res = mock.GetSyncing()
	} else {
		req, err := rpc.NewRequestWithHostID(hostID, "eth_syncing")
		if err != nil {
			return false, err
		}

		res, err = node.rpcHost.ExecRPC(req)
		if err != nil {
			return false, err
		}
	}

	var syncing bool
	if err := json.Unmarshal(res.Result, &syncing); err != nil {
		return true, nil
	}

	return syncing, nil
}

func (node Node) getBlockByNumber(height uint64) (*Block, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetBlockByNumber(height)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("eth_getBlockByNumber", common.Uint64ToHex(height), true)
		if err != nil {
			return nil, err
		}
	}

	block := new(Block)
	if err := json.Unmarshal(res.Result, block); err != nil {
		return nil, err
	}

	return block, nil
}

func (node Node) getUncleByNumberAndIndex(height, index uint64) (*Block, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetUncleByNumberAndIndex(height, index)
	} else {
		params := []interface{}{common.Uint64ToHex(height), common.Uint64ToHex(index)}
		res, err = node.rpcHost.ExecRPCFromArgsSynced("eth_getUncleByBlockNumberAndIndex", params...)
		if err != nil {
			return nil, err
		}
	}

	uncle := new(Block)
	if err := json.Unmarshal(res.Result, uncle); err != nil {
		return nil, err
	}

	return uncle, nil
}

func (node Node) getWork() (string, []string, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetWork()
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("eth_getWork")
		if err != nil {
			return "", nil, err
		}
	}

	var result []string
	if err := json.Unmarshal(res.Result, &result); err != nil {
		return "", nil, err
	} else if len(result) < 4 {
		return "", nil, fmt.Errorf("invalid getwork response: %v", result)
	}

	return res.HostID, result, nil
}

func (node Node) sendEstimateGas(from, to string) (uint64, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.SendEstimateGas(from, to)
	} else {
		tx := map[string]interface{}{"from": from, "to": to}
		res, err = node.rpcHost.ExecRPCFromArgsSynced("eth_estimateGas", tx)
		if err != nil {
			return 0, err
		}
	}

	var hexEstimate string
	err = json.Unmarshal(res.Result, &hexEstimate)
	if err != nil {
		return 0, err
	}

	return common.HexToUint64(hexEstimate)
}

func (node Node) sendSubmitWork(hostID, nonce, hash, mixDigest string) (bool, error) {
	var res *rpc.Response
	if node.mocked {
		res = mock.SendSubmitWork(nonce, hash, mixDigest)
	} else {
		req, err := rpc.NewRequestWithHostID(hostID, "eth_submitWork", nonce, hash, mixDigest)
		if err != nil {
			return false, err
		}

		res, err = node.rpcHost.ExecRPC(req)
		if err != nil {
			return false, err
		}
	}

	var accepted bool
	if err := json.Unmarshal(res.Result, &accepted); err != nil {
		return false, err
	}

	return accepted, nil
}

func (node Node) sendRawTransaction(tx string) (string, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.SendRawTransaction(tx)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("eth_sendRawTransaction", tx)
		if err != nil {
			return "", err
		}
	}

	var txid string
	if err := json.Unmarshal(res.Result, &txid); err != nil {
		return "", err
	}

	return txid, nil
}

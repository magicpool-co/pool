package cfx

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/node/mining/cfx/mock"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func (node Node) getBalance(address string) (*big.Int, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetBalance(address)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgs("cfx_getBalance", address, "latest_state")
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

func (node Node) getLatestBlock(hostID string) (*Block, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetLatestBlock()
	} else {
		res, err = node.rpcHost.ExecRPCFromArgs("cfx_getBlockByEpochNumber", "latest_confirmed", false)
		if err != nil {
			return nil, err
		}
	}

	var block *Block
	if err := json.Unmarshal(res.Result, &block); err != nil {
		return nil, err
	} else if block == nil {
		return nil, fmt.Errorf("nil latest block")
	}

	return block, nil
}

func (node Node) getBlockRewardInfo(epochHeight uint64) ([]*BlockRewardInfo, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetBlockRewardInfo(epochHeight)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgs("cfx_getBlockRewardInfo", common.Uint64ToHex(epochHeight))
		if err != nil {
			return nil, err
		}
	}

	blockRewards := make([]*BlockRewardInfo, 0)
	if err := json.Unmarshal(res.Result, &blockRewards); err != nil {
		return nil, err
	} else if len(blockRewards) == 0 {
		// use the conflux URL as a backup (since it is an archive node and
		// will store block rewards permanently, whereas our full node
		// only stores for around a day)
		req, err := rpc.NewRequest("cfx_getBlockRewardInfo", common.Uint64ToHex(epochHeight))
		if err != nil {
			return nil, err
		}

		err = node.execRPCfromFallback(req, &blockRewards)
		if err != nil {
			return nil, err
		} else if len(blockRewards) == 0 {
			return nil, fmt.Errorf("no block rewards found")
		}
	}

	return blockRewards, nil
}

func (node Node) getBlockRewardInfoMany(epochHeights []uint64) ([][]*BlockRewardInfo, error) {
	var responses []*rpc.Response
	if node.mocked {
		responses = mock.GetBlockRewardInfoMany(epochHeights)
	} else {
		reqs := make([]*rpc.Request, len(epochHeights))
		var err error
		for i, height := range epochHeights {
			reqs[i], err = rpc.NewRequestWithID(i, "cfx_getBlockRewardInfo", common.Uint64ToHex(height))
			if err != nil {
				return nil, err
			}
		}

		// responses, err = node.rpcHost.ExecRPCBulk(reqs)
		responses, err = node.execRPCfromFallbackBulk(reqs)
		if err != nil {
			return nil, err
		} else if len(responses) != len(reqs) {
			return nil, fmt.Errorf("request and response length mismatch: %d and %d", len(responses), len(reqs))
		}
	}

	rewardsList := make([][]*BlockRewardInfo, len(responses))
	for i, res := range responses {
		err := json.Unmarshal(res.Result, &rewardsList[i])
		if err != nil {
			return nil, err
		}
	}

	return rewardsList, nil
}

func (node Node) getBlockByHash(blockHash string) (*Block, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetBlockByHash(blockHash)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgs("cfx_getBlockByHash", blockHash, false)
		if err != nil {
			return nil, err
		}
	}

	var block *Block
	err = json.Unmarshal(res.Result, &block)
	if err != nil {
		return nil, err
	} else if block == nil {
		// use the conflux URL as a backup (since it is an archive node and
		// will store block rewards permanently, whereas our full node
		// only stores for around a day)
		req, err := rpc.NewRequest("cfx_getBlockByHash", blockHash, false)
		if err != nil {
			return nil, err
		}

		err = node.execRPCfromFallback(req, &block)
		if err != nil {
			return nil, err
		} else if block == nil {
			return nil, fmt.Errorf("block not found")
		}
	}

	return block, nil
}

func (node Node) getBlockByHashMany(blockHashes []string) ([]*Block, error) {
	reqs := make([]*rpc.Request, len(blockHashes))
	var responses []*rpc.Response
	if node.mocked {
		responses = mock.GetBlockByHashMany(blockHashes)
	} else {
		var err error
		for i, blockHash := range blockHashes {
			reqs[i], err = rpc.NewRequestWithID(i, "cfx_getBlockByHash", blockHash, false)
			if err != nil {
				return nil, err
			}
		}

		responses, err = node.execRPCfromFallbackBulk(reqs)
		// responses, err = node.rpcHost.ExecRPCBulk(reqs)
		if err != nil {
			return nil, err
		} else if len(responses) != len(reqs) {
			return nil, fmt.Errorf("request and response length mismatch: %d and %d", len(responses), len(reqs))
		}
	}

	blocks := make([]*Block, len(responses))
	for i, res := range responses {
		if res.Error == nil {
			err := json.Unmarshal(res.Result, &blocks[i])
			if err != nil {
				return nil, err
			}
		}

		if blocks[i] == nil {
			// do fallback synchronously to not get the IP blacklisted
			if err := node.execRPCfromFallback(reqs[i], &blocks[i]); err != nil {
				return nil, err
			} else if blocks[i] == nil {
				return nil, fmt.Errorf("unable to find block")
			}
		}
	}

	return blocks, nil
}

func (node Node) getBlocksByEpochMany(epochHeights []uint64) ([][]string, error) {
	var responses []*rpc.Response
	if node.mocked {
		responses = mock.GetBlocksByEpochMany(epochHeights)
	} else {
		reqs := make([]*rpc.Request, len(epochHeights))
		var err error
		for i, height := range epochHeights {
			reqs[i], err = rpc.NewRequestWithID(i, "cfx_getBlocksByEpoch", common.Uint64ToHex(height))
			if err != nil {
				return nil, err
			}
		}

		responses, err = node.execRPCfromFallbackBulk(reqs)
		// responses, err = node.rpcHost.ExecRPCBulk(reqs)
		if err != nil {
			return nil, err
		} else if len(responses) != len(reqs) {
			return nil, fmt.Errorf("request and response length mismatch: %d and %d", len(responses), len(reqs))
		}
	}

	hashesList := make([][]string, len(responses))
	for i, res := range responses {
		err := json.Unmarshal(res.Result, &hashesList[i])
		if err != nil {
			return nil, err
		}
	}

	return hashesList, nil
}

func (node Node) getGasPrice() (*big.Int, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetGasPrice()
	} else {
		res, err = node.rpcHost.ExecRPCFromArgs("cfx_gasPrice")
		if err != nil {
			return nil, err
		}
	}

	var hexPrice string
	err = json.Unmarshal(res.Result, &hexPrice)
	if err != nil {
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
		res, err = node.rpcHost.ExecRPCFromArgs("cfx_getNextNonce", address, "latest_state")
		if err != nil {
			return 0, err
		}
	}

	var hexNonce string
	err = json.Unmarshal(res.Result, &hexNonce)
	if err != nil {
		return 0, err
	}

	return common.HexToUint64(hexNonce)
}

func (node Node) getEpochNumber() (uint64, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetEpochNumber()
	} else {
		res, err = node.rpcHost.ExecRPCFromArgs("cfx_epochNumber", "latest_state")
		if err != nil {
			return 0, err
		}
	}

	var hexEpochNumber string
	err = json.Unmarshal(res.Result, &hexEpochNumber)
	if err != nil {
		return 0, err
	}

	return common.HexToUint64(hexEpochNumber)
}

func (node Node) sendEstimateGas(from, to string, data []byte, amount, gasPrice *big.Int, nonce uint64) (uint64, uint64, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.SendEstimateGas(from, to)
	} else {
		params := []interface{}{
			map[string]interface{}{
				"from":     from,
				"to":       to,
				"value":    "0x" + fmt.Sprintf("%x", amount),
				"data":     "0x" + hex.EncodeToString(data),
				"gasPrice": "0x" + fmt.Sprintf("%x", gasPrice),
				"nonce":    "0x" + fmt.Sprintf("%x", nonce),
			},
		}
		res, err = node.rpcHost.ExecRPCFromArgs("cfx_estimateGasAndCollateral", params...)
		if err != nil {
			return 0, 0, err
		}
	}

	output := new(struct {
		GasUsed               string `json:"gasUsed"`
		StorageCollateralized string `json:"storageCollateralized"`
	})
	err = json.Unmarshal(res.Result, &output)
	if err != nil {
		return 0, 0, err
	}

	gasUsed, err := common.HexToUint64(output.GasUsed)
	if err != nil {
		return 0, 0, err
	}

	storageLimit, err := common.HexToUint64(output.StorageCollateralized)
	if err != nil {
		return 0, 0, err
	}

	return gasUsed, storageLimit, nil
}

func (node Node) sendRawTransaction(tx string) (string, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.SendRawTransaction(tx)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsOnce("cfx_sendRawTransaction", tx)
		if err != nil {
			return "", err
		}
	}

	var txid string
	err = json.Unmarshal(res.Result, &txid)

	return txid, err
}

func (node Node) miningSubscribe() chan *rpc.Request {
	if node.mocked {
		return mock.MiningSubscribe()
	}

	return node.tcpHost.Subscribe("mining.notify")
}

func (node Node) submitBlock(hostID, nonce, hash string) (bool, error) {
	var res *rpc.Response
	if node.mocked {
		res = mock.SubmitBlock(hostID, nonce, hash)
	} else {
		req, err := rpc.NewRequestWithHostID(hostID, "mining.submit", "x", hash, nonce, hash)
		if err != nil {
			return false, err
		}

		res, err = node.tcpHost.Exec(req)
		if err != nil {
			return false, err
		}
	}

	var output []interface{}
	err := json.Unmarshal(res.Result, &output)
	if err != nil {
		return false, err
	}

	var accepted bool
	if len(output) > 0 {
		accepted, _ = output[0].(bool)
	}

	if accepted {
		return true, nil
	} else if len(output) > 1 {
		msg, ok := output[1].(string)
		if ok && strings.Index(msg, "Solution for a stale job!") != -1 {
			return false, fmt.Errorf("stale share: not found")
		}
		return false, fmt.Errorf("block not accepted: %v", output[1])
	}

	return false, fmt.Errorf("block not accepted")
}

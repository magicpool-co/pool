package flux

import (
	"fmt"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/node/mining/flux/mock"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func (node Node) getBlockchainInfo(hostID string) (*BlockchainInfo, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetBlockchainInfo()
	} else {
		if hostID == "" {
			res, err = node.rpcHost.ExecRPCFromArgsSynced("getblockchaininfo")
		} else {
			req, err := rpc.NewRequestWithHostID(hostID, "getblockchaininfo")
			if err != nil {
				return nil, err
			}

			res, err = node.rpcHost.ExecRPC(req)
		}

		if err != nil {
			return nil, err
		}
	}

	info := new(BlockchainInfo)
	if err := json.Unmarshal(res.Result, info); err != nil {
		return nil, err
	}

	return info, nil
}

func (node Node) getRawTransaction(txid string) (*Transaction, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		// @TODO
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("getrawtransaction", txid, 1)
		if err != nil {
			return nil, err
		}
	}

	tx := new(Transaction)
	if err := json.Unmarshal(res.Result, tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (node Node) getBlockHash(height uint64) (string, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetBlockHash(height)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("getblockhash", height)
		if err != nil {
			return "", err
		}
	}

	var hash string
	if err := json.Unmarshal(res.Result, &hash); err != nil {
		return "", err
	}

	return hash, nil
}

func (node Node) getBlockHashMany(heights []uint64) ([]string, error) {
	var responses []*rpc.Response
	if node.mocked {
		responses = mock.GetBlockHashMany(heights)
	} else {
		reqs := make([]*rpc.Request, len(heights))
		var err error
		for i, height := range heights {
			reqs[i], err = rpc.NewRequestWithID(i, "getblockhash", height)
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

	hashes := make([]string, len(responses))
	for i, res := range responses {
		err := json.Unmarshal(res.Result, &hashes[i])
		if err != nil {
			return nil, err
		}
	}

	return hashes, nil
}

func (node Node) getBlock(hash string) (*Block, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetBlock(hash)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("getblock", hash, 2)
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

func (node Node) getBlockMany(hashes []string) ([]*Block, error) {
	var responses []*rpc.Response
	if node.mocked {
		responses = mock.GetBlockMany(hashes)
	} else {
		reqs := make([]*rpc.Request, len(hashes))
		var err error
		for i, hash := range hashes {
			reqs[i], err = rpc.NewRequestWithID(i, "getblock", hash, 2)
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

func (node Node) getBlockTemplate() (string, *BlockTemplate, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.GetBlockTemplate()
	} else {
		capabilities := map[string]interface{}{
			"capabilities": []string{"coinbasetx", "workid", "coinbase/append"},
		}
		res, err = node.rpcHost.ExecRPCFromArgsSynced("getblocktemplate", capabilities)
		if err != nil {
			return "", nil, err
		}
	}

	template := new(BlockTemplate)
	if err := json.Unmarshal(res.Result, template); err != nil {
		return "", nil, err
	} else if len(template.PreviousBlockHash) == 0 {
		return "", nil, fmt.Errorf("invalid getblocktemplate response")
	}

	return res.HostID, template, nil
}

func (node Node) submitBlock(hostID, block string) error {
	var res *rpc.Response
	if node.mocked {
		res = mock.SubmitBlock(hostID, block)
	} else {
		req, err := rpc.NewRequestWithHostID(hostID, "submitblock", block)
		if err != nil {
			return err
		}

		res, err = node.rpcHost.ExecRPC(req)
		if err != nil {
			return err
		}
	}

	var result string
	if err := json.Unmarshal(res.Result, &result); err != nil {
		return err
	} else if len(result) > 0 {
		return fmt.Errorf("submit block error: %s", result)
	}

	return nil
}

func (node Node) signRawTransaction(tx, wif string) (string, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		// @TODO
	} else {
		res, err = node.rpcHost.ExecRPCFromArgs("signrawtransaction", tx, []string{}, []string{wif})
		if err != nil {
			return "", err
		}
	}

	signedTx := new(SignedTransaction)
	if err := json.Unmarshal(res.Result, signedTx); err != nil {
		return "", err
	} else if len(signedTx.Errors) > 0 {
		return "", signedTx.Errors[0]
	}

	return signedTx.Hex, nil
}

func (node Node) sendRawTransaction(tx string) (string, error) {
	var res *rpc.Response
	var err error
	if node.mocked {
		res = mock.SendRawTransaction(tx)
	} else {
		res, err = node.rpcHost.ExecRPCFromArgsSynced("sendrawtransaction", tx)
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

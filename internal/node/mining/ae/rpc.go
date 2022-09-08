package ae

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/magicpool-co/pool/internal/node/mining/ae/mock"
)

func (node Node) getNextNonce(address string) (uint64, error) {
	var result map[string]json.RawMessage
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetNextNonce(address), &result)
	} else {
		err = node.externalHost.ExecHTTP("GET", "/v2/accounts/"+address+"/next-nonce", nil, &result)
	}

	if err != nil {
		return 0, err
	} else if reason, ok := result["reason"]; ok {
		return 0, fmt.Errorf("failed to fetch next nonce: %s", reason)
	}

	var nonce uint64
	err = json.Unmarshal(result["next_nonce"], &nonce)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (node Node) getBalance(address string) (*big.Int, error) {
	var result map[string]json.RawMessage
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetBalance(address), &result)
	} else {
		err = node.externalHost.ExecHTTP("GET", "/v2/accounts/"+address, nil, &result)
	}

	if err != nil {
		return nil, err
	} else if reason, ok := result["reason"]; ok {
		return nil, fmt.Errorf("failed to fetch next nonce: %s", reason)
	}

	balance, ok := new(big.Int).SetString(string(result["balance"]), 10)
	if !ok {
		return nil, fmt.Errorf("unable to parse balance")
	}

	return balance, nil
}

func (node Node) getStatus() (*Status, error) {
	var status *Status
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetStatus(), &status)
	} else {
		err = node.externalHost.ExecHTTP("GET", "/v2/status", nil, &status)
	}

	return status, err
}

func (node Node) getBlock(height uint64) (*Block, error) {
	var block *Block
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetBlock(height), &block)
	} else {
		strHeight := strconv.FormatUint(height, 10)
		err = node.externalHost.ExecHTTP("GET", "/v2/key-blocks/height/"+strHeight, nil, &block)
	}

	return block, err
}

func (node Node) getPendingBlock() (string, *Block, error) {
	var hostID string
	var block *Block
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetPendingBlock(), &block)
	} else {
		hostID, err = node.externalHost.ExecHTTPSticky("", "GET", "/v2/key-blocks/pending", nil, &block)
	}

	return hostID, block, err
}

func (node Node) postBlock(hostID string, block interface{}) error {
	var result map[string]json.RawMessage
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.PostBlock(hostID, block), &result)
	} else {
		_, err = node.internalHost.ExecHTTPSticky(hostID, "POST", "/v2/key-blocks", block, &result)
	}

	if err != nil {
		return err
	} else if reason, ok := result["reason"]; ok {
		return fmt.Errorf("failed to submit block: %s", reason)
	}

	return nil
}

func (node Node) postTransaction(tx string) (string, error) {
	var result map[string]json.RawMessage
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.PostTransaction(tx), &result)
	} else {
		var body []byte
		body, err = json.Marshal(map[string]string{"tx": tx})
		if err != nil {
			return "", err
		}
		err = node.externalHost.ExecHTTP("POST", "/v2/transactions", body, &result)
	}

	if err != nil {
		return "", err
	} else if reason, ok := result["reason"]; ok {
		return "", fmt.Errorf("failed to post tx: %s", reason)
	}

	var txid string
	err = json.Unmarshal(result["tx_hash"], &txid)
	if err != nil {
		return "", err
	}

	return txid, nil
}

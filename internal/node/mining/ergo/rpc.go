package ergo

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/magicpool-co/pool/internal/node/mining/ergo/mock"
)

func (node Node) getAddressFromErgoTree(ergoTree string) (string, error) {
	var address *Address
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetAddressFromErgoTree(), &address)
	} else {
		err = node.httpHost.ExecHTTP("GET", "/utils/ergoTreeToAddress/"+ergoTree, nil, &address)
	}
	if err != nil {
		return "", err
	}

	return address.Address, nil
}

func (node Node) getInfo() (*NodeInfo, error) {
	var info *NodeInfo
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetInfo(), &info)
	} else {
		err = node.httpHost.ExecHTTP("GET", "/info", nil, &info)
	}

	return info, err
}

func (node Node) getWalletBalances() (*Balance, error) {
	var balance *Balance
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetWalletBalances(), &balance)
	} else {
		err = node.httpHost.ExecHTTP("GET", "/wallet/balances", nil, &balance)
	}

	return balance, err
}

func (node Node) getBlocksAtHeight(height uint64) ([]string, error) {
	var headers []string
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetBlockAtHeight(), &headers)
	} else {
		err = node.httpHost.ExecHTTP("GET", "/blocks/at/"+strconv.FormatUint(height, 10), nil, &headers)
	}
	if err != nil {
		return nil, err
	} else if len(headers) == 0 {
		return nil, fmt.Errorf("block at height not found")
	}

	return headers, nil
}

func (node Node) getBlock(header string) (*Block, error) {
	var block *Block
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetBlock(), &block)
	} else {
		err = node.httpHost.ExecHTTP("GET", "/blocks/"+header, nil, &block)
	}

	return block, err
}

func (node Node) getRewardAddress() (string, error) {
	var address *RewardAddress
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetRewardAddress(), &address)
	} else {
		err = node.httpHost.ExecHTTP("GET", "/mining/rewardAddress", nil, &address)
	}
	if err != nil {
		return "", err
	}

	return address.RewardAddress, nil
}

func (node Node) getMiningCandidate() (string, *MiningCandidate, error) {
	var hostID string
	var candidate *MiningCandidate
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetMiningCandidate(), &candidate)
	} else {
		hostID, err = node.httpHost.ExecHTTPSticky("", "GET", "/mining/candidate", nil, &candidate)
	}

	return hostID, candidate, err
}

func (node Node) postMiningSolution(hostID, nonce string) error {
	if node.mocked {
		return nil
	}

	body := map[string]interface{}{"n": nonce}
	var result map[string]interface{}
	_, err := node.httpHost.ExecHTTPSticky(hostID, "POST", "/mining/solution", body, &result)
	if err != nil {
		return err
	} else if len(result) > 0 { // @TODO: this probably should be managed through the error codes instead
		return fmt.Errorf("submit block error: %v", result)
	}

	return nil
}

func (node Node) getWalletStatus() (*WalletStatus, error) {
	var status *WalletStatus
	var err error
	if node.mocked {
		err = json.Unmarshal(mock.GetWalletStatus(), &status)
	} else {
		err = node.httpHost.ExecHTTP("GET", "/wallet/status", nil, &status)
	}

	return status, err
}

func (node Node) postWalletRestore() error {
	if node.mocked {
		return nil
	}

	body := map[string]interface{}{
		"pass":         "rpcrpc",
		"mnemonic":     node.mnemonic,
		"mnemonicPass": "",
	}

	var result string
	err := node.httpHost.ExecHTTP("POST", "/wallet/restore", body, &result)
	if err != nil {
		return err
	} else if result != "OK" {
		return fmt.Errorf("unable to restore wallet")
	}

	return nil
}

func (node Node) postWalletUnlock() error {
	if node.mocked {
		return nil
	}

	body := map[string]interface{}{
		"pass": "rpcrpc",
	}

	var result string
	err := node.httpHost.ExecHTTP("POST", "/wallet/unlock", body, &result)
	if err != nil {
		return err
	} else if result != "OK" {
		return fmt.Errorf("unable to unlock wallet")
	}

	return nil
}

func (node Node) postWalletPaymentSend(address string, amount uint64) (string, error) {
	if node.mocked {
		return "", nil
	}

	body := []map[string]interface{}{
		map[string]interface{}{
			"address": address,
			"amount":  amount,
		},
	}

	var txid string
	err := node.httpHost.ExecHTTP("POST", "/wallet/payment/send", body, &txid)

	return txid, err
}

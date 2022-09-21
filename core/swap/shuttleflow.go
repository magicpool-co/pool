package swap

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/tx/ethtx"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

const (
	zeroAddress = "0x0000000000000000000000000000000000000000"
)

type ShuttleflowClient struct {
	url string
}

func NewShuttleflowClient() *ShuttleflowClient {
	client := &ShuttleflowClient{
		url: "https://shuttleflow.io",
	}

	return client
}

// rpc helpers

func (c *ShuttleflowClient) getSwap(txid, fromChain, toChain string) (*ShuttleflowSwap, error) {
	var direction string
	if fromChain == "cfx" && toChain == "bsc" {
		direction = "out"
	} else if fromChain == "bsc" && toChain == "cfx" {
		direction = "in"
	} else {
		return nil, fmt.Errorf("invalid pair %s-%s", fromChain, toChain)
	}

	req, err := rpc.NewRequest("getUserOperationByHash", txid, direction, "cfx", "bsc")
	if err != nil {
		return nil, err
	}

	res, err := rpc.ExecRPC(c.url+"/rpcshuttleflow", req)
	if err != nil {
		return nil, err
	}

	swap := new(ShuttleflowSwap)
	err = json.Unmarshal(res.Result, &swap)

	return swap, err
}

func (c *ShuttleflowClient) getToken(symbol string) (*ShuttleflowToken, error) {
	req, err := rpc.NewRequest("getTokenList", "bsc")
	if err != nil {
		return nil, err
	}

	res, err := rpc.ExecRPC(c.url+"/rpcsponsor", req)
	if err != nil {
		return nil, err
	}

	tokens := make([]*ShuttleflowToken, 0)
	err = json.Unmarshal(res.Result, &tokens)
	if err != nil {
		return nil, err
	}

	for _, token := range tokens {
		if token.Symbol == symbol || token.ReferenceSymbol == symbol {
			return token, nil
		}
	}

	return nil, fmt.Errorf("unable to find token %s", symbol)
}

func (c *ShuttleflowClient) getDepositAddress(srcAddress, fromChain, toChain, direction string) (string, error) {
	req, err := rpc.NewRequest("getUserWallet", srcAddress, zeroAddress, fromChain, toChain, direction)
	if err != nil {
		return "", err
	}

	res, err := rpc.ExecRPC(c.url+"/rpcshuttleflow", req)
	if err != nil {
		return "", err
	}

	var wallet string
	err = json.Unmarshal(res.Result, &wallet)

	return wallet, err
}

// swaps

func (c *ShuttleflowClient) InitiateSwapFromBSC(bscNode types.PayoutNode, cfxAddress string, amount *big.Int) (string, error) {
	token, err := c.getToken("bCFX")
	if err != nil {
		return "", err
	} else if token.Supported != 1 {
		return "", fmt.Errorf("token bCFX unsupported")
	}

	depositAddress, err := c.getDepositAddress(bscNode.Address(), "cfx", "bsc", "in")
	if err != nil {
		return "", err
	}

	depositAddressBytes, err := common.HexToBytes(depositAddress)
	if err != nil {
		return "", err
	}

	data := ethtx.GenerateContractData("transfer(address,uint256)", depositAddressBytes, amount.Bytes())
	inputs := []*types.TxInput{&types.TxInput{Value: new(big.Int), Data: data}}
	outputs := []*types.TxOutput{&types.TxOutput{Address: token.Reference, Value: new(big.Int)}}
	tx, err := bscNode.CreateTx(inputs, outputs)
	if err != nil {
		return "", err
	}

	return bscNode.BroadcastTx(tx)
}

func (c *ShuttleflowClient) FinalizeSwapFromBSC(cfxNode types.PayoutNode, inTxID string) (string, error) {
	swap, err := c.getSwap(inTxID, "bsc", "cfx")
	if err != nil {
		return "", err
	} else if swap == nil {
		return "", ErrSwapNotReady
	}

	toAddress := swap.TxTo
	rawData := swap.TxInput
	if toAddress == "" || rawData == "" {
		return "", ErrSwapNotReady
	} else if len(rawData) > 2 && rawData[:2] == "0x" {
		rawData = rawData[2:]
	}

	data, err := hex.DecodeString(rawData)
	if err != nil {
		return "", err
	}

	inputs := []*types.TxInput{&types.TxInput{Value: new(big.Int), Data: data}}
	outputs := []*types.TxOutput{&types.TxOutput{Address: toAddress, Value: new(big.Int)}}
	tx, err := cfxNode.CreateTx(inputs, outputs)
	if err != nil {
		return "", err
	}

	return cfxNode.BroadcastTx(tx)
}

func (c *ShuttleflowClient) InitiateSwapFromCFX(cfxNode types.PayoutNode, bscAddress string, amount *big.Int) (string, error) {
	// @TODO: how tf do i find the contract address
	const contractAddress = "cfx:acbsxck3mt6tdcxpbnxunx3wy29fy5k5m6d5cnmpy4"

	token, err := c.getToken("cBNB")
	if err != nil {
		return "", err
	} else if token.Supported != 1 {
		return "", fmt.Errorf("token cBNB unsupported")
	}

	bscAddressBytes, err := common.HexToBytes(bscAddress)
	if err != nil {
		return "", err
	}

	zeroAddressBytes, err := common.HexToBytes(zeroAddress)
	if err != nil {
		return "", err
	}

	data := ethtx.GenerateContractData("deposit(address,address)", bscAddressBytes, zeroAddressBytes)
	inputs := []*types.TxInput{&types.TxInput{Value: amount, Data: data}}
	outputs := []*types.TxOutput{&types.TxOutput{Address: contractAddress, Value: amount}}
	tx, err := cfxNode.CreateTx(inputs, outputs)
	if err != nil {
		return "", err
	}

	return cfxNode.BroadcastTx(tx)
}

func (c *ShuttleflowClient) FinalizeSwapFromCFX(bscNode types.PayoutNode, inTxID string) (string, error) {
	swap, err := c.getSwap(inTxID, "cfx", "bsc")
	if err != nil {
		return "", err
	} else if swap == nil {
		return "", ErrSwapNotReady
	}

	toAddress := swap.TxTo
	rawData := swap.TxInput
	if toAddress == "" || rawData == "" {
		return "", ErrSwapNotReady
	} else if len(rawData) > 2 && rawData[:2] == "0x" {
		rawData = rawData[2:]
	}

	data, err := hex.DecodeString(rawData)
	if err != nil {
		return "", err
	}

	inputs := []*types.TxInput{&types.TxInput{Value: new(big.Int), Data: data}}
	outputs := []*types.TxOutput{&types.TxOutput{Address: toAddress, Value: new(big.Int)}}
	tx, err := bscNode.CreateTx(inputs, outputs)
	if err != nil {
		return "", err
	}

	return bscNode.BroadcastTx(tx)
}

package eth

import (
	"context"
	"encoding/hex"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

func generateHost(url string, logger *log.Logger) (*hostpool.HTTPPool, error) {
	var (
		hostHealthCheck = &hostpool.HTTPHealthCheck{
			RPCRequest: &rpc.Request{
				ID:      []byte(`1`),
				JSONRPC: "2.0",
				Method:  "eth_syncing",
			},
		}
	)

	if url == "" {
		return nil, nil
	}

	host := hostpool.NewHTTPPool(context.Background(), logger, hostHealthCheck, nil)
	err := host.AddHost(url, 0, nil)

	return host, err
}

func New(mainnet bool, url, rawPriv string, erc20 *ERC20, logger *log.Logger) (*Node, error) {
	host, err := generateHost(url, logger)
	if err != nil {
		return nil, err
	}

	privBytes, err := crypto.ObscureHex(rawPriv)
	if err != nil {
		return nil, err
	}

	privKey := secp256k1.PrivKeyFromBytes(privBytes)
	pubKeyBytes := privKey.PubKey().SerializeUncompressed()
	address := "0x" + hex.EncodeToString(crypto.Keccak256(pubKeyBytes[1:])[12:])

	node := &Node{
		mocked:  host == nil,
		mainnet: mainnet,
		address: address,
		privKey: privKey,
		rpcHost: host,
		erc20:   erc20,
		logger:  logger,
	}

	return node, nil
}

type Node struct {
	mocked  bool
	mainnet bool
	address string
	privKey *secp256k1.PrivateKey
	rpcHost *hostpool.HTTPPool
	erc20   *ERC20
	logger  *log.Logger
}

type ERC20 struct {
	Chain    string
	Address  string
	Decimals int
	Units    *types.Number
}

type Block struct {
	Number   string `json:"number"`
	Hash     string `json:"hash"`
	GasLimit string `json:"gasLimit"`
	GasUsed  string `json:"gasUsed"`
	BaseFee  string `json:"baseFeePerGas"`
}

type Transaction struct {
	Hash                 string `json:"hash"`
	BlockHash            string `json:"blockHash"`
	BlockNumber          string `json:"blockNumber"`
	From                 string `json:"from"`
	To                   string `json:"to"`
	Value                string `json:"value"`
	Gas                  string `json:"gas"`
	GasPrice             string `json:"gasPrice"`
	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
	Input                string `json:"input"`
	Nonce                string `json:"nonce"`
	Index                string `json:"transactionIndex"`
	Type                 string `json:"type"`
	V                    string `json:"v"`
	R                    string `json:"r"`
	S                    string `json:"s"`
}

type TransactionReceipt struct {
	TxHash            string `json:"transactionHash"`
	GasUsed           string `json:"gasUsed"`
	EffectiveGasPrice string `json:"effectiveGasPrice"`
	BlockHash         string `json:"blockHash"`
	Status            string `json:"status"`
}

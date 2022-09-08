package eth

import (
	"context"
	"encoding/hex"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

func generateHost(urls []string, tunnel *sshtunnel.SSHTunnel) (*hostpool.HTTPPool, error) {
	var (
		port            = 8545
		hostHealthCheck = &hostpool.HTTPHealthCheck{
			RPCRequest: &rpc.Request{
				JSONRPC: "2.0",
				Method:  "eth_syncing",
			},
		}
	)

	if len(urls) == 0 {
		return nil, nil
	}

	host := hostpool.NewHTTPPool(context.Background(), hostHealthCheck, tunnel)
	for _, url := range urls {
		err := host.AddHost(url, port, nil)
		if err != nil {
			return nil, err
		}
	}

	return host, nil
}

func New(mainnet bool, urls []string, rawPriv string, tunnel *sshtunnel.SSHTunnel, erc20 *ERC20) (*Node, error) {
	host, err := generateHost(urls, tunnel)
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

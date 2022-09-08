package ctxc

import (
	"context"
	"encoding/hex"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sencha-dev/powkit/cuckoo"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func generateHost(urls []string, tunnel *sshtunnel.SSHTunnel) (*hostpool.HTTPPool, error) {
	var (
		httpPort        = 8546
		httpHealthCheck = &hostpool.HTTPHealthCheck{
			RPCRequest: &rpc.Request{
				JSONRPC: "2.0",
				Method:  "eth_syncing",
			},
		}
	)

	if len(urls) == 0 {
		return nil, nil
	}

	host := hostpool.NewHTTPPool(context.Background(), httpHealthCheck, tunnel)
	for _, url := range urls {
		err := host.AddHost(url, httpPort, nil)
		if err != nil {
			return nil, err
		}
	}

	return host, nil
}

func New(mainnet bool, urls []string, rawPriv string, tunnel *sshtunnel.SSHTunnel) (*Node, error) {
	host, err := generateHost(urls, tunnel)
	if err != nil {
		return nil, err
	}

	obscuredPriv, err := crypto.ObscureHex(rawPriv)
	if err != nil {
		return nil, err
	}

	privKey := secp256k1.PrivKeyFromBytes(obscuredPriv)
	pubKeyBytes := privKey.PubKey().SerializeUncompressed()
	address := "0x" + hex.EncodeToString(crypto.Keccak256(pubKeyBytes[1:])[12:])

	node := &Node{
		mocked:  host == nil,
		mainnet: mainnet,
		address: address,
		privKey: privKey,
		rpcHost: host,
		pow:     cuckoo.NewCortex(),
	}

	return node, nil
}

type Node struct {
	mocked  bool
	mainnet bool
	address string
	privKey *secp256k1.PrivateKey
	rpcHost *hostpool.HTTPPool
	pow     *cuckoo.Client
}

type Block struct {
	Number           string         `json:"number"`
	Hash             string         `json:"hash"`
	ParentHash       string         `json:"parentHash"`
	Nonce            string         `json:"nonce"`
	Solution         []uint32       `json:"solution"`
	SHA3Uncles       string         `json:"sha3Uncles"`
	Miner            string         `json:"miner"`
	Difficulty       string         `json:"difficulty"`
	TotalDifficulty  string         `json:"totalDifficulty"`
	ExtraData        string         `json:"extraData"`
	Size             string         `json:"size"`
	StateRoot        string         `json:"stateRoot"`
	Supply           string         `json:"supply"`
	GasLimit         string         `json:"gasLimit"`
	GasUsed          string         `json:"gasUsed"`
	Quota            string         `json:"quota"`
	QuotaUsed        string         `json:"quotaUsed"`
	Timestamp        string         `json:"timestamp"`
	Transactions     []*Transaction `json:"transactions"`
	TransactionsRoot string         `json:"transactionsRoot"`
	ReceiptsRoot     string         `json:"receiptsRoot"`
	Uncles           []string       `json:"uncles"`
}

type Transaction struct {
	Hash        string `json:"hash"`
	BlockHash   string `json:"blockHash"`
	BlockNumber string `json:"blockNumber"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	Gas         string `json:"gas"`
	GasPrice    string `json:"gasPrice"`
	Input       string `json:"input"`
	Nonce       string `json:"nonce"`
	Index       string `json:"transactionIndex"`
	V           string `json:"v"`
	R           string `json:"r"`
	S           string `json:"s"`
}

type TransactionReceipt struct {
	TxHash      string `json:"transactionHash"`
	GasUsed     string `json:"gasUsed"`
	BlockHash   string `json:"blockHash"`
	BlockNumber string `json:"blockNumber"`
	Status      string `json:"status"`
}

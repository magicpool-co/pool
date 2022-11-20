package cfx

import (
	"context"
	"fmt"
	"time"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/goccy/go-json"
	"github.com/sencha-dev/powkit/octopus"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/bech32"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

const (
	mainnetChainID uint64 = 1029
	mainnetPrefix         = "cfx"
	mainnetURL            = "http://main.confluxrpc.com"
	testnetChainID uint64 = 1
	testnetPrefix         = "cfxtest"
	testnetURL            = "http://test.confluxrpc.com"
)

func generateHost(urls []string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*hostpool.HTTPPool, *hostpool.TCPPool, error) {
	var (
		rpcPort            = 12537
		rpcHostHealthCheck = &hostpool.HTTPHealthCheck{
			RPCRequest: &rpc.Request{
				JSONRPC: "2.0",
				Method:  "cfx_getBestBlockHash",
			},
		}
		tcpPort            = 32525
		tcpHostHealthCheck = &hostpool.TCPHealthCheck{
			Interval: time.Second * 15,
			RPCRequest: &rpc.Request{
				JSONRPC: "2.0",
				Method:  "mining.subscribe",
				Params: []json.RawMessage{
					common.MustMarshalJSON("x"),
					common.MustMarshalJSON("x"),
				},
			},
		}
	)

	if len(urls) == 0 {
		return nil, nil, nil
	}

	rpcHost := hostpool.NewHTTPPool(context.Background(), logger, rpcHostHealthCheck, tunnel)
	tcpHost := hostpool.NewTCPPool(context.Background(), logger, tcpHostHealthCheck, tunnel)
	for _, url := range urls {
		err := rpcHost.AddHost(url, rpcPort, nil)
		if err != nil {
			return nil, nil, err
		}

		err = tcpHost.AddHost(url, tcpPort)
		if err != nil {
			return nil, nil, err
		}
	}

	return rpcHost, tcpHost, nil
}

func New(mainnet bool, urls []string, rawPriv string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*Node, error) {
	networkID := mainnetChainID
	networkPrefix := mainnetPrefix
	fallbackURL := mainnetURL
	if !mainnet {
		networkID = testnetChainID
		networkPrefix = testnetPrefix
		fallbackURL = testnetURL
	}

	rpcHost, tcpHost, err := generateHost(urls, logger, tunnel)
	if err != nil {
		return nil, err
	}

	obscuredPriv, err := crypto.ObscureHex(rawPriv)
	if err != nil {
		return nil, err
	}

	privKey := secp256k1.PrivKeyFromBytes(obscuredPriv)
	pubKeyBytes := privKey.PubKey().SerializeUncompressed()
	ethAddress := crypto.Keccak256(pubKeyBytes[1:])[12:]
	if ethAddress[0] != 0x10 {
		return nil, fmt.Errorf("invalid eth address %0x: no 0x10 prefix", ethAddress)
	}

	cfxAddress, err := bech32.EncodeBCH(addressCharset, networkPrefix, addressVersion, ethAddress)
	if err != nil {
		return nil, err
	}

	node := &Node{
		mocked:        rpcHost == nil && tcpHost == nil,
		mainnet:       mainnet,
		networkID:     networkID,
		networkPrefix: networkPrefix,
		fallbackURL:   fallbackURL,
		address:       cfxAddress,
		privKey:       privKey,
		rpcHost:       rpcHost,
		tcpHost:       tcpHost,
		pow:           octopus.NewConflux(),
		logger:        logger,
	}

	return node, nil
}

type Node struct {
	mocked        bool
	mainnet       bool
	networkID     uint64
	networkPrefix string
	fallbackURL   string
	address       string
	privKey       *secp256k1.PrivateKey
	rpcHost       *hostpool.HTTPPool
	tcpHost       *hostpool.TCPPool
	pow           *octopus.Client
	logger        *log.Logger
}

func (node Node) execRPCfromFallback(req *rpc.Request, target interface{}) error {
	res, err := rpc.ExecRPC(node.fallbackURL, req)
	if err != nil {
		return err
	} else if err := json.Unmarshal(res.Result, target); err != nil {
		return err
	}

	return nil
}

func (node Node) execRPCfromFallbackBulk(requests []*rpc.Request) ([]*rpc.Response, error) {
	return rpc.ExecRPCBulk(node.fallbackURL, requests)
}

type Block struct {
	Adaptive              bool     `json:"adaptive"`
	Blame                 string   `json:"blame"`
	BlockNumber           string   `json:"blockNumber"`
	Custom                []string `json:"custom"`
	DeferredLogsBloomHash string   `json:"deferredLogsBloomHash"`
	DeferredReceiptsRoot  string   `json:"deferredReceiptsRoot"`
	DeferredStateRoot     string   `json:"deferredStateRoot"`
	Difficulty            string   `json:"difficulty"`
	EpochNumber           string   `json:"epochNumber"`
	GasLimit              string   `json:"gasLimit"`
	GasUsed               string   `json:"gasUsed"`
	Hash                  string   `json:"hash"`
	Height                string   `json:"height"`
	Miner                 string   `json:"miner"`
	Nonce                 string   `json:"nonce"`
	ParentHash            string   `json:"parentHash"`
	PosReference          string   `json:"posReference"`
	PowQuality            string   `json:"powQuality"`
	RefereeHashes         []string `json:"refereeHashes"`
	Size                  string   `json:"size"`
	Timestamp             string   `json:"timestamp"`
	Transactions          []string `json:"transactions"`
	TransactionsRoot      string   `json:"transactionsRoot"`
}

type BlockRewardInfo struct {
	Author      string `json:"author"`
	BaseReward  string `json:"baseReward"`
	BlockHash   string `json:"blockHash"`
	TotalReward string `json:"totalReward"`
	TxFee       string `json:"txFee"`
}

type Transaction struct {
	BlockHash string `json:"blockHash"`
	ChainId   string `json:"chainId"`
	// ContractCreated string `json:"contractCreated"`
	Data             string `json:"data"`
	EpochHeight      string `json:"epochHeight"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Nonce            string `json:"nonce"`
	R                string `json:"r"`
	S                string `json:"s"`
	Status           string `json:"status"`
	StorageLimit     string `json:"storageLimit"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	V                string `json:"v"`
	Value            string `json:"value"`
}

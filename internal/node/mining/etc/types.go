package etc

import (
	"context"
	"encoding/hex"
	"fmt"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sencha-dev/powkit/ethash"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

type EthType int

const (
	ETC EthType = iota
	ETHW
)

func generateHost(ethType EthType, urls []string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*hostpool.HTTPPool, error) {
	var port int
	switch ethType {
	case ETC:
		port = 8544
	case ETHW:
		port = 8545
	}

	var (
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

	host := hostpool.NewHTTPPool(context.Background(), logger, hostHealthCheck, tunnel)
	for _, url := range urls {
		err := host.AddHost(url, port, nil)
		if err != nil {
			return nil, err
		}
	}

	return host, nil
}

func New(ethType EthType, mainnet bool, urls []string, rawPriv string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*Node, error) {
	switch ethType {
	case ETC, ETHW:
	default:
		return nil, fmt.Errorf("unknown eth type")
	}

	host, err := generateHost(ethType, urls, logger, tunnel)
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
		ethType: ethType,
		mocked:  host == nil,
		mainnet: mainnet,
		address: address,
		privKey: privKey,
		rpcHost: host,
		pow:     ethash.NewEthereumClassic(),
		logger:  logger,
	}

	return node, nil
}

type Node struct {
	ethType EthType
	mocked  bool
	mainnet bool
	address string
	privKey *secp256k1.PrivateKey
	rpcHost *hostpool.HTTPPool
	pow     *ethash.Client
	logger  *log.Logger
}

type Block struct {
	Number           string         `json:"number"`
	Hash             string         `json:"hash"`
	ParentHash       string         `json:"parentHash"`
	Nonce            string         `json:"nonce"`
	MixHash          string         `json:"mixHash"`
	SHA3Uncles       string         `json:"sha3Uncles"`
	Miner            string         `json:"miner"`
	Difficulty       string         `json:"difficulty"`
	TotalDifficulty  string         `json:"totalDifficulty"`
	ExtraData        string         `json:"extraData"`
	Size             string         `json:"size"`
	StateRoot        string         `json:"stateRoot"`
	GasLimit         string         `json:"gasLimit"`
	GasUsed          string         `json:"gasUsed"`
	Timestamp        string         `json:"timestamp"`
	Transactions     []*Transaction `json:"transactions"`
	TransactionsRoot string         `json:"transactionsRoot"`
	ReceiptsRoot     string         `json:"receiptsRoot"`
	Uncles           []string       `json:"uncles"`
	BaseFee          string         `json:"baseFeePerGas"`
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
	Type        string `json:"type"`
	V           string `json:"v"`
	R           string `json:"r"`
	S           string `json:"s"`
}

type TransactionReceipt struct {
	TxHash            string `json:"transactionHash"`
	GasUsed           string `json:"gasUsed"`
	EffectiveGasPrice string `json:"effectiveGasPrice"`
	BlockHash         string `json:"blockHash"`
	BlockNumber       string `json:"blockNumber"`
	Status            string `json:"status"`
}

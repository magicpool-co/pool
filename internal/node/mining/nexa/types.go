package nexa

import (
	"context"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/bech32"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

// general constants
var (
	mainnetPrefix = "nexa"
	testnetPrefix = "nexatest"
)

func generateHost(urls []string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*hostpool.HTTPPool, error) {
	var (
		port        = 7227
		hostOptions = &hostpool.HTTPHostOptions{
			Username: "rpc",
			Password: "rpc",
		}
		hostHealthCheck = &hostpool.HTTPHealthCheck{
			RPCRequest: &rpc.Request{
				JSONRPC: "2.0",
				Method:  "getbestblockhash",
			},
		}
	)

	if len(urls) == 0 {
		return nil, nil
	}

	host := hostpool.NewHTTPPool(context.Background(), logger, hostHealthCheck, tunnel)
	for _, url := range urls {
		err := host.AddHost(url, port, hostOptions)
		if err != nil {
			return nil, err
		}
	}

	return host, nil
}

func New(mainnet bool, urls []string, rawPriv string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*Node, error) {
	prefix := mainnetPrefix
	if !mainnet {
		prefix = testnetPrefix
	}

	host, err := generateHost(urls, logger, tunnel)
	if err != nil {
		return nil, err
	}

	obscuredPriv, err := crypto.ObscureHex(rawPriv)
	if err != nil {
		return nil, err
	} else if err := crypto.ValidateSecp256k1PrivateKey(obscuredPriv); err != nil {
		return nil, err
	}

	privKey := secp256k1.PrivKeyFromBytes(obscuredPriv)
	pubKeyBytes := privKey.PubKey().SerializeUncompressed()
	pubKeyHash := crypto.Ripemd160(crypto.Sha256(pubKeyBytes))
	address, err := bech32.EncodeBCH(addressCharset, prefix, pubKeyAddrID, pubKeyHash)
	if err != nil {
		return nil, err
	}

	node := &Node{
		mocked:  host == nil,
		mainnet: mainnet,
		prefix:  prefix,
		address: address,
		privKey: privKey,
		rpcHost: host,
		logger:  logger,
	}

	return node, nil
}

type Node struct {
	mocked  bool
	mainnet bool
	prefix  string
	address string
	privKey *secp256k1.PrivateKey
	rpcHost *hostpool.HTTPPool
	logger  *log.Logger
}

type BlockchainInfo struct {
	Chain                string  `json:"chain"`
	Blocks               uint64  `json:"blocks"`
	Headers              uint64  `json:"headers"`
	BestBlockHash        string  `json:"bestblockhash"`
	Difficulty           float64 `json:"difficulty"`
	MedianTime           int64   `json:"mediantime"`
	VerificationProgress float64 `json:"verificationprogress"`
	ChainWork            string  `json:"chainwork"`
	Pruned               bool    `json:"pruned"`
}

type Transaction struct {
	Data          string `json:"data"`
	TxID          string `json:"txid"`
	Hash          string `json:"hash"`
	SigOps        int    `json:"sigops"`
	Weight        int    `json:"weight"`
	Height        int64  `json:"height"`
	Confirmations int64  `json:"confirmations"`
	Inputs        []struct {
		Coinbase string `json:"coinbase"`
	} `json:"vin"`
	Outputs []struct {
		Value json.Number `json:"value"`
	} `json:"vout"`
}

type MiningCandidate struct {
	ID               uint64 `json:"id"`
	HeaderCommitment string `json:"headerCommitment"`
	NBits            string `json:"nBits"`
}

type SubmissionResponse struct {
	Height uint64  `json:"height"`
	Hash   string  `json:"hash"`
	Result *string `json:"result"`
}

type Block struct {
	Hash              string         `json:"hash"`
	Confirmations     int64          `json:"confirmations"`
	StrippedSize      uint64         `json:"strippedsize"`
	Size              uint64         `json:"size"`
	Weight            uint64         `json:"weight"`
	Height            uint64         `json:"height"`
	Version           uint64         `json:"version"`
	VersionHex        string         `json:"versionHex"`
	MerkleRoot        string         `json:"merkleroot"`
	Transactions      []*Transaction `json:"tx"`
	Time              int64          `json:"time"`
	MedianTime        int64          `json:"mediantime"`
	Bits              string         `json:"bits"`
	Difficulty        float64        `json:"difficulty"`
	Chainwork         string         `json:"chainwork"`
	HeaderHash        string         `json:"headerhash"`
	MixHash           string         `json:"mixhash"`
	Nonce             uint64         `json:"nonce64"`
	PreviousBlockHash string         `json:"previousblockhash"`
	NextBlockHash     string         `json:"nextblockhash"`
}

package firo

import (
	"context"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/goccy/go-json"
	"github.com/sencha-dev/powkit/firopow"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/base58"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

var (
	mainnetPrefixP2PKH        = []byte{0x52}
	mainnetPrefixP2SH         = []byte{0x07}
	mainnetDevWalletAddresses = []string{"aFA2TbqG9cnhhzX5Yny2pBJRK5EaEqLCH7", "aLgRaYSFk6iVw2FqY1oei8Tdn2aTsGPVmP"}
	mainnetDevWalletAmounts   = []uint64{62500000, 93750000}

	testnetPrefixP2PKH        = []byte{0x41}
	testnetPrefixP2SH         = []byte{0xb2}
	testnetDevWalletAddresses = []string{"TCkC4uoErEyCB4MK3d6ouyJELoXnuyqe9L", "TWDxLLKsFp6qcV1LL4U2uNmW4HwMcapmMU"}
	testnetDevWalletAmounts   = []uint64{62500000, 93750000}
)

func generateHost(urls []string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*hostpool.HTTPPool, error) {
	var (
		port        = 8888
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
	devWalletAddresses := mainnetDevWalletAddresses
	devWalletAmounts := mainnetDevWalletAmounts
	prefixP2PKH := mainnetPrefixP2PKH
	prefixP2SH := mainnetPrefixP2SH
	if !mainnet {
		devWalletAddresses = testnetDevWalletAddresses
		devWalletAmounts = testnetDevWalletAmounts
		prefixP2PKH = testnetPrefixP2PKH
		prefixP2SH = testnetPrefixP2SH
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
	address := base58.CheckEncode(prefixP2PKH, pubKeyHash)

	node := &Node{
		mocked:             host == nil,
		mainnet:            mainnet,
		devWalletAddresses: devWalletAddresses,
		devWalletAmounts:   devWalletAmounts,
		prefixP2PKH:        prefixP2PKH,
		prefixP2SH:         prefixP2SH,
		address:            address,
		privKey:            privKey,
		rpcHost:            host,
		pow:                firopow.NewFiro(),
		logger:             logger,
	}

	return node, nil
}

type Node struct {
	mocked             bool
	mainnet            bool
	devWalletAddresses []string
	devWalletAmounts   []uint64
	prefixP2PKH        []byte
	prefixP2SH         []byte
	address            string
	privKey            *secp256k1.PrivateKey
	rpcHost            *hostpool.HTTPPool
	pow                *firopow.Client
	logger             *log.Logger
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
	Height        int64  `json:"height"`
	Confirmations int64  `json:"confirmations"`
	Inputs        []struct {
		Coinbase string `json:"coinbase"`
	} `json:"vin"`
	Outputs []struct {
		Value        json.Number `json:"value"`
		ScriptPubKey struct {
			Addresses []string `json:"addresses"`
		} `json:"scriptPubKey"`
	} `json:"vout"`
}

type BlockZNode struct {
	Payee  string `json:"payee"`
	Script string `json:"script"`
	Amount uint64 `json:"amount"`
}

type BlockTemplate struct {
	Capabilities []string `json:"capabilities"`
	Version      uint32   `json:"version"`
	Rules        []string `json:"rules"`
	// VBAvailable interface{} `json:"vbavailable"`
	VBRequired        int            `json:"vbrequired"`
	PreviousBlockHash string         `json:"previousblockhash"`
	Transactions      []*Transaction `json:"transactions"`
	CoinbaseAux       struct {
		Flags string `json:"flags"`
	} `json:"coinbaseaux"`
	CoinbaseValue   uint64        `json:"coinbasevalue"`
	LongPollID      string        `json:"longpollid"`
	Target          string        `json:"target"`
	MinTime         int           `json:"mintime"`
	Mutable         []string      `json:"mutable"`
	NonceRange      string        `json:"noncerange"`
	SigOpLimit      int           `json:"sigoplimit"`
	WeightLimit     int           `json:"weightlimit"`
	CurTime         uint32        `json:"curtime"`
	Bits            string        `json:"bits"`
	Height          uint64        `json:"height"`
	ZNode           []*BlockZNode `json:"znode"`
	CoinbasePayload string        `json:"coinbase_payload"`
}

type Block struct {
	Hash              string   `json:"hash"`
	Confirmations     int64    `json:"confirmations"`
	StrippedSize      uint64   `json:"strippedsize"`
	Size              uint64   `json:"size"`
	Weight            uint64   `json:"weight"`
	Height            uint64   `json:"height"`
	Version           uint64   `json:"version"`
	VersionHex        string   `json:"versionHex"`
	MerkleRoot        string   `json:"merkleroot"`
	Transactions      []string `json:"tx"`
	Time              int64    `json:"time"`
	MedianTime        int64    `json:"mediantime"`
	Bits              string   `json:"bits"`
	Difficulty        float64  `json:"difficulty"`
	Chainwork         string   `json:"chainwork"`
	HeaderHash        string   `json:"headerhash"`
	MixHash           string   `json:"mixhash"`
	Nonce             uint64   `json:"nonce64"`
	PreviousBlockHash string   `json:"previousblockhash"`
	NextBlockHash     string   `json:"nextblockhash"`
}

package flux

import (
	"context"
	"fmt"
	"net/http"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/goccy/go-json"
	"github.com/sencha-dev/powkit/equihash"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/base58"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

var (
	mainnetPrefixP2PKH      = []byte{0x1C, 0xB8}
	mainnetPrefixP2SH       = []byte{0x1C, 0xBD}
	mainnetDevWalletAmounts = []uint64{281250000, 468750000, 1125000000}

	testnetPrefixP2PKH      = []byte{0x1D, 0x25}
	testnetPrefixP2SH       = []byte{0x1C, 0xBA}
	testnetDevWalletAmounts = []uint64{562500000, 937500000, 2250000000}
)

func privKeyToWIFUncompressed(privKey *secp256k1.PrivateKey) string {
	wif := append([]byte{0x80}, privKey.Serialize()...)
	checksum := crypto.Sha256d(wif)[:4]
	wif = append(wif, checksum...)

	return base58.Encode(wif)
}

func generateHost(
	urls []string,
	logger *log.Logger,
	tunnel *sshtunnel.SSHTunnel,
) (*hostpool.HTTPPool, error) {
	var (
		port        = 16124
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

func New(
	mainnet bool,
	urls []string,
	rawPriv string,
	logger *log.Logger,
	tunnel *sshtunnel.SSHTunnel,
) (*Node, error) {
	prefixP2PKH := mainnetPrefixP2PKH
	prefixP2SH := mainnetPrefixP2SH
	devWalletAmounts := mainnetDevWalletAmounts
	if !mainnet {
		prefixP2PKH = testnetPrefixP2PKH
		prefixP2SH = testnetPrefixP2SH
		devWalletAmounts = testnetDevWalletAmounts
	}

	host, err := generateHost(urls, logger, tunnel)
	if err != nil {
		return nil, err
	}

	obscuredPriv, err := crypto.ObscureHex(rawPriv)
	if err != nil {
		return nil, err
	}

	privKey := secp256k1.PrivKeyFromBytes(obscuredPriv)
	address := btctx.PrivKeyToAddress(privKey, prefixP2PKH)
	wif := privKeyToWIFUncompressed(privKey)

	node := &Node{
		mocked:           host == nil,
		mainnet:          mainnet,
		prefixP2PKH:      prefixP2PKH,
		prefixP2SH:       prefixP2SH,
		devWalletAmounts: devWalletAmounts,
		address:          address,
		wif:              wif,
		privKey:          privKey,
		rpcHost:          host,
		pow:              equihash.NewFlux(),
		logger:           logger,
	}

	return node, nil
}

type Node struct {
	mocked           bool
	mainnet          bool
	prefixP2PKH      []byte
	prefixP2SH       []byte
	devWalletAmounts []uint64
	address          string
	wif              string
	privKey          *secp256k1.PrivateKey
	rpcHost          *hostpool.HTTPPool
	pow              *equihash.Client
	logger           *log.Logger
}

func (node *Node) HandleHostPoolInfoRequest(w http.ResponseWriter, r *http.Request) {
	if node.rpcHost == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"status": 400, "error": "NoHostPool"}`))
		return
	}

	node.rpcHost.HandleInfoRequest(w, r)
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

type BlockTemplate struct {
	Capabilities []string `json:"capabilities"`
	Version      uint32   `json:"version"`
	Rules        []string `json:"rules"`
	// VBAvailable interface{} `json:"vbavailable"`
	VBRequired           int            `json:"vbrequired"`
	PreviousBlockHash    string         `json:"previousblockhash"`
	FinalSaplingRootHash string         `json:"finalsaplingroothash"`
	Transactions         []*Transaction `json:"transactions"`
	CoinbaseAux          struct {
		Flags string `json:"flags"`
	} `json:"coinbaseaux"`
	CoinbaseValue          uint64   `json:"coinbasevalue"`
	LongPollID             string   `json:"longpollid"`
	Target                 string   `json:"target"`
	MinTime                int      `json:"mintime"`
	Mutable                []string `json:"mutable"`
	NonceRange             string   `json:"noncerange"`
	SigOpLimit             int      `json:"sigoplimit"`
	WeightLimit            int      `json:"weightlimit"`
	CurTime                uint32   `json:"curtime"`
	Bits                   string   `json:"bits"`
	Height                 uint64   `json:"height"`
	MinerReward            uint64   `json:"miner_reward"`
	CumulusFluxnodeAddress string   `json:"cumulus_fluxnode_address"`
	CumulusFluxnodePayout  uint64   `json:"cumulus_fluxnode_payout"`
	NimbusFluxnodeAddress  string   `json:"nimbus_fluxnode_address"`
	NimbusFluxnodePayout   uint64   `json:"nimbus_fluxnode_payout"`
	StratusFluxnodeAddress string   `json:"stratus_fluxnode_address"`
	StratusFluxnodePayout  uint64   `json:"stratus_fluxnode_payout"`
}

type SignedTransactionError struct {
	TxID      string `json:"txid"`
	VOut      uint32 `json:"vout"`
	ScriptSig string `json:"scriptSig"`
	Sequence  uint32 `json:"sequence"`
	ErrorMsg  string `json:"error"`
}

func (s *SignedTransactionError) Error() string {
	return fmt.Sprintf("failed tx signing: %s: %d: %s", s.TxID, s.VOut, s.ErrorMsg)
}

type SignedTransaction struct {
	Hex      string `json:"hex"`
	Complete bool   `json:"complete"`
	Errors   []*SignedTransactionError
}

type Block struct {
	Hash              string         `json:"hash"`
	Confirmations     int64          `json:"confirmations"`
	StrippedSize      uint64         `json:"strippedsize"`
	Size              uint64         `json:"size"`
	Weight            uint64         `json:"weight"`
	Height            uint64         `json:"height"`
	Version           uint64         `json:"version"`
	MerkleRoot        string         `json:"merkleroot"`
	FinalSaplingRoot  string         `json:"finalsaplingroot"`
	Transactions      []*Transaction `json:"tx"`
	Time              int64          `json:"time"`
	MedianTime        int64          `json:"mediantime"`
	Bits              string         `json:"bits"`
	Difficulty        float64        `json:"difficulty"`
	Chainwork         string         `json:"chainwork"`
	Nonce             string         `json:"nonce"`
	Solution          string         `json:"solution"`
	Anchor            string         `json:"anchor"`
	PreviousBlockHash string         `json:"previousblockhash"`
	NextBlockHash     string         `json:"nextblockhash"`
}

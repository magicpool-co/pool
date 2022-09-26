package flux

import (
	"context"
	"fmt"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/goccy/go-json"
	"github.com/sencha-dev/powkit/equihash"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/base58"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

var (
	mainnetPrefixP2PKH = []byte{0x1C, 0xB8}
	mainnetPrefixP2SH  = []byte{0x1C, 0xBD}

	testnetPrefixP2PKH = []byte{0x1D, 0x25}
	testnetPrefixP2SH  = []byte{0x1C, 0xBA}
)

func generateHost(urls []string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*hostpool.HTTPPool, error) {
	var (
		port        = 16124
		hostOptions = &hostpool.HTTPHostOptions{
			Username: "rpc",
			Password: "rpc",
		}
		hostHealthCheck = &hostpool.HTTPHealthCheck{
			RPCRequest: &rpc.Request{
				JSONRPC: "2.0",
				Method:  "getinfo",
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
	prefixP2PKH := mainnetPrefixP2PKH
	prefixP2SH := mainnetPrefixP2SH
	if !mainnet {
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
		mocked:      host == nil,
		mainnet:     mainnet,
		prefixP2PKH: prefixP2PKH,
		prefixP2SH:  prefixP2SH,
		address:     address,
		wif:         crypto.PrivKeyToWIFUncompressed(privKey),
		privKey:     privKey,
		rpcHost:     host,
		pow:         equihash.NewFlux(),
		logger:      logger,
	}

	return node, nil
}

type Node struct {
	mocked      bool
	mainnet     bool
	prefixP2PKH []byte
	prefixP2SH  []byte
	address     string
	wif         string
	privKey     *secp256k1.PrivateKey
	rpcHost     *hostpool.HTTPPool
	pow         *equihash.Client
	logger      *log.Logger
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
	Data   string `json:"data"`
	TxID   string `json:"txid"`
	Hash   string `json:"hash"`
	Inputs []struct {
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
	Confirmations     uint64         `json:"confirmations"`
	StrippedSize      uint64         `json:"strippedsize"`
	Size              uint64         `json:"size"`
	Weight            uint64         `json:"weight"`
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

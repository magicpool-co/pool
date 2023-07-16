package ae

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/sencha-dev/powkit/cuckoo"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/base58"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
)

const (
	mainnetNetworkID   = "ae_mainnet"
	mainnetFallbackURL = "https://mainnet.aeternity.io"

	testnetNetworkID   = "ae_uat"
	testnetFallbackURL = "https://testnet.aeternity.io"
)

func b58Encode(prefix string, data []byte) string {
	return prefix + base58.Encode(append(data, crypto.Sha256d(data)[0:4]...))
}

func generateHost(internal bool, urls []string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*hostpool.HTTPPool, error) {
	var (
		externalPort            = 3013
		internalPort            = 3113
		externalHostHealthCheck = &hostpool.HTTPHealthCheck{
			HTTPMethod: "GET",
			HTTPPath:   "/v2/status",
		}
		internalHostHealthCheck = &hostpool.HTTPHealthCheck{
			HTTPMethod: "GET",
			HTTPPath:   "/v2/debug/network",
		}
	)

	if len(urls) == 0 {
		return nil, nil
	}

	port := externalPort
	healthcheck := externalHostHealthCheck
	if internal {
		port = internalPort
		healthcheck = internalHostHealthCheck
	}

	host := hostpool.NewHTTPPool(context.Background(), logger, healthcheck, tunnel)
	for _, url := range urls {
		err := host.AddHost(url, port, nil)
		if err != nil {
			return nil, err
		}
	}

	return host, nil
}

func New(mainnet bool, urls []string, rawPriv string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*Node, error) {
	networkID := mainnetNetworkID
	fallbackURL := mainnetFallbackURL
	if !mainnet {
		networkID = testnetNetworkID
		fallbackURL = testnetFallbackURL
	}

	externalHost, err := generateHost(false, urls, logger, tunnel)
	if err != nil {
		return nil, err
	}

	internalHost, err := generateHost(true, urls, logger, tunnel)
	if err != nil {
		return nil, err
	}

	obscuredPriv, err := hex.DecodeString(rawPriv)
	// obscuredPriv, err := crypto.ObscureHex(rawPriv)
	if err != nil {
		return nil, err
	}

	privKey, err := crypto.PrivKeyFromBytesED25519(obscuredPriv)
	if err != nil {
		return nil, err
	}

	address := b58Encode(addressPrefix, []byte(fmt.Sprintf("%s", privKey.Public())))

	node := &Node{
		mocked:       externalHost == nil && internalHost == nil,
		mainnet:      mainnet,
		networkID:    networkID,
		fallbackURL:  fallbackURL,
		address:      address,
		privKey:      privKey,
		internalHost: internalHost,
		externalHost: externalHost,
		pow:          cuckoo.NewAeternity(),
		logger:       logger,
	}

	return node, nil
}

type Node struct {
	mocked       bool
	mainnet      bool
	networkID    string
	fallbackURL  string
	address      string
	privKey      ed25519.PrivateKey
	internalHost *hostpool.HTTPPool
	externalHost *hostpool.HTTPPool
	pow          *cuckoo.Client
	logger       *log.Logger
}

func (node *Node) HandleHostPoolInfoRequest(w http.ResponseWriter, r *http.Request) {
	if node.internalHost == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"status": 400, "error": "NoHostPool"}`))
		return
	}

	node.internalHost.HandleInfoRequest(w, r)
}

type Status struct {
	GenesisKeyBlockHash      string  `json:"genesis_key_block_hash"`
	Solutions                uint64  `json:"solutions"`
	Difficulty               uint64  `json:"difficulty"`
	Syncing                  bool    `json:"syncing"`
	SyncProgress             float32 `json:"sync_progress"`
	Listening                bool    `json:"listening"`
	NodeVersion              string  `json:"node_version"`
	NodeRevision             string  `json:"node_revision"`
	PeerCount                uint64  `json:"peer_count"`
	PendingTransactionsCount uint64  `json:"pending_transactions_count"`
	NetworkID                string  `json:"network_id"`
	PeerPubKey               string  `json:"peer_pubkey"`
	TopKeyBlockHash          string  `json:"top_key_block_hash"`
	TopBlockHeight           uint64  `json:"top_block_height"`
}

type Block struct {
	Hash        string   `json:"hash"`
	Height      uint64   `json:"height"`
	PrevHash    string   `json:"prev_hash"`
	PrevKeyHash string   `json:"prev_key_hash"`
	StateHash   string   `json:"state_hash"`
	Miner       string   `json:"miner"`
	Beneficiary string   `json:"beneficiary"`
	Target      uint32   `json:"target"`
	PoW         []uint64 `json:"pow"`
	Nonce       uint64   `json:"nonce"`
	Time        int64    `json:"time"`
	Version     uint64   `json:"version"`
	Info        string   `json:"info"`
}

type DeltaStats struct {
	Data []struct {
		BlockReward json.RawMessage `json:"block_reward"`
		DevReward   uint64          `json:"dev_reward"`
		Height      uint64          `json:"height"`
	} `json:"data"`
}

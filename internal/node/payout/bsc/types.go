package bsc

import (
	"context"
	"encoding/hex"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

func generateHost(url string, logger *log.Logger) (*hostpool.HTTPPool, error) {
	var (
		httpPort        = 443
		httpHealthCheck = &hostpool.HTTPHealthCheck{
			RPCRequest: &rpc.Request{
				JSONRPC: "2.0",
				Method:  "eth_syncing",
			},
		}
	)

	if url == "" {
		return nil, nil
	}

	host := hostpool.NewHTTPPool(context.Background(), logger, httpHealthCheck, nil)
	err := host.AddHost(url, httpPort, nil)

	return host, err
}

func New(mainnet bool, url, rawPriv string, logger *log.Logger) (*Node, error) {
	host, err := generateHost(url, logger)
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
	address := "0x" + hex.EncodeToString(crypto.Keccak256(pubKeyBytes[1:])[12:])

	node := &Node{
		address: address,
		privKey: privKey,
		rpcHost: host,
		logger:  logger,
	}

	return node, nil
}

type Node struct {
	address string
	privKey *secp256k1.PrivateKey
	rpcHost *hostpool.HTTPPool
	logger  *log.Logger
}

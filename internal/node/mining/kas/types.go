package kas

import (
	"context"
	"time"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sencha-dev/powkit/heavyhash"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/node/mining/kas/protowire"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/bech32"
	"github.com/magicpool-co/pool/pkg/hostpool"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
)

var (
	mainnetPrefix = "kaspa"
	testnetPrefix = "kaspatest"
)

func generateHost(urls []string, logger *log.Logger, tunnel *sshtunnel.SSHTunnel) (*hostpool.GRPCPool, error) {
	var (
		port            = 16110
		hostHealthCheck = &hostpool.GRPCHealthCheck{
			Request: &protowire.KaspadMessage{
				Payload: &protowire.KaspadMessage_GetSelectedTipHashRequest{
					GetSelectedTipHashRequest: &protowire.GetSelectedTipHashRequestMessage{},
				},
			},
		}
	)

	if len(urls) == 0 {
		return nil, nil
	}

	factory := func(url string, timeout time.Duration) (hostpool.GRPCClient, error) {
		return protowire.NewClient(url, timeout)
	}

	host := hostpool.NewGRPCPool(context.Background(), factory, logger, hostHealthCheck, tunnel)
	for _, url := range urls {
		err := host.AddHost(url, port)
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

	grpcHost, err := generateHost(urls, logger, tunnel)
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
	pubKeyBytes := privKey.PubKey().SerializeCompressed()
	address, err := bech32.EncodeBCH(addressCharset, prefix, pubKeyECDSAAddrID, pubKeyBytes)
	if err != nil {
		return nil, err
	}

	node := &Node{
		mocked:   grpcHost == nil,
		mainnet:  mainnet,
		prefix:   prefix,
		address:  address,
		privKey:  privKey,
		grpcHost: grpcHost,
		pow:      heavyhash.NewKaspa(),
		logger:   logger,
	}

	return node, nil
}

type Node struct {
	mocked   bool
	mainnet  bool
	prefix   string
	address  string
	privKey  *secp256k1.PrivateKey
	grpcHost *hostpool.GRPCPool
	pow      *heavyhash.Client
	logger   *log.Logger
}

type Block struct {
	Version              uint32
	Parents              [][]string
	HashMerkleRoot       string
	AcceptedIdMerkleRoot string
	UtxoCommitment       string
	Timestamp            int64
	Bits                 uint32
	Nonce                uint64
	DaaScore             uint64
	BlueWork             string
	PruningPoint         string
	BlueScore            uint64
	Transactions         []*Transaction
}

type TransactionOutpoint struct {
	TransactionId string
	Index         uint32
}

type TransactionInput struct {
	PreviousOutpoint *TransactionOutpoint
	SignatureScript  string
	Sequence         uint64
	SigOpCount       uint32
}

type TransactionScriptPubKey struct {
	Version         uint32
	ScriptPublicKey string
}

type TransactionOutput struct {
	Amount          uint64
	ScriptPublicKey *TransactionScriptPubKey
}

type Transaction struct {
	Version      uint32
	Inputs       []*TransactionInput
	Outputs      []*TransactionOutput
	LockTime     uint64
	SubnetworkId string
	Gas          uint64
	Payload      string
}

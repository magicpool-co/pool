package btc

import (
	"github.com/magicpool-co/pool/types"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/base58"
)

var (
	mainnetPrefixP2PKH = []byte{0x00}
	mainnetPrefixP2SH  = []byte{0x05}

	testnetPrefixP2PKH = []byte{0x6f}
	testnetPrefixP2SH  = []byte{0xc4}
)

func New(mainnet bool, rawPriv, blockchairKey string) (*Node, error) {
	prefixP2PKH := mainnetPrefixP2PKH
	prefixP2SH := mainnetPrefixP2SH
	if !mainnet {
		prefixP2PKH = testnetPrefixP2PKH
		prefixP2SH = testnetPrefixP2SH
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
		prefixP2PKH:   prefixP2PKH,
		prefixP2SH:    prefixP2SH,
		address:       address,
		blockchairKey: blockchairKey,
		privKey:       privKey,
	}

	return node, nil
}

type Node struct {
	prefixP2PKH   []byte
	prefixP2SH    []byte
	address       string
	blockchairKey string
	privKey       *secp256k1.PrivateKey
}

func (node Node) Chain() string {
	return "BTC"
}

func (node Node) Address() string {
	return node.address
}

func (node Node) GetUnits() *types.Number {
	return new(types.Number).SetFromValue(1e8)
}

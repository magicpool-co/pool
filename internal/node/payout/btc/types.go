package btc

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
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
	}

	privKey := secp256k1.PrivKeyFromBytes(obscuredPriv)
	address := btctx.PrivKeyToAddress(privKey, prefixP2PKH)

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

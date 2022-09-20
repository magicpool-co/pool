package btc

import (
	"bytes"
	"encoding/hex"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
)

var (
	addressPrefixP2PKH = []byte{0x00}
	addressPrefixP2SH  = []byte{0x05}
)

func GenerateCoinbase(version, lockTime uint32, address string, amount, blockHeight, nTime uint64, extraData, defaultWitness string, prefixP2PKH, prefixP2SH []byte) ([]byte, []byte, error) {
	tx := btctx.NewTransaction(version, lockTime, addressPrefixP2PKH, addressPrefixP2PKH, true)

	serializedBlockHeight := bytes.Join([][]byte{
		crypto.SerializeNumber(blockHeight),
		crypto.SerializeNumber(nTime),
		// extraNonceSize,
		[]byte(extraData),
	}, nil)

	prevTx := "0000000000000000000000000000000000000000000000000000000000000000"
	tx.AddInput(prevTx, 0xFFFFFFFF, 0xFFFFFFFF, serializedBlockHeight)

	scriptPubKey, err := btctx.AddressToScript(address, prefixP2PKH, prefixP2SH, true)
	tx.AddOutput(scriptPubKey, amount)

	if len(defaultWitness) > 0 {
		witness, err := hex.DecodeString(defaultWitness)
		if err != nil {
			return nil, nil, err
		}

		tx.AddOutput(witness, 0)
	}

	serialized, err := tx.Serialize(nil)
	if err != nil {
		return nil, nil, err
	}

	txHash := crypto.ReverseBytes(crypto.Sha256d(serialized))

	return serialized, txHash, nil
}

package rvn

import (
	"bytes"
	"encoding/hex"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
)

const txVersion = 0x1

func GenerateCoinbase(address string, blockReward, blockHeight uint64, extraData, defaultWitness string, prefixP2PKH, prefixP2SH []byte) ([]byte, []byte, error) {
	tx := btctx.NewTransaction(txVersion, 0, prefixP2PKH, prefixP2SH)

	blockHeightSerialBytes, lengthBytes, err := crypto.SerializeBlockHeight(blockHeight)
	if err != nil {
		return nil, nil, err
	}

	serializedBlockHeight := bytes.Join([][]byte{
		lengthBytes,
		crypto.ReverseBytes(blockHeightSerialBytes),
		btctx.EncodeOpCode(0x00),
		[]byte(extraData),
	}, nil)

	prevTx := "0000000000000000000000000000000000000000000000000000000000000000"
	tx.AddInput(prevTx, 0xFFFFFFFF, 0xFFFFFFFF, serializedBlockHeight)

	scriptPubKey, err := btctx.AddressToScript(address, prefixP2PKH, prefixP2SH)
	if err != nil {
		return nil, nil, err
	}

	tx.AddOutput(scriptPubKey, blockReward)

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

package firo

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
)

const txVersion uint32 = 0x50003

func GenerateCoinbase(addresses []string, amounts []uint64, blockHeight, nTime uint64, extraData []byte, extraPayload string, prefixP2PKH, prefixP2SH []byte) ([]byte, []byte, error) {
	if len(addresses) != len(amounts) {
		return nil, nil, fmt.Errorf("address and amount length mismatch")
	} else if len(addresses) == 0 {
		return nil, nil, fmt.Errorf("cannot send transaction without recipients")
	}

	tx := btctx.NewTransaction(txVersion, 0, prefixP2PKH, prefixP2SH)

	extraNonceSize := []byte{0x04}
	if addresses[0] == "a8ULhhDgfdSiXJhSZVdhb8EuDc6R3ogsaM" {
		extraNonceSize = []byte{0x08}
	}

	serializedBlockHeight := bytes.Join([][]byte{
		crypto.SerializeNumber(blockHeight),
		crypto.SerializeNumber(nTime),
		extraNonceSize,
		// crypto.VarIntToBytes(uint64(len(extraData))),
		extraData,
	}, nil)

	prevTx := "0000000000000000000000000000000000000000000000000000000000000000"
	tx.AddInput(prevTx, 0xFFFFFFFF, 0xFFFFFFFF, serializedBlockHeight)

	for i, address := range addresses {
		scriptPubKey, err := btctx.AddressToScript(address, prefixP2PKH, prefixP2SH)
		if err != nil {
			return nil, nil, err
		}

		tx.AddOutput(scriptPubKey, amounts[i])
	}

	extraPayloadBytes, _ := hex.DecodeString(extraPayload)
	serialized, err := tx.Serialize(extraPayloadBytes)
	if err != nil {
		return nil, nil, err
	}

	txHash := crypto.ReverseBytes(crypto.Sha256d(serialized))

	return serialized, txHash, nil
}

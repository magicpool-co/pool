package firo

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/pkg/crypto/wire"
)

const txVersion uint32 = 0x50003

func GenerateCoinbase(addresses []string, amounts []uint64, blockHeight, nTime uint64, extraData []byte, extraPayload string, prefixP2PKH, prefixP2SH []byte) ([]byte, []byte, error) {
	if len(addresses) != len(amounts) {
		return nil, nil, fmt.Errorf("address and amount length mismatch")
	} else if len(addresses) == 0 {
		return nil, nil, fmt.Errorf("cannot send transaction without recipients")
	}

	tx := btctx.NewTransaction(txVersion, 0, prefixP2PKH, prefixP2SH, false)

	var extraNonceSize byte = 0x04
	if addresses[0] == "a8ULhhDgfdSiXJhSZVdhb8EuDc6R3ogsaM" {
		extraNonceSize = 0x08
	}

	var buf bytes.Buffer
	var order = binary.LittleEndian
	if err := wire.WriteSerializedNumber(&buf, order, blockHeight); err != nil {
		return nil, nil, err
	} else if err := wire.WriteSerializedNumber(&buf, order, nTime); err != nil {
		return nil, nil, err
	} else if err := wire.WriteElement(&buf, order, extraNonceSize); err != nil {
		return nil, nil, err
	} else if err := wire.WriteElement(&buf, order, extraData); err != nil {
		return nil, nil, err
	}
	serializedBlockHeight := buf.Bytes()

	prevTx := "0000000000000000000000000000000000000000000000000000000000000000"
	tx.AddInput(prevTx, 0xFFFFFFFF, 0xFFFFFFFF, serializedBlockHeight)

	for i, address := range addresses {
		scriptPubKey, err := btctx.AddressToScript(address, prefixP2PKH, prefixP2SH, false)
		if err != nil {
			return nil, nil, err
		}

		tx.AddOutput(scriptPubKey, amounts[i])
	}

	extraPayloadBytes, err := hex.DecodeString(extraPayload)
	if err != nil {
		return nil, nil, err
	}

	serialized, err := tx.Serialize(extraPayloadBytes)
	if err != nil {
		return nil, nil, err
	}

	txHash := crypto.ReverseBytes(crypto.Sha256d(serialized))

	return serialized, txHash, nil
}

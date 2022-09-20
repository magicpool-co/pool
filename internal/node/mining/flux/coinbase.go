package flux

import (
	"bytes"
	"fmt"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
)

const (
	txVersion             = 0x4
	versionMask    uint32 = 0x80000000
	versionGroupID uint32 = 0x892f2085
	expiryHeight   uint32 = 0
)

func GenerateCoinbase(addresses []string, amounts []uint64, blockHeight uint64, extraData string, prefixP2PKH []byte) ([]byte, []byte, error) {
	if len(addresses) != len(amounts) {
		return nil, nil, fmt.Errorf("address and amount length mismatch")
	} else if len(addresses) == 0 {
		return nil, nil, fmt.Errorf("cannot send transaction without recipients")
	}

	tx := btctx.NewTransaction(txVersion, 0, prefixP2PKH, nil, false)
	tx.SetVersionMask(versionMask)
	tx.SetVersionGroupID(versionGroupID)
	tx.SetExpiryHeight(expiryHeight)

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

	for i, address := range addresses {
		scriptPubKey, err := btctx.AddressToScript(address, prefixP2PKH, nil, false)
		if err != nil {
			return nil, nil, err
		}

		tx.AddOutput(scriptPubKey, amounts[i])
	}

	serialized, err := tx.Serialize(nil)
	if err != nil {
		return nil, nil, err
	}

	txHash := crypto.ReverseBytes(crypto.Sha256d(serialized))

	return serialized, txHash, nil
}

package nexatx

import (
	"encoding/hex"
	"fmt"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/bech32"
	txCommon "github.com/magicpool-co/pool/pkg/crypto/tx"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/types"
)

func privKeyToAddress(privKey *secp256k1.PrivateKey, prefix string) (string, error) {
	pubKeyBytes := privKey.PubKey().SerializeUncompressed()
	pubKeyHash := crypto.Ripemd160(crypto.Sha256(pubKeyBytes))
	return bech32.EncodeBCH(charset, prefix, pubKeyAddrID, pubKeyHash)
}

func GenerateRawTx(baseTx *btctx.Transaction, inputs []*types.TxInput, outputs []*types.TxOutput, fee uint64) (*btctx.Transaction, error) {
	tx := baseTx.ShallowCopy()

	err := txCommon.DistributeFees(inputs, outputs, fee, true)
	if err != nil {
		return nil, err
	}

	for _, inp := range inputs {
		err := tx.AddInput(inp.Hash, inp.Index, 0xFFFFFFFF, nil)
		if err != nil {
			return nil, err
		}
	}

	for _, out := range outputs {
		outputScript, err := AddressToScript(out.Address, string(tx.PrefixP2PKH))
		if err != nil {
			return nil, err
		}
		tx.AddOutput(outputScript, out.Value.Uint64())
	}

	return tx, nil
}

func GenerateSignedTx(privKey *secp256k1.PrivateKey, baseTx *btctx.Transaction, inputs []*types.TxInput, outputs []*types.TxOutput, fee uint64) (*btctx.Transaction, error) {
	rawTx, err := GenerateRawTx(baseTx, inputs, outputs, fee)
	if err != nil {
		return nil, err
	}
	signedTx := baseTx.ShallowCopy()

	address, err := privKeyToAddress(privKey, string(signedTx.PrefixP2PKH))
	if err != nil {
		return nil, err
	}

	inputScript, err := AddressToScript(address, string(signedTx.PrefixP2PKH))
	if err != nil {
		return nil, err
	}

	for i, inp := range inputs {
		inputHash, err := rawTx.CalculateScriptSig(uint32(i), inputScript)
		if err != nil {
			return nil, err
		}

		inputSig := crypto.SchnorrSignBCH(privKey, inputHash).Serialize()
		inputSig = append(inputSig, SIGHASH_ALL)
		scriptSig := generateScriptSig(inputSig)

		err = signedTx.AddInput(inp.Hash, inp.Index, 0xFFFFFFFF, scriptSig)
		if err != nil {
			return nil, err
		}
	}

	for _, out := range outputs {
		outputScript, err := AddressToScript(out.Address, string(signedTx.PrefixP2PKH))
		if err != nil {
			return nil, err
		}
		signedTx.AddOutput(outputScript, out.Value.Uint64())
	}

	return signedTx, nil
}

func GenerateTx(privKey *secp256k1.PrivateKey, baseTx *btctx.Transaction, inputs []*types.TxInput, outputs []*types.TxOutput, feePerByte uint64) ([]byte, error) {
	// generate the tx once to calculate the fee based off of its size
	initialTx, err := GenerateSignedTx(privKey, baseTx, inputs, outputs, 0)
	if err != nil {
		return nil, err
	}

	initialTxSerialized, err := initialTx.Serialize(nil)
	if err != nil {
		return nil, err
	} else if len(initialTxSerialized) > 50000 { // non-standard limit is actually 100000
		return nil, fmt.Errorf("transaction is non-standard with size of %d", len(initialTxSerialized))
	}

	fee := uint64(len(initialTxSerialized)) * feePerByte
	finalTx, err := GenerateSignedTx(privKey, baseTx, inputs, outputs, fee)
	if err != nil {
		return nil, err
	}

	finalTxSerialized, err := finalTx.Serialize(nil)
	if err != nil {
		return nil, err
	}

	return finalTxSerialized, nil
}

func CalculateTxID(tx string) string {
	txBytes, err := hex.DecodeString(tx)
	if err != nil {
		return ""
	}

	txid := crypto.ReverseBytes(crypto.Sha256d(txBytes))

	return hex.EncodeToString(txid)
}

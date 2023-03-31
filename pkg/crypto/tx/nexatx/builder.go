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

func GenerateRawTx(baseTx *Transaction, inputs []*types.TxInput, outputs []*types.TxOutput, fee uint64) (*Transaction, error) {
	tx := baseTx.ShallowCopy()
	err := txCommon.DistributeFees(inputs, outputs, fee, true)
	if err != nil {
		return nil, err
	}

	for _, inp := range inputs {
		err := tx.AddInput(inp.Hash, inp.Index, 0xFFFFFFFF, nil, inp.Value.Uint64())
		if err != nil {
			return nil, err
		}
	}

	for _, out := range outputs {
		outputVersion, outputScript, err := AddressToScript(out.Address, baseTx.Prefix)
		if err != nil {
			return nil, err
		}

		tx.AddOutput(outputVersion, outputScript, out.Value.Uint64())
	}

	return tx, nil
}

func GenerateSignedTx(privKey *secp256k1.PrivateKey, baseTx *Transaction, inputs []*types.TxInput, outputs []*types.TxOutput, fee uint64) (*Transaction, error) {
	rawTx, err := GenerateRawTx(baseTx, inputs, outputs, fee)
	if err != nil {
		return nil, err
	}
	signedTx := baseTx.ShallowCopy()

	address, err := privKeyToAddress(privKey, signedTx.Prefix)
	if err != nil {
		return nil, err
	}

	_, inputScript, err := AddressToScript(address, signedTx.Prefix)
	if err != nil {
		return nil, err
	}

	for i, inp := range inputs {
		inputHash, err := rawTx.CalculateScriptSig(uint32(i), inputScript)
		if err != nil {
			return nil, err
		}

		inputSig := crypto.SchnorrSignBCH(privKey, inputHash).Serialize()
		scriptSig := btctx.GenerateScriptSig(inputSig, privKey.PubKey().SerializeUncompressed())

		err = signedTx.AddInput(inp.Hash, inp.Index, 0xFFFFFFFF, scriptSig, inp.Value.Uint64())
		if err != nil {
			return nil, err
		}
	}

	for _, out := range outputs {
		outputVersion, outputScript, err := AddressToScript(out.Address, signedTx.Prefix)
		if err != nil {
			return nil, err
		}

		signedTx.AddOutput(outputVersion, outputScript, out.Value.Uint64())
	}

	return signedTx, nil
}

func GenerateTx(privKey *secp256k1.PrivateKey, baseTx *Transaction, inputs []*types.TxInput, outputs []*types.TxOutput, feePerByte uint64) ([]byte, error) {
	// generate the tx once to calculate the fee based off of its size
	initialTx, err := GenerateSignedTx(privKey, baseTx, inputs, outputs, 0)
	if err != nil {
		return nil, err
	}

	initialTxSerialized, err := initialTx.Serialize(true)
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

	finalTxSerialized, err := finalTx.Serialize(true)
	if err != nil {
		return nil, err
	}

	return finalTxSerialized, nil
}

func CalculateTxIdem(rawTx string) string {
	txBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		return ""
	}

	tx := new(Transaction)
	err = tx.Deserialize(txBytes)
	if err != nil {
		return ""
	}

	txIdem, err := tx.CalculateTxIdem()
	if err != nil {
		return ""
	}

	return hex.EncodeToString(crypto.ReverseBytes(txIdem))
}

func CalculateTxID(rawTx string) string {
	txBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		return ""
	}

	tx := new(Transaction)
	err = tx.Deserialize(txBytes)
	if err != nil {
		return ""
	}

	txid, err := tx.CalculateTxID()
	if err != nil {
		return ""
	}

	return hex.EncodeToString(crypto.ReverseBytes(txid))
}

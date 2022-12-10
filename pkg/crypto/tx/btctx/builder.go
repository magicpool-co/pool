package btctx

import (
	"encoding/hex"
	"fmt"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1signer "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/base58"
	txCommon "github.com/magicpool-co/pool/pkg/crypto/tx"
	"github.com/magicpool-co/pool/types"
)

func privKeyToAddress(privKey *secp256k1.PrivateKey, prefixP2PKH []byte) string {
	pubKeyBytes := privKey.PubKey().SerializeUncompressed()
	pubKeyHash := crypto.Ripemd160(crypto.Sha256(pubKeyBytes))
	address := base58.CheckEncode(prefixP2PKH, pubKeyHash)

	return address
}

func GenerateRawTx(baseTx *transaction, inputs []*types.TxInput, outputs []*types.TxOutput, fee uint64) (*transaction, error) {
	tx := baseTx.shallowCopy()

	err := txCommon.DistributeFees(inputs, outputs, fee)
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
		outputScript, err := AddressToScript(out.Address, tx.PrefixP2PKH, tx.PrefixP2SH, tx.SegwitEnabled)
		if err != nil {
			return nil, err
		}
		tx.AddOutput(outputScript, out.Value.Uint64())
	}

	return tx, nil
}

func GenerateSignedTx(privKey *secp256k1.PrivateKey, baseTx *transaction, inputs []*types.TxInput, outputs []*types.TxOutput, fee uint64) (*transaction, error) {
	rawTx, err := GenerateRawTx(baseTx, inputs, outputs, fee)
	if err != nil {
		return nil, err
	}
	signedTx := baseTx.shallowCopy()

	address := privKeyToAddress(privKey, signedTx.PrefixP2PKH)
	inputScript, err := AddressToScript(address, signedTx.PrefixP2PKH, signedTx.PrefixP2SH, signedTx.SegwitEnabled)
	if err != nil {
		return nil, err
	}

	for i, inp := range inputs {
		inputHash, err := rawTx.CalculateScriptSig(uint32(i), inputScript)
		if err != nil {
			return nil, err
		}

		inputSig := secp256k1signer.Sign(privKey, inputHash).Serialize()
		inputSig = append(inputSig, SIGHASH_ALL)
		scriptSig := generateScriptSig(inputSig, privKey.PubKey().SerializeUncompressed())

		err = signedTx.AddInput(inp.Hash, inp.Index, 0xFFFFFFFF, scriptSig)
		if err != nil {
			return nil, err
		}
	}

	for _, out := range outputs {
		outputScript, err := AddressToScript(out.Address, signedTx.PrefixP2PKH, signedTx.PrefixP2SH, signedTx.SegwitEnabled)
		if err != nil {
			return nil, err
		}
		signedTx.AddOutput(outputScript, out.Value.Uint64())
	}

	return signedTx, nil
}

func GenerateTx(privKey *secp256k1.PrivateKey, baseTx *transaction, inputs []*types.TxInput, outputs []*types.TxOutput, feePerByte uint64) ([]byte, error) {
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

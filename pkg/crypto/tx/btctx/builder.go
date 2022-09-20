package btctx

import (
	"fmt"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1signer "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/base58"
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

	var sumInputAmount uint64
	for _, inp := range inputs {
		sumInputAmount += inp.Value.Uint64()
		err := tx.AddInput(inp.Hash, inp.Index, 0xFFFFFFFF, nil)
		if err != nil {
			return nil, err
		}
	}

	var sumOutputAmount, sumSplitAmount uint64
	for _, out := range outputs {
		sumOutputAmount += out.Value.Uint64()
		if out.SplitFee {
			sumSplitAmount += out.Value.Uint64()
		}
	}

	if sumOutputAmount != sumInputAmount {
		return nil, fmt.Errorf("amount mismatch: input %d, output %d", sumInputAmount, sumOutputAmount)
	} else if sumSplitAmount < fee {
		return nil, fmt.Errorf("not enough fee space: have %d, want %d", sumSplitAmount, fee)
	}

	var usedFee uint64
	for _, out := range outputs {
		if out.SplitFee {
			out.Fees = (fee * out.Value.Uint64()) / sumSplitAmount
			usedFee += out.Fees
		}
	}

	if usedFee > fee {
		return nil, fmt.Errorf("used too much fees: have %d, want %d", usedFee, fee)
	} else if usedFee < fee {
		for _, out := range outputs {
			if out.SplitFee {
				out.Fees += (fee - usedFee)
				break
			}
		}
	}

	for _, out := range outputs {
		outputScript, err := AddressToScript(out.Address, tx.PrefixP2PKH, tx.PrefixP2SH, tx.SegwitEnabled)
		if err != nil {
			return nil, err
		}
		tx.AddOutput(outputScript, out.Value.Uint64()-out.Fees)
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
		signedTx.AddOutput(outputScript, out.Value.Uint64()-out.Fees)
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
	} else if len(initialTxSerialized) > 50000 {
		// non-standard limit is actually 100000, but we should never come close to that
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

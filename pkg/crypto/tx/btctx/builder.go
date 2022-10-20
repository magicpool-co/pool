package btctx

import (
	"encoding/hex"
	"fmt"
	"math/big"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1signer "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	"github.com/magicpool-co/pool/pkg/common"
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

	var sumOutputAmount uint64
	sumSplitAmount := new(big.Int)
	for _, out := range outputs {
		sumOutputAmount += out.Value.Uint64()
		if out.SplitFee {
			sumSplitAmount.Add(sumSplitAmount, out.Value)
		}
	}

	if sumOutputAmount != sumInputAmount {
		return nil, fmt.Errorf("amount mismatch: input %d, output %d", sumInputAmount, sumOutputAmount)
	}

	usedFees := new(big.Int)
	for _, out := range outputs {
		if !out.SplitFee {
			out.Fee = new(big.Int)
			continue
		}

		out.Fee = new(big.Int).Mul(new(big.Int).SetUint64(fee), out.Value)
		out.Fee.Div(out.Fee, sumSplitAmount)
		out.Value.Sub(out.Value, out.Fee)
		usedFees.Add(usedFees, out.Fee)
	}

	remainder := new(big.Int).Sub(new(big.Int).SetUint64(fee), usedFees)
	if remainder.Cmp(common.Big0) < 0 {
		return nil, fmt.Errorf("fee remainder is negative")
	} else if remainder.Cmp(common.Big0) > 0 {
		for _, output := range outputs {
			if !output.SplitFee {
				continue
			} else if output.Value.Cmp(remainder) > 0 {
				output.Value.Sub(output.Value, remainder)
				remainder = new(big.Int)
				break
			}
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

func CalculateTxID(tx string) string {
	txBytes, err := hex.DecodeString(tx)
	if err != nil {
		return ""
	}

	txid := crypto.ReverseBytes(crypto.Sha256d(txBytes))

	return hex.EncodeToString(txid)
}

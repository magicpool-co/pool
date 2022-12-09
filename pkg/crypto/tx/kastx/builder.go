package kastx

import (
	"encoding/hex"
	"fmt"
	"math/big"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1signer "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"google.golang.org/protobuf/proto"

	"github.com/magicpool-co/pool/internal/node/mining/kas/protowire"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/bech32"
	"github.com/magicpool-co/pool/types"
)

func privKeyToAddress(privKey *secp256k1.PrivateKey, prefix string) (string, error) {
	const addressCharset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	pubKeyBytes := privKey.PubKey().SerializeCompressed()
	return bech32.EncodeBCH(addressCharset, prefix, pubKeyECDSAAddrID, pubKeyBytes)
}

func generateUnsignedTx(inputs []*types.TxInput, outputs []*types.TxOutput, prefix string) (*protowire.RpcTransaction, error) {
	txInputs := make([]*protowire.RpcTransactionInput, len(inputs))
	for i, input := range inputs {
		txInputs[i] = &protowire.RpcTransactionInput{
			PreviousOutpoint: &protowire.RpcOutpoint{
				TransactionId: input.Hash,
				Index:         input.Index,
			},
			SignatureScript: "",
			Sequence:        0,
			SigOpCount:      1,
		}
	}

	txOutputs := make([]*protowire.RpcTransactionOutput, len(outputs))
	for i, output := range outputs {
		script, err := AddressToScript(output.Address, prefix)
		if err != nil {
			return nil, err
		}

		txOutputs[i] = &protowire.RpcTransactionOutput{
			Amount: output.Value.Uint64(),
			ScriptPublicKey: &protowire.RpcScriptPublicKey{
				Version:         0,
				ScriptPublicKey: hex.EncodeToString(script),
			},
		}
	}

	unsignedTx := &protowire.RpcTransaction{
		Version:      0,
		Inputs:       txInputs,
		Outputs:      txOutputs,
		LockTime:     0,
		SubnetworkId: "0000000000000000000000000000000000000000",
		Gas:          0,
		Payload:      "",
	}

	return unsignedTx, nil
}

func signTx(privKey *secp256k1.PrivateKey, tx *protowire.RpcTransaction, inputs []*types.TxInput, prefix string) (*protowire.RpcTransaction, error) {
	address, err := privKeyToAddress(privKey, prefix)
	if err != nil {
		return nil, err
	}

	inputScript, err := AddressToScript(address, prefix)
	if err != nil {
		return nil, err
	}

	for i, inp := range inputs {
		inputHash, err := calculateScriptSig(tx, uint32(i), inp.Value.Uint64(), inputScript)
		if err != nil {
			return nil, err
		}

		inputSig := secp256k1signer.SignCompact(privKey, inputHash, true)
		// remove the recovery code from the secp256k1 signature since kaspa doesnt support it
		if len(inputSig) > 1 {
			inputSig = inputSig[1:]
		}
		inputSig = append(inputSig, byte(SIGHASH_ALL))
		tx.Inputs[i].SignatureScript = hex.EncodeToString(generateScriptSig(inputSig))
	}

	return tx, nil
}

func GenerateTx(privKey *secp256k1.PrivateKey, inputs []*types.TxInput, outputs []*types.TxOutput, prefix string, feePerInput uint64) ([]byte, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("need at least one input")
	} else if len(outputs) == 0 {
		return nil, fmt.Errorf("need at least one output")
	}

	fee := feePerInput * uint64(len(inputs))

	var sumInputAmount uint64
	for _, inp := range inputs {
		sumInputAmount += inp.Value.Uint64()
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

	if remainder.Cmp(new(big.Int)) > 0 {
		return nil, fmt.Errorf("not enough to spend fees")
	}

	unsignedTx, err := generateUnsignedTx(inputs, outputs, prefix)
	if err != nil {
		return nil, err
	}

	signedTx, err := signTx(privKey, unsignedTx, inputs, prefix)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(signedTx)
}

func CalculateTxID(txHex string) string {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return ""
	}

	tx := new(protowire.RpcTransaction)
	err = proto.Unmarshal(txBytes, tx)
	if err != nil {
		return ""
	}

	txid, err := serializeFull(tx)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(txid)
}

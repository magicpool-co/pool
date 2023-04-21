package kastx

import (
	"encoding/hex"
	"fmt"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1signer "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"google.golang.org/protobuf/proto"

	"github.com/magicpool-co/pool/internal/node/mining/kas/protowire"
	"github.com/magicpool-co/pool/pkg/crypto/bech32"
	txCommon "github.com/magicpool-co/pool/pkg/crypto/tx"
	"github.com/magicpool-co/pool/types"
)

const (
	defaultMassPerTxByte           = 1
	defaultMassPerScriptPubKeyByte = 10
	defaultMassPerSigOp            = 1000

	MaximumTxMass = 100000
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

func GenerateTx(
	privKey *secp256k1.PrivateKey,
	inputs []*types.TxInput,
	outputs []*types.TxOutput,
	prefix string,
	feePerInput uint64,
) ([]byte, uint64, error) {
	if len(inputs) == 0 {
		return nil, 0, fmt.Errorf("need at least one input")
	} else if len(outputs) == 0 {
		return nil, 0, fmt.Errorf("need at least one output")
	}

	fee := feePerInput * uint64(len(inputs))
	err := txCommon.DistributeFees(inputs, outputs, fee, true)
	if err != nil {
		return nil, 0, err
	}

	unsignedTx, err := generateUnsignedTx(inputs, outputs, prefix)
	if err != nil {
		return nil, 0, err
	}

	signedTx, err := signTx(privKey, unsignedTx, inputs, prefix)
	if err != nil {
		return nil, 0, err
	}

	txHex, err := proto.Marshal(signedTx)
	if err != nil {
		return nil, 0, err
	}

	txMass := CalculateTxMass(signedTx, txHex)

	return txHex, txMass, nil
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

func CalculateTxMass(tx *protowire.RpcTransaction, txHex []byte) uint64 {
	// calculate mass for size
	size := uint64(len(txHex))
	fmt.Println(size, len(tx.Inputs), len(tx.Outputs))
	massForSize := size * defaultMassPerTxByte

	// calculate mass for scriptPubKey
	var totalScriptPubKeySize uint64
	for _, output := range tx.Outputs {
		totalScriptPubKeySize += 2
		totalScriptPubKeySize += uint64(len(output.ScriptPublicKey.ScriptPublicKey) / 2)
	}
	massForScriptPubKey := totalScriptPubKeySize * defaultMassPerScriptPubKeyByte

	// calculate mass for SigOps
	var totalSigOpCount uint64
	for _, input := range tx.Inputs {
		totalSigOpCount += uint64(input.SigOpCount)
	}
	massForSigOps := totalSigOpCount * defaultMassPerSigOp

	// Sum all components of mass
	return massForSize + massForScriptPubKey + massForSigOps
}

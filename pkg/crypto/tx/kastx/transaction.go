package kastx

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/magicpool-co/pool/internal/node/mining/kas/protowire"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/wire"
)

var (
	transactionSigningHashDomain = []byte("TransactionSigningHash")
	transactionIDDomain          = []byte("TransactionID")

	transactionSigningECDSADomainHash = sha256.Sum256([]byte("TransactionSigningHashECDSA"))
)

func calculatePreviousOutputHash(order binary.ByteOrder, inputs []*protowire.RpcTransactionInput) ([]byte, error) {
	var buf bytes.Buffer
	for _, input := range inputs {
		transactionID, err := hex.DecodeString(input.PreviousOutpoint.TransactionId)
		if err != nil {
			return nil, err
		} else if err := wire.WriteElement(&buf, order, transactionID); err != nil {
			return nil, err
		} else if err := wire.WriteElement(&buf, order, input.PreviousOutpoint.Index); err != nil {
			return nil, err
		}
	}

	return crypto.Blake2b256MAC(buf.Bytes(), transactionSigningHashDomain)
}

func calculateSequencesHash(order binary.ByteOrder, inputs []*protowire.RpcTransactionInput) ([]byte, error) {
	var buf bytes.Buffer
	for _, input := range inputs {
		if err := wire.WriteElement(&buf, order, input.Sequence); err != nil {
			return nil, err
		}
	}

	return crypto.Blake2b256MAC(buf.Bytes(), transactionSigningHashDomain)
}

func calculateSigOpsCountsHash(order binary.ByteOrder, inputs []*protowire.RpcTransactionInput) ([]byte, error) {
	var buf bytes.Buffer
	for _, input := range inputs {
		if err := wire.WriteElement(&buf, order, byte(input.SigOpCount)); err != nil {
			return nil, err
		}
	}

	return crypto.Blake2b256MAC(buf.Bytes(), transactionSigningHashDomain)
}

func calculateOutputsHash(order binary.ByteOrder, outputs []*protowire.RpcTransactionOutput) ([]byte, error) {
	var buf bytes.Buffer
	for _, output := range outputs {
		scriptPubKey, err := hex.DecodeString(output.ScriptPublicKey.ScriptPublicKey)
		if err != nil {
			return nil, err
		} else if err := wire.WriteElement(&buf, order, output.Amount); err != nil {
			return nil, err
		} else if err := wire.WriteElement(&buf, order, uint16(output.ScriptPublicKey.Version)); err != nil {
			return nil, err
		} else if err := wire.WritePrefixedBytes(&buf, order, scriptPubKey); err != nil {
			return nil, err
		}
	}

	return crypto.Blake2b256MAC(buf.Bytes(), transactionSigningHashDomain)
}

func serializePartial(tx *protowire.RpcTransaction, idx uint32, amount uint64, scriptPubKey []byte) ([]byte, error) {
	if len(tx.Inputs) <= int(idx) {
		return nil, fmt.Errorf("index out of bounds")
	}

	var order = binary.LittleEndian
	var buf bytes.Buffer

	if err := wire.WriteElement(&buf, order, uint16(tx.Version)); err != nil { // version
		return nil, err
	}

	previousOutputsHash, err := calculatePreviousOutputHash(order, tx.Inputs)
	if err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, previousOutputsHash); err != nil { // previousOutputsHash
		return nil, err
	}

	sequencesHash, err := calculateSequencesHash(order, tx.Inputs)
	if err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, sequencesHash); err != nil { // sequencesHash
		return nil, err
	}

	sigOpsCountsHash, err := calculateSigOpsCountsHash(order, tx.Inputs)
	if err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, sigOpsCountsHash); err != nil { // sigOpsCountsHash
		return nil, err
	}

	previousOutpointTransactionID, err := hex.DecodeString(tx.Inputs[idx].PreviousOutpoint.TransactionId)
	if err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, previousOutpointTransactionID); err != nil { // input.PreviousOutpoint.TransactionID
		return nil, err
	} else if err := wire.WriteElement(&buf, order, uint32(tx.Inputs[idx].PreviousOutpoint.Index)); err != nil { // input.PreviousOutpoint.Index
		return nil, err
	}

	if err := wire.WriteElement(&buf, order, uint16(0)); err != nil { // input.ScriptPublicKey.Version
		return nil, err
	} else if err := wire.WritePrefixedBytes(&buf, order, scriptPubKey); err != nil { // input.ScriptPublicKey.ScriptPublicKey
		return nil, err
	}

	if err := wire.WriteElement(&buf, order, amount); err != nil { // input.Amount
		return nil, err
	} else if err := wire.WriteElement(&buf, order, tx.Inputs[idx].Sequence); err != nil { // input.Sequence
		return nil, err
	} else if err := wire.WriteElement(&buf, order, byte(tx.Inputs[idx].SigOpCount)); err != nil { // input.SigOpCount
		return nil, err
	}

	outputsHash, err := calculateOutputsHash(order, tx.Outputs)
	if err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, outputsHash); err != nil { // outputsHash
		return nil, err
	}

	if err := wire.WriteElement(&buf, order, tx.LockTime); err != nil { // lockTime
		return nil, err
	} else if err := wire.WriteElement(&buf, order, make([]byte, 20)); err != nil { // subnetworkID
		return nil, err
	} else if err := wire.WriteElement(&buf, order, tx.Gas); err != nil { // gas
		return nil, err
	} else if err := wire.WriteElement(&buf, order, make([]byte, 32)); err != nil { // payload
		return nil, err
	} else if err := wire.WriteElement(&buf, order, uint8(SIGHASH_ALL)); err != nil { // hashType
		return nil, err
	}

	return crypto.Blake2b256MAC(buf.Bytes(), transactionSigningHashDomain)
}

func serializeFull(tx *protowire.RpcTransaction) ([]byte, error) {
	var order = binary.LittleEndian
	var buf bytes.Buffer
	var isCoinbase = tx.SubnetworkId == "0100000000000000000000000000000000000000"

	if err := wire.WriteElement(&buf, order, uint16(tx.Version)); err != nil { // version
		return nil, err
	}

	if err := wire.WriteElement(&buf, order, uint64(len(tx.Inputs))); err != nil { // numInputs
		return nil, err
	}

	for _, input := range tx.Inputs {
		previousOutpointTransactionID, err := hex.DecodeString(input.PreviousOutpoint.TransactionId)
		if err != nil {
			return nil, err
		} else if err := wire.WriteElement(&buf, order, previousOutpointTransactionID); err != nil { // input.PreviousOutpoint.TransactionID
			return nil, err
		} else if err := wire.WriteElement(&buf, order, uint32(input.PreviousOutpoint.Index)); err != nil { // input.PreviousOutpoint.Index
			return nil, err
		}

		if isCoinbase {
			if err := wire.WritePrefixedHexString(&buf, order, input.SignatureScript); err != nil { // input.SignatureScript (empty)
				return nil, err
			}
		} else {
			if err := wire.WritePrefixedBytes(&buf, order, make([]byte, 0)); err != nil { // input.SignatureScript (empty)
				return nil, err
			}
		}

		if err := wire.WriteElement(&buf, order, input.Sequence); err != nil { // input.Sequence
			return nil, err
		}
	}

	if err := wire.WriteElement(&buf, order, uint64(len(tx.Outputs))); err != nil { // numOutputs
		return nil, err
	}

	for _, output := range tx.Outputs {
		scriptPubKey, err := hex.DecodeString(output.ScriptPublicKey.ScriptPublicKey)
		if err != nil {
			return nil, err
		} else if err := wire.WriteElement(&buf, order, output.Amount); err != nil { // output.Amount
			return nil, err
		} else if err := wire.WriteElement(&buf, order, uint16(output.ScriptPublicKey.Version)); err != nil { // output.ScriptPublicKey.Version
			return nil, err
		} else if err := wire.WritePrefixedBytes(&buf, order, scriptPubKey); err != nil { // output.ScriptPublicKey.ScriptPublicKey
			return nil, err
		}
	}

	if err := wire.WriteElement(&buf, order, tx.LockTime); err != nil { // lockTime
		return nil, err
	} else if err := wire.WriteHexString(&buf, order, tx.SubnetworkId); err != nil { // subnetworkID
		return nil, err
	} else if err := wire.WriteElement(&buf, order, tx.Gas); err != nil { // gas
		return nil, err
	} else if err := wire.WritePrefixedHexString(&buf, order, tx.Payload); err != nil { // payload
		return nil, err
	}

	return crypto.Blake2b256MAC(buf.Bytes(), transactionIDDomain)
}

func calculateScriptSig(tx *protowire.RpcTransaction, index uint32, amount uint64, scriptPubKey []byte) ([]byte, error) {
	serialized, err := serializePartial(tx, index, amount, scriptPubKey)
	if err != nil {
		return nil, err
	}

	hasher := sha256.New()
	hasher.Write(transactionSigningECDSADomainHash[:])
	hasher.Write(serialized)

	return hasher.Sum(nil), nil
}

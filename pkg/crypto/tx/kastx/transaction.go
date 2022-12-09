package kastx

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/blake2b"

	"github.com/magicpool-co/pool/internal/node/mining/kas/protowire"
)

var transactionSigningECDSADomainHash = sha256.Sum256([]byte("TransactionSigningHashECDSA"))

func writeUint8(writer []byte, value uint8) int {
	const size = 1
	writer[0] = value

	return size
}

func writeUint16(writer []byte, value uint16) int {
	const size = 2
	binary.LittleEndian.PutUint16(writer[:size], value)

	return size
}

func writeUint32(writer []byte, value uint32) int {
	const size = 4
	binary.LittleEndian.PutUint32(writer[:size], value)

	return size
}

func writeUint64(writer []byte, value uint64) int {
	const size = 8
	binary.LittleEndian.PutUint64(writer[:size], value)

	return size
}

func writeBytes(writer []byte, value []byte) int {
	copy(writer, value)

	return len(value)
}

func writeHex(writer []byte, value string, length int) (int, error) {
	data, err := hex.DecodeString(value)
	if err != nil {
		return 0, err
	} else if len(data) != length {
		return 0, fmt.Errorf("length mismatch: %d, %d", len(data), length)
	}
	copy(writer[:length], data)

	return length, nil
}

func writeOutpoint(writer []byte, outpoint *protowire.RpcOutpoint) (int, error) {
	pos, err := writeHex(writer, outpoint.TransactionId, 32)
	if err != nil {
		return 0, err
	}
	pos += writeUint32(writer[pos:], outpoint.Index)

	return pos, nil
}

func writeTxOutput(writer []byte, output *protowire.RpcTransactionOutput) (int, error) {
	var pos int
	pos += writeUint64(writer[pos:], output.Amount)
	pos += writeUint16(writer[pos:], uint16(output.ScriptPublicKey.Version))
	offset, err := writeHex(writer[pos:], output.ScriptPublicKey.ScriptPublicKey, len(output.ScriptPublicKey.ScriptPublicKey)/2)
	if err != nil {
		return 0, err
	}

	return pos + offset, nil
}

func transactionSigningHash(input []byte) []byte {
	blake, _ := blake2b.New256([]byte("TransactionSigningHash"))
	blake.Write(input)

	return blake.Sum(nil)
}

func transactionIDSigningHash(input []byte) []byte {
	blake, _ := blake2b.New256([]byte("TransactionID"))
	blake.Write(input)

	return blake.Sum(nil)
}

func calculatePreviousOutputHash(inputs []*protowire.RpcTransactionInput) ([]byte, error) {
	var pos int
	writer := make([]byte, len(inputs)*(32+4))
	for _, input := range inputs {
		offset, err := writeHex(writer[pos:], input.PreviousOutpoint.TransactionId, 32)
		if err != nil {
			return nil, err
		}
		pos += offset
		pos += writeUint32(writer[pos:], input.PreviousOutpoint.Index)
	}

	return transactionSigningHash(writer), nil
}

func calculateSequencesHash(inputs []*protowire.RpcTransactionInput) []byte {
	var pos int
	writer := make([]byte, len(inputs)*8)
	for _, input := range inputs {
		pos += writeUint64(writer[pos:], input.Sequence)
	}

	return transactionSigningHash(writer)
}

func calculateSigOpsCountsHash(inputs []*protowire.RpcTransactionInput) []byte {
	writer := make([]byte, len(inputs))
	for i, input := range inputs {
		writer[i] = byte(input.SigOpCount)
	}

	return transactionSigningHash(writer)
}

func calculateOutputsHash(outputs []*protowire.RpcTransactionOutput) ([]byte, error) {
	var length int
	for _, output := range outputs {
		length += 8 + 2 + 8 + len(output.ScriptPublicKey.ScriptPublicKey)/2
	}

	var pos int
	writer := make([]byte, length)
	for _, output := range outputs {
		scriptPubKeyLength := len(output.ScriptPublicKey.ScriptPublicKey) / 2

		pos += writeUint64(writer[pos:], output.Amount)
		pos += writeUint16(writer[pos:], uint16(output.ScriptPublicKey.Version))
		pos += writeUint64(writer[pos:], uint64(scriptPubKeyLength))
		offset, err := writeHex(writer[pos:], output.ScriptPublicKey.ScriptPublicKey, scriptPubKeyLength)
		if err != nil {
			return nil, err
		}
		pos += offset
	}

	return transactionSigningHash(writer), nil
}

func serializePartialSize(tx *protowire.RpcTransaction, scriptPubKey []byte) int {
	var txSize int
	txSize += 2                     // version
	txSize += 32 * 3                // previousOutputsHash, sequencesHash, sigOpCountsHash
	txSize += 32 + 4                // outpoint
	txSize += 2                     // scriptPubKeyVersion
	txSize += 8 + len(scriptPubKey) // scriptPubKey
	txSize += 8                     // amount
	txSize += 8                     // sequence
	txSize += 1                     // sigOpCount
	txSize += 32                    // outputsHash
	txSize += 8                     // lockTime
	txSize += 20                    // subnetworkId
	txSize += 8                     /// gas
	txSize += 32                    // payload
	txSize += 1                     // hashType

	return txSize
}

func serializePartial(tx *protowire.RpcTransaction, idx uint32, amount uint64, scriptPubKey []byte) ([]byte, error) {
	if len(tx.Inputs) <= int(idx) {
		return nil, fmt.Errorf("index out of bounds")
	}

	txSize := serializePartialSize(tx, scriptPubKey)
	txBytes := make([]byte, txSize)
	var pos int

	// version
	pos += writeUint16(txBytes[pos:], uint16(tx.Version))

	// previousOutputsHash
	previousOutputsHash, err := calculatePreviousOutputHash(tx.Inputs)
	if err != nil {
		return nil, err
	}
	pos += writeBytes(txBytes[pos:], previousOutputsHash)

	// sequencesHash
	sequencesHash := calculateSequencesHash(tx.Inputs)
	pos += writeBytes(txBytes[pos:], sequencesHash)

	// sigOpsCountsHash
	sigOpsCountsHash := calculateSigOpsCountsHash(tx.Inputs)
	pos += writeBytes(txBytes[pos:], sigOpsCountsHash)

	// previousOutpoint
	offset, err := writeHex(txBytes[pos:], tx.Inputs[idx].PreviousOutpoint.TransactionId, 32)
	if err != nil {
		return nil, err
	}
	pos += offset
	pos += writeUint32(txBytes[pos:], tx.Inputs[idx].PreviousOutpoint.Index)

	// scriptPubKeyVersion
	pos += writeUint16(txBytes[pos:], 0)

	// scriptPubKey
	pos += writeUint64(txBytes[pos:], uint64(len(scriptPubKey)))
	copy(txBytes[pos:], scriptPubKey)
	pos += len(scriptPubKey)

	// amount
	pos += writeUint64(txBytes[pos:], amount)

	// sequence
	pos += writeUint64(txBytes[pos:], tx.Inputs[idx].Sequence)

	// sigOpCount
	pos += writeUint8(txBytes[pos:], byte(tx.Inputs[idx].SigOpCount))

	// outputsHash
	outputsHash, err := calculateOutputsHash(tx.Outputs)
	if err != nil {
		return nil, err
	}
	pos += writeBytes(txBytes[pos:], outputsHash)

	// lockTime
	pos += writeUint64(txBytes[pos:], tx.LockTime)

	// subnetworkId
	offset, err = writeHex(txBytes[pos:], tx.SubnetworkId, len(tx.SubnetworkId)/2)
	if err != nil {
		return nil, err
	}
	pos += offset

	// gas
	pos += writeUint64(txBytes[pos:], tx.Gas)

	// payload
	pos += writeBytes(txBytes[pos:], make([]byte, 32))

	// hashType
	pos += writeUint8(txBytes[pos:], SIGHASH_ALL)

	if pos != txSize {
		return nil, fmt.Errorf("mismatch for final tx size")
	}

	return transactionSigningHash(txBytes), nil
}

func serializeFullSize(tx *protowire.RpcTransaction) int {
	var txSize int
	txSize += 2 // version
	txSize += 8 // numInputs

	for _ = range tx.Inputs {
		txSize += 32 + 4 // previousOutput
		txSize += 8      // signatureScript
		txSize += 8      // sequence
	}

	txSize += 8 // numOutputs

	for _, output := range tx.Outputs {
		txSize += 8                                               // amount
		txSize += 2                                               // scriptPubKeyVersion
		txSize += 8                                               // scriptPubKeyLength
		txSize += len(output.ScriptPublicKey.ScriptPublicKey) / 2 // scriptPubKey
	}

	txSize += 8  // lockTime
	txSize += 20 // subnetworkId
	txSize += 8  /// gas
	txSize += 8  // payload (empty)

	return txSize
}

func serializeFull(tx *protowire.RpcTransaction) ([]byte, error) {
	txSize := serializeFullSize(tx)
	txBytes := make([]byte, txSize)
	var pos int

	// version
	pos += writeUint16(txBytes[pos:], uint16(tx.Version))

	// numInputs
	pos += writeUint64(txBytes[pos:], uint64(len(tx.Inputs)))

	for _, input := range tx.Inputs {
		// previousOutpoint
		offset, err := writeHex(txBytes[pos:], input.PreviousOutpoint.TransactionId, 32)
		if err != nil {
			return nil, err
		}
		pos += offset
		pos += writeUint32(txBytes[pos:], input.PreviousOutpoint.Index)

		// signatureScript (empty)
		pos += writeUint64(txBytes[pos:], 0)

		// sequence
		pos += writeUint64(txBytes[pos:], input.Sequence)
	}

	// numOutputs
	pos += writeUint64(txBytes[pos:], uint64(len(tx.Outputs)))

	for _, output := range tx.Outputs {
		scriptPubKeyLength := len(output.ScriptPublicKey.ScriptPublicKey) / 2

		pos += writeUint64(txBytes[pos:], output.Amount)
		pos += writeUint16(txBytes[pos:], uint16(output.ScriptPublicKey.Version))
		pos += writeUint64(txBytes[pos:], uint64(scriptPubKeyLength))
		offset, err := writeHex(txBytes[pos:], output.ScriptPublicKey.ScriptPublicKey, scriptPubKeyLength)
		if err != nil {
			return nil, err
		}
		pos += offset
	}

	// lockTime
	pos += writeUint64(txBytes[pos:], tx.LockTime)

	// subnetworkId
	offset, err := writeHex(txBytes[pos:], tx.SubnetworkId, len(tx.SubnetworkId)/2)
	if err != nil {
		return nil, err
	}
	pos += offset

	// gas
	pos += writeUint64(txBytes[pos:], tx.Gas)

	// payload
	pos += writeUint64(txBytes[pos:], 0)

	if pos != txSize {
		return nil, fmt.Errorf("mismatch for final tx size")
	}

	return transactionIDSigningHash(txBytes), nil
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

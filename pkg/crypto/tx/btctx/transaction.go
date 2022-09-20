package btctx

import (
	"bytes"
	"encoding/hex"

	"github.com/magicpool-co/pool/pkg/crypto"
)

type transaction struct {
	PrefixP2PKH    []byte
	PrefixP2SH     []byte
	SegwitEnabled  bool
	Version        uint32
	VersionGroupID *uint32
	LockTime       uint32
	ExpiryHeight   *uint32
	Inputs         []*input
	Outputs        []*output
}

func NewTransaction(version, lockTime uint32, prefixP2PKH, prefixP2SH []byte, segwitEnabled bool) *transaction {
	tx := &transaction{
		PrefixP2PKH:   prefixP2PKH,
		PrefixP2SH:    prefixP2SH,
		SegwitEnabled: segwitEnabled,
		Version:       version,
		LockTime:      lockTime,
	}

	return tx
}

func (tx *transaction) SetVersionMask(versionMask uint32) {
	tx.Version = tx.Version | versionMask
}

func (tx *transaction) SetVersionGroupID(versionGroupID uint32) {
	tx.VersionGroupID = &versionGroupID
}

func (tx *transaction) SetExpiryHeight(expiryHeight uint32) {
	tx.ExpiryHeight = &expiryHeight
}

type witness [][]byte

func (wit witness) SerializeSize() int {
	n := crypto.VarIntSerializeSize(uint64(len(wit)))

	for _, item := range wit {
		n += crypto.VarIntSerializeSize(uint64(len(item))) + len(item)
	}

	return n
}

func (wit witness) Serialize() []byte {
	pos := 0
	out := make([]byte, wit.SerializeSize())

	countLen := uint64(len(wit))
	countLenVarInt := crypto.VarIntToBytes(countLen)
	copy(out[pos:], countLenVarInt)
	pos += len(countLenVarInt)

	for _, w := range wit {
		witLen := uint64(len(w))
		witLenVarInt := crypto.VarIntToBytes(witLen)
		copy(out[pos:], witLenVarInt)
		pos += len(witLenVarInt)

		copy(out[pos:], w)
		pos += len(w)
	}

	return out
}

type input struct {
	PrevHash  []byte
	PrevIndex uint32
	Script    []byte
	Sequence  uint32
	Witness   witness
}

func (inp *input) SerializeSize() int {
	return 40 + crypto.VarIntSerializeSize(uint64(len(inp.Script))) + len(inp.Script)
}

func (inp *input) Serialize() []byte {
	scriptLen := uint64(len(inp.Script))
	scriptLenVarInt := crypto.VarIntToBytes(scriptLen)

	data := bytes.Join([][]byte{
		crypto.ReverseBytes(inp.PrevHash),
		crypto.WriteUint32Le(inp.PrevIndex),
		scriptLenVarInt,
		inp.Script,
		crypto.WriteUint32Le(inp.Sequence),
	}, nil)

	return data
}

type output struct {
	Script []byte
	Value  uint64
}

func (out *output) SerializeSize() int {
	return 8 + crypto.VarIntSerializeSize(uint64(len(out.Script))) + len(out.Script)
}

func (out *output) Serialize() []byte {
	scriptLen := uint64(len(out.Script))
	scriptLenVarInt := crypto.VarIntToBytes(scriptLen)

	data := bytes.Join([][]byte{
		crypto.WriteUint64Le(out.Value),
		scriptLenVarInt,
		out.Script,
	}, nil)

	return data
}

func (tx *transaction) AddInput(hash string, index, sequence uint32, script []byte) error {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return err
	}

	inp := &input{
		PrevHash:  hashBytes,
		PrevIndex: index,
		Sequence:  sequence,
		Script:    script,
	}

	tx.Inputs = append(tx.Inputs, inp)

	return nil
}

func (tx *transaction) AddOutput(script []byte, value uint64) {
	out := &output{
		Script: script,
		Value:  value,
	}

	tx.Outputs = append(tx.Outputs, out)
}

func (tx *transaction) hasWitnesses() bool {
	for _, inp := range tx.Inputs {
		if len(inp.Witness) != 0 {
			return true
		}
	}

	return false
}

func (tx *transaction) shallowCopy() *transaction {
	// As an additional memory optimization, use contiguous backing arrays
	// for the copied inputs and outputs and point the final slice of
	// pointers into the contiguous arrays.  This avoids a lot of small
	// allocations.
	txCopy := &transaction{
		PrefixP2PKH:    tx.PrefixP2PKH,
		PrefixP2SH:     tx.PrefixP2SH,
		SegwitEnabled:  tx.SegwitEnabled,
		Version:        tx.Version,
		VersionGroupID: tx.VersionGroupID,
		Inputs:         make([]*input, len(tx.Inputs)),
		Outputs:        make([]*output, len(tx.Outputs)),
		LockTime:       tx.LockTime,
		ExpiryHeight:   tx.ExpiryHeight,
	}
	inputs := make([]input, len(tx.Inputs))
	for i, oldTxIn := range tx.Inputs {
		inputs[i] = *oldTxIn
		txCopy.Inputs[i] = &inputs[i]
	}
	outputs := make([]output, len(tx.Outputs))
	for i, oldTxOut := range tx.Outputs {
		outputs[i] = *oldTxOut
		txCopy.Outputs[i] = &outputs[i]
	}
	return txCopy
}

func (tx *transaction) SerializeSize(extraPayload []byte) int {
	length := 0

	hasWitnesses := tx.hasWitnesses()
	if hasWitnesses {
		length += 10
	} else {
		length += 8
	}

	if tx.VersionGroupID != nil {
		length += 4 + 11 // @TODO: no idea why this 11 is needed
	}

	length += crypto.VarIntSerializeSize(uint64(len(tx.Inputs)))
	for _, input := range tx.Inputs {
		length += input.SerializeSize()
	}

	length += crypto.VarIntSerializeSize(uint64(len(tx.Outputs)))
	for _, output := range tx.Outputs {
		length += output.SerializeSize()
	}

	if hasWitnesses {
		for _, input := range tx.Inputs {
			length += input.Witness.SerializeSize()
		}
	}

	if tx.ExpiryHeight != nil {
		length += 4
	}

	if len(extraPayload) > 0 {
		length += 1 + len(extraPayload)
	}

	return length
}

func (tx *transaction) Serialize(extraPayload []byte) ([]byte, error) {
	pos := 0
	length := tx.SerializeSize(extraPayload)
	serialized := make([]byte, length)

	// add version
	versionBytes := crypto.WriteUint32Le(tx.Version)
	copy(serialized[pos:], versionBytes)
	pos += len(versionBytes)

	// zcash specific field
	if tx.VersionGroupID != nil {
		versionGroupIDBytes := crypto.WriteUint32Le(*tx.VersionGroupID)
		copy(serialized[pos:], versionGroupIDBytes)
		pos += len(versionGroupIDBytes)
	}

	// if segwit tx, add witness flag
	doWitness := tx.hasWitnesses()
	if doWitness {
		flags := []byte{TxFlagMarker, WitnessFlag}
		copy(serialized[pos:], flags)
		pos += len(flags)
	}

	// add number of inputs
	inputCountBytes := crypto.VarIntToBytes(uint64(len(tx.Inputs)))
	copy(serialized[pos:], inputCountBytes)
	pos += len(inputCountBytes)

	// add inputs
	for _, input := range tx.Inputs {
		inputSerialized := input.Serialize()
		copy(serialized[pos:], inputSerialized)
		pos += len(inputSerialized)
	}

	// add number of outputs
	outputCountBytes := crypto.VarIntToBytes(uint64(len(tx.Outputs)))
	copy(serialized[pos:], outputCountBytes)
	pos += len(outputCountBytes)

	// add outputs
	for _, output := range tx.Outputs {
		outputSerialized := output.Serialize()
		copy(serialized[pos:], outputSerialized)
		pos += len(outputSerialized)
	}

	// if segwit tx, add witness data
	if doWitness {
		for _, input := range tx.Inputs {
			witnessSerialized := input.Witness.Serialize()
			copy(serialized[pos:], witnessSerialized)
			pos += len(witnessSerialized)
		}
	}

	// add locktime
	lockTimeBytes := crypto.WriteUint32Le(uint32(tx.LockTime))
	copy(serialized[pos:], lockTimeBytes)
	pos += len(lockTimeBytes)

	// if zcash, add expiry height
	if tx.ExpiryHeight != nil {
		expiryHeightBytes := crypto.WriteUint32Le(*tx.ExpiryHeight)
		copy(serialized[pos:], expiryHeightBytes)
		pos += len(expiryHeightBytes)
	}

	// if extra payload, add data
	if len(extraPayload) > 0 {
		serialized[pos] = byte(len(extraPayload))
		copy(serialized[pos+1:], extraPayload)
	}

	return serialized, nil
}

func (tx *transaction) CalculateScriptSig(index uint32, script []byte) ([]byte, error) {
	txCopy := tx.shallowCopy()
	for i := range txCopy.Inputs {
		txCopy.Inputs[i].Script = nil
		if uint32(i) == index {
			txCopy.Inputs[i].Script = script
		}
	}

	serialized, err := txCopy.Serialize(nil)
	if err != nil {
		return nil, err
	}

	// add hash type
	serialized = bytes.Join([][]byte{
		serialized,
		crypto.WriteUint32Le(SIGHASH_ALL),
	}, nil)

	return crypto.Sha256d(serialized), nil
}

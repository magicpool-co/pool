// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btctx

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/wire"
)

type Transaction struct {
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

func NewTransaction(
	version, lockTime uint32,
	prefixP2PKH, prefixP2SH []byte,
	segwitEnabled bool,
) *Transaction {
	tx := &Transaction{
		PrefixP2PKH:   prefixP2PKH,
		PrefixP2SH:    prefixP2SH,
		SegwitEnabled: segwitEnabled,
		Version:       version,
		LockTime:      lockTime,
	}

	return tx
}

func (tx *Transaction) SetVersionMask(versionMask uint32) {
	tx.Version = tx.Version | versionMask
}

func (tx *Transaction) SetVersionGroupID(versionGroupID uint32) {
	tx.VersionGroupID = &versionGroupID
}

func (tx *Transaction) SetExpiryHeight(expiryHeight uint32) {
	tx.ExpiryHeight = &expiryHeight
}

type witness [][]byte

func (wit witness) Serialize(buf *bytes.Buffer, order binary.ByteOrder) error {
	return wire.WriteVarByteArray(buf, order, wit)
}

type input struct {
	PrevHash  []byte
	PrevIndex uint32
	Script    []byte
	Sequence  uint32
	Witness   witness
}

func (inp *input) Serialize(buf *bytes.Buffer, order binary.ByteOrder) error {
	if err := wire.WriteElement(buf, order, crypto.ReverseBytes(inp.PrevHash)); err != nil {
		return err
	} else if err := wire.WriteElement(buf, order, inp.PrevIndex); err != nil {
		return err
	} else if err := wire.WriteVarBytes(buf, order, inp.Script); err != nil {
		return err
	} else if err := wire.WriteElement(buf, order, inp.Sequence); err != nil {
		return err
	}

	return nil
}

type output struct {
	Script []byte
	Value  uint64
}

func (out *output) Serialize(buf *bytes.Buffer, order binary.ByteOrder) error {
	if err := wire.WriteElement(buf, order, out.Value); err != nil {
		return err
	} else if err := wire.WriteVarBytes(buf, order, out.Script); err != nil {
		return err
	}

	return nil
}

func (tx *Transaction) AddInput(hash string, index, sequence uint32, script []byte) error {
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

func (tx *Transaction) AddOutput(script []byte, value uint64) {
	out := &output{
		Script: script,
		Value:  value,
	}

	tx.Outputs = append(tx.Outputs, out)
}

func (tx *Transaction) hasWitnesses() bool {
	for _, inp := range tx.Inputs {
		if len(inp.Witness) != 0 {
			return true
		}
	}

	return false
}

func (tx *Transaction) ShallowCopy() *Transaction {
	txCopy := &Transaction{
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

func (tx *Transaction) Serialize(extraPayload []byte) ([]byte, error) {
	var order = binary.LittleEndian
	var buf bytes.Buffer

	// add version
	if err := wire.WriteElement(&buf, order, tx.Version); err != nil {
		return nil, err
	}

	// zcash specific field
	if tx.VersionGroupID != nil {
		if err := wire.WriteElement(&buf, order, *tx.VersionGroupID); err != nil {
			return nil, err
		}
	}

	// if segwit tx, add witness flag
	doWitness := tx.hasWitnesses()
	if doWitness {
		flags := []byte{TxFlagMarker, WitnessFlag}
		if err := wire.WriteElement(&buf, order, flags); err != nil {
			return nil, err
		}
	}

	// add number of inputs
	if err := wire.WriteVarInt(&buf, order, uint64(len(tx.Inputs))); err != nil {
		return nil, err
	}

	// add inputs
	for _, input := range tx.Inputs {
		if err := input.Serialize(&buf, order); err != nil {
			return nil, err
		}
	}

	// add number of outputs
	if err := wire.WriteVarInt(&buf, order, uint64(len(tx.Outputs))); err != nil {
		return nil, err
	}

	// add outputs
	for _, output := range tx.Outputs {
		if err := output.Serialize(&buf, order); err != nil {
			return nil, err
		}
	}

	// if segwit tx, add witness data
	if doWitness {
		for _, input := range tx.Inputs {
			if err := input.Witness.Serialize(&buf, order); err != nil {
				return nil, err
			}
		}
	}

	// add locktime
	if err := wire.WriteElement(&buf, order, tx.LockTime); err != nil {
		return nil, err
	}

	// if zcash, add expiry height
	if tx.ExpiryHeight != nil {
		if err := wire.WriteElement(&buf, order, *tx.ExpiryHeight); err != nil {
			return nil, err
		} else if err := wire.WriteElement(&buf, order, make([]byte, 11)); err != nil {
			// @TODO: no idea why this ^ is needed
			return nil, err
		}
	}

	// if extra payload, add data
	if len(extraPayload) > 0 {
		if err := wire.WriteElement(&buf, order, byte(len(extraPayload))); err != nil {
			return nil, err
		} else if err := wire.WriteElement(&buf, order, extraPayload); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (tx *Transaction) CalculateScriptSig(index uint32, script []byte) ([]byte, error) {
	txCopy := tx.ShallowCopy()
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
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, SIGHASH_ALL)
	serialized = append(serialized, buf...)

	return crypto.Sha256d(serialized), nil
}

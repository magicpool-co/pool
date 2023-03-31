package nexatx

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/wire"
)

var order = binary.LittleEndian

type Transaction struct {
	Prefix   string
	Version  uint8
	LockTime uint32
	Inputs   []*input
	Outputs  []*output
}

func NewTransaction(version uint8, lockTime uint32, prefix string) *Transaction {
	tx := &Transaction{
		Prefix:   prefix,
		Version:  version,
		LockTime: lockTime,
	}

	return tx
}

type input struct {
	Version   uint8
	PrevHash  []byte
	PrevIndex uint32
	Script    []byte
	Sequence  uint32
	Value     uint64
}

func (inp *input) Serialize(buf *bytes.Buffer, order binary.ByteOrder, includeInputScript bool) error {
	if err := wire.WriteElement(buf, order, inp.Version); err != nil {
		return err
	} else if err := wire.WriteElement(buf, order, crypto.ReverseBytes(inp.PrevHash)); err != nil {
		return err
	}

	if includeInputScript {
		if err := wire.WriteVarBytes(buf, order, inp.Script); err != nil {
			return err
		}
	}

	if err := wire.WriteElement(buf, order, inp.Sequence); err != nil {
		return err
	} else if err := wire.WriteElement(buf, order, inp.Value); err != nil {
		return err
	}

	return nil
}

func (inp *input) Deserialize(reader *bytes.Reader, order binary.ByteOrder) error {
	if err := wire.ReadElement(reader, order, &inp.Version); err != nil {
		return err
	}

	inp.PrevHash = make([]byte, 32)
	if err := wire.ReadElement(reader, order, &inp.PrevHash); err != nil {
		return err
	}
	inp.PrevHash = crypto.ReverseBytes(inp.PrevHash)

	var err error
	inp.Script, err = wire.ReadVarBytes(reader, order)
	if err != nil {
		return err
	}

	if err := wire.ReadElement(reader, order, &inp.Sequence); err != nil {
		return err
	} else if err := wire.ReadElement(reader, order, &inp.Value); err != nil {
		return err
	}

	return nil
}

type output struct {
	Version uint8
	Script  []byte
	Value   uint64
}

func (out *output) Serialize(buf *bytes.Buffer, order binary.ByteOrder) error {
	if err := wire.WriteElement(buf, order, out.Version); err != nil {
		return err
	} else if err := wire.WriteElement(buf, order, out.Value); err != nil {
		return err
	} else if err := wire.WriteVarBytes(buf, order, out.Script); err != nil {
		return err
	}

	return nil
}

func (out *output) Deserialize(reader *bytes.Reader, order binary.ByteOrder) error {
	if err := wire.ReadElement(reader, order, &out.Version); err != nil {
		return err
	} else if err := wire.ReadElement(reader, order, &out.Value); err != nil {
		return err
	}

	var err error
	out.Script, err = wire.ReadVarBytes(reader, order)
	if err != nil {
		return err
	}

	return nil
}

func (tx *Transaction) AddInput(hash string, index, sequence uint32, script []byte, value uint64) error {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return err
	}

	// have to convert output into special representation (defined
	// here https://gitlab.com/nexa/nexa/-/blob/nexa1.2.0.0/doc/transaction.md#outpoints-coutpoint)
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, index)
	hashBytes = append(crypto.ReverseBytes(hashBytes), data...)
	hashBytes = crypto.ReverseBytes(crypto.Sha256(hashBytes))

	inp := &input{
		Version:   0,
		PrevHash:  hashBytes,
		PrevIndex: index,
		Sequence:  sequence,
		Script:    script,
		Value:     value,
	}

	tx.Inputs = append(tx.Inputs, inp)

	return nil
}

func (tx *Transaction) AddOutput(version uint8, script []byte, value uint64) {
	out := &output{
		Version: version,
		Script:  script,
		Value:   value,
	}

	tx.Outputs = append(tx.Outputs, out)
}

func (tx *Transaction) ShallowCopy() *Transaction {
	// As an additional memory optimization, use contiguous backing arrays
	// for the copied inputs and outputs and point the final slice of
	// pointers into the contiguous arrays.  This avoids a lot of small
	// allocations.
	txCopy := &Transaction{
		Prefix:   tx.Prefix,
		Version:  tx.Version,
		Inputs:   make([]*input, len(tx.Inputs)),
		Outputs:  make([]*output, len(tx.Outputs)),
		LockTime: tx.LockTime,
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

func (tx *Transaction) Serialize(includeInputScript bool) ([]byte, error) {
	var buf bytes.Buffer

	// add version
	if err := wire.WriteElement(&buf, order, tx.Version); err != nil {
		return nil, err
	}

	// add number of inputs
	if err := wire.WriteVarInt(&buf, order, uint64(len(tx.Inputs))); err != nil {
		return nil, err
	}

	// add inputs
	for _, input := range tx.Inputs {
		if err := input.Serialize(&buf, order, includeInputScript); err != nil {
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

	// add locktime
	if err := wire.WriteElement(&buf, order, tx.LockTime); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (tx *Transaction) Deserialize(data []byte) error {
	reader := bytes.NewReader(data)

	// get version
	if err := wire.ReadElement(reader, order, &tx.Version); err != nil {
		return err
	}

	// get number of inputs
	numInputs, err := wire.ReadVarInt(reader, order)
	if err != nil {
		return err
	}

	// get inputs
	tx.Inputs = make([]*input, numInputs)
	for i := range tx.Inputs {
		tx.Inputs[i] = new(input)
		if err := tx.Inputs[i].Deserialize(reader, order); err != nil {
			return err
		}
	}

	// get number of outputs
	numOutputs, err := wire.ReadVarInt(reader, order)
	if err != nil {
		return err
	}

	// get outputs
	tx.Outputs = make([]*output, numOutputs)
	for i := range tx.Outputs {
		tx.Outputs[i] = new(output)
		if err := tx.Outputs[i].Deserialize(reader, order); err != nil {
			return err
		}
	}

	// get locktime
	if err := wire.ReadElement(reader, order, &tx.LockTime); err != nil {
		return err
	}

	return nil
}

func (tx *Transaction) CalculateScriptSig(index uint32, script []byte) ([]byte, error) {
	// inputs
	var hashPrevoutsBuf, hashInputAmountsBuf, hashSequenceBuf bytes.Buffer
	for _, inp := range tx.Inputs {
		// prevouts
		if err := wire.WriteElement(&hashPrevoutsBuf, order, inp.Version); err != nil {
			return nil, err
		} else if err := wire.WriteElement(&hashPrevoutsBuf, order, crypto.ReverseBytes(inp.PrevHash)); err != nil {
			return nil, err
		}

		// inputs amounts
		if err := wire.WriteElement(&hashInputAmountsBuf, order, inp.Value); err != nil {
			return nil, err
		}

		// sequences
		if err := wire.WriteElement(&hashSequenceBuf, order, inp.Sequence); err != nil {
			return nil, err
		}
	}

	hashPrevouts := crypto.Sha256d(hashPrevoutsBuf.Bytes())
	hashInputAmounts := crypto.Sha256d(hashInputAmountsBuf.Bytes())
	hashSequence := crypto.Sha256d(hashSequenceBuf.Bytes())

	// outputs
	var hashOutputsBuf bytes.Buffer
	for _, out := range tx.Outputs {
		if err := out.Serialize(&hashOutputsBuf, order); err != nil {
			return nil, err
		}
	}

	hashOutputs := crypto.Sha256d(hashOutputsBuf.Bytes())

	// sighash
	var sigHashBuf bytes.Buffer
	if err := wire.WriteElement(&sigHashBuf, order, tx.Version); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&sigHashBuf, order, hashPrevouts); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&sigHashBuf, order, hashInputAmounts); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&sigHashBuf, order, hashSequence); err != nil {
		return nil, err
	} else if err := wire.WriteVarBytes(&sigHashBuf, order, script); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&sigHashBuf, order, hashOutputs); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&sigHashBuf, order, tx.LockTime); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&sigHashBuf, order, byte(SIGHASH_ALL)); err != nil {
		return nil, err
	}

	return crypto.Sha256d(sigHashBuf.Bytes()), nil
}

func (tx *Transaction) CalculateTxIdem() ([]byte, error) {
	serialized, err := tx.Serialize(false)
	if err != nil {
		return nil, err
	}

	txIdem := crypto.Sha256d(serialized)

	return txIdem, nil
}

func (tx *Transaction) CalculateTxID() ([]byte, error) {
	txIdem, err := tx.CalculateTxIdem()
	if err != nil {
		return nil, err
	}

	var satisfierScriptBuf bytes.Buffer
	if err := wire.WriteElement(&satisfierScriptBuf, order, uint32(len(tx.Inputs))); err != nil {
		return nil, err
	}

	for _, inp := range tx.Inputs {
		if err := wire.WriteElement(&satisfierScriptBuf, order, inp.Script); err != nil {
			return nil, err
		} else if err := wire.WriteElement(&satisfierScriptBuf, order, byte(0xFF)); err != nil {
			return nil, err
		}
	}

	satisfierScriptHash := crypto.Sha256d(satisfierScriptBuf.Bytes())
	txid := crypto.Sha256d(append(txIdem, satisfierScriptHash...))

	return txid, nil
}

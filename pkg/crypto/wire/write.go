// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"math"
)

// writeElement writes the little endian representation of element to w.
func WriteElement(w io.Writer, order binary.ByteOrder, element interface{}) error {
	// Attempt to write the element based on the concrete type via fast
	// type assertions first.
	switch e := element.(type) {
	case []byte:
		_, err := w.Write(e)
		return err
	case int8:
		return binarySerializer.PutUint8(w, uint8(e))
	case uint8:
		return binarySerializer.PutUint8(w, e)
	case int16:
		return binarySerializer.PutUint16(w, order, uint16(e))
	case uint16:
		return binarySerializer.PutUint16(w, order, e)
	case int32:
		return binarySerializer.PutUint32(w, order, uint32(e))
	case uint32:
		return binarySerializer.PutUint32(w, order, e)
	case int64:
		return binarySerializer.PutUint64(w, order, uint64(e))
	case uint64:
		return binarySerializer.PutUint64(w, order, e)
	case bool:
		if e {
			return binarySerializer.PutUint8(w, 0x01)
		} else {
			return binarySerializer.PutUint8(w, 0x00)
		}
	}

	// Fall back to the slower binary.Write if a fast path was not available
	// above.
	return binary.Write(w, order, element)
}

// writeElements writes multiple items to w.  It is equivalent to multiple
// calls to writeElement.
func WriteElements(w io.Writer, order binary.ByteOrder, elements ...interface{}) error {
	for _, element := range elements {
		err := WriteElement(w, order, element)
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteVarInt serializes val to w using a variable number of bytes depending
// on its value.
func WriteVarInt(w io.Writer, order binary.ByteOrder, val uint64) error {
	if val < 0xfd {
		return binarySerializer.PutUint8(w, uint8(val))
	}

	if val <= math.MaxUint16 {
		err := binarySerializer.PutUint8(w, 0xfd)
		if err != nil {
			return err
		}
		return binarySerializer.PutUint16(w, order, uint16(val))
	}

	if val <= math.MaxUint32 {
		err := binarySerializer.PutUint8(w, 0xfe)
		if err != nil {
			return err
		}
		return binarySerializer.PutUint32(w, order, uint32(val))
	}

	err := binarySerializer.PutUint8(w, 0xff)
	if err != nil {
		return err
	}

	return binarySerializer.PutUint64(w, order, val)
}

// Special helper for Firo coinbase
func WriteSerializedNumber(w io.Writer, order binary.ByteOrder, val uint64) error {
	var data []byte
	if val >= 1 && val <= 16 {
		data = []byte{byte(0x50 + val)}
	} else if val <= 0x7F {
		data = []byte{1, byte(val)}
	} else {
		data = make([]byte, 1)
		n := val
		for n = val; n > 0x7F; n >>= 8 {
			data = append(data, byte(n&0xFF))
		}
		data[0] = byte(len(data))
		data = append(data, byte(n))
	}

	return WriteElement(w, order, data)
}

// WriteVarString serializes str to w as a variable length integer containing
// the length of the string followed by the bytes that represent the string
// itself.
func WriteVarString(w io.Writer, order binary.ByteOrder, str string) error {
	err := WriteVarInt(w, order, uint64(len(str)))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(str))

	return err
}

// WriteHexString decodes a hex string and writes the decoded bytes.
func WriteHexString(w io.Writer, order binary.ByteOrder, str string) error {
	data, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

// WriteHexString decodes a hex string and writes the decoded bytes in reversed byte order.
func WriteReversedHexString(w io.Writer, order binary.ByteOrder, str string) error {
	data, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	// reverse bytes
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

	_, err = w.Write(data)

	return err
}

// WriteVarHexString decodes a hex string and writes the decoded bytes as PrefixedBytes.
func WritePrefixedHexString(w io.Writer, order binary.ByteOrder, str string) error {
	data, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	return WritePrefixedBytes(w, order, data)
}

// WriteVarHexString decodes a hex string and writes the decoded bytes as VarBytes.
func WriteVarHexString(w io.Writer, order binary.ByteOrder, str string) error {
	data, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	return WriteVarBytes(w, order, data)
}

// WritePrefixedBytes serializes a variable length byte array to w as an int
// containing the number of bytes, followed by the bytes themselves.
func WritePrefixedBytes(w io.Writer, order binary.ByteOrder, bytes []byte) error {
	slen := uint64(len(bytes))
	err := binarySerializer.PutUint64(w, order, slen)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)

	return err
}

// WriteVarBytes serializes a variable length byte array to w as a varInt
// containing the number of bytes, followed by the bytes themselves.
func WriteVarBytes(w io.Writer, order binary.ByteOrder, bytes []byte) error {
	slen := uint64(len(bytes))
	err := WriteVarInt(w, order, slen)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)

	return err
}

func WriteVarByteArray(w io.Writer, order binary.ByteOrder, byteArray [][]byte) error {
	slen := uint64(len(byteArray))
	err := WriteVarInt(w, order, slen)
	if err != nil {
		return err
	}

	for _, bytes := range byteArray {
		if _, err := w.Write(bytes); err != nil {
			return err
		}
	}

	return err
}

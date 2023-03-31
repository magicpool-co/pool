// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"encoding/binary"
	"fmt"
	"io"
)

func ReadElement(r io.Reader, order binary.ByteOrder, element interface{}) error {
	// Attempt to read the element based on the concrete type via fast
	// type assertions first.
	switch e := element.(type) {
	case *int8:
		rv, err := binarySerializer.Uint8(r)
		if err != nil {
			return err
		}

		*e = int8(rv)
		return nil
	case *uint8:
		rv, err := binarySerializer.Uint8(r)
		if err != nil {
			return err
		}

		*e = rv
		return nil
	case *int16:
		rv, err := binarySerializer.Uint16(r, order)
		if err != nil {
			return err
		}

		*e = int16(rv)
		return nil
	case *uint16:
		rv, err := binarySerializer.Uint16(r, order)
		if err != nil {
			return err
		}

		*e = rv
		return nil
	case *int32:
		rv, err := binarySerializer.Uint32(r, order)
		if err != nil {
			return err
		}

		*e = int32(rv)
		return nil
	case *uint32:
		rv, err := binarySerializer.Uint32(r, order)
		if err != nil {
			return err
		}

		*e = rv
		return nil
	case *int64:
		rv, err := binarySerializer.Uint64(r, order)
		if err != nil {
			return err
		}

		*e = int64(rv)
		return nil
	case *uint64:
		rv, err := binarySerializer.Uint64(r, order)
		if err != nil {
			return err
		}

		*e = rv
		return nil
	case *bool:
		rv, err := binarySerializer.Uint8(r)
		if err != nil {
			return err
		}

		if rv == 0x00 {
			*e = false
		} else {
			*e = true
		}
	}

	// Fall back to the slower binary.Read if a fast path was not available
	// above.
	return binary.Read(r, order, element)
}

// ReadVarInt reads a variable length integer from r and returns it as a uint64.
func ReadVarInt(r io.Reader, order binary.ByteOrder) (uint64, error) {
	discriminant, err := binarySerializer.Uint8(r)
	if err != nil {
		return 0, err
	}

	var rv uint64
	switch discriminant {
	case 0xff:
		sv, err := binarySerializer.Uint64(r, order)
		if err != nil {
			return 0, err
		}
		rv = sv

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		min := uint64(0x100000000)
		if rv < min {
			return 0, fmt.Errorf("non canonical VarInt")
		}
	case 0xfe:
		sv, err := binarySerializer.Uint32(r, order)
		if err != nil {
			return 0, err
		}
		rv = uint64(sv)

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		min := uint64(0x10000)
		if rv < min {
			return 0, fmt.Errorf("non canonical VarInt")
		}
	case 0xfd:
		sv, err := binarySerializer.Uint16(r, order)
		if err != nil {
			return 0, err
		}
		rv = uint64(sv)

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		min := uint64(0xfd)
		if rv < min {
			return 0, fmt.Errorf("non canonical VarInt")
		}

	default:
		rv = uint64(discriminant)
	}

	return rv, nil
}

// ReadVarBytes reads the initial varInt and reads the following length of bytes.
func ReadVarBytes(r io.Reader, order binary.ByteOrder) ([]byte, error) {
	length, err := ReadVarInt(r, order)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	err = ReadElement(r, order, data)

	return data, err
}

package blkbuilder

import (
	"bytes"
	"encoding/binary"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/merkle"
	"github.com/magicpool-co/pool/pkg/crypto/wire"
)

func SerializeBitcoinBlockHeader(nonce, nTime, version uint32, bits, prevHash string, txHashes [][]byte) ([]byte, []byte, error) {
	var buf bytes.Buffer
	var order = binary.BigEndian

	merkleRoot := merkle.CalculateRoot(txHashes)
	if err := wire.WriteElement(&buf, order, nonce); err != nil {
		return nil, nil, err
	} else if err := wire.WriteHexString(&buf, order, bits); err != nil {
		return nil, nil, err
	} else if err := wire.WriteElement(&buf, order, nTime); err != nil {
		return nil, nil, err
	} else if err := wire.WriteElement(&buf, order, merkleRoot); err != nil {
		return nil, nil, err
	} else if err := wire.WriteHexString(&buf, order, prevHash); err != nil {
		return nil, nil, err
	} else if err := wire.WriteElement(&buf, order, version); err != nil {
		return nil, nil, err
	}

	header := crypto.ReverseBytes(buf.Bytes())
	headerHash := crypto.ReverseBytes(crypto.Sha256d(header))

	return header, headerHash, nil
}

func SerializeBitcoinBlock(header []byte, txHexes [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	var order = binary.LittleEndian

	if err := wire.WriteElement(&buf, order, header); err != nil {
		return nil, err
	} else if err := wire.WriteVarByteArray(&buf, order, txHexes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

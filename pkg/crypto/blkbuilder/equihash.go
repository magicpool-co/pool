package blkbuilder

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/merkle"
	"github.com/magicpool-co/pool/pkg/crypto/wire"
	"github.com/magicpool-co/pool/types"
)

type EquihashBuilder struct {
	partialHeader []byte
	header        []byte
	headerHash    []byte
	txHexes       [][]byte
}

func NewEquihashBuilder(
	version, nTime uint32,
	bits, prevHash, saplingRoot string,
	txHashes, txHexes [][]byte,
) (*EquihashBuilder, error) {
	merkleRoot := merkle.CalculateRoot(txHashes)

	var buf bytes.Buffer
	var order = binary.BigEndian
	if err := wire.WriteHexString(&buf, order, bits); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, nTime); err != nil {
		return nil, err
	} else if err := wire.WriteHexString(&buf, order, saplingRoot); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, merkleRoot); err != nil {
		return nil, err
	} else if err := wire.WriteHexString(&buf, order, prevHash); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, version); err != nil {
		return nil, err
	}

	builder := &EquihashBuilder{
		partialHeader: buf.Bytes(),
		txHexes:       txHexes,
	}

	return builder, nil
}

func (b *EquihashBuilder) SerializeHeader(work *types.StratumWork) ([]byte, []byte, error) {
	if work.Nonce == nil {
		return nil, nil, fmt.Errorf("no nonce")
	} else if work.EquihashSolution == nil {
		return nil, nil, fmt.Errorf("no solution")
	}

	b.header = make([]byte, 32+len(b.partialHeader))
	copy(b.header, work.Nonce.BytesBE())
	copy(b.header[32:], b.partialHeader)
	b.header = crypto.ReverseBytes(b.header)

	b.headerHash = make([]byte, len(b.header)+len(work.EquihashSolution))
	copy(b.headerHash, b.header)
	copy(b.headerHash[len(b.header):], work.EquihashSolution)
	b.headerHash = crypto.ReverseBytes(crypto.Sha256d(b.headerHash))

	return b.header, b.headerHash, nil
}

func (b *EquihashBuilder) SerializeBlock(work *types.StratumWork) ([]byte, error) {
	if work.EquihashSolution == nil {
		return nil, fmt.Errorf("no solution")
	}

	var buf bytes.Buffer
	var order = binary.BigEndian
	if err := wire.WriteElement(&buf, order, b.header); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, work.EquihashSolution); err != nil {
		return nil, err
	} else if err := wire.WriteVarByteArray(&buf, order, b.txHexes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *EquihashBuilder) PartialJob() []interface{} {
	result := []interface{}{
		hex.EncodeToString(crypto.ReverseBytes(b.partialHeader[104:108])), // version
		hex.EncodeToString(crypto.ReverseBytes(b.partialHeader[72:104])),  // prevHash
		hex.EncodeToString(crypto.ReverseBytes(b.partialHeader[40:72])),   // merkleRoot
		hex.EncodeToString(crypto.ReverseBytes(b.partialHeader[8:40])),    // saplingRoot
		hex.EncodeToString(crypto.ReverseBytes(b.partialHeader[4:8])),     // nTime
		hex.EncodeToString(crypto.ReverseBytes(b.partialHeader[0:4])),     // bits
	}

	return result
}

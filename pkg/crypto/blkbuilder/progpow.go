package blkbuilder

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/merkle"
	"github.com/magicpool-co/pool/pkg/crypto/wire"
	"github.com/magicpool-co/pool/types"
)

type ProgPowBuilder struct {
	header     []byte
	headerHash []byte
	txHexes    [][]byte
}

func NewProgPowBuilder(
	version, nTime, height uint32,
	bits, prevHash string,
	txHashes, txHexes [][]byte,
) (*ProgPowBuilder, error) {
	merkleRoot := merkle.CalculateRoot(txHashes)

	var buf bytes.Buffer
	var order = binary.BigEndian
	if err := wire.WriteElement(&buf, order, height); err != nil {
		return nil, err
	} else if err := wire.WriteHexString(&buf, order, bits); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, nTime); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, merkleRoot); err != nil {
		return nil, err
	} else if err := wire.WriteHexString(&buf, order, prevHash); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, version); err != nil {
		return nil, err
	}

	header := crypto.ReverseBytes(buf.Bytes())
	headerHash := crypto.ReverseBytes(crypto.Sha256d(header))
	builder := &ProgPowBuilder{
		header:     header,
		headerHash: headerHash,
		txHexes:    txHexes,
	}

	return builder, nil
}

func (b *ProgPowBuilder) SerializeHeader(work *types.StratumWork) ([]byte, []byte, error) {
	return b.header, b.headerHash, nil
}

func (b *ProgPowBuilder) SerializeBlock(work *types.StratumWork) ([]byte, error) {
	if work.Nonce == nil {
		return nil, fmt.Errorf("no nonce")
	} else if work.MixDigest == nil {
		return nil, fmt.Errorf("no mix digest")
	}

	var buf bytes.Buffer
	var order = binary.BigEndian
	if err := wire.WriteElement(&buf, order, b.header); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, crypto.ReverseBytes(work.Nonce.BytesBE())); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, crypto.ReverseBytes(work.MixDigest.Bytes())); err != nil {
		return nil, err
	} else if err := wire.WriteVarInt(&buf, order, uint64(len(b.txHexes))); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, bytes.Join(b.txHexes, nil)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *ProgPowBuilder) PartialJob() []interface{} {
	return nil
}

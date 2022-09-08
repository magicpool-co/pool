package blkbuilder

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/magicpool-co/pool/pkg/crypto/util"
	"github.com/magicpool-co/pool/types"
)

type EquihashBuilder struct {
	version     []byte
	prevHash    []byte
	merkleRoot  []byte
	saplingRoot []byte
	nTime       []byte
	bits        []byte
	header      []byte
	headerHash  []byte
	txHexes     [][]byte
}

func NewEquihashBuilder(version, nTime uint32, bits, prevHash, saplingRoot string, txHashes, txHexes [][]byte) (*EquihashBuilder, error) {
	nTimeBytes := util.WriteUint32Be(nTime)
	versionBytes := util.WriteUint32Be(version)

	bitsBytes, err := hex.DecodeString(bits)
	if err != nil {
		return nil, err
	}

	prevHashBytes, err := hex.DecodeString(prevHash)
	if err != nil {
		return nil, err
	}

	saplingRootBytes, err := hex.DecodeString(saplingRoot)
	if err != nil {
		return nil, err
	}

	merkleRootBytes := util.CalculateMerkleRoot(txHashes)

	builder := &EquihashBuilder{
		nTime:       nTimeBytes,
		version:     versionBytes,
		bits:        bitsBytes,
		prevHash:    prevHashBytes,
		merkleRoot:  merkleRootBytes,
		saplingRoot: saplingRootBytes,
		txHexes:     txHexes,
	}

	return builder, nil
}

func (b *EquihashBuilder) SerializeHeader(work *types.StratumWork) ([]byte, []byte, error) {
	if work.Nonce == nil {
		return nil, nil, fmt.Errorf("no nonce")
	} else if work.EquihashSolution == nil {
		return nil, nil, fmt.Errorf("no solution")
	}

	header := make([]byte, 140)
	copy(header[0:32], work.Nonce.BytesBE()) // 4
	copy(header[32:36], b.bits)              // 4
	copy(header[36:40], b.nTime)             // 4
	copy(header[40:72], b.saplingRoot)       // 32
	copy(header[72:104], b.merkleRoot)       // 32
	copy(header[104:136], b.prevHash)        // 32
	copy(header[136:140], b.version)         // 4
	b.header = util.ReverseBytes(header)

	b.headerHash = util.ReverseBytes(util.Sha256d(append(b.header, work.EquihashSolution...)))

	return b.header, b.headerHash, nil
}

func (b *EquihashBuilder) SerializeBlock(work *types.StratumWork) ([]byte, error) {
	if work.EquihashSolution == nil {
		return nil, fmt.Errorf("no solution")
	}

	hex := bytes.Join([][]byte{
		b.header,
		work.EquihashSolution,
		util.VarIntToBytes(uint64(len(b.txHexes))),
		bytes.Join(b.txHexes, nil),
	}, nil)

	return hex, nil
}

func (b *EquihashBuilder) PartialJob() []interface{} {
	result := []interface{}{
		fmt.Sprintf("%0x", util.ReverseBytes(b.version)),
		fmt.Sprintf("%0x", util.ReverseBytes(b.prevHash)),
		fmt.Sprintf("%0x", util.ReverseBytes(b.merkleRoot)),
		fmt.Sprintf("%0x", util.ReverseBytes(b.saplingRoot)),
		fmt.Sprintf("%0x", util.ReverseBytes(b.nTime)),
		fmt.Sprintf("%0x", util.ReverseBytes(b.bits)),
	}

	return result
}

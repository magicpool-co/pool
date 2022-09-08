package blkbuilder

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/magicpool-co/pool/pkg/crypto/util"
	"github.com/magicpool-co/pool/types"
)

type ProgPowBuilder struct {
	version    []byte
	prevHash   []byte
	merkleRoot []byte
	nTime      []byte
	bits       []byte
	height     []byte
	header     []byte
	headerHash []byte
	txHexes    [][]byte
}

func NewProgPowBuilder(version, nTime, height uint32, bits, prevHash string, txHashes, txHexes [][]byte) (*ProgPowBuilder, error) {
	heightBytes := util.WriteUint32Be(height)
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

	merkleRootBytes := util.CalculateMerkleRoot(txHashes)

	header := make([]byte, 80)
	copy(header[0:4], heightBytes)       // 4
	copy(header[4:8], bitsBytes)         // 4
	copy(header[8:12], nTimeBytes)       // 4
	copy(header[12:44], merkleRootBytes) // 32
	copy(header[44:76], prevHashBytes)   // 32
	copy(header[76:80], versionBytes)    // 4
	header = util.ReverseBytes(header)

	headerHash := util.ReverseBytes(util.Sha256d(header))

	builder := &ProgPowBuilder{
		version:    versionBytes,
		prevHash:   prevHashBytes,
		merkleRoot: merkleRootBytes,
		nTime:      nTimeBytes,
		bits:       bitsBytes,
		height:     heightBytes,
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

	hex := bytes.Join([][]byte{
		b.header,
		util.ReverseBytes(work.Nonce.BytesBE()),
		util.ReverseBytes(work.MixDigest.Bytes()),
		util.VarIntToBytes(uint64(len(b.txHexes))),
		bytes.Join(b.txHexes, nil),
	}, nil)

	return hex, nil
}

func (b *ProgPowBuilder) PartialJob() []interface{} {
	return nil
}

func DecodeBlock(raw []byte) {
	// first header
	header := util.ReverseBytes(raw[:80])
	headerHash := hex.EncodeToString(util.ReverseBytes(util.Sha256d(raw[:80])))

	height := hex.EncodeToString(header[0:4])
	bits := hex.EncodeToString(header[4:8])
	nTime := hex.EncodeToString(header[8:12])
	merkleRoot := hex.EncodeToString(header[12:44])
	prevHash := hex.EncodeToString(header[44:76])
	version := hex.EncodeToString(header[76:80])

	// second header
	nonce := hex.EncodeToString(util.ReverseBytes(raw[80:88]))
	mixHash := hex.EncodeToString(raw[88:120])

	var txLen uint64
	var newIndex int
	var err error
	txLenVarIntPrefix := util.VarIntLength(raw[120])
	if txLenVarIntPrefix == 1 {
		txLen = uint64(raw[120])
		newIndex = 121
	} else {
		txLen, err = util.BytesToVarInt(txLenVarIntPrefix, raw[121:121+txLenVarIntPrefix])
		if err != nil {
			panic(err)
		}
		newIndex = 121 + txLenVarIntPrefix
	}

	// @TODO: need to decode transactions (including coinbase, which is chain specific)

	fmt.Println(txLen)

	_ = raw[newIndex:]

	if true {
		fmt.Println(nonce, mixHash, headerHash)
	}
	fmt.Println(height, bits, nTime, merkleRoot, prevHash, version)
}

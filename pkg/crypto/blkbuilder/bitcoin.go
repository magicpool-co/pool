package blkbuilder

import (
	"bytes"
	"encoding/hex"

	"github.com/magicpool-co/pool/pkg/crypto/util"
)

func serializeBitcoinHeader(nonce, bits, nTime, merkleRoot, prevHash, version []byte) []byte {
	header := make([]byte, 80)

	copy(header[0:4], nonce)        // 4
	copy(header[4:8], bits)         // 4
	copy(header[8:12], nTime)       // 4
	copy(header[12:44], merkleRoot) // 32
	copy(header[44:76], prevHash)   // 32
	copy(header[76:80], version)    // 4

	return util.ReverseBytes(header)
}

func SerializeBitcoinBlockHeader(nonce, nTime, version uint32, bits, prevHash string, txHashes [][]byte) ([]byte, []byte, error) {
	nonceBytes := util.WriteUint32Be(nonce)
	nTimeBytes := util.WriteUint32Be(nTime)
	versionBytes := util.WriteUint32Be(version)

	bitsBytes, err := hex.DecodeString(bits)
	if err != nil {
		return nil, nil, err
	}

	prevHashBytes, err := hex.DecodeString(prevHash)
	if err != nil {
		return nil, nil, err
	}

	merkleRoot := util.CalculateMerkleRoot(txHashes)
	header := serializeBitcoinHeader(nonceBytes, bitsBytes, nTimeBytes, merkleRoot, prevHashBytes, versionBytes)
	headerHash := util.ReverseBytes(util.Sha256d(header))

	return header, headerHash, nil
}

func SerializeBitcoinBlock(header []byte, txHexes [][]byte) ([]byte, error) {
	hex := bytes.Join([][]byte{
		header,
		util.VarIntToBytes(uint64(len(txHexes))),
		bytes.Join(txHexes, nil),
	}, nil)

	return hex, nil
}

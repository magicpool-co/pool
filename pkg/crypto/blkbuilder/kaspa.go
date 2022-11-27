package blkbuilder

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

func calculateHeaderSize(parents [][]string, blueWork []byte) int {
	var totalNumParents int
	for _, parent := range parents {
		totalNumParents += len(parent)
	}

	var headerSize int
	headerSize += 2                                       // version
	headerSize += 8 + len(parents)*8 + totalNumParents*32 // parents
	headerSize += 32 * 3                                  // hashMerkleRoot, acceptedIDMerkleRoot, utxoCommitment
	headerSize += 8                                       // timestamp
	headerSize += 4                                       // bits
	headerSize += 8 * 3                                   // nonce, daaScore, blueScore
	headerSize += 8 + len(blueWork)                       // blueWork
	headerSize += 32                                      // pruningPoint

	return headerSize
}

func writeUint16(writer []byte, value uint16) int {
	const size = 2
	binary.LittleEndian.PutUint16(writer[:size], value)

	return size
}

func writeUint32(writer []byte, value uint32) int {
	const size = 4
	binary.LittleEndian.PutUint32(writer[:size], value)

	return size
}

func writeUint64(writer []byte, value uint64) int {
	const size = 8
	binary.LittleEndian.PutUint64(writer[:size], value)

	return size
}

func writeHex(writer []byte, value string, length int) (int, error) {
	data, err := hex.DecodeString(value)
	if err != nil {
		return 0, err
	} else if len(data) != length {
		return 0, fmt.Errorf("length mismatch: %d, %d", len(data), length)
	}
	copy(writer[:length], data)

	return length, nil
}

func SerializeKaspaBlockHeader(version uint16, parents [][]string, hashMerkleRoot, acceptedIDMerkleRoot, utxoCommitment string, timestamp int64, bits uint32, nonce, daaScore, blueScore uint64, blueWork, pruningPoint string) ([]byte, error) {
	padding := len(blueWork) + (len(blueWork) % 2)
	for len(blueWork) < padding {
		blueWork = "0" + blueWork
	}

	blueWorkBytes, err := hex.DecodeString(blueWork)
	if err != nil {
		return nil, err
	}

	headerSize := calculateHeaderSize(parents, blueWorkBytes)
	header := make([]byte, headerSize)
	var pos int

	pos += writeUint16(header[pos:], version)
	pos += writeUint64(header[pos:], uint64(len(parents)))
	for _, parent := range parents {
		pos += writeUint64(header[pos:], uint64(len(parent)))
		for _, hash := range parent {
			offset, err := writeHex(header[pos:], hash, 32)
			if err != nil {
				return nil, err
			}
			pos += offset
		}
	}

	offset, err := writeHex(header[pos:], hashMerkleRoot, 32)
	if err != nil {
		return nil, err
	}
	pos += offset

	offset, err = writeHex(header[pos:], acceptedIDMerkleRoot, 32)
	if err != nil {
		return nil, err
	}
	pos += offset

	offset, err = writeHex(header[pos:], utxoCommitment, 32)
	if err != nil {
		return nil, err
	}
	pos += offset

	pos += writeUint64(header[pos:], uint64(timestamp))
	pos += writeUint32(header[pos:], bits)
	pos += writeUint64(header[pos:], nonce)
	pos += writeUint64(header[pos:], daaScore)
	pos += writeUint64(header[pos:], blueScore)

	pos += writeUint64(header[pos:], uint64(len(blueWorkBytes)))
	copy(header[pos:], blueWorkBytes)
	pos += len(blueWorkBytes)

	offset, err = writeHex(header[pos:], pruningPoint, 32)
	if err != nil {
		return nil, err
	}
	pos += offset

	if pos != headerSize {
		return nil, fmt.Errorf("final position and size mismatch: have %d, want %d", pos, headerSize)
	}

	hasher, err := blake2b.New(32, []byte("BlockHash"))
	if err != nil {
		return nil, err
	}
	hasher.Write(header)
	header = hasher.Sum(nil)

	return header, nil
}

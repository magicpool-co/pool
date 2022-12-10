package blkbuilder

import (
	"bytes"
	"encoding/binary"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/wire"
)

func SerializeKaspaBlockHeader(version uint16, parents [][]string, hashMerkleRoot, acceptedIDMerkleRoot, utxoCommitment string, timestamp int64, bits uint32, nonce, daaScore, blueScore uint64, blueWork, pruningPoint string) ([]byte, error) {
	padding := len(blueWork) + (len(blueWork) % 2)
	for len(blueWork) < padding {
		blueWork = "0" + blueWork
	}

	var buf bytes.Buffer
	var order = binary.LittleEndian

	if err := wire.WriteElement(&buf, order, version); err != nil {
		return nil, err
	}

	if err := wire.WriteElement(&buf, order, uint64(len(parents))); err != nil {
		return nil, err
	}

	for _, parent := range parents {
		if err := wire.WriteElement(&buf, order, uint64(len(parent))); err != nil {
			return nil, err
		}

		for _, hash := range parent {
			if err := wire.WriteHexString(&buf, order, hash); err != nil {
				return nil, err
			}
		}
	}

	if err := wire.WriteHexString(&buf, order, hashMerkleRoot); err != nil {
		return nil, err
	} else if err := wire.WriteHexString(&buf, order, acceptedIDMerkleRoot); err != nil {
		return nil, err
	} else if err := wire.WriteHexString(&buf, order, utxoCommitment); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, timestamp); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, bits); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, nonce); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, daaScore); err != nil {
		return nil, err
	} else if err := wire.WriteElement(&buf, order, blueScore); err != nil {
		return nil, err
	} else if err := wire.WritePrefixedHexString(&buf, order, blueWork); err != nil {
		return nil, err
	} else if err := wire.WriteHexString(&buf, order, pruningPoint); err != nil {
		return nil, err
	}

	return crypto.Blake2b256MAC(buf.Bytes(), []byte("BlockHash"))
}

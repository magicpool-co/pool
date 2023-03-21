package nexatx

import (
	"bytes"
	"fmt"

	"github.com/magicpool-co/pool/pkg/crypto/bech32"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
)

const (
	charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

	pubKeyAddrID     = 0x00
	scriptHashAddrID = 0x08
	templateAddrID   = 0x98

	pubKeySize     = 20
	scriptHashSize = 20
	templateSize   = 24

	// tx flags
	SIGHASH_ALL = 0x1
)

func generateScriptSig(sig []byte) []byte {
	return bytes.Join([][]byte{
		btctx.EncodeScriptData(sig),
	}, nil)
}

func AddressToScript(addr string, addrPrefix string) ([]byte, error) {
	prefix, version, decoded, err := bech32.DecodeBCH(charset, addr)
	if err != nil {
		return nil, err
	} else if prefix != addrPrefix {
		return nil, fmt.Errorf("prefix mismtach")
	}

	var script []byte
	switch version {
	case pubKeyAddrID:
		if len(decoded) != pubKeySize {
			return nil, fmt.Errorf("length mismatch")
		}
		script = btctx.CompileP2PKH(decoded)
	case templateAddrID:
		if len(decoded) == 0 || int(decoded[0]) != len(decoded)-1 {
			return nil, fmt.Errorf("length mismatch")
		}
		// @TODO: we should run this through a script engine to
		// double check it is valid script
		script = decoded[1:]
	default:
		return nil, fmt.Errorf("unknown address version")
	}

	return script, nil
}

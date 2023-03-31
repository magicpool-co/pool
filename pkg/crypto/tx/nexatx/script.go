package nexatx

import (
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
	SIGHASH_ALL = 0x0
)

func AddressToScript(addr string, addrPrefix string) (uint8, []byte, error) {
	prefix, version, decoded, err := bech32.DecodeBCH(charset, addr)
	if err != nil {
		return 0, nil, err
	} else if prefix != addrPrefix {
		return 0, nil, fmt.Errorf("prefix mismtach")
	}

	var scriptVersion uint8
	var script []byte
	switch version {
	case pubKeyAddrID:
		if len(decoded) != pubKeySize {
			return 0, nil, fmt.Errorf("length mismatch")
		}
		scriptVersion = 0
		script = btctx.CompileP2PKH(decoded)
	case templateAddrID:
		if len(decoded) == 0 || int(decoded[0]) != len(decoded)-1 {
			return 0, nil, fmt.Errorf("length mismatch")
		}
		// @TODO: we should run this through a script engine to
		// double check it is valid script
		scriptVersion = 1
		script = decoded[1:]
	default:
		return 0, nil, fmt.Errorf("unknown address version")
	}

	return scriptVersion, script, nil
}

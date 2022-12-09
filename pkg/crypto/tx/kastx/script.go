package kastx

import (
	"bytes"
	"fmt"

	"github.com/magicpool-co/pool/pkg/crypto/bech32"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
)

const (
	charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

	pubKeyAddrID      = 0x00
	pubKeyECDSAAddrID = 0x01
	scriptHashAddrID  = 0x08

	pubKeySize      = 32
	pubKeySizeECDSA = 33
)

const (
	// opcodes
	OP_EQUAL           = 0x87
	OP_BLAKE_2B        = 0xaa
	OP_CHECKSIG_ECDSA  = 0xab
	OP_CHECKSIG        = 0xac
	OP_CHECKSIG_VERIFY = 0xad

	// tx flags
	SIGHASH_ALL = 0x1
)

func compileP2PK(serializedPubKey []byte) []byte {
	return bytes.Join([][]byte{
		btctx.EncodeScriptData(serializedPubKey),
		btctx.EncodeOpCode(OP_CHECKSIG),
	}, nil)
}

func compileP2PKE(serializedPubKey []byte) []byte {
	return bytes.Join([][]byte{
		btctx.EncodeScriptData(serializedPubKey),
		btctx.EncodeOpCode(OP_CHECKSIG_ECDSA),
	}, nil)
}

func compileP2SH(scriptHash []byte) []byte {
	return bytes.Join([][]byte{
		btctx.EncodeOpCode(OP_BLAKE_2B),
		btctx.EncodeScriptData(scriptHash),
		btctx.EncodeOpCode(OP_EQUAL),
	}, nil)
}

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
		script = compileP2PK(decoded)
	case pubKeyECDSAAddrID:
		if len(decoded) != pubKeySizeECDSA {
			return nil, fmt.Errorf("length mismatch")
		}
		script = compileP2PKE(decoded)
	case scriptHashAddrID:
		if len(decoded) != pubKeySize {
			return nil, fmt.Errorf("length mismatch")
		}
		script = compileP2SH(decoded)
	default:
		return nil, fmt.Errorf("unknown address version")
	}

	return script, nil
}

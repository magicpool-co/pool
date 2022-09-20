package btctx

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil/bech32"
	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/magicpool-co/pool/pkg/crypto/base58"
	"github.com/magicpool-co/pool/pkg/crypto/util"
)

const (
	// opcodes
	OP_0 = 0x00

	OP_DATA_1      = 0x01
	OP_DATA_20     = 0x14
	OP_DATA_21     = 0x15
	OP_DATA_32     = 0x20
	OP_DATA_33     = 0x21
	OP_PUSHDATA1   = 0x4c
	OP_PUSHDATA2   = 0x4d
	OP_PUSHDATA4   = 0x4e
	OP_DUP         = 0x76
	OP_EQUAL       = 0x87
	OP_EQUALVERIFY = 0x88
	OP_HASH160     = 0xa9
	OP_CHECKSIG    = 0xac

	// tx flags
	TxFlagMarker         = 0x00
	WitnessFlag          = 0x01
	SIGHASH_ALL          = 0x1
	SIGHASH_NONE         = 0x2
	SIGHASH_SINGLE       = 0x3
	SIGHASH_ANYONECANPAY = 0x80
)

func EncodeOpCode(opCode int) []byte {
	b := make([]byte, 1)
	b[0] = byte(opCode)

	return b
}

func compileP2PK(serializedPubKey []byte) []byte {
	return bytes.Join([][]byte{
		encodeScriptData(serializedPubKey),
		EncodeOpCode(OP_CHECKSIG),
	}, nil)
}

func compileP2PKH(pubKeyHash []byte) []byte {
	return bytes.Join([][]byte{
		EncodeOpCode(OP_DUP),
		EncodeOpCode(OP_HASH160),
		encodeScriptData(pubKeyHash),
		EncodeOpCode(OP_EQUALVERIFY),
		EncodeOpCode(OP_CHECKSIG),
	}, nil)
}

func compileP2SH(scriptHash []byte) []byte {
	return bytes.Join([][]byte{
		EncodeOpCode(OP_HASH160),
		encodeScriptData(scriptHash),
		EncodeOpCode(OP_EQUAL),
	}, nil)
}

func compileP2WPKH(pubKeyHash []byte) []byte {
	return bytes.Join([][]byte{
		EncodeOpCode(OP_0),
		encodeScriptData(pubKeyHash),
	}, nil)
}

func compileP2WSH(scriptHash []byte) []byte {
	return bytes.Join([][]byte{
		EncodeOpCode(OP_0),
		encodeScriptData(scriptHash),
	}, nil)
}

func compileCoinbaseScript(blockHeight int32, extraNonce uint64) []byte {
	return bytes.Join([][]byte{
		util.WriteUint64Le(uint64(blockHeight)),
		util.WriteUint64Le(uint64(extraNonce)),
	}, nil)
}

func generateScriptSig(sig []byte, pub []byte) []byte {
	return bytes.Join([][]byte{
		encodeScriptData(sig),
		encodeScriptData(pub),
	}, nil)
}

func encodeScriptData(data []byte) []byte {
	dataLen := len(data)

	if dataLen < OP_PUSHDATA1 {
		data = bytes.Join([][]byte{
			[]byte{byte((OP_DATA_1 - 1) + dataLen)},
			data,
		}, nil)
	} else if dataLen <= 0xff {
		data = bytes.Join([][]byte{
			[]byte{OP_PUSHDATA1},
			[]byte{byte(dataLen)},
			data,
		}, nil)
	} else if dataLen <= 0xffff {
		buf := util.WriteUint16Le(uint16(dataLen))
		data = bytes.Join([][]byte{
			[]byte{OP_PUSHDATA2},
			buf,
			data,
		}, nil)
	} else {
		buf := util.WriteUint32Le(uint32(dataLen))
		data = bytes.Join([][]byte{
			[]byte{OP_PUSHDATA4},
			buf,
			data,
		}, nil)
	}

	return data
}

func AddressToScript(addr string, p2pkhPrefix, p2shPrefix []byte, segwit bool) ([]byte, error) {
	// segwit (P2WPKH or P2WSH)
	oneIndex := strings.LastIndexByte(addr, '1')
	if oneIndex > 1 {
		prefix := addr[:oneIndex+1]
		if strings.ToLower(prefix) == "bc1" {
			if !segwit {
				return nil, fmt.Errorf("segwit not supported")
			}

			witnessVer, witnessProg, err := decodeSegWitAddress(addr)
			if err != nil {
				return nil, err
			}

			// We currently only support P2WPKH and P2WSH, which is
			// witness version 0 and P2TR which is witness version 1.
			if witnessVer != 0 && witnessVer != 1 {
				return nil, fmt.Errorf("unsupported witness version %d", witnessVer)
			}

			switch len(witnessProg) {
			case 20:
				return compileP2WPKH(witnessProg), nil
			case 32:
				if witnessVer == 1 {
					return nil, fmt.Errorf("taproot not supported")
				}

				return compileP2WSH(witnessProg), nil
			default:
				return nil, fmt.Errorf("unsupported witness prog len %d", len(witnessProg))
			}
		}
	}

	// P2PK
	if len(addr) == 130 || len(addr) == 66 {
		serializedPubKey, err := hex.DecodeString(addr)
		if err != nil {
			return nil, err
		}

		pubKey, err := secp256k1.ParsePubKey(serializedPubKey)
		if err != nil {
			return nil, err
		}

		var pubKeyBytes []byte
		switch serializedPubKey[0] {
		case 0x02, 0x03:
			pubKeyBytes = pubKey.SerializeCompressed()
		default:
			pubKeyBytes = pubKey.SerializeUncompressed()
		}

		return compileP2PK(pubKeyBytes), nil
	}

	// verify pubKeyHash is valid ripemd160
	prefix, pubKeyHash, err := base58.CheckDecode(addr)
	if err != nil {
		return nil, err
	}

	if p2pkhPrefix != nil && bytes.Compare(prefix, p2pkhPrefix) == 0 {
		return compileP2PKH(pubKeyHash), nil
	} else if p2shPrefix != nil && bytes.Compare(prefix, p2shPrefix) == 0 {
		return compileP2SH(pubKeyHash), nil
	}

	return nil, fmt.Errorf("unknown address type %s", addr)
}

func decodeSegWitAddress(address string) (byte, []byte, error) {
	// Decode the bech32 encoded address.
	_, data, bech32version, err := bech32.DecodeGeneric(address)
	if err != nil {
		return 0, nil, err
	}

	// The first byte of the decoded address is the witness version, it must
	// exist.
	if len(data) < 1 {
		return 0, nil, fmt.Errorf("no witness version")
	}

	// ...and be <= 16.
	version := data[0]
	if version > 16 {
		return 0, nil, fmt.Errorf("invalid witness version: %v", version)
	}

	// The remaining characters of the address returned are grouped into
	// words of 5 bits. In order to restore the original witness program
	// bytes, we'll need to regroup into 8 bit words.
	regrouped, err := bech32.ConvertBits(data[1:], 5, 8, false)
	if err != nil {
		return 0, nil, err
	}

	// The regrouped data must be between 2 and 40 bytes.
	if len(regrouped) < 2 || len(regrouped) > 40 {
		return 0, nil, fmt.Errorf("invalid data length")
	}

	// For witness version 0, address MUST be exactly 20 or 32 bytes.
	if version == 0 && len(regrouped) != 20 && len(regrouped) != 32 {
		return 0, nil, fmt.Errorf("invalid data length for witness "+
			"version 0: %v", len(regrouped))
	}

	// For witness version 0, the bech32 encoding must be used.
	if version == 0 && bech32version != bech32.Version0 {
		return 0, nil, fmt.Errorf("invalid checksum expected bech32 " +
			"encoding for address with witness version 0")
	}

	// For witness version 1, the bech32m encoding must be used.
	if version == 1 && bech32version != bech32.VersionM {
		return 0, nil, fmt.Errorf("invalid checksum expected bech32m " +
			"encoding for address with witness version 1")
	}

	return version, regrouped, nil
}

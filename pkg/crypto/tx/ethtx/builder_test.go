package ethtx

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func TestNewTx(t *testing.T) {
	tests := []struct {
		rawPriv  string
		address  string
		value    *big.Int
		baseFee  *big.Int
		gasLimit uint64
		nonce    uint64
		chainID  uint64
		txHex    []byte
		txFee    *big.Int
		txid     string
	}{
		{
			rawPriv:  "83d590f5efeacd03a05137273f2be70522cb6bfd85acc41682ef85c65a8e7500",
			address:  "0xae8c89152d34206b5bbaaebee2a50e163466f73d",
			value:    new(big.Int).SetUint64(0x0419308899983463),
			baseFee:  new(big.Int).SetUint64(0x4663cb82),
			gasLimit: 21000,
			nonce:    0x1,
			chainID:  1029,
			txHex: []byte{
				0x02, 0xf8, 0x73, 0x82, 0x04, 0x05, 0x01, 0x84, 0xb2, 0xd0, 0x5e, 0x00, 0x84, 0xf9, 0x34, 0x29,
				0x82, 0x82, 0x52, 0x08, 0x94, 0xae, 0x8c, 0x89, 0x15, 0x2d, 0x34, 0x20, 0x6b, 0x5b, 0xba, 0xae,
				0xbe, 0xe2, 0xa5, 0x0e, 0x16, 0x34, 0x66, 0xf7, 0x3d, 0x88, 0x04, 0x18, 0xe0, 0xae, 0x1a, 0xab,
				0x44, 0x53, 0x80, 0xc0, 0x80, 0xa0, 0x69, 0x8d, 0x3d, 0x8b, 0x27, 0xf5, 0x68, 0x9a, 0x2c, 0x1d,
				0x2d, 0x7f, 0x0b, 0xb5, 0x7a, 0x1d, 0xd5, 0xe3, 0x4a, 0xf9, 0xf9, 0x53, 0xf2, 0x38, 0x10, 0xc8,
				0x76, 0x41, 0xd4, 0x65, 0x83, 0x39, 0x9f, 0x16, 0xa4, 0x0a, 0xce, 0x1a, 0x59, 0xf9, 0xb1, 0x0f,
				0xce, 0x00, 0x2b, 0xf1, 0x48, 0x6d, 0x48, 0x5a, 0x75, 0x4b, 0x95, 0xf9, 0xb9, 0x9c, 0x39, 0xea,
				0x45, 0xc2, 0x14, 0x19, 0x31, 0xc1,
			},
			txFee: new(big.Int).SetUint64(87799850922000),
			txid:  "0xcc481bf7bfa305055f6258136f6e8e06d7466a32322a7828db245dc8df1fef5e",
		},
		{
			rawPriv:  "049b55cb98b79f3178050969c568000efcbfc746e08ac490ffc1369a4d9b6cc0",
			address:  "0xae8c89152d34206b5bbaaebee2a50e163466f73d",
			value:    new(big.Int).SetUint64(0x01419308899983463),
			baseFee:  new(big.Int).SetUint64(0x95134f91),
			gasLimit: 11000,
			nonce:    0x142,
			chainID:  1029,
			txHex: []byte{
				0x02, 0xf8, 0x77, 0x82, 0x04, 0x05, 0x82, 0x01, 0x42, 0x84, 0xb2, 0xd0, 0x5e, 0x00, 0x85, 0x01,
				0x47, 0xe3, 0xad, 0x91, 0x82, 0x2a, 0xf8, 0x94, 0xae, 0x8c, 0x89, 0x15, 0x2d, 0x34, 0x20, 0x6b,
				0x5b, 0xba, 0xae, 0xbe, 0xe2, 0xa5, 0x0e, 0x16, 0x34, 0x66, 0xf7, 0x3d, 0x88, 0x14, 0x18, 0xf9,
				0x7f, 0x9a, 0x8e, 0x45, 0xeb, 0x80, 0xc0, 0x80, 0xa0, 0x25, 0xad, 0x58, 0xd4, 0x97, 0x50, 0x21,
				0xdb, 0x32, 0x3d, 0xe7, 0x14, 0x62, 0x9e, 0x90, 0x88, 0x37, 0x06, 0xed, 0x56, 0xa0, 0x94, 0x05,
				0x35, 0xd4, 0x4f, 0xae, 0x08, 0xcb, 0x03, 0xc3, 0x66, 0xa0, 0x78, 0xad, 0xba, 0x75, 0x8b, 0xae,
				0x4f, 0x40, 0xff, 0xcd, 0x7c, 0xf4, 0x89, 0x58, 0x50, 0xb9, 0x0a, 0xe7, 0xc5, 0xa1, 0xc0, 0xb0,
				0xa7, 0x69, 0x87, 0xe7, 0x27, 0x3b, 0x6a, 0xf1, 0xe5, 0x61,
			},
			txFee: new(big.Int).SetUint64(60511778107000),
			txid:  "0x0d526ac4e7f253bf1c17f6babfea671560221daced64ce172f5d660ff5dc987f",
		},
		{
			rawPriv:  "fd531800d692ddeac5f444b6c71579ef8a55de1660da622deae4f212e21f47fc",
			address:  "0xae8c89152d34206b5bbaaebee2a50e163466f73d",
			value:    new(big.Int).SetUint64(0x09421512415132463),
			baseFee:  new(big.Int).SetUint64(0x5c23a91f),
			gasLimit: 13500,
			nonce:    0x999,
			chainID:  1029,
			txHex: []byte{
				0x02, 0xf8, 0x77, 0x82, 0x04, 0x05, 0x82, 0x09, 0x99, 0x84, 0xb2, 0xd0, 0x5e, 0x00, 0x85, 0x01,
				0x0e, 0xf4, 0x07, 0x1f, 0x82, 0x34, 0xbc, 0x94, 0xae, 0x8c, 0x89, 0x15, 0x2d, 0x34, 0x20, 0x6b,
				0x5b, 0xba, 0xae, 0xbe, 0xe2, 0xa5, 0x0e, 0x16, 0x34, 0x66, 0xf7, 0x3d, 0x88, 0x94, 0x21, 0x19,
				0x53, 0x88, 0x6b, 0x9d, 0x9f, 0x80, 0xc0, 0x01, 0xa0, 0x65, 0x97, 0x2f, 0x18, 0xa1, 0xb0, 0x41,
				0x95, 0x32, 0x58, 0xe5, 0x9e, 0x3a, 0x40, 0x79, 0x7c, 0x18, 0xd3, 0x9e, 0xa2, 0x31, 0x21, 0x2d,
				0xdd, 0x2f, 0x5e, 0x53, 0xdf, 0x0e, 0x6e, 0xf9, 0xc4, 0xa0, 0x6c, 0x0f, 0x58, 0xbe, 0xbe, 0x40,
				0x5c, 0xac, 0xd8, 0x18, 0x4c, 0x30, 0xfa, 0xb6, 0x00, 0x86, 0xc9, 0x94, 0xed, 0xff, 0xd2, 0x68,
				0x21, 0x84, 0xe0, 0x31, 0x7a, 0x59, 0x85, 0x29, 0x56, 0x35,
			},
			txFee: new(big.Int).SetUint64(61368852514500),
			txid:  "0x4cbb99cbc23c85c834c9087c0aa23f3a2c10e229696eedec9747104f25a5662e",
		},
	}

	for i, tt := range tests {
		rawPrivBytes, err := hex.DecodeString(tt.rawPriv)
		if err != nil {
			t.Errorf("failed on %d: decode rawPriv: %v", i, err)
			continue
		}

		privKey := secp256k1.PrivKeyFromBytes(rawPrivBytes)
		txHex, txFee, err := NewTx(privKey, tt.address, nil, tt.value,
			tt.baseFee, tt.gasLimit, tt.nonce, tt.chainID)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
			continue
		} else if txHex != "0x"+hex.EncodeToString(tt.txHex) {
			t.Errorf("failed on %d: tx hex mismatch: have %s, want 0x%x", i, txHex, tt.txHex)
			continue
		} else if txFee.Cmp(tt.txFee) != 0 {
			t.Errorf("failed on %d: tx fee mismatch: have %s, want %s", i, txFee, tt.txFee)
			continue
		}

		txid := CalculateTxID(txHex)
		if txid != tt.txid {
			t.Errorf("failed on %d: txid mismatch: have %s, want %s", i, txid, tt.txid)
		}
	}
}

func TestNewLegacyTx(t *testing.T) {
	tests := []struct {
		rawPriv  string
		address  string
		value    *big.Int
		gasPrice *big.Int
		gasLimit uint64
		nonce    uint64
		chainID  uint64
		txHex    []byte
		txFee    *big.Int
		txid     string
	}{
		{
			rawPriv:  "83d590f5efeacd03a05137273f2be70522cb6bfd85acc41682ef85c65a8e7500",
			address:  "0xae8c89152d34206b5bbaaebee2a50e163466f73d",
			value:    new(big.Int).SetUint64(0x0419308899983463),
			gasPrice: new(big.Int).SetUint64(0x0021000000),
			gasLimit: 21000,
			nonce:    0x1,
			chainID:  1029,
			txHex: []byte{
				0xf8, 0x6d, 0x01, 0x84, 0x21, 0x00, 0x00, 0x00, 0x82, 0x52, 0x08, 0x94, 0xae, 0x8c, 0x89, 0x15,
				0x2d, 0x34, 0x20, 0x6b, 0x5b, 0xba, 0xae, 0xbe, 0xe2, 0xa5, 0x0e, 0x16, 0x34, 0x66, 0xf7, 0x3d,
				0x88, 0x04, 0x19, 0x25, 0xf5, 0x91, 0x98, 0x34, 0x63, 0x80, 0x82, 0x08, 0x2d, 0xa0, 0x80, 0xa5,
				0x50, 0x6b, 0xc0, 0xc3, 0x17, 0x9a, 0x03, 0xd8, 0x04, 0x5a, 0x1f, 0x7d, 0xde, 0x03, 0x18, 0x9c,
				0x37, 0x4a, 0x50, 0x74, 0xb1, 0x90, 0x7d, 0x2e, 0xb6, 0xa1, 0x29, 0x74, 0x31, 0xff, 0xa0, 0x78,
				0x78, 0x92, 0x12, 0xf3, 0x11, 0xd5, 0x0a, 0x12, 0x4f, 0x44, 0xb4, 0xe1, 0x57, 0x23, 0x98, 0x8f,
				0xe3, 0x2a, 0xcc, 0x55, 0x31, 0x4f, 0x2e, 0x76, 0x62, 0x29, 0x04, 0xe7, 0xb4, 0x34, 0xb0,
			},
			txFee: new(big.Int).SetUint64(11626610688000),
			txid:  "0xaba17968e980860c8a406505152e6f082348ff7aba8cd44fe7c8a33acc02ea58",
		},
		{
			rawPriv:  "049b55cb98b79f3178050969c568000efcbfc746e08ac490ffc1369a4d9b6cc0",
			address:  "0xae8c89152d34206b5bbaaebee2a50e163466f73d",
			value:    new(big.Int).SetUint64(0x01419308899983463),
			gasPrice: new(big.Int).SetUint64(0x0011020500),
			gasLimit: 11000,
			nonce:    0x142,
			chainID:  1029,
			txHex: []byte{
				0xf8, 0x6f, 0x82, 0x01, 0x42, 0x84, 0x11, 0x02, 0x05, 0x00, 0x82, 0x2a, 0xf8, 0x94, 0xae, 0x8c,
				0x89, 0x15, 0x2d, 0x34, 0x20, 0x6b, 0x5b, 0xba, 0xae, 0xbe, 0xe2, 0xa5, 0x0e, 0x16, 0x34, 0x66,
				0xf7, 0x3d, 0x88, 0x14, 0x19, 0x2d, 0xad, 0xca, 0xd1, 0x5c, 0x63, 0x80, 0x82, 0x08, 0x2d, 0xa0,
				0xb1, 0xba, 0x1f, 0x86, 0xd3, 0xf1, 0x2b, 0x6d, 0x99, 0xbc, 0x33, 0xf5, 0xef, 0x23, 0x35, 0x5d,
				0xb0, 0x72, 0x5b, 0x73, 0xcd, 0xe1, 0xf0, 0xaa, 0xec, 0xfa, 0xd0, 0x6c, 0xc1, 0x10, 0x4b, 0xdf,
				0xa0, 0x2d, 0xe8, 0x07, 0xc6, 0x31, 0x48, 0xf9, 0x49, 0x89, 0x52, 0x7b, 0x5f, 0x93, 0x7a, 0xa2,
				0x97, 0xee, 0x3e, 0x4c, 0xd1, 0x6c, 0x5b, 0xc6, 0xed, 0x03, 0xeb, 0x2a, 0x20, 0x52, 0xb7, 0x7b,
				0x92,
			},
			txFee: new(big.Int).SetUint64(3138795264000),
			txid:  "0xa2cb0d154acdfd365f5ff954d67b93a3eea1e5d0d3b91c5e30257f019a67a4e3",
		},
		{
			rawPriv:  "fd531800d692ddeac5f444b6c71579ef8a55de1660da622deae4f212e21f47fc",
			address:  "0xae8c89152d34206b5bbaaebee2a50e163466f73d",
			value:    new(big.Int).SetUint64(0x09421512415132463),
			gasPrice: new(big.Int).SetUint64(0x009152231),
			gasLimit: 13500,
			nonce:    0x999,
			chainID:  1029,
			txHex: []byte{
				0xf8, 0x6f, 0x82, 0x09, 0x99, 0x84, 0x09, 0x15, 0x22, 0x31, 0x82, 0x34, 0xbc, 0x94, 0xae, 0x8c,
				0x89, 0x15, 0x2d, 0x34, 0x20, 0x6b, 0x5b, 0xba, 0xae, 0xbe, 0xe2, 0xa5, 0x0e, 0x16, 0x34, 0x66,
				0xf7, 0x3d, 0x88, 0x94, 0x21, 0x4f, 0x45, 0x1e, 0x9c, 0x14, 0x67, 0x80, 0x82, 0x08, 0x2d, 0xa0,
				0x90, 0x11, 0xc8, 0x51, 0x2a, 0x89, 0x28, 0xd1, 0x01, 0x1d, 0xb9, 0x36, 0x32, 0x97, 0xb7, 0x50,
				0x55, 0x5f, 0x23, 0x24, 0xf0, 0xd1, 0x13, 0xcc, 0xdb, 0x23, 0x43, 0xf7, 0x5a, 0xe4, 0xbd, 0x50,
				0xa0, 0x6e, 0x49, 0xb2, 0x7c, 0xea, 0xf3, 0x48, 0x60, 0xd5, 0x64, 0x1b, 0x2d, 0xbe, 0x4e, 0x1f,
				0x43, 0xd2, 0x26, 0x56, 0xbe, 0x1d, 0xf6, 0x62, 0xb6, 0x74, 0xc2, 0xc6, 0x8d, 0xa5, 0xea, 0xc5,
				0x45,
			},
			txFee: new(big.Int).SetUint64(2057129365500),
			txid:  "0xa3f1321ae38c93b4070c731800a112b03a5a5964e84a54b97e77c04a0ca7cf8f",
		},
	}

	for i, tt := range tests {
		rawPrivBytes, err := hex.DecodeString(tt.rawPriv)
		if err != nil {
			t.Errorf("failed on %d: decode rawPriv: %v", i, err)
			continue
		}

		privKey := secp256k1.PrivKeyFromBytes(rawPrivBytes)
		txHex, txFee, err := NewLegacyTx(privKey, tt.address, nil, tt.value,
			tt.gasPrice, tt.gasLimit, tt.nonce, tt.chainID)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
			continue
		} else if txHex != "0x"+hex.EncodeToString(tt.txHex) {
			t.Errorf("failed on %d: tx hex mismatch: have %s, want 0x%x", i, txHex, tt.txHex)
			continue
		} else if txFee.Cmp(tt.txFee) != 0 {
			t.Errorf("failed on %d: tx fee mismatch: have %s, want %s", i, txFee, tt.txFee)
			continue
		}

		txid := CalculateTxID(txHex)
		if txid != tt.txid {
			t.Errorf("failed on %d: txid mismatch: have %s, want %s", i, txid, tt.txid)
		}
	}
}

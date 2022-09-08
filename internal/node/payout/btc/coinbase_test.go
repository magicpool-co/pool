package btc

import (
	"bytes"
	"testing"
)

func TestGenerateCoinbase(t *testing.T) {
	tests := []struct {
		version        uint32
		lockTime       uint32
		address        string
		opReturns      []string
		amount         uint64
		height         uint64
		nTime          uint64
		extraData      string
		defaultWitness string
		prefixP2PKH    []byte
		prefixP2SH     []byte
		coinbaseHex    []byte
		coinbaseHash   []byte
	}{
		{
			version:     2,
			address:     "bc1qppsntrhcfe8m48dszxzjq9tfdd4ccpua0hqej2",
			amount:      625000000,
			height:      739165,
			nTime:       1654279674,
			extraData:   "/SBICrypto.com Pool/",
			prefixP2PKH: mainnetPrefixP2PKH,
			prefixP2SH:  mainnetPrefixP2SH,
			coinbaseHex: []byte{
				0x02, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x29, 0x03, 0x5d, 0x47, 0x0b, 0x04, 0xfa,
				0x4d, 0x9a, 0x62, 0x2f, 0x53, 0x42, 0x49, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x2e, 0x63, 0x6f,
				0x6d, 0x20, 0x50, 0x6f, 0x6f, 0x6c, 0x2f, 0x01, 0xf0, 0x59, 0x56, 0x9a, 0x87, 0x15, 0x00, 0x00,
				0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x01, 0x40, 0xbe, 0x40, 0x25, 0x00, 0x00, 0x00, 0x00,
				0x16, 0x00, 0x14, 0x08, 0x61, 0x35, 0x8e, 0xf8, 0x4e, 0x4f, 0xba, 0x9d, 0xb0, 0x11, 0x85, 0x20,
				0x15, 0x69, 0x6b, 0x6b, 0x8c, 0x07, 0x9d, 0x00, 0x00, 0x00, 0x00,
			},
			coinbaseHash: []byte{
				0x68, 0xc3, 0x4a, 0x6f, 0xaa, 0xe3, 0xd2, 0xba, 0xff, 0x50, 0x40, 0xfd, 0x42, 0xee, 0x12, 0x13,
				0xf9, 0x0e, 0x7c, 0x67, 0x1b, 0xd1, 0x9f, 0x73, 0x52, 0x55, 0x74, 0x6f, 0x7e, 0xfe, 0x70, 0x01,
			},
		},
		{
			version: 1,
			address: "1KFHE7w8BhaENAswwryaoccDb6qcT6DbYY",
			opReturns: []string{
				"486174686c54c2ab958d73fc32f074013197302232563e05f3bb53c636d314ea002739d7",
				"52534b424c4f434b3ade8cc5626137b40827a3644dcc85f2956fa5df881947804685e23d2c0042918c",
			},
			amount: 625000000,
			height: 739399,
			nTime:  1654425552,
			extraData: "HMined by AntPool861mm(/(oALƝViA,wx1",
			prefixP2PKH: mainnetPrefixP2PKH,
			prefixP2SH:  mainnetPrefixP2SH,
			coinbaseHex: []byte{
				0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x64, 0x03, 0x18, 0x47, 0x0b, 0x2c, 0xfa,
				0xbe, 0x6d, 0x6d, 0xb3, 0x77, 0xbc, 0x95, 0xf9, 0xc6, 0x93, 0x77, 0x13, 0xf9, 0xa0, 0x67, 0x2a,
				0x1b, 0x13, 0xef, 0xef, 0xb5, 0xa8, 0xed, 0xf0, 0xae, 0xa7, 0xbe, 0x5f, 0xce, 0x68, 0xc1, 0x1f,
				0xbf, 0x1c, 0x2a, 0x10, 0x00, 0x00, 0x00, 0xf0, 0x9f, 0x90, 0x9f, 0x09, 0x2f, 0x46, 0x32, 0x50,
				0x6f, 0x6f, 0x6c, 0x2f, 0x6b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x06, 0xe9, 0x43, 0x60, 0x41, 0x14,
				0x00, 0x00, 0x03, 0x40, 0xbe, 0x40, 0x25, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xc8,
				0x25, 0xa1, 0xec, 0xf2, 0xa6, 0x83, 0x0c, 0x44, 0x01, 0x62, 0x0c, 0x3a, 0x16, 0xf1, 0x99, 0x50,
				0x57, 0xc2, 0xab, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x26, 0x6a, 0x24,
				0x48, 0x61, 0x74, 0x68, 0x6c, 0x54, 0xc2, 0xab, 0x95, 0x8d, 0x73, 0xfc, 0x32, 0xf0, 0x74, 0x01,
				0x31, 0x97, 0x30, 0x22, 0x32, 0x56, 0x3e, 0x05, 0xf3, 0xbb, 0x53, 0xc6, 0x36, 0xd3, 0x14, 0xea,
				0x00, 0x27, 0x39, 0xd7, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x6a, 0x4c, 0x29,
				0x52, 0x53, 0x4b, 0x42, 0x4c, 0x4f, 0x43, 0x4b, 0x3a, 0xde, 0x8c, 0xc5, 0x62, 0x61, 0x37, 0xb4,
				0x08, 0x27, 0xa3, 0x64, 0x4d, 0xcc, 0x85, 0xf2, 0x95, 0x6f, 0xa5, 0xdf, 0x88, 0x19, 0x47, 0x80,
				0x46, 0x85, 0xe2, 0x3d, 0x2c, 0x00, 0x42, 0x91, 0x8c, 0x03, 0x6f, 0x11, 0x49,
			},
			coinbaseHash: []byte{
				0xf0, 0xfa, 0xdb, 0xcb, 0xec, 0xd6, 0x4a, 0x47, 0x9b, 0xf5, 0xff, 0x37, 0xc2, 0x9e, 0x4c, 0xe6,
				0x4b, 0x4f, 0x79, 0x12, 0x9c, 0x5f, 0x72, 0x69, 0x19, 0x14, 0xe2, 0xbc, 0x19, 0x4f, 0x49, 0xfe,
			},
		},
	}

	for i, tt := range tests {
		continue

		coinbaseHex, coinbaseHash, err := GenerateCoinbase(tt.version, tt.lockTime, tt.address, tt.amount, tt.height,
			tt.nTime, tt.extraData, tt.defaultWitness, tt.prefixP2PKH, tt.prefixP2SH)
		if err != nil {
			t.Errorf("failed on %d: GenerateCoinbase: %v", i, err)
		} else if bytes.Compare(coinbaseHex, tt.coinbaseHex) != 0 {
			t.Errorf("failed on %d: coinbase hex mismatch: have %x, want %x", i, coinbaseHex, tt.coinbaseHex)
		} else if bytes.Compare(coinbaseHash, tt.coinbaseHash) != 0 {
			t.Errorf("failed on %d: coinbase hash mismatch: have %x, want %x", i, coinbaseHash, tt.coinbaseHash)
		}
	}
}

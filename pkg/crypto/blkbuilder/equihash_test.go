package blkbuilder

import (
	"bytes"
	"testing"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

func TestSerializeEquihash(t *testing.T) {
	tests := []struct {
		height      uint32 // unused
		nTime       uint32
		version     uint32
		bits        string
		prevHash    string
		saplingRoot string
		txHashes    [][]byte
		txHexes     [][]byte
		work        *types.StratumWork
		header      []byte
		headerHash  []byte
		block       []byte
	}{
		{
			height:      1134187,
			nTime:       1654165695,
			version:     4,
			bits:        "1d1b9fa2",
			prevHash:    "0000000f443abd39a22948760ac1c0465655d4d02d73153deb1d39a496e38da5",
			saplingRoot: "22758d2483e36192e2ca6beef539801f3d6218eb6e24125da16d9cdc7a8e98d9",
			txHashes: [][]byte{
				[]byte{
					0x37, 0xe7, 0x93, 0x4e, 0xa9, 0x8f, 0x8e, 0xe7, 0x82, 0x92, 0xd2, 0xb9, 0x19, 0x93, 0xc8, 0x60,
					0xc4, 0x78, 0x65, 0x37, 0x17, 0xd4, 0x6c, 0x01, 0xbe, 0x8e, 0x60, 0xfe, 0xcb, 0x2a, 0xcf, 0x29,
				},
				[]byte{
					0x06, 0xa1, 0x30, 0x7e, 0xda, 0xd0, 0xd4, 0xc4, 0xf0, 0x76, 0x52, 0xba, 0xf5, 0xc1, 0x06, 0x5f,
					0x9b, 0x59, 0x8d, 0x41, 0x22, 0x54, 0x6f, 0x73, 0x3e, 0x0c, 0x2e, 0x0a, 0x21, 0xab, 0xf8, 0xfa,
				},
				[]byte{
					0xdb, 0x91, 0x1a, 0xce, 0x4b, 0x24, 0x7d, 0x1e, 0x77, 0xaf, 0x37, 0x7a, 0xe3, 0x66, 0x7f, 0xe3,
					0x43, 0xdc, 0xb3, 0x86, 0x2f, 0x6f, 0x27, 0xfc, 0xc7, 0x6b, 0x13, 0xb7, 0x42, 0x94, 0x42, 0x6a,
				},
				[]byte{
					0x3c, 0x3b, 0x5d, 0x60, 0xe5, 0x9f, 0x8e, 0x48, 0x74, 0xb4, 0xd7, 0x26, 0x47, 0x76, 0xee, 0x4e,
					0x9e, 0x58, 0xf8, 0x52, 0xa4, 0x79, 0x7c, 0x33, 0xe4, 0x7b, 0xcf, 0xf3, 0x04, 0x57, 0x50, 0x50,
				},
			},
			txHexes: [][]byte{
				[]byte{
					0x04, 0x00, 0x00, 0x80, 0x85, 0x20, 0x2f, 0x89, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x20, 0x03, 0x6b,
					0x4e, 0x11, 0x00, 0x32, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x20, 0x68, 0x74, 0x74, 0x70, 0x73,
					0x3a, 0x2f, 0x2f, 0x32, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0xff, 0xff,
					0xff, 0xff, 0x04, 0x80, 0x75, 0x84, 0xdf, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0x04,
					0xe2, 0x69, 0x9c, 0xec, 0x5f, 0x44, 0x28, 0x05, 0x40, 0xfb, 0x75, 0x2c, 0x76, 0x60, 0xaa, 0x3b,
					0xa8, 0x57, 0xcc, 0x88, 0xac, 0xa0, 0x11, 0x87, 0x21, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9,
					0x14, 0xc8, 0x2d, 0x43, 0x3d, 0x5b, 0x39, 0x0d, 0x31, 0xf9, 0x15, 0xe8, 0x03, 0xf0, 0x78, 0x90,
					0x81, 0x90, 0x6a, 0x01, 0x8d, 0x88, 0xac, 0x60, 0x1d, 0xe1, 0x37, 0x00, 0x00, 0x00, 0x00, 0x19,
					0x76, 0xa9, 0x14, 0x68, 0xc0, 0x11, 0x16, 0xec, 0x9b, 0x98, 0x11, 0x26, 0x7d, 0xad, 0xf5, 0xc3,
					0x77, 0x35, 0xff, 0xe5, 0x7a, 0xb9, 0x26, 0x88, 0xac, 0x80, 0x46, 0x1c, 0x86, 0x00, 0x00, 0x00,
					0x00, 0x19, 0x76, 0xa9, 0x14, 0xe1, 0x2b, 0xd3, 0x11, 0x53, 0x8d, 0x45, 0xd2, 0xaa, 0x34, 0xb0,
					0xa6, 0x32, 0xfb, 0x8c, 0x53, 0xaa, 0xaa, 0x2b, 0x19, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				[]byte{
					0x05, 0x00, 0x00, 0x00, 0x04, 0xc1, 0x24, 0x42, 0x86, 0x3b, 0x15, 0xb5, 0x4e, 0x59, 0x0d, 0x1d,
					0xe7, 0xb9, 0xbc, 0x91, 0xcb, 0xb1, 0x96, 0xdf, 0x11, 0xeb, 0xb1, 0x3d, 0xd2, 0x49, 0xb6, 0x49,
					0x2f, 0x14, 0x0a, 0x24, 0x63, 0x00, 0x00, 0x00, 0x00, 0xbf, 0x90, 0x98, 0x62, 0x03, 0x9f, 0x8d,
					0x98, 0x62, 0x01, 0x0e, 0x36, 0x35, 0x2e, 0x31, 0x30, 0x38, 0x2e, 0x31, 0x39, 0x38, 0x2e, 0x31,
					0x31, 0x39, 0x41, 0x1c, 0x3c, 0x75, 0xbe, 0x47, 0xc4, 0x32, 0xf0, 0xad, 0x5d, 0xd9, 0x50, 0x40,
					0x1f, 0x5d, 0x61, 0x0a, 0x92, 0xc4, 0x09, 0xe5, 0x71, 0xf8, 0xb2, 0x41, 0xe2, 0xd5, 0x40, 0x3a,
					0xdf, 0xdc, 0xb3, 0xa9, 0x5e, 0x91, 0xaa, 0xdc, 0x92, 0x1d, 0x92, 0x4d, 0xeb, 0x5a, 0xde, 0x28,
					0x5c, 0xdf, 0x8d, 0x3a, 0xaa, 0xd5, 0xd6, 0xa6, 0xdc, 0xcd, 0x96, 0x29, 0xe0, 0x74, 0x36, 0xcf,
					0x74, 0x78, 0x79, 0x31, 0x41, 0x1b, 0xbc, 0xbb, 0xaa, 0x30, 0xa6, 0xac, 0x10, 0x73, 0xb7, 0xc8,
					0xdf, 0xff, 0xb8, 0x80, 0xe5, 0x02, 0xec, 0x0a, 0x03, 0x73, 0x27, 0xf2, 0x5c, 0xb3, 0x84, 0x63,
					0x45, 0x7b, 0xb8, 0x11, 0x38, 0xe0, 0x3c, 0x19, 0xc6, 0x42, 0x39, 0xd3, 0x59, 0x00, 0x81, 0xd0,
					0x1c, 0xf3, 0x57, 0x4e, 0x66, 0x3e, 0xa4, 0x70, 0xe5, 0xcc, 0x1a, 0xa0, 0x56, 0x33, 0x62, 0x80,
					0x3c, 0xca, 0x73, 0x23, 0x02, 0xa9,
				},
				[]byte{
					0x05, 0x00, 0x00, 0x00, 0x04, 0xee, 0x12, 0xe3, 0x87, 0xc5, 0x38, 0xfe, 0x48, 0xe7, 0xa8, 0xb6,
					0x01, 0xd7, 0x93, 0x77, 0xa4, 0x2a, 0xf1, 0xe5, 0x75, 0x9d, 0x39, 0xef, 0x12, 0x8a, 0x95, 0xc7,
					0x53, 0x46, 0xe1, 0x54, 0x53, 0x00, 0x00, 0x00, 0x00, 0xbf, 0x90, 0x98, 0x62, 0x01, 0xb6, 0x90,
					0x98, 0x62, 0x01, 0x0d, 0x39, 0x34, 0x2e, 0x32, 0x33, 0x2e, 0x32, 0x35, 0x31, 0x2e, 0x31, 0x36,
					0x39, 0x41, 0x1b, 0xfb, 0xbf, 0xcc, 0xb9, 0x77, 0xa7, 0x95, 0xc8, 0xdb, 0x4a, 0x57, 0x10, 0xce,
					0xd5, 0x1a, 0x12, 0x51, 0xa8, 0x97, 0x12, 0xbd, 0xac, 0xf3, 0xce, 0x45, 0xea, 0xa0, 0xc7, 0x71,
					0x3a, 0xee, 0x8d, 0x64, 0xba, 0xf5, 0x17, 0x36, 0x8e, 0x34, 0x3b, 0xbb, 0xb4, 0x0d, 0x63, 0x78,
					0xe4, 0x8d, 0x99, 0xda, 0xee, 0xf5, 0xfe, 0x5f, 0xea, 0xff, 0x91, 0xb8, 0xbd, 0x0f, 0x86, 0xb2,
					0x88, 0x3f, 0x91, 0x41, 0x1c, 0x2e, 0x19, 0xdf, 0x0f, 0xe6, 0x60, 0xf4, 0x4c, 0x93, 0xde, 0x18,
					0x7d, 0x0e, 0x42, 0x4f, 0x6b, 0x4c, 0x6b, 0xa1, 0x46, 0x5a, 0x82, 0x6a, 0xc5, 0xbb, 0x40, 0x76,
					0xe5, 0x14, 0x62, 0xe7, 0x98, 0x24, 0x61, 0xd1, 0x6a, 0xee, 0xd4, 0x00, 0x0b, 0x1e, 0xb3, 0x77,
					0x5b, 0x48, 0xfa, 0x14, 0x22, 0xd3, 0xf9, 0x14, 0xb9, 0x42, 0xd4, 0xc2, 0x97, 0x93, 0x62, 0x75,
					0x0c, 0x6d, 0x57, 0x8f, 0x15,
				},
				[]byte{
					0x05, 0x00, 0x00, 0x00, 0x04, 0xf2, 0xbe, 0x7c, 0xbd, 0x7a, 0xd2, 0x42, 0xbd, 0x62, 0x81, 0xd4,
					0x95, 0xa9, 0x0b, 0xa9, 0xd7, 0xce, 0xf4, 0x23, 0x6d, 0xf0, 0xe6, 0xcb, 0xbd, 0x0c, 0x7f, 0x4b,
					0x9c, 0x94, 0x95, 0x26, 0x93, 0x00, 0x00, 0x00, 0x00, 0xbf, 0x90, 0x98, 0x62, 0x02, 0x56, 0x8c,
					0x98, 0x62, 0x01, 0x0d, 0x31, 0x39, 0x35, 0x2e, 0x33, 0x2e, 0x32, 0x32, 0x33, 0x2e, 0x31, 0x36,
					0x32, 0x41, 0x1b, 0x4b, 0x00, 0xc4, 0x7b, 0xce, 0xd3, 0x17, 0xd4, 0xea, 0x45, 0x68, 0x9d, 0x94,
					0x7e, 0x94, 0xce, 0x24, 0x38, 0x67, 0x58, 0x81, 0x9a, 0x82, 0xd8, 0x42, 0x67, 0xd0, 0x32, 0x63,
					0x1b, 0x9c, 0xe4, 0x35, 0x95, 0xad, 0x27, 0x56, 0xfb, 0xe3, 0x3d, 0x50, 0xbc, 0xc3, 0x14, 0xae,
					0x8d, 0x2b, 0x4a, 0x9e, 0x0b, 0x5e, 0xdc, 0xdf, 0xc9, 0x2c, 0xf2, 0x5e, 0xed, 0x0a, 0xdd, 0x95,
					0xaa, 0x4a, 0x18, 0x41, 0x1b, 0x24, 0xcd, 0x3e, 0x26, 0xb2, 0x9a, 0x87, 0xd7, 0xf2, 0x41, 0xec,
					0x52, 0x71, 0x51, 0x97, 0x6b, 0x51, 0x30, 0xc8, 0x3a, 0x3c, 0x84, 0x92, 0xba, 0xdb, 0x17, 0x85,
					0xaa, 0xcc, 0x6d, 0xf9, 0x95, 0x42, 0x31, 0xb6, 0xe8, 0xd0, 0xe6, 0xd4, 0x2c, 0xe9, 0xe3, 0xbb,
					0x34, 0xe0, 0x6b, 0xc5, 0xcb, 0x55, 0x8b, 0x70, 0x56, 0xc5, 0xb4, 0x83, 0x73, 0x66, 0x12, 0xa5,
					0x45, 0xde, 0xe9, 0x42, 0x16,
				},
			},
			work: &types.StratumWork{
				Nonce: new(types.Number).SetFromBytes(common.MustParseHex("07d7fd03000000000000000000000000000000000000000000000000c1c83500")),
				EquihashSolution: []byte{
					0x34, 0x0d, 0x09, 0x5c, 0x5c, 0xcc, 0x34, 0xed, 0x3d, 0x87, 0xc3, 0xd7, 0xff, 0x8b, 0x60, 0x99,
					0xa9, 0x6b, 0xb8, 0xba, 0xc9, 0x72, 0x8c, 0x1e, 0x84, 0x5f, 0xaa, 0x14, 0x87, 0xfa, 0xef, 0xae,
					0x16, 0x06, 0xa1, 0x03, 0x4d, 0xcf, 0x38, 0x1f, 0x2c, 0xfb, 0x09, 0x56, 0x13, 0x98, 0x8c, 0x49,
					0xa7, 0x3b, 0x3f, 0xcd, 0xaa,
				},
			},
			header: []byte{
				0x04, 0x00, 0x00, 0x00, 0xa5, 0x8d, 0xe3, 0x96, 0xa4, 0x39, 0x1d, 0xeb, 0x3d, 0x15, 0x73, 0x2d,
				0xd0, 0xd4, 0x55, 0x56, 0x46, 0xc0, 0xc1, 0x0a, 0x76, 0x48, 0x29, 0xa2, 0x39, 0xbd, 0x3a, 0x44,
				0x0f, 0x00, 0x00, 0x00, 0xe2, 0x7a, 0x5a, 0x53, 0xcc, 0x84, 0x29, 0x0c, 0x50, 0xa0, 0xfb, 0x5f,
				0xf1, 0x9a, 0xd7, 0x7a, 0x7f, 0x14, 0xe7, 0xe0, 0xcf, 0x63, 0x91, 0x7e, 0x6a, 0x3b, 0x33, 0x12,
				0xb6, 0xcb, 0x53, 0x95, 0xd9, 0x98, 0x8e, 0x7a, 0xdc, 0x9c, 0x6d, 0xa1, 0x5d, 0x12, 0x24, 0x6e,
				0xeb, 0x18, 0x62, 0x3d, 0x1f, 0x80, 0x39, 0xf5, 0xee, 0x6b, 0xca, 0xe2, 0x92, 0x61, 0xe3, 0x83,
				0x24, 0x8d, 0x75, 0x22, 0xbf, 0x90, 0x98, 0x62, 0xa2, 0x9f, 0x1b, 0x1d, 0x07, 0xd7, 0xfd, 0x03,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc1, 0xc8, 0x35, 0x00,
			},
			headerHash: []byte{
				0x00, 0x00, 0x00, 0x10, 0x5e, 0x37, 0xa8, 0xd5, 0x20, 0xba, 0xc1, 0x4d, 0xb7, 0x69, 0x04, 0xd3,
				0xb4, 0xd4, 0x77, 0x5e, 0x93, 0xe3, 0x0e, 0x8a, 0xf9, 0x3f, 0xc6, 0x3d, 0x05, 0x74, 0x20, 0x08,
			},
			block: []byte{
				0x04, 0x00, 0x00, 0x00, 0xa5, 0x8d, 0xe3, 0x96, 0xa4, 0x39, 0x1d, 0xeb, 0x3d, 0x15, 0x73, 0x2d,
				0xd0, 0xd4, 0x55, 0x56, 0x46, 0xc0, 0xc1, 0x0a, 0x76, 0x48, 0x29, 0xa2, 0x39, 0xbd, 0x3a, 0x44,
				0x0f, 0x00, 0x00, 0x00, 0xe2, 0x7a, 0x5a, 0x53, 0xcc, 0x84, 0x29, 0x0c, 0x50, 0xa0, 0xfb, 0x5f,
				0xf1, 0x9a, 0xd7, 0x7a, 0x7f, 0x14, 0xe7, 0xe0, 0xcf, 0x63, 0x91, 0x7e, 0x6a, 0x3b, 0x33, 0x12,
				0xb6, 0xcb, 0x53, 0x95, 0xd9, 0x98, 0x8e, 0x7a, 0xdc, 0x9c, 0x6d, 0xa1, 0x5d, 0x12, 0x24, 0x6e,
				0xeb, 0x18, 0x62, 0x3d, 0x1f, 0x80, 0x39, 0xf5, 0xee, 0x6b, 0xca, 0xe2, 0x92, 0x61, 0xe3, 0x83,
				0x24, 0x8d, 0x75, 0x22, 0xbf, 0x90, 0x98, 0x62, 0xa2, 0x9f, 0x1b, 0x1d, 0x07, 0xd7, 0xfd, 0x03,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc1, 0xc8, 0x35, 0x00, 0x34, 0x0d, 0x09, 0x5c,
				0x5c, 0xcc, 0x34, 0xed, 0x3d, 0x87, 0xc3, 0xd7, 0xff, 0x8b, 0x60, 0x99, 0xa9, 0x6b, 0xb8, 0xba,
				0xc9, 0x72, 0x8c, 0x1e, 0x84, 0x5f, 0xaa, 0x14, 0x87, 0xfa, 0xef, 0xae, 0x16, 0x06, 0xa1, 0x03,
				0x4d, 0xcf, 0x38, 0x1f, 0x2c, 0xfb, 0x09, 0x56, 0x13, 0x98, 0x8c, 0x49, 0xa7, 0x3b, 0x3f, 0xcd,
				0xaa, 0x04, 0x04, 0x00, 0x00, 0x80, 0x85, 0x20, 0x2f, 0x89, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x20,
				0x03, 0x6b, 0x4e, 0x11, 0x00, 0x32, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x20, 0x68, 0x74, 0x74,
				0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x32, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x2e, 0x63, 0x6f, 0x6d,
				0xff, 0xff, 0xff, 0xff, 0x04, 0x80, 0x75, 0x84, 0xdf, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9,
				0x14, 0x04, 0xe2, 0x69, 0x9c, 0xec, 0x5f, 0x44, 0x28, 0x05, 0x40, 0xfb, 0x75, 0x2c, 0x76, 0x60,
				0xaa, 0x3b, 0xa8, 0x57, 0xcc, 0x88, 0xac, 0xa0, 0x11, 0x87, 0x21, 0x00, 0x00, 0x00, 0x00, 0x19,
				0x76, 0xa9, 0x14, 0xc8, 0x2d, 0x43, 0x3d, 0x5b, 0x39, 0x0d, 0x31, 0xf9, 0x15, 0xe8, 0x03, 0xf0,
				0x78, 0x90, 0x81, 0x90, 0x6a, 0x01, 0x8d, 0x88, 0xac, 0x60, 0x1d, 0xe1, 0x37, 0x00, 0x00, 0x00,
				0x00, 0x19, 0x76, 0xa9, 0x14, 0x68, 0xc0, 0x11, 0x16, 0xec, 0x9b, 0x98, 0x11, 0x26, 0x7d, 0xad,
				0xf5, 0xc3, 0x77, 0x35, 0xff, 0xe5, 0x7a, 0xb9, 0x26, 0x88, 0xac, 0x80, 0x46, 0x1c, 0x86, 0x00,
				0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xe1, 0x2b, 0xd3, 0x11, 0x53, 0x8d, 0x45, 0xd2, 0xaa,
				0x34, 0xb0, 0xa6, 0x32, 0xfb, 0x8c, 0x53, 0xaa, 0xaa, 0x2b, 0x19, 0x88, 0xac, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x00, 0x00, 0x04, 0xc1, 0x24, 0x42, 0x86, 0x3b, 0x15, 0xb5, 0x4e, 0x59, 0x0d, 0x1d,
				0xe7, 0xb9, 0xbc, 0x91, 0xcb, 0xb1, 0x96, 0xdf, 0x11, 0xeb, 0xb1, 0x3d, 0xd2, 0x49, 0xb6, 0x49,
				0x2f, 0x14, 0x0a, 0x24, 0x63, 0x00, 0x00, 0x00, 0x00, 0xbf, 0x90, 0x98, 0x62, 0x03, 0x9f, 0x8d,
				0x98, 0x62, 0x01, 0x0e, 0x36, 0x35, 0x2e, 0x31, 0x30, 0x38, 0x2e, 0x31, 0x39, 0x38, 0x2e, 0x31,
				0x31, 0x39, 0x41, 0x1c, 0x3c, 0x75, 0xbe, 0x47, 0xc4, 0x32, 0xf0, 0xad, 0x5d, 0xd9, 0x50, 0x40,
				0x1f, 0x5d, 0x61, 0x0a, 0x92, 0xc4, 0x09, 0xe5, 0x71, 0xf8, 0xb2, 0x41, 0xe2, 0xd5, 0x40, 0x3a,
				0xdf, 0xdc, 0xb3, 0xa9, 0x5e, 0x91, 0xaa, 0xdc, 0x92, 0x1d, 0x92, 0x4d, 0xeb, 0x5a, 0xde, 0x28,
				0x5c, 0xdf, 0x8d, 0x3a, 0xaa, 0xd5, 0xd6, 0xa6, 0xdc, 0xcd, 0x96, 0x29, 0xe0, 0x74, 0x36, 0xcf,
				0x74, 0x78, 0x79, 0x31, 0x41, 0x1b, 0xbc, 0xbb, 0xaa, 0x30, 0xa6, 0xac, 0x10, 0x73, 0xb7, 0xc8,
				0xdf, 0xff, 0xb8, 0x80, 0xe5, 0x02, 0xec, 0x0a, 0x03, 0x73, 0x27, 0xf2, 0x5c, 0xb3, 0x84, 0x63,
				0x45, 0x7b, 0xb8, 0x11, 0x38, 0xe0, 0x3c, 0x19, 0xc6, 0x42, 0x39, 0xd3, 0x59, 0x00, 0x81, 0xd0,
				0x1c, 0xf3, 0x57, 0x4e, 0x66, 0x3e, 0xa4, 0x70, 0xe5, 0xcc, 0x1a, 0xa0, 0x56, 0x33, 0x62, 0x80,
				0x3c, 0xca, 0x73, 0x23, 0x02, 0xa9, 0x05, 0x00, 0x00, 0x00, 0x04, 0xee, 0x12, 0xe3, 0x87, 0xc5,
				0x38, 0xfe, 0x48, 0xe7, 0xa8, 0xb6, 0x01, 0xd7, 0x93, 0x77, 0xa4, 0x2a, 0xf1, 0xe5, 0x75, 0x9d,
				0x39, 0xef, 0x12, 0x8a, 0x95, 0xc7, 0x53, 0x46, 0xe1, 0x54, 0x53, 0x00, 0x00, 0x00, 0x00, 0xbf,
				0x90, 0x98, 0x62, 0x01, 0xb6, 0x90, 0x98, 0x62, 0x01, 0x0d, 0x39, 0x34, 0x2e, 0x32, 0x33, 0x2e,
				0x32, 0x35, 0x31, 0x2e, 0x31, 0x36, 0x39, 0x41, 0x1b, 0xfb, 0xbf, 0xcc, 0xb9, 0x77, 0xa7, 0x95,
				0xc8, 0xdb, 0x4a, 0x57, 0x10, 0xce, 0xd5, 0x1a, 0x12, 0x51, 0xa8, 0x97, 0x12, 0xbd, 0xac, 0xf3,
				0xce, 0x45, 0xea, 0xa0, 0xc7, 0x71, 0x3a, 0xee, 0x8d, 0x64, 0xba, 0xf5, 0x17, 0x36, 0x8e, 0x34,
				0x3b, 0xbb, 0xb4, 0x0d, 0x63, 0x78, 0xe4, 0x8d, 0x99, 0xda, 0xee, 0xf5, 0xfe, 0x5f, 0xea, 0xff,
				0x91, 0xb8, 0xbd, 0x0f, 0x86, 0xb2, 0x88, 0x3f, 0x91, 0x41, 0x1c, 0x2e, 0x19, 0xdf, 0x0f, 0xe6,
				0x60, 0xf4, 0x4c, 0x93, 0xde, 0x18, 0x7d, 0x0e, 0x42, 0x4f, 0x6b, 0x4c, 0x6b, 0xa1, 0x46, 0x5a,
				0x82, 0x6a, 0xc5, 0xbb, 0x40, 0x76, 0xe5, 0x14, 0x62, 0xe7, 0x98, 0x24, 0x61, 0xd1, 0x6a, 0xee,
				0xd4, 0x00, 0x0b, 0x1e, 0xb3, 0x77, 0x5b, 0x48, 0xfa, 0x14, 0x22, 0xd3, 0xf9, 0x14, 0xb9, 0x42,
				0xd4, 0xc2, 0x97, 0x93, 0x62, 0x75, 0x0c, 0x6d, 0x57, 0x8f, 0x15, 0x05, 0x00, 0x00, 0x00, 0x04,
				0xf2, 0xbe, 0x7c, 0xbd, 0x7a, 0xd2, 0x42, 0xbd, 0x62, 0x81, 0xd4, 0x95, 0xa9, 0x0b, 0xa9, 0xd7,
				0xce, 0xf4, 0x23, 0x6d, 0xf0, 0xe6, 0xcb, 0xbd, 0x0c, 0x7f, 0x4b, 0x9c, 0x94, 0x95, 0x26, 0x93,
				0x00, 0x00, 0x00, 0x00, 0xbf, 0x90, 0x98, 0x62, 0x02, 0x56, 0x8c, 0x98, 0x62, 0x01, 0x0d, 0x31,
				0x39, 0x35, 0x2e, 0x33, 0x2e, 0x32, 0x32, 0x33, 0x2e, 0x31, 0x36, 0x32, 0x41, 0x1b, 0x4b, 0x00,
				0xc4, 0x7b, 0xce, 0xd3, 0x17, 0xd4, 0xea, 0x45, 0x68, 0x9d, 0x94, 0x7e, 0x94, 0xce, 0x24, 0x38,
				0x67, 0x58, 0x81, 0x9a, 0x82, 0xd8, 0x42, 0x67, 0xd0, 0x32, 0x63, 0x1b, 0x9c, 0xe4, 0x35, 0x95,
				0xad, 0x27, 0x56, 0xfb, 0xe3, 0x3d, 0x50, 0xbc, 0xc3, 0x14, 0xae, 0x8d, 0x2b, 0x4a, 0x9e, 0x0b,
				0x5e, 0xdc, 0xdf, 0xc9, 0x2c, 0xf2, 0x5e, 0xed, 0x0a, 0xdd, 0x95, 0xaa, 0x4a, 0x18, 0x41, 0x1b,
				0x24, 0xcd, 0x3e, 0x26, 0xb2, 0x9a, 0x87, 0xd7, 0xf2, 0x41, 0xec, 0x52, 0x71, 0x51, 0x97, 0x6b,
				0x51, 0x30, 0xc8, 0x3a, 0x3c, 0x84, 0x92, 0xba, 0xdb, 0x17, 0x85, 0xaa, 0xcc, 0x6d, 0xf9, 0x95,
				0x42, 0x31, 0xb6, 0xe8, 0xd0, 0xe6, 0xd4, 0x2c, 0xe9, 0xe3, 0xbb, 0x34, 0xe0, 0x6b, 0xc5, 0xcb,
				0x55, 0x8b, 0x70, 0x56, 0xc5, 0xb4, 0x83, 0x73, 0x66, 0x12, 0xa5, 0x45, 0xde, 0xe9, 0x42, 0x16,
			},
		},
		{
			height:      1134309,
			nTime:       1654180391,
			version:     4,
			bits:        "1d1afbb2",
			prevHash:    "000000097e8820dedf4951ba1f4389a03aa19b771488104e7c6abeda6c269f38",
			saplingRoot: "22758d2483e36192e2ca6beef539801f3d6218eb6e24125da16d9cdc7a8e98d9",
			txHashes: [][]byte{
				[]byte{
					0xc2, 0x1d, 0x92, 0x8a, 0xc6, 0xdb, 0x70, 0x3a, 0x99, 0x5f, 0x3d, 0xd0, 0x3d, 0xf2, 0xa2, 0x90,
					0xf5, 0x44, 0x81, 0xad, 0x12, 0x9c, 0xc4, 0x33, 0xd6, 0xfd, 0xed, 0x2a, 0xbd, 0x82, 0xdd, 0x24,
				},
			},
			txHexes: [][]byte{
				[]byte{
					0x04, 0x00, 0x00, 0x80, 0x85, 0x20, 0x2f, 0x89, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x29, 0x03, 0xe5,
					0x4e, 0x11, 0x00, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x50, 0x6f, 0x6f, 0x6c, 0x20, 0x68, 0x74, 0x74,
					0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x66, 0x6c, 0x75, 0x78, 0x2e, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x70,
					0x6f, 0x6f, 0x6c, 0x2e, 0x6f, 0x72, 0x67, 0xff, 0xff, 0xff, 0xff, 0x06, 0xc0, 0x0c, 0x0c, 0xdb,
					0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xaa, 0xc1, 0xa7, 0xfa, 0x91, 0xab, 0x04, 0x13,
					0xf7, 0x99, 0xdd, 0xba, 0x40, 0xa4, 0x1e, 0xb8, 0x4b, 0x78, 0x0e, 0x13, 0x88, 0xac, 0xa0, 0x11,
					0x87, 0x21, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xb5, 0x69, 0x86, 0x38, 0x57, 0x77,
					0x23, 0xbf, 0xdd, 0x4e, 0xf9, 0x2d, 0xe1, 0xf1, 0x30, 0xee, 0xf0, 0xb7, 0xb2, 0x7b, 0x88, 0xac,
					0x60, 0x1d, 0xe1, 0x37, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xbf, 0x64, 0xd0, 0xac,
					0x3f, 0x8c, 0xe0, 0xc7, 0x02, 0xb2, 0xe9, 0xd6, 0x78, 0xa3, 0x5e, 0x4a, 0x78, 0x02, 0x80, 0x1d,
					0x88, 0xac, 0x80, 0x46, 0x1c, 0x86, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xfb, 0xe7,
					0x14, 0xf8, 0xcc, 0x06, 0xa7, 0x43, 0x65, 0xe2, 0x3b, 0x47, 0x8a, 0x37, 0x17, 0x2c, 0x31, 0x70,
					0x8d, 0xc3, 0x88, 0xac, 0x60, 0x34, 0x3c, 0x02, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14,
					0x0a, 0x6c, 0xa8, 0x4c, 0x78, 0x4f, 0xcf, 0x8d, 0x06, 0xaa, 0x3b, 0xf1, 0x9d, 0x67, 0x91, 0x73,
					0x42, 0xe9, 0x59, 0x75, 0x88, 0xac, 0x60, 0x34, 0x3c, 0x02, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76,
					0xa9, 0x14, 0x1a, 0x22, 0x1c, 0x9f, 0x3b, 0xc1, 0xa4, 0xc2, 0x05, 0x62, 0x7b, 0x06, 0x31, 0xb3,
					0xab, 0x23, 0x78, 0x31, 0x6d, 0xda, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
			},
			work: &types.StratumWork{
				Nonce: new(types.Number).SetFromBytes(common.MustParseHex("2000c19a000000000000000000000000000000000000000000000000ba098e10")),
				EquihashSolution: []byte{
					0x34, 0x47, 0x72, 0x9e, 0x61, 0x2e, 0xd3, 0xc5, 0xc2, 0xde, 0xc2, 0x67, 0x30, 0xdb, 0x48, 0x6d,
					0x5e, 0x7d, 0x78, 0xd1, 0x3b, 0x07, 0x35, 0xb2, 0xcd, 0xdb, 0x54, 0x61, 0x9f, 0x6f, 0x31, 0xa6,
					0x46, 0x17, 0x8f, 0xb3, 0x5f, 0x3f, 0x84, 0x63, 0x6c, 0x54, 0x90, 0xed, 0xd4, 0xf6, 0xf9, 0xa5,
					0x01, 0x0f, 0x2a, 0xaa, 0xc8,
				},
			},
			header: []byte{
				0x04, 0x00, 0x00, 0x00, 0x38, 0x9f, 0x26, 0x6c, 0xda, 0xbe, 0x6a, 0x7c, 0x4e, 0x10, 0x88, 0x14,
				0x77, 0x9b, 0xa1, 0x3a, 0xa0, 0x89, 0x43, 0x1f, 0xba, 0x51, 0x49, 0xdf, 0xde, 0x20, 0x88, 0x7e,
				0x09, 0x00, 0x00, 0x00, 0x24, 0xdd, 0x82, 0xbd, 0x2a, 0xed, 0xfd, 0xd6, 0x33, 0xc4, 0x9c, 0x12,
				0xad, 0x81, 0x44, 0xf5, 0x90, 0xa2, 0xf2, 0x3d, 0xd0, 0x3d, 0x5f, 0x99, 0x3a, 0x70, 0xdb, 0xc6,
				0x8a, 0x92, 0x1d, 0xc2, 0xd9, 0x98, 0x8e, 0x7a, 0xdc, 0x9c, 0x6d, 0xa1, 0x5d, 0x12, 0x24, 0x6e,
				0xeb, 0x18, 0x62, 0x3d, 0x1f, 0x80, 0x39, 0xf5, 0xee, 0x6b, 0xca, 0xe2, 0x92, 0x61, 0xe3, 0x83,
				0x24, 0x8d, 0x75, 0x22, 0x27, 0xca, 0x98, 0x62, 0xb2, 0xfb, 0x1a, 0x1d, 0x20, 0x00, 0xc1, 0x9a,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xba, 0x09, 0x8e, 0x10,
			},
			headerHash: []byte{
				0x00, 0x00, 0x00, 0x0a, 0x55, 0x1c, 0x54, 0x02, 0x46, 0xac, 0x83, 0x32, 0x23, 0xa0, 0xc7, 0xe0,
				0xbc, 0x8e, 0xe1, 0x12, 0xf1, 0x59, 0xd9, 0x92, 0xe1, 0x5d, 0x7b, 0x54, 0x9d, 0xff, 0x64, 0x42,
			},
			block: []byte{
				0x04, 0x00, 0x00, 0x00, 0x38, 0x9f, 0x26, 0x6c, 0xda, 0xbe, 0x6a, 0x7c, 0x4e, 0x10, 0x88, 0x14,
				0x77, 0x9b, 0xa1, 0x3a, 0xa0, 0x89, 0x43, 0x1f, 0xba, 0x51, 0x49, 0xdf, 0xde, 0x20, 0x88, 0x7e,
				0x09, 0x00, 0x00, 0x00, 0x24, 0xdd, 0x82, 0xbd, 0x2a, 0xed, 0xfd, 0xd6, 0x33, 0xc4, 0x9c, 0x12,
				0xad, 0x81, 0x44, 0xf5, 0x90, 0xa2, 0xf2, 0x3d, 0xd0, 0x3d, 0x5f, 0x99, 0x3a, 0x70, 0xdb, 0xc6,
				0x8a, 0x92, 0x1d, 0xc2, 0xd9, 0x98, 0x8e, 0x7a, 0xdc, 0x9c, 0x6d, 0xa1, 0x5d, 0x12, 0x24, 0x6e,
				0xeb, 0x18, 0x62, 0x3d, 0x1f, 0x80, 0x39, 0xf5, 0xee, 0x6b, 0xca, 0xe2, 0x92, 0x61, 0xe3, 0x83,
				0x24, 0x8d, 0x75, 0x22, 0x27, 0xca, 0x98, 0x62, 0xb2, 0xfb, 0x1a, 0x1d, 0x20, 0x00, 0xc1, 0x9a,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xba, 0x09, 0x8e, 0x10, 0x34, 0x47, 0x72, 0x9e,
				0x61, 0x2e, 0xd3, 0xc5, 0xc2, 0xde, 0xc2, 0x67, 0x30, 0xdb, 0x48, 0x6d, 0x5e, 0x7d, 0x78, 0xd1,
				0x3b, 0x07, 0x35, 0xb2, 0xcd, 0xdb, 0x54, 0x61, 0x9f, 0x6f, 0x31, 0xa6, 0x46, 0x17, 0x8f, 0xb3,
				0x5f, 0x3f, 0x84, 0x63, 0x6c, 0x54, 0x90, 0xed, 0xd4, 0xf6, 0xf9, 0xa5, 0x01, 0x0f, 0x2a, 0xaa,
				0xc8, 0x01, 0x04, 0x00, 0x00, 0x80, 0x85, 0x20, 0x2f, 0x89, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x29,
				0x03, 0xe5, 0x4e, 0x11, 0x00, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x50, 0x6f, 0x6f, 0x6c, 0x20, 0x68,
				0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x66, 0x6c, 0x75, 0x78, 0x2e, 0x6d, 0x69, 0x6e, 0x65,
				0x72, 0x70, 0x6f, 0x6f, 0x6c, 0x2e, 0x6f, 0x72, 0x67, 0xff, 0xff, 0xff, 0xff, 0x06, 0xc0, 0x0c,
				0x0c, 0xdb, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xaa, 0xc1, 0xa7, 0xfa, 0x91, 0xab,
				0x04, 0x13, 0xf7, 0x99, 0xdd, 0xba, 0x40, 0xa4, 0x1e, 0xb8, 0x4b, 0x78, 0x0e, 0x13, 0x88, 0xac,
				0xa0, 0x11, 0x87, 0x21, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xb5, 0x69, 0x86, 0x38,
				0x57, 0x77, 0x23, 0xbf, 0xdd, 0x4e, 0xf9, 0x2d, 0xe1, 0xf1, 0x30, 0xee, 0xf0, 0xb7, 0xb2, 0x7b,
				0x88, 0xac, 0x60, 0x1d, 0xe1, 0x37, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xbf, 0x64,
				0xd0, 0xac, 0x3f, 0x8c, 0xe0, 0xc7, 0x02, 0xb2, 0xe9, 0xd6, 0x78, 0xa3, 0x5e, 0x4a, 0x78, 0x02,
				0x80, 0x1d, 0x88, 0xac, 0x80, 0x46, 0x1c, 0x86, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14,
				0xfb, 0xe7, 0x14, 0xf8, 0xcc, 0x06, 0xa7, 0x43, 0x65, 0xe2, 0x3b, 0x47, 0x8a, 0x37, 0x17, 0x2c,
				0x31, 0x70, 0x8d, 0xc3, 0x88, 0xac, 0x60, 0x34, 0x3c, 0x02, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76,
				0xa9, 0x14, 0x0a, 0x6c, 0xa8, 0x4c, 0x78, 0x4f, 0xcf, 0x8d, 0x06, 0xaa, 0x3b, 0xf1, 0x9d, 0x67,
				0x91, 0x73, 0x42, 0xe9, 0x59, 0x75, 0x88, 0xac, 0x60, 0x34, 0x3c, 0x02, 0x00, 0x00, 0x00, 0x00,
				0x19, 0x76, 0xa9, 0x14, 0x1a, 0x22, 0x1c, 0x9f, 0x3b, 0xc1, 0xa4, 0xc2, 0x05, 0x62, 0x7b, 0x06,
				0x31, 0xb3, 0xab, 0x23, 0x78, 0x31, 0x6d, 0xda, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
	}

	for i, tt := range tests {
		builder, err := NewEquihashBuilder(tt.version, tt.nTime, tt.bits, tt.prevHash, tt.saplingRoot, tt.txHashes, tt.txHexes)
		if err != nil {
			t.Errorf("failed on %d: NewEquihashBuilder: %v", i, err)
		}

		header, headerHash, err := builder.SerializeHeader(tt.work)
		if err != nil {
			t.Errorf("failed on %d: SerializeHeader: %v", i, err)
		} else if bytes.Compare(header, tt.header) != 0 {
			t.Errorf("failed on %d: header mismatch: have %x, want %x", i, header, tt.header)
		} else if bytes.Compare(headerHash, tt.headerHash) != 0 {
			t.Errorf("failed on %d: header hash mismatch: have %x, want %x", i, headerHash, tt.headerHash)
		}

		block, err := builder.SerializeBlock(tt.work)
		if err != nil {
			t.Errorf("failed on %d: SerializeBlock: %v", i, err)
		} else if bytes.Compare(block, tt.block) != 0 {
			t.Errorf("failed on %d: block mismatch: have %x, want %x", i, block, tt.block)
		}
	}
}

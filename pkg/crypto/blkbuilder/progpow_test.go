package blkbuilder

import (
	"bytes"
	"testing"

	"github.com/magicpool-co/pool/types"
)

func TestSerializeProgPow(t *testing.T) {
	tests := []struct {
		height     uint32
		nTime      uint32
		version    uint32
		bits       string
		prevHash   string
		txHashes   [][]byte
		header     []byte
		headerHash []byte
		work       *types.StratumWork
		txHexes    [][]byte
		block      []byte
	}{
		// RVN
		{
			height:   1874596,
			nTime:    1628302731,
			version:  805306368,
			bits:     "1b00f916",
			prevHash: "000000000000665a397c703be65c838d490d291e8d97e6ecf4006b89304c8387",
			txHashes: [][]byte{
				[]byte{
					0x29, 0xf1, 0x69, 0x6e, 0x5f, 0x52, 0x60, 0xf1, 0xe1, 0xda, 0xc7, 0x93, 0xad, 0x29, 0xdb, 0x8b,
					0x0a, 0xe9, 0x78, 0x9e, 0x08, 0xac, 0x6d, 0x72, 0x33, 0x24, 0xf4, 0xd7, 0xba, 0x35, 0x08, 0x58,
				},
			},
			header: []byte{
				0x00, 0x00, 0x00, 0x30, 0x87, 0x83, 0x4c, 0x30, 0x89, 0x6b, 0x00, 0xf4, 0xec, 0xe6, 0x97, 0x8d,
				0x1e, 0x29, 0x0d, 0x49, 0x8d, 0x83, 0x5c, 0xe6, 0x3b, 0x70, 0x7c, 0x39, 0x5a, 0x66, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x58, 0x08, 0x35, 0xba, 0xd7, 0xf4, 0x24, 0x33, 0x72, 0x6d, 0xac, 0x08,
				0x9e, 0x78, 0xe9, 0x0a, 0x8b, 0xdb, 0x29, 0xad, 0x93, 0xc7, 0xda, 0xe1, 0xf1, 0x60, 0x52, 0x5f,
				0x6e, 0x69, 0xf1, 0x29, 0x8b, 0xed, 0x0d, 0x61, 0x16, 0xf9, 0x00, 0x1b, 0xa4, 0x9a, 0x1c, 0x00,
			},
			headerHash: []byte{
				0x55, 0xd0, 0xd7, 0x73, 0x0e, 0x83, 0xf5, 0xef, 0x56, 0x2c, 0x1a, 0x77, 0x3f, 0x96, 0xd2, 0x21,
				0x3b, 0x30, 0x94, 0xc3, 0x67, 0x93, 0x30, 0xea, 0x5c, 0xc8, 0xbd, 0x5b, 0x24, 0xc2, 0xad, 0x58,
			},
			work: &types.StratumWork{
				MixDigest: new(types.Hash).SetFromBytes([]byte{
					0x57, 0x46, 0x43, 0x42, 0xdc, 0x5f, 0xb0, 0x1b, 0x6d, 0xb1, 0xf4, 0xc5, 0x83, 0xf0, 0x2f, 0x3c,
					0xae, 0x1f, 0xf4, 0x96, 0xb8, 0x8b, 0x45, 0x1f, 0xc5, 0xd5, 0xe3, 0xd8, 0x7d, 0x4e, 0x0c, 0x4d,
				}),
				Nonce: new(types.Number).SetFromValue(0x6a28f8000c5cc063),
			},
			txHexes: [][]byte{
				[]byte{
					0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x0e, 0x03, 0xa4, 0x9a, 0x1c, 0x00, 0x2f,
					0x42, 0x65, 0x65, 0x70, 0x6f, 0x6f, 0x6c, 0x2f, 0xff, 0xff, 0xff, 0xff, 0x01, 0x00, 0x88, 0x52,
					0x6a, 0x74, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0x41, 0x6e, 0x56, 0x05, 0x75, 0xb8, 0xf4,
					0x5d, 0xf6, 0xf7, 0xaf, 0x74, 0x7c, 0x4d, 0x98, 0x92, 0x1e, 0xf5, 0xd9, 0xd1, 0x88, 0xac, 0x00,
					0x00, 0x00, 0x00,
				},
			},
			block: []byte{
				0x00, 0x00, 0x00, 0x30, 0x87, 0x83, 0x4c, 0x30, 0x89, 0x6b, 0x00, 0xf4, 0xec, 0xe6, 0x97, 0x8d,
				0x1e, 0x29, 0x0d, 0x49, 0x8d, 0x83, 0x5c, 0xe6, 0x3b, 0x70, 0x7c, 0x39, 0x5a, 0x66, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x58, 0x08, 0x35, 0xba, 0xd7, 0xf4, 0x24, 0x33, 0x72, 0x6d, 0xac, 0x08,
				0x9e, 0x78, 0xe9, 0x0a, 0x8b, 0xdb, 0x29, 0xad, 0x93, 0xc7, 0xda, 0xe1, 0xf1, 0x60, 0x52, 0x5f,
				0x6e, 0x69, 0xf1, 0x29, 0x8b, 0xed, 0x0d, 0x61, 0x16, 0xf9, 0x00, 0x1b, 0xa4, 0x9a, 0x1c, 0x00,
				0x63, 0xc0, 0x5c, 0x0c, 0x00, 0xf8, 0x28, 0x6a, 0x4d, 0x0c, 0x4e, 0x7d, 0xd8, 0xe3, 0xd5, 0xc5,
				0x1f, 0x45, 0x8b, 0xb8, 0x96, 0xf4, 0x1f, 0xae, 0x3c, 0x2f, 0xf0, 0x83, 0xc5, 0xf4, 0xb1, 0x6d,
				0x1b, 0xb0, 0x5f, 0xdc, 0x42, 0x43, 0x46, 0x57, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff,
				0xff, 0xff, 0x0e, 0x03, 0xa4, 0x9a, 0x1c, 0x00, 0x2f, 0x42, 0x65, 0x65, 0x70, 0x6f, 0x6f, 0x6c,
				0x2f, 0xff, 0xff, 0xff, 0xff, 0x01, 0x00, 0x88, 0x52, 0x6a, 0x74, 0x00, 0x00, 0x00, 0x19, 0x76,
				0xa9, 0x14, 0x41, 0x6e, 0x56, 0x05, 0x75, 0xb8, 0xf4, 0x5d, 0xf6, 0xf7, 0xaf, 0x74, 0x7c, 0x4d,
				0x98, 0x92, 0x1e, 0xf5, 0xd9, 0xd1, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			height:   1874415,
			nTime:    1628291468,
			version:  805306368,
			bits:     "1b00f442",
			prevHash: "000000000000c160b2474557ba884b80a31de6aed149ed65f59dc504fd94d711",
			txHashes: [][]byte{
				[]byte{
					0xa0, 0x1a, 0x50, 0x11, 0xc3, 0xee, 0xbc, 0x4a, 0xe8, 0xed, 0xfe, 0x21, 0x26, 0xdb, 0xbe, 0x65,
					0xd6, 0x79, 0x22, 0xe3, 0x23, 0x67, 0xa7, 0xf5, 0x10, 0xdb, 0x20, 0xb5, 0xff, 0x07, 0xa4, 0x83,
				},
				[]byte{
					0x31, 0x8e, 0xad, 0xb7, 0x28, 0x95, 0xc6, 0x70, 0xed, 0x50, 0x6f, 0x45, 0x00, 0xc8, 0x0a, 0x7d,
					0x7c, 0xb4, 0x0a, 0xbb, 0x04, 0xeb, 0x9d, 0x76, 0xec, 0xf6, 0x05, 0x74, 0xc5, 0x9d, 0x35, 0x48,
				},
			},
			header: []byte{
				0x00, 0x00, 0x00, 0x30, 0x11, 0xd7, 0x94, 0xfd, 0x04, 0xc5, 0x9d, 0xf5, 0x65, 0xed, 0x49, 0xd1,
				0xae, 0xe6, 0x1d, 0xa3, 0x80, 0x4b, 0x88, 0xba, 0x57, 0x45, 0x47, 0xb2, 0x60, 0xc1, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0xfd, 0xdd, 0xbb, 0xe5, 0xbd, 0x18, 0xf6, 0xcc, 0x17, 0x66, 0x9a, 0xab,
				0xde, 0x1b, 0xb7, 0xcb, 0xf0, 0xcc, 0x3c, 0x25, 0x20, 0xd4, 0xa1, 0x07, 0x39, 0x7a, 0x80, 0xab,
				0x95, 0x1b, 0x5a, 0xfe, 0x8c, 0xc1, 0x0d, 0x61, 0x42, 0xf4, 0x00, 0x1b, 0xef, 0x99, 0x1c, 0x00,
			},
			headerHash: []byte{
				0xfd, 0xd9, 0x62, 0x5e, 0xa1, 0x63, 0xd7, 0xa7, 0x1f, 0x32, 0x2e, 0x94, 0x0f, 0xa1, 0xc2, 0xdc,
				0x57, 0x59, 0x89, 0x00, 0x93, 0x9c, 0x11, 0xba, 0xdf, 0x4f, 0x84, 0x64, 0xee, 0x7b, 0xc2, 0xe5,
			},
			work: &types.StratumWork{
				MixDigest: new(types.Hash).SetFromBytes([]byte{
					0x05, 0x92, 0x88, 0x73, 0x5a, 0xb6, 0xe1, 0x93, 0x7c, 0x28, 0x03, 0xb1, 0xd3, 0xa7, 0x8c, 0x05,
					0x43, 0xe8, 0xf1, 0xab, 0x37, 0x02, 0x28, 0x80, 0x22, 0xa4, 0x25, 0x7d, 0xa1, 0x22, 0x60, 0x30,
				}),
				Nonce: new(types.Number).SetFromValue(0x703791004c2fa52a),
			},
			txHexes: [][]byte{
				[]byte{
					0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x0e, 0x03, 0xef, 0x99, 0x1c, 0x00, 0x2f,
					0x42, 0x65, 0x65, 0x70, 0x6f, 0x6f, 0x6c, 0x2f, 0xff, 0xff, 0xff, 0xff, 0x01, 0xa2, 0x39, 0x5f,
					0x6a, 0x74, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0x41, 0x6e, 0x56, 0x05, 0x75, 0xb8, 0xf4,
					0x5d, 0xf6, 0xf7, 0xaf, 0x74, 0x7c, 0x4d, 0x98, 0x92, 0x1e, 0xf5, 0xd9, 0xd1, 0x88, 0xac, 0x00,
					0x00, 0x00, 0x00,
				},
				[]byte{
					0x01, 0x00, 0x00, 0x00, 0x05, 0xe2, 0x61, 0xc9, 0xa9, 0x45, 0x5e, 0xd0, 0x85, 0x95, 0xcb, 0x31,
					0xe9, 0xe6, 0x39, 0xe8, 0x9f, 0x8f, 0xa2, 0x39, 0x45, 0x03, 0x91, 0x82, 0x58, 0x90, 0xad, 0x1d,
					0xdf, 0xb9, 0x61, 0x4d, 0xc2, 0x08, 0x00, 0x00, 0x00, 0x6b, 0x48, 0x30, 0x45, 0x02, 0x21, 0x00,
					0xf8, 0xe8, 0x7c, 0xeb, 0xda, 0xc6, 0x68, 0x6b, 0xbe, 0x7d, 0x6b, 0x92, 0x51, 0x05, 0xba, 0xc7,
					0x41, 0x96, 0x82, 0xbe, 0x9c, 0xa1, 0xca, 0xc9, 0x14, 0xe4, 0xcc, 0xcd, 0xcd, 0x29, 0x5e, 0xd9,
					0x02, 0x20, 0x0e, 0x7c, 0xfe, 0x43, 0x83, 0x07, 0xd5, 0x9c, 0x26, 0xdb, 0x3e, 0x3b, 0x1f, 0x49,
					0x66, 0x2b, 0xf9, 0x22, 0x73, 0x3b, 0xc1, 0xc9, 0xe9, 0xc6, 0xca, 0x65, 0xeb, 0x25, 0xa1, 0x5a,
					0xa2, 0xfa, 0x01, 0x21, 0x02, 0x95, 0x9b, 0xdc, 0x48, 0x87, 0xb1, 0x7b, 0x08, 0x8f, 0x10, 0x8a,
					0xf2, 0x11, 0xfa, 0xce, 0xaa, 0xe0, 0xef, 0x99, 0x46, 0x57, 0xe8, 0x14, 0x5a, 0x04, 0xb2, 0xa8,
					0x8c, 0xca, 0x66, 0xbf, 0xde, 0xe5, 0xff, 0xff, 0xff, 0x5e, 0xb2, 0x55, 0x7e, 0x3f, 0x98, 0xf2,
					0xbc, 0x59, 0x71, 0x6e, 0x66, 0x99, 0xcf, 0xaa, 0xf1, 0xa6, 0x00, 0xc7, 0xb9, 0x00, 0x1b, 0x8e,
					0x6c, 0xfb, 0xda, 0x41, 0x51, 0x61, 0x67, 0xd4, 0x75, 0x04, 0x00, 0x00, 0x00, 0x6a, 0x47, 0x30,
					0x44, 0x02, 0x20, 0x39, 0xdf, 0x89, 0x9f, 0x34, 0x8f, 0x45, 0x3a, 0x8c, 0xa4, 0x53, 0x28, 0x0a,
					0xc0, 0x5c, 0x98, 0x45, 0x70, 0x27, 0x6f, 0xea, 0xd0, 0x1a, 0xc1, 0x7e, 0xd3, 0x21, 0xd1, 0xb1,
					0x42, 0x09, 0x52, 0x02, 0x20, 0x27, 0x16, 0x4f, 0x4b, 0x59, 0x7f, 0x68, 0xa8, 0x59, 0x36, 0x77,
					0x9f, 0x45, 0xef, 0xd6, 0xd9, 0x10, 0x22, 0xf3, 0xfa, 0x45, 0x2d, 0xb0, 0xb8, 0xdd, 0xf7, 0x4b,
					0xa2, 0xf2, 0xb7, 0x4e, 0x11, 0x01, 0x21, 0x02, 0x95, 0x9b, 0xdc, 0x48, 0x87, 0xb1, 0x7b, 0x08,
					0x8f, 0x10, 0x8a, 0xf2, 0x11, 0xfa, 0xce, 0xaa, 0xe0, 0xef, 0x99, 0x46, 0x57, 0xe8, 0x14, 0x5a,
					0x04, 0xb2, 0xa8, 0x8c, 0xca, 0x66, 0xbf, 0xde, 0x8e, 0xff, 0xff, 0xff, 0x63, 0xfb, 0xaa, 0x80,
					0xe6, 0x6b, 0x23, 0x68, 0x5c, 0x1c, 0x59, 0x3f, 0x6a, 0x46, 0x78, 0xef, 0xc4, 0x5a, 0x51, 0xf7,
					0x7a, 0x53, 0x76, 0x94, 0x85, 0x25, 0xbe, 0xd6, 0x2a, 0x3d, 0x8f, 0xc5, 0x03, 0x00, 0x00, 0x00,
					0x6a, 0x47, 0x30, 0x44, 0x02, 0x20, 0x69, 0x10, 0x91, 0x68, 0xf4, 0x6e, 0xb5, 0xf5, 0x9c, 0x74,
					0xe5, 0xbd, 0x3d, 0xfd, 0x12, 0x50, 0x21, 0x4d, 0x56, 0xb2, 0x07, 0x49, 0xa5, 0x92, 0x80, 0x5c,
					0xad, 0xad, 0xfa, 0x74, 0xb8, 0x9f, 0x02, 0x20, 0x3f, 0xc1, 0xca, 0xb8, 0x4f, 0xb4, 0x32, 0x81,
					0xe0, 0x3b, 0xf2, 0x07, 0x18, 0x7b, 0x8c, 0x3a, 0x9d, 0xb4, 0x95, 0x92, 0xbd, 0x28, 0xd8, 0xda,
					0xaf, 0x85, 0x3c, 0x23, 0x4b, 0x75, 0x82, 0x2c, 0x01, 0x21, 0x02, 0x95, 0x9b, 0xdc, 0x48, 0x87,
					0xb1, 0x7b, 0x08, 0x8f, 0x10, 0x8a, 0xf2, 0x11, 0xfa, 0xce, 0xaa, 0xe0, 0xef, 0x99, 0x46, 0x57,
					0xe8, 0x14, 0x5a, 0x04, 0xb2, 0xa8, 0x8c, 0xca, 0x66, 0xbf, 0xde, 0x30, 0xff, 0xff, 0xff, 0x06,
					0xb8, 0x45, 0xf2, 0xf6, 0x07, 0x70, 0x2b, 0x32, 0x33, 0xd6, 0x9b, 0x51, 0xe7, 0x55, 0x41, 0x93,
					0x96, 0xf9, 0xb0, 0x49, 0xa8, 0x7c, 0xd7, 0xad, 0x10, 0xa6, 0x72, 0x29, 0x9a, 0xf7, 0x77, 0x01,
					0x00, 0x00, 0x00, 0x6a, 0x47, 0x30, 0x44, 0x02, 0x20, 0x6a, 0x9b, 0xfd, 0x78, 0x64, 0x62, 0x0d,
					0x86, 0x17, 0x31, 0x52, 0xc1, 0x80, 0xb3, 0xac, 0xa4, 0x46, 0x79, 0xef, 0x40, 0x28, 0x4c, 0x08,
					0x5f, 0x50, 0x22, 0x11, 0x87, 0x7c, 0x0d, 0x3f, 0x13, 0x02, 0x20, 0x5f, 0x4f, 0xf8, 0xb6, 0x8f,
					0x72, 0xcd, 0x3f, 0xc2, 0x31, 0x3a, 0x65, 0x1b, 0x71, 0xf9, 0xe3, 0x3d, 0x03, 0x1e, 0x9d, 0x33,
					0x3c, 0x51, 0x63, 0x5e, 0xb2, 0xab, 0x7b, 0xfd, 0x60, 0x03, 0xe0, 0x01, 0x21, 0x02, 0x95, 0x9b,
					0xdc, 0x48, 0x87, 0xb1, 0x7b, 0x08, 0x8f, 0x10, 0x8a, 0xf2, 0x11, 0xfa, 0xce, 0xaa, 0xe0, 0xef,
					0x99, 0x46, 0x57, 0xe8, 0x14, 0x5a, 0x04, 0xb2, 0xa8, 0x8c, 0xca, 0x66, 0xbf, 0xde, 0x3d, 0xff,
					0xff, 0xff, 0x2e, 0x0d, 0x29, 0x1f, 0x4c, 0x70, 0x71, 0x2d, 0xd6, 0x21, 0x71, 0x27, 0x0e, 0x49,
					0x8f, 0xd5, 0x8c, 0x4e, 0x3b, 0x47, 0x4b, 0x42, 0x62, 0xd8, 0x16, 0x77, 0xeb, 0xd2, 0x46, 0xff,
					0xbd, 0xd4, 0x04, 0x00, 0x00, 0x00, 0x6a, 0x47, 0x30, 0x44, 0x02, 0x20, 0x1f, 0x26, 0xf4, 0x0e,
					0x70, 0x9a, 0xde, 0xde, 0xf4, 0x76, 0xed, 0xd9, 0xd4, 0x3d, 0x70, 0xa0, 0xad, 0x56, 0xf6, 0x7e,
					0xb9, 0x0b, 0x9d, 0xba, 0x05, 0xff, 0x33, 0x8f, 0xef, 0x7c, 0x95, 0x3d, 0x02, 0x20, 0x01, 0xe3,
					0x5c, 0xa3, 0x4a, 0x26, 0xa5, 0x3d, 0x7d, 0x34, 0xd7, 0x91, 0xb5, 0x66, 0x6a, 0x60, 0x0f, 0xee,
					0x43, 0x84, 0x30, 0x44, 0x71, 0x84, 0x4f, 0xcb, 0x90, 0x10, 0xbc, 0xaf, 0xa7, 0x99, 0x01, 0x21,
					0x02, 0x95, 0x9b, 0xdc, 0x48, 0x87, 0xb1, 0x7b, 0x08, 0x8f, 0x10, 0x8a, 0xf2, 0x11, 0xfa, 0xce,
					0xaa, 0xe0, 0xef, 0x99, 0x46, 0x57, 0xe8, 0x14, 0x5a, 0x04, 0xb2, 0xa8, 0x8c, 0xca, 0x66, 0xbf,
					0xde, 0x6c, 0xff, 0xff, 0xff, 0x02, 0x00, 0x82, 0x35, 0x7a, 0x0a, 0x00, 0x00, 0x00, 0x19, 0x76,
					0xa9, 0x14, 0xee, 0x75, 0x7a, 0x01, 0x45, 0xab, 0xfb, 0x49, 0x4c, 0xba, 0x53, 0xfa, 0xd3, 0x04,
					0x1b, 0xa7, 0xf0, 0x07, 0x05, 0x46, 0x88, 0xac, 0x4e, 0xbd, 0x39, 0xf9, 0x01, 0x00, 0x00, 0x00,
					0x19, 0x76, 0xa9, 0x14, 0x3b, 0x53, 0x1c, 0x6e, 0xb7, 0x4d, 0x5f, 0xbf, 0xea, 0x1f, 0xea, 0xdf,
					0xa0, 0xd7, 0x55, 0x71, 0x40, 0x78, 0xe2, 0xad, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00,
				},
			},
			block: []byte{
				0x00, 0x00, 0x00, 0x30, 0x11, 0xd7, 0x94, 0xfd, 0x04, 0xc5, 0x9d, 0xf5, 0x65, 0xed, 0x49, 0xd1,
				0xae, 0xe6, 0x1d, 0xa3, 0x80, 0x4b, 0x88, 0xba, 0x57, 0x45, 0x47, 0xb2, 0x60, 0xc1, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0xfd, 0xdd, 0xbb, 0xe5, 0xbd, 0x18, 0xf6, 0xcc, 0x17, 0x66, 0x9a, 0xab,
				0xde, 0x1b, 0xb7, 0xcb, 0xf0, 0xcc, 0x3c, 0x25, 0x20, 0xd4, 0xa1, 0x07, 0x39, 0x7a, 0x80, 0xab,
				0x95, 0x1b, 0x5a, 0xfe, 0x8c, 0xc1, 0x0d, 0x61, 0x42, 0xf4, 0x00, 0x1b, 0xef, 0x99, 0x1c, 0x00,
				0x2a, 0xa5, 0x2f, 0x4c, 0x00, 0x91, 0x37, 0x70, 0x30, 0x60, 0x22, 0xa1, 0x7d, 0x25, 0xa4, 0x22,
				0x80, 0x28, 0x02, 0x37, 0xab, 0xf1, 0xe8, 0x43, 0x05, 0x8c, 0xa7, 0xd3, 0xb1, 0x03, 0x28, 0x7c,
				0x93, 0xe1, 0xb6, 0x5a, 0x73, 0x88, 0x92, 0x05, 0x02, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff,
				0xff, 0xff, 0x0e, 0x03, 0xef, 0x99, 0x1c, 0x00, 0x2f, 0x42, 0x65, 0x65, 0x70, 0x6f, 0x6f, 0x6c,
				0x2f, 0xff, 0xff, 0xff, 0xff, 0x01, 0xa2, 0x39, 0x5f, 0x6a, 0x74, 0x00, 0x00, 0x00, 0x19, 0x76,
				0xa9, 0x14, 0x41, 0x6e, 0x56, 0x05, 0x75, 0xb8, 0xf4, 0x5d, 0xf6, 0xf7, 0xaf, 0x74, 0x7c, 0x4d,
				0x98, 0x92, 0x1e, 0xf5, 0xd9, 0xd1, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
				0x05, 0xe2, 0x61, 0xc9, 0xa9, 0x45, 0x5e, 0xd0, 0x85, 0x95, 0xcb, 0x31, 0xe9, 0xe6, 0x39, 0xe8,
				0x9f, 0x8f, 0xa2, 0x39, 0x45, 0x03, 0x91, 0x82, 0x58, 0x90, 0xad, 0x1d, 0xdf, 0xb9, 0x61, 0x4d,
				0xc2, 0x08, 0x00, 0x00, 0x00, 0x6b, 0x48, 0x30, 0x45, 0x02, 0x21, 0x00, 0xf8, 0xe8, 0x7c, 0xeb,
				0xda, 0xc6, 0x68, 0x6b, 0xbe, 0x7d, 0x6b, 0x92, 0x51, 0x05, 0xba, 0xc7, 0x41, 0x96, 0x82, 0xbe,
				0x9c, 0xa1, 0xca, 0xc9, 0x14, 0xe4, 0xcc, 0xcd, 0xcd, 0x29, 0x5e, 0xd9, 0x02, 0x20, 0x0e, 0x7c,
				0xfe, 0x43, 0x83, 0x07, 0xd5, 0x9c, 0x26, 0xdb, 0x3e, 0x3b, 0x1f, 0x49, 0x66, 0x2b, 0xf9, 0x22,
				0x73, 0x3b, 0xc1, 0xc9, 0xe9, 0xc6, 0xca, 0x65, 0xeb, 0x25, 0xa1, 0x5a, 0xa2, 0xfa, 0x01, 0x21,
				0x02, 0x95, 0x9b, 0xdc, 0x48, 0x87, 0xb1, 0x7b, 0x08, 0x8f, 0x10, 0x8a, 0xf2, 0x11, 0xfa, 0xce,
				0xaa, 0xe0, 0xef, 0x99, 0x46, 0x57, 0xe8, 0x14, 0x5a, 0x04, 0xb2, 0xa8, 0x8c, 0xca, 0x66, 0xbf,
				0xde, 0xe5, 0xff, 0xff, 0xff, 0x5e, 0xb2, 0x55, 0x7e, 0x3f, 0x98, 0xf2, 0xbc, 0x59, 0x71, 0x6e,
				0x66, 0x99, 0xcf, 0xaa, 0xf1, 0xa6, 0x00, 0xc7, 0xb9, 0x00, 0x1b, 0x8e, 0x6c, 0xfb, 0xda, 0x41,
				0x51, 0x61, 0x67, 0xd4, 0x75, 0x04, 0x00, 0x00, 0x00, 0x6a, 0x47, 0x30, 0x44, 0x02, 0x20, 0x39,
				0xdf, 0x89, 0x9f, 0x34, 0x8f, 0x45, 0x3a, 0x8c, 0xa4, 0x53, 0x28, 0x0a, 0xc0, 0x5c, 0x98, 0x45,
				0x70, 0x27, 0x6f, 0xea, 0xd0, 0x1a, 0xc1, 0x7e, 0xd3, 0x21, 0xd1, 0xb1, 0x42, 0x09, 0x52, 0x02,
				0x20, 0x27, 0x16, 0x4f, 0x4b, 0x59, 0x7f, 0x68, 0xa8, 0x59, 0x36, 0x77, 0x9f, 0x45, 0xef, 0xd6,
				0xd9, 0x10, 0x22, 0xf3, 0xfa, 0x45, 0x2d, 0xb0, 0xb8, 0xdd, 0xf7, 0x4b, 0xa2, 0xf2, 0xb7, 0x4e,
				0x11, 0x01, 0x21, 0x02, 0x95, 0x9b, 0xdc, 0x48, 0x87, 0xb1, 0x7b, 0x08, 0x8f, 0x10, 0x8a, 0xf2,
				0x11, 0xfa, 0xce, 0xaa, 0xe0, 0xef, 0x99, 0x46, 0x57, 0xe8, 0x14, 0x5a, 0x04, 0xb2, 0xa8, 0x8c,
				0xca, 0x66, 0xbf, 0xde, 0x8e, 0xff, 0xff, 0xff, 0x63, 0xfb, 0xaa, 0x80, 0xe6, 0x6b, 0x23, 0x68,
				0x5c, 0x1c, 0x59, 0x3f, 0x6a, 0x46, 0x78, 0xef, 0xc4, 0x5a, 0x51, 0xf7, 0x7a, 0x53, 0x76, 0x94,
				0x85, 0x25, 0xbe, 0xd6, 0x2a, 0x3d, 0x8f, 0xc5, 0x03, 0x00, 0x00, 0x00, 0x6a, 0x47, 0x30, 0x44,
				0x02, 0x20, 0x69, 0x10, 0x91, 0x68, 0xf4, 0x6e, 0xb5, 0xf5, 0x9c, 0x74, 0xe5, 0xbd, 0x3d, 0xfd,
				0x12, 0x50, 0x21, 0x4d, 0x56, 0xb2, 0x07, 0x49, 0xa5, 0x92, 0x80, 0x5c, 0xad, 0xad, 0xfa, 0x74,
				0xb8, 0x9f, 0x02, 0x20, 0x3f, 0xc1, 0xca, 0xb8, 0x4f, 0xb4, 0x32, 0x81, 0xe0, 0x3b, 0xf2, 0x07,
				0x18, 0x7b, 0x8c, 0x3a, 0x9d, 0xb4, 0x95, 0x92, 0xbd, 0x28, 0xd8, 0xda, 0xaf, 0x85, 0x3c, 0x23,
				0x4b, 0x75, 0x82, 0x2c, 0x01, 0x21, 0x02, 0x95, 0x9b, 0xdc, 0x48, 0x87, 0xb1, 0x7b, 0x08, 0x8f,
				0x10, 0x8a, 0xf2, 0x11, 0xfa, 0xce, 0xaa, 0xe0, 0xef, 0x99, 0x46, 0x57, 0xe8, 0x14, 0x5a, 0x04,
				0xb2, 0xa8, 0x8c, 0xca, 0x66, 0xbf, 0xde, 0x30, 0xff, 0xff, 0xff, 0x06, 0xb8, 0x45, 0xf2, 0xf6,
				0x07, 0x70, 0x2b, 0x32, 0x33, 0xd6, 0x9b, 0x51, 0xe7, 0x55, 0x41, 0x93, 0x96, 0xf9, 0xb0, 0x49,
				0xa8, 0x7c, 0xd7, 0xad, 0x10, 0xa6, 0x72, 0x29, 0x9a, 0xf7, 0x77, 0x01, 0x00, 0x00, 0x00, 0x6a,
				0x47, 0x30, 0x44, 0x02, 0x20, 0x6a, 0x9b, 0xfd, 0x78, 0x64, 0x62, 0x0d, 0x86, 0x17, 0x31, 0x52,
				0xc1, 0x80, 0xb3, 0xac, 0xa4, 0x46, 0x79, 0xef, 0x40, 0x28, 0x4c, 0x08, 0x5f, 0x50, 0x22, 0x11,
				0x87, 0x7c, 0x0d, 0x3f, 0x13, 0x02, 0x20, 0x5f, 0x4f, 0xf8, 0xb6, 0x8f, 0x72, 0xcd, 0x3f, 0xc2,
				0x31, 0x3a, 0x65, 0x1b, 0x71, 0xf9, 0xe3, 0x3d, 0x03, 0x1e, 0x9d, 0x33, 0x3c, 0x51, 0x63, 0x5e,
				0xb2, 0xab, 0x7b, 0xfd, 0x60, 0x03, 0xe0, 0x01, 0x21, 0x02, 0x95, 0x9b, 0xdc, 0x48, 0x87, 0xb1,
				0x7b, 0x08, 0x8f, 0x10, 0x8a, 0xf2, 0x11, 0xfa, 0xce, 0xaa, 0xe0, 0xef, 0x99, 0x46, 0x57, 0xe8,
				0x14, 0x5a, 0x04, 0xb2, 0xa8, 0x8c, 0xca, 0x66, 0xbf, 0xde, 0x3d, 0xff, 0xff, 0xff, 0x2e, 0x0d,
				0x29, 0x1f, 0x4c, 0x70, 0x71, 0x2d, 0xd6, 0x21, 0x71, 0x27, 0x0e, 0x49, 0x8f, 0xd5, 0x8c, 0x4e,
				0x3b, 0x47, 0x4b, 0x42, 0x62, 0xd8, 0x16, 0x77, 0xeb, 0xd2, 0x46, 0xff, 0xbd, 0xd4, 0x04, 0x00,
				0x00, 0x00, 0x6a, 0x47, 0x30, 0x44, 0x02, 0x20, 0x1f, 0x26, 0xf4, 0x0e, 0x70, 0x9a, 0xde, 0xde,
				0xf4, 0x76, 0xed, 0xd9, 0xd4, 0x3d, 0x70, 0xa0, 0xad, 0x56, 0xf6, 0x7e, 0xb9, 0x0b, 0x9d, 0xba,
				0x05, 0xff, 0x33, 0x8f, 0xef, 0x7c, 0x95, 0x3d, 0x02, 0x20, 0x01, 0xe3, 0x5c, 0xa3, 0x4a, 0x26,
				0xa5, 0x3d, 0x7d, 0x34, 0xd7, 0x91, 0xb5, 0x66, 0x6a, 0x60, 0x0f, 0xee, 0x43, 0x84, 0x30, 0x44,
				0x71, 0x84, 0x4f, 0xcb, 0x90, 0x10, 0xbc, 0xaf, 0xa7, 0x99, 0x01, 0x21, 0x02, 0x95, 0x9b, 0xdc,
				0x48, 0x87, 0xb1, 0x7b, 0x08, 0x8f, 0x10, 0x8a, 0xf2, 0x11, 0xfa, 0xce, 0xaa, 0xe0, 0xef, 0x99,
				0x46, 0x57, 0xe8, 0x14, 0x5a, 0x04, 0xb2, 0xa8, 0x8c, 0xca, 0x66, 0xbf, 0xde, 0x6c, 0xff, 0xff,
				0xff, 0x02, 0x00, 0x82, 0x35, 0x7a, 0x0a, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xee, 0x75,
				0x7a, 0x01, 0x45, 0xab, 0xfb, 0x49, 0x4c, 0xba, 0x53, 0xfa, 0xd3, 0x04, 0x1b, 0xa7, 0xf0, 0x07,
				0x05, 0x46, 0x88, 0xac, 0x4e, 0xbd, 0x39, 0xf9, 0x01, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14,
				0x3b, 0x53, 0x1c, 0x6e, 0xb7, 0x4d, 0x5f, 0xbf, 0xea, 0x1f, 0xea, 0xdf, 0xa0, 0xd7, 0x55, 0x71,
				0x40, 0x78, 0xe2, 0xad, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00,
			},
		},
		// FIRO
		{
			height:   481899,
			nTime:    1654104368,
			version:  536875008,
			bits:     "1b0c56ee",
			prevHash: "4e6fd7d014186bdde1c8a6ae39cd503dbb0121f66a0f51ac82d4b74829e5423f",
			txHashes: [][]byte{
				[]byte{
					0xe7, 0xf9, 0x38, 0x2a, 0x0c, 0xdc, 0x0d, 0x4d, 0x9e, 0xd4, 0xd5, 0x6a, 0xab, 0xbc, 0x58, 0x65,
					0x1e, 0xb3, 0x11, 0xff, 0x52, 0x82, 0x49, 0x93, 0xd1, 0x69, 0xf3, 0xae, 0x14, 0xe5, 0x2a, 0xd4,
				},
			},
			header: []byte{
				0x00, 0x10, 0x00, 0x20, 0x3f, 0x42, 0xe5, 0x29, 0x48, 0xb7, 0xd4, 0x82, 0xac, 0x51, 0x0f, 0x6a,
				0xf6, 0x21, 0x01, 0xbb, 0x3d, 0x50, 0xcd, 0x39, 0xae, 0xa6, 0xc8, 0xe1, 0xdd, 0x6b, 0x18, 0x14,
				0xd0, 0xd7, 0x6f, 0x4e, 0xd4, 0x2a, 0xe5, 0x14, 0xae, 0xf3, 0x69, 0xd1, 0x93, 0x49, 0x82, 0x52,
				0xff, 0x11, 0xb3, 0x1e, 0x65, 0x58, 0xbc, 0xab, 0x6a, 0xd5, 0xd4, 0x9e, 0x4d, 0x0d, 0xdc, 0x0c,
				0x2a, 0x38, 0xf9, 0xe7, 0x30, 0xa1, 0x97, 0x62, 0xee, 0x56, 0x0c, 0x1b, 0x6b, 0x5a, 0x07, 0x00,
			},
			headerHash: []byte{
				0x0f, 0xbb, 0x41, 0x9b, 0x0c, 0xa4, 0x4e, 0x66, 0x81, 0x7f, 0xa2, 0x7c, 0x82, 0x8b, 0xb0, 0xe3,
				0x54, 0x78, 0xf3, 0x4f, 0x35, 0xc4, 0x17, 0x8b, 0x96, 0x99, 0xd7, 0x02, 0xb4, 0x86, 0x73, 0xc9,
			},
			work: &types.StratumWork{
				MixDigest: new(types.Hash).SetFromBytes([]byte{
					0x22, 0xe1, 0x01, 0x37, 0x46, 0xdb, 0x44, 0xa3, 0x08, 0xfc, 0x9c, 0x11, 0x8e, 0x6a, 0x07, 0x45,
					0xd6, 0x40, 0x2d, 0xaa, 0xe5, 0x5b, 0x74, 0xd3, 0x74, 0xf1, 0x17, 0x83, 0x83, 0x44, 0xdd, 0xd8,
				}),
				Nonce: new(types.Number).SetFromValue(0xa016e00027755951),
			},
			txHexes: [][]byte{
				[]byte{
					0x03, 0x00, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x1b, 0x03, 0x6b, 0x5a, 0x07, 0x04, 0x30,
					0xa1, 0x97, 0x62, 0x04, 0x00, 0x00, 0x00, 0x00, 0x0c, 0x2f, 0x57, 0x6f, 0x6f, 0x6c, 0x79, 0x50,
					0x6f, 0x6f, 0x6c, 0x79, 0x2f, 0xff, 0xff, 0xff, 0xff, 0x03, 0xe0, 0x05, 0x2d, 0x0b, 0x00, 0x00,
					0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xa5, 0xfe, 0xd1, 0x2b, 0x0f, 0xeb, 0x74, 0x16, 0x4a, 0x20,
					0x4a, 0x11, 0xfc, 0xa3, 0xcd, 0x72, 0x2f, 0x37, 0x27, 0x58, 0x88, 0xac, 0x60, 0xb8, 0x13, 0x1a,
					0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0x48, 0xd6, 0x09, 0xea, 0x2d, 0x5a, 0x5b, 0xff,
					0x42, 0x8e, 0x0c, 0x23, 0x04, 0xb9, 0xf7, 0xfb, 0xd5, 0x68, 0xcf, 0xae, 0x88, 0xac, 0x40, 0xbe,
					0x40, 0x25, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0x26, 0x23, 0x40, 0xfc, 0x2a, 0x99,
					0x26, 0x3e, 0x43, 0xb4, 0x85, 0x9e, 0x0d, 0x4d, 0x48, 0x14, 0x14, 0xc5, 0xae, 0xdc, 0x88, 0xac,
					0x00, 0x00, 0x00, 0x00, 0x46, 0x02, 0x00, 0x6b, 0x5a, 0x07, 0x00, 0xf7, 0xd9, 0x80, 0xf7, 0x30,
					0x25, 0x97, 0x2c, 0xf5, 0x62, 0x98, 0x85, 0xf5, 0x43, 0x71, 0x2f, 0xf6, 0x6c, 0xb1, 0x83, 0x95,
					0x36, 0x2e, 0x87, 0x1e, 0x1e, 0xf8, 0x69, 0xf6, 0x74, 0xdd, 0x87, 0xd9, 0xb0, 0x6e, 0xba, 0x31,
					0xfc, 0xe1, 0xd8, 0x5b, 0xec, 0x4d, 0x67, 0x41, 0xeb, 0x05, 0x6c, 0x51, 0x7b, 0xa5, 0x8f, 0x6f,
					0x03, 0xde, 0xa0, 0x4d, 0x68, 0x09, 0xf2, 0x80, 0xf9, 0xad, 0xfa,
				},
			},
			block: []byte{
				0x00, 0x10, 0x00, 0x20, 0x3f, 0x42, 0xe5, 0x29, 0x48, 0xb7, 0xd4, 0x82, 0xac, 0x51, 0x0f, 0x6a,
				0xf6, 0x21, 0x01, 0xbb, 0x3d, 0x50, 0xcd, 0x39, 0xae, 0xa6, 0xc8, 0xe1, 0xdd, 0x6b, 0x18, 0x14,
				0xd0, 0xd7, 0x6f, 0x4e, 0xd4, 0x2a, 0xe5, 0x14, 0xae, 0xf3, 0x69, 0xd1, 0x93, 0x49, 0x82, 0x52,
				0xff, 0x11, 0xb3, 0x1e, 0x65, 0x58, 0xbc, 0xab, 0x6a, 0xd5, 0xd4, 0x9e, 0x4d, 0x0d, 0xdc, 0x0c,
				0x2a, 0x38, 0xf9, 0xe7, 0x30, 0xa1, 0x97, 0x62, 0xee, 0x56, 0x0c, 0x1b, 0x6b, 0x5a, 0x07, 0x00,
				0x51, 0x59, 0x75, 0x27, 0x00, 0xe0, 0x16, 0xa0, 0xd8, 0xdd, 0x44, 0x83, 0x83, 0x17, 0xf1, 0x74,
				0xd3, 0x74, 0x5b, 0xe5, 0xaa, 0x2d, 0x40, 0xd6, 0x45, 0x07, 0x6a, 0x8e, 0x11, 0x9c, 0xfc, 0x08,
				0xa3, 0x44, 0xdb, 0x46, 0x37, 0x01, 0xe1, 0x22, 0x01, 0x03, 0x00, 0x05, 0x00, 0x01, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff,
				0xff, 0xff, 0x1b, 0x03, 0x6b, 0x5a, 0x07, 0x04, 0x30, 0xa1, 0x97, 0x62, 0x04, 0x00, 0x00, 0x00,
				0x00, 0x0c, 0x2f, 0x57, 0x6f, 0x6f, 0x6c, 0x79, 0x50, 0x6f, 0x6f, 0x6c, 0x79, 0x2f, 0xff, 0xff,
				0xff, 0xff, 0x03, 0xe0, 0x05, 0x2d, 0x0b, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xa5,
				0xfe, 0xd1, 0x2b, 0x0f, 0xeb, 0x74, 0x16, 0x4a, 0x20, 0x4a, 0x11, 0xfc, 0xa3, 0xcd, 0x72, 0x2f,
				0x37, 0x27, 0x58, 0x88, 0xac, 0x60, 0xb8, 0x13, 0x1a, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9,
				0x14, 0x48, 0xd6, 0x09, 0xea, 0x2d, 0x5a, 0x5b, 0xff, 0x42, 0x8e, 0x0c, 0x23, 0x04, 0xb9, 0xf7,
				0xfb, 0xd5, 0x68, 0xcf, 0xae, 0x88, 0xac, 0x40, 0xbe, 0x40, 0x25, 0x00, 0x00, 0x00, 0x00, 0x19,
				0x76, 0xa9, 0x14, 0x26, 0x23, 0x40, 0xfc, 0x2a, 0x99, 0x26, 0x3e, 0x43, 0xb4, 0x85, 0x9e, 0x0d,
				0x4d, 0x48, 0x14, 0x14, 0xc5, 0xae, 0xdc, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00, 0x46, 0x02, 0x00,
				0x6b, 0x5a, 0x07, 0x00, 0xf7, 0xd9, 0x80, 0xf7, 0x30, 0x25, 0x97, 0x2c, 0xf5, 0x62, 0x98, 0x85,
				0xf5, 0x43, 0x71, 0x2f, 0xf6, 0x6c, 0xb1, 0x83, 0x95, 0x36, 0x2e, 0x87, 0x1e, 0x1e, 0xf8, 0x69,
				0xf6, 0x74, 0xdd, 0x87, 0xd9, 0xb0, 0x6e, 0xba, 0x31, 0xfc, 0xe1, 0xd8, 0x5b, 0xec, 0x4d, 0x67,
				0x41, 0xeb, 0x05, 0x6c, 0x51, 0x7b, 0xa5, 0x8f, 0x6f, 0x03, 0xde, 0xa0, 0x4d, 0x68, 0x09, 0xf2,
				0x80, 0xf9, 0xad, 0xfa,
			},
		},
		{
			height:   482105,
			nTime:    1654168961,
			version:  536875008,
			bits:     "1b0c57f5",
			prevHash: "15c816235988763abc9b029259130d38f6cfad50eefd19818c03eff7b50445a5",
			txHashes: [][]byte{
				[]byte{
					0xe4, 0x3c, 0x49, 0x67, 0x96, 0x2a, 0xe9, 0xe9, 0xf0, 0x79, 0x12, 0x94, 0xcb, 0x96, 0x96, 0x82,
					0xee, 0x45, 0x41, 0x51, 0x6d, 0x53, 0xe4, 0x48, 0xed, 0x69, 0xd9, 0x20, 0x4f, 0x40, 0xd6, 0xb1,
				},
				[]byte{
					0x46, 0xc6, 0x96, 0x54, 0xc5, 0x8f, 0x5a, 0x27, 0x3e, 0x0b, 0x33, 0xe6, 0xe5, 0x54, 0x8c, 0x63,
					0xad, 0x99, 0xda, 0x35, 0x5b, 0x53, 0x01, 0x32, 0xba, 0xde, 0x70, 0xf4, 0xa0, 0x9a, 0x7f, 0x81,
				},
				[]byte{
					0xee, 0xd0, 0x28, 0xb4, 0xb0, 0x71, 0x2f, 0x34, 0x15, 0xf7, 0x1f, 0xb9, 0x59, 0x02, 0xcd, 0xb0,
					0xab, 0xff, 0x62, 0x41, 0x99, 0x8a, 0xf5, 0xf9, 0x1c, 0xe9, 0xf4, 0x2e, 0x90, 0xd4, 0x23, 0xc4,
				},
			},
			header: []byte{
				0x00, 0x10, 0x00, 0x20, 0xa5, 0x45, 0x04, 0xb5, 0xf7, 0xef, 0x03, 0x8c, 0x81, 0x19, 0xfd, 0xee,
				0x50, 0xad, 0xcf, 0xf6, 0x38, 0x0d, 0x13, 0x59, 0x92, 0x02, 0x9b, 0xbc, 0x3a, 0x76, 0x88, 0x59,
				0x23, 0x16, 0xc8, 0x15, 0x8d, 0x2c, 0x54, 0x62, 0xca, 0x50, 0x6a, 0x5b, 0x0b, 0xd5, 0xd0, 0xe1,
				0xe4, 0x15, 0xdf, 0x08, 0xa3, 0xae, 0xa0, 0x0f, 0xda, 0x8d, 0xca, 0xc7, 0xa1, 0xd8, 0xef, 0x1f,
				0x44, 0xf3, 0x07, 0x54, 0x81, 0x9d, 0x98, 0x62, 0xf5, 0x57, 0x0c, 0x1b, 0x39, 0x5b, 0x07, 0x00,
			},
			headerHash: []byte{
				0x1e, 0x41, 0x4a, 0x02, 0x6b, 0xc8, 0x16, 0xf4, 0x02, 0x6c, 0xd8, 0x8e, 0xdb, 0x9c, 0x6b, 0xbd,
				0xed, 0x30, 0xd4, 0x3f, 0x01, 0x10, 0xe4, 0x8f, 0x2d, 0xf3, 0x85, 0x80, 0xc5, 0x25, 0xb3, 0x31,
			},
			work: &types.StratumWork{
				MixDigest: new(types.Hash).SetFromBytes([]byte{
					0xb8, 0xcc, 0x2b, 0x05, 0x7b, 0x84, 0xeb, 0xf8, 0x8d, 0x8a, 0xcc, 0x51, 0xe1, 0x15, 0x26, 0x8f,
					0x10, 0xea, 0xca, 0x9f, 0xe0, 0x65, 0x15, 0x31, 0x3f, 0x87, 0x07, 0x89, 0xf2, 0xe3, 0xcd, 0x45,
				}),
				Nonce: new(types.Number).SetFromValue(0x988000015901d178),
			},
			txHexes: [][]byte{
				[]byte{
					0x03, 0x00, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0x2e, 0x03, 0x39, 0x5b, 0x07, 0x04, 0x81,
					0x9d, 0x98, 0x62, 0x08, 0x07, 0x00, 0x00, 0x00, 0x00, 0x34, 0x29, 0xa9, 0x1b, 0x32, 0x4d, 0x69,
					0x6e, 0x65, 0x72, 0x73, 0x20, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x32, 0x6d, 0x69,
					0x6e, 0x65, 0x72, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0xff, 0xff, 0xff, 0xff, 0x03, 0x22, 0xbf, 0x40,
					0x25, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0x55, 0x15, 0x4e, 0xc4, 0x38, 0x5f, 0x71,
					0xc4, 0xa2, 0x84, 0x73, 0x1c, 0xaf, 0xaf, 0x5d, 0x19, 0x40, 0x6c, 0x03, 0x05, 0x88, 0xac, 0x60,
					0xb8, 0x13, 0x1a, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0x12, 0x94, 0x9a, 0xd1, 0xf4,
					0x3b, 0x77, 0x0d, 0x08, 0x60, 0x46, 0x6c, 0xfb, 0x5a, 0x30, 0xd8, 0x7f, 0xb1, 0x66, 0xf7, 0x88,
					0xac, 0xe0, 0x05, 0x2d, 0x0b, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xa5, 0xfe, 0xd1,
					0x2b, 0x0f, 0xeb, 0x74, 0x16, 0x4a, 0x20, 0x4a, 0x11, 0xfc, 0xa3, 0xcd, 0x72, 0x2f, 0x37, 0x27,
					0x58, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00, 0x46, 0x02, 0x00, 0x39, 0x5b, 0x07, 0x00, 0xf8, 0x49,
					0xee, 0x0f, 0xb3, 0x30, 0x25, 0x0e, 0xe8, 0x83, 0x43, 0x13, 0x1d, 0x88, 0x9c, 0x4f, 0xa0, 0xe2,
					0x5d, 0x17, 0x8a, 0x87, 0x21, 0xad, 0x35, 0x28, 0x45, 0x7f, 0x6d, 0x68, 0x31, 0x4c, 0x6b, 0x3c,
					0x8b, 0x2b, 0xe9, 0x73, 0xe6, 0x84, 0x64, 0x33, 0xda, 0xbb, 0xab, 0x82, 0x97, 0x38, 0xe5, 0x29,
					0xdc, 0x3b, 0xd9, 0xdc, 0x09, 0x8b, 0xa8, 0xcb, 0x93, 0x08, 0x67, 0x84, 0x06, 0xba,
				},
				[]byte{
					0x03, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfd, 0x49, 0x01, 0x01, 0x00, 0x39,
					0x5b, 0x07, 0x00, 0x01, 0x00, 0x01, 0x83, 0xff, 0x8c, 0x19, 0x1c, 0x9f, 0x73, 0xbd, 0x3e, 0x35,
					0x5a, 0x2c, 0x98, 0x30, 0x9e, 0x9a, 0x7f, 0x33, 0xc5, 0x2a, 0x47, 0x10, 0x8b, 0x53, 0x3b, 0x5e,
					0x00, 0x96, 0x4f, 0x8f, 0xd1, 0x9f, 0x32, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x32, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				},
				[]byte{
					0x01, 0x00, 0x00, 0x00, 0x01, 0x66, 0x1e, 0xfb, 0x57, 0x4e, 0x6f, 0x9c, 0xed, 0xde, 0xd3, 0x43,
					0x7d, 0x5c, 0xa8, 0xfd, 0x1b, 0x87, 0x6b, 0xe6, 0xdf, 0x24, 0x88, 0x43, 0x54, 0x2f, 0x62, 0xfc,
					0x15, 0x5b, 0x17, 0xb8, 0x93, 0x03, 0x00, 0x00, 0x00, 0x6b, 0x48, 0x30, 0x45, 0x02, 0x21, 0x00,
					0xc6, 0x9f, 0x35, 0x81, 0xc2, 0x4b, 0x6b, 0x5d, 0x18, 0x24, 0x3f, 0xd6, 0xdc, 0x36, 0x15, 0x5e,
					0x74, 0x70, 0xe6, 0x21, 0x5c, 0x9e, 0xa5, 0x14, 0x66, 0x9b, 0x5b, 0xce, 0x1a, 0x5c, 0x33, 0x9b,
					0x02, 0x20, 0x2f, 0x82, 0xc9, 0xef, 0x6d, 0x30, 0x4c, 0x40, 0x4f, 0x6b, 0x59, 0xa9, 0xe5, 0x50,
					0xaf, 0x4a, 0x89, 0xdf, 0x37, 0x36, 0x68, 0x19, 0x60, 0x09, 0xc7, 0xe2, 0x80, 0x6c, 0xe3, 0xf1,
					0xbd, 0x85, 0x01, 0x21, 0x02, 0xd8, 0xbe, 0xd3, 0x94, 0xcf, 0xa4, 0x2e, 0x5e, 0x70, 0x7e, 0x5c,
					0x67, 0x5a, 0xe5, 0x09, 0x86, 0x0c, 0xaa, 0xa0, 0x72, 0x83, 0x9b, 0x31, 0x49, 0x81, 0x83, 0x91,
					0x8d, 0xa5, 0x1c, 0xdc, 0x2b, 0xfe, 0xff, 0xff, 0xff, 0x02, 0x54, 0x7c, 0x39, 0x02, 0x00, 0x00,
					0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xfd, 0x68, 0xce, 0xaf, 0x03, 0xf0, 0xce, 0x11, 0xa6, 0x46,
					0x5e, 0xf2, 0x5b, 0x21, 0x99, 0x8c, 0xd3, 0x91, 0x8c, 0xfd, 0x88, 0xac, 0x00, 0x65, 0xcd, 0x1d,
					0x00, 0x00, 0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0x43, 0x3d, 0xd2, 0x70, 0x29, 0xc3, 0x44, 0x99,
					0x17, 0x63, 0x89, 0x67, 0xac, 0x55, 0xb1, 0x96, 0x59, 0xea, 0x69, 0x2a, 0x88, 0xac, 0x37, 0x5b,
					0x07, 0x00,
				},
			},
			block: []byte{
				0x00, 0x10, 0x00, 0x20, 0xa5, 0x45, 0x04, 0xb5, 0xf7, 0xef, 0x03, 0x8c, 0x81, 0x19, 0xfd, 0xee,
				0x50, 0xad, 0xcf, 0xf6, 0x38, 0x0d, 0x13, 0x59, 0x92, 0x02, 0x9b, 0xbc, 0x3a, 0x76, 0x88, 0x59,
				0x23, 0x16, 0xc8, 0x15, 0x8d, 0x2c, 0x54, 0x62, 0xca, 0x50, 0x6a, 0x5b, 0x0b, 0xd5, 0xd0, 0xe1,
				0xe4, 0x15, 0xdf, 0x08, 0xa3, 0xae, 0xa0, 0x0f, 0xda, 0x8d, 0xca, 0xc7, 0xa1, 0xd8, 0xef, 0x1f,
				0x44, 0xf3, 0x07, 0x54, 0x81, 0x9d, 0x98, 0x62, 0xf5, 0x57, 0x0c, 0x1b, 0x39, 0x5b, 0x07, 0x00,
				0x78, 0xd1, 0x01, 0x59, 0x01, 0x00, 0x80, 0x98, 0x45, 0xcd, 0xe3, 0xf2, 0x89, 0x07, 0x87, 0x3f,
				0x31, 0x15, 0x65, 0xe0, 0x9f, 0xca, 0xea, 0x10, 0x8f, 0x26, 0x15, 0xe1, 0x51, 0xcc, 0x8a, 0x8d,
				0xf8, 0xeb, 0x84, 0x7b, 0x05, 0x2b, 0xcc, 0xb8, 0x03, 0x03, 0x00, 0x05, 0x00, 0x01, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff,
				0xff, 0xff, 0x2e, 0x03, 0x39, 0x5b, 0x07, 0x04, 0x81, 0x9d, 0x98, 0x62, 0x08, 0x07, 0x00, 0x00,
				0x00, 0x00, 0x34, 0x29, 0xa9, 0x1b, 0x32, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x20, 0x68, 0x74,
				0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x32, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x2e, 0x63, 0x6f,
				0x6d, 0xff, 0xff, 0xff, 0xff, 0x03, 0x22, 0xbf, 0x40, 0x25, 0x00, 0x00, 0x00, 0x00, 0x19, 0x76,
				0xa9, 0x14, 0x55, 0x15, 0x4e, 0xc4, 0x38, 0x5f, 0x71, 0xc4, 0xa2, 0x84, 0x73, 0x1c, 0xaf, 0xaf,
				0x5d, 0x19, 0x40, 0x6c, 0x03, 0x05, 0x88, 0xac, 0x60, 0xb8, 0x13, 0x1a, 0x00, 0x00, 0x00, 0x00,
				0x19, 0x76, 0xa9, 0x14, 0x12, 0x94, 0x9a, 0xd1, 0xf4, 0x3b, 0x77, 0x0d, 0x08, 0x60, 0x46, 0x6c,
				0xfb, 0x5a, 0x30, 0xd8, 0x7f, 0xb1, 0x66, 0xf7, 0x88, 0xac, 0xe0, 0x05, 0x2d, 0x0b, 0x00, 0x00,
				0x00, 0x00, 0x19, 0x76, 0xa9, 0x14, 0xa5, 0xfe, 0xd1, 0x2b, 0x0f, 0xeb, 0x74, 0x16, 0x4a, 0x20,
				0x4a, 0x11, 0xfc, 0xa3, 0xcd, 0x72, 0x2f, 0x37, 0x27, 0x58, 0x88, 0xac, 0x00, 0x00, 0x00, 0x00,
				0x46, 0x02, 0x00, 0x39, 0x5b, 0x07, 0x00, 0xf8, 0x49, 0xee, 0x0f, 0xb3, 0x30, 0x25, 0x0e, 0xe8,
				0x83, 0x43, 0x13, 0x1d, 0x88, 0x9c, 0x4f, 0xa0, 0xe2, 0x5d, 0x17, 0x8a, 0x87, 0x21, 0xad, 0x35,
				0x28, 0x45, 0x7f, 0x6d, 0x68, 0x31, 0x4c, 0x6b, 0x3c, 0x8b, 0x2b, 0xe9, 0x73, 0xe6, 0x84, 0x64,
				0x33, 0xda, 0xbb, 0xab, 0x82, 0x97, 0x38, 0xe5, 0x29, 0xdc, 0x3b, 0xd9, 0xdc, 0x09, 0x8b, 0xa8,
				0xcb, 0x93, 0x08, 0x67, 0x84, 0x06, 0xba, 0x03, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xfd, 0x49, 0x01, 0x01, 0x00, 0x39, 0x5b, 0x07, 0x00, 0x01, 0x00, 0x01, 0x83, 0xff, 0x8c,
				0x19, 0x1c, 0x9f, 0x73, 0xbd, 0x3e, 0x35, 0x5a, 0x2c, 0x98, 0x30, 0x9e, 0x9a, 0x7f, 0x33, 0xc5,
				0x2a, 0x47, 0x10, 0x8b, 0x53, 0x3b, 0x5e, 0x00, 0x96, 0x4f, 0x8f, 0xd1, 0x9f, 0x32, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x32, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
				0x00, 0x01, 0x66, 0x1e, 0xfb, 0x57, 0x4e, 0x6f, 0x9c, 0xed, 0xde, 0xd3, 0x43, 0x7d, 0x5c, 0xa8,
				0xfd, 0x1b, 0x87, 0x6b, 0xe6, 0xdf, 0x24, 0x88, 0x43, 0x54, 0x2f, 0x62, 0xfc, 0x15, 0x5b, 0x17,
				0xb8, 0x93, 0x03, 0x00, 0x00, 0x00, 0x6b, 0x48, 0x30, 0x45, 0x02, 0x21, 0x00, 0xc6, 0x9f, 0x35,
				0x81, 0xc2, 0x4b, 0x6b, 0x5d, 0x18, 0x24, 0x3f, 0xd6, 0xdc, 0x36, 0x15, 0x5e, 0x74, 0x70, 0xe6,
				0x21, 0x5c, 0x9e, 0xa5, 0x14, 0x66, 0x9b, 0x5b, 0xce, 0x1a, 0x5c, 0x33, 0x9b, 0x02, 0x20, 0x2f,
				0x82, 0xc9, 0xef, 0x6d, 0x30, 0x4c, 0x40, 0x4f, 0x6b, 0x59, 0xa9, 0xe5, 0x50, 0xaf, 0x4a, 0x89,
				0xdf, 0x37, 0x36, 0x68, 0x19, 0x60, 0x09, 0xc7, 0xe2, 0x80, 0x6c, 0xe3, 0xf1, 0xbd, 0x85, 0x01,
				0x21, 0x02, 0xd8, 0xbe, 0xd3, 0x94, 0xcf, 0xa4, 0x2e, 0x5e, 0x70, 0x7e, 0x5c, 0x67, 0x5a, 0xe5,
				0x09, 0x86, 0x0c, 0xaa, 0xa0, 0x72, 0x83, 0x9b, 0x31, 0x49, 0x81, 0x83, 0x91, 0x8d, 0xa5, 0x1c,
				0xdc, 0x2b, 0xfe, 0xff, 0xff, 0xff, 0x02, 0x54, 0x7c, 0x39, 0x02, 0x00, 0x00, 0x00, 0x00, 0x19,
				0x76, 0xa9, 0x14, 0xfd, 0x68, 0xce, 0xaf, 0x03, 0xf0, 0xce, 0x11, 0xa6, 0x46, 0x5e, 0xf2, 0x5b,
				0x21, 0x99, 0x8c, 0xd3, 0x91, 0x8c, 0xfd, 0x88, 0xac, 0x00, 0x65, 0xcd, 0x1d, 0x00, 0x00, 0x00,
				0x00, 0x19, 0x76, 0xa9, 0x14, 0x43, 0x3d, 0xd2, 0x70, 0x29, 0xc3, 0x44, 0x99, 0x17, 0x63, 0x89,
				0x67, 0xac, 0x55, 0xb1, 0x96, 0x59, 0xea, 0x69, 0x2a, 0x88, 0xac, 0x37, 0x5b, 0x07, 0x00,
			},
		},
	}

	for i, tt := range tests {
		builder, err := NewProgPowBuilder(tt.version, tt.nTime, tt.height, tt.bits, tt.prevHash, tt.txHashes, tt.txHexes)
		if err != nil {
			t.Errorf("failed on %d: NewProgPowBuilder: %v", i, err)
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

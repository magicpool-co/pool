package bech32

import (
	"bytes"
	"testing"

	"github.com/magicpool-co/pool/pkg/common"
)

const (
	cfxCharset = "abcdefghjkmnprstuvwxyz0123456789"
	cfxVersion = 0

	kasCharset           = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	kasPubKeyAddrID      = 0x00
	kasPubKeyECDSAAddrID = 0x01
	kasScriptHashAddrID  = 0x08
)

func TestEncodeDecodeBCH(t *testing.T) {
	tests := []struct {
		charset string
		prefix  string
		version byte
		body    []byte
		encoded string
	}{
		// CFX
		{
			charset: cfxCharset,
			prefix:  "cfx",
			version: cfxVersion,
			body:    common.MustParseHex("1e16a24b714a8f29f805d832fbcf5ee82a108493"),
			encoded: "cfx:aatbrjwnsffj8mt2a1pdf88tn5ycyeeewpzeb10ka3",
		},
		{
			charset: cfxCharset,
			prefix:  "cfxtest",
			version: cfxVersion,
			body:    common.MustParseHex("1e16a24b714a8f29f805d832fbcf5ee82a108493"),
			encoded: "cfxtest:aatbrjwnsffj8mt2a1pdf88tn5ycyeeewp9twhudex",
		},
		{
			charset: cfxCharset,
			prefix:  "cfx",
			version: cfxVersion,
			body:    common.MustParseHex("85d80245dc02f5a89589e1f19c5c718e405b56cd"),
			encoded: "cfx:acc7uawf5ubtnmezvhu9dhc6sghea0403y2dgpyfjp",
		},
		{
			charset: cfxCharset,
			prefix:  "cfxtest",
			version: cfxVersion,
			body:    common.MustParseHex("85d80245dc02f5a89589e1f19c5c718e405b56cd"),
			encoded: "cfxtest:acc7uawf5ubtnmezvhu9dhc6sghea0403ywjz6wtpg",
		},
		{
			charset: cfxCharset,
			prefix:  "cfx",
			version: cfxVersion,
			body:    common.MustParseHex("1a2f80341409639ea6a35bbcab8299066109aa55"),
			encoded: "cfx:aarc9abycue0hhzgyrr53m6cxedgccrmmyybjgh4xg",
		},
		{
			charset: cfxCharset,
			prefix:  "cfxtest",
			body:    common.MustParseHex("1a2f80341409639ea6a35bbcab8299066109aa55"),
			encoded: "cfxtest:aarc9abycue0hhzgyrr53m6cxedgccrmmy8m50bu1p",
		},
		{
			charset: cfxCharset,
			prefix:  "cfx",
			version: cfxVersion,
			body:    common.MustParseHex("19c742cec42b9e4eff3b84cdedcde2f58a36f44f"),
			encoded: "cfx:aap6su0s2uz36x19hscp55sr6n42yr1yk6r2rx2eh7",
		},
		{
			charset: cfxCharset,
			prefix:  "cfxtest",
			body:    common.MustParseHex("19c742cec42b9e4eff3b84cdedcde2f58a36f44f"),
			encoded: "cfxtest:aap6su0s2uz36x19hscp55sr6n42yr1yk6hx8d8sd1",
		},
		{
			charset: cfxCharset,
			prefix:  "cfx",
			version: cfxVersion,
			body:    common.MustParseHex("84980a94d94f54ac335109393c08c866a21b1b0e"),
			encoded: "cfx:acckucyy5fhzknbxmeexwtaj3bxmeg25b2b50pta6v",
		},
		{
			charset: cfxCharset,
			prefix:  "cfxtest",
			body:    common.MustParseHex("84980a94d94f54ac335109393c08c866a21b1b0e"),
			encoded: "cfxtest:acckucyy5fhzknbxmeexwtaj3bxmeg25b2nuf6km25",
		},
		{
			charset: cfxCharset,
			prefix:  "cfx",
			version: cfxVersion,
			body:    common.MustParseHex("1cdf3969a428a750b89b33cf93c96560e2bd17d1"),
			encoded: "cfx:aasr8snkyuymsyf2xp369e8kpzusftj14ec1n0vxj1",
		},
		{
			charset: cfxCharset,
			prefix:  "cfxtest",
			version: cfxVersion,
			body:    common.MustParseHex("1cdf3969a428a750b89b33cf93c96560e2bd17d1"),
			encoded: "cfxtest:aasr8snkyuymsyf2xp369e8kpzusftj14ej62g13p7",
		},
		{
			charset: cfxCharset,
			prefix:  "cfx",
			version: cfxVersion,
			body:    common.MustParseHex("0888000000000000000000000000000000000002"),
			encoded: "cfx:aaejuaaaaaaaaaaaaaaaaaaaaaaaaaaaajrwuc9jnb",
		},
		{
			charset: cfxCharset,
			prefix:  "cfxtest",
			version: cfxVersion,
			body:    common.MustParseHex("0888000000000000000000000000000000000002"),
			encoded: "cfxtest:aaejuaaaaaaaaaaaaaaaaaaaaaaaaaaaajh3dw3ctn",
		},

		// KAS
		{
			charset: kasCharset,
			prefix:  "a",
			version: 0,
			body:    []byte(""),
			encoded: "a:qqeq69uvrh",
		},
		{
			charset: kasCharset,
			prefix:  "a",
			version: 8,
			body:    []byte(""),
			encoded: "a:pq99546ray",
		},
		{
			charset: kasCharset,
			prefix:  "a",
			version: 120,
			body:    []byte(""),
			encoded: "a:0qf6jrhtdq",
		},
		{
			charset: kasCharset,
			prefix:  "b",
			version: 8,
			body:    []byte(" "),
			encoded: "b:pqsqzsjd64fv",
		},
		{
			charset: kasCharset,
			prefix:  "b",
			version: 8,
			body:    []byte("-"),
			encoded: "b:pqksmhczf8ud",
		},
		{
			charset: kasCharset,
			prefix:  "b",
			version: 8,
			body:    []byte("0"),
			encoded: "b:pqcq53eqrk0e",
		},
		{
			charset: kasCharset,
			prefix:  "b",
			body:    []byte("1"),
			version: 8,
			encoded: "b:pqcshg75y0vf",
		},
		{
			charset: kasCharset,
			prefix:  "b",
			version: 8,
			body:    []byte("-1"),
			encoded: "b:pqknzl4e9y0zy",
		},
		{
			charset: kasCharset,
			prefix:  "b",
			version: 8,
			body:    []byte("11"),
			encoded: "b:pqcnzt888ytdg",
		},
		{
			charset: kasCharset,
			prefix:  "b",
			version: 8,
			body:    []byte("abc"),
			encoded: "b:ppskycc8txxxn2w",
		},
		{
			charset: kasCharset,
			prefix:  "b",
			version: 8,
			body:    []byte("1234598760"),
			encoded: "b:pqcnyve5x5unsdekxqeusxeyu2",
		},
		{
			charset: kasCharset,
			prefix:  "b",
			version: 8,
			body:    []byte("abcdefghijklmnopqrstuvwxyz"),
			encoded: "b:ppskycmyv4nxw6rfdf4kcmtwdac8zunnw36hvamc09aqtpppz8lk",
		},
		{
			charset: kasCharset,
			prefix:  "b",
			version: 8,
			body:    []byte("000000000000000000000000000000000000000000"),
			encoded: "b:pqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrq7ag684l3",
		},
		{
			charset: kasCharset,
			prefix:  "kaspa",
			version: kasPubKeyAddrID,
			body: []byte{
				0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27,
				0x3e, 0xfc, 0xc5, 0x4c, 0xe7, 0xd2, 0xa4, 0x91,
				0xbb, 0x4a, 0x0e, 0x84, 0xe3, 0x4c, 0xce, 0x70,
				0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc, 0xe3, 0x4c,
			},
			encoded: "kaspa:qr35ennsep3hxfe7lnz5ee7j5jgmkjswsn35ennsep3hxfe7ln35cdv0dy335",
		},
		{
			charset: kasCharset,
			prefix:  "kaspa",
			version: kasPubKeyAddrID,
			body: []byte{
				0x0e, 0xf0, 0x30, 0x10, 0x7f, 0xd2, 0x6e, 0x0b,
				0x6b, 0xf4, 0x05, 0x12, 0xbc, 0xa2, 0xce, 0xb1,
				0xdd, 0x80, 0xad, 0xaa, 0xe3, 0x4c, 0xce, 0x70,
				0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc, 0xe3, 0x4c,
			},
			encoded: "kaspa:qq80qvqs0lfxuzmt7sz3909ze6camq9d4t35ennsep3hxfe7ln35cvfqgz3z8",
		},
		{
			charset: kasCharset,
			prefix:  "kaspatest",
			version: kasPubKeyAddrID,
			body: []byte{
				0x78, 0xb3, 0x16, 0xa0, 0x86, 0x47, 0xd5, 0xb7,
				0x72, 0x83, 0xe5, 0x12, 0xd3, 0x60, 0x3f, 0x1f,
				0x1c, 0x8d, 0xe6, 0x8f, 0xe3, 0x4c, 0xce, 0x70,
				0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc, 0xe3, 0x4c,
			},
			encoded: "kaspatest:qputx94qseratdmjs0j395mq8u03er0x3l35ennsep3hxfe7ln35ckquw528z",
		},
		{
			charset: kasCharset,
			prefix:  "kaspa",
			version: kasPubKeyECDSAAddrID,
			body: []byte{
				0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27,
				0x3e, 0xfc, 0xc5, 0x4c, 0xe7, 0xd2, 0xa4, 0x91,
				0xbb, 0x4a, 0x0e, 0x84, 0xe3, 0x4c, 0xce, 0x70,
				0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc, 0xe3, 0x4c,
				0xaa,
			},
			encoded: "kaspa:q835ennsep3hxfe7lnz5ee7j5jgmkjswsn35ennsep3hxfe7ln35e2sm7yrlr4w",
		},
		{
			charset: kasCharset,
			prefix:  "kaspa",
			version: kasScriptHashAddrID,
			body: []byte{
				0xc0, 0xa7, 0x82, 0xa0, 0x69, 0x79, 0xf1, 0xbe,
				0xb5, 0xc7, 0x78, 0x42, 0x15, 0xcb, 0x0e, 0x78,
				0x10, 0x06, 0x4c, 0xd7, 0x83, 0x35, 0x30, 0x98,
				0xd7, 0x71, 0xf2, 0x78, 0xc0, 0xb0, 0xe4, 0xd1,
			},
			encoded: "kaspa:prq20q4qd9ulr044cauyy9wtpeupqpjv67pn2vyc6acly7xqkrjdzmh8rj9f4",
		},
		{
			charset: kasCharset,
			prefix:  "kaspa",
			version: kasScriptHashAddrID,
			body: []byte{
				0xe8, 0xc3, 0x00, 0xc8, 0x79, 0x86, 0xef, 0xa8,
				0x4c, 0x37, 0xc0, 0x51, 0x99, 0x29, 0x01, 0x9e,
				0xf8, 0x6e, 0xb5, 0xb4, 0xe8, 0xc3, 0x00, 0xc8,
				0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37, 0xe8, 0xc3,
			},
			encoded: "kaspa:pr5vxqxg0xrwl2zvxlq9rxffqx00sm44kn5vxqxg0xrwl2zvxl5vxyhvsake2",
		},
		{
			charset: kasCharset,
			prefix:  "kaspatest",
			version: kasScriptHashAddrID,
			body: []byte{
				0xc5, 0x79, 0x34, 0x2c, 0x2c, 0x4c, 0x92, 0x20,
				0x20, 0x5e, 0x2c, 0xdc, 0x28, 0x56, 0x17, 0x04,
				0x0c, 0x92, 0x4a, 0x0a, 0xe8, 0xc3, 0x00, 0xc8,
				0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37, 0xe8, 0xc3,
			},
			encoded: "kaspatest:przhjdpv93xfygpqtckdc2zkzuzqeyj2pt5vxqxg0xrwl2zvxl5vx35yyy2h9",
		},
	}

	for i, tt := range tests {
		encoded, err := EncodeBCH(tt.charset, tt.prefix, tt.version, tt.body)
		if err != nil {
			t.Errorf("failed on %d: encode: %v", i, err)
		} else if encoded != tt.encoded {
			t.Errorf("failed on %d: encode: encoded mismatch: have %s, want %s", i, encoded, tt.encoded)
		}

		prefix, version, body, err := DecodeBCH(tt.charset, tt.encoded)
		if err != nil {
			t.Errorf("failed on %d: decode: %v", i, err)
		} else if prefix != tt.prefix {
			t.Errorf("failed on %d: decode: prefix mismatch: have %s, want %s", i, prefix, tt.prefix)
		} else if version != tt.version {
			t.Errorf("failed on %d: decode: version mismatch: have 0x%x, want 0x%x", i, version, tt.version)
		} else if bytes.Compare(body, tt.body) != 0 {
			t.Errorf("failed on %d: decode: body mismatch: have %x, want %x", i, body, tt.body)
		}
	}
}

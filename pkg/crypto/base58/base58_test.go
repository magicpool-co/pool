/* btckeygenie v1.0.0
 * https://github.com/vsergeev/btckeygenie
 * License: MIT
 */

package base58

import (
	"bytes"
	"testing"
)

func TestBase58EncodeDecode(t *testing.T) {
	tests := []struct {
		decoded []byte
		encoded string
		valid   bool
	}{
		{
			decoded: []byte{0x4e, 0x19},
			encoded: "6wi",
			valid:   true,
		},
		{
			decoded: []byte{0x3a, 0xb7},
			encoded: "5UA",
			valid:   true,
		},
		{
			decoded: []byte{0xae, 0x0d, 0xdc, 0x9b},
			encoded: "5T3W5p",
			valid:   true,
		},
		{
			decoded: []byte{0x65, 0xe0, 0xb4, 0xc9},
			encoded: "3c3E6L",
			valid:   true,
		},
		{
			decoded: []byte{0x25, 0x79, 0x36, 0x86, 0xe9, 0xf2, 0x5b, 0x6b},
			encoded: "7GYJp3ZThFG",
			valid:   true,
		},
		{
			decoded: []byte{0x94, 0xb9, 0xac, 0x08, 0x4a, 0x0d, 0x65, 0xf5},
			encoded: "RspedB5CMo2",
			valid:   true,
		},
		{
			encoded: "5T3IW5p", // Invalid character I
			valid:   false,
		},
		{
			encoded: "6Owi", // Invalid character O
			valid:   false,
		},
	}

	for i, tt := range tests {
		if len(tt.decoded) > 0 {
			encoded := Encode(tt.decoded)
			if encoded != tt.encoded {
				t.Errorf("failed on %d: encoded mismatch: have %s, want %s", i, encoded, tt.encoded)
			}
		}

		decoded, err := Decode(tt.encoded)
		if err != nil {
			if tt.valid {
				t.Errorf("failed on %d: decode: %v", i, err)
			}
		} else if !bytes.Equal(decoded, tt.decoded) {
			t.Errorf("failed on %d: decoded mismatch: have %x, want %x", i, decoded, tt.decoded)
		}
	}
}

func TestBase58CheckEncodeDecode(t *testing.T) {
	tests := []struct {
		version []byte
		decoded []byte
		encoded string
		valid   bool
	}{
		{
			version: []byte{0x00},
			decoded: []byte{
				0x01, 0x09, 0x66, 0x77, 0x60, 0x06, 0x95, 0x3d,
				0x55, 0x67, 0x43, 0x9e, 0x5e, 0x39, 0xf8, 0x6a,
				0x0d, 0x27, 0x3b, 0xee,
			},
			encoded: "16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM",
			valid:   true,
		},
		{
			version: []byte{0x00},
			decoded: []byte{
				0x00, 0x00, 0x00, 0x00, 0x60, 0x06, 0x95, 0x3d,
				0x55, 0x67, 0x43, 0x9e, 0x5e, 0x39, 0xf8, 0x6a,
				0x0d, 0x27, 0x3b, 0xee,
			},
			encoded: "111112LbMksD9tCRVsyW67atmDssDkHHG",
			valid:   true,
		},
		{
			version: []byte{0x80},
			decoded: []byte{
				0x0c, 0x28, 0xfc, 0xa3, 0x86, 0xc7, 0xa2, 0x27,
				0x60, 0x0b, 0x2f, 0xe5, 0x0b, 0x7c, 0xae, 0x11,
				0xec, 0x86, 0xd3, 0xbf,
			},
			encoded: "tXNWF26KmH2nSm8LudobRzMF4ggVkfTjff",
			valid:   true,
		},
		{
			encoded: "5T3IW5p",
			valid:   false,
		},
		{
			encoded: "6wi",
			valid:   false,
		},
		{
			encoded: "6UwLL9Risc3QfPqBUvKofHmBQ7wMtjzm",
			valid:   false,
		},
	}

	for i, tt := range tests {
		if len(tt.decoded) > 0 {
			encoded := CheckEncode(tt.version, tt.decoded)
			if encoded != tt.encoded {
				t.Errorf("failed on %d: encoded mismatch: have %s, want %s", i, encoded, tt.encoded)
			}
		}

		version, decoded, err := CheckDecode(tt.encoded)
		if err != nil {
			if tt.valid {
				t.Errorf("failed on %d: decode: %v", i, err)
			}
		} else if !bytes.Equal(version, tt.version) {
			t.Errorf("failed on %d: version mismatch: have %x, want %x", i, version, tt.version)
		} else if !bytes.Equal(decoded, tt.decoded) {
			t.Errorf("failed on %d: decoded mismatch: have %x, want %x", i, decoded, tt.decoded)
		}
	}
}

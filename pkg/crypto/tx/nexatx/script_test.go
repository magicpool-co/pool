package nexatx

import (
	"bytes"
	"testing"

	"github.com/magicpool-co/pool/pkg/common"
)

func TestAddressToScript(t *testing.T) {
	tests := []struct {
		addr         string
		prefix       string
		version      uint8
		scriptPubKey []byte
	}{
		{
			addr:         "nexa:qqusg34nkr7tnupmz8wragk5rrcenkgn7cn42h8p38",
			prefix:       "nexa",
			version:      0,
			scriptPubKey: common.MustParseHex("76a914390446b3b0fcb9f03b11dc3ea2d418f199d913f688ac"),
		},
		{
			addr:         "nexa:nqtsq5g5kzgzgwd3qnunzj8ufuhu93mshclse3y9zeve0fg0",
			prefix:       "nexa",
			version:      1,
			scriptPubKey: common.MustParseHex("005114b0902439b104f93148fc4f2fc2c770be3f0cc485"),
		},
		{
			addr:         "nexa:nqtsq5g5cu4s27u72al62eylgw7pjul9psxxmkhl9lt0lygh",
			prefix:       "nexa",
			version:      1,
			scriptPubKey: common.MustParseHex("005114c72b057b9e577fa5649f43bc1973e50c0c6ddaff"),
		},
		{
			addr:         "nexatest:qzrpqsursz4dprxly2zfrvh8qgu64zvfe56x2f5t62",
			prefix:       "nexatest",
			version:      0,
			scriptPubKey: common.MustParseHex("76a9148610438380aad08cdf228491b2e70239aa8989cd88ac"),
		},
		{
			addr:         "nexatest:nqtsq5g5vjp8vv50e6kdyp3723hn23tvkgn0qj6kj0q7l32p",
			prefix:       "nexatest",
			version:      1,
			scriptPubKey: common.MustParseHex("005114648276328fceacd2063e546f35456cb226f04b56"),
		},
		{
			addr:         "nexatest:nqtsq5g53h8gcg58sdmtsyxh42sdg3c5ar6lu0f3hl0720kx",
			prefix:       "nexatest",
			version:      1,
			scriptPubKey: common.MustParseHex("0051148dce8c22878376b810d7aaa0d44714e8f5fe3d31"),
		},
	}

	for i, tt := range tests {
		version, scriptPubKey, err := AddressToScript(tt.addr, tt.prefix)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if version != tt.version {
			t.Errorf("failed on %d: version: have %d, want %d", i, version, tt.version)
		} else if bytes.Compare(scriptPubKey, tt.scriptPubKey) != 0 {
			t.Errorf("failed on %d: scriptPubKey: have %x, want %x", i, scriptPubKey, tt.scriptPubKey)
		}
	}
}

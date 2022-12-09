package kastx

import (
	"bytes"
	"testing"

	"github.com/magicpool-co/pool/pkg/common"
)

func TestAddressToScript(t *testing.T) {
	tests := []struct {
		addr         string
		prefix       string
		scriptPubKey []byte
	}{
		{
			addr:         "kaspa:qqg0m2cq9nls4s2mlj9estz6v947l8j6vmcvv4057clh688088ftg7ce6p895",
			prefix:       "kaspa",
			scriptPubKey: common.MustParseHex("2010fdab002cff0ac15bfc8b982c5a616bef9e5a66f0c655f4f63f7d1cef39d2b4ac"),
		},
		{
			addr:         "kaspa:qpy827u4r43hp36nu2w78dphwgzjr3e9xdwwvm7k7dalyhpfkr84qucn4ecud",
			prefix:       "kaspa",
			scriptPubKey: common.MustParseHex("2048757b951d6370c753e29de3b437720521c725335ce66fd6f37bf25c29b0cf50ac"),
		},
		{
			addr:         "kaspatest:qq4fnmnql7ffuw2nuhxh6pxqva4c8n8wxt8df5uwwyr38n3wkpzyzmnlwha8g",
			prefix:       "kaspatest",
			scriptPubKey: common.MustParseHex("202a99ee60ff929e3953e5cd7d04c0676b83ccee32ced4d38e710713ce2eb04441ac"),
		},
		{
			addr:         "kaspatest:qyp55var4ed9sjqy6x52qp8xmpnpcewn3vstqxxxvl43kzl9q9qhwpgyfy07x9e",
			prefix:       "kaspatest",
			scriptPubKey: common.MustParseHex("21034a33a3ae5a584804d1a8a004e6d8661c65d38b20b018c667eb1b0be501417705ab"),
		},
		{
			addr:         "kaspasim:qzpj2cfa9m40w9m2cmr8pvfuqpp32mzzwsuw6ukhfduqpp32mzzws59e8fapc",
			prefix:       "kaspasim",
			scriptPubKey: common.MustParseHex("208325613d2eeaf7176ac6c670b13c0043156c427438ed72d74b7800862ad884e8ac"),
		},
		{
			addr:         "kaspasim:qr7w7nqsdnc3zddm6u8s9fex4ysk95hm3v30q353ymuqpp32mzzws59e8fapc",
			prefix:       "kaspasim",
			scriptPubKey: common.MustParseHex("20fcef4c106cf11135bbd70f02a726a92162d2fb8b22f0469126f800862ad884e8ac"),
		},
	}

	for i, tt := range tests {
		scriptPubKey, err := AddressToScript(tt.addr, tt.prefix)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if bytes.Compare(scriptPubKey, tt.scriptPubKey) != 0 {
			t.Errorf("failed on %d: have %x, want %x", i, scriptPubKey, tt.scriptPubKey)
		}
	}
}

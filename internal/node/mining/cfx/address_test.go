package cfx

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestETHAddressToCFX(t *testing.T) {
	tests := []struct {
		ethAddress    string
		networkPrefix string
		cfxAddress    string
	}{
		{
			ethAddress:    "0x1e16a24b714a8f29f805d832fbcf5ee82a108493",
			networkPrefix: "cfx",
			cfxAddress:    "cfx:aatbrjwnsffj8mt2a1pdf88tn5ycyeeewpzeb10ka3",
		},
		{
			ethAddress:    "0x1e16a24b714a8f29f805d832fbcf5ee82a108493",
			networkPrefix: "cfxtest",
			cfxAddress:    "cfxtest:aatbrjwnsffj8mt2a1pdf88tn5ycyeeewp9twhudex",
		},
		{
			ethAddress:    "0x85d80245dc02f5a89589e1f19c5c718e405b56cd",
			networkPrefix: "cfx",
			cfxAddress:    "cfx:acc7uawf5ubtnmezvhu9dhc6sghea0403y2dgpyfjp",
		},
		{
			ethAddress:    "0x85d80245dc02f5a89589e1f19c5c718e405b56cd",
			networkPrefix: "cfxtest",
			cfxAddress:    "cfxtest:acc7uawf5ubtnmezvhu9dhc6sghea0403ywjz6wtpg",
		},
		{
			ethAddress:    "0x1a2f80341409639ea6a35bbcab8299066109aa55",
			networkPrefix: "cfx",
			cfxAddress:    "cfx:aarc9abycue0hhzgyrr53m6cxedgccrmmyybjgh4xg",
		},
		{
			ethAddress:    "0x1a2f80341409639ea6a35bbcab8299066109aa55",
			networkPrefix: "cfxtest",
			cfxAddress:    "cfxtest:aarc9abycue0hhzgyrr53m6cxedgccrmmy8m50bu1p",
		},
		{
			ethAddress:    "0x19c742cec42b9e4eff3b84cdedcde2f58a36f44f",
			networkPrefix: "cfx",
			cfxAddress:    "cfx:aap6su0s2uz36x19hscp55sr6n42yr1yk6r2rx2eh7",
		},
		{
			ethAddress:    "0x19c742cec42b9e4eff3b84cdedcde2f58a36f44f",
			networkPrefix: "cfxtest",
			cfxAddress:    "cfxtest:aap6su0s2uz36x19hscp55sr6n42yr1yk6hx8d8sd1",
		},
		{
			ethAddress:    "0x84980a94d94f54ac335109393c08c866a21b1b0e",
			networkPrefix: "cfx",
			cfxAddress:    "cfx:acckucyy5fhzknbxmeexwtaj3bxmeg25b2b50pta6v",
		},
		{
			ethAddress:    "0x84980a94d94f54ac335109393c08c866a21b1b0e",
			networkPrefix: "cfxtest",
			cfxAddress:    "cfxtest:acckucyy5fhzknbxmeexwtaj3bxmeg25b2nuf6km25",
		},
		{
			ethAddress:    "0x1cdf3969a428a750b89b33cf93c96560e2bd17d1",
			networkPrefix: "cfx",
			cfxAddress:    "cfx:aasr8snkyuymsyf2xp369e8kpzusftj14ec1n0vxj1",
		},
		{
			ethAddress:    "0x1cdf3969a428a750b89b33cf93c96560e2bd17d1",
			networkPrefix: "cfxtest",
			cfxAddress:    "cfxtest:aasr8snkyuymsyf2xp369e8kpzusftj14ej62g13p7",
		},
		{
			ethAddress:    "0x0888000000000000000000000000000000000002",
			networkPrefix: "cfx",
			cfxAddress:    "cfx:aaejuaaaaaaaaaaaaaaaaaaaaaaaaaaaajrwuc9jnb",
		},
		{
			ethAddress:    "0x0888000000000000000000000000000000000002",
			networkPrefix: "cfxtest",
			cfxAddress:    "cfxtest:aaejuaaaaaaaaaaaaaaaaaaaaaaaaaaaajh3dw3ctn",
		},
	}

	for i, tt := range tests {
		ethAddress, err := hex.DecodeString(strings.ReplaceAll(tt.ethAddress, "0x", ""))
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
			continue
		}

		cfxAddress, err := ETHAddressToCFX(ethAddress, tt.networkPrefix)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if cfxAddress != tt.cfxAddress {
			t.Errorf("failed on %d: cfx address mismatch: have %s, want %s", i, cfxAddress, tt.cfxAddress)
		}
	}
}

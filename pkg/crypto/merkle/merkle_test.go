package merkle

import (
	"encoding/hex"
	"testing"
)

func TestCalculateRoot(t *testing.T) {
	tests := []struct {
		hashes     []string
		merkleRoot string
	}{
		{
			hashes: []string{
				"fb522b04ef3088b6da01b0d1b7b873fc5141bc62ae2e3628eb4dc505766e37dd",
			},
			merkleRoot: "fb522b04ef3088b6da01b0d1b7b873fc5141bc62ae2e3628eb4dc505766e37dd",
		},
		{
			hashes: []string{
				"baf184e90cc2ed9168914e904942e74936b7513e296c910bd03efc62b0f497b5",
				"d1cf29f18031da7d68ade484ddbf15fee12222a33c53b013cf0bbb10384692a7",
				"99d9546e11eb24bc6af553d38f7e0bddbfccee61f600af3e629f1ea9589b26f3",
				"5a9475a91e1a13edae487e74f73ca1b00cdbc2cbff37398489f5c7b600fa636b",
				"37bef58f7c89e53a0b546d7d526628657a9ce7bf2a086044570b3a506745f273",
				"e9a6b6297d084179ae496b6b308f6068c31b7bad4704449bae0c4d14194223cd",
				"656ebb3483bd16451f11a171ce05be1652283b51cd347068392b5f1d2b3b1e49",
			},
			merkleRoot: "1f6795e0c01f66ce452eaea94fd312aa90d62244f9cfb2804bcf93371710ed2e",
		},
		{
			hashes: []string{
				"b86fdc10feb62a76d78bacc922d0cd72cace8056bda9a9956484a9868a6999b2",
				"61530c725c43f0057986b07ea7588f6c6389817358c6f9869eecfb278934b5ca",
				"25653bf2beeae1cf02ed9f3a29b210e66b6b60d25e605793ef206dde0509e3ce",
				"867a0b05c77554b19058022cf157de7b9d82a2453139c74fcd2565000d1ce309",
				"19ec5070662c5f703d8b7128a3911a6d98d43265de0cd4ae82730dfac24741c1",
				"89da5e2e67021fb3e8abfd01c486545227b068e5190b71190ac70863fd820b5b",
				"fbaf5f1dc54e8e02b8c6afccc2c31e201982c92c91f7d6e87129fc99f242850a",
				"cbd1d7a95f618f00f59a8e14e4e49a508f52dfa23789048d249fed38f7f5715f",
				"156dd2f7ca73f8da9e9726b1353747827b9da0106b27c1f71ea8c097b9006f82",
				"d9ea24c85d9632a17a2c0e650bfe4fa09bb2d6e280887492181e72c64a6af93b",
				"e202d8ce30c613e03de94f780e14100b29bf0a4f4cd7eba14a512ab91f1eb4e5",
				"ab60da5a41b3085a894a172a26725cbf0c68ca199184cd13765c1ea222a5b09a",
				"ffa72c24f1b3e5c1cbe1db6bf5a2908b2bd1efac91d4f5cb917c8be0d2ce9ee2",
				"5fb22b6aa42a5f26a59b6237dbe23300245d2de2688b50ca9f6062c5059aad16",
				"40cd74dfd4d273135c5894f6ad2182f79b16b2e33dc0d3f40593760503c0cc36",
				"04e09d2ac435503e877e661d8dd36a14834a5a8d63b5cd1ff5b3809e1d611f74",
				"e82fef26275c64af8cba2ff83bf9c9b3ab3f0430b584063bdac6b1f80fc370f5",
				"5dbe4739410250402fd6872e1734291dbd6b7e57d60a28bfded0752f8504f27c",
				"d4a1bf48f873873c5e8c9fd8877efd7bacc3a7f751d40b83f097cb940388f9fa",
				"65b90cdffc163de746ff8f302c70c3b44d45cfa86c6418b74cca5b83fbc3764c",
				"25e092cec073aae67cf103e5e27aa8575f47d2a8adaad38df61d6e7a4fe17ad4",
				"f45fc99f8ab2d42e482163e32443b80b7093d1cd9fcf0eb2676400884a4d1e57",
				"69d5e365bd84a9bbd54e58fdfe21fa233fc9199de9804d3a71642574b2ad9e59",
				"0db79a5770962dddb7e6ba5e5fca1828281832f26377867b7233299364823fbe",
				"3e4b758b8046d01b0a2bf078d140a0f45464017178dd18235253c146bac5c6e4",
				"b6c12f4145759fd7c047e504de46201d5e2494ab2dc306d56a57622a9720e511",
			},
			merkleRoot: "327d9063184e0161db7a77cbfd5816e44f9353a9b045f8293c338a3187fcd0ec",
		},
	}

	for i, tt := range tests {
		hashesBytes := make([][]byte, len(tt.hashes))
		for i, hash := range tt.hashes {
			var err error
			hashesBytes[i], err = hex.DecodeString(hash)
			if err != nil {
				t.Errorf("failed on %d: cannot decode hash: %v", i, err)
				return
			}
		}

		merkleRoot := CalculateRoot(hashesBytes)
		if hex.EncodeToString(merkleRoot) != tt.merkleRoot {
			t.Errorf("failed on %d: have %x, want %s", i, merkleRoot, tt.merkleRoot)
			return
		}
	}
}

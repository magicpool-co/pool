package blkbuilder

import (
	"bytes"
	"testing"
)

func TestSerializeKaspaHeader(t *testing.T) {
	tests := []struct {
		version              uint16
		parents              [][]string
		hashMerkleRoot       string
		acceptedIDMerkleRoot string
		utxoCommitment       string
		timestamp            int64
		bits                 uint32
		nonce                uint64
		daaScore             uint64
		blueScore            uint64
		blueWork             string
		pruningPoint         string
		initialHeader        []byte
		finalHeader          []byte
	}{
		{
			version: 1,
			parents: [][]string{
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"bafee3d9fb38f13784b3910964c4b469621a9a9128d67c034e586f558304e68e"},
				[]string{"6a81a712e2269bfa6c765bb6786096f475a268889be7ff3268b510690ec9a4cc"},
				[]string{"c9494053ec7dc5e5b42d13e2053bc294444e960bfa940c93567e31ac34369d32"},
				[]string{"2e5be6575ae28912a4b20e3f2c64db019c173a4aa52507d98ee2b7a258013196"},
				[]string{"2e5be6575ae28912a4b20e3f2c64db019c173a4aa52507d98ee2b7a258013196"},
				[]string{"2e5be6575ae28912a4b20e3f2c64db019c173a4aa52507d98ee2b7a258013196"},
				[]string{"cbe1f6a9d98a9c25a9e112f5692c36cb28d5cdd4fcd97effe75b645b09fac42f"},
				[]string{"cbe1f6a9d98a9c25a9e112f5692c36cb28d5cdd4fcd97effe75b645b09fac42f"},
				[]string{"7a748a62ae5c58e08ead9c270baff45a777c8d9ce766c8af40ccc1cc60159aae"},
				[]string{"7a748a62ae5c58e08ead9c270baff45a777c8d9ce766c8af40ccc1cc60159aae"},
				[]string{"b955ccc8ec6fdbfa4de994b942a7c94df2f5f247d1f379d81cc8c8962f073780"},
				[]string{"4b9fb4075a2c9e3d79a9cae9d34231f8119186ac037acd91b929c951d2f0489d"},
				[]string{"4a673c24bb30bb0cf26b752c666286deec108f6482834cd0fa10a09b37d54245"},
				[]string{"4a673c24bb30bb0cf26b752c666286deec108f6482834cd0fa10a09b37d54245"},
				[]string{"4a673c24bb30bb0cf26b752c666286deec108f6482834cd0fa10a09b37d54245"},
				[]string{"a4fab1eaa182069d7e349eb5696817e9ad67b31e1e9913d2b0e37e418aca893b"},
				[]string{"b5096607c01cb42ca73abb432365d4229967c8fa13664e0393db2523b49e8f07"},
				[]string{"b5096607c01cb42ca73abb432365d4229967c8fa13664e0393db2523b49e8f07"},
				[]string{"b5096607c01cb42ca73abb432365d4229967c8fa13664e0393db2523b49e8f07"},
				[]string{"b5096607c01cb42ca73abb432365d4229967c8fa13664e0393db2523b49e8f07"},
				[]string{"b5096607c01cb42ca73abb432365d4229967c8fa13664e0393db2523b49e8f07"},
			},
			hashMerkleRoot:       "3aae9bd437ca151774a04c72df3c2f6f194b5f65f09e53b54969330f080a9f4f",
			acceptedIDMerkleRoot: "103bfb5134c94c420846b4a480982a2a9b466b6cfc6d45b60bc10eccfed3c305",
			utxoCommitment:       "f32424c5aeb8ab1c5c72b547cf8cee55eec9f0633b13878c93611939a0195b96",
			timestamp:            1661062150793,
			bits:                 453325233,
			nonce:                123456789,
			daaScore:             24606947,
			blueScore:            23102453,
			blueWork:             "7b09bfb044de1ae41",
			pruningPoint:         "37f4aeda7e595d2ddf6dabf6d21b4738eaa31cc2191e856c2969edd12bb459e0",
			initialHeader: []byte{
				0x3e, 0x1d, 0x2e, 0x3d, 0xa8, 0x53, 0xf5, 0x02,
				0x32, 0xba, 0x64, 0x0f, 0x7a, 0x99, 0x5f, 0x58,
				0x31, 0xa6, 0xa9, 0xde, 0x61, 0x64, 0xcc, 0xa9,
				0x88, 0x10, 0xb5, 0xfd, 0x13, 0x2b, 0x6b, 0x0a,
			},
			finalHeader: []byte{
				0x3d, 0xce, 0xb0, 0xac, 0x62, 0xe4, 0xc3, 0xb3,
				0x61, 0xcd, 0xed, 0xb7, 0xb8, 0xa0, 0x0c, 0xa8,
				0x86, 0x76, 0xd9, 0xe7, 0xdc, 0x45, 0xef, 0x9a,
				0x1f, 0xe7, 0x32, 0xee, 0x89, 0x35, 0xbb, 0x5e,
			},
		},
		{
			version: 1,
			parents: [][]string{
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{
					"896bd6d52a3b5dcf06002dfc85b812f693658df2a8f571b3e46a016216b1dabe",
					"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692",
				},
				[]string{"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692"},
				[]string{"04f4e4e52f5f74d9c0d2a5c7cde5ac7a7052eaed313f4b2263d7f8432d9f8692"},
				[]string{"0a483ebb08da3d417aa1b9f4194094517672e85d6d6bca538ea623309adafe05"},
				[]string{"0a483ebb08da3d417aa1b9f4194094517672e85d6d6bca538ea623309adafe05"},
				[]string{"d78c8d16436a01c0dc1a3b070594dbaccb24b38f2f7337a4a94707ada2bd686b"},
				[]string{"d78c8d16436a01c0dc1a3b070594dbaccb24b38f2f7337a4a94707ada2bd686b"},
				[]string{"c8e703eedf2b1fba28c27ab5d6130b59e79d57e99bb748b12d24d33cee3c27a9"},
				[]string{"c8e703eedf2b1fba28c27ab5d6130b59e79d57e99bb748b12d24d33cee3c27a9"},
				[]string{"185633b428b93f08ac166809455afb6324251a6959fb1a2587a6eef56f937c95"},
				[]string{"c991b2f4bb47a72ebf8518b46b432ee2baf4e7b0c48a83e4dce9fe028a63f8dd"},
				[]string{"c991b2f4bb47a72ebf8518b46b432ee2baf4e7b0c48a83e4dce9fe028a63f8dd"},
				[]string{"632bbcc9db1651eab6c625938fa347fd08053e1669e3910bc87369691faaf1fe"},
				[]string{"02c2a9360e552f1a6b9b98763a0646fcbf6b7efcbc0b0f3d9e5bdb52811f0a40"},
				[]string{"e9b1fc70f1267210c1b89f5248607ce24022be61c621b0ab8acd3b4b5c99c682"},
				[]string{"e9b1fc70f1267210c1b89f5248607ce24022be61c621b0ab8acd3b4b5c99c682"},
				[]string{"e9b1fc70f1267210c1b89f5248607ce24022be61c621b0ab8acd3b4b5c99c682"},
				[]string{"e9b1fc70f1267210c1b89f5248607ce24022be61c621b0ab8acd3b4b5c99c682"},
				[]string{"8dbac88b73cba28f5e2f93b6702882a37688821a8e7c2db04cb062f7bb34ba0d"},
				[]string{"8dbac88b73cba28f5e2f93b6702882a37688821a8e7c2db04cb062f7bb34ba0d"},
				[]string{"e7e0916d4b3959e102eaa5341a02c9c37f8de2bf775b2b9ee7d3ac0c63f91e2b"},
				[]string{"e4c39fce1ab255f30b333925e079c94c1b6bdebd1b579f12fd8407756dbd8dd6"},
				[]string{"bfb6b3861f2bd5c3473c0ed53cf3aa1f94a565c9738300d7af2272115dcbaa95"},
				[]string{"82237f6135c67f1b792dd126e3ff66bd9a914b5cd4885aa7c95b35af029fae9a"},
				[]string{"4218eb0a880575eed9ca425e5fc559059a09eb43dc95b71fe304ab16642e173a"},
				[]string{"ae3d8b968db5a4bb8446a8a256d65f35be6180c6e1f03a66cef86be87684fa0b"},
			},
			hashMerkleRoot:       "d89861472cceed0cf1c8735a778fd0aef683c1e81fb0ece7db3695d965eb43e6",
			acceptedIDMerkleRoot: "175834d576e2c7f7333f83dc79b89b3a0ef78da7bb8143e1b0455eda2f13234d",
			utxoCommitment:       "3ba2d3e6e2b207ecd9251ee91869bb5a2ec1e8d717b48f7bb3530f2aee0220c5",
			timestamp:            1668902735647,
			bits:                 505134858,
			nonce:                7428945058704835731,
			daaScore:             19100422,
			blueScore:            19043081,
			blueWork:             "109528c0e51a",
			pruningPoint:         "f562a6cd834ee9eb97c79ae98e2be1b352e5c8fef2f1ec0ab2e9de913b7f8368",
			initialHeader: []byte{
				0x7d, 0xd0, 0xfa, 0x7e, 0xd2, 0x9f, 0xd0, 0xcf,
				0xf6, 0xbd, 0xd9, 0xdf, 0x05, 0xb3, 0x0e, 0x48,
				0xcb, 0xf5, 0x62, 0xf0, 0xe9, 0x0f, 0x87, 0xbc,
				0x5e, 0x98, 0xd9, 0x4e, 0x76, 0xeb, 0xe2, 0xe0,
			},
			finalHeader: []byte{
				0xc5, 0xfb, 0x18, 0x6a, 0x8f, 0x76, 0xa9, 0xfb,
				0xb5, 0x6f, 0x19, 0x02, 0x16, 0x51, 0x39, 0x6c,
				0x36, 0xc4, 0xcb, 0xba, 0x12, 0x18, 0x49, 0x18,
				0xe4, 0x6c, 0x81, 0x43, 0xc5, 0xef, 0xd8, 0xa7,
			},
		},
	}

	for i, tt := range tests {
		initialHeader, err := SerializeKaspaBlockHeader(tt.version, tt.parents, tt.hashMerkleRoot,
			tt.acceptedIDMerkleRoot, tt.utxoCommitment, 0, tt.bits,
			0, tt.daaScore, tt.blueScore, tt.blueWork, tt.pruningPoint)
		if err != nil {
			t.Errorf("failed on %d: initial: %v", i, err)
		} else if bytes.Compare(initialHeader, tt.initialHeader) != 0 {
			t.Errorf("failed on %d: initial header mismatch: have %x, want %x", i, initialHeader, tt.initialHeader)
		}

		finalHeader, err := SerializeKaspaBlockHeader(tt.version, tt.parents, tt.hashMerkleRoot,
			tt.acceptedIDMerkleRoot, tt.utxoCommitment, tt.timestamp, tt.bits,
			tt.nonce, tt.daaScore, tt.blueScore, tt.blueWork, tt.pruningPoint)
		if err != nil {
			t.Errorf("failed on %d: final: %v", i, err)
		} else if bytes.Compare(finalHeader, tt.finalHeader) != 0 {
			t.Errorf("failed on %d: final header mismatch: have %x, want %x", i, finalHeader, tt.finalHeader)
		}
	}
}

package kastx

import (
	"bytes"
	"testing"

	"github.com/magicpool-co/pool/internal/node/mining/kas/protowire"
	"github.com/magicpool-co/pool/pkg/common"
)

var (
	testTx = &protowire.RpcTransaction{
		Version: 0,
		Inputs: []*protowire.RpcTransactionInput{
			&protowire.RpcTransactionInput{
				PreviousOutpoint: &protowire.RpcOutpoint{
					TransactionId: "880eb9819a31821d9d2399e2f35e2433b72637e393d71ecc9b8d0250f49153c3",
					Index:         0,
				},
				Sequence: 0,
			},
			&protowire.RpcTransactionInput{
				PreviousOutpoint: &protowire.RpcOutpoint{
					TransactionId: "880eb9819a31821d9d2399e2f35e2433b72637e393d71ecc9b8d0250f49153c3",
					Index:         1,
				},
				Sequence: 1,
			},
			&protowire.RpcTransactionInput{
				PreviousOutpoint: &protowire.RpcOutpoint{
					TransactionId: "880eb9819a31821d9d2399e2f35e2433b72637e393d71ecc9b8d0250f49153c3",
					Index:         2,
				},
				Sequence: 2,
			},
		},
		Outputs: []*protowire.RpcTransactionOutput{
			&protowire.RpcTransactionOutput{
				Amount: 300,
				ScriptPublicKey: &protowire.RpcScriptPublicKey{
					Version:         0,
					ScriptPublicKey: "20fcef4c106cf11135bbd70f02a726a92162d2fb8b22f0469126f800862ad884e8ac",
				},
			},
			&protowire.RpcTransactionOutput{
				Amount: 300,
				ScriptPublicKey: &protowire.RpcScriptPublicKey{
					Version:         0,
					ScriptPublicKey: "208325613d2eeaf7176ac6c670b13c0043156c427438ed72d74b7800862ad884e8ac",
				},
			},
		},
		LockTime:     1615462089000,
		SubnetworkId: "0000000000000000000000000000000000000000",
	}
)

func TestSerializeFull(t *testing.T) {
	tests := []struct {
		tx         *protowire.RpcTransaction
		serialized []byte
	}{
		{
			tx: &protowire.RpcTransaction{
				Version: 0,
				Outputs: []*protowire.RpcTransactionOutput{
					&protowire.RpcTransactionOutput{
						Amount: 34922893143,
						ScriptPublicKey: &protowire.RpcScriptPublicKey{
							Version:         0,
							ScriptPublicKey: "20fc0c592acf1e509c8fa680d1ee11d492c243f420542129a7d17b02caa0ca5340ac",
						},
					},
					&protowire.RpcTransactionOutput{
						Amount: 34922823143,
						ScriptPublicKey: &protowire.RpcScriptPublicKey{
							Version:         0,
							ScriptPublicKey: "20c62cf30e4e57c5922086460235d5157a62f2666aa22707cb9b588f6211fc0ed6ac",
						},
					},
				},
				LockTime:     0,
				SubnetworkId: "0100000000000000000000000000000000000000",
				Gas:          0,
				Payload:      "6105940100000000e7fd8f210800000000002220e3a134d07b6719befe296576fdea05a14f555f3491b4c13229b7fc77d3aff5b7ac302e31322e372f6575312e6163632d706f6f6c2e7077",
			},
			serialized: common.MustParseHex("9a02063591f93629c01f65ebf5d9ee271a19d631bb001faa686b846e7f3762ce"),
		},
	}

	for i, tt := range tests {
		serialized, err := serializeFull(tt.tx)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if bytes.Compare(serialized, tt.serialized) != 0 {
			t.Errorf("failed on %d: have %x, want %x", i, serialized, tt.serialized)
		}
	}
}

func TestCalculateScriptSig(t *testing.T) {
	tests := []struct {
		tx           *protowire.RpcTransaction
		index        uint32
		amount       uint64
		scriptPubKey []byte
		scriptSig    []byte
	}{
		{
			tx:           testTx,
			index:        0,
			amount:       100,
			scriptPubKey: common.MustParseHex("208325613d2eeaf7176ac6c670b13c0043156c427438ed72d74b7800862ad884e8ac"),
			scriptSig:    common.MustParseHex("1d679268414c20ffe952e3c255befd892e60e86ae1657fce8a20225e5dc87d64"),
		},
		{
			tx:           testTx,
			index:        1,
			amount:       200,
			scriptPubKey: common.MustParseHex("20fcef4c106cf11135bbd70f02a726a92162d2fb8b22f0469126f800862ad884e8ac"),
			scriptSig:    common.MustParseHex("2d87f5eac48ad95b58f9859356679c497cbab64a6f6967d119490471147bacd0"),
		},
		{
			tx:           testTx,
			index:        2,
			amount:       300,
			scriptPubKey: common.MustParseHex("20fcef4c106cf11135bbd70f02a726a92162d2fb8b22f0469126f800862ad884e8ac"),
			scriptSig:    common.MustParseHex("c4269fb253f8801b9c8240f84e8ec2b4db8bdc2ad1fc81fc0e6897b3c5c1f223"),
		},
	}

	for i, tt := range tests {
		scriptSig, err := calculateScriptSig(tt.tx, tt.index, tt.amount, tt.scriptPubKey)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if bytes.Compare(scriptSig, tt.scriptSig) != 0 {
			t.Errorf("failed on %d: have %x, want %x", i, scriptSig, tt.scriptSig)
		}
	}
}

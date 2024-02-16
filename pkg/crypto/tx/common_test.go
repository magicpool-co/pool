package tx

import (
	"math/big"
	"testing"

	"github.com/magicpool-co/pool/types"
)

func TestDistributeFees(t *testing.T) {
	tests := []struct {
		inputs  []*types.TxInput
		outputs []*types.TxOutput
		fee     uint64
		strict  bool
		fees    []*big.Int
		err     error
	}{
		{
			inputs: []*types.TxInput{
				&types.TxInput{
					Value: new(big.Int).SetUint64(0xffffffff),
				},
			},
			outputs: []*types.TxOutput{
				&types.TxOutput{
					Value:    new(big.Int).SetUint64(0xffffffff),
					SplitFee: true,
				},
			},
			fee:    0x512512,
			strict: false,
			fees:   []*big.Int{new(big.Int).SetUint64(0x512512)},
		},
		{
			inputs: []*types.TxInput{
				&types.TxInput{
					Value: new(big.Int).SetUint64(0xffffffff),
				},
			},
			outputs: []*types.TxOutput{
				&types.TxOutput{
					Value:    new(big.Int).SetUint64(0xffffffff),
					SplitFee: true,
				},
			},
			fee:    0x512512,
			strict: true,
			fees:   []*big.Int{new(big.Int).SetUint64(0x512512)},
		},
		{
			inputs: []*types.TxInput{
				&types.TxInput{
					Value: new(big.Int).SetUint64(0xffffffff),
				},
				&types.TxInput{
					Value: new(big.Int).SetUint64(0x4d22991abd),
				},
			},
			outputs: []*types.TxOutput{
				&types.TxOutput{
					Value:    new(big.Int).SetUint64(0xffffffff),
					SplitFee: true,
				},
				&types.TxOutput{
					Value:    new(big.Int).SetUint64(0x3d1f211abd),
					SplitFee: true,
				},
				&types.TxOutput{
					Value:    new(big.Int).SetUint64(0xffa7dd2df),
					SplitFee: false,
				},
				&types.TxOutput{
					Value:    new(big.Int).SetUint64(0x8fa2d21),
					SplitFee: true,
				},
			},
			fee:    0x512512,
			strict: true,
			fees: []*big.Int{
				new(big.Int).SetUint64(0x14e34),
				new(big.Int).SetUint64(0x4fcb25),
				new(big.Int).SetUint64(0),
				new(big.Int).SetUint64(0xbb8),
			},
		},
		{
			inputs:  []*types.TxInput{},
			outputs: []*types.TxOutput{},
			fee:     0x512512,
			strict:  false,
			err:     ErrFeesNotDistributed,
		},
		{
			inputs: []*types.TxInput{
				&types.TxInput{
					Value: new(big.Int).SetUint64(0xffffffff),
				},
			},
			outputs: []*types.TxOutput{
				&types.TxOutput{
					Value:    new(big.Int).SetUint64(0xfffffff1),
					SplitFee: true,
				},
			},
			fee:    0x512512,
			strict: true,
			err:    ErrInputOutputAmountMismatch,
		},
	}

	for i, tt := range tests {
		err := DistributeFees(tt.inputs, tt.outputs, tt.fee, tt.strict)
		if err != tt.err {
			t.Errorf("failed on %d: err mismatch: have %v, want %v", i, err, tt.err)
			continue
		} else if err != nil {
			continue
		}

		for j, output := range tt.outputs {
			if output.Fee.Cmp(tt.fees[j]) != 0 {
				t.Errorf("failed on %d: output %d: fee mismatch: have %s, want %s",
					i, j, output.Fee, tt.fees[j])
			}
		}
	}
}

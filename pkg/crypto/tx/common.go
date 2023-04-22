package tx

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

var (
	ErrInputOutputAmountMismatch = fmt.Errorf("input and output sum mismatch")
	ErrNegativeFeeRemainder      = fmt.Errorf("negative fee remainder")
	ErrFeesNotDistributed        = fmt.Errorf("fees could not be distributed")
	ErrTxTooBig                  = fmt.Errorf("tx too big")
)

func DistributeFees(inputs []*types.TxInput, outputs []*types.TxOutput, fee uint64, strict bool) error {
	var sumInputAmount uint64
	for _, inp := range inputs {
		sumInputAmount += inp.Value.Uint64()
	}

	var sumOutputAmount uint64
	sumSplitAmount := new(big.Int)
	for _, out := range outputs {
		sumOutputAmount += out.Value.Uint64()
		if out.SplitFee {
			sumSplitAmount.Add(sumSplitAmount, out.Value)
		}
	}

	if strict && sumInputAmount != sumOutputAmount {
		return ErrInputOutputAmountMismatch
	}

	usedFees := new(big.Int)
	for _, out := range outputs {
		if !out.SplitFee {
			out.Fee = new(big.Int)
			continue
		}

		out.Fee = new(big.Int).Mul(new(big.Int).SetUint64(fee), out.Value)
		out.Fee.Div(out.Fee, sumSplitAmount)
		out.Value.Sub(out.Value, out.Fee)
		usedFees.Add(usedFees, out.Fee)
	}

	remainder := new(big.Int).Sub(new(big.Int).SetUint64(fee), usedFees)
	if remainder.Cmp(common.Big0) < 0 {
		return ErrNegativeFeeRemainder
	} else if remainder.Cmp(common.Big0) > 0 {
		for _, output := range outputs {
			if !output.SplitFee {
				continue
			} else if output.Value.Cmp(remainder) > 0 {
				output.Value.Sub(output.Value, remainder)
				remainder = new(big.Int)
				break
			}
		}
	}

	if remainder.Cmp(new(big.Int)) > 0 {
		return ErrFeesNotDistributed
	}

	return nil
}

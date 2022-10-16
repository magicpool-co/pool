package ergo

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://explorer.ergoplatform.com/en/transactions/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://explorer.ergoplatform.com/en/addresses/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	balance, err := node.getWalletBalances()
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetUint64(balance.Balance), nil
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	return nil, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, error) {
	if len(outputs) == 0 {
		return "", fmt.Errorf("need at least one output")
	}

	const fee = 1000000
	sumValue := new(big.Int)
	for _, output := range outputs {
		if output.Value.Cmp(common.Big0) <= 0 {
			return "", fmt.Errorf("output value is not greater than zero")
		} else if !output.SplitFee {
			continue
		}
		sumValue.Add(sumValue, output.Value)
	}

	usedFees := new(big.Int)
	for _, output := range outputs {
		if !output.SplitFee {
			continue
		}

		proportionalFee := new(big.Int).Mul(new(big.Int).SetUint64(fee), output.Value)
		proportionalFee.Div(proportionalFee, sumValue)
		output.Value.Sub(output.Value, proportionalFee)
		usedFees.Add(usedFees, proportionalFee)
	}

	remainder := new(big.Int).Sub(new(big.Int).SetUint64(fee), usedFees)
	if remainder.Cmp(common.Big0) < 0 {
		return "", fmt.Errorf("fee remainder is negative")
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

	if remainder.Cmp(common.Big0) > 0 {
		return "", fmt.Errorf("not enough value to cover the fee remainder")
	}

	addresses := make([]string, len(outputs))
	amounts := make([]uint64, len(outputs))
	for i, output := range outputs {
		addresses[i] = output.Address
		amounts[i] = output.Value.Uint64()
	}

	tx, err := node.postWalletTransactionGenerate(addresses, amounts, fee)
	if err != nil {
		return "", err
	}
	txHex := hex.EncodeToString(tx)

	/*txid, err := node.postWalletTransactionCheck(tx)
	if err != nil {
		return "", err
	}*/

	return txHex, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	txBytes, err := hex.DecodeString(tx)
	if err != nil {
		return "", err
	}

	return node.postWalletTransactionSend(txBytes)
}

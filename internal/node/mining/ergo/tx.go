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

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, string, error) {
	if len(outputs) == 0 {
		return "", "", fmt.Errorf("need at least one output")
	}

	const fee = 1000000
	sumValue := new(big.Int)
	for _, output := range outputs {
		if output.Value.Cmp(common.Big0) <= 0 {
			return "", "", fmt.Errorf("output value is not greater than zero")
		}
		sumValue.Add(sumValue, output.Value)
	}

	usedFees := new(big.Int)
	for _, output := range outputs {
		if !output.SplitFee {
			output.Fee = new(big.Int)
			continue
		}

		output.Fee = new(big.Int).Mul(new(big.Int).SetUint64(fee), output.Value)
		output.Fee.Div(output.Fee, sumValue)
		output.Value.Sub(output.Value, output.Fee)
		usedFees.Add(usedFees, output.Fee)
	}

	remainder := new(big.Int).Sub(new(big.Int).SetUint64(fee), usedFees)
	if remainder.Cmp(common.Big0) < 0 {
		return "", "", fmt.Errorf("fee remainder is negative")
	} else if remainder.Cmp(common.Big0) > 0 {
		for _, output := range outputs {
			if output.Value.Cmp(remainder) > 0 {
				output.Value.Sub(output.Value, remainder)
				remainder = new(big.Int)
				break
			}
		}
	}

	if remainder.Cmp(common.Big0) > 0 {
		return "", "", fmt.Errorf("not enough value to cover the fee remainder")
	}

	addresses := make([]string, len(outputs))
	amounts := make([]uint64, len(outputs))
	for i, output := range outputs {
		addresses[i] = output.Address
		amounts[i] = output.Value.Uint64()
	}

	txBytes, err := node.postWalletTransactionGenerate(addresses, amounts, fee)
	if err != nil {
		return "", "", err
	}
	tx := hex.EncodeToString(txBytes)

	txid, err := node.postWalletTransactionCheck(txBytes)
	if err != nil {
		return "", "", err
	}

	return txid, tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	txBytes, err := hex.DecodeString(tx)
	if err != nil {
		return "", err
	}

	return node.postWalletTransactionSend(txBytes)
}

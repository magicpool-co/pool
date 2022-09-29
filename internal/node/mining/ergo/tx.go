package ergo

import (
	"fmt"
	"math/big"

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

	const fee = 2000000
	var feeCounter uint64
	for _, output := range outputs {
		if output.SplitFee {
			feeCounter++
		}
	}

	var feeDistribution uint64
	var remainderDistributed, forceRemainder bool
	if feeCounter != 0 {
		feeDistribution = fee / feeCounter
	} else {
		feeCounter = 1
		forceRemainder = true
	}
	remainder := fee - (feeDistribution * feeCounter)

	addresses := make([]string, len(outputs))
	amounts := make([]uint64, len(outputs))
	for i, output := range outputs {
		value := output.Value.Uint64()
		if output.SplitFee {
			value -= feeDistribution
		}

		// distribute remainder, force remainder if no one is set to split fees
		if !remainderDistributed && (output.SplitFee || forceRemainder) {
			value -= remainder
			remainderDistributed = true
		}

		if value <= 0 {
			return "", fmt.Errorf("not enough balance to pay fees")
		}

		addresses[i] = output.Address
		amounts[i] = value
	}

	return node.postWalletPaymentSend(addresses, amounts)
}

func (node Node) BroadcastTx(txid string) (string, error) {
	return txid, nil
}

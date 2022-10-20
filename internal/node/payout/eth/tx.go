package eth

import (
	"encoding/hex"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/tx/ethtx"
	"github.com/magicpool-co/pool/types"
)

const (
	baseFeeChangeDenominator = 8          // Bounds the amount the base fee can change between blocks.
	elasticityMultiplier     = 2          // Bounds the maximum gas limit an EIP-1559 block may have.
	initialBaseFee           = 1000000000 // Initial base fee for EIP-1559 blocks.
)

// BigMax returns the larger of x or y.
func bigMax(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 {
		return y
	}
	return x
}

func calcNextBaseFee(height, gasLimit, gasUsed uint64, currentBaseFee *big.Int) *big.Int {
	if height < 10_499_401 {
		return new(big.Int).SetUint64(initialBaseFee)
	}

	var parentGasTarget = gasLimit / elasticityMultiplier
	var parentGasTargetBig = new(big.Int).SetUint64(parentGasTarget)
	var baseFeeChangeDenominator = new(big.Int).SetUint64(baseFeeChangeDenominator)

	// If the parent gasUsed is the same as the target, the baseFee remains unchanged.
	if gasUsed == parentGasTarget {
		return new(big.Int).Set(currentBaseFee)
	} else if gasUsed > parentGasTarget {
		// If the parent block used more gas than its target, the baseFee should increase.
		gasUsedDelta := new(big.Int).SetUint64(gasUsed - parentGasTarget)
		x := new(big.Int).Mul(currentBaseFee, gasUsedDelta)
		y := x.Div(x, parentGasTargetBig)
		baseFeeDelta := bigMax(x.Div(y, baseFeeChangeDenominator), common.Big1)

		return x.Add(currentBaseFee, baseFeeDelta)
	} else {
		// Otherwise if the parent block used less gas than its target, the baseFee should decrease.
		gasUsedDelta := new(big.Int).SetUint64(parentGasTarget - gasUsed)
		x := new(big.Int).Mul(currentBaseFee, gasUsedDelta)
		y := x.Div(x, parentGasTargetBig)
		baseFeeDelta := x.Div(y, baseFeeChangeDenominator)

		return bigMax(x.Sub(currentBaseFee, baseFeeDelta), common.Big0)
	}
}

func (node Node) getBaseFee() (*big.Int, error) {
	height, err := node.getBlockNumber()
	if err != nil {
		return nil, err
	}

	block, err := node.getBlockByNumber(height)
	if err != nil {
		return nil, err
	}

	gasLimit, err := common.HexToUint64(block.GasLimit)
	if err != nil {
		return nil, err
	}

	gasUsed, err := common.HexToUint64(block.GasUsed)
	if err != nil {
		return nil, err
	}

	currentBaseFee, err := common.HexToBig(block.BaseFee)
	if err != nil {
		return nil, err
	}
	nextBaseFee := calcNextBaseFee(height, gasLimit, gasUsed, currentBaseFee)

	return nextBaseFee, nil
}

func (node Node) GetTxExplorerURL(txid string) string {
	return "https://etherscan.io/tx/" + txid
}

func (node Node) GetAddressExplorerURL(address string) string {
	return "https://etherscan.io/address/" + address
}

func (node Node) GetBalance() (*big.Int, error) {
	if node.erc20 != nil {
		data := ethtx.GenerateContractData("balanceOf(address)", ethCommon.HexToAddress(node.address).Bytes())
		params := []interface{}{
			map[string]interface{}{"to": node.erc20.Address, "data": "0x" + hex.EncodeToString(data)},
			"latest",
		}
		return node.sendCall(params)
	}

	return node.getBalance(node.address)
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	tx, err := node.getTransactionByHash(txid)
	if err != nil || tx == nil || len(tx.BlockNumber) <= 2 {
		return nil, err
	}

	txReceipt, err := node.getTransactionReceipt(txid)
	if err != nil {
		return nil, err
	}

	txType, err := common.HexToUint64(tx.Type)
	if err != nil {
		return nil, err
	}

	height, err := common.HexToUint64(tx.BlockNumber)
	if err != nil {
		return nil, err
	}

	value, err := common.HexToBig(tx.Value)
	if err != nil {
		return nil, err
	}

	gasUsed, err := common.HexToBig(txReceipt.GasUsed)
	if err != nil {
		return nil, err
	}

	gasTotal, err := common.HexToBig(tx.Gas)
	if err != nil {
		return nil, err
	}

	gasPrice, err := common.HexToBig(tx.GasPrice)
	if err != nil {
		return nil, err
	}

	confirmed := txReceipt.Status == "0x1"
	gasLeftover := new(big.Int).Sub(gasTotal, gasUsed)
	fees := new(big.Int).Mul(gasUsed, gasPrice)
	feeBalance := new(big.Int).Mul(gasLeftover, gasPrice)

	// handle type 2 (EIP1559) transaction fees
	if txType == 2 {
		block, err := node.getBlockByNumber(height)
		if err != nil {
			return nil, err
		}

		baseFeePerGas, err := common.HexToBig(block.BaseFee)
		if err != nil {
			return nil, err
		}

		maxFeePerGas, err := common.HexToBig(tx.MaxFeePerGas)
		if err != nil {
			return nil, err
		}

		maxPriorityFeePerGas, err := common.HexToBig(tx.MaxPriorityFeePerGas)
		if err != nil {
			return nil, err
		}

		// maxPriorityFeePerGas is included in maxFeePerGas and will always be spent
		// unless baseFeePerGas is greater than maxFeePerGas (which would be unlikely)
		feeSavings := new(big.Int).Sub(maxFeePerGas, baseFeePerGas)
		feeSavings.Sub(feeSavings, maxPriorityFeePerGas)
		if feeSavings.Cmp(common.Big0) < 0 {
			return nil, fmt.Errorf("invalid fee balance calc: base %s, max %s, priority %s",
				baseFeePerGas, maxFeePerGas, maxPriorityFeePerGas)
		}

		feeSavings.Mul(feeSavings, gasUsed)
		gasSavings := new(big.Int).Mul(maxFeePerGas, gasLeftover)
		feeBalance = new(big.Int).Add(feeSavings, gasSavings)
	}

	res := &types.TxResponse{
		Hash:        tx.Hash,
		BlockNumber: height,
		Value:       value,
		Fee:         fees,
		FeeBalance:  feeBalance,
		Confirmed:   confirmed,
	}

	return res, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, string, error) {
	if len(inputs) != 1 || len(outputs) != 1 {
		return "", "", fmt.Errorf("must have exactly one input and output")
	} else if inputs[0].Value.Cmp(outputs[0].Value) != 0 {
		return "", "", fmt.Errorf("inputs and outputs must have same value")
	}
	input := inputs[0]
	output := outputs[0]

	chainID, err := node.getChainID()
	if err != nil {
		return "", "", err
	}

	nonce, err := node.getPendingNonce(node.address)
	if err != nil {
		return "", "", err
	}
	// handle for future nonces
	nonce += uint64(input.Index)

	var toAddress string
	var value *big.Int
	var gasLimit uint64
	var data []byte
	if node.erc20 != nil {
		toAddress = node.erc20.Address
		value = new(big.Int)
		gasLimit = 100000
		outputAddress := ethCommon.HexToAddress(output.Address)
		data = ethtx.GenerateContractData("transfer(address,uint256)", outputAddress.Bytes(), input.Value.Bytes())
	} else {
		toAddress = output.Address
		value = new(big.Int).Set(input.Value)
		gasLimit, err = node.sendEstimateGas(node.address, toAddress)
		if err != nil {
			return "", "", err
		} else if gasLimit != 21000 {
			gasLimit += 30000
		}
	}

	baseFee, err := node.getBaseFee()
	if err != nil {
		return "", "", err
	}

	tx, fee, err := ethtx.NewTx(node.privKey.ToECDSA(), toAddress, data, value, baseFee, gasLimit, nonce, chainID)
	if err != nil {
		return "", "", err
	} else if node.erc20 != nil && (input.FeeBalance == nil || input.FeeBalance.Cmp(fee) > 0) {
		return "", "", fmt.Errorf("insufficient fee balance")
	}
	txid := ethtx.CalculateTxID(tx)

	if node.erc20 == nil {
		output.Value.Sub(output.Value, fee)
	}
	output.Fee = fee

	return txid, tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.sendRawTransaction(tx)
}

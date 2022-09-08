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

func (node Node) GetBalance(address string) (*big.Int, error) {
	if node.erc20 != nil {
		data := ethtx.GenerateContractData("balanceOf(address)", ethCommon.HexToAddress(address).Bytes())
		params := []interface{}{
			map[string]interface{}{"to": node.erc20.Address, "data": "0x" + hex.EncodeToString(data)},
			"latest",
		}
		return node.sendCall(params)
	}

	return node.getBalance(address)
}

func (node Node) GetTx(txid string) (*types.TxResponse, error) {
	return nil, nil
}

func (node Node) CreateTx(inputs []*types.TxInput, outputs []*types.TxOutput) (string, error) {
	if len(inputs) != 1 || len(outputs) != 1 {
		return "", fmt.Errorf("must have exactly one input and output")
	} else if inputs[0].Value.Cmp(outputs[0].Value) != 0 {
		return "", fmt.Errorf("inputs and outputs must have same value")
	}
	input := inputs[0]
	output := outputs[0]

	chainID, err := node.getChainID()
	if err != nil {
		return "", err
	}

	nonce, err := node.getPendingNonce(node.address)
	if err != nil {
		return "", err
	}

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
		gasLimit, err := node.sendEstimateGas(node.address, toAddress)
		if err != nil {
			return "", err
		} else if gasLimit != 21000 {
			gasLimit += 30000
		}
	}

	baseFee, err := node.getBaseFee()
	if err != nil {
		return "", err
	}

	tx, fees, err := ethtx.NewTx(node.privKey.ToECDSA(), toAddress, data, value, baseFee, gasLimit, nonce, chainID)
	if err != nil {
		return "", err
	} else if node.erc20 != nil && (input.FeeBalance == nil || input.FeeBalance.Cmp(fees) > 0) {
		return "", fmt.Errorf("insufficient fee balance")
	}

	return tx, nil
}

func (node Node) BroadcastTx(tx string) (string, error) {
	return node.sendRawTransaction(tx)
}

package ethtx

import (
	"fmt"

	"encoding/hex"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	ethCommon "github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/magicpool-co/pool/pkg/crypto"
)

func GenerateContractData(function string, args ...[]byte) []byte {
	data := crypto.Keccak256([]byte(function))[:4]
	for _, arg := range args {
		data = append(data, ethCommon.LeftPadBytes(arg, 32)...)
	}

	return data
}

func NewTx(
	privKey *secp256k1.PrivateKey,
	address string,
	data []byte,
	value, baseFee *big.Int,
	gasLimit, nonce, chainID uint64,
) (string, *big.Int, error) {
	toAddress := ethCommon.HexToAddress(address)

	// maxFee = (baseFee * 2) + priorityTip
	priorityTip := new(big.Int).SetUint64(3 * uint64(1e9))
	maxFee := new(big.Int).Mul(baseFee, big.NewInt(1))
	maxFee.Add(maxFee, priorityTip)
	fees := new(big.Int).Mul(maxFee, new(big.Int).SetUint64(gasLimit))
	if value.Cmp(fees) <= 0 {
		return "", nil, fmt.Errorf("fees greater than value")
	}
	value = new(big.Int).Sub(value, fees)

	signer := ethTypes.NewLondonSigner(new(big.Int).SetUint64(chainID))
	tx := ethTypes.NewTx(&ethTypes.DynamicFeeTx{
		ChainID:   new(big.Int).SetUint64(chainID),
		Nonce:     nonce,
		GasFeeCap: maxFee,
		GasTipCap: priorityTip,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     value,
		Data:      data,
	})

	signedTx, err := ethTypes.SignTx(tx, signer, privKey.ToECDSA())
	if err != nil {
		return "", nil, err
	}

	txBin, err := signedTx.MarshalBinary()
	if err != nil {
		return "", nil, err
	}

	encodedTx := make([]byte, len(txBin)*2+2)
	copy(encodedTx, "0x")
	hex.Encode(encodedTx[2:], txBin)

	return string(encodedTx), fees, nil
}

func NewLegacyTx(
	privKey *secp256k1.PrivateKey,
	address string,
	data []byte,
	value, gasPrice *big.Int, gasLimit, nonce, chainID uint64,
) (string, *big.Int, error) {
	toAddress := ethCommon.HexToAddress(address)

	fees := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasLimit))
	if value.Cmp(fees) <= 0 {
		return "", nil, fmt.Errorf("fees greater than value")
	}
	value = new(big.Int).Sub(value, fees)

	signer := ethTypes.NewLondonSigner(new(big.Int).SetUint64(chainID))
	tx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddress,
		Value:    value,
		Data:     data,
	})

	signedTx, err := ethTypes.SignTx(tx, signer, privKey.ToECDSA())
	if err != nil {
		return "", nil, err
	}

	txBin, err := signedTx.MarshalBinary()
	if err != nil {
		return "", nil, err
	}

	encodedTx := make([]byte, len(txBin)*2+2)
	copy(encodedTx, "0x")
	hex.Encode(encodedTx[2:], txBin)

	return string(encodedTx), fees, nil
}

func CalculateTxID(tx string) string {
	if len(tx) > 2 && tx[:2] == "0x" {
		tx = tx[2:]
	}

	txBytes, err := hex.DecodeString(tx)
	if err != nil {
		return ""
	}
	txid := crypto.Keccak256(txBytes)

	return "0x" + hex.EncodeToString(txid)
}

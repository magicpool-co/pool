package ethtx

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/crypto/sha3"
)

func GenerateContractData(function string, args ...[]byte) []byte {
	transferFnSignature := []byte(function)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]

	var data []byte
	data = append(data, methodID...)
	for _, arg := range args {
		data = append(data, ethCommon.LeftPadBytes(arg, 32)...)
	}

	return data
}

func NewTx(privKey *ecdsa.PrivateKey, address string, data []byte, value, baseFee *big.Int, gasLimit, nonce, chainID uint64) (string, *big.Int, error) {
	toAddress := ethCommon.HexToAddress(address)

	priorityTip := new(big.Int).SetUint64(3 * uint64(1e9))
	maxFee := new(big.Int).Mul(baseFee, big.NewInt(2))
	maxFee.Add(maxFee, priorityTip)
	fees := new(big.Int).Mul(maxFee, new(big.Int).SetUint64(gasLimit))

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

	signedTx, err := ethTypes.SignTx(tx, signer, privKey)
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

func NewLegacyTx(privKey *ecdsa.PrivateKey, address string, data []byte, value, gasPrice *big.Int, gasLimit, nonce, chainID uint64) (string, error) {
	toAddress := ethCommon.HexToAddress(address)

	signer := ethTypes.NewLondonSigner(new(big.Int).SetUint64(chainID))
	tx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddress,
		Value:    value,
		Data:     data,
	})

	signedTx, err := ethTypes.SignTx(tx, signer, privKey)
	if err != nil {
		return "", err
	}

	txBin, err := signedTx.MarshalBinary()
	if err != nil {
		return "", err
	}

	encodedTx := make([]byte, len(txBin)*2+2)
	copy(encodedTx, "0x")
	hex.Encode(encodedTx[2:], txBin)

	return string(encodedTx), nil
}

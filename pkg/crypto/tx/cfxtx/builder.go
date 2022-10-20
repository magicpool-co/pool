package cfxtx

import (
	"fmt"

	"crypto/ecdsa"
	"encoding/hex"
	"math/big"

	cfxTypes "github.com/Conflux-Chain/go-conflux-sdk/types"
	cfxAddress "github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/magicpool-co/pool/pkg/crypto"
)

func NewTx(privKey *ecdsa.PrivateKey, address string, data []byte, value, gasPrice *big.Int, gasLimit, storageLimit, nonce, chainID, epochNumber uint64) (string, *big.Int, error) {
	fromAddress, err := cfxAddress.NewFromBytes(ethCrypto.PubkeyToAddress(privKey.PublicKey).Bytes())
	if err != nil {
		return "", nil, err
	}

	toAddress, err := cfxAddress.NewFromBase32(address)
	if err != nil {
		return "", nil, err
	}

	// minimum gas price is one, in case it sets to zero
	if gasPrice.Cmp(new(big.Int).SetUint64(1)) != 1 {
		gasPrice = new(big.Int).SetUint64(1)
	}

	fee := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasLimit))
	if value.Cmp(fee) <= 0 {
		return "", nil, fmt.Errorf("fee greater than value")
	}
	value = new(big.Int).Sub(value, fee)

	tx := cfxTypes.UnsignedTransaction{
		UnsignedTransactionBase: cfxTypes.UnsignedTransactionBase{
			From:         &fromAddress,
			Value:        cfxTypes.NewBigInt(value.Uint64()),
			Nonce:        cfxTypes.NewBigInt(nonce),
			ChainID:      cfxTypes.NewUint(uint(chainID)),
			EpochHeight:  cfxTypes.NewUint64(epochNumber),
			Gas:          cfxTypes.NewBigInt(gasLimit),
			StorageLimit: cfxTypes.NewUint64(storageLimit * 10 / 9),
			GasPrice:     cfxTypes.NewBigInt(gasPrice.Uint64()),
		},
		To:   &toAddress,
		Data: data,
	}

	txHash, err := tx.Hash()
	if err != nil {
		return "", nil, err
	}

	sig, err := ethCrypto.Sign(txHash, privKey)
	if err != nil {
		return "", nil, err
	}

	txBin, err := tx.EncodeWithSignature(sig[64], sig[0:32], sig[32:64])
	if err != nil {
		return "", nil, err
	}

	encodedTx := make([]byte, len(txBin)*2+2)
	copy(encodedTx, "0x")
	hex.Encode(encodedTx[2:], txBin)

	return string(encodedTx), fee, nil
}

func CalculateTxID(tx string) string {
	txBytes, err := hex.DecodeString(tx)
	if err != nil {
		return ""
	}
	txid := crypto.Keccak256(txBytes)

	return "0x" + hex.EncodeToString(txid)
}

package cfxtx

import (
	"fmt"

	"encoding/hex"
	"math/big"

	cfxTypes "github.com/Conflux-Chain/go-conflux-sdk/types"
	cfxAddress "github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1signer "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	"github.com/magicpool-co/pool/pkg/crypto"
)

func signSecp256k1ETH(privKey *secp256k1.PrivateKey, msg []byte) []byte {
	const RecoveryIDOffset = 64

	sig := secp256k1signer.SignCompact(privKey, msg, false)
	v := sig[0] - 27
	copy(sig, sig[1:])
	sig[RecoveryIDOffset] = v

	return sig
}

func NewTx(
	privKey *secp256k1.PrivateKey,
	address string,
	data []byte,
	value, gasPrice *big.Int,
	gasLimit, storageLimit, nonce, chainID, epochNumber uint64,
) (string, *big.Int, error) {
	pubKeyBytes := privKey.PubKey().SerializeUncompressed()
	fromAddress, err := cfxAddress.NewFromBytes(crypto.Keccak256(pubKeyBytes[1:])[12:])
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

	sig := signSecp256k1ETH(privKey, txHash)
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

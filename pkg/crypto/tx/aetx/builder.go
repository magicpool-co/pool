package aetx

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/aeternity/rlp-go"

	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/base58"
)

const (
	addressPrefix = "ak_"
	txPrefix      = "tx_"
	txHashPrefix  = "th_"

	rlpMessageVersion  uint64 = 1
	rlpMessageVersion2 uint64 = 2

	IDTagAccount uint8 = 1

	ObjectTagAccount           uint64 = 10
	ObjectTagSignedTransaction uint64 = 11
	ObjectTagSpendTransaction  uint64 = 12
)

var (
	gasBase    = new(big.Int).SetUint64(15000)
	gasPerByte = new(big.Int).SetUint64(20)
	gasPrice   = new(big.Int).SetUint64(1e9)
)

/* helpers */

func calcChecksum(data []byte) []byte {
	checksum := crypto.Sha256d(data)
	return checksum[0:4]
}

func b58Encode(prefix string, data []byte) string {
	data = append(data, calcChecksum(data)...)
	return prefix + base58.Encode(data)
}

func b58Decode(input string) ([]byte, error) {
	if len(input) <= 8 || string(input[2]) != "_" {
		return nil, fmt.Errorf("invalid object encoding")
	}
	prefix := input[0:3]

	raw, err := base58.Decode(input[3:])
	if err != nil {
		return nil, err
	} else if len(raw) < 5 {
		return nil, fmt.Errorf("invalid input, %s cannot be decoded", input)
	}

	output := raw[:len(raw)-4]
	chk := b58Encode(prefix, output)
	if input != chk {
		return nil, fmt.Errorf("invalid checksum, expected %s got %s", chk, input)
	}

	return output, nil
}

func b64Encode(prefix string, data []byte) string {
	data = append(data, calcChecksum(data)...)
	return prefix + base64.StdEncoding.EncodeToString(data)
}

func b64Decode(input string) ([]byte, error) {
	if len(input) <= 8 || string(input[2]) != "_" {
		return nil, fmt.Errorf("invalid object encoding")
	}
	prefix := input[0:3]

	raw, err := base64.StdEncoding.DecodeString(input[3:])
	if err != nil {
		return nil, err
	} else if len(raw) < 5 {
		return nil, fmt.Errorf("invalid input, %s cannot be decoded", input)
	}

	output := raw[:len(raw)-4]
	chk := b64Encode(prefix, output)
	if input != chk {
		return nil, fmt.Errorf("invalid checksum, expected %s got %s", chk, input)
	}

	return output, nil
}

func buildRLPMessage(tag, version uint64, fields ...interface{}) ([]byte, error) {
	data := []interface{}{tag, version}
	data = append(data, fields...)

	return rlp.EncodeToBytes(data)
}

func buildIDTag(IDTag uint8, encodedHash string) ([]uint8, error) {
	if IDTag != IDTagAccount {
		return nil, fmt.Errorf("unknown id tag")
	}

	raw, err := b58Decode(encodedHash)
	if err != nil {
		return nil, err
	}

	v := []uint8{IDTag}
	for _, x := range raw {
		v = append(v, uint8(x))
	}

	return v, nil
}

func readIDTag(v []uint8) (string, error) {
	hash := []byte{}
	for _, x := range v[1:] {
		hash = append(hash, byte(x))
	}

	switch v[0] {
	case IDTagAccount:
	default:
		return "", fmt.Errorf("unknown id tag")
	}

	encodedHash := b58Encode(addressPrefix, hash)

	return encodedHash, nil
}

func calcFee(tx *transaction) (*big.Int, error) {
	rlpEncoded, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}

	length := new(big.Int).SetUint64(uint64(len(rlpEncoded)))
	gas := new(big.Int).Mul(length, gasPerByte)
	gas.Add(gas, gasBase)
	fee := new(big.Int).Mul(gas, gasPrice)

	return fee, nil
}

func signTx(priv ed25519.PrivateKey, tx *transaction, networkID string) (*signedTransaction, error) {
	txRaw, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	sig := ed25519.Sign(priv, append([]byte(networkID), txRaw...))

	signedTransaction := &signedTransaction{
		Signatures:  [][]byte{sig},
		Transaction: tx,
	}

	return signedTransaction, nil
}

func NewTx(privKey ed25519.PrivateKey, networkID, fromAddress, toAddress string, value *big.Int, nonce uint64) (string, *big.Int, error) {
	tx := &transaction{
		SenderID:    fromAddress,
		RecipientID: toAddress,
		Amount:      value,
		Fee:         new(big.Int).SetUint64(2e14),
		TTL:         0,
		Nonce:       nonce,
	}

	// dynamically update fee in case it changes size of tx
	var fee *big.Int
	for {
		var err error
		fee, err = calcFee(tx)
		if err != nil || tx.Fee.Cmp(fee) == 0 {
			break
		}
		tx.Fee = fee
	}

	signedTransaction, err := signTx(privKey, tx, networkID)
	if err != nil {
		return "", nil, err
	}

	encodedTx, err := rlp.EncodeToBytes(signedTransaction)
	if err != nil {
		return "", nil, err
	}

	txBin := b64Encode(txPrefix, encodedTx)

	return txBin, fee, nil
}

func CalculateTxID(tx string) string {
	txBytes, err := b64Decode(tx)
	if err != nil {
		return ""
	}

	return b58Encode(txHashPrefix, crypto.Blake2b256(txBytes))
}

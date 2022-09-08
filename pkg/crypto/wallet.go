package crypto

import (
	"crypto/ed25519"
	"fmt"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/magicpool-co/pool/pkg/crypto/base58"
)

func PrivKeyFromBytesED25519(rawPriv []byte) (ed25519.PrivateKey, error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid private key")
		}
	}()

	priv := ed25519.PrivateKey(rawPriv)

	return priv, err
}

func PrivKeyToWIFUncompressed(privKey *secp256k1.PrivateKey) string {
	encodeLen := 1 + 32 + 4
	a := make([]byte, 0, encodeLen)
	a = append(a, 0x80)
	a = append(a, privKey.Serialize()...)
	cksum := Sha256d(a)[:4]
	a = append(a, cksum...)

	return base58.Encode(a)
}

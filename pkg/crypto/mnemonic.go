package crypto

import (
	"crypto/sha512"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/text/unicode/norm"
)

func MnemonicToSeed(mnemonic, pass string) []byte {
	mnemonic = norm.NFKD.String(mnemonic)
	pass = norm.NFKD.String(pass)

	return pbkdf2.Key([]byte(mnemonic), []byte("mnemonic"+pass), 2048, 64, sha512.New)
}

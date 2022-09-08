package main

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"strings"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/magicpool-co/pool/pkg/crypto"
)

var (
	ErrNotImplemented = fmt.Errorf("Not implemented")
)

func generateSecp256k1Priv(obscure bool) ([]byte, []byte, error) {
	privKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	rawPriv := privKey.D.Bytes()
	obscuredPriv := rawPriv

	if obscure {
		obscuredPriv, err = crypto.ObscureHex(hex.EncodeToString(rawPriv))
		if err != nil {
			return nil, nil, err
		}

		err = crypto.ValidateSecp256k1PrivateKey(obscuredPriv)
		if err != nil {
			return nil, nil, err
		}
	}

	return rawPriv, obscuredPriv, nil
}

func generateEd25519Priv(obscure bool) ([]byte, error) {
	_, rawPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	if obscure {
		rawPriv, err = crypto.ObscureHex(hex.EncodeToString(rawPriv))
	}

	return rawPriv, err
}

func main() {
	argChain := flag.String("chain", "", "The chain to generate a secret for")
	argObscure := flag.Bool("obscure", false, "Whether or not to obscure the private key")

	flag.Parse()

	chain := strings.ToUpper(*argChain)

	var rawPriv, obscuredPriv []byte
	var err error
	switch chain {
	case "BTC", "RVN", "FLUX", "FIRO":
		rawPriv, obscuredPriv, err = generateSecp256k1Priv(*argObscure)
		rawPriv = obscuredPriv
	case "ETH", "ETC", "CTXC":
		rawPriv, obscuredPriv, err = generateSecp256k1Priv(*argObscure)
		rawPriv = obscuredPriv
	case "CFX":
		for {
			rawPriv, obscuredPriv, err = generateSecp256k1Priv(*argObscure)
			privKey := secp256k1.PrivKeyFromBytes(obscuredPriv)
			pubKeyBytes := privKey.PubKey().SerializeUncompressed()
			ethAddress := crypto.Keccak256(pubKeyBytes[1:])[12:]
			if ethAddress[0] == 0x10 {
				break
			}
		}
	case "AE":
		rawPriv, err = generateEd25519Priv(*argObscure)
	case "ERGO":
		err = ErrNotImplemented
	default:
		err = ErrNotImplemented
	}

	if err != nil {
		fmt.Printf("failed with error: %v\n", err)
	} else {
		fmt.Println(hex.EncodeToString(rawPriv))
	}
}

package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/magicpool-co/pool/pkg/crypto"
)

func generateSecp256k1Priv(obscure bool) ([]byte, error) {
	privKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	} else if !obscure {
		return privKey.D.Bytes(), nil
	}

	return crypto.ObscureHex(hex.EncodeToString(privKey.D.Bytes()))
}

func main() {
	argChain := flag.String("chain", "", "The chain to generate a secret for")
	argObscure := flag.Bool("obscure", false, "Whether or not to obscure the private key")

	flag.Parse()

	chain := strings.ToUpper(*argChain)

	var rawPriv []byte
	var err error
	switch chain {
	case "BTC", "ERG", "FIRO", "FLUX", "KAS", "NEXA", "RVN":
		rawPriv, err = generateSecp256k1Priv(*argObscure)
	case "BSC", "ETC", "ETH":
		rawPriv, err = generateSecp256k1Priv(*argObscure)
	case "CFX":
		for {
			rawPriv, err = generateSecp256k1Priv(*argObscure)
			if err != nil {
				break
			}

			privKey := secp256k1.PrivKeyFromBytes(rawPriv)
			pubKeyBytes := privKey.PubKey().SerializeUncompressed()
			ethAddress := crypto.Keccak256(pubKeyBytes[1:])[12:]
			if ethAddress[0] == 0x10 {
				break
			}
		}
	default:
		log.Fatalf("chain not supported")
	}

	if err != nil {
		log.Fatalf("failed with error: %v", err)
	}

	log.Printf(hex.EncodeToString(rawPriv))
}

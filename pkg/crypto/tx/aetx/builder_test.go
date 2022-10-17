package aetx

import (
	"math/big"
	"testing"

	"github.com/magicpool-co/pool/pkg/crypto"
)

func TestNewTx(t *testing.T) {
	tests := []struct {
		priv        []byte
		networkID   string
		fromAddress string
		toAddress   string
		amount      *big.Int
		nonce       uint64
		hash        string
		tx          string
	}{
		{
			priv: []byte{
				0x74, 0xe8, 0x41, 0x09, 0xd0, 0x8d, 0x23, 0xd7,
				0xd2, 0xad, 0x8c, 0xdc, 0x07, 0x10, 0x2b, 0x62,
				0xcc, 0xc9, 0xb6, 0x27, 0x39, 0xea, 0xec, 0x58,
				0x71, 0xa1, 0xab, 0x01, 0x82, 0x4c, 0x01, 0x21,
				0xc0, 0x39, 0x2a, 0x15, 0x51, 0xfe, 0x58, 0xa2,
				0x8a, 0x1e, 0x6b, 0xc8, 0xfe, 0x65, 0x89, 0x27,
				0x09, 0x2a, 0x0d, 0x3c, 0xce, 0xd8, 0xd7, 0x44,
				0xfc, 0x5a, 0x8e, 0x76, 0x52, 0xca, 0x08, 0x35,
			},
			networkID:   "ae_uat",
			fromAddress: "ak_2Tf5zs6vginLjcqksfwRi2WNkL6YZsytgN1npXUdTz4ZeirfFz",
			toAddress:   "ak_2Tf5zs6vginLjcqksfwRi2WNkL6YZsytgN1npXUdTz4ZUhUP2Q",
			amount:      new(big.Int).SetUint64(1e12),
			nonce:       1,
			hash:        "th_2d6MhS8XWKzEPkBysJUc59K6npK42cS4aDpRmwQqAMMJPFvVgW",
			tx:          "tx_+KALAfhCuEDTAUnoKY99nZZ2u1X78jLRrZ7koKTEK6tkTRNSI/xhD1nzTeI5xWJ5cdFRFxNq5NC8/Mrp/5NN4tPw+F/cwekMuFj4VgwBoQHAOSoVUf5Yoooea8j+ZYknCSoNPM7Y10T8Wo52UsoINaEBwDkqFVH+WKKKHmvI/mWJJwkqDTzO2NdE/FqOdlLKCDSF6NSlEACGDz492LAAAAGAYZaYZQ==",
		},
		{
			priv: []byte{
				0x74, 0xe8, 0x41, 0x09, 0xd0, 0x8d, 0x23, 0xd7,
				0xd2, 0xad, 0x8c, 0xdc, 0x07, 0x10, 0x2b, 0x62,
				0xcc, 0xc9, 0xb6, 0x27, 0x39, 0xea, 0xec, 0x58,
				0x71, 0xa1, 0xab, 0x01, 0x82, 0x4c, 0x01, 0x21,
				0xc0, 0x39, 0x2a, 0x15, 0x51, 0xfe, 0x58, 0xa2,
				0x8a, 0x1e, 0x6b, 0xc8, 0xfe, 0x65, 0x89, 0x27,
				0x09, 0x2a, 0x0d, 0x3c, 0xce, 0xd8, 0xd7, 0x44,
				0xfc, 0x5a, 0x8e, 0x76, 0x52, 0xca, 0x08, 0x35,
			},
			networkID:   "ae_uat",
			fromAddress: "ak_2Tf5zs6vginLjcqksfwRi2WNkL6YZsytgN1npXUdTz4ZUhUP2Q",
			toAddress:   "ak_2Tf5zs6vginLjcqksfwRi2WNkL6YZsytgN1npXUdTz4ZeirfFz",
			amount:      new(big.Int).SetUint64(1e19),
			nonce:       100,
			hash:        "th_2L1Bphd5HaSTyQ3Xb6SZn1LgwD26hFWQ6uapdSyyjB87fs4zEK",
			tx:          "tx_+KMLAfhCuECm4BdljgsR3xbFt/E3xz4QRMtTdscWF4UvcrIParSRC75FHeAK2vLZe8xpuF03gaYT2Y+QRtygcelaRguyA3UFuFv4WQwBoQHAOSoVUf5Yoooea8j+ZYknCSoNPM7Y10T8Wo52UsoINKEBwDkqFVH+WKKKHmvI/mWJJwkqDTzO2NdE/FqOdlLKCDWIiscjBInoAACGD0w2IAgAAGSAg3gocg==",
		},
		{
			priv: []byte{
				0x74, 0xe8, 0x41, 0x09, 0xd0, 0x8d, 0x23, 0xd7,
				0xd2, 0xad, 0x8c, 0xdc, 0x07, 0x10, 0x2b, 0x62,
				0xcc, 0xc9, 0xb6, 0x27, 0x39, 0xea, 0xec, 0x58,
				0x71, 0xa1, 0xab, 0x01, 0x82, 0x4c, 0x01, 0x21,
				0xc0, 0x39, 0x2a, 0x15, 0x51, 0xfe, 0x58, 0xa2,
				0x8a, 0x1e, 0x6b, 0xc8, 0xfe, 0x65, 0x89, 0x27,
				0x09, 0x2a, 0x0d, 0x3c, 0xce, 0xd8, 0xd7, 0x44,
				0xfc, 0x5a, 0x8e, 0x76, 0x52, 0xca, 0x08, 0x35,
			},
			networkID:   "ae_mainnet",
			fromAddress: "ak_2Tf5zs6vginLjcqksfwRi2WNkL6YZsytgN1npXUdTz4ZeirfFz",
			toAddress:   "ak_2Tf5zs6vginLjcqksfwRi2WNkL6YZsytgN1npXUdTz4ZUhUP2Q",
			amount:      new(big.Int).SetUint64(1e12),
			nonce:       0,
			hash:        "th_BNxzVo1LJH4jUUtpXydhSmmzaDubfpqFDLzNw3ER4HGEkXWio",
			tx:          "tx_+KALAfhCuEA8KH7y4hCtKv4d1JqVw2/cLdtItMW7VssWlYKJN6k0niFR80wBY3TCxjvlMfVEoS/mu8P6RH5TXAr4tRP/V/ULuFj4VgwBoQHAOSoVUf5Yoooea8j+ZYknCSoNPM7Y10T8Wo52UsoINaEBwDkqFVH+WKKKHmvI/mWJJwkqDTzO2NdE/FqOdlLKCDSF6NSlEACGDz492LAAAACAZd4/TQ==",
		},
	}

	for i, tt := range tests {
		priv, err := crypto.PrivKeyFromBytesED25519(tt.priv)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
			continue
		}

		tx, _, err := NewTx(priv, tt.networkID, tt.fromAddress, tt.toAddress, tt.amount, tt.nonce)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if tx != tt.tx {
			t.Errorf("failed on %d: tx mismatch: have %s, want %s", i, tx, tt.tx)
		} else if txid := CalculateTxID(tx); txid != tt.hash {
			t.Errorf("failed on %d: txid mismatch: have %s, want %s", i, txid, tt.hash)
		}
	}
}

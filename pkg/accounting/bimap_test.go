package accounting

import (
	"math/big"
	"testing"

	"github.com/magicpool-co/pool/pkg/db"
)

var (
	validIn  = new(big.Int).SetUint64(2151231)
	validOut = new(big.Int).SetUint64(42123)
)

func TestBimapAddByInput(t *testing.T) {
	tests := []struct {
		balances        []*db.Balance
		inputSums       map[string]*big.Int
		outputSums      map[string]*big.Int
		outputTradeSums map[string]*big.Int
		err             bool
	}{
		{
			balances: []*db.Balance{
				&db.Balance{
					InCoin:   "ETH",
					InValue:  db.NullBigInt{validIn, true},
					OutCoin:  "BTC",
					OutValue: db.NullBigInt{validOut, true},
				},
				&db.Balance{
					InCoin:   "ETC",
					InValue:  db.NullBigInt{validIn, true},
					OutCoin:  "BTC",
					OutValue: db.NullBigInt{nil, false},
				},
				&db.Balance{
					InCoin:   "RVN",
					InValue:  db.NullBigInt{validIn, true},
					OutCoin:  "ETH",
					OutValue: db.NullBigInt{validOut, true},
				},
				&db.Balance{
					InCoin:   "RVN",
					InValue:  db.NullBigInt{validIn, true},
					OutCoin:  "USDC",
					OutValue: db.NullBigInt{validOut, true},
				},
				&db.Balance{
					InCoin:   "ETH",
					InValue:  db.NullBigInt{validIn, true},
					OutCoin:  "USDC",
					OutValue: db.NullBigInt{validOut, true},
				},
			},
			inputSums: map[string]*big.Int{
				"ETH": new(big.Int).SetUint64(4302462),
				"ETC": new(big.Int).SetUint64(2151231),
				"RVN": new(big.Int).SetUint64(4302462),
			},
			outputSums: map[string]*big.Int{
				"ETH":  new(big.Int).SetUint64(42123),
				"BTC":  new(big.Int).SetUint64(42123),
				"USDC": new(big.Int).SetUint64(84246),
			},
			outputTradeSums: map[string]*big.Int{
				"ETH":  new(big.Int).SetUint64(42123),
				"BTC":  new(big.Int).SetUint64(42123),
				"USDC": new(big.Int).SetUint64(84246),
			},
			err: false,
		},
	}

	for i, tt := range tests {
		m := NewBimap()

		for j, balance := range tt.balances {
			err := m.AddByInput(balance, balance.OutValue.BigInt)
			if (err != nil) != tt.err {
				t.Errorf("failed on %d: AddByInput %d: have %v, want %v",
					i, j, err, tt.err)
			}
		}

		for coin, expected := range tt.inputSums {
			actual, ok := m.GetInputSum(coin)
			if !ok && expected.Cmp(new(big.Int)) != 0 {
				t.Errorf("failed on %d: GetInputSum %s: have %s, want %s",
					i, coin, actual, expected)
			} else if ok && expected.Cmp(actual) != 0 {
				t.Errorf("failed on %d: GetInputSum %s: have %s, want %s",
					i, coin, actual, expected)
			}
		}

		for coin, expected := range tt.outputSums {
			actual, ok := m.GetOutputSum(coin)
			if !ok && expected.Cmp(new(big.Int)) != 0 {
				t.Errorf("failed on %d: GetOutputSum %s: have %s, want %s",
					i, coin, actual, expected)
			} else if ok && expected.Cmp(actual) != 0 {
				t.Errorf("failed on %d: GetOutputSum %s: have %s, want %s",
					i, coin, actual, expected)
			}
		}
	}
}

package accounting

import (
	"database/sql"
	"math/big"
	"testing"

	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/types"
	"github.com/magicpool-co/pool/pkg/utils"
)

var (
	globalRates = map[string]map[string]float64{
		"ETH": map[string]float64{
			"BTC":  0.06194548,
			"USDC": 3557.04,
		},
		"ETC": map[string]float64{
			"ETH":  0.01520084,
			"BTC":  0.00094162,
			"USDC": 54.07,
		},
		"RVN": map[string]float64{
			"ETH":  0.00003069,
			"BTC":  0.00000190,
			"USDC": 0.1092,
		},
	}

	globalUnits = map[string]*big.Int{
		"ETH":  new(big.Int).SetUint64(1e18),
		"ETC":  new(big.Int).SetUint64(1e18),
		"RVN":  new(big.Int).SetUint64(1e8),
		"BTC":  new(big.Int).SetUint64(1e8),
		"USDC": new(big.Int).SetUint64(1e6),
	}
)

func generateBalanceSlim(inCoin, outCoin, inValue, outValue string) *db.Balance {
	outValueBig := db.NullBigInt{}
	inValueBig := db.NullBigInt{}

	if len(inValue) != 0 {
		inValueBig = db.NullBigInt{utils.MustParseBigInt(inValue), true}
	}

	if len(outValue) != 0 {
		outValueBig = db.NullBigInt{utils.MustParseBigInt(outValue), true}
	}

	balance := &db.Balance{
		InCoin:   inCoin,
		InValue:  inValueBig,
		OutCoin:  outCoin,
		OutValue: outValueBig,
	}

	return balance
}

func generateTrade(stage int, fromCoin, toCoin, value string) *db.SwitchTrade {
	trade := &db.SwitchTrade{
		Stage:    stage,
		FromCoin: fromCoin,
		ToCoin:   toCoin,
		Value:    db.NullBigInt{utils.MustParseBigInt(value), true},
	}

	return trade
}

func TestSwitchAccountantSwitching(t *testing.T) {
	tests := []struct {
		rates               map[string]map[string]float64
		units               map[string]*big.Int
		balances            []*db.Balance
		expectedBalancesQty int
	}{
		{
			rates:               globalRates,
			units:               globalUnits,
			balances:            []*db.Balance{},
			expectedBalancesQty: 0,
		},
		{
			rates: globalRates,
			units: globalUnits,
			balances: []*db.Balance{
				generateBalanceSlim("ETH", "BTC", "30000000000000000000", ""),
			},
			expectedBalancesQty: 1,
		},
		{
			rates: globalRates,
			units: globalUnits,
			balances: []*db.Balance{
				generateBalanceSlim("RVN", "BTC", "10000000000000", ""),
			},
			expectedBalancesQty: 1,
		},
		{
			rates: globalRates,
			units: globalUnits,
			balances: []*db.Balance{
				generateBalanceSlim("ETC", "BTC", "40000000000000000000", ""),
				generateBalanceSlim("ETC", "BTC", "70000000000000000000", ""),
			},
			expectedBalancesQty: 0,
		},
		{
			rates: globalRates,
			units: globalUnits,
			balances: []*db.Balance{
				generateBalanceSlim("ETC", "BTC", "20005345340062400534", ""),
				generateBalanceSlim("ETC", "BTC", "50052051200530006230", ""),
				generateBalanceSlim("ETC", "BTC", "30000034550000345053", ""),
			},
			expectedBalancesQty: 0,
		},
		{
			rates: globalRates,
			units: globalUnits,
			balances: []*db.Balance{
				generateBalanceSlim("ETC", "BTC", "20005345340062400534", ""),
				generateBalanceSlim("ETC", "BTC", "50052051200530006230", ""),
				generateBalanceSlim("ETC", "BTC", "30000034550000345053", ""),
				generateBalanceSlim("ETC", "ETH", "30000034550000345053", ""),
				generateBalanceSlim("RVN", "ETH", "500000000000", ""),
				generateBalanceSlim("ETH", "USDC", "2000000345500003450", ""),
			},
			expectedBalancesQty: 1,
		},
		{
			rates: globalRates,
			units: globalUnits,
			balances: []*db.Balance{
				generateBalanceSlim("ETC", "BTC", "20005345340062400534", ""),
				generateBalanceSlim("ETC", "BTC", "50052051200530006230", ""),
				generateBalanceSlim("ETC", "BTC", "300000034550000345053", ""),
				generateBalanceSlim("ETC", "ETH", "300000034550000345053", ""),
				generateBalanceSlim("RVN", "ETH", "5000000000000", ""),
				generateBalanceSlim("ETH", "USDC", "2000000345500003450", ""),
			},
			expectedBalancesQty: 6,
		},
	}

	for i, tt := range tests {
		accountant := NewSwitchAccountant()
		accountant.AddRates(tt.rates)
		accountant.AddUnits(tt.units)

		for j, balance := range tt.balances {
			if err := accountant.AddBalance(balance, true); err != nil {
				t.Errorf("failed on %d: add balance %d: %v", i, j, err)
			}
		}

		if err := accountant.CheckThresholds(); err != nil {
			t.Errorf("failed on %d: check threshold: %v", i, err)
		} else if len(accountant.Balances()) != tt.expectedBalancesQty {
			t.Errorf("failed on %d: balance count mismatch: have %d, want %d",
				i, len(accountant.Balances()), tt.expectedBalancesQty)
		}
	}
}

func TestSwitchAccountantDeposit(t *testing.T) {
	tests := []struct {
		balances    []*db.Balance
		inputValues map[string]*big.Int
	}{
		{
			balances: []*db.Balance{
				generateBalanceSlim("ETC", "BTC", "40000000000000000000", ""),
				generateBalanceSlim("ETC", "BTC", "70000000000000000000", ""),
			},
			inputValues: map[string]*big.Int{
				"ETH": utils.MustParseBigInt("0"),
				"ETC": utils.MustParseBigInt("110000000000000000000"),
				"RVN": utils.MustParseBigInt("0"),
			},
		},
		{
			balances: []*db.Balance{
				generateBalanceSlim("ETC", "BTC", "20005345340062400534", ""),
				generateBalanceSlim("ETC", "BTC", "50052051200530006230", ""),
				generateBalanceSlim("ETC", "BTC", "300000034550000345053", ""),
				generateBalanceSlim("ETC", "ETH", "30000034550000345053", ""),
				generateBalanceSlim("RVN", "ETH", "5000000000000", ""),
				generateBalanceSlim("ETH", "USDC", "2000000345500003450", ""),
			},
			inputValues: map[string]*big.Int{
				"ETH": utils.MustParseBigInt("2000000345500003450"),
				"ETC": utils.MustParseBigInt("400057465640593096870"),
				"RVN": utils.MustParseBigInt("5000000000000"),
			},
		},
	}

	for i, tt := range tests {
		accountant := NewSwitchAccountant()

		for j, balance := range tt.balances {
			if err := accountant.AddBalance(balance, false); err != nil {
				t.Errorf("failed on %d: add balance %d: %v", i, j, err)
			}
		}

		for coin, actual := range accountant.InputValues() {
			if expected, ok := tt.inputValues[coin]; !ok {
				t.Errorf("failed on %d: cannot find %s", i, coin)
			} else if actual.Cmp(expected) != 0 {
				t.Errorf("failed on %d: %s value mismatch: have %s, want %s",
					i, coin, actual, expected)
			}
		}
	}
}

func TestSwitchAccountantTrade(t *testing.T) {
	tests := []struct {
		balances []*db.Balance
		deposits []*db.SwitchDeposit
		trades   [][]*db.SwitchTrade
	}{
		{
			balances: []*db.Balance{
				{
					ID:          1,
					MinerID:     sql.NullInt64{1, true},
					InCoin:      "ETH",
					InValue:     db.NullBigInt{utils.MustParseBigInt("30000000000000000000"), true},
					OutCoin:     "BTC",
					OutValue:    db.NullBigInt{utils.MustParseBigInt("183299999"), true},
					OutType:     int(types.StandardBalance),
					OutPayoutID: 1,
				},
			},
			deposits: []*db.SwitchDeposit{
				&db.SwitchDeposit{
					CoinID: "ETH",
					Value:  db.NullBigInt{utils.MustParseBigInt("27999999999968800000"), true},
					Fees:   db.NullBigInt{utils.MustParseBigInt("2000000000031200000"), true},
				},
			},
			trades: [][]*db.SwitchTrade{
				{
					generateTrade(1, "ETH", "BTC", "27999999999968800000"),
				},
			},
		},
	}

	for i, tt := range tests {
		accountant := NewSwitchAccountant()

		for j, balance := range tt.balances {
			if err := accountant.AddBalance(balance, false); err != nil {
				t.Errorf("failed on %d: add balance %d: %v", i, j, err)
			}
		}

		for j, deposit := range tt.deposits {
			if err := accountant.AddDeposit(deposit); err != nil {
				t.Errorf("failed on %d: add deposit %d: %v", i, j, err)
			}
		}

		allTrades, err := accountant.GenerateTrades(1)
		if err != nil {
			t.Errorf("failed on %d: generate trades: %v", i, err)
			return
		}

		for j, actualTrades := range allTrades {
			expectedTrades := tt.trades[j]
			for k, actual := range actualTrades {
				expected := expectedTrades[k]
				if expected == nil {
					t.Errorf("test %d: trade length mismatch: unable to find %d,%d", i, j, k)
				}

				if expected.Stage != actual.Stage {
					t.Errorf("test %d: stage mismatch %d,%d: have %d, want %d",
						i, j, j, actual.Stage, expected.Stage)
				}

				if expected.FromCoin != actual.FromCoin {
					t.Errorf("test %d: from_coin mismatch %d,%d: have %s, want %s",
						i, j, j, actual.FromCoin, expected.FromCoin)
				}

				if expected.ToCoin != actual.ToCoin {
					t.Errorf("test %d: to_coin mismatch %d,%d: have %s, want %s",
						i, j, j, actual.ToCoin, expected.ToCoin)
				}

				if expected.Value.BigInt.Cmp(actual.Value.BigInt) != 0 {
					t.Errorf("test %d: value mismatch %d,%d: have %s, want %s",
						i, j, j, actual.Value.BigInt, expected.Value.BigInt)
				}
			}
		}
	}
}

func TestSwitchAccountantWithdrawal(t *testing.T) {
	tests := []struct {
		balances    []*db.Balance
		payouts     []*db.Payout
		deposits    []*db.SwitchDeposit
		paths       [][]*db.SwitchTrade
		withdrawals []*db.SwitchWithdrawal
	}{
		{
			payouts: []*db.Payout{
				{
					ID:                1,
					CoinID:            "BTC",
					MinerID:           sql.NullInt64{1, true},
					Address:           "0x63806bA7A87c51Ae09F183E853f6ea336D0a2783",
					Value:             db.NullBigInt{new(big.Int), true},
					PoolFees:          db.NullBigInt{new(big.Int), true},
					ExchangeFees:      db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin:    "BTC",
					FeeBalancePending: false,
					InFeeBalance:      db.NullBigInt{new(big.Int), true},
					OutFeeBalance:     db.NullBigInt{new(big.Int), true},
				},
			},
			deposits: []*db.SwitchDeposit{
				&db.SwitchDeposit{
					CoinID: "ETH",
					Value:  db.NullBigInt{utils.MustParseBigInt("27999999999968800000"), true},
					Fees:   db.NullBigInt{utils.MustParseBigInt("2000000000031200000"), true},
				},
			},
			withdrawals: []*db.SwitchWithdrawal{
				&db.SwitchWithdrawal{
					CoinID:    "BTC",
					Value:     db.NullBigInt{utils.MustParseBigInt("183299999"), true},
					Fees:      db.NullBigInt{utils.MustParseBigInt("54123"), true},
					TradeFees: db.NullBigInt{utils.MustParseBigInt("56524"), true},
				},
			},
			paths: [][]*db.SwitchTrade{
				[]*db.SwitchTrade{
					&db.SwitchTrade{
						FromCoin: "ETH",
						ToCoin:   "BTC",
						Value:    db.NullBigInt{utils.MustParseBigInt("27999999999968800000"), true},
						Proceeds: db.NullBigInt{utils.MustParseBigInt("183354122"), true},
						Fees:     db.NullBigInt{utils.MustParseBigInt("56524"), true},
					},
				},
			},
			balances: []*db.Balance{
				{
					ID:           1,
					MinerID:      sql.NullInt64{1, true},
					InCoin:       "ETH",
					InValue:      db.NullBigInt{utils.MustParseBigInt("30000000000000000000"), true},
					PoolFees:     db.NullBigInt{new(big.Int), true},
					ExchangeFees: db.NullBigInt{new(big.Int), true},
					OutCoin:      "BTC",
					OutValue:     db.NullBigInt{utils.MustParseBigInt("183299999"), true},
					OutType:      int(types.StandardBalance),
					OutPayoutID:  1,
				},
			},
		},
		{
			payouts: []*db.Payout{
				{
					ID:                1,
					CoinID:            "BTC",
					MinerID:           sql.NullInt64{1, true},
					Address:           "0x63806bA7A87c51Ae09F183E853f6ea336D0a2783",
					Value:             db.NullBigInt{new(big.Int), true},
					PoolFees:          db.NullBigInt{new(big.Int), true},
					ExchangeFees:      db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin:    "BTC",
					FeeBalancePending: false,
					InFeeBalance:      db.NullBigInt{new(big.Int), true},
					OutFeeBalance:     db.NullBigInt{new(big.Int), true},
				},
				{
					ID:                2,
					CoinID:            "ETH",
					MinerID:           sql.NullInt64{1, true},
					Address:           "0x63806bA7A87c51Ae09F183E853f6ea336D0a2783",
					Value:             db.NullBigInt{new(big.Int), true},
					PoolFees:          db.NullBigInt{new(big.Int), true},
					ExchangeFees:      db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin:    "ETH",
					FeeBalancePending: false,
					InFeeBalance:      db.NullBigInt{new(big.Int), true},
					OutFeeBalance:     db.NullBigInt{new(big.Int), true},
				},
				{
					ID:                3,
					CoinID:            "BTC",
					MinerID:           sql.NullInt64{2, true},
					Address:           "0x63806bA7A87c51Ae09F183E853f6ea336D0a2783",
					Value:             db.NullBigInt{new(big.Int), true},
					PoolFees:          db.NullBigInt{new(big.Int), true},
					ExchangeFees:      db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin:    "BTC",
					FeeBalancePending: false,
					InFeeBalance:      db.NullBigInt{new(big.Int), true},
					OutFeeBalance:     db.NullBigInt{new(big.Int), true},
				},
			},
			deposits: []*db.SwitchDeposit{
				&db.SwitchDeposit{
					CoinID: "ETC",
					Value:  db.NullBigInt{utils.MustParseBigInt("151931159250000000000"), true},
					Fees:   db.NullBigInt{utils.MustParseBigInt("0"), true},
				},
				&db.SwitchDeposit{
					CoinID: "ETH",
					Value:  db.NullBigInt{utils.MustParseBigInt("1351222550000000000"), true},
					Fees:   db.NullBigInt{utils.MustParseBigInt("0"), true},
				},
			},
			withdrawals: []*db.SwitchWithdrawal{
				&db.SwitchWithdrawal{
					CoinID:    "BTC",
					Value:     db.NullBigInt{utils.MustParseBigInt("10232333"), true},
					Fees:      db.NullBigInt{utils.MustParseBigInt("30000"), true},
					TradeFees: db.NullBigInt{utils.MustParseBigInt("25717"), true},
				},
				&db.SwitchWithdrawal{
					CoinID:    "ETH",
					Value:     db.NullBigInt{utils.MustParseBigInt("1295872690000000000"), true},
					Fees:      db.NullBigInt{utils.MustParseBigInt("13400000000000000"), true},
					TradeFees: db.NullBigInt{utils.MustParseBigInt("3281370000000000"), true},
				},
			},
			paths: [][]*db.SwitchTrade{
				[]*db.SwitchTrade{
					&db.SwitchTrade{
						FromCoin: "ETC",
						ToCoin:   "BTC",
						Value:    db.NullBigInt{utils.MustParseBigInt("13187159772104700000"), true},
						Proceeds: db.NullBigInt{utils.MustParseBigInt("878396"), true},
						Fees:     db.NullBigInt{utils.MustParseBigInt("2200"), true},
					},
				},
				[]*db.SwitchTrade{
					&db.SwitchTrade{
						FromCoin: "ETC",
						ToCoin:   "ETH",
						Value:    db.NullBigInt{utils.MustParseBigInt("138743999477895300000"), true},
						Proceeds: db.NullBigInt{utils.MustParseBigInt("1309272690000000000"), true},
						Fees:     db.NullBigInt{utils.MustParseBigInt("3281370000000000"), true},
					},
				},
				[]*db.SwitchTrade{
					&db.SwitchTrade{
						FromCoin: "ETH",
						ToCoin:   "BTC",
						Value:    db.NullBigInt{utils.MustParseBigInt("1351222550000000000"), true},
						Proceeds: db.NullBigInt{utils.MustParseBigInt("9383937"), true},
						Fees:     db.NullBigInt{utils.MustParseBigInt("23517"), true},
					},
				},
			},
			balances: []*db.Balance{
				{
					ID:           1,
					MinerID:      sql.NullInt64{1, true},
					InCoin:       "ETC",
					InValue:      db.NullBigInt{utils.MustParseBigInt("13187159772104700000"), true},
					PoolFees:     db.NullBigInt{new(big.Int), true},
					ExchangeFees: db.NullBigInt{new(big.Int), true},
					OutCoin:      "BTC",
					OutValue:     db.NullBigInt{utils.MustParseBigInt("875829"), true},
					OutType:      int(types.StandardBalance),
					OutPayoutID:  1,
				},
				{
					ID:           2,
					MinerID:      sql.NullInt64{1, true},
					InCoin:       "ETC",
					InValue:      db.NullBigInt{utils.MustParseBigInt("138743999477895300000"), true},
					PoolFees:     db.NullBigInt{new(big.Int), true},
					ExchangeFees: db.NullBigInt{new(big.Int), true},
					OutCoin:      "ETH",
					OutValue:     db.NullBigInt{utils.MustParseBigInt("1295872690000000000"), true},
					OutType:      int(types.StandardBalance),
					OutPayoutID:  2,
				},
				{
					ID:           3,
					MinerID:      sql.NullInt64{2, true},
					InCoin:       "ETH",
					InValue:      db.NullBigInt{utils.MustParseBigInt("1351222550000000000"), true},
					PoolFees:     db.NullBigInt{new(big.Int), true},
					ExchangeFees: db.NullBigInt{new(big.Int), true},
					OutCoin:      "BTC",
					OutValue:     db.NullBigInt{utils.MustParseBigInt("9356504"), true},
					OutType:      int(types.StandardBalance),
					OutPayoutID:  3,
				},
			},
		},
	}

	for i, tt := range tests {
		accountant := NewSwitchAccountant()

		for j, balance := range tt.balances {
			copy := &db.Balance{
				ID:           balance.ID,
				InCoin:       balance.InCoin,
				InValue:      balance.InValue,
				PoolFees:     balance.PoolFees,
				ExchangeFees: balance.ExchangeFees,
				OutCoin:      balance.OutCoin,
				OutType:      balance.OutType,
				OutPayoutID:  balance.OutPayoutID,
			}

			if err := accountant.AddBalance(copy, false); err != nil {
				t.Errorf("failed on %d: add balance %d: %v", i, j, err)
			}
		}

		for j, payout := range tt.payouts {
			if err := accountant.AddPayout(payout); err != nil {
				t.Errorf("failed on %d: add payout %d: %v", i, j, err)
			}
		}

		for j, deposit := range tt.deposits {
			if err := accountant.AddDeposit(deposit); err != nil {
				t.Errorf("failed on %d: add deposit %d: %v", i, j, err)
			}
		}

		for j, path := range tt.paths {
			if err := accountant.AddFinalPath(path); err != nil {
				t.Errorf("failed on %d: add final path %d: %v", i, j, err)
			}
		}

		for j, withdrawal := range tt.withdrawals {
			if err := accountant.AddWithdrawal(withdrawal); err != nil {
				t.Errorf("failed on %d: add withdrawal %d: %v", i, j, err)
			}
		}

		if err := accountant.Distribute(); err != nil {
			t.Errorf("failed on %d: distribute: %v", i, err)
		}

		for j, actual := range accountant.Balances() {
			expected := tt.balances[j]

			if expected.OutValue.BigInt.Cmp(actual.OutValue.BigInt) != 0 {
				t.Errorf("failed on %d: balance mismatch %d: have %s, want %s",
					i, j, actual.OutValue.BigInt, expected.OutValue.BigInt)
			}
		}
	}
}

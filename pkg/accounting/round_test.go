package accounting

import (
	"database/sql"
	"math/big"
	"testing"

	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/types"
	"github.com/magicpool-co/pool/pkg/utils"
)

func TestRoundAccountant(t *testing.T) {
	tests := []struct {
		ethRate    float64
		spendable  *big.Int
		rounds     []*db.Round
		shares     map[uint64][]*db.Share
		miners     []*db.Miner
		recipients []*db.Recipient
		payouts    map[uint64]*db.Payout
		balances   []*db.Balance
	}{
		{
			ethRate:   0.705,
			spendable: utils.MustParseBigInt("3200092000012004030"),
			rounds: []*db.Round{
				&db.Round{
					ID:     1,
					CoinID: "ETH",
					Value:  db.NullBigInt{new(big.Int).SetUint64(3200092000012004030), true},
				},
			},
			shares: map[uint64][]*db.Share{
				1: []*db.Share{
					&db.Share{RoundID: 1, MinerID: 1, Count: 1235123},
					&db.Share{RoundID: 1, MinerID: 2, Count: 2131500},
					&db.Share{RoundID: 1, MinerID: 3, Count: 5432},
					&db.Share{RoundID: 1, MinerID: 4, Count: 645734},
					&db.Share{RoundID: 1, MinerID: 5, Count: 5346356},
				},
			},
			miners: []*db.Miner{
				&db.Miner{ID: 1, CoinID: "ETH"},
				&db.Miner{ID: 2, CoinID: "USDC"},
				&db.Miner{ID: 3, CoinID: "BTC"},
				&db.Miner{ID: 4, CoinID: "BTC"},
				&db.Miner{ID: 5, CoinID: "ETH"},
			},
			recipients: []*db.Recipient{
				&db.Recipient{ID: 6, CoinID: "ETH", Fraction: 90},
				&db.Recipient{ID: 7, CoinID: "BTC", Fraction: 5},
				&db.Recipient{ID: 8, CoinID: "USDC", Fraction: 5},
			},
			balances: []*db.Balance{
				&db.Balance{
					MinerID:  sql.NullInt64{1, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("417868599751233906"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("4220894946982160"), true},
					OutType:  int(types.StandardBalance),
					OutValue: db.NullBigInt{utils.MustParseBigInt("417868599751233906"), true},
				},
				&db.Balance{
					MinerID:  sql.NullInt64{2, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("685671169937320129"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("6925971413508284"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					MinerID:  sql.NullInt64{3, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("1837762096446024"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("18563253499454"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					MinerID:  sql.NullInt64{4, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("218465660822252741"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("2206723846689421"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					MinerID:  sql.NullInt64{5, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("1808786894496829773"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("18270574691887169"), true},
					OutType:  int(types.StandardBalance),
					OutValue: db.NullBigInt{utils.MustParseBigInt("1808786894496829773"), true},
				},
				&db.Balance{
					MinerID:  sql.NullInt64{2, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("35460992907801420"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("358191847553549"), true},
					OutType:  int(types.FeeBalance),
					OutValue: db.NullBigInt{utils.MustParseBigInt("35460992907801420"), true},
				},
				&db.Balance{
					RecipientID: sql.NullInt64{6, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("28800828000108035"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutType:     int(types.StandardBalance),
					OutValue:    db.NullBigInt{utils.MustParseBigInt("28800828000108035"), true},
				},
				&db.Balance{
					RecipientID: sql.NullInt64{7, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("1600046000006001"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutType:     int(types.StandardBalance),
				},
				&db.Balance{
					RecipientID: sql.NullInt64{8, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("0"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutType:     int(types.StandardBalance),
				},
				&db.Balance{
					RecipientID: sql.NullInt64{8, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("1600046000006001"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutType:     int(types.FeeBalance),
					OutValue:    db.NullBigInt{utils.MustParseBigInt("1600046000006001"), true},
				},
			},
			payouts: map[uint64]*db.Payout{
				1: &db.Payout{
					MinerID:        sql.NullInt64{1, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				2: &db.Payout{
					MinerID:        sql.NullInt64{2, true},
					CoinID:         "USDC",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				3: &db.Payout{
					MinerID:        sql.NullInt64{3, true},
					CoinID:         "BTC",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "BTC",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				4: &db.Payout{
					MinerID:        sql.NullInt64{4, true},
					CoinID:         "BTC",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "BTC",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				5: &db.Payout{
					MinerID:        sql.NullInt64{5, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				6: &db.Payout{
					RecipientID:    sql.NullInt64{6, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				7: &db.Payout{
					RecipientID:    sql.NullInt64{7, true},
					CoinID:         "BTC",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "BTC",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				8: &db.Payout{
					MinerID:        sql.NullInt64{8, true},
					CoinID:         "USDC",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
			},
		},
		{
			ethRate:   0.705,
			spendable: utils.MustParseBigInt("1205123132322041235"),
			rounds: []*db.Round{
				&db.Round{
					ID:     1,
					CoinID: "ETC",
					Value:  db.NullBigInt{new(big.Int).SetUint64(1205123132322041235), true},
				},
			},
			shares: map[uint64][]*db.Share{
				1: []*db.Share{
					&db.Share{RoundID: 1, MinerID: 1, Count: 7345},
					&db.Share{RoundID: 1, MinerID: 2, Count: 734334},
					&db.Share{RoundID: 1, MinerID: 3, Count: 745664},
					&db.Share{RoundID: 1, MinerID: 5, Count: 5346356},
				},
			},
			miners: []*db.Miner{
				&db.Miner{ID: 1, CoinID: "ETH"},
				&db.Miner{ID: 2, CoinID: "USDC"},
				&db.Miner{ID: 3, CoinID: "BTC"},
				&db.Miner{ID: 4, CoinID: "BTC"},
				&db.Miner{ID: 5, CoinID: "ETH"},
			},
			recipients: []*db.Recipient{
				&db.Recipient{ID: 6, CoinID: "ETH", Fraction: 90},
				&db.Recipient{ID: 7, CoinID: "BTC", Fraction: 5},
				&db.Recipient{ID: 8, CoinID: "USDC", Fraction: 5},
			},
			balances: []*db.Balance{
				&db.Balance{
					MinerID:  sql.NullInt64{1, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("1282338176269741"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("12952910871411"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					MinerID:  sql.NullInt64{2, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("92743843352629144"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("936806498511406"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					MinerID:  sql.NullInt64{3, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("130182901820285724"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("1314978806265512"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					MinerID:  sql.NullInt64{5, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("933401824741834795"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("9428301260018533"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					MinerID:  sql.NullInt64{2, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("35460992907801420"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("358191847553549"), true},
					OutType:  int(types.FeeBalance),
				},
				&db.Balance{
					RecipientID: sql.NullInt64{6, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("10846108190898371"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutType:     int(types.StandardBalance),
				},
				&db.Balance{
					RecipientID: sql.NullInt64{7, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("602561566161020"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutType:     int(types.StandardBalance),
				},
				&db.Balance{
					RecipientID: sql.NullInt64{8, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("0"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutType:     int(types.StandardBalance),
				},
				&db.Balance{
					RecipientID: sql.NullInt64{8, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("602561566161020"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutType:     int(types.FeeBalance),
				},
			},
			payouts: map[uint64]*db.Payout{
				1: &db.Payout{
					MinerID:        sql.NullInt64{1, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				2: &db.Payout{
					MinerID:        sql.NullInt64{2, true},
					CoinID:         "USDC",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				3: &db.Payout{
					MinerID:        sql.NullInt64{3, true},
					CoinID:         "BTC",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "BTC",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				4: &db.Payout{
					MinerID:        sql.NullInt64{4, true},
					CoinID:         "BTC",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "BTC",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				5: &db.Payout{
					MinerID:        sql.NullInt64{5, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				6: &db.Payout{
					RecipientID:    sql.NullInt64{6, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				7: &db.Payout{
					RecipientID:    sql.NullInt64{7, true},
					CoinID:         "BTC",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "BTC",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				8: &db.Payout{
					MinerID:        sql.NullInt64{8, true},
					CoinID:         "USDC",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
			},
		},
		{
			ethRate:   0.01209455,
			spendable: utils.MustParseBigInt("3200693000000000000"),
			rounds: []*db.Round{
				&db.Round{
					ID:     1,
					CoinID: "ETC",
					Value:  db.NullBigInt{new(big.Int).SetUint64(3200693000000000000), true},
				},
			},
			shares: map[uint64][]*db.Share{
				1: []*db.Share{
					&db.Share{RoundID: 1, MinerID: 1, Count: 11264},
					&db.Share{RoundID: 1, MinerID: 2, Count: 88736},
				},
			},
			miners: []*db.Miner{
				&db.Miner{ID: 1, CoinID: "ETH"},
				&db.Miner{ID: 2, CoinID: "ETH"},
			},
			recipients: []*db.Recipient{
				&db.Recipient{ID: 6, CoinID: "ETH", Fraction: 90},
				&db.Recipient{ID: 7, CoinID: "ETH", Fraction: 10},
			},
			balances: []*db.Balance{
				&db.Balance{
					MinerID:  sql.NullInt64{1, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("356920798924800000"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("3605260595200000"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					MinerID:  sql.NullInt64{2, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("2811765271075200000"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("28401669404800000"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					RecipientID: sql.NullInt64{6, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("28806237000000000"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutType:     int(types.StandardBalance),
				},
				&db.Balance{
					RecipientID: sql.NullInt64{7, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("3200693000000000"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutType:     int(types.StandardBalance),
				},
			},
			payouts: map[uint64]*db.Payout{
				1: &db.Payout{
					MinerID:        sql.NullInt64{1, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int).SetUint64(2445333413796000), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				2: &db.Payout{
					MinerID:        sql.NullInt64{2, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int).SetUint64(2445333413796000), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				6: &db.Payout{
					RecipientID:    sql.NullInt64{6, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int).SetUint64(47139892217568949), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int).SetUint64(1905033855762097), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				7: &db.Payout{
					RecipientID:    sql.NullInt64{7, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int).SetUint64(5237765801856067), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int).SetUint64(211670428416693), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
			},
		},
		{
			ethRate:   0.01209455,
			spendable: utils.MustParseBigInt("4401386000000000000"),
			rounds: []*db.Round{
				&db.Round{
					ID:       1,
					CoinID:   "ETH",
					Value:    db.NullBigInt{new(big.Int).SetUint64(2200693000000000000), true},
					MEVValue: db.NullBigInt{new(big.Int).SetUint64(2200693000000000000), true},
				},
			},
			shares: map[uint64][]*db.Share{
				1: []*db.Share{
					&db.Share{RoundID: 1, MinerID: 1, Count: 88736},
					&db.Share{RoundID: 1, MinerID: 2, Count: 12341},
				},
			},
			miners: []*db.Miner{
				&db.Miner{ID: 1, CoinID: "ETH", Address: "0x7dD8E752F5e606Aca3A40DD1dEaDF363dbbCa100"},
				&db.Miner{ID: 2, CoinID: "ETH"},
			},
			recipients: []*db.Recipient{
				&db.Recipient{ID: 6, CoinID: "ETH", Fraction: 90},
				&db.Recipient{ID: 7, CoinID: "ETH", Fraction: 10},
			},
			balances: []*db.Balance{
				&db.Balance{
					MinerID:  sql.NullInt64{1, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("3846610622095709213"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("17387993771402000"), true},
					OutValue: db.NullBigInt{utils.MustParseBigInt("3846610622095709213"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					MinerID:  sql.NullInt64{2, true},
					InValue:  db.NullBigInt{utils.MustParseBigInt("532013510291559900"), true},
					PoolFees: db.NullBigInt{utils.MustParseBigInt("5373873841328887"), true},
					OutValue: db.NullBigInt{utils.MustParseBigInt("532013510291559900"), true},
					OutType:  int(types.StandardBalance),
				},
				&db.Balance{
					RecipientID: sql.NullInt64{6, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("20485680851457799"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutValue:    db.NullBigInt{utils.MustParseBigInt("20485680851457799"), true},
					OutType:     int(types.StandardBalance),
				},
				&db.Balance{
					RecipientID: sql.NullInt64{7, true},
					InValue:     db.NullBigInt{utils.MustParseBigInt("2276186761273088"), true},
					PoolFees:    db.NullBigInt{utils.MustParseBigInt("0"), true},
					OutValue:    db.NullBigInt{utils.MustParseBigInt("2276186761273088"), true},
					OutType:     int(types.StandardBalance),
				},
			},
			payouts: map[uint64]*db.Payout{
				1: &db.Payout{
					MinerID:        sql.NullInt64{1, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				2: &db.Payout{
					MinerID:        sql.NullInt64{2, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				6: &db.Payout{
					RecipientID:    sql.NullInt64{6, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
				7: &db.Payout{
					RecipientID:    sql.NullInt64{7, true},
					CoinID:         "ETH",
					Value:          db.NullBigInt{new(big.Int), true},
					PoolFees:       db.NullBigInt{new(big.Int), true},
					ExchangeFees:   db.NullBigInt{new(big.Int), true},
					FeeBalanceCoin: "ETH",
					InFeeBalance:   db.NullBigInt{new(big.Int), true},
					OutFeeBalance:  db.NullBigInt{new(big.Int), true},
				},
			},
		},
	}

	for i, tt := range tests {
		accountant := NewRoundAccountant(tt.ethRate)
		for _, round := range tt.rounds {
			if err := accountant.AddRound(round); err != nil {
				t.Errorf("failed on %d: AddRound: %v", i, err)
			}
		}

		if err := accountant.FinalizeSetRounds(); err != nil {
			t.Errorf("failed on %d: FinalizeSetRounds: %v", i, err)
		}

		if accountant.GetSpendableAmount().Cmp(tt.spendable) != 0 {
			t.Errorf("failed on %d: spendable mismatch: have %s, want %s",
				i, accountant.GetSpendableAmount(), tt.spendable)
		}

		if rounds, err := accountant.GetRounds(); err != nil {
			t.Errorf("failed on %d: GetRounds: %v", i, err)
		} else {
			for _, round := range rounds {
				if err := accountant.CreditRound(round, tt.shares[round.ID]); err != nil {
					t.Errorf("failed on %d: CreditRound %d: %v", i, round.ID, err)
				}
			}
		}

		if err := accountant.FinalizeCreditRounds(); err != nil {
			t.Errorf("failed on %d: FinalizeCreditRounds: %v", i, err)
		}

		for _, miner := range tt.miners {
			if err := accountant.CreditMiner(miner, tt.payouts[miner.ID]); err != nil {
				t.Errorf("failed on %d: CreditMiner %d: %v", i, miner.ID, err)
			}
		}

		if err := accountant.FinalizeCreditMiners(); err != nil {
			t.Errorf("failed on %d: FinalizeCreditMiners: %v", i, err)
		}

		for _, recipient := range tt.recipients {
			if err := accountant.CreditRecipient(recipient, tt.payouts[recipient.ID]); err != nil {
				t.Errorf("failed on %d: CreditRecipient %d: %v", i, recipient.ID, err)
			}
		}

		if err := accountant.FinalizeCreditRecipients(); err != nil {
			t.Errorf("failed on %d: FinalizeCreditRecipients: %v", i, err)
		}

		if err := accountant.ValidateBooks(); err != nil {
			t.Errorf("failed on %d: %v", i, err)
		}

		balances, err := accountant.GetBalances()
		if err != nil {
			t.Errorf("failed on %d: GetBalances: %v", i, err)
		}

		if len(balances) != len(tt.balances) {
			t.Errorf("failed on %d: balance count mismatch: have %d, want %d",
				i, len(balances), len(tt.balances))
			continue
		}

		for j, actualBalance := range balances {
			expectedBalance := tt.balances[j]

			// check miner id
			if actualBalance.MinerID != expectedBalance.MinerID {
				t.Errorf("failed on %d balance %d: MinerID mismatch: have %v, want %v",
					i, j, actualBalance.MinerID, expectedBalance.MinerID)
			}

			// check recipient id
			if actualBalance.RecipientID != expectedBalance.RecipientID {
				t.Errorf("failed on %d balance %d: RecipientID mismatch: have %v, want %v",
					i, j, actualBalance.RecipientID, expectedBalance.RecipientID)
			}

			// check in value
			if actualBalance.InValue.BigInt.Cmp(expectedBalance.InValue.BigInt) != 0 {
				t.Errorf("failed on %d balance %d: inValue mismatch: have %s, want %s",
					i, j, actualBalance.InValue.BigInt, expectedBalance.InValue.BigInt)
			}

			// check pool fees
			if actualBalance.PoolFees.BigInt.Cmp(expectedBalance.PoolFees.BigInt) != 0 {
				t.Errorf("failed on %d balance %d: poolFees mismatch: have %s, want %s",
					i, j, actualBalance.PoolFees.BigInt, expectedBalance.PoolFees.BigInt)
			}

			// check out type
			if actualBalance.OutType != expectedBalance.OutType {
				t.Errorf("failed on %d balance %d: outType mismatch: have %d, want %d",
					i, j, actualBalance.OutType, expectedBalance.OutType)
			}

			// check out value
			if actualBalance.OutValue.BigInt == nil && expectedBalance.OutValue.BigInt != nil {
				t.Errorf("failed on %d balance %d: outValue nil mismatch: have nil, want not nil",
					i, j)
			} else if actualBalance.OutValue.BigInt != nil && expectedBalance.OutValue.BigInt == nil {
				t.Errorf("failed on %d balance %d: outValue nil mismatch: have not nil, want nil",
					i, j)
			} else if actualBalance.OutValue.BigInt.Cmp(expectedBalance.OutValue.BigInt) != 0 {
				t.Errorf("failed on %d balance %d: outValue mismatch: have %s, want %s",
					i, j, actualBalance.OutValue.BigInt, expectedBalance.OutValue.BigInt)
			}
		}
	}
}

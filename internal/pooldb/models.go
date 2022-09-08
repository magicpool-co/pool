package pooldb

import (
	"time"

	"github.com/magicpool-co/pool/pkg/dbcl"
)

type Chain struct {
	ID string `json:"id"`

	Mineable   bool `json:"mineable"`
	Switchable bool `json:"switchable"`
	Payable    bool `json:"payable"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Node struct {
	ID uint64 `db:"id"`

	ChainID string  `db:"chain_id"`
	Region  string  `db:"region"`
	URL     string  `db:"url"`
	Version *string `db:"version"`

	Mainnet bool    `db:"mainnet"`
	Enabled bool    `db:"enabled"`
	Backup  bool    `db:"backup"`
	Active  bool    `db:"active"`
	Synced  bool    `db:"synced"`
	Height  *uint64 `db:"height"`

	NeedsBackup   bool `db:"needs_backup"`
	PendingBackup bool `db:"pending_backup"`
	NeedsUpdate   bool `db:"needs_update"`
	PendingUpdate bool `db:"pending_update"`
	NeedsResize   bool `db:"needs_resize"`
	PendingResize bool `db:"pending_resize"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DownAt    *time.Time `db:"down_at"`
	BackupAt  *time.Time `db:"backup_at"`
}

/* pool */

type Miner struct {
	ID      uint64 `db:"id"`
	ChainID string `db:"chain_id"`
	Address string `db:"address"`

	Active    bool       `db:"active"`
	LastLogin *time.Time `db:"last_login"`
	LastShare *time.Time `db:"last_share"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Worker struct {
	ID      uint64 `db:"id"`
	MinerID uint64 `db:"miner_id"`
	Name    string `db:"name"`

	Active    bool       `db:"active"`
	LastLogin *time.Time `db:"last_login"`
	LastShare *time.Time `db:"last_share"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type IPAddress struct {
	MinerID   uint64 `db:"miner_id"`
	IPAddress string `db:"ip_address"`

	Active    bool      `db:"active"`
	LastShare time.Time `db:"last_share"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Round struct {
	ID       uint64  `db:"id"`
	ChainID  string  `db:"chain_id"`
	MinerID  uint64  `db:"miner_id"`
	WorkerID *uint64 `db:"worker_id"`

	Height      uint64  `db:"height"`
	UncleHeight *uint64 `db:"uncle_height"`
	EpochHeight *uint64 `db:"epoch_height"`

	Hash         string          `db:"hash"`
	Nonce        *uint64         `db:"nonce"`
	MixDigest    *string         `db:"mix_digest"`
	Solution     *string         `db:"solution"`
	CoinbaseTxID *string         `db:"coinbase_txid"`
	Value        dbcl.NullBigInt `db:"value"`

	AcceptedShares uint64  `db:"accepted_shares"`
	RejectedShares uint64  `db:"rejected_shares"`
	InvalidShares  uint64  `db:"invalid_shares"`
	Difficulty     uint64  `db:"difficulty"`
	Luck           float32 `db:"luck"`

	Pending bool `db:"pending"`
	Uncle   bool `db:"uncle"`
	Orphan  bool `db:"orphan"`
	Mature  bool `db:"mature"`
	Spent   bool `db:"spent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Share struct {
	ID       uint64  `db:"id"`
	RoundID  uint64  `db:"round_id"`
	MinerID  uint64  `db:"miner_id"`
	WorkerID *uint64 `db:"worker_id"`

	Count uint64 `db:"count"`

	CreatedAt time.Time `db:"created_at"`
}

/* switch models */

type Switch struct {
	ID     uint64 `db:"id"`
	Status int    `db:"status"`

	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	CompletedAt sql.NullTime `db:"completed_at"`
}

type SwitchDeposit struct {
	ID       uint64 `db:"id"`
	CoinID   string `db:"coin_id"`
	SwitchID uint64 `db:"switch_id"`
	TxID     string `db:"txid"`

	BittrexID   sql.NullString `db:"bittrex_id"`
	BittrexTxID sql.NullString `db:"bittrex_txid"`

	Value          NullBigInt      `db:"value"`
	Fees           NullBigInt      `db:"fees"`
	CostBasisPrice sql.NullFloat64 `db:"cost_basis_price"`

	Registered bool `db:"registered"`
	Pending    bool `db:"pending"`
	Spent      bool `db:"spent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type SwitchTrade struct {
	ID     uint64 `db:"id"`
	PathID uint64 `db:"path_id"`
	Stage  int    `db:"stage"`

	NextTradeID sql.NullInt64  `db:"next_trade_id"`
	SwitchID    uint64         `db:"switch_id"`
	BittrexID   sql.NullString `db:"bittrex_id"`
	FromCoin    string         `db:"from_coin"`
	ToCoin      string         `db:"to_coin"`
	Market      string         `db:"market"`
	Direction   int            `db:"direction"`

	Value           NullBigInt      `db:"value"`
	Fees            NullBigInt      `db:"fees"`
	Proceeds        NullBigInt      `db:"proceeds"`
	Slippage        sql.NullFloat64 `db:"slippage"`
	CostBasisPrice  sql.NullFloat64 `db:"cost_basis_price"`
	FairMarketPrice sql.NullFloat64 `db:"fair_market_price"`

	Initiated bool `db:"initiated"`
	Open      bool `db:"open"`
	Filled    bool `db:"filled"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type SwitchWithdrawal struct {
	ID        uint64 `db:"id"`
	SwitchID  uint64 `db:"switch_id"`
	CoinID    string `db:"coin_id"`
	BittrexID string `db:"bittrex_id"`

	Value     NullBigInt `db:"value"`
	Fees      NullBigInt `db:"fees"`
	TradeFees NullBigInt `db:"trade_fees"`

	TxID      sql.NullString `db:"txid"`
	Height    sql.NullInt64  `db:"height"`
	Confirmed bool           `db:"confirmed"`
	Spent     bool           `db:"spent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

/* payout models */

type Payout struct {
	ID     uint64 `db:"id"`
	CoinID string `db:"coin_id"`

	MinerID     sql.NullInt64 `db:"miner_id"`
	RecipientID sql.NullInt64 `db:"recipient_id"`
	Address     string        `db:"address"`

	Value        NullBigInt `db:"value"`
	PoolFees     NullBigInt `db:"pool_fees"`
	ExchangeFees NullBigInt `db:"exchange_fees"`
	Spent        bool       `db:"spent"`
	Confirmed    bool       `db:"confirmed"`

	TxID   sql.NullString `db:"txid"`
	Height sql.NullInt64  `db:"height"`
	TxFees NullBigInt     `db:"tx_fees"`

	FeeBalanceCoin    string     `db:"fee_balance_coin"`
	FeeBalancePending bool       `db:"fee_balance_pending"`
	InFeeBalance      NullBigInt `db:"in_fee_balance"`
	OutFeeBalance     NullBigInt `db:"out_fee_balance"`

	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	SpentAt     sql.NullTime `db:"spent_at"`
	ConfirmedAt sql.NullTime `db:"confirmed_at"`
}

/* balance models */

type Balance struct {
	ID       uint64        `db:"id"`
	SwitchID sql.NullInt64 `db:"switch_id"`

	MinerID     sql.NullInt64 `db:"miner_id"`
	RecipientID sql.NullInt64 `db:"recipient_id"`

	InCoin      string        `db:"in_coin"`
	InValue     NullBigInt    `db:"in_value"`
	InRoundID   sql.NullInt64 `db:"in_round_id"`
	InDepositID sql.NullInt64 `db:"in_deposit_id"`
	InPayoutID  sql.NullInt64 `db:"in_payout_id"`

	Pending bool `db:"pending"`

	PoolFees     NullBigInt `db:"pool_fees"`
	ExchangeFees NullBigInt `db:"exchange_fees"`

	OutType     int        `db:"out_type"`
	OutCoin     string     `db:"out_coin"`
	OutValue    NullBigInt `db:"out_value"`
	OutPayoutID uint64     `db:"out_payout_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

/* utxo models */

type UTXO struct {
	ID     uint64 `db:"id"`
	CoinID string `db:"coin_id"`

	Value NullBigInt `db:"value"`
	Spent bool       `db:"spent"`

	InTxID         string        `db:"in_txid"`
	InIndex        uint32        `db:"in_index"`
	InRoundID      sql.NullInt64 `db:"in_round_id"`
	InDepositID    sql.NullInt64 `db:"in_deposit_id"`
	InWithdrawalID sql.NullInt64 `db:"in_withdrawal_id"`
	InPayoutID     sql.NullInt64 `db:"in_payout_id"`

	OutTxID      sql.NullString `db:"out_txid"`
	OutDepositID sql.NullInt64  `db:"out_deposit_id"`
	OutPayoutID  sql.NullInt64  `db:"out_payout_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

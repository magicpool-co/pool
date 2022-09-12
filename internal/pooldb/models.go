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
	Active  bool   `db:"active"`

	RecipientFeePercent *uint64 `db:"recipient_fee_percent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Worker struct {
	ID      uint64 `db:"id"`
	MinerID uint64 `db:"miner_id"`
	Name    string `db:"name"`
	Active  bool   `db:"active"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type IPAddress struct {
	ChainID   string `db:"chain_id"`
	MinerID   uint64 `db:"miner_id"`
	WorkerID  uint64 `db:"worker_id"`
	IPAddress string `db:"ip_address"`

	Active    bool      `db:"active"`
	LastShare time.Time `db:"last_share"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Round struct {
	ID      uint64 `db:"id"`
	ChainID string `db:"chain_id"`
	MinerID uint64 `db:"miner_id"`

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
	ID      uint64 `db:"id"`
	RoundID uint64 `db:"round_id"`
	MinerID uint64 `db:"miner_id"`

	Count uint64 `db:"count"`

	CreatedAt time.Time `db:"created_at"`
}

/* switch */

type Switch struct {
	ID         uint64 `db:"id"`
	ExchangeID string `db:"exchange_id"`
	Status     int    `db:"status"`

	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	CompletedAt *time.Time `db:"completed_at"`
}

type SwitchDeposit struct {
	ID        uint64 `db:"id"`
	SwitchID  uint64 `db:"switch_id"`
	ChainID   string `db:"chain_id"`
	NetworkID string `db:"network_id"`

	DepositTxID       string  `db:"deposit_txid"`
	ExchangeTxID      *string `db:"exchange_txid"`
	ExchangeDepositID *string `db:"exchange_deposit_id"`

	Value      dbcl.NullBigInt `db:"value"`
	Fees       dbcl.NullBigInt `db:"fees"`
	Registered bool            `db:"registered"`
	Pending    bool            `db:"pending"`
	Spent      bool            `db:"spent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type SwitchTrade struct {
	ID       uint64 `db:"id"`
	SwitchID uint64 `db:"switch_id"`
	PathID   uint64 `db:"path_id"`
	Stage    int    `db:"stage"`

	ExchangeTradeID *string `db:"exchange_trade_id"`
	NextTradeID     *uint64 `db:"next_trade_id"`

	FromChainID string `db:"from_chain_id"`
	ToChainID   string `db:"to_chain_id"`
	Market      string `db:"market"`
	Direction   int    `db:"direction"`

	Value     dbcl.NullBigInt `db:"value"`
	Fees      dbcl.NullBigInt `db:"fees"`
	Proceeds  dbcl.NullBigInt `db:"proceeds"`
	Slippage  *float64        `db:"slippage"`
	Initiated bool            `db:"initiated"`
	Open      bool            `db:"open"`
	Filled    bool            `db:"filled"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type SwitchWithdrawal struct {
	ID        uint64 `db:"id"`
	SwitchID  uint64 `db:"switch_id"`
	ChainID   string `db:"chain_id"`
	NetworkID string `db:"network_id"`

	ExchangeTxID         *string `db:"exchange_txid"`
	ExchangeWithdrawalID string  `db:"exchange_withdrawal_id"`

	Value          dbcl.NullBigInt `db:"value"`
	TradeFees      dbcl.NullBigInt `db:"trade_fees"`
	WithdrawalFees dbcl.NullBigInt `db:"withdrawal_fees"`
	Pending        bool            `db:"pending"`
	Spent          bool            `db:"spent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

/* balance */

type BalanceInput struct {
	ID      uint64 `db:"id"`
	RoundID uint64 `db:"round_id"`
	ChainID string `db:"chain_id"`
	MinerID uint64 `db:"miner_id"`

	OutputBalanceID *uint64 `db:"output_balance_id"`

	Value   dbcl.NullBigInt `db:"value"`
	Pending bool            `db:"pending"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type BalanceOutput struct {
	ID      uint64 `db:"id"`
	MinerID uint64 `db:"miner_id"`
	ChainID string `db:"chain_id"`

	InDepositID *uint64 `db:"in_deposit_id"`
	InPayoutID  *uint64 `db:"in_payout_id"`
	OutPayoutID *uint64 `db:"out_payout_id"`

	Value        dbcl.NullBigInt `db:"value"`
	PoolFees     dbcl.NullBigInt `db:"pool_fees"`
	ExchangeFees dbcl.NullBigInt `db:"exchange_fees"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Payout struct {
	ID      uint64 `db:"id"`
	MinerID uint64 `db:"miner_id"`
	ChainID string `db:"chain_id"`
	Address string `db:"address"`

	TxID   string  `db:"txid"`
	Height *uint64 `db:"height"`

	Value        dbcl.NullBigInt `db:"value"`
	PoolFees     dbcl.NullBigInt `db:"pool_fees"`
	ExchangeFees dbcl.NullBigInt `db:"exchange_fees"`
	TxFees       dbcl.NullBigInt `db:"tx_fees"`
	Confirmed    bool            `db:"confirmed"`
	Failed       bool            `db:"failed"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UTXO struct {
	ID      uint64 `db:"id"`
	ChainID string `db:"chain_id"`

	InTxID         string  `db:"in_txid"`
	InIndex        uint32  `db:"in_index"`
	InRoundID      *uint64 `db:"in_round_id"`
	InDepositID    *uint64 `db:"in_deposit_id"`
	InWithdrawalID *uint64 `db:"in_withdrawal_id"`
	InPayoutID     *uint64 `db:"in_payout_id"`

	OutTxID      *string `db:"out_txid"`
	OutDepositID *uint64 `db:"out_deposit_id"`
	OutPayoutID  *uint64 `db:"out_payout_id"`

	Value dbcl.NullBigInt `db:"value"`
	Spent bool            `db:"spent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

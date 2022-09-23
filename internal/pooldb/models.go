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
	URL string `db:"url"`

	ChainID string  `db:"chain_id"`
	Region  string  `db:"region"`
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

/* exchange */

type ExchangeBatch struct {
	ID         uint64 `db:"id"`
	ExchangeID int    `db:"exchange_id"`
	Status     int    `db:"status"`

	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	CompletedAt *time.Time `db:"completed_at"`
}

type ExchangeInput struct {
	ID            uint64 `db:"id"`
	BatchID       uint64 `db:"batch_id"`
	InputChainID  string `db:"input_chain_id"`
	OutputChainID string `db:"output_chain_id"`

	Value dbcl.NullBigInt `db:"value"`

	CreatedAt time.Time `db:"created_at"`
}

type ExchangeDeposit struct {
	ID        uint64 `db:"id"`
	BatchID   uint64 `db:"batch_id"`
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

type ExchangeTrade struct {
	ID      uint64 `db:"id"`
	BatchID uint64 `db:"batch_id"`
	Path    int    `db:"path"`
	Stage   int    `db:"stage"`

	ExchangeTradeID *string `db:"exchange_trade_id"`

	FromChainID string `db:"from_chain_id"`
	ToChainID   string `db:"to_chain_id"`
	Market      string `db:"market"`
	Direction   int    `db:"direction"`

	Value                 dbcl.NullBigInt `db:"value"`
	Proceeds              dbcl.NullBigInt `db:"proceeds"`
	TradeFees             dbcl.NullBigInt `db:"trade_fees"`
	CumulativeDepositFees dbcl.NullBigInt `db:"cumulative_deposit_fees"`
	CumulativeTradeFees   dbcl.NullBigInt `db:"cumulative_trade_fees"`

	OrderPrice *float64 `db:"order_price"`
	FillPrice  *float64 `db:"fill_price"`
	Slippage   *float64 `db:"slippage"`
	Initiated  bool     `db:"initiated"`
	Open       bool     `db:"open"`
	Filled     bool     `db:"filled"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ExchangeWithdrawal struct {
	ID        uint64 `db:"id"`
	BatchID   uint64 `db:"batch_id"`
	ChainID   string `db:"chain_id"`
	NetworkID string `db:"network_id"`

	ExchangeTxID         *string `db:"exchange_txid"`
	ExchangeWithdrawalID string  `db:"exchange_withdrawal_id"`

	Value          dbcl.NullBigInt `db:"value"`
	DepositFees    dbcl.NullBigInt `db:"deposit_fees"`
	TradeFees      dbcl.NullBigInt `db:"trade_fees"`
	WithdrawalFees dbcl.NullBigInt `db:"withdrawal_fees"`
	CumulativeFees dbcl.NullBigInt `db:"cumulative_fees"`
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

	OutputChainID   string  `db:"output_chain_id"`
	OutputBalanceID *uint64 `db:"output_balance_id"`
	ExchangeBatchID *uint64 `db:"exchange_batch_id"`

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

	Value dbcl.NullBigInt `db:"value"`
	TxID  string          `db:"txid"`
	Index uint32          `db:"index"`
	Spent bool            `db:"spent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

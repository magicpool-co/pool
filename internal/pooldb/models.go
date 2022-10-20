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
	ID        uint64          `db:"id"`
	ChainID   string          `db:"chain_id"`
	Address   string          `db:"address"`
	Threshold dbcl.NullBigInt `db:"threshold"`
	Active    bool            `db:"active"`

	RecipientFeePercent *uint64 `db:"recipient_fee_percent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	// column not present in the table, only
	// helpful for a specific join query (GetActiveWorkers)
	LastShare time.Time `db:"last_share"`
}

type Worker struct {
	ID      uint64 `db:"id"`
	MinerID uint64 `db:"miner_id"`
	Name    string `db:"name"`
	Active  bool   `db:"active"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	// column not present in the table, only
	// helpful for a specific join query (GetWorkersByMiner)
	LastShare time.Time `db:"last_share"`
}

type IPAddress struct {
	ChainID   string `db:"chain_id"`
	MinerID   uint64 `db:"miner_id"`
	WorkerID  uint64 `db:"worker_id"`
	IPAddress string `db:"ip_address"`

	Active    bool      `db:"active"`
	Expired   bool      `db:"expired"`
	LastShare time.Time `db:"last_share"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Round struct {
	ID      uint64 `db:"id"`
	ChainID string `db:"chain_id"`
	MinerID uint64 `db:"miner_id"`
	// column not present in the table, only
	// helpful for a specific join query (GetRounds)
	Miner *string `db:"miner"`

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
	Luck           float64 `db:"luck"`

	// columns not present in the table, only helpful for
	// a specific join query (GetRoundsByMiner)
	MinerAcceptedShares uint64          `db:"miner_accepted_shares"`
	MinerValue          dbcl.NullBigInt `db:"miner_value"`

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

/* utxo */

type UTXO struct {
	ID            uint64  `db:"id"`
	ChainID       string  `db:"chain_id"`
	TransactionID *uint64 `db:"transaction_id"`

	Value  dbcl.NullBigInt `db:"value"`
	TxID   string          `db:"txid"`
	Index  uint32          `db:"idx"`
	Active bool            `db:"active"`
	Spent  bool            `db:"spent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Transaction struct {
	ID      uint64 `db:"id"`
	ChainID string `db:"chain_id"`

	TxID   string  `db:"txid"`
	TxHex  string  `db:"tx_hex"`
	Height *uint64 `db:"height"`

	Value        dbcl.NullBigInt `db:"value"`
	Fee          dbcl.NullBigInt `db:"fee"`
	FeeBalance   dbcl.NullBigInt `db:"fee_balance"`
	Remainder    dbcl.NullBigInt `db:"remainder"`
	RemainderIdx uint32          `db:"remainder_idx"`
	Spent        bool            `db:"spent"`
	Confirmed    bool            `db:"confirmed"`
	Failed       bool            `db:"failed"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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
	ID         uint64 `db:"id"`
	BatchID    uint64 `db:"batch_id"`
	InChainID  string `db:"in_chain_id"`
	OutChainID string `db:"out_chain_id"`

	Value dbcl.NullBigInt `db:"value"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ExchangeDeposit struct {
	ID        uint64 `db:"id"`
	BatchID   uint64 `db:"batch_id"`
	ChainID   string `db:"chain_id"`
	NetworkID string `db:"network_id"`

	TransactionID     uint64  `db:"transaction_id"`
	DepositTxID       string  `db:"deposit_txid"`
	ExchangeTxID      *string `db:"exchange_txid"`
	ExchangeDepositID *string `db:"exchange_deposit_id"`

	Value      dbcl.NullBigInt `db:"value"`
	Fees       dbcl.NullBigInt `db:"fees"`
	Registered bool            `db:"registered"`
	Confirmed  bool            `db:"confirmed"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ExchangeTrade struct {
	ID      uint64 `db:"id"`
	BatchID uint64 `db:"batch_id"`
	PathID  int    `db:"path_id"`
	StageID int    `db:"stage_id"`

	ExchangeTradeID *string `db:"exchange_trade_id"`

	InitialChainID string `db:"initial_chain_id"`
	FromChainID    string `db:"from_chain_id"`
	ToChainID      string `db:"to_chain_id"`
	Market         string `db:"market"`
	Direction      int    `db:"direction"`

	Value                 dbcl.NullBigInt `db:"value"`
	Proceeds              dbcl.NullBigInt `db:"proceeds"`
	TradeFees             dbcl.NullBigInt `db:"trade_fees"`
	CumulativeDepositFees dbcl.NullBigInt `db:"cumulative_deposit_fees"`
	CumulativeTradeFees   dbcl.NullBigInt `db:"cumulative_trade_fees"`

	OrderPrice          *float64 `db:"order_price"`
	FillPrice           *float64 `db:"fill_price"`
	CumulativeFillPrice *float64 `db:"cumulative_fill_price"`
	Slippage            *float64 `db:"slippage"`
	Initiated           bool     `db:"initiated"`
	Confirmed           bool     `db:"confirmed"`

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
	Confirmed      bool            `db:"confirmed"`
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

	OutChainID      string  `db:"out_chain_id"`
	BalanceOutputID *uint64 `db:"balance_output_id"`
	BatchID         *uint64 `db:"batch_id"`

	Value    dbcl.NullBigInt `db:"value"`
	PoolFees dbcl.NullBigInt `db:"pool_fees"`
	Pending  bool            `db:"pending"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type BalanceOutput struct {
	ID      uint64 `db:"id"`
	ChainID string `db:"chain_id"`
	MinerID uint64 `db:"miner_id"`

	InBatchID   *uint64 `db:"in_batch_id"`
	InDepositID *uint64 `db:"in_deposit_id"`
	InPayoutID  *uint64 `db:"in_payout_id"`
	OutPayoutID *uint64 `db:"out_payout_id"`

	Value        dbcl.NullBigInt `db:"value"`
	PoolFees     dbcl.NullBigInt `db:"pool_fees"`
	ExchangeFees dbcl.NullBigInt `db:"exchange_fees"`
	Spent        bool            `db:"spent"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Payout struct {
	ID      uint64 `db:"id"`
	ChainID string `db:"chain_id"`
	MinerID uint64 `db:"miner_id"`
	Address string `db:"address"`

	TransactionID uint64  `db:"transaction_id"`
	TxID          string  `db:"txid"`
	Height        *uint64 `db:"height"`

	Value        dbcl.NullBigInt `db:"value"`
	FeeBalance   dbcl.NullBigInt `db:"fee_balance"`
	PoolFees     dbcl.NullBigInt `db:"pool_fees"`
	ExchangeFees dbcl.NullBigInt `db:"exchange_fees"`
	TxFees       dbcl.NullBigInt `db:"tx_fees"`
	Confirmed    bool            `db:"confirmed"`
	Failed       bool            `db:"failed"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

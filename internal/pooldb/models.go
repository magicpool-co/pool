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

package tsdb

import (
	"time"
)

type RawBlock struct {
	ID      uint64 `db:"id"`
	ChainID string `db:"chain_id"`

	Hash       string  `db:"hash"`
	Height     uint64  `db:"height"`
	Value      float64 `db:"value"`
	Difficulty float64 `db:"difficulty"`
	UncleCount uint64  `db:"uncle_count"`
	TxCount    uint64  `db:"tx_count"`

	Timestamp time.Time `db:"timestamp"`
}

type Block struct {
	ChainID string `db:"chain_id"`

	Value            float64 `db:"value"`
	Difficulty       float64 `db:"difficulty"`
	BlockTime        float64 `db:"block_time"`
	Hashrate         float64 `db:"hashrate"`
	UncleRate        float64 `db:"uncle_rate"`
	Profitability    float64 `db:"profitability"`
	AvgProfitability float64 `db:"avg_profitability"`

	// columns not present in the table,
	// only helpful for summary query
	ProfitabilityBTC    float64 `db:"profitability_btc"`
	AvgProfitabilityBTC float64 `db:"avg_profitability_btc"`

	Pending    bool      `db:"pending"`
	Count      uint64    `db:"count"`
	UncleCount uint64    `db:"uncle_count"`
	TxCount    uint64    `db:"tx_count"`
	Period     int       `db:"period"`
	StartTime  time.Time `db:"start_time"`
	EndTime    time.Time `db:"end_time"`
}

type Price struct {
	ChainID string `db:"chain_id"`

	PriceUSD float64 `db:"price_usd"`
	PriceBTC float64 `db:"price_btc"`
	PriceETH float64 `db:"price_eth"`

	Timestamp time.Time `db:"timestamp"`
}

type Round struct {
	ChainID string `db:"chain_id"`

	Value            float64 `db:"value"`
	Difficulty       float64 `db:"difficulty"`
	RoundTime        float64 `db:"round_time"`
	AcceptedShares   float64 `db:"accepted_shares"`
	RejectedShares   float64 `db:"rejected_shares"`
	InvalidShares    float64 `db:"invalid_shares"`
	Hashrate         float64 `db:"hashrate"`
	UncleRate        float64 `db:"uncle_rate"`
	Luck             float64 `db:"luck"`
	AvgLuck          float64 `db:"avg_luck"`
	Profitability    float64 `db:"profitability"`
	AvgProfitability float64 `db:"avg_profitability"`

	Pending    bool      `db:"pending"`
	Count      uint64    `db:"count"`
	UncleCount uint64    `db:"uncle_count"`
	Period     int       `db:"period"`
	StartTime  time.Time `db:"start_time"`
	EndTime    time.Time `db:"end_time"`
}

type Share struct {
	ChainID  string  `db:"chain_id"`
	MinerID  *uint64 `db:"miner_id"`
	WorkerID *uint64 `db:"worker_id"`

	Miners           uint64  `db:"miners"`
	Workers          uint64  `db:"workers"`
	AcceptedShares   uint64  `db:"accepted_shares"`
	RejectedShares   uint64  `db:"rejected_shares"`
	InvalidShares    uint64  `db:"invalid_shares"`
	Hashrate         float64 `db:"hashrate"`
	AvgHashrate      float64 `db:"avg_hashrate"`
	ReportedHashrate float64 `db:"reported_hashrate"`

	Pending   bool      `db:"pending"`
	Count     uint64    `db:"count"`
	Period    int       `db:"period"`
	StartTime time.Time `db:"start_time"`
	EndTime   time.Time `db:"end_time"`
}

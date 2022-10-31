package types

import (
	"fmt"
	"strings"
	"time"
)

/* node */

type AccountingType int

const (
	UTXOStructure AccountingType = iota
	AccountStructure
)

/* pool */

type ShareStatus int

const (
	AcceptedShare ShareStatus = iota
	RejectedShare
	InvalidShare
)

/* exchange */

type ExchangeID int

const (
	BinanceID ExchangeID = iota
	KucoinID
	BittrexID
)

type TradeDirection int

func (d TradeDirection) String() string {
	switch d {
	case TradeBuy:
		return "BUY"
	case TradeSell:
		return "SELL"
	default:
		return ""
	}
}

const (
	TradeBuy TradeDirection = iota
	TradeSell
)

/* transaction */

type TransactionType int

const (
	DepositTx TransactionType = iota
	PayoutTx
)

/* chart */

type PeriodType int

const (
	Period15m PeriodType = iota
	Period1h
	Period4h
	Period1d
)

func ParsePeriodType(raw string) (PeriodType, error) {
	switch strings.ToLower(raw) {
	case "15m":
		return Period15m, nil
	case "1h":
		return Period1h, nil
	case "4h":
		return Period4h, nil
	case "1d":
		return Period1d, nil
	default:
		return 0, fmt.Errorf("invalid period type")
	}
}

func (t PeriodType) AverageWindow() int {
	return int(t.Average() / t.Rollup())
}

func (t PeriodType) RetentionWindow() int {
	return int(t.Retention() / t.Rollup())
}

func (t PeriodType) Rollup() time.Duration {
	switch t {
	case 0:
		return time.Minute * 15
	case 1:
		return time.Hour
	case 2:
		return time.Hour * 4
	case 3:
		return time.Hour * 24
	default:
		return time.Minute
	}
}

func (t PeriodType) Average() time.Duration {
	switch t {
	case 0:
		return time.Hour * 4
	case 1:
		return time.Hour * 8
	case 2:
		return time.Hour * 24
	case 3:
		return time.Hour * 24 * 30
	default:
		return time.Hour
	}
}

func (t PeriodType) Retention() time.Duration {
	switch t {
	case 0:
		return time.Hour * 24
	case 1:
		return time.Hour * 24 * 7
	case 2:
		return time.Hour * 24 * 30
	case 3:
		return time.Hour * 24 * 365
	default:
		return time.Hour
	}
}

func (t PeriodType) GenerateRange(endTime time.Time) map[time.Time]bool {
	index := make(map[time.Time]bool)
	for i := t.RetentionWindow(); i >= 0; i-- {
		entry := endTime.Add(-t.Rollup() * time.Duration(i))
		index[entry] = false
	}

	return index
}

type NetworkMetric string

const (
	NetworkValue         NetworkMetric = "value"
	NetworkDifficulty    NetworkMetric = "difficulty"
	NetworkBlockTime     NetworkMetric = "block_time"
	NetworkHashrate      NetworkMetric = "hashrate"
	NetworkProfitability NetworkMetric = "profitability"
)

func ParseNetworkMetric(raw string) (NetworkMetric, error) {
	switch strings.ToLower(raw) {
	case "value":
		return NetworkValue, nil
	case "difficulty":
		return NetworkDifficulty, nil
	case "block_time":
		return NetworkBlockTime, nil
	case "hashrate":
		return NetworkHashrate, nil
	case "profitability":
		return NetworkProfitability, nil
	default:
		return "", fmt.Errorf("unknown metric type")
	}
}

type ShareMetric string

const (
	ShareHashrate         ShareMetric = "hashrate"
	ShareAverageHashrate  ShareMetric = "avg_hashrate"
	ShareReportedHashrate ShareMetric = "reported_hashrate"
)

func ParseShareMetric(raw string) (ShareMetric, error) {
	switch strings.ToLower(raw) {
	case "hashrate":
		return ShareHashrate, nil
	case "avg_hashrate":
		return ShareAverageHashrate, nil
	case "reported_hashrate":
		return ShareReportedHashrate, nil
	default:
		return "", fmt.Errorf("unknown metric type")
	}
}

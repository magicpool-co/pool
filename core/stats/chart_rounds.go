package stats

import (
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/types"
)

type Round struct {
	Timestamp        time.Time `json:"timestamp"`
	Miners           uint64    `json:"miners"`
	Workers          uint64    `json:"workers"`
	AcceptedShares   uint64    `json:"acceptedShares"`
	RejectedShares   uint64    `json:"rejectedShares"`
	InvalidShares    uint64    `json:"invalidShares"`
	Hashrate         float64   `json:"hashrate"`
	AvgHashrate      float64   `json:"avgHashrate"`
	ReportedHashrate float64   `json:"reportedHashrate"`
}

func processRawRounds(rawRounds []*tsdb.Round) [][]interface{} {
	return nil
}

func (c *Client) GetRoundCharts(chain string, period types.PeriodType) (interface{}, error) {
	raw, err := tsdb.GetRounds(c.tsdb.Reader(), chain, int(period))
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"keys":   nil,
		"points": processRawRounds(raw),
	}

	return data, nil
}

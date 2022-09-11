package charter

import (
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/dbcl"
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

func processRawRounds(rawRounds []*tsdb.Round) []*Round {
	return nil
}

func FetchRounds(tsdbClient *dbcl.Client, period types.PeriodType) (map[string][]*Round, error) {
	idx := make(map[string][]*Round)
	for _, chain := range chains {
		raw, err := tsdb.GetRounds(tsdbClient.Reader(), chain, int(period))
		if err != nil {
			return nil, err
		}

		idx[chain] = processRawRounds(raw)
	}

	return idx, nil
}

package charter

import (
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type Round struct {
	Miners           uint64  `db:"miners"`
	Workers          uint64  `db:"workers"`
	AcceptedShares   uint64  `db:"acceptedShares"`
	RejectedShares   uint64  `db:"rejectedShares"`
	InvalidShares    uint64  `db:"invalidShares"`
	Hashrate         float64 `db:"hashrate"`
	AvgHashrate      float64 `db:"avgHashrate"`
	ReportedHashrate float64 `db:"reportedHashrate"`

	Timestamp time.Time `db:"timestamp"`
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

package charter

import (
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type Block struct {
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

func processRawBlocks(rawBlocks []*tsdb.Block) []*Block {
	return nil
}

func FetchBlocks(tsdbClient *dbcl.Client, period types.PeriodType) (map[string][]*Block, error) {
	idx := make(map[string][]*Block)
	for _, chain := range chains {
		rawBlocks, err := tsdb.GetBlocks(tsdbClient.Reader(), chain, int(period))
		if err != nil {
			return nil, err
		}

		idx[chain] = processRawBlocks(rawBlocks)
	}

	return idx, nil
}

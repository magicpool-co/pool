package charter

import (
	"sort"
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type Share struct {
	Timestamp        int64   `json:"timestamp"`
	Miners           uint64  `json:"miners"`
	Workers          uint64  `json:"workers"`
	AcceptedShares   uint64  `json:"acceptedShares"`
	RejectedShares   uint64  `json:"rejectedShares"`
	InvalidShares    uint64  `json:"invalidShares"`
	Hashrate         float64 `json:"hashrate"`
	AvgHashrate      float64 `json:"avgHashrate"`
	ReportedHashrate float64 `json:"reportedHashrate"`
}

func processRawShares(items []*tsdb.Share, period types.PeriodType) []*Share {
	var index map[time.Time]bool
	if len(items) == 0 {
		index = period.GenerateRange(time.Now())
	} else {
		endTime := items[0].EndTime
		if newEndTime := items[len(items)-1].EndTime; newEndTime.After(endTime) {
			endTime = newEndTime
		}

		index = period.GenerateRange(endTime)
	}

	shares := make([]*Share, 0)
	for _, item := range items {
		if exists := index[item.EndTime]; exists {
			continue
		}

		index[item.EndTime] = true
		share := &Share{
			Timestamp:        item.EndTime.Unix(),
			Miners:           item.Miners,
			Workers:          item.Workers,
			AcceptedShares:   item.AcceptedShares,
			RejectedShares:   item.RejectedShares,
			InvalidShares:    item.InvalidShares,
			Hashrate:         processFloat(item.Hashrate),
			AvgHashrate:      processFloat(item.AvgHashrate),
			ReportedHashrate: processFloat(item.ReportedHashrate),
		}

		shares = append(shares, share)
	}

	for timestamp, exists := range index {
		if !exists {
			shares = append(shares, &Share{Timestamp: timestamp.Unix()})
		}
	}

	sort.Slice(shares, func(i, j int) bool {
		return shares[i].Timestamp < shares[j].Timestamp
	})

	return shares
}

func FetchGlobalShares(tsdbClient *dbcl.Client, period types.PeriodType) (map[string][]*Share, error) {
	idx := make(map[string][]*Share)
	for _, chain := range chains {
		raw, err := tsdb.GetGlobalShares(tsdbClient.Reader(), chain, int(period))
		if err != nil {
			return nil, err
		}

		idx[chain] = processRawShares(raw, period)
	}

	return idx, nil
}

func FetchMinerShares(tsdbClient *dbcl.Client, minerID uint64, period types.PeriodType) (map[string][]*Share, error) {
	idx := make(map[string][]*Share)
	for _, chain := range chains {
		raw, err := tsdb.GetMinerShares(tsdbClient.Reader(), minerID, chain, int(period))
		if err != nil {
			return nil, err
		}

		idx[chain] = processRawShares(raw, period)
	}

	return idx, nil
}

func FetchWorkerShares(tsdbClient *dbcl.Client, workerID uint64, period types.PeriodType) (map[string][]*Share, error) {
	idx := make(map[string][]*Share)
	for _, chain := range chains {
		raw, err := tsdb.GetWorkerShares(tsdbClient.Reader(), workerID, chain, int(period))
		if err != nil {
			return nil, err
		}

		idx[chain] = processRawShares(raw, period)
	}

	return idx, nil
}

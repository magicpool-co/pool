package charter

import (
	"sort"
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

var shareKeys = []string{
	"timestamp",
	"miners",
	"workers",
	"acceptedShares",
	"rejectedShares",
	"invalidShares",
	"hashrate",
	"avgHashrate",
	"reportedHashrate",
}

func convertShare(item *tsdb.Share) []interface{} {
	data := []interface{}{
		item.EndTime.Unix(),
		item.Miners,
		item.Workers,
		item.AcceptedShares,
		item.RejectedShares,
		item.InvalidShares,
		processFloat(item.Hashrate),
		processFloat(item.AvgHashrate),
		processFloat(item.ReportedHashrate),
	}

	return data
}

func processRawShares(items []*tsdb.Share, period types.PeriodType) [][]interface{} {
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

	shares := make([][]interface{}, 0)
	for _, item := range items {
		if exists := index[item.EndTime]; !exists {
			shares = append(shares, convertShare(item))
			index[item.EndTime] = true
		}
	}

	for timestamp, exists := range index {
		if !exists {
			shares = append(shares, convertShare(&tsdb.Share{EndTime: timestamp}))
		}
	}

	sort.Slice(shares, func(i, j int) bool {
		return shares[i][0].(int64) < shares[j][0].(int64)
	})

	return shares
}

func FetchGlobalShares(tsdbClient *dbcl.Client, period types.PeriodType) (interface{}, error) {
	idx := make(map[string][][]interface{})
	for _, chain := range chains {
		raw, err := tsdb.GetGlobalShares(tsdbClient.Reader(), chain, int(period))
		if err != nil {
			return nil, err
		}

		idx[chain] = processRawShares(raw, period)
	}

	data := map[string]interface{}{
		"keys":   shareKeys,
		"chains": idx,
	}

	return data, nil
}

func FetchMinerShares(tsdbClient *dbcl.Client, minerID uint64, period types.PeriodType) (interface{}, error) {
	idx := make(map[string][][]interface{})
	for _, chain := range chains {
		raw, err := tsdb.GetMinerShares(tsdbClient.Reader(), minerID, chain, int(period))
		if err != nil {
			return nil, err
		}

		idx[chain] = processRawShares(raw, period)
	}

	data := map[string]interface{}{
		"keys":   shareKeys,
		"chains": idx,
	}

	return data, nil
}

func FetchWorkerShares(tsdbClient *dbcl.Client, workerID uint64, period types.PeriodType) (interface{}, error) {
	idx := make(map[string][][]interface{})
	for _, chain := range chains {
		raw, err := tsdb.GetWorkerShares(tsdbClient.Reader(), workerID, chain, int(period))
		if err != nil {
			return nil, err
		}

		idx[chain] = processRawShares(raw, period)
	}

	data := map[string]interface{}{
		"keys":   shareKeys,
		"chains": idx,
	}

	return data, nil
}

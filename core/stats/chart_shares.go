package stats

import (
	"sort"
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
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
		common.SafeRoundedFloat(item.Hashrate),
		common.SafeRoundedFloat(item.AvgHashrate),
		common.SafeRoundedFloat(item.ReportedHashrate),
	}

	return data
}

func processRawShares(items []*tsdb.Share, period types.PeriodType) [][]interface{} {
	var endTime time.Time
	if len(items) == 0 {
		endTime = time.Now()
	} else {
		endTime = items[0].EndTime
		if newEndTime := items[len(items)-1].EndTime; newEndTime.After(endTime) {
			endTime = newEndTime
		}
	}

	index := period.GenerateRange(common.NormalizeDate(endTime, period.Rollup(), true))
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

func (c *Client) GetGlobalShareCharts(chain string, period types.PeriodType) (interface{}, error) {
	raw, err := tsdb.GetGlobalShares(c.tsdb.Reader(), chain, int(period))
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"keys": shareKeys,
		"data": processRawShares(raw, period),
	}

	return data, nil
}

func (c *Client) GetMinerShareCharts(minerIDs []uint64, chain string, period types.PeriodType) (interface{}, error) {
	if len(minerIDs) == 0 {
		return nil, nil
	}

	raw, err := tsdb.GetMinerShares(c.tsdb.Reader(), minerIDs[0], chain, int(period))
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"keys": shareKeys,
		"data": processRawShares(raw, period),
	}

	return data, nil
}

func (c *Client) GetWorkerShareCharts(workerID uint64, chain string, period types.PeriodType) (interface{}, error) {
	raw, err := tsdb.GetWorkerShares(c.tsdb.Reader(), workerID, chain, int(period))
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"keys":   shareKeys,
		"points": processRawShares(raw, period),
	}

	return data, nil
}

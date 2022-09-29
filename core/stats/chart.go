package stats

import (
	"sort"
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

/* block chart */

func (c *Client) GetBlockChart(chain string, period types.PeriodType) (*BlockChart, error) {
	items, err := tsdb.GetBlocks(c.tsdb.Reader(), chain, int(period))
	if err != nil {
		return nil, err
	}

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
	chart := &BlockChart{
		Timestamp:        make([]int64, 0),
		Value:            make([]float64, 0),
		Difficulty:       make([]float64, 0),
		BlockTime:        make([]float64, 0),
		Hashrate:         make([]float64, 0),
		UncleRate:        make([]float64, 0),
		Profitability:    make([]float64, 0),
		AvgProfitability: make([]float64, 0),
		BlockCount:       make([]uint64, 0),
		UncleCount:       make([]uint64, 0),
		TxCount:          make([]uint64, 0),
	}

	for _, item := range items {
		if exists := index[item.EndTime]; !exists {
			chart.AddPoint(item)
			index[item.EndTime] = true
		}
	}

	for timestamp, exists := range index {
		if !exists {
			chart.AddPoint(&tsdb.Block{EndTime: timestamp})
		}
	}

	sort.Sort(chart)

	return chart, nil
}

/* round chart */

func (c *Client) GetRoundChart(chain string, period types.PeriodType) (*RoundChart, error) {
	items, err := tsdb.GetRounds(c.tsdb.Reader(), chain, int(period))
	if err != nil {
		return nil, err
	}

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
	chart := &RoundChart{
		Timestamp:        make([]int64, 0),
		Value:            make([]float64, 0),
		Difficulty:       make([]float64, 0),
		RoundTime:        make([]float64, 0),
		Hashrate:         make([]float64, 0),
		UncleRate:        make([]float64, 0),
		Luck:             make([]float64, 0),
		AvgLuck:          make([]float64, 0),
		Profitability:    make([]float64, 0),
		AvgProfitability: make([]float64, 0),
	}

	for _, item := range items {
		if exists := index[item.EndTime]; !exists {
			chart.AddPoint(item)
			index[item.EndTime] = true
		}
	}

	for timestamp, exists := range index {
		if !exists {
			chart.AddPoint(&tsdb.Round{EndTime: timestamp})
		}
	}

	sort.Sort(chart)

	return chart, nil
}

/* share chart */

func getShareChart(items []*tsdb.Share, period types.PeriodType) *ShareChart {
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
	chart := &ShareChart{
		Timestamp:        make([]int64, 0),
		Miners:           make([]uint64, 0),
		Workers:          make([]uint64, 0),
		AcceptedShares:   make([]uint64, 0),
		RejectedShares:   make([]uint64, 0),
		InvalidShares:    make([]uint64, 0),
		Hashrate:         make([]float64, 0),
		AvgHashrate:      make([]float64, 0),
		ReportedHashrate: make([]float64, 0),
	}

	for _, item := range items {
		if exists := index[item.EndTime]; !exists {
			chart.AddPoint(item)
			index[item.EndTime] = true
		}
	}

	for timestamp, exists := range index {
		if !exists {
			chart.AddPoint(&tsdb.Share{EndTime: timestamp})
		}
	}

	sort.Sort(chart)

	return chart
}

func (c *Client) GetGlobalShareChart(chain string, period types.PeriodType) (*ShareChart, error) {
	items, err := tsdb.GetGlobalShares(c.tsdb.Reader(), chain, int(period))
	if err != nil {
		return nil, err
	}

	return getShareChart(items, period), nil
}

func (c *Client) GetMinerShareChart(minerIDs []uint64, chain string, period types.PeriodType) (*ShareChart, error) {
	if len(minerIDs) == 0 {
		return nil, nil
	}

	items, err := tsdb.GetMinerShares(c.tsdb.Reader(), minerIDs[0], chain, int(period))
	if err != nil {
		return nil, err
	}

	return getShareChart(items, period), nil
}

func (c *Client) GetWorkerShareChart(workerID uint64, chain string, period types.PeriodType) (*ShareChart, error) {
	items, err := tsdb.GetWorkerShares(c.tsdb.Reader(), workerID, chain, int(period))
	if err != nil {
		return nil, err
	}

	return getShareChart(items, period), nil
}

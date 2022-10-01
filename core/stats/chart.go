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

func (c *Client) GetBlockProfitabilityChart(period types.PeriodType) (map[string]*BlockChart, error) {
	items, err := tsdb.GetBlocksProfitability(c.tsdb.Reader(), int(period))
	if err != nil {
		return nil, err
	}

	itemsIdx := make(map[string][]*tsdb.Block)
	for _, item := range items {
		if _, ok := itemsIdx[item.ChainID]; !ok {
			itemsIdx[item.ChainID] = make([]*tsdb.Block, 0)
		}
		itemsIdx[item.ChainID] = append(itemsIdx[item.ChainID], item)
	}

	chartIdx := make(map[string]*BlockChart)
	for chain, items := range itemsIdx {
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
		chartIdx[chain] = &BlockChart{
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
				chartIdx[chain].AddPoint(item)
				index[item.EndTime] = true
			}
		}

		for timestamp, exists := range index {
			if !exists {
				chartIdx[chain].AddPoint(&tsdb.Block{EndTime: timestamp})
			}
		}

		sort.Sort(chartIdx[chain])

		chartIdx[chain].Value = nil
		chartIdx[chain].Difficulty = nil
		chartIdx[chain].BlockTime = nil
		chartIdx[chain].Hashrate = nil
		chartIdx[chain].UncleRate = nil
		chartIdx[chain].BlockCount = nil
		chartIdx[chain].UncleCount = nil
		chartIdx[chain].TxCount = nil
	}

	return chartIdx, nil
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

func sumShares(items []*tsdb.Share) []*tsdb.Share {
	idx := make(map[time.Time]*tsdb.Share)
	duplicateIdx := make(map[time.Time][]*tsdb.Share)
	for _, item := range items {
		if _, ok := idx[item.EndTime]; !ok {
			idx[item.EndTime] = item
			continue
		} else if _, ok := duplicateIdx[item.EndTime]; !ok {
			duplicateIdx[item.EndTime] = make([]*tsdb.Share, 0)
		}

		duplicateIdx[item.EndTime] = append(duplicateIdx[item.EndTime], item)
	}

	for timestamp, duplicateItems := range duplicateIdx {
		for _, item := range duplicateItems {
			idx[timestamp].Miners += item.Miners
			idx[timestamp].Workers += item.Workers
			idx[timestamp].AcceptedShares += item.AcceptedShares
			idx[timestamp].RejectedShares += item.RejectedShares
			idx[timestamp].InvalidShares += item.InvalidShares
			idx[timestamp].Hashrate += item.Hashrate
			idx[timestamp].AvgHashrate += item.AvgHashrate
			idx[timestamp].ReportedHashrate += item.ReportedHashrate
		}
	}

	var count int
	uniqueItems := make([]*tsdb.Share, len(idx))
	for _, item := range idx {
		uniqueItems[count] = item
		count++
	}

	sort.Slice(uniqueItems, func(i, j int) bool {
		return uniqueItems[i].EndTime.Before(uniqueItems[j].EndTime)
	})

	return uniqueItems
}

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
	items, err := tsdb.GetMinerShares(c.tsdb.Reader(), minerIDs, chain, int(period))
	if err != nil {
		return nil, err
	}

	// sum shares if more than one miner
	if len(minerIDs) > 1 {
		items = sumShares(items)
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

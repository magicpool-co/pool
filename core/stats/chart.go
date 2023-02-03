package stats

import (
	"fmt"
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

	var firstTime, lastTime time.Time
	for _, item := range items {
		if exists := index[item.EndTime]; !exists {
			chart.AddPoint(item)
			index[item.EndTime] = true

			if firstTime.IsZero() || item.EndTime.Before(firstTime) {
				firstTime = item.EndTime
			}
			if lastTime.IsZero() || item.EndTime.Before(lastTime) {
				lastTime = item.EndTime
			}
		}
	}

	for timestamp, exists := range index {
		if !exists && !timestamp.Before(firstTime) && !timestamp.After(lastTime) {
			chart.AddPoint(&tsdb.Block{EndTime: timestamp})
		}
	}

	sort.Sort(chart)

	return chart, nil
}

func (c *Client) GetBlockSingleMetricChart(metric types.NetworkMetric, period types.PeriodType, average bool) (*ChartSingle, error) {
	var items []*tsdb.Block
	var err error
	switch metric {
	case types.NetworkProfitability:
		items, err = tsdb.GetBlocksProfitability(c.tsdb.Reader(), int(period))
	case types.NetworkValue:
		items, err = tsdb.GetBlocksAdjustedValue(c.tsdb.Reader(), int(period))
	case types.NetworkEmission:
		items, err = tsdb.GetBlocksAdjustedEmission(c.tsdb.Reader(), int(period))
	default:
		items, err = tsdb.GetBlocksSingleMetric(c.tsdb.Reader(), string(metric), int(period))
	}
	if err != nil {
		return nil, err
	}

	itemsIdx := make(map[time.Time]map[string]*tsdb.Block)
	chainIdx := make(map[string]time.Time)
	for _, item := range items {
		if !c.chains[item.ChainID] {
			continue
		} else if _, ok := itemsIdx[item.EndTime]; !ok {
			itemsIdx[item.EndTime] = make(map[string]*tsdb.Block)
		}
		itemsIdx[item.EndTime][item.ChainID] = item

		if item.EndTime.After(chainIdx[item.ChainID]) {
			chainIdx[item.ChainID] = item.EndTime
		}
	}

	var startTime, endTime time.Time
	for _, timestamp := range chainIdx {
		if startTime.IsZero() || timestamp.Before(startTime) {
			startTime = timestamp
		}
		if endTime.IsZero() || timestamp.Before(endTime) {
			endTime = timestamp
		}
	}

	if endTime.IsZero() {
		endTime = time.Now()
	}

	index := period.GenerateRange(common.NormalizeDate(endTime, period.Rollup(), true))
	for timestamp := range index {
		if _, ok := itemsIdx[timestamp]; !ok && !timestamp.Before(startTime) {
			itemsIdx[timestamp] = make(map[string]*tsdb.Block)
		}
	}

	for timestamp, chainIdx := range itemsIdx {
		if _, ok := index[timestamp]; !ok {
			delete(itemsIdx, timestamp)
			continue
		}

		for chain := range chainIdx {
			if _, ok := itemsIdx[timestamp][chain]; !ok {
				itemsIdx[timestamp][chain] = &tsdb.Block{EndTime: timestamp}
			}
		}
	}

	var count int
	timestamps := make([]time.Time, len(itemsIdx))
	for timestamp := range itemsIdx {
		timestamps[count] = timestamp
		count++
	}

	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	values := make(map[string][]float64)
	for chain := range chainIdx {
		values[chain] = make([]float64, len(timestamps))
	}

	for i, timestamp := range timestamps {
		for chain, item := range itemsIdx[timestamp] {
			var value float64
			switch metric {
			case types.NetworkValue:
				value = item.Value
			case types.NetworkDifficulty:
				value = item.Difficulty
			case types.NetworkBlockTime:
				value = item.BlockTime
			case types.NetworkHashrate:
				value = item.Hashrate
			case types.NetworkProfitability:
				value = item.Profitability
				if average {
					value = item.AvgProfitability
				}
			default:
				return nil, fmt.Errorf("unknown metric type")
			}

			if value == 0 && i > 0 {
				value = values[chain][i-1]
			}

			values[chain][i] = value
		}
	}

	parsedTimestamps := make([]int64, len(timestamps))
	for i, timestamp := range timestamps {
		parsedTimestamps[i] = timestamp.Unix()
	}

	chart := &ChartSingle{
		Timestamps: parsedTimestamps,
		Values:     values,
	}

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

	var firstTime, lastTime time.Time
	for _, item := range items {
		if exists := index[item.EndTime]; !exists {
			chart.AddPoint(item)
			index[item.EndTime] = true

			if firstTime.IsZero() || item.EndTime.Before(firstTime) {
				firstTime = item.EndTime
			}
			if lastTime.IsZero() || item.EndTime.Before(lastTime) {
				lastTime = item.EndTime
			}
		}
	}

	for timestamp, exists := range index {
		if !exists && !timestamp.Before(firstTime) && !timestamp.After(lastTime) {
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

func sumSharesSingle(metric types.ShareMetric, items []*tsdb.Share) ([]*tsdb.Share, error) {
	compoundIdx := make(map[time.Time]map[string]*tsdb.Share)
	compoundDuplicateIdx := make(map[time.Time]map[string][]*tsdb.Share)
	for _, item := range items {
		if _, ok := compoundIdx[item.EndTime]; !ok {
			compoundIdx[item.EndTime] = make(map[string]*tsdb.Share)
			compoundDuplicateIdx[item.EndTime] = make(map[string][]*tsdb.Share)
		}

		if _, ok := compoundIdx[item.EndTime][item.ChainID]; !ok {
			compoundIdx[item.EndTime][item.ChainID] = item
			continue
		} else if _, ok := compoundDuplicateIdx[item.EndTime][item.ChainID]; !ok {
			compoundDuplicateIdx[item.EndTime][item.ChainID] = make([]*tsdb.Share, 0)
		}

		compoundDuplicateIdx[item.EndTime][item.ChainID] = append(compoundDuplicateIdx[item.EndTime][item.ChainID], item)
	}

	for timestamp, duplicateIdx := range compoundDuplicateIdx {
		for chain, duplicateItems := range duplicateIdx {
			for _, item := range duplicateItems {
				switch metric {
				case types.ShareHashrate:
					compoundIdx[timestamp][chain].Hashrate += item.Hashrate
				case types.ShareAverageHashrate:
					compoundIdx[timestamp][chain].AvgHashrate += item.AvgHashrate
				case types.ShareReportedHashrate:
					compoundIdx[timestamp][chain].ReportedHashrate += item.ReportedHashrate
				default:
					return nil, fmt.Errorf("unknown metric type")
				}
			}
		}
	}

	uniqueItems := make([]*tsdb.Share, 0)
	for _, idx := range compoundIdx {
		for _, item := range idx {
			uniqueItems = append(uniqueItems, item)
		}
	}

	sort.Slice(uniqueItems, func(i, j int) bool {
		return uniqueItems[i].EndTime.Before(uniqueItems[j].EndTime)
	})

	return uniqueItems, nil
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

	var firstTime, lastTime time.Time
	for _, item := range items {
		if exists := index[item.EndTime]; !exists {
			chart.AddPoint(item)
			index[item.EndTime] = true

			if firstTime.IsZero() || item.EndTime.Before(firstTime) {
				firstTime = item.EndTime
			}
			if lastTime.IsZero() || item.EndTime.Before(lastTime) {
				lastTime = item.EndTime
			}
		}
	}

	for timestamp, exists := range index {
		if !exists && !timestamp.Before(firstTime) && !timestamp.After(lastTime) {
			chart.AddPoint(&tsdb.Share{EndTime: timestamp})
		}
	}

	sort.Sort(chart)

	return chart
}

func getShareChartSingle(metric types.ShareMetric, items []*tsdb.Share, period types.PeriodType) (*ChartSingle, error) {
	itemsIdx := make(map[time.Time]map[string]*tsdb.Share)
	chainIdx := make(map[string]time.Time)
	for _, item := range items {
		if _, ok := itemsIdx[item.EndTime]; !ok {
			itemsIdx[item.EndTime] = make(map[string]*tsdb.Share)
		}
		itemsIdx[item.EndTime][item.ChainID] = item

		if item.EndTime.After(chainIdx[item.ChainID]) {
			chainIdx[item.ChainID] = item.EndTime
		}
	}

	var startTime, endTime time.Time
	for _, timestamp := range chainIdx {
		if startTime.IsZero() || timestamp.Before(startTime) {
			startTime = timestamp
		}
		// if endTime.IsZero() || timestamp.Before(endTime) {
		if endTime.IsZero() || timestamp.After(endTime) {
			endTime = timestamp
		}
	}

	if endTime.IsZero() {
		endTime = time.Now()
	}

	index := period.GenerateRange(common.NormalizeDate(endTime, period.Rollup(), true))
	for timestamp := range index {
		if _, ok := itemsIdx[timestamp]; !ok && !timestamp.Before(startTime) {
			itemsIdx[timestamp] = make(map[string]*tsdb.Share)
		}
	}

	for timestamp, chainIdx := range itemsIdx {
		if _, ok := index[timestamp]; !ok {
			delete(itemsIdx, timestamp)
			continue
		}

		for chain := range chainIdx {
			if _, ok := itemsIdx[timestamp][chain]; !ok {
				itemsIdx[timestamp][chain] = &tsdb.Share{EndTime: timestamp}
			}
		}
	}

	var count int
	timestamps := make([]time.Time, len(itemsIdx))
	for timestamp := range itemsIdx {
		timestamps[count] = timestamp
		count++
	}

	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	values := make(map[string][]float64)
	for chain := range chainIdx {
		values[chain] = make([]float64, len(timestamps))
	}

	for i, timestamp := range timestamps {
		for chain, item := range itemsIdx[timestamp] {
			var value float64
			switch metric {
			case types.ShareHashrate:
				value = item.Hashrate
			case types.ShareAverageHashrate:
				value = item.AvgHashrate
			case types.ShareReportedHashrate:
				value = item.ReportedHashrate
			default:
				return nil, fmt.Errorf("unknown metric type")
			}

			if value == 0 && i > 0 {
				value = values[chain][i-1]
			}

			values[chain][i] = value
		}
	}

	parsedTimestamps := make([]int64, len(timestamps))
	for i, timestamp := range timestamps {
		parsedTimestamps[i] = timestamp.Unix()
	}

	chart := &ChartSingle{
		Timestamps: parsedTimestamps,
		Values:     values,
	}

	return chart, nil
}

func (c *Client) GetGlobalShareChart(chain string, period types.PeriodType) (*ShareChart, error) {
	items, err := tsdb.GetGlobalShares(c.tsdb.Reader(), chain, int(period))
	if err != nil {
		return nil, err
	}

	return getShareChart(items, period), nil
}

func (c *Client) GetGlobalShareSingleMetricChart(metric types.ShareMetric, period types.PeriodType) (*ChartSingle, error) {
	items, err := tsdb.GetGlobalSharesSingleMetric(c.tsdb.Reader(), string(metric), int(period))
	if err != nil {
		return nil, err
	}

	return getShareChartSingle(metric, items, period)
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

func (c *Client) GetMinerShareSingleMetricChart(minerIDs []uint64, metric types.ShareMetric, period types.PeriodType) (*ChartSingle, error) {
	items, err := tsdb.GetMinerSharesSingleMetric(c.tsdb.Reader(), minerIDs, string(metric), int(period))
	if err != nil {
		return nil, err
	}

	// sum shares if more than one miner
	if len(minerIDs) > 1 {
		items, err = sumSharesSingle(metric, items)
		if err != nil {
			return nil, err
		}
	}

	return getShareChartSingle(metric, items, period)
}

func (c *Client) GetWorkerShareChart(workerID uint64, chain string, period types.PeriodType) (*ShareChart, error) {
	items, err := tsdb.GetWorkerShares(c.tsdb.Reader(), workerID, chain, int(period))
	if err != nil {
		return nil, err
	}

	return getShareChart(items, period), nil
}

func (c *Client) GetWorkerShareSingleMetricChart(workerID uint64, metric types.ShareMetric, period types.PeriodType) (*ChartSingle, error) {
	items, err := tsdb.GetWorkerSharesSingleMetric(c.tsdb.Reader(), workerID, string(metric), int(period))
	if err != nil {
		return nil, err
	}

	return getShareChartSingle(metric, items, period)
}

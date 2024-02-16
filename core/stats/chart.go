package stats

import (
	"fmt"
	"sort"
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

type rangeItem struct {
	ChainID   string
	Timestamp time.Time
	Data      interface{}
}

func (c *Client) normalizeRange(
	items []rangeItem,
	period types.PeriodType,
) ([]time.Time, map[string]time.Time, map[time.Time]map[string]rangeItem) {
	itemsIdx := make(map[time.Time]map[string]rangeItem)
	chainIdx := make(map[string]time.Time)
	for _, item := range items {
		var ok bool
		item.ChainID, ok = c.processChainID(item.ChainID)
		if !ok {
			continue
		}

		if _, ok := itemsIdx[item.Timestamp]; !ok {
			itemsIdx[item.Timestamp] = make(map[string]rangeItem)
		}
		itemsIdx[item.Timestamp][item.ChainID] = item

		if item.Timestamp.After(chainIdx[item.ChainID]) {
			chainIdx[item.ChainID] = item.Timestamp
		}
	}

	var startTime, endTime time.Time
	for _, timestamp := range chainIdx {
		if startTime.IsZero() || timestamp.Before(startTime) {
			startTime = timestamp
		}
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
			itemsIdx[timestamp] = make(map[string]rangeItem)
		}
	}

	for timestamp, chainIdx := range itemsIdx {
		if _, ok := index[timestamp]; !ok {
			delete(itemsIdx, timestamp)
			continue
		}

		for chain := range chainIdx {
			if _, ok := itemsIdx[timestamp][chain]; !ok {
				itemsIdx[timestamp][chain] = rangeItem{Timestamp: timestamp}
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

	return timestamps, chainIdx, itemsIdx
}

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

func (c *Client) GetBlockSingleMetricChart(
	metric types.NetworkMetric,
	period types.PeriodType,
	average bool,
) (*ChartSingle, error) {
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

	rangeItems := make([]rangeItem, len(items))
	for i, item := range items {
		rangeItems[i] = rangeItem{
			ChainID:   item.ChainID,
			Timestamp: item.EndTime,
			Data:      item,
		}
	}

	timestamps, chainIdx, rangeItemsIdx := c.normalizeRange(rangeItems, period)
	values := make(map[string][]float64)
	for chain := range chainIdx {
		values[chain] = make([]float64, len(timestamps))
	}

	for i, timestamp := range timestamps {
		for chain, rangeItem := range rangeItemsIdx[timestamp] {
			if rangeItem.Data == nil {
				continue
			}

			item, ok := rangeItem.Data.(*tsdb.Block)
			if !ok {
				continue
			}
			var value float64
			switch metric {
			case types.NetworkValue, types.NetworkEmission:
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

/* share chart */

func sumShares(items []*tsdb.Share) []*tsdb.Share {
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
				compoundIdx[timestamp][chain].Miners += item.Miners
				compoundIdx[timestamp][chain].Workers += item.Workers
				compoundIdx[timestamp][chain].AcceptedAdjustedShares += item.AcceptedAdjustedShares
				compoundIdx[timestamp][chain].RejectedAdjustedShares += item.RejectedAdjustedShares
				compoundIdx[timestamp][chain].InvalidAdjustedShares += item.InvalidAdjustedShares
				compoundIdx[timestamp][chain].Hashrate += item.Hashrate
				compoundIdx[timestamp][chain].AvgHashrate += item.AvgHashrate
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
		Timestamp:      make([]int64, 0),
		Miners:         make([]uint64, 0),
		Workers:        make([]uint64, 0),
		AcceptedShares: make([]uint64, 0),
		RejectedShares: make([]uint64, 0),
		InvalidShares:  make([]uint64, 0),
		Hashrate:       make([]float64, 0),
		AvgHashrate:    make([]float64, 0),
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

func (c *Client) getShareChartSingle(
	metric types.ShareMetric,
	items []*tsdb.Share,
	period types.PeriodType,
) (*ChartSingle, error) {
	rangeItems := make([]rangeItem, len(items))
	for i, item := range items {
		rangeItems[i] = rangeItem{
			ChainID:   item.ChainID,
			Timestamp: item.EndTime,
			Data:      item,
		}
	}

	timestamps, chainIdx, rangeItemsIdx := c.normalizeRange(rangeItems, period)
	values := make(map[string][]float64)
	for chain := range chainIdx {
		values[chain] = make([]float64, len(timestamps))
	}

	for i, timestamp := range timestamps {
		for chain, rangeItem := range rangeItemsIdx[timestamp] {
			if rangeItem.Data == nil {
				continue
			}

			item, ok := rangeItem.Data.(*tsdb.Share)
			if !ok {
				continue
			}

			var buffValues bool
			var value float64
			switch metric {
			case types.ShareHashrate:
				value = item.Hashrate
				buffValues = true
			case types.ShareAverageHashrate:
				value = item.AvgHashrate
				buffValues = true
			case types.ShareAcceptedCount:
				value = float64(item.AcceptedAdjustedShares)
			case types.ShareRejectedCount:
				value = float64(item.RejectedAdjustedShares)
			case types.ShareRejectedRate:
				if item.AcceptedShares > 0 {
					denominator := float64(item.AcceptedAdjustedShares + item.RejectedAdjustedShares)
					value = 100 * (float64(item.RejectedAdjustedShares) / denominator)
				}
			default:
				return nil, fmt.Errorf("unknown metric type")
			}

			if buffValues && value == 0 && i > 0 {
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

func (c *Client) GetGlobalShareChart(
	chain string,
	period types.PeriodType,
) (*ShareChart, error) {
	items, err := tsdb.GetGlobalShares(c.tsdb.Reader(), chain, int(period))
	if err != nil {
		return nil, err
	}

	return getShareChart(items, period), nil
}

func (c *Client) GetGlobalShareSingleMetricChart(
	metric types.ShareMetric,
	period types.PeriodType,
) (*ChartSingle, error) {
	items, err := tsdb.GetGlobalSharesSingleMetric(c.tsdb.Reader(), string(metric), int(period))
	if err != nil {
		return nil, err
	}

	return c.getShareChartSingle(metric, items, period)
}

func (c *Client) GetMinerShareChart(
	minerIDs []uint64,
	chain string,
	period types.PeriodType,
) (*ShareChart, error) {
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

func (c *Client) GetMinerShareSingleMetricChart(
	minerIDs []uint64,
	metric types.ShareMetric,
	period types.PeriodType,
) (*ChartSingle, error) {
	items, err := tsdb.GetMinerSharesSingleMetric(c.tsdb.Reader(), minerIDs, string(metric), int(period))
	if err != nil {
		return nil, err
	}

	// sum shares if more than one miner
	if len(minerIDs) > 1 {
		items = sumShares(items)
	}

	return c.getShareChartSingle(metric, items, period)
}

func (c *Client) GetWorkerShareChart(
	workerID uint64,
	chain string,
	period types.PeriodType,
) (*ShareChart, error) {
	items, err := tsdb.GetWorkerShares(c.tsdb.Reader(), workerID, chain, int(period))
	if err != nil {
		return nil, err
	}

	return getShareChart(items, period), nil
}

func (c *Client) GetWorkerShareSingleMetricChart(
	workerID uint64,
	metric types.ShareMetric,
	period types.PeriodType,
) (*ChartSingle, error) {
	items, err := tsdb.GetWorkerSharesSingleMetric(c.tsdb.Reader(), workerID, string(metric), int(period))
	if err != nil {
		return nil, err
	}

	return c.getShareChartSingle(metric, items, period)
}

/* share chart */

func sumEarnings(items []*tsdb.Earning) ([]*tsdb.Earning, error) {
	compoundIdx := make(map[time.Time]map[string]*tsdb.Earning)
	compoundDuplicateIdx := make(map[time.Time]map[string][]*tsdb.Earning)
	for _, item := range items {
		if _, ok := compoundIdx[item.EndTime]; !ok {
			compoundIdx[item.EndTime] = make(map[string]*tsdb.Earning)
			compoundDuplicateIdx[item.EndTime] = make(map[string][]*tsdb.Earning)
		}

		if _, ok := compoundIdx[item.EndTime][item.ChainID]; !ok {
			compoundIdx[item.EndTime][item.ChainID] = item
			continue
		} else if _, ok := compoundDuplicateIdx[item.EndTime][item.ChainID]; !ok {
			compoundDuplicateIdx[item.EndTime][item.ChainID] = make([]*tsdb.Earning, 0)
		}

		compoundDuplicateIdx[item.EndTime][item.ChainID] = append(compoundDuplicateIdx[item.EndTime][item.ChainID], item)
	}

	for timestamp, duplicateIdx := range compoundDuplicateIdx {
		for chain, duplicateItems := range duplicateIdx {
			for _, item := range duplicateItems {
				compoundIdx[timestamp][chain].Value += item.Value
				compoundIdx[timestamp][chain].AvgValue += item.AvgValue
			}
		}
	}

	uniqueItems := make([]*tsdb.Earning, 0)
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

func (c *Client) getEarningChartSingle(
	items []*tsdb.Earning,
	period types.PeriodType,
) (*ChartSingle, error) {
	rangeItems := make([]rangeItem, len(items))
	for i, item := range items {
		rangeItems[i] = rangeItem{
			ChainID:   item.ChainID,
			Timestamp: item.EndTime,
			Data:      item,
		}
	}

	timestamps, chainIdx, rangeItemsIdx := c.normalizeRange(rangeItems, period)
	values := make(map[string][]float64)
	for chain := range chainIdx {
		values[chain] = make([]float64, len(timestamps))
	}

	for i, timestamp := range timestamps {
		for chain, rangeItem := range rangeItemsIdx[timestamp] {
			if rangeItem.Data == nil {
				continue
			}

			item, ok := rangeItem.Data.(*tsdb.Earning)
			if !ok {
				continue
			}

			values[chain][i] = item.Value
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

func (c *Client) GetGlobalEarningChart(period types.PeriodType) (*ChartSingle, error) {
	items, err := tsdb.GetGlobalEarningsSingleMetric(c.tsdb.Reader(), "value", int(period))
	if err != nil {
		return nil, err
	}

	return c.getEarningChartSingle(items, period)
}

func (c *Client) GetMinerEarningChart(
	minerIDs []uint64,
	metric types.EarningMetric,
	period types.PeriodType,
) (*ChartSingle, error) {
	items, err := tsdb.GetMinerEarningsSingleMetric(c.tsdb.Reader(), minerIDs, string(metric), int(period))
	if err != nil {
		return nil, err
	}

	// sum shares if more than one miner
	if len(minerIDs) > 1 {
		items, err = sumEarnings(items)
		if err != nil {
			return nil, err
		}
	}

	return c.getEarningChartSingle(items, period)
}

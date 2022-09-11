package charter

import (
	"sort"
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type Block struct {
	Timestamp        int64   `json:"timestamp"`
	Value            float64 `json:"value"`
	Difficulty       float64 `json:"difficulty"`
	BlockTime        float64 `json:"blockTime"`
	Hashrate         float64 `json:"hashrate"`
	UncleRate        float64 `json:"uncleRate"`
	Profitability    float64 `json:"profitability"`
	AvgProfitability float64 `json:"avgProfitability"`
	BlockCount       uint64  `json:"blockCount"`
	UncleCount       uint64  `json:"uncleCount"`
	TxCount          uint64  `json:"txCount"`
}

func processRawBlocks(items []*tsdb.Block, period types.PeriodType) []*Block {
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

	blocks := make([]*Block, 0)
	for _, item := range items {
		if exists := index[item.EndTime]; exists {
			continue
		}

		index[item.EndTime] = true
		block := &Block{
			Timestamp:        item.EndTime.Unix(),
			Value:            processFloat(item.Value),
			Difficulty:       processFloat(item.Difficulty),
			BlockTime:        processFloat(item.BlockTime),
			Hashrate:         processFloat(item.Hashrate),
			UncleRate:        processFloat(item.UncleRate),
			Profitability:    processFloat(item.Profitability),
			AvgProfitability: processFloat(item.AvgProfitability),
			BlockCount:       item.Count,
			UncleCount:       item.UncleCount,
			TxCount:          item.TxCount,
		}

		blocks = append(blocks, block)
	}

	for timestamp, exists := range index {
		if !exists {
			blocks = append(blocks, &Block{Timestamp: timestamp.Unix()})
		}
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Timestamp < blocks[j].Timestamp
	})

	return blocks
}

func FetchBlocks(tsdbClient *dbcl.Client, period types.PeriodType) (map[string][]*Block, error) {
	idx := make(map[string][]*Block)
	for _, chain := range chains {
		raw, err := tsdb.GetBlocks(tsdbClient.Reader(), chain, int(period))
		if err != nil {
			return nil, err
		}

		idx[chain] = processRawBlocks(raw, period)
	}

	return idx, nil
}

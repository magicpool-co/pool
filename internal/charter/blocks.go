package charter

import (
	"sort"
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

var blockKeys = []string{
	"timestamp",
	"value",
	"difficulty",
	"blockTime",
	"hashrate",
	"uncleRate",
	"profitability",
	"avgProfitability",
	"blockCount",
	"uncleCount",
	"txCount",
}

func convertBlock(block *tsdb.Block) []interface{} {
	data := []interface{}{
		block.EndTime.Unix(),
		common.SafeRoundedFloat(block.Value),
		common.SafeRoundedFloat(block.Difficulty),
		common.SafeRoundedFloat(block.BlockTime),
		common.SafeRoundedFloat(block.Hashrate),
		common.SafeRoundedFloat(block.UncleRate),
		common.SafeRoundedFloat(block.Profitability),
		common.SafeRoundedFloat(block.AvgProfitability),
		block.Count,
		block.UncleCount,
		block.TxCount,
	}

	return data
}

func processRawBlocks(items []*tsdb.Block, period types.PeriodType) [][]interface{} {
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
	blocks := make([][]interface{}, 0)
	for _, item := range items {
		if exists := index[item.EndTime]; !exists {
			blocks = append(blocks, convertBlock(item))
			index[item.EndTime] = true
		}
	}

	for timestamp, exists := range index {
		if !exists {
			blocks = append(blocks, convertBlock(&tsdb.Block{EndTime: timestamp}))
		}
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i][0].(int64) < blocks[j][0].(int64)
	})

	return blocks
}

func FetchBlocks(tsdbClient *dbcl.Client, chain string, period types.PeriodType) (interface{}, error) {
	raw, err := tsdb.GetBlocks(tsdbClient.Reader(), chain, int(period))
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"keys":   blockKeys,
		"points": processRawBlocks(raw, period),
	}

	return data, nil
}

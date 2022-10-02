package chart

import (
	"fmt"
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

var (
	blockStart         = time.Unix(1661990400, 0)
	blockDelay         = time.Minute * 5
	blockPeriod        = types.Period15m
	blockRollupPeriods = []types.PeriodType{types.Period4h, types.Period1d}
)

func (c *Client) rollupBlocks(node types.MiningNode, endTime time.Time) error {
	lastTime, err := tsdb.GetBlockMaxEndTime(c.tsdb.Reader(), node.Chain(), int(blockPeriod))
	if err != nil {
		return err
	} else if !lastTime.Before(endTime) {
		return nil
	}

	startTime := endTime.Add(blockPeriod.Rollup() * -1)
	block, err := tsdb.GetRawBlockRollup(c.tsdb.Reader(), node.Chain(), startTime, endTime)
	if err != nil {
		return err
	} else if block == nil {
		block = &tsdb.Block{
			ChainID: node.Chain(),
		}
	}

	block.Period = int(blockPeriod)
	block.StartTime = startTime
	block.EndTime = endTime
	block.Pending = false
	if block.Count > 0 {
		block.UncleRate = float64(block.UncleCount) / float64(block.Count+block.UncleCount)
	}
	block.Hashrate = node.CalculateHashrate(block.BlockTime, block.Difficulty)
	if block.Hashrate > 0 {
		block.Profitability = (block.Value / block.BlockTime) / block.Hashrate
	}
	block.AvgProfitability, err = tsdb.GetBlocksAverageSlow(c.tsdb.Reader(), block.EndTime, node.Chain(), int(blockPeriod), blockPeriod.Average())
	if err != nil {
		return err
	}

	tx, err := c.tsdb.Begin()
	if err != nil {
		return err
	}
	defer tx.SafeRollback()

	err = tsdb.InsertBlocks(tx, block)
	if err != nil {
		return err
	}

	block.Pending = true
	block.Value *= float64(block.Count)
	block.Difficulty *= float64(block.Count)
	block.BlockTime *= float64(block.Count)
	block.Hashrate *= float64(block.Count)
	block.Profitability *= float64(block.Count)
	block.AvgProfitability = 0

	for _, rollupPeriod := range blockRollupPeriods {
		block.Period = int(rollupPeriod)
		block.StartTime = common.NormalizeDate(block.StartTime, rollupPeriod.Rollup(), true)
		block.EndTime = common.NormalizeDate(block.StartTime, rollupPeriod.Rollup(), false)

		err = tsdb.InsertPartialBlocks(tx, block)
		if err != nil {
			return err
		}
	}

	return tx.SafeCommit()
}

func (c *Client) finalizeBlocks(node types.MiningNode, endTime time.Time) error {
	for _, rollupPeriod := range blockRollupPeriods {
		// finalize summed statistics
		blocks, err := tsdb.GetPendingBlocksAtEndTime(c.tsdb.Reader(), endTime, node.Chain(), int(rollupPeriod))
		if err != nil {
			return err
		}

		for _, block := range blocks {
			block.Pending = false
			if block.Count > 0 {
				block.Value /= float64(block.Count)
				block.Difficulty /= float64(block.Count)
				block.BlockTime /= float64(block.Count)
				block.UncleRate = float64(block.UncleCount) / float64(block.Count+block.UncleCount)
			}
			block.Hashrate = node.CalculateHashrate(block.BlockTime, block.Difficulty)
			if block.Hashrate > 0 {
				block.Profitability = (block.Value / block.BlockTime) / block.Hashrate
			}
		}

		if err := tsdb.InsertFinalBlocks(c.tsdb.Writer(), blocks...); err != nil {
			return err
		}

		// finalize averages after updated statistics
		for _, block := range blocks {
			block.AvgProfitability, err = tsdb.GetBlocksAverageSlow(c.tsdb.Reader(), block.EndTime,
				node.Chain(), int(rollupPeriod), rollupPeriod.Average())
			if err != nil {
				return err
			}
		}

		if err := tsdb.InsertFinalBlocks(c.tsdb.Writer(), blocks...); err != nil {
			return err
		}

	}

	return nil
}

func (c *Client) truncateBlocks(node types.MiningNode, endTime time.Time) error {
	cutoffTime, err := tsdb.GetRawBlockMaxTimestampBeforeTime(c.tsdb.Reader(), node.Chain(), endTime)
	if err != nil {
		return err
	} else if !cutoffTime.IsZero() {
		err = tsdb.DeleteRawBlocksBeforeTime(c.tsdb.Writer(), node.Chain(), cutoffTime)
		if err != nil {
			return err
		}
	}

	for _, rollupPeriod := range append([]types.PeriodType{blockPeriod}, blockRollupPeriods...) {
		timestamp := endTime.Add(rollupPeriod.Retention() * -1)
		err = tsdb.DeleteBlocksBeforeEndTime(c.tsdb.Writer(), timestamp, node.Chain(), int(rollupPeriod))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) CollectBlocks(node types.MiningNode) error {
	lastHeight, err := tsdb.GetRawBlockMaxHeight(c.tsdb.Reader(), node.Chain())
	if err != nil {
		return err
	}

	currentHeight, syncing, err := node.GetStatus()
	if err != nil {
		return err
	} else if syncing {
		return fmt.Errorf("node is syncing")
	} else if lastHeight == 0 {
		switch node.Chain() {
		case "CFX":
			lastHeight = 53100000
		case "CTXC":
			lastHeight = 6840000
		case "ETC":
			lastHeight = 15850000
		case "FLUX":
			lastHeight = 1198000
		case "FIRO":
			lastHeight = 530000
		case "RVN":
			lastHeight = 2432500
		default:
			lastHeight = currentHeight - 1
		}
	} else if lastHeight >= currentHeight {
		return nil
	}

	switch node.Chain() {
	case "CFX":
		if currentHeight-lastHeight > 10000 {
			currentHeight = lastHeight + 10000
		}
	case "CTXC", "ETC":
		if currentHeight-lastHeight > 1000 {
			currentHeight = lastHeight + 1000
		}
	case "FLUX", "FIRO", "RVN":
		if currentHeight-lastHeight > 500 {
			currentHeight = lastHeight + 500
		}
	}

	blocks, err := node.GetBlocks(lastHeight+1, currentHeight)
	if err != nil {
		return err
	} else if len(blocks) == 0 {
		return nil
	}

	err = tsdb.InsertRawBlocks(c.tsdb.Writer(), blocks...)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) FetchBlockIntervals(chain string) ([]time.Time, error) {
	lastTime, err := c.redis.GetChartBlocksLastTime(chain)
	if err != nil || lastTime.IsZero() {
		lastTime, err = tsdb.GetBlockMaxEndTime(c.tsdb.Reader(), chain, int(blockPeriod))
		if err != nil {
			return nil, err
		}
	}

	// handle initialization of intervals when there are no blocks in tsdb
	if lastTime.IsZero() {
		lastTime = blockStart
		err := c.redis.SetChartBlocksLastTime(chain, lastTime)

		return nil, err
	}

	// protect against charting when no blocks actually exist
	lastRawTime, err := tsdb.GetRawBlockMaxTimestamp(c.tsdb.Reader(), chain)
	if err != nil {
		return nil, err
	} else if lastRawTime.IsZero() || lastTime.Sub(lastRawTime) > blockDelay {
		return nil, nil
	}

	intervals := make([]time.Time, 0)
	for {
		endTime := lastTime.Add(blockPeriod.Rollup())
		if time.Now().UTC().Sub(endTime) < blockDelay {
			break
		}

		intervals = append(intervals, endTime)
		lastTime = endTime
	}

	if len(intervals) > 25 {
		intervals = intervals[:25]
	}

	return intervals, nil
}

func (c *Client) ProcessBlocks(timestamp time.Time, node types.MiningNode) error {
	err := c.rollupBlocks(node, timestamp)
	if err != nil {
		return fmt.Errorf("rollup: %s: %v", node.Chain(), err)
	}

	err = c.finalizeBlocks(node, timestamp)
	if err != nil {
		return fmt.Errorf("finalize: %s: %v", node.Chain(), err)
	}

	err = c.truncateBlocks(node, timestamp)
	if err != nil {
		return fmt.Errorf("truncate: %s: %v", node.Chain(), err)
	}

	err = c.redis.SetChartBlocksLastTime(node.Chain(), timestamp)
	if err != nil {
		return fmt.Errorf("set: %s: %v", node.Chain(), err)
	}

	return nil
}

package worker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bsm/redislock"
	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

var (
	blockDelay         = time.Minute * 5
	blockPeriod        = types.Period15m
	blockRollupPeriods = []types.PeriodType{types.Period4h, types.Period1d}
)

type ChartBlockJob struct {
	locker *redislock.Client
	logger *log.Logger
	redis  *redis.Client
	tsdb   *dbcl.Client
	nodes  []types.MiningNode
}

func getMarketRate(chain string, timestamp time.Time) float64 {
	const baseURL = "https://api.coingecko.com/api/v3"
	var tickers = map[string]string{
		"AE":   "aeternity",
		"BTC":  "bitcoin",
		"CFX":  "conflux-token",
		"CTXC": "cortex",
		"ETC":  "ethereum-classic",
		"ETH":  "ethereum",
		"FIRO": "zcoin",
		"FLUX": "flux",
		"RVN":  "ravencoin",
	}

	ticker, ok := tickers[chain]
	if !ok {
		return 0
	}

	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", baseURL, ticker)
	res, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer res.Body.Close()

	data := make(map[string]map[string]float64)
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return 0
	} else if _, ok := data[ticker]; !ok {
		return 0
	}

	return data[ticker]["usd"]
}

func (j *ChartBlockJob) fetchIntervals(chain string) ([]time.Time, error) {
	lastTime, err := j.redis.GetChartBlocksLastTime(chain)
	if err != nil || lastTime.IsZero() {
		lastTime, err = tsdb.GetBlockMaxEndTime(j.tsdb.Reader(), chain, int(blockPeriod))
		if err != nil {
			return nil, err
		}
	}

	// handle initialization of intervals when there are no blocks in tsdb
	if lastTime.IsZero() {
		lastTime = common.NormalizeDate(time.Now(), blockPeriod.Rollup(), false)
		err := j.redis.SetChartBlocksLastTime(chain, lastTime)

		return nil, err
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

	return intervals, nil
}

func (j *ChartBlockJob) collect(node types.MiningNode) error {
	lastHeight, err := tsdb.GetRawBlockMaxHeight(j.tsdb.Reader(), node.Chain())
	if err != nil {
		return err
	}

	currentHeight, syncing, err := node.GetStatus()
	if err != nil {
		return err
	} else if syncing {
		return fmt.Errorf("node is syncing")
	} else if lastHeight == 0 {
		lastHeight = currentHeight - 1
	} else if lastHeight >= currentHeight {
		return nil
	}

	blocks, err := node.GetBlocks(lastHeight+1, currentHeight)
	if err != nil {
		return err
	} else if len(blocks) == 0 {
		return nil
	}

	err = tsdb.InsertRawBlocks(j.tsdb.Writer(), blocks...)
	if err != nil {
		return err
	}

	return nil
}

func (j *ChartBlockJob) rollup(node types.MiningNode, endTime time.Time) error {
	lastTime, err := tsdb.GetBlockMaxEndTime(j.tsdb.Reader(), node.Chain(), int(blockPeriod))
	if err != nil {
		return err
	} else if !lastTime.Before(endTime) {
		return nil
	}

	startTime := endTime.Add(blockPeriod.Rollup() * -1)
	block, err := tsdb.GetRawBlockRollup(j.tsdb.Reader(), node.Chain(), startTime, endTime)
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
		marketRate := getMarketRate(node.Chain(), block.EndTime)
		block.Profitability = marketRate * (block.Value / block.BlockTime) / block.Hashrate
	}
	block.AvgProfitability, err = tsdb.GetBlocksAverageSlow(j.tsdb.Reader(), block.EndTime, node.Chain(), int(blockPeriod), blockPeriod.Average())
	if err != nil {
		return err
	}

	tx, err := j.tsdb.Begin()
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

func (j *ChartBlockJob) finalize(node types.MiningNode, endTime time.Time) error {
	for _, rollupPeriod := range blockRollupPeriods {
		// finalize summed statistics
		blocks, err := tsdb.GetPendingBlocksAtEndTime(j.tsdb.Reader(), endTime, node.Chain(), int(rollupPeriod))
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
				marketRate := getMarketRate(node.Chain(), block.EndTime)
				block.Profitability = marketRate * (block.Value / block.BlockTime) / block.Hashrate
			}
		}

		if err := tsdb.InsertFinalBlocks(j.tsdb.Writer(), blocks...); err != nil {
			return err
		}

		// finalize averages after updated statistics
		for _, block := range blocks {
			block.AvgProfitability, err = tsdb.GetBlocksAverageSlow(j.tsdb.Reader(), block.EndTime,
				node.Chain(), int(rollupPeriod), rollupPeriod.Average())
			if err != nil {
				return err
			}
		}

		if err := tsdb.InsertFinalBlocks(j.tsdb.Writer(), blocks...); err != nil {
			return err
		}

	}

	return nil
}

func (j *ChartBlockJob) truncate(node types.MiningNode, endTime time.Time) error {
	cutoffTime, err := tsdb.GetRawBlockMaxTimestampBeforeTime(j.tsdb.Reader(), node.Chain(), endTime)
	if err != nil {
		return err
	} else if !cutoffTime.IsZero() {
		err = tsdb.DeleteRawBlocksBeforeTime(j.tsdb.Writer(), node.Chain(), cutoffTime)
		if err != nil {
			return err
		}
	}

	for _, rollupPeriod := range blockRollupPeriods {
		timestamp := endTime.Add(rollupPeriod.Retention() * -1)
		err = tsdb.DeleteBlocksBeforeEndTime(j.tsdb.Writer(), timestamp, node.Chain(), int(rollupPeriod))
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *ChartBlockJob) Run() {
	defer recoverPanic(j.logger)

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:chrtblk", time.Minute*30, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	for _, node := range j.nodes {
		if err := j.collect(node); err != nil {
			j.logger.Error(fmt.Errorf("block: collect: %s: %v", node.Chain(), err))
			continue
		}

		intervals, err := j.fetchIntervals(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("block: interval: %s: %v", node.Chain(), err))
			continue
		}

		for _, interval := range intervals {
			if err := j.rollup(node, interval); err != nil {
				j.logger.Error(fmt.Errorf("block: rollup: %s: %v", node.Chain(), err))
				break
			}

			if err := j.finalize(node, interval); err != nil {
				j.logger.Error(fmt.Errorf("block: finalize: %s: %v", node.Chain(), err))
				break
			}

			if err := j.truncate(node, interval); err != nil {
				j.logger.Error(fmt.Errorf("block: truncate: %s: %v", node.Chain(), err))
				break
			}

			if err := j.redis.SetChartBlocksLastTime(node.Chain(), interval); err != nil {
				j.logger.Error(fmt.Errorf("block: delete: %s: %v", node.Chain(), err))
				break
			}
		}
	}
}

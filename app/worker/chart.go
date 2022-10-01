package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/core/chart"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

var (
	blockDelay         = time.Minute * 5
	blockPeriod        = types.Period15m
	blockRollupPeriods = []types.PeriodType{types.Period4h, types.Period1d}
)

type ChartJob struct {
	locker *redislock.Client
	logger *log.Logger
	redis  *redis.Client
	pooldb *dbcl.Client
	tsdb   *dbcl.Client
	nodes  []types.MiningNode
}

func (j *ChartJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:chart", time.Minute*30, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	client := chart.New(j.pooldb, j.tsdb, j.redis)

	// shares
	for _, node := range j.nodes {
		intervals, err := client.FetchShareIntervals(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("share: interval: %s: %v", node.Chain(), err))
			continue
		}

		for _, interval := range intervals {
			err := client.ProcessShares(interval, node)
			if err != nil {
				j.logger.Error(fmt.Errorf("share: %v", err))
				break
			}
		}
	}

	// blocks
	for _, node := range j.nodes {
		err := client.CollectBlocks(node)
		if err != nil {
			j.logger.Error(fmt.Errorf("block: collect: %s: %v", node.Chain(), err))
			continue
		}

		intervals, err := client.FetchBlockIntervals(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("block: interval: %s: %v", node.Chain(), err))
			continue
		}

		for _, interval := range intervals {
			err := client.ProcessBlocks(interval, node)
			if err != nil {
				j.logger.Error(fmt.Errorf("block: %v", err))
				break
			}
		}
	}

	// rounds
	for _, node := range j.nodes {
		intervals, err := client.FetchRoundIntervals(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("round: interval: %s: %v", node.Chain(), err))
			continue
		}

		for _, interval := range intervals {
			err := client.ProcessRounds(interval, node)
			if err != nil {
				j.logger.Error(fmt.Errorf("round: %v", err))
				break
			}
		}
	}

	// prices
	for _, node := range j.nodes {
		err := client.ProcessPrices(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("price: interval: %s: %v", node.Chain(), err))
		}
	}
}

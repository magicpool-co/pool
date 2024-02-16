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
	lock, err := retrieveLock("cron:chart", time.Minute*30, j.locker)
	if lock == nil {
		if err != nil {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(context.Background())

	client := chart.New(j.pooldb, j.tsdb, j.redis)

	// shares
	for _, node := range j.nodes {
		for _, chain := range []string{node.Chain(), "S" + node.Chain()} {
			intervals, err := client.FetchShareIntervals(chain)
			if err != nil {
				j.logger.Error(fmt.Errorf("share: interval: %s: %v", chain, err))
				continue
			}

			for _, interval := range intervals {
				err := client.ProcessShares(chain, interval, node)
				if err != nil {
					j.logger.Error(fmt.Errorf("share: %v", err))
					break
				}
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

	// earnings
	for _, node := range j.nodes {
		intervals, err := client.FetchEarningIntervals(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("earning: interval: %s: %v", node.Chain(), err))
			continue
		}

		for _, interval := range intervals {
			err := client.ProcessEarnings(interval, node)
			if err != nil {
				j.logger.Error(fmt.Errorf("earning: %v", err))
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

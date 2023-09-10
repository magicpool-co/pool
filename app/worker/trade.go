package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/core/trade"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type TradeJob struct {
	locker    *redislock.Client
	logger    *log.Logger
	pooldb    *dbcl.Client
	redis     *redis.Client
	nodes     []types.PayoutNode
	exchanges []types.Exchange
	telegram  *telegram.Client
}

func (j *TradeJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:trade", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	client := trade.New(j.pooldb, j.redis, j.nodes, j.exchanges, j.telegram)
	// if err := client.CheckForNewBatches(); err != nil {
	// 	j.logger.Error(fmt.Errorf("check: %v", err))
	// }

	for _, exchangeID := range []types.ExchangeID{types.KucoinID, types.MEXCGlobalID} {
		batches, err := pooldb.GetActiveExchangeBatches(j.pooldb.Reader(), uint64(exchangeID))
		if err != nil {
			j.logger.Error(fmt.Errorf("fetch: %v", err))
		}

		for _, batch := range batches {
			if err := client.ProcessBatch(batch.ID); err != nil {
				j.logger.Error(fmt.Errorf("process: %d: %v", batch.ID, err))
			}
		}
	}
}

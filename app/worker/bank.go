package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/core/bank"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type BankJob struct {
	locker   *redislock.Client
	logger   *log.Logger
	pooldb   *dbcl.Client
	redis    *redis.Client
	nodes    []types.PayoutNode
	telegram *telegram.Client
}

func (j *BankJob) Run() {
	defer j.logger.RecoverPanic()
	lock, err := retrieveLock("cron:bank", time.Minute*5, j.locker)
	if lock == nil {
		if err != nil {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(context.Background())

	client := bank.New(j.pooldb, j.redis, j.telegram)
	for _, node := range j.nodes {
		err = client.BroadcastOutgoingTxs(node)
		if err != nil {
			j.logger.Error(fmt.Errorf("bank: broadcast: %s: %v", node.Chain(), err))
		}

		err = client.ConfirmOutgoingTxs(node)
		if err != nil {
			j.logger.Error(fmt.Errorf("bank: confirm: %s: %v", node.Chain(), err))
		}
	}
}

package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/core/payout"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type PayoutJob struct {
	locker   *redislock.Client
	logger   *log.Logger
	pooldb   *dbcl.Client
	nodes    []types.PayoutNode
	telegram *telegram.Client
}

func (j *PayoutJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:payout", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	client := payout.New(j.pooldb, j.telegram)

	for _, node := range j.nodes {
		if err := client.InitiatePayouts(node); err != nil {
			j.logger.Error(fmt.Errorf("payout: initiate: %s: %v", node.Chain(), err))
		} else if err := client.FinalizePayouts(node); err != nil {
			j.logger.Error(fmt.Errorf("payout: finalize: %s: %v", node.Chain(), err))
		}
	}
}

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

func (j *BankJob) broadcastAndConfirm(client *bank.Client, node types.PayoutNode) error {
	bankLock, err := client.FetchLock(node.Chain())
	if err != nil {
		return err
	}
	defer bankLock.Release(context.Background())

	err = client.BroadcastOutgoingTxs(node)
	if err != nil {
		return fmt.Errorf("bank: broadcast: %s: %v", node.Chain(), err)
	}

	err = client.ConfirmOutgoingTxs(node)
	if err != nil {
		return fmt.Errorf("bank: confirm: %s: %v", node.Chain(), err)
	}

	return nil
}

func (j *BankJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:bank", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	client := bank.New(j.pooldb, j.redis, j.telegram)
	for _, node := range j.nodes {
		err := j.broadcastAndConfirm(client, node)
		if err != nil {
			j.logger.Error(err)
		}
	}
}

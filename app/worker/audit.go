package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/core/audit"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type AuditJob struct {
	locker *redislock.Client
	logger *log.Logger
	pooldb *dbcl.Client
	nodes  []types.PayoutNode
}

func (j *AuditJob) Run() {
	defer j.logger.RecoverPanic()
	lock, err := retrieveLock("cron:audit", time.Minute*5, j.locker)
	if lock == nil {
		if err != nil {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(context.Background())

	for _, node := range j.nodes {
		if err := audit.CheckWallet(j.pooldb, node); err != nil {
			j.logger.Error(fmt.Errorf("audit: %s: %v", node.Chain(), err))
			continue
		}
	}
}

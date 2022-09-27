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

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:audit", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	for _, node := range j.nodes {
		if err := audit.CheckWallet(j.pooldb, node); err != nil {
			j.logger.Error(fmt.Errorf("audit: %s: %v", node.Chain(), err))
			continue
		}
	}
}

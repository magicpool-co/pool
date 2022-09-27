package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/core/credit"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type BlockUnlockJob struct {
	locker *redislock.Client
	logger *log.Logger
	pooldb *dbcl.Client
	nodes  []types.MiningNode
}

func (j *BlockUnlockJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:blkunlock", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
	}
	defer lock.Release(ctx)

	for _, node := range j.nodes {
		if err := credit.UnlockRounds(node, j.pooldb); err != nil {
			j.logger.Error(fmt.Errorf("unlock: %s: %v", node.Chain(), err))
			continue
		}
	}
}

type BlockCreditJob struct {
	locker *redislock.Client
	logger *log.Logger
	pooldb *dbcl.Client
	nodes  []types.MiningNode
}

func (j *BlockCreditJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:blkunlock", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	for _, node := range j.nodes {
		rounds, err := pooldb.GetMatureUnspentRounds(j.pooldb.Reader(), node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("credit: fetch rounds: %s: %v", node.Chain(), err))
			continue
		}

		for _, round := range rounds {
			shares, err := pooldb.GetSharesByRound(j.pooldb.Reader(), round.ID)
			if err != nil {
				j.logger.Error(fmt.Errorf("credit: fetch shares: %s: %v", node.Chain(), err))
				break
			} else if err := credit.CreditRound(j.pooldb, round, shares); err != nil {
				j.logger.Error(fmt.Errorf("credit: %s: %v", node.Chain(), err))
				break
			}
		}
	}
}

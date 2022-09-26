package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"

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

func (j *BlockUnlockJob) unlock(node types.MiningNode) error {
	height, syncing, err := node.GetStatus()
	if err != nil {
		return err
	} else if syncing {
		return nil
	}

	pendingRounds, err := pooldb.GetPendingRoundsByChain(j.pooldb.Reader(), node.Chain(), height-node.GetImmatureDepth())
	if err != nil {
		return err
	}

	for _, round := range pendingRounds {
		err := node.UnlockRound(round)
		if err != nil {
			return err
		}

		cols := []string{"uncle", "orphan", "pending", "mature", "spent", "height",
			"epoch_height", "uncle_height", "hash", "value", "created_at"}
		err = pooldb.UpdateRound(j.pooldb.Writer(), round, cols)
		if err != nil {
			return err
		}
	}

	immatureRounds, err := pooldb.GetImmatureRoundsByChain(j.pooldb.Reader(), node.Chain(), height-node.GetMatureDepth())
	if err != nil {
		return err
	}

	for _, round := range immatureRounds {
		round.Mature = true
		err = pooldb.UpdateRound(j.pooldb.Writer(), round, []string{"mature"})
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *BlockUnlockJob) Run() {
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
		if err := j.unlock(node); err != nil {
			j.logger.Error(fmt.Errorf("unlock: %s: %v", node.Chain(), err))
			continue
		}
	}
}

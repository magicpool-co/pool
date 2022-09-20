package worker

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/internal/accounting"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type BlockCreditJob struct {
	locker *redislock.Client
	logger *log.Logger
	pooldb *dbcl.Client
	nodes  []types.MiningNode
}

func (j *BlockCreditJob) credit(round *pooldb.Round, shares []*pooldb.Share) error {
	// create a miner index of proportional share values
	minerIdx := make(map[uint64]uint64)
	for _, share := range shares {
		minerIdx[share.MinerID] += share.Count
	}

	// fetch the recipients and create a recipient index of proportional fee values
	recipients, err := pooldb.GetRecipients(j.pooldb.Reader())
	if err != nil {
		return err
	}

	recipientIdx := make(map[uint64]uint64)
	for _, recipient := range recipients {
		if recipient.RecipientFeePercent == nil {
			return fmt.Errorf("no recipient fee set for %d", recipient.ID)
		}
		recipientIdx[recipient.ID] += types.Uint64Value(recipient.RecipientFeePercent)
	}

	// distribute the proceeds to miners and recipients
	minerValues, recipientValues, err := accounting.CreditRound(round.Value.BigInt, minerIdx, recipientIdx)
	if err != nil {
		return err
	}

	// merge the two values maps since miners and recipients are globally
	// unique (since recipients are stored as miners)
	compoundValues := make(map[uint64]*big.Int)
	for minerID, value := range minerValues {
		compoundValues[minerID] = value
	}

	for recipientID, value := range recipientValues {
		if _, ok := compoundValues[recipientID]; ok {
			compoundValues[recipientID].Add(compoundValues[recipientID], value)
		} else {
			compoundValues[recipientID] = value
		}
	}

	// fetch miners to check their payout chain
	minerIDs := make([]uint64, 0)
	for minerID := range minerValues {
		minerIDs = append(minerIDs, minerID)
	}

	miners, err := pooldb.GetMiners(j.pooldb.Reader(), minerIDs)
	if err != nil {
		return err
	}

	// create compound index for miners and recipients
	compoundIdx := make(map[uint64]*pooldb.Miner)
	for _, miner := range append(miners, recipients...) {
		compoundIdx[miner.ID] = miner
	}

	// process the balance inputs for all round distributions
	usedValue := new(big.Int)
	inputs := make([]*pooldb.BalanceInput, 0)
	for minerID, value := range compoundValues {
		miner, ok := compoundIdx[minerID]
		if !ok {
			return fmt.Errorf("no miner found for %d", minerID)
		} else if value == nil || value.Cmp(common.Big0) == 0 {
			continue
		}
		usedValue.Add(usedValue, value)

		input := &pooldb.BalanceInput{
			RoundID: round.ID,
			ChainID: round.ChainID,
			MinerID: miner.ID,

			Value:   dbcl.NullBigInt{Valid: true, BigInt: value},
			Pending: miner.ChainID != round.ChainID,
		}
		inputs = append(inputs, input)
		delete(compoundIdx, miner.ID)
		delete(compoundValues, miner.ID)
	}

	if len(compoundIdx) > 0 {
		return fmt.Errorf("unable to find %d miners in idx", len(compoundIdx))
	} else if len(compoundValues) > 0 {
		return fmt.Errorf("unable to find %d miners in values", len(compoundValues))
	} else if usedValue.Cmp(round.Value.BigInt) != 0 {
		return fmt.Errorf("crediting mismatch: have %s, want %s", usedValue, round.Value.BigInt)
	}

	tx, err := j.pooldb.Begin()
	if err != nil {
		return err
	}
	defer tx.SafeRollback()

	round.Spent = true
	if err := pooldb.InsertBalanceInputs(tx, inputs...); err != nil {
		return err
	} else if pooldb.UpdateRound(tx, round, []string{"spent"}); err != nil {
		return err
	}

	return tx.SafeCommit()
}

func (j *BlockCreditJob) Run() {
	defer recoverPanic(j.logger)

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
			} else if err := j.credit(round, shares); err != nil {
				j.logger.Error(fmt.Errorf("credit: %s: %v", node.Chain(), err))
				break
			}
		}
	}
}

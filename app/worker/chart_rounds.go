package worker

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

var (
	roundPeriod        = types.Period1h
	roundRollupPeriods = []types.PeriodType{types.Period1d}
)

type ChartRoundJob struct {
	locker *redislock.Client
	logger *log.Logger
	redis  *redis.Client
	pooldb *dbcl.Client
	tsdb   *dbcl.Client
	nodes  []types.MiningNode
}

func (j *ChartRoundJob) fetchIntervals(chain string) ([]time.Time, error) {
	lastTime, err := j.redis.GetChartRoundsLastTime(chain)
	if err != nil || lastTime.IsZero() {
		lastTime, err = tsdb.GetRoundMaxEndTime(j.tsdb.Reader(), chain, int(roundPeriod))
		if err != nil {
			return nil, err
		}
	}

	// handle initialization of intervals when there are no rounds in tsdb
	if lastTime.IsZero() {
		lastTime, err = pooldb.GetRoundMinTimestamp(j.pooldb.Reader(), chain)
		if err != nil {
			return nil, err
		} else if lastTime.IsZero() {
			return nil, nil
		}

		lastTime = common.NormalizeDate(lastTime, roundPeriod.Rollup(), false)
	}

	intervals := make([]time.Time, 0)
	for {
		startTime := lastTime
		endTime := lastTime.Add(roundPeriod.Rollup())
		if time.Now().Before(endTime) {
			break
		}

		// make sure no pending round are in the period to avoid
		// collecting stats on a round that has incomplete statistics
		pendingRoundCount, err := pooldb.GetPendingRoundCountBetweenTime(j.pooldb.Reader(), chain, startTime, endTime)
		if err != nil {
			return nil, err
		} else if pendingRoundCount > 0 {
			break
		}

		intervals = append(intervals, endTime)
		lastTime = endTime
	}

	return intervals, nil
}

func (j *ChartRoundJob) rollup(node types.MiningNode, endTime time.Time) error {
	lastTime, err := tsdb.GetRoundMaxEndTime(j.tsdb.Reader(), node.Chain(), int(roundPeriod))
	if err != nil {
		return err
	} else if !lastTime.Before(endTime) {
		return nil
	}

	startTime := endTime.Add(roundPeriod.Rollup() * -1)
	unitsBigFloat := new(big.Float).SetInt(node.GetUnits().Big())
	// startTime needs to be exclusive to avoid duplicating the statistics
	// for rounds with a timestamp of both startTime and endTime
	poolRounds, err := pooldb.GetRoundsBetweenTime(j.pooldb.Reader(), node.Chain(), startTime, endTime)
	if err != nil {
		return err
	}

	// need to get the prior round for calculating round time
	previousPoolRound, err := pooldb.GetLastRoundBeforeTime(j.pooldb.Reader(), node.Chain(), startTime)
	if err != nil {
		return err
	} else if previousPoolRound != nil {
		poolRounds = append([]*pooldb.Round{previousPoolRound}, poolRounds...)
	}

	var count, uncleCount uint64
	var value, roundTime, difficulty, acceptedShares, rejectedShares, invalidShares float64
	var lastRoundTime time.Time
	for _, poolRound := range poolRounds {
		if lastRoundTime.IsZero() {
			lastRoundTime = poolRound.CreatedAt
			continue
		}

		var roundValue float64
		if poolRound.Value.Valid {
			roundValueBig := new(big.Float).SetInt(poolRound.Value.BigInt)
			roundValue, _ = new(big.Float).Quo(roundValueBig, unitsBigFloat).Float64()
		}

		count += 1
		if poolRound.Uncle {
			uncleCount += 1
		}
		value += roundValue
		roundTime += float64(poolRound.CreatedAt.Sub(lastRoundTime)/time.Millisecond) / 1_000
		difficulty += float64(poolRound.Difficulty)
		acceptedShares += float64(poolRound.AcceptedShares)
		rejectedShares += float64(poolRound.RejectedShares)
		invalidShares += float64(poolRound.InvalidShares)
		lastRoundTime = poolRound.CreatedAt
	}

	var uncleRate float64
	if count > 0 {
		value /= float64(count)
		roundTime /= float64(count)
		difficulty /= float64(count)
		acceptedShares /= float64(count)
		rejectedShares /= float64(count)
		invalidShares /= float64(count)
		uncleRate = float64(uncleCount) / float64(count)
	}

	var hashrate, luck, profitability float64
	if acceptedShares > 0 {
		minedDifficulty := float64(node.GetShareDifficulty().Value()) * acceptedShares
		marketRate := getMarketRate(node.Chain(), endTime)

		hashrate = node.CalculateHashrate(roundTime, difficulty)
		luck = 100 * (difficulty / minedDifficulty)
		profitability = marketRate * (value / roundTime) / hashrate
	}

	avgLuck, err := tsdb.GetRoundsAverageLuckSlow(j.tsdb.Reader(), endTime, node.Chain(), int(roundPeriod), roundPeriod.Average())
	if err != nil {
		return err
	}

	avgProfitability, err := tsdb.GetRoundsAverageProfitabilitySlow(j.tsdb.Reader(), endTime, node.Chain(), int(roundPeriod), roundPeriod.Average())
	if err != nil {
		return err
	}

	round := &tsdb.Round{
		ChainID: node.Chain(),

		Value:            value,
		Difficulty:       difficulty,
		RoundTime:        roundTime,
		AcceptedShares:   acceptedShares,
		RejectedShares:   rejectedShares,
		InvalidShares:    invalidShares,
		Hashrate:         hashrate,
		UncleRate:        uncleRate,
		Luck:             luck,
		AvgLuck:          avgLuck,
		Profitability:    profitability,
		AvgProfitability: avgProfitability,

		Pending:   false,
		Period:    int(roundPeriod),
		StartTime: startTime,
		EndTime:   endTime,
	}

	tx, err := j.tsdb.Begin()
	if err != nil {
		return err
	}
	defer tx.SafeRollback()

	err = tsdb.InsertRounds(tx, round)
	if err != nil {
		return err
	}

	round.Pending = true
	round.Value *= float64(round.Count)
	round.Difficulty *= float64(round.Count)
	round.RoundTime *= float64(round.Count)
	round.AcceptedShares *= float64(round.Count)
	round.RejectedShares *= float64(round.Count)
	round.InvalidShares *= float64(round.Count)
	round.Hashrate, round.UncleRate = 0, 0
	round.Luck, round.AvgLuck = 0, 0
	round.Profitability, round.AvgProfitability = 0, 0

	for _, rollupPeriod := range roundRollupPeriods {
		round.Period = int(rollupPeriod)
		round.StartTime = common.NormalizeDate(round.StartTime, rollupPeriod.Rollup(), true)
		round.EndTime = common.NormalizeDate(round.StartTime, rollupPeriod.Rollup(), false)

		err = tsdb.InsertPartialRounds(tx, round)
		if err != nil {
			return err
		}
	}

	return tx.SafeCommit()
}

func (j *ChartRoundJob) finalize(node types.MiningNode, endTime time.Time) error {
	for _, rollupPeriod := range roundRollupPeriods {
		// finalize summed statistics
		rounds, err := tsdb.GetPendingRoundsAtEndTime(j.tsdb.Reader(), endTime, node.Chain(), int(rollupPeriod))
		if err != nil {
			return err
		}

		for _, round := range rounds {
			round.Pending = false
			round.AvgLuck = 0
			round.AvgProfitability = 0
			if round.Count > 0 {
				round.Value /= float64(round.Count)
				round.Difficulty /= float64(round.Count)
				round.RoundTime /= float64(round.Count)
				round.AcceptedShares /= float64(round.Count)
				round.RejectedShares /= float64(round.Count)
				round.InvalidShares /= float64(round.Count)
				round.UncleRate = float64(round.UncleCount) / float64(round.Count)
			}
			if round.AcceptedShares > 0 {
				minedDifficulty := float64(node.GetShareDifficulty().Value()) * round.AcceptedShares
				marketRate := getMarketRate(node.Chain(), endTime)
				round.Hashrate = node.CalculateHashrate(round.RoundTime, round.Difficulty)
				round.Luck = 100 * (round.Difficulty / minedDifficulty)
				round.Profitability = marketRate * (round.Value / round.RoundTime) / round.Hashrate
			}
		}

		if err := tsdb.InsertFinalRounds(j.tsdb.Writer(), rounds...); err != nil {
			return err
		}

		// finalize averages after updated statistics
		for _, round := range rounds {
			round.AvgLuck, err = tsdb.GetRoundsAverageLuckSlow(j.tsdb.Reader(), endTime, node.Chain(),
				int(rollupPeriod), rollupPeriod.Average())
			if err != nil {
				return err
			}
			round.AvgProfitability, err = tsdb.GetRoundsAverageProfitabilitySlow(j.tsdb.Reader(),
				endTime, node.Chain(), int(rollupPeriod), rollupPeriod.Average())
			if err != nil {
				return err
			}
		}

		if err := tsdb.InsertFinalRounds(j.tsdb.Writer(), rounds...); err != nil {
			return err
		}
	}

	return nil
}

func (j *ChartRoundJob) truncate(node types.MiningNode, endTime time.Time) error {
	for _, rollupPeriod := range roundRollupPeriods {
		timestamp := endTime.Add(rollupPeriod.Retention() * -1)
		err := tsdb.DeleteRoundsBeforeEndTime(j.tsdb.Writer(), timestamp, node.Chain(), int(rollupPeriod))
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *ChartRoundJob) Run() {
	defer recoverPanic(j.logger)

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:chrtrnd", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	for _, node := range j.nodes {
		intervals, err := j.fetchIntervals(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("round: interval: %s: %v", node.Chain(), err))
			continue
		}

		for _, interval := range intervals {
			if err := j.rollup(node, interval); err != nil {
				j.logger.Error(fmt.Errorf("round: rollup: %s: %v", node.Chain(), err))
				break
			}

			if err := j.finalize(node, interval); err != nil {
				j.logger.Error(fmt.Errorf("round: finalize: %s: %v", node.Chain(), err))
				break
			}

			if err := j.truncate(node, interval); err != nil {
				j.logger.Error(fmt.Errorf("round: truncate: %s: %v", node.Chain(), err))
				break
			}

			if err := j.redis.SetChartRoundsLastTime(node.Chain(), interval); err != nil {
				j.logger.Error(fmt.Errorf("round: delete: %s: %v", node.Chain(), err))
				break
			}
		}
	}
}

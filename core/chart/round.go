package chart

import (
	"fmt"
	"math/big"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

var (
	roundPeriod        = types.Period1h
	roundRollupPeriods = []types.PeriodType{types.Period1d}
)

func (c *Client) FetchRoundIntervals(chain string) ([]time.Time, error) {
	lastTime, err := c.redis.GetChartRoundsLastTime(chain)
	if err != nil || lastTime.IsZero() {
		lastTime, err = tsdb.GetRoundMaxEndTime(c.tsdb.Reader(), chain, int(roundPeriod))
		if err != nil {
			return nil, err
		}
	}

	// handle initialization of intervals when there are no rounds in tsdb
	if lastTime.IsZero() {
		lastTime, err = pooldb.GetRoundMinTimestamp(c.pooldb.Reader(), chain)
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
		pendingRoundCount, err := pooldb.GetPendingRoundCountBetweenTime(c.pooldb.Reader(), chain, startTime, endTime)
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

func (c *Client) rollupRounds(node types.MiningNode, endTime time.Time) error {
	lastTime, err := tsdb.GetRoundMaxEndTime(c.tsdb.Reader(), node.Chain(), int(roundPeriod))
	if err != nil {
		return err
	} else if !lastTime.Before(endTime) {
		return nil
	}

	startTime := endTime.Add(roundPeriod.Rollup() * -1)
	unitsBigFloat := new(big.Float).SetInt(node.GetUnits().Big())
	// startTime needs to be exclusive to avoid duplicating the statistics
	// for rounds with a timestamp of both startTime and endTime
	poolRounds, err := pooldb.GetRoundsBetweenTime(c.pooldb.Reader(), node.Chain(), startTime, endTime)
	if err != nil {
		return err
	}

	// need to get the prior round for calculating round time
	previousPoolRound, err := pooldb.GetLastRoundBeforeTime(c.pooldb.Reader(), node.Chain(), startTime)
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

	avgLuck, err := tsdb.GetRoundsAverageLuckSlow(c.tsdb.Reader(), endTime, node.Chain(), int(roundPeriod), roundPeriod.Average())
	if err != nil {
		return err
	}

	avgProfitability, err := tsdb.GetRoundsAverageProfitabilitySlow(c.tsdb.Reader(), endTime, node.Chain(), int(roundPeriod), roundPeriod.Average())
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

	tx, err := c.tsdb.Begin()
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

func (c *Client) finalizeRounds(node types.MiningNode, endTime time.Time) error {
	for _, rollupPeriod := range roundRollupPeriods {
		// finalize summed statistics
		rounds, err := tsdb.GetPendingRoundsAtEndTime(c.tsdb.Reader(), endTime, node.Chain(), int(rollupPeriod))
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

		if err := tsdb.InsertFinalRounds(c.tsdb.Writer(), rounds...); err != nil {
			return err
		}

		// finalize averages after updated statistics
		for _, round := range rounds {
			round.AvgLuck, err = tsdb.GetRoundsAverageLuckSlow(c.tsdb.Reader(), endTime, node.Chain(),
				int(rollupPeriod), rollupPeriod.Average())
			if err != nil {
				return err
			}
			round.AvgProfitability, err = tsdb.GetRoundsAverageProfitabilitySlow(c.tsdb.Reader(),
				endTime, node.Chain(), int(rollupPeriod), rollupPeriod.Average())
			if err != nil {
				return err
			}
		}

		if err := tsdb.InsertFinalRounds(c.tsdb.Writer(), rounds...); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) truncateRounds(node types.MiningNode, endTime time.Time) error {
	for _, rollupPeriod := range roundRollupPeriods {
		timestamp := endTime.Add(rollupPeriod.Retention() * -1)
		err := tsdb.DeleteRoundsBeforeEndTime(c.tsdb.Writer(), timestamp, node.Chain(), int(rollupPeriod))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) ProcessRounds(timestamp time.Time, node types.MiningNode) error {
	err := c.rollupRounds(node, timestamp)
	if err != nil {
		return fmt.Errorf("rollup: %s: %v", node.Chain(), err)
	}

	err = c.finalizeRounds(node, timestamp)
	if err != nil {
		return fmt.Errorf("finalize: %s: %v", node.Chain(), err)
	}

	err = c.truncateRounds(node, timestamp)
	if err != nil {
		return fmt.Errorf("truncate: %s: %v", node.Chain(), err)
	}

	err = c.redis.SetChartRoundsLastTime(node.Chain(), timestamp)
	if err != nil {
		return fmt.Errorf("time: %s: %v", node.Chain(), err)
	}

	return nil
}

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
	earningDelay  = time.Hour * 4
	earningPeriod = types.Period1d
)

func (c *Client) FetchEarningIntervals(chain string) ([]time.Time, error) {
	lastTime, err := c.redis.GetChartEarningsLastTime(chain)
	if err != nil || lastTime.IsZero() {
		lastTime, err = tsdb.GetEarningMaxEndTime(c.tsdb.Reader(), chain, int(earningPeriod))
		if err != nil {
			return nil, err
		}
	}

	// handle initialization of intervals when there are no rounds in tsdb
	if lastTime.IsZero() {
		lastTime, err = pooldb.GetBalanceInputMinTimestamp(c.pooldb.Reader(), chain)
		if err != nil {
			return nil, err
		} else if lastTime.IsZero() {
			return nil, nil
		}

		lastTime = common.NormalizeDate(lastTime, earningPeriod.Rollup(), false)
	}

	intervals := make([]time.Time, 0)
	for {
		endTime := lastTime.Add(earningPeriod.Rollup())
		if time.Since(endTime) < earningDelay {
			break
		}

		intervals = append(intervals, endTime)
		lastTime = endTime
	}

	if len(intervals) > 10 {
		intervals = intervals[:10]
	}

	return intervals, nil
}

func (c *Client) rollupEarnings(node types.MiningNode, endTime time.Time) error {
	lastTime, err := tsdb.GetEarningMaxEndTime(c.tsdb.Reader(), node.Chain(), int(earningPeriod))
	if err != nil {
		return err
	} else if !lastTime.Before(endTime) {
		return nil
	}

	startTime := endTime.Add(earningPeriod.Rollup() * -1)
	balanceInputSums, err := pooldb.GetBalanceInputSumFromRange(c.pooldb.Reader(), node.Chain(), startTime, endTime)
	if err != nil {
		return err
	}

	globalAvg, err := tsdb.GetGlobalEarningsAverage(c.tsdb.Reader(), endTime,
		node.Chain(), int(earningPeriod), earningPeriod.Average())
	if err != nil {
		return err
	}

	minerAvg, err := tsdb.GetMinerEarningsAverage(c.tsdb.Reader(), endTime,
		node.Chain(), int(earningPeriod), earningPeriod.Average())
	if err != nil {
		return err
	}

	units := node.GetUnits().Big()
	globalSum := new(big.Int)
	minerEarnings := make([]*tsdb.Earning, len(balanceInputSums))
	for i, balanceInputSum := range balanceInputSums {
		globalSum.Add(globalSum, balanceInputSum.Value.BigInt)
		minerEarnings[i] = &tsdb.Earning{
			ChainID: node.Chain(),
			MinerID: types.Uint64Ptr(balanceInputSum.MinerID),

			Value:    common.BigIntToFloat64(balanceInputSum.Value.BigInt, units),
			AvgValue: minerAvg[balanceInputSum.MinerID],

			Pending:   false,
			Count:     1,
			Period:    int(earningPeriod),
			StartTime: startTime,
			EndTime:   endTime,
		}
	}

	globalEarning := &tsdb.Earning{
		ChainID: node.Chain(),

		Value:    common.BigIntToFloat64(globalSum, units),
		AvgValue: globalAvg,

		Pending:   false,
		Count:     1,
		Period:    int(earningPeriod),
		StartTime: startTime,
		EndTime:   endTime,
	}

	tx, err := c.tsdb.Begin()
	if err != nil {
		return err
	}
	defer tx.SafeRollback()

	if err := tsdb.InsertGlobalEarnings(tx, globalEarning); err != nil {
		return err
	} else if err := tsdb.InsertMinerEarnings(tx, minerEarnings...); err != nil {
		return err
	}

	err = tx.SafeCommit()
	if err != nil {
		return err
	}

	err = c.redis.SetChartEarningsLastTime(node.Chain(), endTime)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ProcessEarnings(timestamp time.Time, node types.MiningNode) error {
	err := c.rollupEarnings(node, timestamp)
	if err != nil {
		return fmt.Errorf("rollup: %s: %v", node.Chain(), err)
	}

	return nil
}

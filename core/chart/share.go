package chart

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

var (
	sharePeriod        = types.Period15m
	shareSeconds       = types.Period15m.Rollup() / time.Second
	shareRollupPeriods = []types.PeriodType{types.Period4h, types.Period1d}
)

func parseInterval(interval string) (time.Time, error) {
	parsedInterval, err := strconv.ParseInt(interval, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	timestamp := time.Unix(parsedInterval, 0)
	if timestamp.IsZero() {
		return time.Time{}, fmt.Errorf("interval is zero")
	}

	return timestamp, nil
}

func getInitialShareAverages(q dbcl.Querier, ts time.Time, chain string, period types.PeriodType) (float64, map[uint64]float64, map[uint64]float64, error) {
	globalAvg, err := tsdb.GetGlobalSharesAverage(q, ts, chain, int(period), period.Average())
	if err != nil {
		return 0, nil, nil, err
	}

	minerAvg, err := tsdb.GetMinerSharesAverage(q, ts, chain, int(period), period.Average())
	if err != nil {
		return 0, nil, nil, err
	}

	workerAvg, err := tsdb.GetWorkerSharesAverage(q, ts, chain, int(period), period.Average())
	if err != nil {
		return 0, nil, nil, err
	}

	return globalAvg, minerAvg, workerAvg, nil
}

func (c *Client) rollupShares(chain string, node types.MiningNode, interval string) error {
	endTime, err := parseInterval(interval)
	if err != nil {
		return err
	}

	// @TODO: we should probs check the db just in case if its not set
	lastTime, err := c.redis.GetChartSharesLastTime(chain)
	if err != nil {
		return err
	} else if !lastTime.Before(endTime) {
		return nil
	}

	accepted, acceptedAdjusted, err := c.redis.GetIntervalAcceptedShares(chain, interval)
	if err != nil {
		return err
	}
	rejected, rejectedAdjusted, err := c.redis.GetIntervalRejectedShares(chain, interval)
	if err != nil {
		return err
	}
	invalid, invalidAdjusted, err := c.redis.GetIntervalInvalidShares(chain, interval)
	if err != nil {
		return err
	}
	globalAvg, minerAvg, workerAvg, err := getInitialShareAverages(c.tsdb.Reader(), endTime, chain, sharePeriod)
	if err != nil {
		return err
	}

	globalShare := &tsdb.Share{
		ChainID:     node.Chain(),
		AvgHashrate: globalAvg,
		Count:       1,
		Period:      int(sharePeriod),
		StartTime:   endTime.Add(sharePeriod.Rollup() * -1),
		EndTime:     endTime,
	}

	// @TODO: add minerAvg:workerAvg to uniqueIDs by building the compound ID somehow

	uniqueIDs := make(map[string]bool)
	for compoundID := range accepted {
		uniqueIDs[compoundID] = true
	}
	for compoundID := range rejected {
		uniqueIDs[compoundID] = true
	}
	for compoundID := range invalid {
		uniqueIDs[compoundID] = true
	}
	delete(uniqueIDs, "")

	minerSharesIdx := make(map[uint64]*tsdb.Share)
	workerSharesIdx := make(map[uint64]*tsdb.Share)
	for compoundID := range uniqueIDs {
		parts := strings.Split(compoundID, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid compoundID %s", compoundID)
		}
		minerID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return err
		} else if _, ok := minerSharesIdx[minerID]; !ok {
			minerSharesIdx[minerID] = &tsdb.Share{
				ChainID:     chain,
				MinerID:     types.Uint64Ptr(minerID),
				Miners:      1,
				AvgHashrate: minerAvg[minerID],
				Count:       1,
				Period:      int(sharePeriod),
				StartTime:   endTime.Add(sharePeriod.Rollup() * -1),
				EndTime:     endTime,
			}
		}
		minerSharesIdx[minerID].AcceptedShares += accepted[compoundID]
		minerSharesIdx[minerID].AcceptedAdjustedShares += acceptedAdjusted[compoundID]
		if _, ok := acceptedAdjusted[compoundID]; !ok {
			minerSharesIdx[minerID].AcceptedAdjustedShares += accepted[compoundID]
		}

		minerSharesIdx[minerID].RejectedShares += rejected[compoundID]
		minerSharesIdx[minerID].RejectedAdjustedShares += rejectedAdjusted[compoundID]
		if _, ok := rejectedAdjusted[compoundID]; !ok {
			minerSharesIdx[minerID].RejectedAdjustedShares += rejected[compoundID]
		}

		minerSharesIdx[minerID].InvalidShares += invalid[compoundID]
		minerSharesIdx[minerID].InvalidAdjustedShares += invalidAdjusted[compoundID]
		if _, ok := invalidAdjusted[compoundID]; !ok {
			minerSharesIdx[minerID].InvalidAdjustedShares += invalid[compoundID]
		}

		workerID, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return err
		} else if workerID == 0 {
			continue
		} else if _, ok := workerSharesIdx[workerID]; !ok {
			// add miner shares worker count
			minerSharesIdx[minerID].Workers++
			workerSharesIdx[workerID] = &tsdb.Share{
				ChainID:     chain,
				WorkerID:    types.Uint64Ptr(workerID),
				Workers:     1,
				AvgHashrate: workerAvg[workerID],
				Count:       1,
				Period:      int(sharePeriod),
				StartTime:   endTime.Add(sharePeriod.Rollup() * -1),
				EndTime:     endTime,
			}
		}
		workerSharesIdx[workerID].AcceptedShares += accepted[compoundID]
		workerSharesIdx[workerID].AcceptedAdjustedShares += acceptedAdjusted[compoundID]
		if _, ok := acceptedAdjusted[compoundID]; !ok {
			workerSharesIdx[workerID].AcceptedAdjustedShares += accepted[compoundID]
		}

		workerSharesIdx[workerID].RejectedShares += rejected[compoundID]
		workerSharesIdx[workerID].RejectedAdjustedShares += rejectedAdjusted[compoundID]
		if _, ok := rejectedAdjusted[compoundID]; !ok {
			workerSharesIdx[workerID].RejectedAdjustedShares += rejected[compoundID]
		}

		workerSharesIdx[workerID].InvalidShares += invalid[compoundID]
		workerSharesIdx[workerID].InvalidAdjustedShares += invalidAdjusted[compoundID]
		if _, ok := invalidAdjusted[compoundID]; !ok {
			workerSharesIdx[workerID].InvalidAdjustedShares += invalid[compoundID]
		}
	}

	// make sure not to insert when there is nothing to insert
	if len(minerSharesIdx) != 0 {
		minerShares := make([]*tsdb.Share, 0)
		for _, minerShare := range minerSharesIdx {
			minerShare.Hashrate = float64(minerShare.AcceptedShares) * node.GetAdjustedShareDifficulty() / float64(shareSeconds)
			minerShares = append(minerShares, minerShare)

			// sum all miner values for global share
			globalShare.AcceptedShares += minerShare.AcceptedShares
			globalShare.AcceptedAdjustedShares += minerShare.AcceptedAdjustedShares
			globalShare.RejectedShares += minerShare.RejectedShares
			globalShare.RejectedAdjustedShares += minerShare.RejectedAdjustedShares
			globalShare.InvalidShares += minerShare.InvalidShares
			globalShare.InvalidAdjustedShares += minerShare.InvalidAdjustedShares
			globalShare.Hashrate += minerShare.Hashrate
		}

		workerShares := make([]*tsdb.Share, 0)
		for _, workerShare := range workerSharesIdx {
			workerShare.Hashrate = float64(workerShare.AcceptedShares) * node.GetAdjustedShareDifficulty() / float64(shareSeconds)
			workerShares = append(workerShares, workerShare)
		}

		// add miner and worker count to global share
		globalShare.Miners = uint64(len(minerShares))
		globalShare.Workers = uint64(len(workerShares))

		tx, err := c.tsdb.Begin()
		if err != nil {
			return err
		}
		defer tx.SafeRollback()

		if err := tsdb.InsertGlobalShares(tx, globalShare); err != nil {
			return err
		} else if err := tsdb.InsertMinerShares(tx, minerShares...); err != nil {
			return err
		} else if err := tsdb.InsertWorkerShares(tx, workerShares...); err != nil {
			return err
		}

		fullShareList := [][]*tsdb.Share{[]*tsdb.Share{globalShare}, minerShares, workerShares}
		for _, shareList := range fullShareList {
			for _, share := range shareList {
				share.Pending = true
				share.Hashrate *= float64(share.Count)
				share.AvgHashrate = 0
			}
		}

		for _, rollupPeriod := range shareRollupPeriods {
			for _, shareList := range fullShareList {
				for _, share := range shareList {
					share.Period = int(rollupPeriod)
					share.StartTime = common.NormalizeDate(share.StartTime, rollupPeriod.Rollup(), true)
					share.EndTime = common.NormalizeDate(share.StartTime, rollupPeriod.Rollup(), false)
				}
			}

			if err := tsdb.InsertPartialGlobalShares(tx, globalShare); err != nil {
				return err
			} else if err := tsdb.InsertPartialMinerShares(tx, minerShares...); err != nil {
				return err
			} else if err := tsdb.InsertPartialWorkerShares(tx, workerShares...); err != nil {
				return err
			}
		}

		if err := tx.SafeCommit(); err != nil {
			return err
		}

		// set top 100 minerIDs by hashrate to avoid heavy DB queries
		sort.Slice(minerShares, func(i, j int) bool {
			return minerShares[i].Hashrate < minerShares[j].Hashrate
		})

		topMinerCount := len(minerShares)
		if topMinerCount > 100 {
			topMinerCount = 100
		}

		topMinerIDs := make([]uint64, topMinerCount)
		for i := len(minerShares) - 1; i > len(minerShares)-topMinerCount-1; i-- {
			topMinerIDs[i] = types.Uint64Value(minerShares[i].MinerID)
		}

		if err := c.redis.SetTopMinerIDs(chain, topMinerIDs); err != nil {
			return err
		}
	}

	if err := c.redis.SetChartSharesLastTime(chain, endTime); err != nil {
		return err
	}

	return nil
}

func finalizeShare(share *tsdb.Share) {
	share.Pending = false
	share.AvgHashrate = 0
	if share.Count > 0 {
		share.Hashrate /= float64(share.Count)
	}
}

func (c *Client) finalizeShares(chain string, endTime time.Time) error {
	for _, rollupPeriod := range shareRollupPeriods {
		// finalize summed statistics
		globalShares, err := tsdb.GetPendingGlobalSharesByEndTime(c.tsdb.Reader(), endTime, chain, int(rollupPeriod))
		if err != nil {
			return err
		}
		minerShares, err := tsdb.GetPendingMinerSharesByEndTime(c.tsdb.Reader(), endTime, chain, int(rollupPeriod))
		if err != nil {
			return err
		}
		workerShares, err := tsdb.GetPendingWorkerSharesByEndTime(c.tsdb.Reader(), endTime, chain, int(rollupPeriod))
		if err != nil {
			return err
		}

		for _, globalShare := range globalShares {
			finalizeShare(globalShare)
		}
		for _, minerShare := range minerShares {
			finalizeShare(minerShare)
		}
		for _, workerShare := range workerShares {
			finalizeShare(workerShare)
		}

		if err := tsdb.InsertFinalGlobalShares(c.tsdb.Writer(), globalShares...); err != nil {
			return err
		} else if err := tsdb.InsertFinalMinerShares(c.tsdb.Writer(), minerShares...); err != nil {
			return err
		} else if err := tsdb.InsertFinalWorkerShares(c.tsdb.Writer(), workerShares...); err != nil {
			return err
		}

		// finalize averages after updated statistics
		globalAvg, minerAvg, workerAvg, err := getInitialShareAverages(c.tsdb.Reader(), endTime, chain, rollupPeriod)
		if err != nil {
			return err
		}

		for _, globalShare := range globalShares {
			globalShare.AvgHashrate = globalAvg
		}
		for _, minerShare := range minerShares {
			minerShare.AvgHashrate = minerAvg[types.Uint64Value(minerShare.MinerID)]
		}
		for _, workerShare := range workerShares {
			workerShare.AvgHashrate = workerAvg[types.Uint64Value(workerShare.WorkerID)]

		}

		if err := tsdb.InsertFinalGlobalShares(c.tsdb.Writer(), globalShares...); err != nil {
			return err
		} else if err := tsdb.InsertFinalMinerShares(c.tsdb.Writer(), minerShares...); err != nil {
			return err
		} else if err := tsdb.InsertFinalWorkerShares(c.tsdb.Writer(), workerShares...); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) truncateShares(chain string, endTime time.Time) error {
	for _, rollupPeriod := range append([]types.PeriodType{sharePeriod}, shareRollupPeriods...) {
		timestamp := endTime.Add(rollupPeriod.Retention() * -1)
		if err := tsdb.DeleteGlobalSharesBeforeEndTime(c.tsdb.Writer(), timestamp, chain, int(rollupPeriod)); err != nil {
			return err
		} else if err := tsdb.DeleteMinerSharesBeforeEndTime(c.tsdb.Writer(), timestamp, chain, int(rollupPeriod)); err != nil {
			return err
		} else if err := tsdb.DeleteWorkerSharesBeforeEndTime(c.tsdb.Writer(), timestamp, chain, int(rollupPeriod)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) FetchShareIntervals(chain string) ([]string, error) {
	return c.redis.GetIntervals(chain)
}

func (c *Client) ProcessShares(chain, interval string, node types.MiningNode) error {
	timestamp, err := parseInterval(interval)
	if err != nil {
		return fmt.Errorf("parse: %s: %v", chain, err)
	} else if !timestamp.Before(time.Now()) {
		return nil
	}

	err = c.rollupShares(chain, node, interval)
	if err != nil {
		return fmt.Errorf("rollup: %s: %v", chain, err)
	}

	err = c.finalizeShares(chain, timestamp)
	if err != nil {
		return fmt.Errorf("finalize: %s: %v", chain, err)
	}

	err = c.truncateShares(chain, timestamp)
	if err != nil {
		return fmt.Errorf("truncate: %s: %v", chain, err)
	}

	err = c.redis.DeleteInterval(chain, interval)
	if err != nil {
		return fmt.Errorf("delete: %s: %v", chain, err)
	}

	return nil
}

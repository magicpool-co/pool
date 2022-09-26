package worker

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/redis"
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

type ChartShareJob struct {
	locker *redislock.Client
	logger *log.Logger
	redis  *redis.Client
	tsdb   *dbcl.Client
	nodes  []types.MiningNode
}

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

func getInitialAverages(q dbcl.Querier, ts time.Time, chain string, period types.PeriodType) (float64, map[uint64]float64, map[uint64]float64, error) {
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

func (j *ChartShareJob) rollup(node types.MiningNode, interval string) error {
	endTime, err := parseInterval(interval)
	if err != nil {
		return err
	}

	// @TODO: we should probs check the db just in case if its not set
	lastTime, err := j.redis.GetChartSharesLastTime(node.Chain())
	if err != nil {
		return err
	} else if !lastTime.Before(endTime) {
		return nil
	}

	accepted, err := j.redis.GetIntervalAcceptedShares(node.Chain(), interval)
	if err != nil {
		return err
	}
	rejected, err := j.redis.GetIntervalRejectedShares(node.Chain(), interval)
	if err != nil {
		return err
	}
	invalid, err := j.redis.GetIntervalInvalidShares(node.Chain(), interval)
	if err != nil {
		return err
	}
	reported, err := j.redis.GetIntervalReportedHashrates(node.Chain(), interval)
	if err != nil {
		return err
	}
	globalAvg, minerAvg, workerAvg, err := getInitialAverages(j.tsdb.Reader(), endTime, node.Chain(), sharePeriod)
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
	for compoundID := range reported {
		uniqueIDs[compoundID] = true
	}

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
				ChainID:     node.Chain(),
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
		minerSharesIdx[minerID].RejectedShares += rejected[compoundID]
		minerSharesIdx[minerID].InvalidShares += invalid[compoundID]
		minerSharesIdx[minerID].ReportedHashrate += reported[compoundID]

		workerID, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return err
		} else if workerID == 0 {
			continue
		} else if _, ok := workerSharesIdx[workerID]; !ok {
			// add miner shares worker count
			minerSharesIdx[minerID].Workers++
			workerSharesIdx[workerID] = &tsdb.Share{
				ChainID:     node.Chain(),
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
		workerSharesIdx[workerID].RejectedShares += rejected[compoundID]
		workerSharesIdx[workerID].InvalidShares += invalid[compoundID]
		workerSharesIdx[workerID].ReportedHashrate += reported[compoundID]
	}

	// make sure not to insert when there is nothing to insert
	if len(minerSharesIdx) != 0 {
		minerShares := make([]*tsdb.Share, 0)
		for _, minerShare := range minerSharesIdx {
			minerShare.Hashrate = float64(minerShare.AcceptedShares) * node.GetAdjustedShareDifficulty() / float64(shareSeconds)
			minerShares = append(minerShares, minerShare)

			// sum all miner values for global share
			globalShare.AcceptedShares += minerShare.AcceptedShares
			globalShare.RejectedShares += minerShare.RejectedShares
			globalShare.InvalidShares += minerShare.InvalidShares
			globalShare.Hashrate += minerShare.Hashrate
			globalShare.ReportedHashrate += minerShare.ReportedHashrate
		}

		workerShares := make([]*tsdb.Share, 0)
		for _, workerShare := range workerSharesIdx {
			workerShare.Hashrate = float64(workerShare.AcceptedShares) * node.GetAdjustedShareDifficulty() / float64(shareSeconds)
			workerShares = append(workerShares, workerShare)
		}

		// add miner and worker count to global share
		globalShare.Miners = uint64(len(minerShares))
		globalShare.Workers = uint64(len(workerShares))

		tx, err := j.tsdb.Begin()
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
				share.ReportedHashrate *= float64(share.Count)
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
	}

	if err := j.redis.SetChartSharesLastTime(node.Chain(), endTime); err != nil {
		return err
	}

	return nil
}

func finalizeShare(share *tsdb.Share) {
	share.Pending = false
	share.AvgHashrate = 0
	if share.Count > 0 {
		share.Hashrate /= float64(share.Count)
		share.ReportedHashrate /= float64(share.Count)
	}
}

func (j *ChartShareJob) finalize(node types.MiningNode, endTime time.Time) error {
	for _, rollupPeriod := range shareRollupPeriods {
		// finalize summed statistics
		globalShares, err := tsdb.GetPendingGlobalSharesByEndTime(j.tsdb.Reader(), endTime, node.Chain(), int(rollupPeriod))
		if err != nil {
			return err
		}
		minerShares, err := tsdb.GetPendingMinerSharesByEndTime(j.tsdb.Reader(), endTime, node.Chain(), int(rollupPeriod))
		if err != nil {
			return err
		}
		workerShares, err := tsdb.GetPendingWorkerSharesByEndTime(j.tsdb.Reader(), endTime, node.Chain(), int(rollupPeriod))
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

		if err := tsdb.InsertFinalGlobalShares(j.tsdb.Writer(), globalShares...); err != nil {
			return err
		} else if err := tsdb.InsertFinalMinerShares(j.tsdb.Writer(), minerShares...); err != nil {
			return err
		} else if err := tsdb.InsertFinalWorkerShares(j.tsdb.Writer(), workerShares...); err != nil {
			return err
		}

		// finalize averages after updated statistics
		globalAvg, minerAvg, workerAvg, err := getInitialAverages(j.tsdb.Reader(), endTime, node.Chain(), rollupPeriod)
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

		if err := tsdb.InsertFinalGlobalShares(j.tsdb.Writer(), globalShares...); err != nil {
			return err
		} else if err := tsdb.InsertFinalMinerShares(j.tsdb.Writer(), minerShares...); err != nil {
			return err
		} else if err := tsdb.InsertFinalWorkerShares(j.tsdb.Writer(), workerShares...); err != nil {
			return err
		}
	}

	return nil
}

func (j *ChartShareJob) truncate(node types.MiningNode, endTime time.Time) error {
	for _, rollupPeriod := range append([]types.PeriodType{sharePeriod}, shareRollupPeriods...) {
		timestamp := endTime.Add(rollupPeriod.Retention() * -1)
		if err := tsdb.DeleteGlobalSharesBeforeEndTime(j.tsdb.Writer(), timestamp, node.Chain(), int(rollupPeriod)); err != nil {
			return err
		} else if err := tsdb.DeleteMinerSharesBeforeEndTime(j.tsdb.Writer(), timestamp, node.Chain(), int(rollupPeriod)); err != nil {
			return err
		} else if err := tsdb.DeleteWorkerSharesBeforeEndTime(j.tsdb.Writer(), timestamp, node.Chain(), int(rollupPeriod)); err != nil {
			return err
		}
	}

	return nil
}

func (j *ChartShareJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:chrtshr", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	for _, node := range j.nodes {
		intervals, err := j.redis.GetIntervals(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("share: interval: %s: %v", node.Chain(), err))
			continue
		}

		for _, interval := range intervals {
			timestamp, err := parseInterval(interval)
			if err != nil {
				j.logger.Error(fmt.Errorf("share: parse: %s: %v", node.Chain(), err))
				break
			} else if !timestamp.Before(time.Now()) {
				continue
			}

			if err := j.rollup(node, interval); err != nil {
				j.logger.Error(fmt.Errorf("share: rollup: %s: %v", node.Chain(), err))
				break
			}

			if err := j.finalize(node, timestamp); err != nil {
				j.logger.Error(fmt.Errorf("share: finalize: %s: %v", node.Chain(), err))
				break
			}

			if err := j.truncate(node, timestamp); err != nil {
				j.logger.Error(fmt.Errorf("share: truncate: %s: %v", node.Chain(), err))
				break
			}

			if err := j.redis.DeleteInterval(node.Chain(), interval); err != nil {
				j.logger.Error(fmt.Errorf("share: delete: %s: %v", node.Chain(), err))
				break
			}
		}
	}

}

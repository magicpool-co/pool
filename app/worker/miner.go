package worker

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/core/mailer"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func parseCompoundID(compoundID string) (uint64, uint64, string, error) {
	parts := strings.Split(compoundID, ":")
	if len(parts) != 3 {
		return 0, 0, "", fmt.Errorf("invalid compoundID: %s", compoundID)
	}

	minerID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || minerID == 0 {
		return 0, 0, "", fmt.Errorf("invalid compoundID: %s", compoundID)
	}

	workerID, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, "", fmt.Errorf("invalid compoundID: %s", compoundID)
	}

	// need to replace "|" with ":" for IPv6 compatibility
	ipAddress := strings.ReplaceAll(parts[2], "|", ":")

	return minerID, workerID, ipAddress, nil
}

type MinerJob struct {
	locker *redislock.Client
	logger *log.Logger
	pooldb *dbcl.Client
	redis  *redis.Client
	nodes  []types.MiningNode
}

func (j *MinerJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:miner", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	for _, node := range j.nodes {
		ipAddressIdx, err := j.redis.GetMinerIPAddresses(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("ip: fetch ips: %s: %v", node.Chain(), err))
		}

		inactiveIpAddressIdx, err := j.redis.GetMinerIPAddressesInactive(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("ip: fetch inactive ips: %s: %v", node.Chain(), err))
		}

		rttIdx, err := j.redis.GetMinerLatencies(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("ip: fetch latencies: %s: %v", node.Chain(), err))
		}

		// process the index into a slice of addresses
		addresses := make([]*pooldb.IPAddress, 0)
		inactiveWorkers := make([]uint64, 0)
		addToInactiveIPs := make([]string, 0)
		removeFromActiveIPs := make([]string, 0)
		removeFromInactiveIPs := make([]string, 0)
		for compoundID, timestamp := range ipAddressIdx {
			minerID, workerID, ipAddress, err := parseCompoundID(compoundID)
			if err != nil {
				j.logger.Error(fmt.Errorf("ip: %v", err))
				continue
			}

			var rtt *float64
			if rawRtt, ok := rttIdx[compoundID]; ok {
				rtt = types.Float64Ptr(rawRtt / 1000)
			}

			lastShare := time.Unix(int64(timestamp), 0)
			timeSinceLastShare := time.Now().Sub(lastShare)
			active := timeSinceLastShare < time.Hour
			expired := timeSinceLastShare > time.Hour*24

			// check to see if the worker is already marked as inactive
			if _, ok := inactiveIpAddressIdx[compoundID]; ok {
				if !active && !expired {
					// if the worker is still inactive but not expired, do nothing
					continue
				} else if expired {
					// if the worker is newly expired, remove it from the active ips
					removeFromActiveIPs = append(removeFromActiveIPs, compoundID)
				}

				// the state must have changed (since its no longer inactive but not expired),
				// so we can remove it from the inactive IPs
				removeFromInactiveIPs = append(removeFromInactiveIPs, compoundID)
			} else if !active {
				// the worker is inactive, but not yet added to
				// the inactive worker list, so we add it
				addToInactiveIPs = append(addToInactiveIPs, compoundID)
				inactiveWorkers = append(inactiveWorkers, workerID)
			}

			addresses = append(addresses, &pooldb.IPAddress{
				MinerID:   minerID,
				WorkerID:  workerID,
				ChainID:   node.Chain(),
				IPAddress: ipAddress,

				Active:        active,
				Expired:       expired,
				LastShare:     lastShare,
				RoundTripTime: rtt,
			})
		}

		// insert the ip addresses, set old addresses to inactive or expired, clear the redis sorted set
		if err := pooldb.InsertIPAddresses(j.pooldb.Writer(), addresses...); err != nil {
			j.logger.Error(fmt.Errorf("ip: insert: %s: %v", node.Chain(), err))
		} else if err := pooldb.UpdateWorkerSetInactive(j.pooldb.Writer(), inactiveWorkers); err != nil {
			j.logger.Error(fmt.Errorf("ip: set inactive: %s: %v", node.Chain(), err))
		} else if err := j.redis.AddMinerIPAddressesInactive(node.Chain(), addToInactiveIPs); err != nil {
			j.logger.Error(fmt.Errorf("ip: add inactive: %s: %v", node.Chain(), err))
		} else if err := j.redis.RemoveMinerIPAddressesInactive(node.Chain(), removeFromInactiveIPs); err != nil {
			j.logger.Error(fmt.Errorf("ip: remove inactive: %s: %v", node.Chain(), err))
		} else if err := j.redis.RemoveMinerIPAddresses(node.Chain(), removeFromActiveIPs); err != nil {
			j.logger.Error(fmt.Errorf("ip: remove active: %s: %v", node.Chain(), err))
		}
	}
}

type MinerNotifyJob struct {
	locker   *redislock.Client
	logger   *log.Logger
	pooldb   *dbcl.Client
	redis    *redis.Client
	nodes    []types.MiningNode
	mailer   *mailer.Client
	telegram *telegram.Client
}

func (j *MinerNotifyJob) notifyMiner(miner *pooldb.Miner, workers []*pooldb.Worker) error {
	if miner.Email == nil {
		return fmt.Errorf("no email for miner")
	}

	address := miner.Address
	if parts := strings.Split(address, ":"); len(parts) == 0 {
		address = strings.ToLower(miner.ChainID) + ":" + address
	}

	dbTx, err := j.pooldb.Begin()
	if err != nil {
		return err
	}
	defer dbTx.SafeRollback()

	workerIdx := make(map[string]time.Time)
	for _, worker := range workers {
		workerIdx[worker.Name] = worker.LastShare

		worker.Notified = true
		err = pooldb.UpdateWorker(dbTx, worker, []string{"notified"})
		if err != nil {
			return err
		}
	}

	err = dbTx.SafeCommit()
	if err != nil {
		return err
	}

	return j.mailer.SendEmailForWorkers("tug@sencha.dev", address, workerIdx)
}

func (j *MinerNotifyJob) Run() {
	defer j.logger.RecoverPanic()

	ctx := context.Background()
	lock, err := j.locker.Obtain(ctx, "cron:minerntfy", time.Minute*5, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			j.logger.Error(err)
		}
		return
	}
	defer lock.Release(ctx)

	inactiveIpAddressIdx := make(map[string]bool)
	for _, node := range j.nodes {
		partialInactiveIdx, err := j.redis.GetMinerIPAddressesInactive(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("notify: fetch inactive ips: %s: %v", node.Chain(), err))
			return
		}

		for compoundID := range partialInactiveIdx {
			inactiveIpAddressIdx[compoundID] = true
		}
	}

	workerIdx := make(map[uint64]bool)
	for compoundID := range inactiveIpAddressIdx {
		_, workerID, _, err := parseCompoundID(compoundID)
		if err != nil {
			j.logger.Error(fmt.Errorf("notify: %v", err))
			continue
		}

		workerIdx[workerID] = true
	}

	workerIDs := make([]uint64, 0)
	for workerID := range workerIdx {
		workerIDs = append(workerIDs, workerID)
	}

	workers, err := pooldb.GetWorkersWithLastShares(j.pooldb.Reader(), workerIDs)
	if err != nil {
		j.logger.Error(fmt.Errorf("notify: get workers: %v", err))
		return
	}

	workersToNotify := make(map[uint64][]*pooldb.Worker, 0)
	for _, worker := range workers {
		if !worker.Active && !worker.Notified {
			if _, ok := workersToNotify[worker.MinerID]; !ok {
				workersToNotify[worker.MinerID] = make([]*pooldb.Worker, 0)
			}

			workersToNotify[worker.MinerID] = append(workersToNotify[worker.MinerID], worker)
		}
	}

	minerIDs := make([]uint64, 0)
	for minerID := range workersToNotify {
		minerIDs = append(minerIDs, minerID)
	}

	miners, err := pooldb.GetMiners(j.pooldb.Reader(), minerIDs)
	if err != nil {
		j.logger.Error(fmt.Errorf("notify: get miners: %v", err))
	}

	minerIdx := make(map[uint64]*pooldb.Miner, 0)
	for _, miner := range miners {
		// if miner.EnabledWorkerNotifications && miner.Email != nil {
		minerIdx[miner.ID] = miner
		// }
	}

	for minerID, workers := range workersToNotify {
		miner, ok := minerIdx[minerID]
		if !ok || miner == nil {
			continue
		}

		err := j.notifyMiner(miner, workers)
		if err != nil {
			j.logger.Error(fmt.Errorf("notify: notify miner: %v", err))
			continue
		}
	}

	if err := pooldb.UpdateWorkerSetActive(j.pooldb.Writer()); err != nil {
		j.logger.Error(fmt.Errorf("notify: set active: %v", err))
	}
}

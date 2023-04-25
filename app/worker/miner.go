package worker

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

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
		addToInactiveIPs := make([]string, 0)
		removeFromActiveIPs := make([]string, 0)
		removeFromInactiveIPs := make([]string, 0)
		for compoundID, timestamp := range ipAddressIdx {
			parts := strings.Split(compoundID, ":")
			if len(parts) != 3 {
				j.logger.Error(fmt.Errorf("ip: invalid compoundID: %s", compoundID))
				continue
			}

			minerID, err := strconv.ParseUint(parts[0], 10, 64)
			if err != nil || minerID == 0 {
				j.logger.Error(fmt.Errorf("ip: invalid compoundID: %s", compoundID))
				continue
			}

			workerID, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				j.logger.Error(fmt.Errorf("ip: invalid compoundID: %s", compoundID))
				continue
			}

			// need to replace "|" with ":" for IPv6 compatibility
			ipAddress := strings.ReplaceAll(parts[2], "|", ":")

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
		} else if err := j.redis.AddMinerIPAddressesInactive(node.Chain(), addToInactiveIPs); err != nil {
			j.logger.Error(fmt.Errorf("ip: add inactive: %s: %v", node.Chain(), err))
		} else if err := j.redis.RemoveMinerIPAddressesInactive(node.Chain(), removeFromInactiveIPs); err != nil {
			j.logger.Error(fmt.Errorf("ip: remove inactive: %s: %v", node.Chain(), err))
		} else if err := j.redis.RemoveMinerIPAddresses(node.Chain(), removeFromActiveIPs); err != nil {
			j.logger.Error(fmt.Errorf("ip: remove active: %s: %v", node.Chain(), err))
		}
	}
}

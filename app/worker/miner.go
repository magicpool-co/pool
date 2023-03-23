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
			j.logger.Error(fmt.Errorf("ip: fetch: %s: %v", node.Chain(), err))
		}

		rttIdx, err := j.redis.GetMinerLatencies(node.Chain())
		if err != nil {
			j.logger.Error(fmt.Errorf("ip: fetch: %s: %v", node.Chain(), err))
		}

		// process the index into a slice of addresses
		addresses := make([]*pooldb.IPAddress, 0)
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

			address := &pooldb.IPAddress{
				MinerID:   minerID,
				WorkerID:  workerID,
				ChainID:   node.Chain(),
				IPAddress: ipAddress,

				Active:        true,
				Expired:       false,
				LastShare:     time.Unix(int64(timestamp), 0),
				RoundTripTime: rtt,
			}
			addresses = append(addresses, address)
		}

		// insert the ip addresses, set old addresses to inactive or expired, clear the redis sorted set
		if err := pooldb.InsertIPAddresses(j.pooldb.Writer(), addresses...); err != nil {
			j.logger.Error(fmt.Errorf("ip: insert: %s: %v", node.Chain(), err))
		} else if err := pooldb.UpdateIPAddressesSetInactive(j.pooldb.Writer(), time.Hour); err != nil {
			j.logger.Error(fmt.Errorf("ip: update: %s: %v", node.Chain(), err))
		} else if err := pooldb.UpdateIPAddressesSetExpired(j.pooldb.Writer(), time.Hour*24); err != nil {
			j.logger.Error(fmt.Errorf("ip: update: %s: %v", node.Chain(), err))
		} else if err := j.redis.DeleteMinerIPAddresses(node.Chain()); err != nil {
			j.logger.Error(fmt.Errorf("ip: delete ips: %s: %v", node.Chain(), err))
		} else if err := j.redis.DeleteMinerLatencies(node.Chain()); err != nil {
			j.logger.Error(fmt.Errorf("ip: delete latencies: %s: %v", node.Chain(), err))
		}
	}
}

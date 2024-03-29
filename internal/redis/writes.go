package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/magicpool-co/pool/internal/tsdb"
)

/* channels */

func (c *Client) WriteToStreamMinerIndexChannel(msg string) error {
	key := c.getStreamMinerIndexChannelKey()
	return c.streamClusterClient.Publish(context.Background(), key, msg).Err()
}

func (c *Client) WriteToStreamMinerChannel(minerID uint64, msg string) error {
	key := c.getStreamMinerChannelKey(minerID)
	return c.streamClusterClient.Publish(context.Background(), key, msg).Err()
}

func (c *Client) WriteToStreamDebugIndexChannel(msg string) error {
	key := c.getStreamDebugIndexChannelKey()
	return c.streamClusterClient.Publish(context.Background(), key, msg).Err()
}

func (c *Client) WriteToStreamDebugChannel(ip, msg string) error {
	key := c.getStreamDebugChannelKey(ip)
	return c.streamClusterClient.Publish(context.Background(), key, msg).Err()
}

/* miner */

func (c *Client) SetMinerID(miner string, minerID uint64) error {
	return c.baseSet(c.getMinersKey(miner), strconv.FormatUint(minerID, 10))
}

func (c *Client) SetMinerIPAddressesBulk(chain string, values map[string]int64) error {
	members := make([]redis.Z, 0)
	for k, v := range values {
		members = append(members, redis.Z{Member: k, Score: float64(v)})
	}

	return c.baseZAddBatch(c.getMinerIPAddressesKey(chain), members)
}

func (c *Client) SetMinerDifficultiesBulk(chain string, values map[string]int64) error {
	members := make([]redis.Z, 0)
	for k, v := range values {
		members = append(members, redis.Z{Member: k, Score: float64(v)})
	}

	return c.baseZAddBatch(c.getMinerDifficultiesKey(chain), members)
}

func (c *Client) SetMinerLatenciesBulk(chain string, values map[string]int64) error {
	members := make([]redis.Z, 0)
	for k, v := range values {
		members = append(members, redis.Z{Member: k, Score: float64(v)})
	}

	return c.baseZAddBatch(c.getMinerLatenciesKey(chain), members)
}

func (c *Client) RemoveMinerIPAddresses(chain string, values []string) error {
	if len(values) == 0 {
		return nil
	}

	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	members := make([]interface{}, len(values))
	for i, value := range values {
		members[i] = value
	}

	// remove inactive ip addresses
	pipe.ZRem(ctx, c.getMinerIPAddressesKey(chain), members...)

	// remove inactive ip addresses from latencies
	pipe.ZRem(ctx, c.getMinerLatenciesKey(chain), members...)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) AddMinerIPAddressesInactive(chain string, values []string) error {
	return c.baseSAddMany(c.getMinerIPAddressesInactiveKey(chain), values...)
}

func (c *Client) RemoveMinerIPAddressesInactive(chain string, values []string) error {
	return c.baseSRemMany(c.getMinerIPAddressesInactiveKey(chain), values...)
}

func (c *Client) SetWorkerID(minerID uint64, worker string, workerID uint64) error {
	return c.baseSet(c.getWorkersKey(minerID, worker), strconv.FormatUint(workerID, 10))
}

func (c *Client) SetTopMinerIDsBulk(chain string, values map[uint64]float64, increment bool) error {
	if !increment {
		members := make([]redis.Z, 0)
		for k, v := range values {
			members = append(members, redis.Z{Member: strconv.FormatUint(k, 10), Score: float64(v)})
		}

		return c.baseZAddBatch(c.getTopMinersKey(chain), members)
	}

	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	for id, value := range values {
		pipe.ZIncrBy(ctx, c.getTopMinersKey(chain), value, strconv.FormatUint(id, 10))
	}

	_, err := pipe.Exec(ctx)

	return err
}

/* share index */

func (c *Client) AddShareIndexHeight(chain string, height uint64) error {
	_, err := c.baseSAdd(c.getShareIndexKey(chain), strconv.FormatUint(height, 10))
	return err
}

func (c *Client) DeleteShareIndexHeight(chain string, height uint64) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	// remove shares set
	pipe.Del(ctx, c.getUniqueSharesKey(chain, height))

	// remove height from the set
	pipe.SRem(ctx, c.getShareIndexKey(chain), strconv.FormatUint(height, 10))

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) AddUniqueShare(chain string, height uint64, hash string) (bool, error) {
	return c.baseSAdd(c.getUniqueSharesKey(chain, height), hash)
}

/* rounds */

func (c *Client) AddAcceptedShare(
	chain, interval, compoundID string,
	soloMinerID uint64,
	count int,
	window int64,
) error {
	if count <= 0 {
		return nil
	}

	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	// if there is no solo miner ID, push the shares to the global list
	// and increment the global round share counter. otherwise, increment
	// the miner specific round share counter
	if soloMinerID == 0 {
		for i := 0; i < count; i++ {
			pipe.LPush(ctx, c.getRoundSharesKey(chain), compoundID)
		}
		pipe.LTrim(ctx, c.getRoundSharesKey(chain), 0, window-1)
		pipe.IncrBy(ctx, c.getRoundAcceptedSharesKey(chain), int64(count))
	} else {
		pipe.IncrBy(ctx, c.getRoundSoloAcceptedSharesKey(chain, soloMinerID), int64(count))
	}

	pipe.ZIncrBy(ctx, c.getIntervalAcceptedSharesKey(chain, interval), float64(count), compoundID)
	pipe.ZIncrBy(ctx, c.getIntervalAcceptedAdjustedSharesKey(chain, interval), 1, compoundID)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) AddRejectedShare(
	chain, interval, compoundID string,
	soloMinerID uint64,
	count int,
) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	if soloMinerID == 0 {
		pipe.IncrBy(ctx, c.getRoundRejectedSharesKey(chain), int64(count))
	} else {
		pipe.IncrBy(ctx, c.getRoundSoloRejectedSharesKey(chain, soloMinerID), int64(count))
	}

	pipe.ZIncrBy(ctx, c.getIntervalRejectedSharesKey(chain, interval), float64(count), compoundID)
	pipe.ZIncrBy(ctx, c.getIntervalRejectedAdjustedSharesKey(chain, interval), 1, compoundID)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) AddInvalidShare(
	chain, interval, compoundID string,
	soloMinerID uint64,
	count int,
) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	if soloMinerID == 0 {
		pipe.IncrBy(ctx, c.getRoundInvalidSharesKey(chain), int64(count))
	} else {
		pipe.IncrBy(ctx, c.getRoundSoloInvalidSharesKey(chain, soloMinerID), int64(count))
	}

	pipe.ZIncrBy(ctx, c.getIntervalInvalidSharesKey(chain, interval), float64(count), compoundID)
	pipe.ZIncrBy(ctx, c.getIntervalInvalidAdjustedSharesKey(chain, interval), 1, compoundID)

	_, err := pipe.Exec(ctx)

	return err
}

/* interval */

func (c *Client) AddInterval(chain, interval string) error {
	_, err := c.baseSAdd(c.getIntervalsKey(chain), interval)
	return err
}

func (c *Client) DeleteInterval(chain, interval string) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	// remove miner accepted, rejected, and last shares
	pipe.Del(ctx, c.getIntervalAcceptedSharesKey(chain, interval))
	pipe.Del(ctx, c.getIntervalAcceptedAdjustedSharesKey(chain, interval))
	pipe.Del(ctx, c.getIntervalRejectedSharesKey(chain, interval))
	pipe.Del(ctx, c.getIntervalRejectedAdjustedSharesKey(chain, interval))
	pipe.Del(ctx, c.getIntervalInvalidSharesKey(chain, interval))
	pipe.Del(ctx, c.getIntervalInvalidAdjustedSharesKey(chain, interval))

	// remove interval from the set
	pipe.SRem(ctx, c.getIntervalsKey(chain), interval)

	_, err := pipe.Exec(ctx)

	return err
}

/* charts */

func (c *Client) SetChartSharesLastTime(chain string, timestamp time.Time) error {
	encoded := strconv.FormatInt(timestamp.UTC().Unix(), 10)
	return c.baseSet(c.getChartSharesLastTimeKey(chain), encoded)
}

func (c *Client) SetChartBlocksLastTime(chain string, timestamp time.Time) error {
	encoded := strconv.FormatInt(timestamp.UTC().Unix(), 10)
	return c.baseSet(c.getChartBlocksLastTimeKey(chain), encoded)
}

func (c *Client) SetChartRoundsLastTime(chain string, timestamp time.Time) error {
	encoded := strconv.FormatInt(timestamp.UTC().Unix(), 10)
	return c.baseSet(c.getChartRoundsLastTimeKey(chain), encoded)
}

func (c *Client) SetChartEarningsLastTime(chain string, timestamp time.Time) error {
	encoded := strconv.FormatInt(timestamp.UTC().Unix(), 10)
	return c.baseSet(c.getChartEarningsLastTimeKey(chain), encoded)
}

/* cached stats */

func (c *Client) SetCachedGlobalLastShares(shares []*tsdb.Share, exp time.Duration) error {
	encoded, err := encode(shares)
	if err != nil {
		return err
	}

	return c.baseSetExp(c.getCachedGlobalLastSharesKey(), encoded, exp)
}

func (c *Client) SetCachedGlobalLastProfits(blocks []*tsdb.Block, exp time.Duration) error {
	encoded, err := encode(blocks)
	if err != nil {
		return err
	}

	return c.baseSetExp(c.getCachedGlobalLastProfitsKey(), encoded, exp)
}

func (c *Client) SetCachedLuckByChain(chain string, luck float64, exp time.Duration) error {
	encoded := strconv.FormatFloat(luck, 'f', 8, 64)

	return c.baseSetExp(c.getCachedLuckByChainKey(chain), encoded, exp)
}

func (c *Client) SetCachedMinersByChain(chain string, miners int64, exp time.Duration) error {
	encoded := strconv.FormatInt(miners, 10)

	return c.baseSetExp(c.getCachedMinersByChainKey(chain), encoded, exp)
}

func (c *Client) SetCachedWorkersByChain(chain string, workers int64, exp time.Duration) error {
	encoded := strconv.FormatInt(workers, 10)

	return c.baseSetExp(c.getCachedWorkersByChainKey(chain), encoded, exp)
}

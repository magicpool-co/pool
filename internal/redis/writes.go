package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/magicpool-co/pool/internal/tsdb"
)

/* miner */

func (c *Client) SetMinerID(miner string, minerID uint64) error {
	return c.baseSet(c.getMinersKey(miner), strconv.FormatUint(minerID, 10))
}

func (c *Client) SetMinerIPAddressesBulk(chain string, values map[string]int64) error {
	members := make([]*redis.Z, 0)
	for k, v := range values {
		members = append(members, &redis.Z{Member: k, Score: float64(v)})
	}

	return c.baseZAddBatch(c.getMinerIPAddressesKey(chain), members)
}

func (c *Client) SetMinerLatenciesBulk(chain string, values map[string]int64) error {
	members := make([]*redis.Z, 0)
	for k, v := range values {
		members = append(members, &redis.Z{Member: k, Score: float64(v)})
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

func (c *Client) SetTopMinerIDs(chain string, minerIDs []uint64) error {
	values := make([]interface{}, len(minerIDs))
	for i, minerID := range minerIDs {
		values[i] = strconv.FormatUint(minerID, 10)
	}

	return c.baseResetList(c.getTopMinersKey(chain), values)
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

func (c *Client) AddAcceptedShare(chain, interval, compoundID string, count, window int64) error {
	if count <= 0 {
		return nil
	}

	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	for i := 0; i < int(count); i++ {
		pipe.LPush(ctx, c.getRoundSharesKey(chain), compoundID)
	}
	pipe.LTrim(ctx, c.getRoundSharesKey(chain), 0, window-1)
	pipe.IncrBy(ctx, c.getRoundAcceptedSharesKey(chain), count)
	pipe.ZIncrBy(ctx, c.getIntervalAcceptedSharesKey(chain, interval), float64(count), compoundID)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) AddRejectedShare(chain, interval, compoundID string) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	pipe.Incr(ctx, c.getRoundRejectedSharesKey(chain))
	pipe.ZIncrBy(ctx, c.getIntervalRejectedSharesKey(chain, interval), 1, compoundID)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) AddInvalidShare(chain, interval, compoundID string) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	pipe.Incr(ctx, c.getRoundInvalidSharesKey(chain))
	pipe.ZIncrBy(ctx, c.getIntervalInvalidSharesKey(chain, interval), 1, compoundID)

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
	pipe.Del(ctx, c.getIntervalRejectedSharesKey(chain, interval))
	pipe.Del(ctx, c.getIntervalInvalidSharesKey(chain, interval))

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

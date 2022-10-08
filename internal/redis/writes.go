package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
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

func (c *Client) DeleteMinerIPAddresses(chain string) error {
	return c.baseDel(c.getMinerIPAddressesKey(chain))
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

/* rounds */

func (c *Client) AddAcceptedShare(chain, interval, compoundID string, window int64) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	pipe.LPush(ctx, c.getRoundSharesKey(chain), compoundID)
	pipe.LTrim(ctx, c.getRoundSharesKey(chain), 0, window-1)
	pipe.Incr(ctx, c.getRoundAcceptedSharesKey(chain))
	pipe.ZIncrBy(ctx, c.getIntervalAcceptedSharesKey(chain, interval), 1, compoundID)

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
	return c.writeClient.SAdd(context.Background(), c.getIntervalsKey(chain), interval).Err()
}

func (c *Client) DeleteInterval(chain, interval string) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	// remove miner reported hashrate, accepted, rejected, and last shares
	pipe.Del(ctx, c.getIntervalAcceptedSharesKey(chain, interval))
	pipe.Del(ctx, c.getIntervalRejectedSharesKey(chain, interval))
	pipe.Del(ctx, c.getIntervalInvalidSharesKey(chain, interval))
	pipe.Del(ctx, c.getIntervalReportedHashratesKey(chain, interval))

	// remove interval from the set
	pipe.SRem(ctx, c.getIntervalsKey(chain), interval)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) SetIntervalReportedHashrateBatch(chain, interval string, values map[string]float64) error {
	members := make([]*redis.Z, 0)
	for k, v := range values {
		members = append(members, &redis.Z{Member: k, Score: v})
	}

	return c.baseZAddBatch(c.getIntervalReportedHashratesKey(chain, interval), members)
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

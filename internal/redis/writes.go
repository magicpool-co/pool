package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

func (c *Client) SetMinerID(addressChain string, minerID uint64) error {
	key := c.getKey("mnr", "id", addressChain)
	return c.writeClient.Set(context.Background(), key, strconv.FormatUint(minerID, 10), 0).Err()
}

func (c *Client) SetWorkerID(minerID uint64, workerName string, workerID uint64) error {
	key := c.getKey("wkr", "id", strconv.FormatUint(minerID, 10), workerName)
	return c.writeClient.Set(context.Background(), key, strconv.FormatUint(workerID, 10), 0).Err()
}

func (c *Client) AddMinerIPAddress(minerID uint64, ipAddress string) error {
	key := c.getKey("mnr", strconv.FormatUint(minerID, 10), "ip", ipAddress)
	return c.writeClient.SAdd(context.Background(), key, ipAddress).Err()
}

func (c *Client) SetChartSharesLastTime(chain string, period int, value time.Time) error {
	key := c.getKey("chrt", "sh", chain, strconv.Itoa(period))
	encoded := strconv.FormatInt(value.UTC().Unix(), 10)
	return c.writeClient.Set(context.Background(), key, encoded, 0).Err()
}

func (c *Client) SetChartBlocksLastTime(chain string, period int, value time.Time) error {
	key := c.getKey("chrt", "blk", chain, strconv.Itoa(period))
	encoded := strconv.FormatInt(value.UTC().Unix(), 10)
	return c.writeClient.Set(context.Background(), key, encoded, 0).Err()
}

func (c *Client) SetChartRoundsLastTime(chain string, period int, value time.Time) error {
	key := c.getKey("chrt", "rnd", chain, strconv.Itoa(period))
	encoded := strconv.FormatInt(value.UTC().Unix(), 10)
	return c.writeClient.Set(context.Background(), key, encoded, 0).Err()
}

func (c *Client) ResetRoundAcceptedShares(chain string) error {
	key := c.getKey(chain, "pstats", "round", "ash")
	return c.writeClient.Set(context.Background(), key, "0", 0).Err()
}

func (c *Client) ResetRoundRejectedShares(chain string) error {
	key := c.getKey(chain, "pstats", "round", "rsh")
	return c.writeClient.Set(context.Background(), key, "0", 0).Err()
}

func (c *Client) ResetRoundInvalidShares(chain string) error {
	key := c.getKey(chain, "pstats", "round", "ish")
	return c.writeClient.Set(context.Background(), key, "0", 0).Err()
}

func (c *Client) AddInterval(chain, interval string) error {
	key := c.getKey(chain, "intrvl")
	return c.writeClient.SAdd(context.Background(), key, interval).Err()
}

func (c *Client) DeleteInterval(chain, interval string) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	// remove miner reported hashrate, accepted, rejected, and last shares
	pipe.Del(ctx, c.getKey(chain, "pstats", interval, "rhr"))
	pipe.Del(ctx, c.getKey(chain, "pstats", interval, "ash"))
	pipe.Del(ctx, c.getKey(chain, "pstats", interval, "rsh"))
	pipe.Del(ctx, c.getKey(chain, "pstats", interval, "ish"))

	// remove interval from the set
	pipe.SRem(ctx, c.getKey(chain, "intrvl"), interval)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) AddAcceptedShare(chain, interval, id string, window int64) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	// push miner to round shares and trim to window size, increment round accepted shares
	pipe.LPush(ctx, c.getKey(chain, "pstats", "window", "ash"), id)
	pipe.LTrim(ctx, c.getKey(chain, "pstats", "window", "ash"), 0, window-1)
	pipe.Incr(ctx, c.getKey(chain, "pstats", "round", "ash"))

	// increment miner accepted shares
	pipe.ZIncrBy(ctx, c.getKey(chain, "pstats", interval, "ash"), 1, id)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) AddRejectedShare(chain, interval, id string) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	// incr round rejected shares
	pipe.Incr(ctx, c.getKey(chain, "pstats", "round", "rsh"))

	// increment miner rejected shares
	pipe.ZIncrBy(ctx, c.getKey(chain, "pstats", interval, "rsh"), 1, id)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) AddInvalidShare(chain, interval, id string) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	// incr round invalid shares
	pipe.Incr(ctx, c.getKey(chain, "pstats", "round", "ish"))

	// increment miner invalid shares
	pipe.ZIncrBy(ctx, c.getKey(chain, "pstats", interval, "ish"), 1, id)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) SetReportedHashrateBulk(chain, interval string, values map[string]float64) error {
	key := c.getKey(chain, "pstats", interval, "rhr")

	members := make([]*redis.Z, 0)
	for k, v := range values {
		members = append(members, &redis.Z{Member: k, Score: v})
	}

	batchSize, count := 500, len(members)
	for i := 0; i <= (count / batchSize); i++ {
		start := i * batchSize
		end := (i + 1) * batchSize
		if end >= count {
			end = count
		}

		if start >= end {
			continue
		}

		if err := c.writeClient.ZAdd(context.Background(), key, members[start:end]...).Err(); err != nil {
			return err
		}
	}

	return nil
}

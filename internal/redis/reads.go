package redis

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

/* miners */

func (c *Client) GetMinerID(miner string) (uint64, error) {
	return c.baseGetUint64(c.getMinersKey(miner))
}

func (c *Client) GetMinerIPAddresses(chain string) (map[string]float64, error) {
	return c.baseZRangeWithScores(c.getMinerIPAddressesKey(chain))
}

func (c *Client) GetMinerLatencies(chain string) (map[string]float64, error) {
	return c.baseZRangeWithScores(c.getMinerLatenciesKey(chain))
}

func (c *Client) GetWorkerID(minerID uint64, worker string) (uint64, error) {
	return c.baseGetUint64(c.getWorkersKey(minerID, worker))
}

func (c *Client) GetTopMinerIDs(chain string) ([]uint64, error) {
	ctx := context.Background()
	results, err := c.readClient.LRange(ctx, c.getTopMinersKey(chain), 0, -1).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	values := make([]uint64, len(results))
	for i, result := range results {
		values[i], err = strconv.ParseUint(result, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

/* share index */

func (c *Client) GetShareIndexes(chain string) ([]string, error) {
	return c.baseSMembers(c.getShareIndexKey(chain))
}

/* rounds */

func (c *Client) GetRoundShares(chain string) (map[uint64]uint64, error) {
	key := c.getRoundSharesKey(chain)
	raw, err := c.readClient.LRange(context.Background(), key, 0, -1).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	values := make(map[uint64]uint64)
	for _, k := range raw {
		parts := strings.Split(k, ":")
		id, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}
		values[id]++
	}

	return values, nil
}

func (c *Client) GetRoundShareCounts(chain string) (uint64, uint64, uint64, error) {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	acceptedRaw := pipe.Get(ctx, c.getRoundAcceptedSharesKey(chain))
	rejectedRaw := pipe.Get(ctx, c.getRoundRejectedSharesKey(chain))
	invalidRaw := pipe.Get(ctx, c.getRoundInvalidSharesKey(chain))

	pipe.Set(ctx, c.getRoundAcceptedSharesKey(chain), "0", 0)
	pipe.Set(ctx, c.getRoundRejectedSharesKey(chain), "0", 0)
	pipe.Set(ctx, c.getRoundInvalidSharesKey(chain), "0", 0)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	var acceptedStr, rejectedStr, invalidStr string
	if acceptedStr, err = acceptedRaw.Result(); err != nil && err != redis.Nil {
		return 0, 0, 0, err
	} else if rejectedStr, err = rejectedRaw.Result(); err != nil && err != redis.Nil {
		return 0, 0, 0, err
	} else if invalidStr, err = invalidRaw.Result(); err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	var accepted, rejected, invalid uint64
	if accepted, err = strconv.ParseUint(acceptedStr, 10, 64); err != nil && acceptedStr != "" {
		return 0, 0, 0, err
	} else if rejected, err = strconv.ParseUint(rejectedStr, 10, 64); err != nil && rejectedStr != "" {
		return 0, 0, 0, err
	} else if invalid, err = strconv.ParseUint(invalidStr, 10, 64); err != nil && invalidStr != "" {
		return 0, 0, 0, err
	}

	return accepted, rejected, invalid, nil
}

/* intervals */

func (c *Client) GetIntervals(chain string) ([]string, error) {
	return c.baseSMembers(c.getIntervalsKey(chain))
}

func (c *Client) GetIntervalAcceptedShares(chain, interval string) (map[string]uint64, error) {
	return c.baseZRangeWithScoresUint64(c.getIntervalAcceptedSharesKey(chain, interval))
}

func (c *Client) GetIntervalRejectedShares(chain, interval string) (map[string]uint64, error) {
	return c.baseZRangeWithScoresUint64(c.getIntervalRejectedSharesKey(chain, interval))
}

func (c *Client) GetIntervalInvalidShares(chain, interval string) (map[string]uint64, error) {
	return c.baseZRangeWithScoresUint64(c.getIntervalInvalidSharesKey(chain, interval))
}

func (c *Client) GetIntervalReportedHashrates(chain, interval string) (map[string]float64, error) {
	return c.baseZRangeWithScores(c.getIntervalReportedHashratesKey(chain, interval))
}

/* chart */

func (c *Client) GetChartSharesLastTime(chain string) (time.Time, error) {
	return c.baseGetTime(c.getChartSharesLastTimeKey(chain))
}

func (c *Client) GetChartBlocksLastTime(chain string) (time.Time, error) {
	return c.baseGetTime(c.getChartBlocksLastTimeKey(chain))
}

func (c *Client) GetChartRoundsLastTime(chain string) (time.Time, error) {
	return c.baseGetTime(c.getChartRoundsLastTimeKey(chain))
}

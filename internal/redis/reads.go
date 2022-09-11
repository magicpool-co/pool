package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

func (c *Client) baseGetUint64(key string) (uint64, error) {
	value, err := c.readClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return strconv.ParseUint(value, 10, 64)
}

func (c *Client) baseGetTime(key string) (time.Time, error) {
	raw, err := c.readClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return time.Time{}, nil
	} else if err != nil {
		return time.Time{}, err
	}

	encoded, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(encoded, 0).UTC(), nil
}

func (c *Client) baseIsMember(key, member string) (bool, error) {
	value, err := c.readClient.SIsMember(context.Background(), key, member).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return value, nil
}

func (c *Client) baseZRangeWithScores(key string) (map[string]float64, error) {
	args := redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  -1,
	}

	results, err := c.readClient.ZRangeArgsWithScores(context.Background(), args).Result()
	if err != nil {
		return nil, err
	}

	values := make(map[string]float64, len(results))
	for _, result := range results {
		if member, ok := result.Member.(string); !ok {
			return nil, fmt.Errorf("unable to cast member %v", result.Member)
		} else {
			values[member] = result.Score
		}
	}

	return values, nil
}

func (c *Client) baseZRangeWithScoresUint64(key string) (map[string]uint64, error) {
	raw, err := c.baseZRangeWithScores(key)
	if err != nil {
		return nil, err
	}

	values := make(map[string]uint64)
	for k, v := range raw {
		values[k] = uint64(v)
	}

	return values, nil
}

func (c *Client) GetMinerID(addressChain string) (uint64, error) {
	key := c.getKey("mnr", "id", addressChain)
	return c.baseGetUint64(key)
}

func (c *Client) GetWorkerID(minerID uint64, workerName string) (uint64, error) {
	key := c.getKey("wkr", "id", strconv.FormatUint(minerID, 10), workerName)
	return c.baseGetUint64(key)
}

func (c *Client) GetMinerIPAddressExists(minerID uint64, ipAddress string) (bool, error) {
	key := c.getKey("mnr", strconv.FormatUint(minerID, 10), "ip")
	return c.baseIsMember(key, ipAddress)
}

func (c *Client) GetChartSharesLastTime(chain string, period int) (time.Time, error) {
	key := c.getKey("chrt", "sh", chain, strconv.Itoa(period))
	return c.baseGetTime(key)
}

func (c *Client) GetChartBlocksLastTime(chain string, period int) (time.Time, error) {
	key := c.getKey("chrt", "blk", chain, strconv.Itoa(period))
	return c.baseGetTime(key)
}

func (c *Client) GetChartRoundsLastTime(chain string, period int) (time.Time, error) {
	key := c.getKey("chrt", "rnd", chain, strconv.Itoa(period))
	return c.baseGetTime(key)
}

func (c *Client) GetShares(chain string) ([]string, error) {
	key := c.getKey(chain, "sh")
	return c.readClient.LRange(context.Background(), key, 0, -1).Result()
}

func (c *Client) GetIntervals(chain string) ([]string, error) {
	key := c.getKey(chain, "intrvl")
	return c.readClient.SMembers(context.Background(), key).Result()
}

func (c *Client) FetchShares(chain string) ([]string, error) {
	key := c.getKey(chain, "pstats", "window", "ash")
	return c.readClient.LRange(context.Background(), key, 0, -1).Result()
}

func (c *Client) FetchRoundAcceptedShares(chain string) (uint64, error) {
	key := c.getKey(chain, "pstats", "round", "ash")
	return c.baseGetUint64(key)
}

func (c *Client) FetchRoundRejectedShares(chain string) (uint64, error) {
	key := c.getKey(chain, "pstats", "round", "rsh")
	return c.baseGetUint64(key)
}

func (c *Client) FetchRoundInvalidShares(chain string) (uint64, error) {
	key := c.getKey(chain, "pstats", "round", "ish")
	return c.baseGetUint64(key)
}

func (c *Client) FetchMinerAcceptedShares(chain, interval string) (map[string]uint64, error) {
	key := c.getKey(chain, "pstats", interval, "ash")
	return c.baseZRangeWithScoresUint64(key)
}

func (c *Client) FetchMinerRejectedShares(chain, interval string) (map[string]uint64, error) {
	key := c.getKey(chain, "pstats", interval, "rsh")
	return c.baseZRangeWithScoresUint64(key)
}

func (c *Client) FetchMinerInvalidShares(chain, interval string) (map[string]uint64, error) {
	key := c.getKey(chain, "pstats", interval, "ish")
	return c.baseZRangeWithScoresUint64(key)
}

func (c *Client) FetchMinerReportedHashrates(chain, interval string) (map[string]float64, error) {
	key := c.getKey(chain, "pstats", interval, "rhr")
	return c.baseZRangeWithScores(key)
}

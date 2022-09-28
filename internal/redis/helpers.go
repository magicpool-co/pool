package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

/* reads */

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

func (c *Client) baseSMembers(key string) ([]string, error) {
	values, err := c.readClient.SMembers(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return values, nil
}

func (c *Client) baseZRangeWithScores(key string) (map[string]float64, error) {
	args := redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  -1,
	}

	results, err := c.readClient.ZRangeArgsWithScores(context.Background(), args).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
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

/* writes */

func (c *Client) baseSet(key, value string) error {
	return c.writeClient.Set(context.Background(), key, value, 0).Err()
}

func (c *Client) baseDel(key string) error {
	return c.writeClient.Del(context.Background(), key).Err()
}

func (c *Client) baseZAddBatch(key string, members []*redis.Z) error {
	if len(members) == 0 {
		return nil
	}

	const batchSize = 500
	count := len(members)
	for start := 0; start < count; start += batchSize {
		end := start + batchSize
		if end > count {
			end = count
		}

		_, err := c.writeClient.ZAdd(context.Background(), key, members[start:end]...).Result()
		if err != nil {
			return err
		}
	}

	return nil
}

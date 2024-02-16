package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

/* encoding */

func encode(v interface{}) (string, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return "", err
	}

	return string(buf.Bytes()), nil
}

func decode(val string, v interface{}) error {
	buf := bytes.NewBuffer([]byte(val))
	enc := gob.NewDecoder(buf)

	return enc.Decode(v)
}

/* reads */

func (c *Client) baseGet(key string) (string, error) {
	value, err := c.readClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return value, nil
}

func (c *Client) baseGetInt64(key string) (int64, error) {
	value, err := c.readClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return -1, nil
	} else if err != nil {
		return 0, err
	}

	return strconv.ParseInt(value, 10, 64)
}

func (c *Client) baseGetUint64(key string) (uint64, error) {
	value, err := c.readClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return strconv.ParseUint(value, 10, 64)
}

func (c *Client) baseGetFloat64(key string) (float64, error) {
	value, err := c.readClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(value, 64)
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

func (c *Client) baseZRange(key string, limit int64, reverse bool) ([]string, error) {
	args := redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  limit,
		Rev:   reverse,
	}

	results, err := c.readClient.ZRangeArgs(context.Background(), args).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return results, nil
}

func (c *Client) baseZRangeWithScores(key string) (map[string]float64, error) {
	args := redis.ZRangeArgs{
		Key:   key,
		Start: 0,
		Stop:  -1,
	}

	ctx := context.Background()
	results, err := c.readClient.ZRangeArgsWithScores(ctx, args).Result()
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

func (c *Client) baseSetExp(key, value string, expiration time.Duration) error {
	return c.writeClient.Set(context.Background(), key, value, expiration).Err()
}

func (c *Client) baseDel(key string) error {
	return c.writeClient.Del(context.Background(), key).Err()
}

func (c *Client) baseSAdd(key, value string) (bool, error) {
	res, err := c.writeClient.SAdd(context.Background(), key, value).Result()
	return res == 1, err
}

func (c *Client) baseSAddMany(key string, values ...string) error {
	if len(values) == 0 {
		return nil
	}

	members := make([]interface{}, len(values))
	for i, value := range values {
		members[i] = value
	}

	return c.writeClient.SAdd(context.Background(), key, members...).Err()
}

func (c *Client) baseSRemMany(key string, values ...string) error {
	if len(values) == 0 {
		return nil
	}

	members := make([]interface{}, len(values))
	for i, value := range values {
		members[i] = value
	}

	return c.writeClient.SRem(context.Background(), key, members...).Err()
}

func (c *Client) baseResetList(key string, values []interface{}) error {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	pipe.Del(ctx, key)
	pipe.LPush(ctx, key, values...)

	_, err := pipe.Exec(ctx)

	return err
}

func (c *Client) baseZAddBatch(key string, members []redis.Z) error {
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

		ctx := context.Background()
		_, err := c.writeClient.ZAdd(ctx, key, members[start:end]...).Result()
		if err != nil {
			return err
		}
	}

	return nil
}

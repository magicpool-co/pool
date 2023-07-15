package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	env                 string
	readClient          *redis.Client
	writeClient         *redis.Client
	streamClusterClient *redis.ClusterClient
}

func newClient(addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        "",
		DB:              0,
		PoolSize:        25,
		PoolTimeout:     2 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
		ReadTimeout:     2 * time.Minute,
		WriteTimeout:    1 * time.Minute,
	})

	return client
}

func newClusterClient(addrs []string) *redis.ClusterClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           addrs,
		Password:        "",
		PoolSize:        25,
		PoolTimeout:     2 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
		ReadTimeout:     2 * time.Minute,
		WriteTimeout:    1 * time.Minute,
	})

	return client
}

func New(args map[string]string) (*Client, error) {
	var argKeys = []string{"REDIS_WRITE_HOST", "REDIS_READ_HOST", "REDIS_PORT"}
	for _, k := range argKeys {
		if _, ok := args[k]; !ok {
			return nil, fmt.Errorf("%s is a required argument", k)
		}
	}

	env := args["ENVIRONMENT"]
	if env == "" {
		env = "dev"
	}

	writeHost := args["REDIS_WRITE_HOST"]
	readHost := args["REDIS_READ_HOST"]
	port := args["REDIS_PORT"]

	var streamClusterClient *redis.ClusterClient
	if streamHosts, ok := args["REDIS_STREAM_HOSTS"]; ok {
		addrs := strings.Split(streamHosts, ",")
		for i := range addrs {
			addrs[i] += ":" + port
		}
		streamClusterClient = newClusterClient(addrs)
	}

	client := &Client{
		env:                 env,
		readClient:          newClient(readHost + ":" + port),
		writeClient:         newClient(writeHost + ":" + port),
		streamClusterClient: streamClusterClient,
	}

	return client, client.Ping()
}

func (c *Client) Ping() error {
	ctx := context.Background()
	err := c.writeClient.Ping(ctx).Err()
	if err != nil {
		return err
	}

	err = c.readClient.Ping(ctx).Err()
	if err != nil {
		return err
	} else if c.streamClusterClient == nil {
		return nil
	}

	err = c.streamClusterClient.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
		return shard.Ping(ctx).Err()
	})

	return err
}

func (c *Client) NewRateLimiter() *redis_rate.Limiter {
	return redis_rate.NewLimiter(c.writeClient)
}

func (c *Client) NewLocker() *redislock.Client {
	return redislock.New(c.writeClient)
}

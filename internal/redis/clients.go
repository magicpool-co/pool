package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
)

type Client struct {
	env         string
	readClient  *redis.Client
	writeClient *redis.Client
}

func newClient(addr string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "",
		DB:           db,
		PoolSize:     25,
		PoolTimeout:  2 * time.Minute,
		IdleTimeout:  10 * time.Minute,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	})

	return client
}

func New(args map[string]string) (*Client, error) {
	var argKeys = []string{"REDIS_WRITE_HOST", "REDIS_READ_HOST", "REDIS_PORT", "REDIS_DB"}
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
	db, err := strconv.Atoi(args["REDIS_DB"])
	if err != nil {
		return nil, err
	}

	writeClient := newClient(writeHost+":"+port, db)
	readClient := newClient(readHost+":"+port, db)

	client := &Client{
		env:         env,
		readClient:  readClient,
		writeClient: writeClient,
	}

	return client, client.Ping()
}

func (c *Client) getKey(args ...string) string {
	key := c.env
	for _, arg := range args {
		key += ":" + strings.ToLower(arg)
	}

	return key
}

func (c *Client) Ping() error {
	err := c.writeClient.Ping(context.Background()).Err()
	if err != nil {
		return err
	}

	return c.readClient.Ping(context.Background()).Err()
}

func (c *Client) NewRateLimiter() *redis_rate.Limiter {
	return redis_rate.NewLimiter(c.writeClient)
}

func (c *Client) NewLocker() *redislock.Client {
	return redislock.New(c.writeClient)
}

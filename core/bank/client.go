package bank

import (
	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/pkg/dbcl"
)

type Client struct {
	pooldb *dbcl.Client
	redis  *redis.Client
	locker *redislock.Client
}

func New(pooldbClient *dbcl.Client, redisClient *redis.Client) *Client {
	client := &Client{
		pooldb: pooldbClient,
		redis:  redisClient,
		locker: redisClient.NewLocker(),
	}

	return client
}

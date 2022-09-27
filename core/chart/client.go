package chart

import (
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/pkg/dbcl"
)

type Client struct {
	pooldb *dbcl.Client
	tsdb   *dbcl.Client
	redis  *redis.Client
}

func New(pooldbClient, tsdbClient *dbcl.Client, redisClient *redis.Client) *Client {
	client := &Client{
		pooldb: pooldbClient,
		tsdb:   tsdbClient,
		redis:  redisClient,
	}

	return client
}

package stats

import (
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/pkg/dbcl"
)

type Client struct {
	pooldb *dbcl.Client
	tsdb   *dbcl.Client
	redis  *redis.Client
	chains map[string]bool
}

func New(pooldbClient, tsdbClient *dbcl.Client, redisClient *redis.Client, chains []string) *Client {
	chainIdx := make(map[string]bool)
	for _, chain := range chains {
		chainIdx[chain] = true
	}

	client := &Client{
		pooldb: pooldbClient,
		tsdb:   tsdbClient,
		redis:  redisClient,
		chains: chainIdx,
	}

	return client
}

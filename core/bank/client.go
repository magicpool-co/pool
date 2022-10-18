package bank

import (
	"context"
	"strings"
	"time"

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

func (c *Client) fetchLock(chain string) (*redislock.Lock, error) {
	ctx := context.Background()
	key := "payout:" + strings.ToLower(chain) + ":prep"
	lock, err := c.locker.Obtain(ctx, key, time.Minute*15, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			return nil, err
		}
		return nil, nil
	}

	return lock, nil
}

package bank

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bsm/redislock"

	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/dbcl"
)

type Client struct {
	pooldb   *dbcl.Client
	redis    *redis.Client
	locker   *redislock.Client
	telegram *telegram.Client
}

func New(
	pooldbClient *dbcl.Client,
	redisClient *redis.Client,
	telegramClient *telegram.Client,
) *Client {
	client := &Client{
		pooldb:   pooldbClient,
		redis:    redisClient,
		locker:   redisClient.NewLocker(),
		telegram: telegramClient,
	}

	return client
}

func (c *Client) FetchLock(chain string) (*redislock.Lock, error) {
	ctx := context.Background()
	key := "tx:" + strings.ToLower(chain) + ":bank"
	lock, err := c.locker.Obtain(ctx, key, time.Minute*15, nil)
	if err != nil {
		if err != redislock.ErrNotObtained {
			return nil, err
		}
		return nil, fmt.Errorf("unable to retrieve lock")
	}

	return lock, nil
}

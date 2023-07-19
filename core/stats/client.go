package stats

import (
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

const (
	cacheDuration = time.Minute
)

type Client struct {
	useCache bool
	pooldb   *dbcl.Client
	tsdb     *dbcl.Client
	redis    *redis.Client
	chains   map[string]bool
}

func New(
	pooldbClient, tsdbClient *dbcl.Client,
	redisClient *redis.Client,
	chains []string,
	cacheEnabled bool,
) *Client {
	chainIdx := make(map[string]bool)
	for _, chain := range chains {
		chainIdx[chain] = true
	}

	client := &Client{
		useCache: cacheEnabled,
		pooldb:   pooldbClient,
		tsdb:     tsdbClient,
		redis:    redisClient,
		chains:   chainIdx,
	}

	return client
}

func (c *Client) processChainID(chainID string) (string, bool) {
	if c.chains[chainID] {
		return chainID, true
	} else if len(chainID) > 1 && c.chains[chainID[1:]] {
		return chainID[1:] + " SOLO", true
	}

	return "", false
}

func (c *Client) getGlobalSharesLast() ([]*tsdb.Share, error) {
	if c.useCache {
		shares, err := c.redis.GetCachedGlobalLastShares()
		if err == nil && len(shares) > 0 {
			return shares, nil
		}
	}

	shares, err := tsdb.GetGlobalSharesLast(c.tsdb.Reader(), int(types.Period15m))
	if err != nil {
		return nil, err
	}

	if c.useCache && len(shares) > 0 {
		go c.redis.SetCachedGlobalLastShares(shares, cacheDuration)
	}

	return shares, err
}

func (c *Client) getBlocksWithProfitabilityLast() ([]*tsdb.Block, error) {
	if c.useCache {
		blocks, err := c.redis.GetCachedGlobalLastProfits()
		if err == nil && len(blocks) > 0 {
			return blocks, nil
		}
	}

	blocks, err := tsdb.GetBlocksWithProfitabilityLast(c.tsdb.Reader(), int(types.Period15m))
	if err != nil {
		return nil, err
	}

	if c.useCache && len(blocks) > 0 {
		go c.redis.SetCachedGlobalLastProfits(blocks, cacheDuration)
	}

	return blocks, err
}

func (c *Client) getRoundLuckByChain(chain string, solo bool) (float64, error) {
	soloChain := chain
	if solo {
		chain = "S" + chain
	}

	if c.useCache {
		luck, err := c.redis.GetCachedLuckByChain(soloChain)
		if err == nil && luck > 0 {
			return luck, nil
		}
	}

	luck, err := pooldb.GetRoundLuckByChain(c.pooldb.Reader(), chain, solo, time.Hour*24*7)
	if err != nil {
		return 0.0, err
	}

	if c.useCache {
		go c.redis.SetCachedLuckByChain(soloChain, luck, cacheDuration)
	}

	return luck, err
}

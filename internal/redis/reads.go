package redis

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/magicpool-co/pool/internal/tsdb"
)

/* channels */

type PubSub struct {
	ctx    context.Context
	cancel context.CancelFunc
	pubsub *redis.PubSub
}

func (p *PubSub) Channel() <-chan *redis.Message {
	return p.pubsub.Channel()
}

func (p *PubSub) Close() {
	p.cancel()
	p.pubsub.Close()
}

func (c *Client) getClusterChannel(key string) (*PubSub, error) {
	if c.streamClusterClient == nil {
		return nil, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	rawPubsub := c.streamClusterClient.SSubscribe(ctx, key)
	_, err := rawPubsub.Receive(ctx)
	if err != nil {
		return nil, err
	}

	pubsub := &PubSub{
		ctx:    ctx,
		cancel: cancel,
		pubsub: rawPubsub,
	}

	return pubsub, nil
}

func (c *Client) GetStreamIndexChannel() (*PubSub, error) {
	return c.getClusterChannel(c.getStreamIndexChannelKey())
}

func (c *Client) GetStreamMinerChannel(minerID uint64) (*PubSub, error) {
	return c.getClusterChannel(c.getStreamMinerChannelKey(minerID))
}

/* miners */

func (c *Client) GetMinerID(miner string) (uint64, error) {
	return c.baseGetUint64(c.getMinersKey(miner))
}

func (c *Client) GetMinerIPAddresses(chain string) (map[string]float64, error) {
	return c.baseZRangeWithScores(c.getMinerIPAddressesKey(chain))
}

func (c *Client) GetMinerLatencies(chain string) (map[string]float64, error) {
	return c.baseZRangeWithScores(c.getMinerLatenciesKey(chain))
}

func (c *Client) GetMinerIPAddressesInactive(chain string) (map[string]bool, error) {
	values, err := c.baseSMembers(c.getMinerIPAddressesInactiveKey(chain))
	if err != nil {
		return nil, err
	}

	idx := make(map[string]bool)
	for _, value := range values {
		idx[value] = true
	}

	return idx, nil
}

func (c *Client) GetWorkerID(minerID uint64, worker string) (uint64, error) {
	return c.baseGetUint64(c.getWorkersKey(minerID, worker))
}

func (c *Client) GetTopMinerIDs(chain string) ([]uint64, error) {
	ctx := context.Background()
	results, err := c.readClient.LRange(ctx, c.getTopMinersKey(chain), 0, -1).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	values := make([]uint64, len(results))
	for i, result := range results {
		values[i], err = strconv.ParseUint(result, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

/* share index */

func (c *Client) GetShareIndexes(chain string) ([]string, error) {
	return c.baseSMembers(c.getShareIndexKey(chain))
}

/* rounds */

func (c *Client) GetRoundShares(chain string) (map[uint64]uint64, error) {
	key := c.getRoundSharesKey(chain)
	raw, err := c.readClient.LRange(context.Background(), key, 0, -1).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	values := make(map[uint64]uint64)
	for _, k := range raw {
		parts := strings.Split(k, ":")
		id, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}
		values[id]++
	}

	return values, nil
}

func (c *Client) GetRoundShareCounts(chain string) (uint64, uint64, uint64, error) {
	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	acceptedRaw := pipe.Get(ctx, c.getRoundAcceptedSharesKey(chain))
	rejectedRaw := pipe.Get(ctx, c.getRoundRejectedSharesKey(chain))
	invalidRaw := pipe.Get(ctx, c.getRoundInvalidSharesKey(chain))

	pipe.Set(ctx, c.getRoundAcceptedSharesKey(chain), "0", 0)
	pipe.Set(ctx, c.getRoundRejectedSharesKey(chain), "0", 0)
	pipe.Set(ctx, c.getRoundInvalidSharesKey(chain), "0", 0)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	var acceptedStr, rejectedStr, invalidStr string
	if acceptedStr, err = acceptedRaw.Result(); err != nil && err != redis.Nil {
		return 0, 0, 0, err
	} else if rejectedStr, err = rejectedRaw.Result(); err != nil && err != redis.Nil {
		return 0, 0, 0, err
	} else if invalidStr, err = invalidRaw.Result(); err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	var accepted, rejected, invalid uint64
	if accepted, err = strconv.ParseUint(acceptedStr, 10, 64); err != nil && acceptedStr != "" {
		return 0, 0, 0, err
	} else if rejected, err = strconv.ParseUint(rejectedStr, 10, 64); err != nil && rejectedStr != "" {
		return 0, 0, 0, err
	} else if invalid, err = strconv.ParseUint(invalidStr, 10, 64); err != nil && invalidStr != "" {
		return 0, 0, 0, err
	}

	return accepted, rejected, invalid, nil
}

/* intervals */

func (c *Client) GetIntervals(chain string) ([]string, error) {
	return c.baseSMembers(c.getIntervalsKey(chain))
}

func (c *Client) GetIntervalAcceptedShares(chain, interval string) (map[string]uint64, map[string]uint64, error) {
	raw, err := c.baseZRangeWithScoresUint64(c.getIntervalAcceptedSharesKey(chain, interval))
	if err != nil {
		return nil, nil, err
	}

	adjusted, err := c.baseZRangeWithScoresUint64(c.getIntervalAcceptedAdjustedSharesKey(chain, interval))
	if err != nil {
		return nil, nil, err
	}

	return raw, adjusted, nil
}

func (c *Client) GetIntervalRejectedShares(chain, interval string) (map[string]uint64, map[string]uint64, error) {
	raw, err := c.baseZRangeWithScoresUint64(c.getIntervalRejectedSharesKey(chain, interval))
	if err != nil {
		return nil, nil, err
	}

	adjusted, err := c.baseZRangeWithScoresUint64(c.getIntervalRejectedAdjustedSharesKey(chain, interval))
	if err != nil {
		return nil, nil, err
	}

	return raw, adjusted, nil
}

func (c *Client) GetIntervalInvalidShares(chain, interval string) (map[string]uint64, map[string]uint64, error) {
	raw, err := c.baseZRangeWithScoresUint64(c.getIntervalInvalidSharesKey(chain, interval))
	if err != nil {
		return nil, nil, err
	}

	adjusted, err := c.baseZRangeWithScoresUint64(c.getIntervalInvalidAdjustedSharesKey(chain, interval))
	if err != nil {
		return nil, nil, err
	}

	return raw, adjusted, nil
}

/* chart */

func (c *Client) GetChartSharesLastTime(chain string) (time.Time, error) {
	return c.baseGetTime(c.getChartSharesLastTimeKey(chain))
}

func (c *Client) GetChartBlocksLastTime(chain string) (time.Time, error) {
	return c.baseGetTime(c.getChartBlocksLastTimeKey(chain))
}

func (c *Client) GetChartRoundsLastTime(chain string) (time.Time, error) {
	return c.baseGetTime(c.getChartRoundsLastTimeKey(chain))
}

/* cached stats */

func (c *Client) GetCachedGlobalLastShares() ([]*tsdb.Share, error) {
	encoded, err := c.baseGet(c.getCachedGlobalLastSharesKey())
	if err != nil {
		return nil, err
	} else if len(encoded) == 0 {
		return nil, nil
	}

	var shares []*tsdb.Share
	err = decode(encoded, &shares)
	if err != nil {
		return nil, err
	}

	return shares, nil
}

func (c *Client) GetCachedGlobalLastProfits() ([]*tsdb.Block, error) {
	encoded, err := c.baseGet(c.getCachedGlobalLastProfitsKey())
	if err != nil {
		return nil, err
	} else if len(encoded) == 0 {
		return nil, nil
	}

	var blocks []*tsdb.Block
	err = decode(encoded, &blocks)
	if err != nil {
		return nil, err
	}

	return blocks, nil
}

func (c *Client) GetCachedLuckByChain(chain string) (float64, error) {
	return c.baseGetFloat64(c.getCachedLuckByChainKey(chain))
}

func (c *Client) GetCachedMinersByChain(chain string) (int64, error) {
	return c.baseGetInt64(c.getCachedMinersByChainKey(chain))
}

func (c *Client) GetCachedWorkersByChain(chain string) (int64, error) {
	return c.baseGetInt64(c.getCachedWorkersByChainKey(chain))
}

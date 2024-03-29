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
	if p.pubsub == nil {
		return nil
	}

	return p.pubsub.Channel()
}

func (p *PubSub) Close() {
	if p.pubsub != nil {
		p.cancel()
		p.pubsub.Close()
	}
}

func (c *Client) getClusterChannel(key string) (*PubSub, error) {
	if c.streamClusterClient == nil {
		return &PubSub{}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	rawPubsub := c.streamClusterClient.Subscribe(ctx, key)
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

func (c *Client) GetStreamMinerIndexChannel() (*PubSub, error) {
	return c.getClusterChannel(c.getStreamMinerIndexChannelKey())
}

func (c *Client) GetStreamMinerChannel(minerID uint64) (*PubSub, error) {
	return c.getClusterChannel(c.getStreamMinerChannelKey(minerID))
}

func (c *Client) GetStreamDebugIndexChannel() (*PubSub, error) {
	return c.getClusterChannel(c.getStreamDebugIndexChannelKey())
}

/* miners */

func (c *Client) GetMinerID(miner string) (uint64, error) {
	return c.baseGetUint64(c.getMinersKey(miner))
}

func (c *Client) GetMinerIPAddresses(chain string) (map[string]float64, error) {
	return c.baseZRangeWithScores(c.getMinerIPAddressesKey(chain))
}

func (c *Client) GetMinerDifficulties(chain string) (map[string]float64, error) {
	return c.baseZRangeWithScores(c.getMinerDifficultiesKey(chain))
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
	raw, err := c.baseZRange(c.getTopMinersKey(chain), 250, true)
	if err != nil {
		return nil, err
	}

	values := make([]uint64, len(raw))
	for i, id := range raw {
		values[i], err = strconv.ParseUint(id, 10, 64)
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

func (c *Client) GetRoundSoloShares(chain string, minerID uint64) (uint64, error) {
	return c.baseGetUint64(c.getRoundSoloAcceptedSharesKey(chain, minerID))
}

func (c *Client) GetRoundShareCounts(chain string, soloMinerID uint64) (
	uint64, uint64, uint64,
	error,
) {
	var acceptedKey, rejectedKey, invalidKey string
	if soloMinerID == 0 {
		acceptedKey = c.getRoundAcceptedSharesKey(chain)
		rejectedKey = c.getRoundRejectedSharesKey(chain)
		invalidKey = c.getRoundInvalidSharesKey(chain)
	} else {
		acceptedKey = c.getRoundSoloAcceptedSharesKey(chain, soloMinerID)
		rejectedKey = c.getRoundSoloRejectedSharesKey(chain, soloMinerID)
		invalidKey = c.getRoundSoloInvalidSharesKey(chain, soloMinerID)
	}

	ctx := context.Background()
	pipe := c.writeClient.Pipeline()

	acceptedRaw := pipe.Get(ctx, acceptedKey)
	rejectedRaw := pipe.Get(ctx, rejectedKey)
	invalidRaw := pipe.Get(ctx, invalidKey)

	pipe.Set(ctx, acceptedKey, "0", 0)
	pipe.Set(ctx, rejectedKey, "0", 0)
	pipe.Set(ctx, invalidKey, "0", 0)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	acceptedStr, err := acceptedRaw.Result()
	if err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	accepted, err := strconv.ParseUint(acceptedStr, 10, 64)
	if err != nil && acceptedStr != "" {
		return 0, 0, 0, err
	}

	rejectedStr, err := rejectedRaw.Result()
	if err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	rejected, err := strconv.ParseUint(rejectedStr, 10, 64)
	if err != nil && rejectedStr != "" {
		return 0, 0, 0, err
	}

	invalidStr, err := invalidRaw.Result()
	if err != nil && err != redis.Nil {
		return 0, 0, 0, err
	}

	invalid, err := strconv.ParseUint(invalidStr, 10, 64)
	if err != nil && invalidStr != "" {
		return 0, 0, 0, err
	}

	return accepted, rejected, invalid, nil
}

/* intervals */

func (c *Client) GetIntervals(chain string) ([]string, error) {
	return c.baseSMembers(c.getIntervalsKey(chain))
}

func (c *Client) GetIntervalAcceptedShares(chain, interval string) (
	map[string]uint64,
	map[string]uint64,
	error,
) {
	key := c.getIntervalAcceptedSharesKey(chain, interval)
	raw, err := c.baseZRangeWithScoresUint64(key)
	if err != nil {
		return nil, nil, err
	}

	key = c.getIntervalAcceptedAdjustedSharesKey(chain, interval)
	adjusted, err := c.baseZRangeWithScoresUint64(key)
	if err != nil {
		return nil, nil, err
	}

	return raw, adjusted, nil
}

func (c *Client) GetIntervalRejectedShares(chain, interval string) (
	map[string]uint64,
	map[string]uint64,
	error,
) {
	key := c.getIntervalRejectedSharesKey(chain, interval)
	raw, err := c.baseZRangeWithScoresUint64(key)
	if err != nil {
		return nil, nil, err
	}

	key = c.getIntervalRejectedAdjustedSharesKey(chain, interval)
	adjusted, err := c.baseZRangeWithScoresUint64(key)
	if err != nil {
		return nil, nil, err
	}

	return raw, adjusted, nil
}

func (c *Client) GetIntervalInvalidShares(chain, interval string) (
	map[string]uint64,
	map[string]uint64,
	error,
) {
	key := c.getIntervalInvalidSharesKey(chain, interval)
	raw, err := c.baseZRangeWithScoresUint64(key)
	if err != nil {
		return nil, nil, err
	}

	key = c.getIntervalInvalidAdjustedSharesKey(chain, interval)
	adjusted, err := c.baseZRangeWithScoresUint64(key)
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

func (c *Client) GetChartEarningsLastTime(chain string) (time.Time, error) {
	return c.baseGetTime(c.getChartEarningsLastTimeKey(chain))
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

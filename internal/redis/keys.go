package redis

import (
	"strconv"
	"strings"
)

func (c *Client) getKey(args ...string) string {
	key := c.env
	for _, arg := range args {
		key += ":" + arg
	}

	return key
}

/* channels */

func (c *Client) getStreamIndexChannelKey() string {
	return c.getKey("strm", "idx")
}

func (c *Client) getStreamMinerChannelKey(minerID uint64) string {
	return c.getKey("strm", "mnr", strconv.FormatUint(minerID, 10))
}

/* miners/workers */

func (c *Client) getMinersKey(minerID string) string {
	return c.getKey("pool", "mnrs", "ids", minerID)
}

func (c *Client) getWorkersKey(minerID uint64, worker string) string {
	compoundID := strconv.FormatUint(minerID, 10) + ":" + worker
	return c.getKey("pool", "wrkrs", "ids", compoundID)
}

func (c *Client) getMinerIPAddressesKey(chain string) string {
	return c.getKey("pool", "mnrs", strings.ToLower(chain), "ips")
}

func (c *Client) getMinerIPAddressesInactiveKey(chain string) string {
	return c.getKey("pool", "mnrs", strings.ToLower(chain), "ips", "inctv")
}

func (c *Client) getMinerLatenciesKey(chain string) string {
	return c.getKey("pool", "mnrs", strings.ToLower(chain), "ltncs")
}

func (c *Client) getTopMinersKey(chain string) string {
	return c.getKey("pool", "mnrs", strings.ToLower(chain), "top")
}

/* share index */

func (c *Client) getShareIndexKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "ush")
}

func (c *Client) getUniqueSharesKey(chain string, height uint64) string {
	return c.getKey("pool", strings.ToLower(chain), "ush", strconv.FormatUint(height, 10))
}

/* rounds */

func (c *Client) getRoundSharesKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "shr")
}

func (c *Client) getRoundAcceptedSharesKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "ash")
}

func (c *Client) getRoundSoloAcceptedSharesKey(chain string, minerID uint64) string {
	return c.getKey("pool", strings.ToLower(chain), "ash", "solo", strconv.FormatUint(minerID, 10))
}

func (c *Client) getRoundRejectedSharesKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "rsh")
}

func (c *Client) getRoundSoloRejectedSharesKey(chain string, minerID uint64) string {
	return c.getKey("pool", strings.ToLower(chain), "rsh", "solo", strconv.FormatUint(minerID, 10))
}

func (c *Client) getRoundInvalidSharesKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "ish")
}

func (c *Client) getRoundSoloInvalidSharesKey(chain string, minerID uint64) string {
	return c.getKey("pool", strings.ToLower(chain), "ish", "solo", strconv.FormatUint(minerID, 10))
}

/* intervals */

func (c *Client) getIntervalsKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "intrvls")
}

func (c *Client) getIntervalAcceptedSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "ash", interval)
}

func (c *Client) getIntervalAcceptedSoloSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "ash", interval, "solo")
}

func (c *Client) getIntervalAcceptedAdjustedSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "ashadj", interval)
}

func (c *Client) getIntervalAcceptedSoloAdjustedSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "ashadj", interval, "solo")
}

func (c *Client) getIntervalRejectedSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "rsh", interval)
}

func (c *Client) getIntervalRejectedSoloSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "rsh", interval, "solo")
}

func (c *Client) getIntervalRejectedAdjustedSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "rshadj", interval)
}

func (c *Client) getIntervalRejectedSoloAdjustedSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "rshadj", interval, "solo")
}

func (c *Client) getIntervalInvalidSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "ish", interval)
}

func (c *Client) getIntervalInvalidSoloSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "ish", interval, "solo")
}

func (c *Client) getIntervalInvalidAdjustedSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "ishadj", interval)
}

func (c *Client) getIntervalInvalidSoloAdjustedSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "ishadj", interval, "solo")
}

/* chart */

func (c *Client) getChartSharesLastTimeKey(chain string) string {
	return c.getKey("chrt", "sh", strings.ToLower(chain))
}

func (c *Client) getChartBlocksLastTimeKey(chain string) string {
	return c.getKey("chrt", "blk", strings.ToLower(chain))
}

func (c *Client) getChartRoundsLastTimeKey(chain string) string {
	return c.getKey("chrt", "rnd", strings.ToLower(chain))
}

/* cached stats */

func (c *Client) getCachedGlobalLastSharesKey() string {
	return c.getKey("cache", "shrs")
}

func (c *Client) getCachedGlobalLastProfitsKey() string {
	return c.getKey("cache", "prfts")
}

func (c *Client) getCachedLuckByChainKey(chain string) string {
	return c.getKey("cache", "luck", strings.ToLower(chain))
}

func (c *Client) getCachedMinersByChainKey(chain string) string {
	return c.getKey("cache", "mnrs", strings.ToLower(chain))
}

func (c *Client) getCachedWorkersByChainKey(chain string) string {
	return c.getKey("cache", "wrkrs", strings.ToLower(chain))
}

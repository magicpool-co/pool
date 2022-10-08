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

func (c *Client) getTopMinersKey(chain string) string {
	return c.getKey("pool", "mnrs", strings.ToLower(chain), "top")
}

/* rounds */

func (c *Client) getRoundSharesKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "shr")
}

func (c *Client) getRoundAcceptedSharesKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "ash")
}

func (c *Client) getRoundRejectedSharesKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "rsh")
}

func (c *Client) getRoundInvalidSharesKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "ish")
}

/* intervals */

func (c *Client) getIntervalsKey(chain string) string {
	return c.getKey("pool", strings.ToLower(chain), "intrvls")
}

func (c *Client) getIntervalAcceptedSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "ash", interval)
}

func (c *Client) getIntervalRejectedSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "rsh", interval)
}

func (c *Client) getIntervalInvalidSharesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "ish", interval)
}

func (c *Client) getIntervalReportedHashratesKey(chain, interval string) string {
	return c.getKey("pool", strings.ToLower(chain), "rhr", interval)
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

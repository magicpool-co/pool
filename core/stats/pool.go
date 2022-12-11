package stats

import (
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/types"
)

var (
	names = map[string]string{
		"CFX":  "Conflux",
		"CTXC": "Cortex",
		"ERGO": "ERGO",
		"ETC":  "Ethereum Classic",
		"ETHW": "EthereumPow",
		"FLUX": "Flux",
		"FIRO": "Firo",
		"KAS":  "Kaspa",
		"RVN":  "Ravencoin",
	}
)

func (c *Client) GetPoolStats(nodes []types.MiningNode) ([]*PoolStats, error) {
	dbShares, err := tsdb.GetGlobalSharesLast(c.tsdb.Reader(), int(types.Period15m))
	if err != nil {
		return nil, err
	}

	dbSharesIdx := make(map[string]*tsdb.Share)
	for _, dbShare := range dbShares {
		dbSharesIdx[dbShare.ChainID] = dbShare
	}

	stats := make([]*PoolStats, len(nodes))
	for i, node := range nodes {
		chain := node.Chain()
		miners, err := pooldb.GetActiveMinersCount(c.pooldb.Reader(), chain)
		if err != nil {
			return nil, err
		}

		workers, err := pooldb.GetActiveWorkersCount(c.pooldb.Reader(), chain)
		if err != nil {
			return nil, err
		}

		var hashrate float64
		if dbShare, ok := dbSharesIdx[chain]; ok {
			hashrate = dbShare.Hashrate
		}

		luck, err := pooldb.GetRoundLuckByChain(c.pooldb.Reader(), chain, time.Hour*24*30)
		if err != nil {
			return nil, err
		}

		luck /= float64(node.GetShareDifficulty().Value())
		luck *= 100

		stats[i] = &PoolStats{
			Name:     names[chain],
			Symbol:   chain,
			Miners:   miners,
			Workers:  workers,
			Hashrate: newNumberFromFloat64(hashrate, "H/s", true),
			Luck:     newNumberFromFloat64(luck, "%", false),
		}
	}

	return stats, nil
}

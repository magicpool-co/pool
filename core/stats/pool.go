package stats

import (
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) GetPoolSummary(nodes []types.MiningNode) ([]*PoolSummary, error) {
	dbShares, err := tsdb.GetGlobalSharesLast(c.tsdb.Reader(), int(types.Period15m))
	if err != nil {
		return nil, err
	}

	dbSharesIdx := make(map[string]*tsdb.Share)
	for _, dbShare := range dbShares {
		dbSharesIdx[dbShare.ChainID] = dbShare
	}

	dbBlocks, err := tsdb.GetBlocksProfitabilityLast(c.tsdb.Reader(), int(types.Period15m))
	if err != nil {
		return nil, err
	}

	dbBlocksIdx := make(map[string]*tsdb.Block)
	for _, dbBlock := range dbBlocks {
		dbBlocksIdx[dbBlock.ChainID] = dbBlock
	}

	stats := make([]*PoolSummary, len(nodes))
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

		var profitUsd, profitBtc float64
		if dbBlock, ok := dbBlocksIdx[chain]; ok {
			profitUsd, profitBtc = dbBlock.AvgProfitability, dbBlock.AvgProfitabilityBTC
		}

		luck, err := pooldb.GetRoundLuckByChain(c.pooldb.Reader(), chain, time.Hour*24*30)
		if err != nil {
			return nil, err
		}

		luck /= float64(node.GetShareDifficulty().Value())
		luck *= 100

		stats[i] = &PoolSummary{
			Name:      node.Name(),
			Symbol:    chain,
			Miners:    miners,
			Workers:   workers,
			Hashrate:  newNumberFromFloat64(hashrate, "H/s", true),
			Luck:      newNumberFromFloat64(luck, "%", false),
			ProfitUSD: newNumberFromFloat64WithPrecision(profitUsd, 32, " $/H/s", false),
			ProfitBTC: newNumberFromFloat64WithPrecision(profitBtc, 32, " BTC/H/s", false),
		}
	}

	return stats, nil
}

package stats

import (
	"time"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) GetPoolSummary(nodes []types.MiningNode) ([]*PoolSummary, error) {
	dbShares, err := c.getGlobalSharesLast()
	if err != nil {
		return nil, err
	}

	dbSharesIdx := make(map[string]*tsdb.Share)
	for _, dbShare := range dbShares {
		dbSharesIdx[dbShare.ChainID] = dbShare
	}

	dbBlocks, err := c.getBlocksWithProfitabilityLast()
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
		miners, err := c.getActiveMinersCount(chain)
		if err != nil {
			return nil, err
		}

		workers, err := c.getActiveWorkersCount(chain)
		if err != nil {
			return nil, err
		}

		var hashrate float64
		if dbShare, ok := dbSharesIdx[chain]; ok {
			hashrate = dbShare.Hashrate
		}

		var networkDifficulty, networkHashrate, blockReward, blockTime, profitUsd, profitBtc float64
		var ttf time.Duration
		if dbBlock, ok := dbBlocksIdx[chain]; ok {
			networkDifficulty, networkHashrate = dbBlock.Difficulty, dbBlock.Hashrate
			blockReward, blockTime = dbBlock.Value, dbBlock.BlockTime
			profitUsd, profitBtc = dbBlock.AvgProfitabilityUSD, dbBlock.AvgProfitabilityBTC

			if blockTime > 0 && hashrate > 0 && networkHashrate > 0 {
				ttf = time.Duration(blockTime * (networkHashrate / hashrate) * float64(time.Second))
			}
		}

		luck, err := c.getRoundLuckByChain(chain)
		if err != nil {
			return nil, err
		}

		if chain == "NEXA" {
			luck /= 0.2
		} else {
			luck /= float64(node.GetShareDifficulty(1).Value())
		}

		luck *= 100

		stats[i] = &PoolSummary{
			Name:               node.Name(),
			Symbol:             chain,
			Fee:                newNumberFromFloat64WithPrecision(0.01, 2, "%", false),
			Miners:             miners,
			Workers:            workers,
			Hashrate:           newNumberFromFloat64(hashrate, "H/s", true),
			Luck:               newNumberFromFloat64(luck, "%", false),
			TTF:                newNumberFromDuration(ttf),
			ProfitUSD:          newNumberFromFloat64WithPrecision(profitUsd, 32, " $/H/s", false),
			ProfitBTC:          newNumberFromFloat64WithPrecision(profitBtc, 32, " BTC/H/s", false),
			NetworkDifficulty:  newNumberFromFloat64(networkDifficulty, "", true),
			NetworkHashrate:    newNumberFromFloat64(networkHashrate, "H/s", true),
			NetworkBlockReward: newNumberFromFloat64(blockReward, " "+chain, false),
		}
	}

	return stats, nil
}

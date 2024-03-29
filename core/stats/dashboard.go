package stats

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

var (
	dashboardAggPeriod   = int(types.Period15m)
	dashboardAggDuration = time.Hour * 24
)

func (c *Client) processHashrateInfo(shares []*tsdb.Share) map[string]*HashrateInfo {
	idx := make(map[string]*HashrateInfo)
	for _, share := range shares {
		var ok bool
		share.ChainID, ok = c.processChainID(share.ChainID)
		if !ok {
			continue
		}

		var units string
		switch share.ChainID {
		case "FLUX":
			units = "Sol/s"
		default:
			units = "H/s"
		}

		idx[share.ChainID] = &HashrateInfo{
			Hashrate:    newNumberFromFloat64(share.Hashrate, units, true),
			AvgHashrate: newNumberFromFloat64(share.AvgHashrate, units, true),
		}
	}

	return idx
}

func (c *Client) processShareInfo(shares []*tsdb.Share) map[string]*ShareInfo {
	idx := make(map[string]*ShareInfo)
	for _, share := range shares {
		var ok bool
		share.ChainID, ok = c.processChainID(share.ChainID)
		if !ok {
			continue
		}

		var acceptedRate, rejectedRate, invalidRate float64
		sumShares := share.AcceptedAdjustedShares + share.RejectedAdjustedShares + share.InvalidAdjustedShares
		sumSharesFloat := float64(sumShares)
		if sumSharesFloat > 0 {
			acceptedRate = 100 * float64(share.AcceptedAdjustedShares) / sumSharesFloat
			rejectedRate = 100 * float64(share.RejectedAdjustedShares) / sumSharesFloat
			invalidRate = 100 * float64(share.InvalidAdjustedShares) / sumSharesFloat
		}

		idx[share.ChainID] = &ShareInfo{
			AcceptedShares:    newNumberFromFloat64(float64(share.AcceptedAdjustedShares), "", false),
			AcceptedShareRate: newNumberFromFloat64(acceptedRate, "%", false),
			RejectedShares:    newNumberFromFloat64(float64(share.RejectedAdjustedShares), "", false),
			RejectedShareRate: newNumberFromFloat64(rejectedRate, "%", false),
			InvalidShares:     newNumberFromFloat64(float64(share.InvalidAdjustedShares), "", false),
			InvalidShareRate:  newNumberFromFloat64(invalidRate, "%", false),
		}
	}

	return idx
}

func (c *Client) GetMinerDashboard(minerIdx map[uint64]string) (*Dashboard, error) {
	minerIDs := make([]uint64, 0)
	for minerID := range minerIdx {
		minerIDs = append(minerIDs, minerID)
	}

	// fetch last shares
	lastShares, err := tsdb.GetMinersSharesLast(c.tsdb.Reader(), minerIDs, dashboardAggPeriod)
	if err != nil {
		return nil, err
	}

	// fetch sum shares
	sumShares, err := tsdb.GetMinersSharesSum(c.tsdb.Reader(), minerIDs, dashboardAggPeriod, dashboardAggDuration)
	if err != nil {
		return nil, err
	}

	// fetch last profitabilities
	lastProfits, err := c.getBlocksWithProfitabilityLast()
	if err != nil {
		return nil, err
	}

	// fetch sum active/inactive workers
	activeWorkers, inactiveWorkers, err := pooldb.GetWorkerCountByMinerIDs(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return nil, err
	}

	// fetch balance sums
	balanceSums, err := pooldb.GetBalanceSumsByMinerIDs(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return nil, err
	}

	// calculate immature, pending, and unpaid balance sums
	rawImmatureBalances := make(map[string]*big.Int)
	rawPendingBalances := make(map[string]*big.Int)
	rawUnpaidBalances := make(map[string]*big.Int)
	for _, balanceSum := range balanceSums {
		chain := balanceSum.ChainID
		minerChain, ok := minerIdx[balanceSum.MinerID]
		if !ok {
			return nil, fmt.Errorf("miner chain not found for id %d", balanceSum.MinerID)
		}

		// // process balance sum immature balance
		immature := balanceSum.ImmatureValue
		if immature.Valid && immature.BigInt.Cmp(common.Big0) > 0 {
			if _, ok := rawImmatureBalances[chain]; !ok {
				rawImmatureBalances[chain] = new(big.Int)
			}

			rawImmatureBalances[chain].Add(rawImmatureBalances[chain], immature.BigInt)
		}

		// process balance sum mature balance
		mature := balanceSum.MatureValue
		if mature.Valid && mature.BigInt.Cmp(common.Big0) > 0 {
			if chain == minerChain {
				if _, ok := rawUnpaidBalances[chain]; !ok {
					rawUnpaidBalances[chain] = new(big.Int)
				}

				rawUnpaidBalances[chain].Add(rawUnpaidBalances[chain], mature.BigInt)
			} else {
				if _, ok := rawPendingBalances[chain]; !ok {
					rawPendingBalances[chain] = new(big.Int)
				}

				rawPendingBalances[chain].Add(rawPendingBalances[chain], mature.BigInt)
			}
		}
	}

	// convert raw balances to processed balances
	immatureBalances := make(map[string]Number, len(rawImmatureBalances))
	for chain, balance := range rawImmatureBalances {
		immatureBalances[chain], err = newNumberFromBigInt(balance, chain)
		if err != nil {
			return nil, err
		}
	}

	pendingBalances := make(map[string]Number, len(rawPendingBalances))
	for chain, balance := range rawPendingBalances {
		pendingBalances[chain], err = newNumberFromBigInt(balance, chain)
		if err != nil {
			return nil, err
		}
	}

	unpaidBalances := make(map[string]Number, len(rawUnpaidBalances))
	for chain, balance := range rawUnpaidBalances {
		unpaidBalances[chain], err = newNumberFromBigInt(balance, chain)
		if err != nil {
			return nil, err
		}
	}

	// calculate hashrate and share info
	hashrateInfo := c.processHashrateInfo(lastShares)
	shareInfo := c.processShareInfo(sumShares)

	// convert profitabilities into index
	profitIndex := make(map[string]*tsdb.Block)
	for _, block := range lastProfits {
		profitIndex[block.ChainID] = block
	}

	// calculate projected earnings
	projectedEarningsNative := make(map[string]float64)
	projectedEarningsUSD := make(map[string]float64)
	projectedEarningsBTC := make(map[string]float64)
	projectedEarningsETH := make(map[string]float64)
	for chain, hashrateValue := range hashrateInfo {
		chain = strings.Split(chain, " ")[0]
		block, ok := profitIndex[chain]
		if ok {
			hashrate := hashrateValue.AvgHashrate.Value
			projectedEarningsNative[chain] += hashrate * block.AvgProfitability
			projectedEarningsUSD[chain] += hashrate * block.AvgProfitabilityUSD
			projectedEarningsBTC[chain] += hashrate * block.AvgProfitabilityBTC
			projectedEarningsETH[chain] += hashrate * block.AvgProfitabilityETH
		}
	}

	dashboard := &Dashboard{
		ActiveWorkers:           newNumberFromUint64Ptr(activeWorkers),
		InactiveWorkers:         newNumberFromUint64Ptr(inactiveWorkers),
		HashrateInfo:            hashrateInfo,
		ShareInfo:               shareInfo,
		ImmatureBalance:         immatureBalances,
		PendingBalance:          pendingBalances,
		UnpaidBalance:           unpaidBalances,
		ProjectedEarningsNative: projectedEarningsNative,
		ProjectedEarningsUSD:    projectedEarningsUSD,
		ProjectedEarningsBTC:    projectedEarningsBTC,
		ProjectedEarningsETH:    projectedEarningsETH,
	}

	return dashboard, nil
}

func (c *Client) GetWorkerDashboard(workerID uint64) (*Dashboard, error) {
	sumShares, err := tsdb.GetWorkerSharesSum(c.tsdb.Reader(),
		[]uint64{workerID}, dashboardAggPeriod, dashboardAggDuration)
	if err != nil {
		return nil, err
	}

	lastShares, err := tsdb.GetWorkerSharesLast(c.tsdb.Reader(), workerID, dashboardAggPeriod)
	if err != nil {
		return nil, err
	}

	dashboard := &Dashboard{
		HashrateInfo: c.processHashrateInfo(lastShares),
		ShareInfo:    c.processShareInfo(sumShares),
	}

	return dashboard, nil
}

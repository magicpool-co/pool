package stats

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	// "github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

var (
	dashboardAggPeriod   = int(types.Period15m)
	dashboardAggDuration = time.Hour * 24
)

func processHashrateInfo(shares []*tsdb.Share) map[string]*HashrateInfo {
	idx := make(map[string]*HashrateInfo)
	for _, share := range shares {
		var units string
		switch share.ChainID {
		case "FLUX":
			units = "S/s"
		case "AE", "CTXC":
			units = "Gps"
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

func processShareInfo(shares []*tsdb.Share) map[string]*ShareInfo {
	idx := make(map[string]*ShareInfo)
	for _, share := range shares {
		var acceptedRate, rejectedRate, invalidRate float64
		sumShares := float64(share.AcceptedShares + share.RejectedShares + share.InvalidShares)
		if sumShares > 0 {
			acceptedRate = 100 * float64(share.AcceptedShares) / sumShares
			rejectedRate = 100 * float64(share.RejectedShares) / sumShares
			invalidRate = 100 * float64(share.InvalidShares) / sumShares
		}

		idx[share.ChainID] = &ShareInfo{
			AcceptedShares:    newNumberFromFloat64(float64(share.AcceptedShares), "", false),
			AcceptedShareRate: newNumberFromFloat64(acceptedRate, "%", false),
			RejectedShares:    newNumberFromFloat64(float64(share.RejectedShares), "", false),
			RejectedShareRate: newNumberFromFloat64(rejectedRate, "%", false),
			InvalidShares:     newNumberFromFloat64(float64(share.InvalidShares), "", false),
			InvalidShareRate:  newNumberFromFloat64(invalidRate, "%", false),
		}
	}

	return idx
}

func (c *Client) GetGlobalDashboard() (*Dashboard, error) {
	sumShares, err := tsdb.GetGlobalSharesSum(c.tsdb.Reader(), dashboardAggPeriod, dashboardAggDuration)
	if err != nil {
		return nil, err
	}

	lastShares, err := tsdb.GetGlobalSharesLast(c.tsdb.Reader(), dashboardAggPeriod)
	if err != nil {
		return nil, err
	}

	activeMiners, err := pooldb.GetActiveMinersCount(c.pooldb.Reader(), "")
	if err != nil {
		return nil, err
	}

	activeWorkers, err := pooldb.GetActiveWorkersCount(c.pooldb.Reader(), "")
	if err != nil {
		return nil, err
	}

	dashboard := &Dashboard{
		Miners:        newNumberFromUint64Ptr(activeMiners),
		ActiveWorkers: newNumberFromUint64Ptr(activeWorkers),
		HashrateInfo:  processHashrateInfo(lastShares),
		ShareInfo:     processShareInfo(sumShares),
	}

	return dashboard, nil
}

func (c *Client) GetMinerDashboard(minerIDs []uint64, chains []string) (*Dashboard, error) {
	if len(minerIDs) != len(chains) {
		return nil, fmt.Errorf("minerIDs and chains count mismatch")
	}

	minerChainIdx := make(map[uint64]string)
	for i, minerID := range minerIDs {
		minerChainIdx[minerID] = strings.ToUpper(chains[i])
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

	// fetch sum active workers
	activeWorkers, err := pooldb.GetActiveWorkersByMinersCount(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return nil, err
	}

	// fetch sum inactive workers
	inactiveWorkers, err := pooldb.GetInactiveWorkersByMinersCount(c.pooldb.Reader(), minerIDs)
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
		minerChain, ok := minerChainIdx[balanceSum.MinerID]
		if !ok || minerChain == "" {
			continue
		}

		// // process balance sum immature balance
		// immature := balanceSum.ImmatureValue
		// if immature.Valid && immature.BigInt.Cmp(common.Big0) > 0 {
		// 	if _, ok := rawImmatureBalances[minerChain]; !ok {
		// 		rawImmatureBalances[minerChain] = new(big.Int)
		// 	}

		// 	rawImmatureBalances[minerChain].Add(rawImmatureBalances[minerChain], immature.BigInt)
		// }

		// // process balance sum mature balance
		// mature := balanceSum.MatureValue
		// if mature.Valid && mature.BigInt.Cmp(common.Big0) > 0 {
		// 	if balanceSum.ChainID == minerChain {
		// 		if _, ok := rawUnpaidBalances[minerChain]; !ok {
		// 			rawUnpaidBalances[minerChain] = new(big.Int)
		// 		}

		// 		rawUnpaidBalances[minerChain].Add(rawUnpaidBalances[minerChain], mature.BigInt)
		// 	} else {
		// 		if _, ok := rawPendingBalances[minerChain]; !ok {
		// 			rawPendingBalances[minerChain] = new(big.Int)
		// 		}

		// 		rawPendingBalances[minerChain].Add(rawPendingBalances[minerChain], mature.BigInt)
		// 	}
		// }
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
	hashrateInfo := processHashrateInfo(lastShares)
	shareInfo := processShareInfo(sumShares)

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
		block, ok := profitIndex[chain]
		if ok {
			hashrate := hashrateValue.AvgHashrate.Value
			projectedEarningsNative[chain] = hashrate * block.AvgProfitability
			projectedEarningsUSD[chain] = hashrate * block.AvgProfitabilityUSD
			projectedEarningsBTC[chain] = hashrate * block.AvgProfitabilityBTC
			projectedEarningsETH[chain] = hashrate * block.AvgProfitabilityETH
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
	sumShares, err := tsdb.GetWorkerSharesSum(c.tsdb.Reader(), []uint64{workerID}, dashboardAggPeriod, dashboardAggDuration)
	if err != nil {
		return nil, err
	}

	lastShares, err := tsdb.GetWorkerSharesLast(c.tsdb.Reader(), workerID, dashboardAggPeriod)
	if err != nil {
		return nil, err
	}

	dashboard := &Dashboard{
		HashrateInfo: processHashrateInfo(lastShares),
		ShareInfo:    processShareInfo(sumShares),
	}

	return dashboard, nil
}

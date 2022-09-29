package stats

import (
	"math/big"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
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
			Hashrate:         newNumberFromFloat64(share.Hashrate, units, true),
			AvgHashrate:      newNumberFromFloat64(share.AvgHashrate, units, true),
			ReportedHashrate: newNumberFromFloat64(share.ReportedHashrate, units, true),
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

	activeMiners, err := pooldb.GetActiveMinersCount(c.pooldb.Reader())
	if err != nil {
		return nil, err
	}

	activeWorkers, err := pooldb.GetActiveWorkersCount(c.pooldb.Reader())
	if err != nil {
		return nil, err
	}

	dashboard := &Dashboard{
		Miners:        newNumberFromUint64Ptr(activeMiners),
		ActiveWorkers: newNumberFromUint64Ptr(activeWorkers),
		HashrateInfo:  processHashrateInfo(lastShares),
		SharesInfo:    processShareInfo(sumShares),
	}

	return dashboard, nil
}

func (c *Client) GetMinerDashboard(minerIDs []uint64) (*Dashboard, error) {
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

	// fetch pending balances for all minerIDs
	pendingBalances, err := pooldb.GetPendingBalanceInputSumsByMiners(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return nil, err
	}

	// sum pending balances by chain
	pendingBalanceBig := make(map[string]*big.Int)
	for _, balance := range pendingBalances {
		if !balance.Value.Valid {
			continue
		} else if _, ok := pendingBalanceBig[balance.ChainID]; !ok {
			pendingBalanceBig[balance.ChainID] = new(big.Int)
		}
		pendingBalanceBig[balance.ChainID].Add(pendingBalanceBig[balance.ChainID], balance.Value.BigInt)
	}

	// convert pending balances to number type
	pendingBalance := make(map[string]Number)
	for chain, balance := range pendingBalanceBig {
		var err error
		pendingBalance[chain], err = newNumberFromBigInt(balance, chain)
		if err != nil {
			return nil, err
		}
	}

	// fetch unpaid balances for all minerIDs
	unpaidBalances, err := pooldb.GetUnpaidBalanceOutputSumsByMiners(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return nil, err
	}

	// sum unpaid balances by chain
	unpaidBalanceBig := make(map[string]*big.Int)
	for _, balance := range unpaidBalances {
		if !balance.Value.Valid {
			continue
		} else if _, ok := unpaidBalanceBig[balance.ChainID]; !ok {
			unpaidBalanceBig[balance.ChainID] = new(big.Int)
		}
		unpaidBalanceBig[balance.ChainID].Add(unpaidBalanceBig[balance.ChainID], balance.Value.BigInt)
	}

	// convert unpaid balances to number type
	unpaidBalance := make(map[string]Number)
	for chain, balance := range unpaidBalanceBig {
		var err error
		unpaidBalance[chain], err = newNumberFromBigInt(balance, chain)
		if err != nil {
			return nil, err
		}
	}

	dashboard := &Dashboard{
		ActiveWorkers:   newNumberFromUint64Ptr(activeWorkers),
		InactiveWorkers: newNumberFromUint64Ptr(inactiveWorkers),
		HashrateInfo:    processHashrateInfo(lastShares),
		SharesInfo:      processShareInfo(sumShares),
		PendingBalance:  pendingBalance,
		UnpaidBalance:   unpaidBalance,
	}

	return dashboard, nil
}

func (c *Client) GetWorkerDashboard(workerID uint64) (*Dashboard, error) {
	sumShares, err := tsdb.GetWorkerSharesSum(c.tsdb.Reader(), workerID, dashboardAggPeriod, dashboardAggDuration)
	if err != nil {
		return nil, err
	}

	lastShares, err := tsdb.GetWorkerSharesLast(c.tsdb.Reader(), workerID, dashboardAggPeriod)
	if err != nil {
		return nil, err
	}

	dashboard := &Dashboard{
		HashrateInfo: processHashrateInfo(lastShares),
		SharesInfo:   processShareInfo(sumShares),
	}

	return dashboard, nil
}

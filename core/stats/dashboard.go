package stats

import (
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
		Hashrate:      processHashrateInfo(lastShares),
		Shares:        processShareInfo(sumShares),
	}

	return dashboard, nil
}

func (c *Client) GetMinerDashboard(minerIDs []uint64) (*Dashboard, error) {
	totalSumShares, totalLastShares := make([]*tsdb.Share, 0), make([]*tsdb.Share, 0)
	var totalActiveWorkers, totalInactiveWorkers uint64
	for _, minerID := range minerIDs {
		sumShares, err := tsdb.GetMinerSharesSum(c.tsdb.Reader(), minerID, dashboardAggPeriod, dashboardAggDuration)
		if err != nil {
			return nil, err
		}
		totalSumShares = append(totalSumShares, sumShares...)

		lastShares, err := tsdb.GetMinerSharesLast(c.tsdb.Reader(), minerID, dashboardAggPeriod)
		if err != nil {
			return nil, err
		}
		totalLastShares = append(totalLastShares, lastShares...)

		activeWorkers, err := pooldb.GetActiveWorkersByMinerCount(c.pooldb.Reader(), minerID)
		if err != nil {
			return nil, err
		}
		totalActiveWorkers += activeWorkers

		inactiveWorkers, err := pooldb.GetInactiveWorkersByMinerCount(c.pooldb.Reader(), minerID)
		if err != nil {
			return nil, err
		}
		totalInactiveWorkers += inactiveWorkers
	}

	dashboard := &Dashboard{
		ActiveWorkers:   newNumberFromUint64Ptr(totalActiveWorkers),
		InactiveWorkers: newNumberFromUint64Ptr(totalInactiveWorkers),
		Hashrate:        processHashrateInfo(totalLastShares),
		Shares:          processShareInfo(totalSumShares),
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
		Hashrate: processHashrateInfo(lastShares),
		Shares:   processShareInfo(sumShares),
	}

	return dashboard, nil
}

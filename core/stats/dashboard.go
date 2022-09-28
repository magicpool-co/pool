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
		idx[share.ChainID] = &HashrateInfo{
			Hashrate:         newNumberFromFloat64(share.Hashrate, "H/s", true),
			AvgHashrate:      newNumberFromFloat64(share.AvgHashrate, "H/s", true),
			ReportedHashrate: newNumberFromFloat64(share.ReportedHashrate, "H/s", true),
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

	activeMiners, err := pooldb.GetActiveMinerCount(c.pooldb.Reader())
	if err != nil {
		return nil, err
	}

	activeWorkers, err := pooldb.GetActiveWorkerCount(c.pooldb.Reader())
	if err != nil {
		return nil, err
	}

	dashboard := &Dashboard{
		MinersCount:  newNumberFromFloat64Ptr(float64(activeMiners), "", false),
		WorkersCount: newNumberFromFloat64Ptr(float64(activeWorkers), "", false),
		Hashrate:     processHashrateInfo(lastShares),
		Shares:       processShareInfo(sumShares),
	}

	return dashboard, nil
}

func (c *Client) GetMinerDashboard(minerID uint64) (*Dashboard, error) {
	sumShares, err := tsdb.GetMinerSharesSum(c.tsdb.Reader(), minerID, dashboardAggPeriod, dashboardAggDuration)
	if err != nil {
		return nil, err
	}

	lastShares, err := tsdb.GetMinerSharesLast(c.tsdb.Reader(), minerID, dashboardAggPeriod)
	if err != nil {
		return nil, err
	}

	dbWorkers, err := pooldb.GetWorkersByMiner(c.pooldb.Reader(), minerID)
	if err != nil {
		return nil, err
	}

	activeWorkers := make([]*Worker, 0)
	inactiveWorkers := make([]*Worker, 0)
	for _, dbWorker := range dbWorkers {
		worker := &Worker{
			Name:     dbWorker.Name,
			Active:   dbWorker.Active,
			LastSeen: dbWorker.LastShare.Unix(),
		}

		if worker.Active {
			activeWorkers = append(activeWorkers, worker)
		} else {
			inactiveWorkers = append(inactiveWorkers, worker)
		}
	}

	dashboard := &Dashboard{
		WorkersActive:   activeWorkers,
		WorkersInactive: inactiveWorkers,
		Hashrate:        processHashrateInfo(lastShares),
		Shares:          processShareInfo(sumShares),
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

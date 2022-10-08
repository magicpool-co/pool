package stats

import (
	"fmt"
	"sort"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/types"
)

func (c *Client) GetMiners(chain string, page, size uint64) ([]*Miner, uint64, error) {
	topMinerIDs, err := c.redis.GetTopMinerIDs(chain)
	if err != nil {
		return nil, 0, err
	}

	timestamp, err := c.redis.GetChartSharesLastTime(chain)
	if err != nil {
		return nil, 0, err
	} else if timestamp.IsZero() {
		return nil, 0, fmt.Errorf("no share timestamp found")
	}

	offset := int(page * size)
	limit := int(offset + int(size))
	count := uint64(len(topMinerIDs))
	if offset > len(topMinerIDs)-1 {
		return nil, count, nil
	} else if limit > len(topMinerIDs) {
		limit = len(topMinerIDs)
	}

	minerIDs := topMinerIDs[offset:limit]
	dbShares, err := tsdb.GetMinerSharesByEndTime(c.tsdb.Reader(),
		timestamp, minerIDs, chain, int(types.Period15m))
	if err != nil {
		return nil, 0, err
	}

	sort.Slice(dbShares, func(i, j int) bool {
		return dbShares[i].Hashrate > dbShares[j].Hashrate
	})

	dbMiners, err := pooldb.GetActiveMiners(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return nil, 0, err
	}

	dbMinerIdx := make(map[uint64]*pooldb.Miner)
	for _, dbMiner := range dbMiners {
		dbMinerIdx[dbMiner.ID] = dbMiner
	}

	miners := make([]*Miner, 0)
	for _, dbShare := range dbShares {
		minerID := types.Uint64Value(dbShare.MinerID)
		dbMiner, ok := dbMinerIdx[minerID]
		if !ok {
			continue
		}

		miner := &Miner{
			ID:           dbMiner.ID,
			Chain:        dbMiner.ChainID,
			Address:      dbMiner.Address,
			Active:       true,
			HashrateInfo: processHashrateInfo([]*tsdb.Share{dbShare}),
			FirstSeen:    dbMiner.CreatedAt.Unix(),
			LastSeen:     dbMiner.LastShare.Unix(),
		}
		miners = append(miners, miner)
	}

	return miners, count, nil
}

func (c *Client) GetWorkers(minerID, page, size uint64) (*WorkerList, error) {
	dbWorkers, err := pooldb.GetWorkersByMiner(c.pooldb.Reader(), minerID)
	if err != nil {
		return nil, err
	}

	workerIDs := make([]uint64, len(dbWorkers))
	for i, dbWorker := range dbWorkers {
		workerIDs[i] = dbWorker.ID
	}

	timestamp, err := c.redis.GetChartSharesLastTime("ETC")
	if err != nil {
		return nil, err
	} else if timestamp.IsZero() {
		return nil, fmt.Errorf("no share timestamp found")
	}

	dbShares, err := tsdb.GetWorkerSharesAllChainsByEndTime(c.tsdb.Reader(),
		timestamp, workerIDs, int(types.Period15m))
	if err != nil {
		return nil, err
	}

	dbSharesIdx := make(map[uint64][]*tsdb.Share)
	for _, dbShare := range dbShares {
		workerID := types.Uint64Value(dbShare.WorkerID)
		if _, ok := dbSharesIdx[workerID]; !ok {
			dbSharesIdx[workerID] = make([]*tsdb.Share, 0)
		}
		dbSharesIdx[workerID] = append(dbSharesIdx[workerID], dbShare)
	}

	activeWorkers := make([]*Worker, 0)
	inactiveWorkers := make([]*Worker, 0)
	for _, dbWorker := range dbWorkers {
		worker := &Worker{
			Name:         dbWorker.Name,
			Active:       dbWorker.Active,
			HashrateInfo: processHashrateInfo(dbSharesIdx[dbWorker.ID]),
			FirstSeen:    dbWorker.CreatedAt.Unix(),
			LastSeen:     dbWorker.LastShare.Unix(),
		}

		if worker.Active {
			activeWorkers = append(activeWorkers, worker)
		} else {
			inactiveWorkers = append(inactiveWorkers, worker)
		}
	}

	sort.Slice(activeWorkers, func(i, j int) bool {
		return activeWorkers[i].Name < activeWorkers[j].Name
	})

	sort.Slice(inactiveWorkers, func(i, j int) bool {
		return inactiveWorkers[i].Name < inactiveWorkers[j].Name
	})

	workerList := &WorkerList{
		Active:   activeWorkers,
		Inactive: inactiveWorkers,
	}

	return workerList, nil
}

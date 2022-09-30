package stats

import (
	"sort"

	"github.com/magicpool-co/pool/internal/pooldb"
)

func (c *Client) GetMiners(page, size uint64) ([]*Miner, uint64, error) {
	count, err := pooldb.GetActiveMinersCount(c.pooldb.Reader())
	if err != nil {
		return nil, 0, err
	}

	dbMiners, err := pooldb.GetActiveMiners(c.pooldb.Reader(), page, size)
	if err != nil {
		return nil, 0, err
	}

	minerIDs := make([]uint64, len(dbMiners))
	minerIdx := make(map[uint64]*Miner)
	for _, dbMiner := range dbMiners {
		minerIDs = append(minerIDs, dbMiner.ID)
		minerIdx[dbMiner.ID] = &Miner{
			ID:        dbMiner.ID,
			Chain:     dbMiner.ChainID,
			Address:   dbMiner.Address,
			Active:    dbMiner.Active,
			FirstSeen: dbMiner.CreatedAt.Unix(),
			LastSeen:  dbMiner.LastShare.Unix(),
		}
	}

	var i int
	miners := make([]*Miner, len(minerIdx))
	for _, miner := range minerIdx {
		miners[i] = miner
		i++
	}

	sort.Slice(miners, func(i, j int) bool {
		return miners[i].ID < miners[j].ID
	})

	return miners, count, nil
}

func (c *Client) GetWorkers(minerID, page, size uint64) (*WorkerList, error) {
	dbWorkers, err := pooldb.GetWorkersByMiner(c.pooldb.Reader(), minerID)
	if err != nil {
		return nil, err
	}

	activeWorkers := make([]*Worker, 0)
	inactiveWorkers := make([]*Worker, 0)
	for _, dbWorker := range dbWorkers {
		worker := &Worker{
			Name:      dbWorker.Name,
			Active:    dbWorker.Active,
			FirstSeen: dbWorker.CreatedAt.Unix(),
			LastSeen:  dbWorker.LastShare.Unix(),
		}

		if worker.Active {
			activeWorkers = append(activeWorkers, worker)
		} else {
			inactiveWorkers = append(inactiveWorkers, worker)
		}
	}

	workerList := &WorkerList{
		Active:   activeWorkers,
		Inactive: inactiveWorkers,
	}

	return workerList, nil
}

package tsdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/magicpool-co/pool/pkg/dbcl"
)

func GetBlocks(q dbcl.Querier, chain string, period int) ([]*Block, error) {
	const query = `SELECT *
	FROM blocks
	WHERE
		chain_id = ?
	AND
		period = ?
	AND
		pending = FALSE;`

	output := []*Block{}
	err := q.Select(&output, query, chain, period)

	return output, err
}

func GetRounds(q dbcl.Querier, chain string, period int) ([]*Round, error) {
	const query = `SELECT *
	FROM rounds
	WHERE
		chain_id = ?
	AND
		period = ?
	AND
		pending = FALSE;`

	output := []*Round{}
	err := q.Select(&output, query, chain, period)

	return output, err
}

func GetGlobalShares(q dbcl.Querier, chain string, period int) ([]*Share, error) {
	const query = `SELECT *
	FROM global_shares
	WHERE
		chain_id = ?
	AND
		period = ?
	AND
		pending = FALSE;`

	output := []*Share{}
	err := q.Select(&output, query, chain, period)

	return output, err
}

func GetPendingGlobalSharesByEndTime(q dbcl.Querier, timestamp time.Time, chain string, period int) ([]*Share, error) {
	const query = `SELECT
		chain_id, hashrate, reported_hashrate, count, period, start_time, end_time
	FROM global_shares
	WHERE
		end_time = ?
	AND
		chain_id = ?
	AND
		period = ?
	AND
		pending = TRUE;`

	output := []*Share{}
	err := q.Select(&output, query, timestamp, chain, period)

	return output, err
}

func GetMinerShares(q dbcl.Querier, minerID uint64, chain string, period int) ([]*Share, error) {
	const query = `SELECT *
	FROM miner_shares
	WHERE
		miner_id = ?
	AND
		chain_id = ?
	AND
		period = ?
	AND
		pending = FALSE;`

	output := []*Share{}
	err := q.Select(&output, query, minerID, chain, period)

	return output, err
}

func GetPendingMinerSharesByEndTime(q dbcl.Querier, timestamp time.Time, chain string, period int) ([]*Share, error) {
	const query = `SELECT
		chain_id, miner_id, hashrate, reported_hashrate, count, period, start_time, end_time
	FROM miner_shares
	WHERE
		end_time = ?
	AND
		chain_id = ?
	AND
		period = ?
	AND
		pending = TRUE;`

	output := []*Share{}
	err := q.Select(&output, query, timestamp, chain, period)

	return output, err
}

func GetWorkerShares(q dbcl.Querier, workerID uint64, chain string, period int) ([]*Share, error) {
	const query = `SELECT *
	FROM worker_shares 
	WHERE
		worker_id = ?
	AND
		chain_id = ?
	AND
		period = ?
	AND
		pending = FALSE;`

	output := []*Share{}
	err := q.Select(&output, query, workerID, chain, period)

	return output, err
}

func GetPendingWorkerSharesByEndTime(q dbcl.Querier, timestamp time.Time, chain string, period int) ([]*Share, error) {
	const query = `SELECT
		chain_id, worker_id, hashrate, reported_hashrate, count, period, start_time, end_time
	FROM worker_shares
	WHERE
		end_time = ?
	AND
		chain_id = ?
	AND
		period = ?
	AND
		pending = TRUE;`

	output := []*Share{}
	err := q.Select(&output, query, timestamp, chain, period)

	return output, err
}

func GetGlobalShareMaxEndTime(q dbcl.Querier, chain string, period int) (time.Time, error) {
	const query = `SELECT MAX(end_time)
	FROM global_shares 
	WHERE
		chain_id = ?
	AND
		period = ?;`

	return dbcl.GetTime(q, query, chain, period)
}

func GetRawBlockMaxHeight(q dbcl.Querier, chain string) (uint64, error) {
	const query = `SELECT IFNULL(MAX(height), 0)
	FROM raw_blocks
	WHERE
		chain_id = ?;`

	return dbcl.GetUint64(q, query, chain)
}

func GetRawBlockRollup(q dbcl.Querier, chain string, start, end time.Time) (*Block, error) {
	const query = `WITH cte as (
		SELECT
			id,
			chain_id,
			value,
			difficulty,
			uncle_count,
			tx_count,
			timestamp,
			LAG(timestamp, 1) OVER (ORDER BY timestamp) prev_timestamp
		FROM raw_blocks
		WHERE
			chain_id = ?
	) SELECT
		chain_id,
		IFNULL(AVG(value), 0) value,
		IFNULL(AVG(difficulty), 0) difficulty,
		IFNULL(AVG(TIMESTAMPDIFF(MICROSECOND, prev_timestamp, timestamp) / 1000000), 0) block_time,
		IFNULL(SUM(uncle_count), 0) uncle_count,
		IFNULL(SUM(tx_count), 0) tx_count,
		IFNULL(COUNT(id), 0) count
	FROM cte
	WHERE
		timestamp BETWEEN ? AND ?
	GROUP BY chain_id;`

	output := new(Block)
	err := q.Get(output, query, chain, start, end)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, err
}

func GetPendingRoundsAtEndTime(q dbcl.Querier, timestamp time.Time, chain string, period int) ([]*Round, error) {
	const query = `SELECT *
	FROM rounds
	WHERE
		end_time = ?
	AND
		chain_id = ?
	AND
		period = ?
	AND
		pending = true;`

	output := []*Round{}
	err := q.Select(&output, query, timestamp, chain, period)

	return output, err
}

func GetPendingBlocksAtEndTime(q dbcl.Querier, timestamp time.Time, chain string, period int) ([]*Block, error) {
	const query = `SELECT *
	FROM blocks
	WHERE
		end_time = ?
	AND
		chain_id = ?
	AND
		period = ?
	AND
		pending = true;`

	output := []*Block{}
	err := q.Select(&output, query, timestamp, chain, period)

	return output, err
}

func GetBlockMaxEndTime(q dbcl.Querier, chain string, period int) (time.Time, error) {
	const query = `SELECT MAX(end_time) 
	FROM blocks
	WHERE
		chain_id = ?
	AND
		period = ?;`

	return dbcl.GetTime(q, query, chain, period)
}

func GetRawBlockMaxTimestampBeforeTime(q dbcl.Querier, chain string, timestamp time.Time) (time.Time, error) {
	const query = `SELECT MAX(timestamp) 
	FROM raw_blocks
	WHERE
		timestamp < ?
	AND
		chain_id = ?;`

	return dbcl.GetTime(q, query, timestamp, chain)
}

func GetRoundMaxEndTime(q dbcl.Querier, chain string, period int) (time.Time, error) {
	const query = `SELECT MAX(end_time) 
	FROM rounds 
	WHERE
		chain_id = ?
	AND
		period = ?;`

	return dbcl.GetTime(q, query, chain, period)
}

func GetGlobalSharesAverageFast(q dbcl.Querier, timestamp time.Time, chain string, period, windowSize int, duration time.Duration) (float64, error) {
	var query = fmt.Sprintf(`WITH last AS (
    	SELECT hashrate, avg_hashrate
	    FROM global_shares
	    WHERE
	    	end_time = ?
		AND
			chain_id = ?
		AND
			period = ?
	), first as (
		SELECT hashrate
		FROM global_shares
		WHERE
			end_time = DATE_SUB(?, %s)
		AND
			chain_id = ?
		AND
			period = ?
	) SELECT
		((last.avg_hashrate * ?) - first.hashrate + last.hashrate) / ? as avg_hashrate
	FROM last
	JOIN first;`, dbcl.ConvertDurationToInterval(duration))

	return dbcl.GetFloat64(q, query, timestamp, chain, period, timestamp, chain, period, windowSize, windowSize)
}

func GetMinerSharesAverageFast(q dbcl.Querier, timestamp time.Time, chain string, period, windowSize int, duration time.Duration) (map[uint64]float64, error) {
	var query = fmt.Sprintf(`WITH last AS (
    	SELECT miner_id, hashrate, avg_hashrate
	    FROM miner_shares
	    WHERE
	    	end_time = ?
		AND
			chain_id = ?
		AND
			period = ?
	), first as (
		SELECT miner_id, hashrate
		FROM miner_shares
		WHERE
			end_time = DATE_SUB(?, %s)
		AND
			chain_id = ?
		AND
			period = ?
	) SELECT
		last.miner_id, 
		((last.avg_hashrate * ?) - first.hashrate + last.hashrate) / ? as avg_hashrate
	FROM last
	JOIN first ON last.miner_id = first.miner_id;`, dbcl.ConvertDurationToInterval(duration))

	items := []*Share{}
	err := q.Select(&items, query, timestamp, chain, period, timestamp, chain, period, windowSize, windowSize)
	if err != nil {
		return nil, err
	}

	output := make(map[uint64]float64, len(items))
	for _, item := range items {
		if item.MinerID != nil {
			output[*item.MinerID] = item.AvgHashrate
		}
	}

	return output, err
}

func GetWorkerSharesAverageFast(q dbcl.Querier, timestamp time.Time, chain string, period, windowSize int, duration time.Duration) (map[uint64]float64, error) {
	var query = fmt.Sprintf(`WITH last AS (
    	SELECT worker_id, hashrate, avg_hashrate
	    FROM worker_shares
	    WHERE
	    	end_time = ?
		AND
			chain_id = ?
		AND
			period = ?
	), first as (
		SELECT worker_id, hashrate
		FROM worker_shares
		WHERE
			end_time = DATE_SUB(?, %s)
		AND
			chain_id = ?
		AND
			period = ?
	) SELECT
		last.worker_id, 
		((last.avg_hashrate * ?) - first.hashrate + last.hashrate) / ? as avg_hashrate
	FROM last
	JOIN first ON last.worker_id = first.worker_id;`, dbcl.ConvertDurationToInterval(duration))

	items := []*Share{}
	err := q.Select(&items, query, timestamp, chain, period, timestamp, chain, period, windowSize, windowSize)
	if err != nil {
		return nil, err
	}

	output := make(map[uint64]float64, len(items))
	for _, item := range items {
		if item.WorkerID != nil {
			output[*item.WorkerID] = item.AvgHashrate
		}
	}

	return output, err
}

func GetBlocksAverageSlow(q dbcl.Querier, timestamp time.Time, chain string, period int, duration time.Duration) (float64, error) {
	var query = fmt.Sprintf(`SELECT IFNULL(AVG(profitability), 0)
	FROM blocks
    WHERE
    	end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		chain_id = ?
	AND
		period = ?;`, dbcl.ConvertDurationToInterval(duration))

	return dbcl.GetFloat64(q, query, timestamp, timestamp, chain, period)
}

func GetRoundsAverageLuckSlow(q dbcl.Querier, timestamp time.Time, chain string, period int, duration time.Duration) (float64, error) {
	var query = fmt.Sprintf(`SELECT IFNULL(AVG(luck), 0)
	FROM rounds
    WHERE
    	end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		chain_id = ?
	AND
		period = ?;`, dbcl.ConvertDurationToInterval(duration))

	return dbcl.GetFloat64(q, query, timestamp, timestamp, chain, period)
}

func GetRoundsAverageProfitabilitySlow(q dbcl.Querier, timestamp time.Time, chain string, period int, duration time.Duration) (float64, error) {
	var query = fmt.Sprintf(`SELECT IFNULL(AVG(profitability), 0)
	FROM rounds
    WHERE
    	end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		chain_id = ?
	AND
		period = ?;`, dbcl.ConvertDurationToInterval(duration))

	return dbcl.GetFloat64(q, query, timestamp, timestamp, chain, period)
}

func GetGlobalSharesAverage(q dbcl.Querier, timestamp time.Time, chain string, period int, duration time.Duration) (float64, error) {
	var query = fmt.Sprintf(`SELECT IFNULL(AVG(hashrate), 0)
	FROM global_shares
    WHERE
    	end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		chain_id = ?
	AND
		period = ?;`, dbcl.ConvertDurationToInterval(duration))

	return dbcl.GetFloat64(q, query, timestamp, timestamp, chain, period)
}

func GetGlobalSharesAverageSlow(q dbcl.Querier, timestamp time.Time, chain string, period int, duration time.Duration) (float64, error) {
	var query = fmt.Sprintf(`SELECT IFNULL(AVG(hashrate), 0)
	FROM global_shares
    WHERE
    	end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		chain_id = ?
	AND
		period = ?;`, dbcl.ConvertDurationToInterval(duration))

	return dbcl.GetFloat64(q, query, timestamp, timestamp, chain, period)
}

func GetMinerSharesAverage(q dbcl.Querier, timestamp time.Time, chain string, period int, duration time.Duration) (map[uint64]float64, error) {
	var query = fmt.Sprintf(`SELECT miner_id, IFNULL(AVG(hashrate), 0) avg_hashrate
	FROM miner_shares
    WHERE
    	end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		chain_id = ?
	AND
		period = ?
	GROUP BY miner_id`, dbcl.ConvertDurationToInterval(duration))

	items := []*Share{}
	err := q.Select(&items, query, timestamp, timestamp, chain, period)
	if err != nil {
		return nil, err
	}

	output := make(map[uint64]float64, len(items))
	for _, item := range items {
		if item.MinerID == nil {
			return nil, fmt.Errorf("empty minerID")
		}
		output[*item.MinerID] = item.AvgHashrate
	}

	return output, err
}

func GetMinerSharesAverageSlow(q dbcl.Querier, minerID uint64, timestamp time.Time, chain string, period int, duration time.Duration) (float64, error) {
	var query = fmt.Sprintf(`SELECT IFNULL(AVG(hashrate), 0)
	FROM miner_shares
    WHERE
    	end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		miner_id = ?
	AND
		chain_id = ?
	AND
		period = ?;`, dbcl.ConvertDurationToInterval(duration))

	return dbcl.GetFloat64(q, query, timestamp, timestamp, minerID, chain, period)
}

func GetWorkerSharesAverage(q dbcl.Querier, timestamp time.Time, chain string, period int, duration time.Duration) (map[uint64]float64, error) {
	var query = fmt.Sprintf(`SELECT worker_id, IFNULL(AVG(hashrate), 0) avg_hashrate
	FROM worker_shares
    WHERE
    	end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		chain_id = ?
	AND
		period = ?
	GROUP BY worker_id`, dbcl.ConvertDurationToInterval(duration))

	items := []*Share{}
	err := q.Select(&items, query, timestamp, timestamp, chain, period)
	if err != nil {
		return nil, err
	}

	output := make(map[uint64]float64, len(items))
	for _, item := range items {
		if item.WorkerID == nil {
			return nil, fmt.Errorf("empty workerID")
		}
		output[*item.WorkerID] = item.AvgHashrate
	}

	return output, err
}

func GetWorkerSharesAverageSlow(q dbcl.Querier, workerID uint64, timestamp time.Time, chain string, period int, duration time.Duration) (float64, error) {
	var query = fmt.Sprintf(`SELECT IFNULL(AVG(hashrate), 0)
	FROM worker_shares
    WHERE
    	end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		worker_id = ?
	AND
		chain_id = ?
	AND
		period = ?;`, dbcl.ConvertDurationToInterval(duration))

	return dbcl.GetFloat64(q, query, timestamp, timestamp, workerID, chain, period)
}

/* sums */

func GetGlobalSharesSum(q dbcl.Querier, period int, duration time.Duration) ([]*Share, error) {
	var query = fmt.Sprintf(`SELECT
		chain_id,
		IFNULL(SUM(accepted_shares), 0) accepted_shares,
		IFNULL(SUM(rejected_shares), 0) rejected_shares,
		IFNULL(SUM(invalid_shares), 0) invalid_shares
	FROM global_shares
    WHERE
		end_time BETWEEN DATE_SUB(CURRENT_TIMESTAMP, %s) AND CURRENT_TIMESTAMP
	AND
		period = ?
	GROUP BY chain_id`, dbcl.ConvertDurationToInterval(duration))

	output := []*Share{}
	err := q.Select(&output, query, period)

	return output, err
}

func GetGlobalSharesLast(q dbcl.Querier, period int) ([]*Share, error) {
	const query = `SELECT
		chain_id,
		miners,
		workers,
		hashrate,
		avg_hashrate,
		reported_hashrate
	FROM global_shares
	WHERE
		end_time = (
			SELECT MAX(end_time)
			FROM global_shares
			WHERE
				period = ?
		)
	AND
		period = ?;`

	output := []*Share{}
	err := q.Select(&output, query, period, period)

	return output, err
}

func GetMinerSharesSum(q dbcl.Querier, minerID uint64, period int, duration time.Duration) ([]*Share, error) {
	var query = fmt.Sprintf(`SELECT
		chain_id,
		IFNULL(SUM(accepted_shares), 0) accepted_shares,
		IFNULL(SUM(rejected_shares), 0) rejected_shares,
		IFNULL(SUM(invalid_shares), 0) invalid_shares
	FROM miner_shares
    WHERE
		end_time BETWEEN DATE_SUB(CURRENT_TIMESTAMP, %s) AND CURRENT_TIMESTAMP
	AND
		miner_id = ?
	AND
		period = ?
	GROUP BY chain_id`, dbcl.ConvertDurationToInterval(duration))

	output := []*Share{}
	err := q.Select(&output, query, minerID, period)

	return output, err
}

func GetMinerSharesLast(q dbcl.Querier, minerID uint64, period int) ([]*Share, error) {
	const query = `SELECT
		chain_id,
		hashrate,
		avg_hashrate,
		reported_hashrate
	FROM miner_shares
	WHERE
		end_time = (
			SELECT MAX(end_time)
			FROM miner_shares
			WHERE
				miner_id = ?
			AND
				period = ?
		)
    AND
		miner_id = ?
	AND
		period = ?;`

	output := []*Share{}
	err := q.Select(&output, query, minerID, period, minerID, period)

	return output, err
}

func GetWorkerSharesSum(q dbcl.Querier, workerID uint64, period int, duration time.Duration) ([]*Share, error) {
	var query = fmt.Sprintf(`SELECT
		chain_id,
		IFNULL(SUM(accepted_shares), 0) accepted_shares,
		IFNULL(SUM(rejected_shares), 0) rejected_shares,
		IFNULL(SUM(invalid_shares), 0) invalid_shares
	FROM worker_shares
    WHERE
		end_time BETWEEN DATE_SUB(CURRENT_TIMESTAMP, %s) AND CURRENT_TIMESTAMP
	AND
		worker_id = ?
	AND
		period = ?
	GROUP BY chain_id`, dbcl.ConvertDurationToInterval(duration))

	output := []*Share{}
	err := q.Select(&output, query, workerID, period)

	return output, err
}

func GetWorkerSharesLast(q dbcl.Querier, workerID uint64, period int) ([]*Share, error) {
	const query = `SELECT
		chain_id,
		hashrate,
		avg_hashrate,
		reported_hashrate
	FROM worker_shares
    WHERE
		end_time = (
			SELECT MAX(end_time)
			FROM worker_shares
			WHERE
				worker_id = ?
			AND
				period = ?
		)
    AND
		worker_id = ?
    AND
		period = ?;`

	output := []*Share{}
	err := q.Select(&output, query, workerID, period, workerID, period)

	return output, err
}

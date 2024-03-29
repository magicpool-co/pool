package tsdb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

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

func GetBlocksSingleMetric(q dbcl.Querier, metric string, period int) ([]*Block, error) {
	var query = fmt.Sprintf(`SELECT chain_id, %s, end_time
	FROM blocks
	WHERE
		period = ?
	AND
		pending = FALSE;`, metric)

	output := []*Block{}
	err := q.Select(&output, query, period)

	return output, err
}

func GetBlocksAdjustedValue(q dbcl.Querier, period int) ([]*Block, error) {
	const query = `SELECT
		blocks.chain_id,
		blocks.value * prices.price_usd value,
		blocks.end_time
	FROM blocks
	JOIN prices ON blocks.end_time = prices.timestamp AND blocks.chain_id = prices.chain_id
	WHERE
		period = ?
	AND
		pending = FALSE;`

	output := []*Block{}
	err := q.Select(&output, query, period)

	return output, err
}

func GetBlocksAdjustedEmission(q dbcl.Querier, period int) ([]*Block, error) {
	const query = `SELECT
		blocks.chain_id,
		blocks.value * blocks.count * prices.price_usd value,
		blocks.end_time
	FROM blocks
	JOIN prices ON blocks.end_time = prices.timestamp AND blocks.chain_id = prices.chain_id
	WHERE
		period = ?
	AND
		pending = FALSE;`

	output := []*Block{}
	err := q.Select(&output, query, period)

	return output, err
}

func GetBlocksProfitability(q dbcl.Querier, period int) ([]*Block, error) {
	const query = `SELECT
		blocks.chain_id,
		blocks.profitability * prices.price_usd profitability,
		blocks.avg_profitability * prices.price_usd avg_profitability,
		blocks.end_time
	FROM blocks
	JOIN prices ON blocks.end_time = prices.timestamp AND blocks.chain_id = prices.chain_id
	WHERE
		period = ?
	AND
		pending = FALSE;`

	output := []*Block{}
	err := q.Select(&output, query, period)

	return output, err
}

func GetBlocksWithProfitabilityLast(q dbcl.Querier, period int) ([]*Block, error) {
	const query = `WITH block_times AS (
	    SELECT chain_id, MAX(end_time) AS end_time
	    FROM blocks
	    WHERE period = ?
	    GROUP BY chain_id
	), price_times AS (
	    SELECT chain_id, MAX(timestamp) AS timestamp
	    FROM prices
	    GROUP BY chain_id
	), prices AS (
	    SELECT prices.chain_id, price_usd, price_btc, price_eth
	    FROM prices
	    JOIN price_times ON prices.chain_id = price_times.chain_id AND prices.timestamp = price_times.timestamp
	) SELECT
	      blocks.chain_id,
	      blocks.hashrate,
	      blocks.difficulty,
	      blocks.value,
	      blocks.block_time,
	      blocks.profitability,
	      blocks.avg_profitability,
	      blocks.profitability * prices.price_usd AS profitability_usd,
	      blocks.avg_profitability * prices.price_usd AS avg_profitability_usd,
	      blocks.profitability * prices.price_btc AS profitability_btc,
	      blocks.avg_profitability * prices.price_btc AS avg_profitability_btc,
	      blocks.profitability * prices.price_eth AS profitability_eth,
	      blocks.avg_profitability * prices.price_eth AS avg_profitability_eth
	FROM blocks
	JOIN block_times ON blocks.chain_id = block_times.chain_id AND blocks.end_time = block_times.end_time
	JOIN prices ON blocks.chain_id = prices.chain_id
	WHERE
	        period = ?
	  AND
	        pending = FALSE;`

	output := []*Block{}
	err := q.Select(&output, query, period, period)

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

func GetGlobalSharesSingleMetric(q dbcl.Querier, metric string, period int) ([]*Share, error) {
	var query = fmt.Sprintf(`SELECT chain_id, %s, end_time
	FROM global_shares
	WHERE
		period = ?
	AND
		pending = FALSE;`, metric)

	output := []*Share{}
	err := q.Select(&output, query, period)

	return output, err
}

func GetPendingGlobalSharesByEndTime(
	q dbcl.Querier,
	timestamp time.Time,
	chain string,
	period int,
) ([]*Share, error) {
	const query = `SELECT
		chain_id, hashrate, count, period, start_time, end_time
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

func GetMinerShares(q dbcl.Querier, minerIDs []uint64, chain string, period int) ([]*Share, error) {
	const rawQuery = `SELECT *
	FROM miner_shares
	WHERE
		miner_id IN (?)
	AND
		chain_id = ?
	AND
		period = ?
	AND
		pending = FALSE;`

	if len(minerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs, chain, period)
	if err != nil {
		return nil, err
	}

	output := []*Share{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

func GetMinerSharesSingleMetric(
	q dbcl.Querier,
	minerIDs []uint64,
	metric string,
	period int,
) ([]*Share, error) {
	var rawQuery = fmt.Sprintf(`SELECT chain_id, %s, end_time
	FROM miner_shares
	WHERE
		miner_id IN (?)
	AND
		period = ?
	AND
		pending = FALSE;`, metric)

	if len(minerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs, period)
	if err != nil {
		return nil, err
	}

	output := []*Share{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

func GetPendingMinerSharesByEndTime(
	q dbcl.Querier,
	timestamp time.Time,
	chain string, period int,
) ([]*Share, error) {
	const query = `SELECT
		chain_id, miner_id, hashrate, count, period, start_time, end_time
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

func GetMinerSharesByEndTime(
	q dbcl.Querier,
	timestamp time.Time,
	minerIDs []uint64,
	chain string,
	period int,
) ([]*Share, error) {
	const rawQuery = `SELECT
		miner_id,
		chain_id,
		hashrate,
		avg_hashrate
	FROM miner_shares
    WHERE
		end_time = ?
	AND
		miner_id IN (?)
	AND
		chain_id = ?
	AND
		period = ?;`

	if len(minerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, timestamp, minerIDs, chain, period)
	if err != nil {
		return nil, err
	}

	output := []*Share{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

func GetWorkerShares(
	q dbcl.Querier,
	workerID uint64,
	chain string,
	period int,
) ([]*Share, error) {
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

func GetWorkerSharesSingleMetric(
	q dbcl.Querier,
	workerID uint64,
	metric string,
	period int,
) ([]*Share, error) {
	var query = fmt.Sprintf(`SELECT chain_id, %s, end_time
	FROM worker_shares
	WHERE
		worker_id = ?
	AND
		period = ?
	AND
		pending = FALSE;`, metric)

	output := []*Share{}
	err := q.Select(&output, query, workerID, period)

	return output, err
}

func GetPendingWorkerSharesByEndTime(
	q dbcl.Querier,
	timestamp time.Time,
	chain string, period int,
) ([]*Share, error) {
	const query = `SELECT
		chain_id, worker_id, hashrate, count, period, start_time, end_time
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

func GetWorkerSharesAllChainsByEndTime(
	q dbcl.Querier,
	timestamp time.Time,
	workerIDs []uint64,
	period int,
) ([]*Share, error) {
	const rawQuery = `SELECT
		worker_id,
		chain_id,
		hashrate,
		avg_hashrate
	FROM worker_shares
    WHERE
		end_time = ?
	AND
		worker_id IN (?)
	AND
		period = ?;`

	if len(workerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, timestamp, workerIDs, period)
	if err != nil {
		return nil, err
	}

	output := []*Share{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

func GetRawBlockMaxTimestamp(q dbcl.Querier, chain string) (time.Time, error) {
	const query = `SELECT MAX(timestamp)
	FROM raw_blocks
	WHERE
		chain_id = ?;`

	return dbcl.GetTime(q, query, chain)
}

func GetRawBlockMaxHeight(q dbcl.Querier, chain string) (uint64, error) {
	const query = `SELECT IFNULL(MAX(height), 0)
	FROM raw_blocks
	WHERE
		chain_id = ?;`

	return dbcl.GetUint64(q, query, chain)
}

func GetRawBlockMaxHashByHeight(q dbcl.Querier, chain string) (string, error) {
	const query = `WITH cte as (
			SELECT MAX(height) height
		FROM
			raw_blocks
		WHERE
			chain_id = ?
	) SELECT IFNULL(hash, "")
	FROM raw_blocks
	JOIN cte ON raw_blocks.height = cte.height
	ORDER BY id DESC
	LIMIT 1;`

	return dbcl.GetString(q, query, chain)
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

func GetPendingBlocksAtEndTime(
	q dbcl.Querier,
	timestamp time.Time,
	chain string,
	period int,
) ([]*Block, error) {
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

func GetPriceMaxTimestamp(q dbcl.Querier, chain string) (time.Time, error) {
	const query = `SELECT MAX(timestamp) 
	FROM prices
	WHERE
		chain_id = ?;`

	return dbcl.GetTime(q, query, chain)
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

func GetEarningMaxEndTime(q dbcl.Querier, chain string, period int) (time.Time, error) {
	const query = `SELECT MAX(end_time)
	FROM global_earnings
	WHERE
		chain_id = ?
	AND
		period = ?;`

	return dbcl.GetTime(q, query, chain, period)
}

func GetGlobalEarningsSingleMetric(q dbcl.Querier, metric string, period int) ([]*Earning, error) {
	var query = fmt.Sprintf(`SELECT chain_id, %s, end_time
	FROM global_earnings
	WHERE
		period = ?
	AND
		pending = FALSE;`, metric)

	output := []*Earning{}
	err := q.Select(&output, query, period)

	return output, err
}

func GetMinerEarningsSingleMetric(
	q dbcl.Querier,
	minerIDs []uint64,
	metric string,
	period int,
) ([]*Earning, error) {
	var rawQuery = fmt.Sprintf(`SELECT chain_id, %s, end_time
	FROM miner_earnings
	WHERE
		miner_id IN (?)
	AND
		period = ?
	AND
		pending = FALSE;`, metric)

	if len(minerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs, period)
	if err != nil {
		return nil, err
	}

	output := []*Earning{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

/* averages */

func GetBlocksAverageSlow(
	q dbcl.Querier,
	timestamp time.Time,
	chain string,
	period int,
	duration time.Duration,
) (float64, error) {
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

func GetGlobalSharesAverage(
	q dbcl.Querier,
	timestamp time.Time,
	chain string,
	period int,
	duration time.Duration,
) (float64, error) {
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

func GetMinerSharesAverage(
	q dbcl.Querier,
	timestamp time.Time,
	chain string,
	period int,
	duration time.Duration,
) (map[uint64]float64, error) {
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

func GetWorkerSharesAverage(
	q dbcl.Querier,
	timestamp time.Time,
	chain string,
	period int,
	duration time.Duration,
) (map[uint64]float64, error) {
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

func GetGlobalEarningsAverage(
	q dbcl.Querier,
	timestamp time.Time,
	chain string,
	period int,
	duration time.Duration,
) (float64, error) {
	var query = fmt.Sprintf(`SELECT IFNULL(AVG(value), 0) avg_value
	FROM global_earnings
	WHERE
		end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		chain_id = ?
	AND
		period = ?;`, dbcl.ConvertDurationToInterval(duration))

	return dbcl.GetFloat64(q, query, timestamp, timestamp, chain, period)
}

func GetMinerEarningsAverage(
	q dbcl.Querier,
	timestamp time.Time,
	chain string,
	period int,
	duration time.Duration,
) (map[uint64]float64, error) {
	var query = fmt.Sprintf(`SELECT miner_id, IFNULL(AVG(value), 0) avg_value
	FROM miner_earnings
	WHERE
		end_time BETWEEN DATE_SUB(?, %s) AND ?
	AND
		chain_id = ?
	AND
		period = ?
	GROUP BY miner_id`, dbcl.ConvertDurationToInterval(duration))

	items := []*Earning{}
	err := q.Select(&items, query, timestamp, timestamp, chain, period)
	if err != nil {
		return nil, err
	}

	output := make(map[uint64]float64, len(items))
	for _, item := range items {
		if item.MinerID == nil {
			return nil, fmt.Errorf("empty minerID")
		}
		output[*item.MinerID] = item.AvgValue
	}

	return output, err
}

/* sums */

func GetGlobalSharesLast(q dbcl.Querier, period int) ([]*Share, error) {
	const query = `SELECT
		chain_id,
		miners,
		workers,
		hashrate,
		avg_hashrate
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

func GetMinersSharesSum(
	q dbcl.Querier,
	minerIDs []uint64,
	period int,
	duration time.Duration,
) ([]*Share, error) {
	var rawQuery = fmt.Sprintf(`SELECT
		chain_id,
		IFNULL(SUM(accepted_adjusted_shares), 0) accepted_adjusted_shares,
		IFNULL(SUM(rejected_adjusted_shares), 0) rejected_adjusted_shares,
		IFNULL(SUM(invalid_adjusted_shares), 0) invalid_adjusted_shares
	FROM miner_shares
    WHERE
		end_time BETWEEN DATE_SUB(CURRENT_TIMESTAMP, %s) AND CURRENT_TIMESTAMP
	AND
		miner_id IN (?)
	AND
		period = ?
	GROUP BY chain_id`, dbcl.ConvertDurationToInterval(duration))

	if len(minerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs, period)
	if err != nil {
		return nil, err
	}

	output := []*Share{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

func GetMinersSharesLast(
	q dbcl.Querier,
	minerIDs []uint64,
	period int,
) ([]*Share, error) {
	const rawQuery = `SELECT
		chain_id,
		IFNULL(SUM(hashrate), 0) hashrate,
		IFNULL(SUM(avg_hashrate), 0) avg_hashrate
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
		miner_id IN (?)
	AND
		period = ?
	GROUP BY chain_id;`

	if len(minerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs[0], period, minerIDs, period)
	if err != nil {
		return nil, err
	}

	output := []*Share{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

func GetWorkerSharesSum(
	q dbcl.Querier,
	workerIDs []uint64,
	period int,
	duration time.Duration,
) ([]*Share, error) {
	var rawQuery = fmt.Sprintf(`SELECT
		worker_id,
		chain_id,
		IFNULL(SUM(accepted_adjusted_shares), 0) accepted_adjusted_shares,
		IFNULL(SUM(rejected_adjusted_shares), 0) rejected_adjusted_shares,
		IFNULL(SUM(invalid_adjusted_shares), 0) invalid_adjusted_shares
	FROM worker_shares
    WHERE
		end_time BETWEEN DATE_SUB(CURRENT_TIMESTAMP, %s) AND CURRENT_TIMESTAMP
	AND
		worker_id IN (?)
	AND
		period = ?
	GROUP BY worker_id, chain_id`, dbcl.ConvertDurationToInterval(duration))

	if len(workerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, workerIDs, period)
	if err != nil {
		return nil, err
	}

	output := []*Share{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

func GetWorkerSharesLast(q dbcl.Querier, workerID uint64, period int) ([]*Share, error) {
	const query = `SELECT
		chain_id,
		hashrate,
		avg_hashrate
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

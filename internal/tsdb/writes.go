package tsdb

import (
	"fmt"
	"time"

	"github.com/magicpool-co/pool/pkg/dbcl"
)

func InsertRawBlocks(q dbcl.Querier, objects ...*RawBlock) error {
	const table = "raw_blocks"
	cols := []string{"chain_id", "hash", "height", "value", "difficulty", "uncle_count", "tx_count", "timestamp"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

func DeleteRawBlocksBeforeTime(q dbcl.Querier, chain string, timestamp time.Time) error {
	const query = `DELETE FROM raw_blocks
	WHERE
		timestamp < ?
	AND
		chain_id = ?;`

	_, err := q.Exec(query, timestamp, chain)

	return err
}

func DeleteBlocksBeforeEndTime(q dbcl.Querier, timestamp time.Time, chain string, period int) error {
	const query = `DELETE FROM blocks
	WHERE
		end_time < ?
	AND
		chain_id = ?
	AND
		period = ?;`

	_, err := q.Exec(query, timestamp, chain, period)

	return err
}

func DeleteRoundsBeforeEndTime(q dbcl.Querier, timestamp time.Time, chain string, period int) error {
	const query = `DELETE FROM rounds
	WHERE
		end_time < ?
	AND
		chain_id = ?
	AND
		period = ?;`

	_, err := q.Exec(query, timestamp, chain, period)

	return err
}

func DeleteGlobalSharesBeforeEndTime(q dbcl.Querier, timestamp time.Time, chain string, period int) error {
	const query = `DELETE FROM global_shares
	WHERE
		end_time < ?
	AND
		chain_id = ?
	AND
		period = ?;`

	_, err := q.Exec(query, timestamp, chain, period)

	return err
}

func DeleteMinerSharesBeforeEndTime(q dbcl.Querier, timestamp time.Time, chain string, period int) error {
	const query = `DELETE FROM miner_shares
	WHERE
		end_time < ?
	AND
		chain_id = ?
	AND
		period = ?;`

	_, err := q.Exec(query, timestamp, chain, period)

	return err
}

func DeleteWorkerSharesBeforeEndTime(q dbcl.Querier, timestamp time.Time, chain string, period int) error {
	const query = `DELETE FROM worker_shares
	WHERE
		end_time < ?
	AND
		chain_id = ?
	AND
		period = ?;`

	_, err := q.Exec(query, timestamp, chain, period)

	return err
}

func InsertPrices(q dbcl.Querier, objects ...*Price) error {
	const table = "prices"
	cols := []string{"chain_id", "price_usd", "price_btc", "price_eth", "timestamp"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

func InsertBlocks(q dbcl.Querier, objects ...*Block) error {
	const table = "blocks"
	cols := []string{"chain_id", "value", "difficulty", "block_time", "hashrate", "uncle_rate",
		"profitability", "avg_profitability", "pending", "count", "uncle_count", "tx_count", "period",
		"start_time", "end_time"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

func InsertPartialBlocks(q dbcl.Querier, objects ...*Block) error {
	const table = "blocks"
	insertCols := []string{"chain_id", "value", "difficulty", "block_time", "hashrate", "uncle_rate",
		"profitability", "avg_profitability", "pending", "count", "uncle_count", "tx_count", "period",
		"start_time", "end_time"}
	updateCols := []string{"value", "difficulty", "block_time", "hashrate", "uncle_rate", "profitability",
		"count", "uncle_count", "tx_count"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateAdd(q, table, insertCols, updateCols, rawObjects)
}

func InsertFinalBlocks(q dbcl.Querier, objects ...*Block) error {
	const table = "blocks"
	insertCols := []string{"chain_id", "value", "difficulty", "block_time", "hashrate", "uncle_rate",
		"profitability", "avg_profitability", "pending", "count", "uncle_count", "tx_count", "period",
		"start_time", "end_time"}
	updateCols := []string{"value", "difficulty", "block_time", "hashrate", "uncle_rate",
		"profitability", "avg_profitability", "pending"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateOverwrite(q, table, insertCols, updateCols, rawObjects)
}

func InsertRounds(q dbcl.Querier, objects ...*Round) error {
	const table = "rounds"
	cols := []string{"chain_id", "value", "difficulty", "round_time", "accepted_shares", "rejected_shares",
		"invalid_shares", "hashrate", "uncle_rate", "luck", "avg_luck", "profitability", "avg_profitability",
		"pending", "count", "uncle_count", "period", "start_time", "end_time"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

func InsertPartialRounds(q dbcl.Querier, objects ...*Round) error {
	const table = "rounds"
	insertCols := []string{"chain_id", "value", "difficulty", "round_time", "accepted_shares", "rejected_shares",
		"invalid_shares", "hashrate", "uncle_rate", "luck", "avg_luck", "profitability", "avg_profitability",
		"pending", "count", "uncle_count", "period", "start_time", "end_time"}
	updateCols := []string{"value", "difficulty", "round_time", "accepted_shares", "rejected_shares", "invalid_shares"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateAdd(q, table, insertCols, updateCols, rawObjects)
}

func InsertFinalRounds(q dbcl.Querier, objects ...*Round) error {
	const table = "rounds"
	insertCols := []string{"chain_id", "value", "difficulty", "round_time", "accepted_shares", "rejected_shares",
		"invalid_shares", "hashrate", "uncle_rate", "luck", "avg_luck", "profitability", "avg_profitability",
		"pending", "count", "uncle_count", "period", "start_time", "end_time"}
	updateCols := []string{"value", "difficulty", "round_time", "accepted_shares", "rejected_shares", "invalid_shares",
		"hashrate", "uncle_rate", "luck", "avg_luck", "profitability", "avg_profitability", "pending"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateOverwrite(q, table, insertCols, updateCols, rawObjects)
}

func InsertGlobalShares(q dbcl.Querier, objects ...*Share) error {
	const table = "global_shares"
	cols := []string{"chain_id", "miners", "workers", "accepted_shares", "rejected_shares", "invalid_shares",
		"hashrate", "avg_hashrate", "pending", "count", "period", "start_time", "end_time"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

func InsertPartialGlobalShares(q dbcl.Querier, objects ...*Share) error {
	const table = "global_shares"
	insertCols := []string{"chain_id", "miners", "workers", "accepted_shares", "rejected_shares", "invalid_shares",
		"hashrate", "avg_hashrate", "pending", "count", "period", "start_time", "end_time"}
	updateCols := []string{"miners", "workers", "accepted_shares", "invalid_shares", "rejected_shares", "hashrate",
		"avg_hashrate", "count"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateAdd(q, table, insertCols, updateCols, rawObjects)
}

func InsertFinalGlobalShares(q dbcl.Querier, objects ...*Share) error {
	const table = "global_shares"
	insertCols := []string{"chain_id", "miners", "workers", "accepted_shares", "rejected_shares", "invalid_shares",
		"hashrate", "avg_hashrate", "pending", "count", "period", "start_time", "end_time"}
	updateCols := []string{"hashrate", "avg_hashrate", "pending"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateOverwrite(q, table, insertCols, updateCols, rawObjects)
}

func InsertMinerShares(q dbcl.Querier, objects ...*Share) error {
	const table = "miner_shares"
	cols := []string{"chain_id", "miner_id", "workers", "accepted_shares", "rejected_shares", "invalid_shares",
		"hashrate", "avg_hashrate", "pending", "count", "period", "start_time", "end_time"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		if object.MinerID == nil {
			return fmt.Errorf("minerID is nil")
		}
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

func InsertPartialMinerShares(q dbcl.Querier, objects ...*Share) error {
	const table = "miner_shares"
	insertCols := []string{"chain_id", "miner_id", "workers", "accepted_shares", "rejected_shares", "invalid_shares",
		"hashrate", "avg_hashrate", "pending", "count", "period", "start_time", "end_time"}
	updateCols := []string{"workers", "accepted_shares", "rejected_shares", "hashrate", "count"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateAdd(q, table, insertCols, updateCols, rawObjects)
}

func InsertFinalMinerShares(q dbcl.Querier, objects ...*Share) error {
	const table = "miner_shares"
	insertCols := []string{"chain_id", "miner_id", "workers", "accepted_shares", "rejected_shares", "invalid_shares",
		"hashrate", "avg_hashrate", "pending", "count", "period", "start_time", "end_time"}
	updateCols := []string{"hashrate", "avg_hashrate", "pending"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateOverwrite(q, table, insertCols, updateCols, rawObjects)
}

func InsertWorkerShares(q dbcl.Querier, objects ...*Share) error {
	const table = "worker_shares"
	cols := []string{"chain_id", "worker_id", "accepted_shares", "rejected_shares", "invalid_shares",
		"hashrate", "avg_hashrate", "pending", "count", "period", "start_time", "end_time"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		if object.WorkerID == nil {
			return fmt.Errorf("workerID is nil")
		}
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

func InsertPartialWorkerShares(q dbcl.Querier, objects ...*Share) error {
	const table = "worker_shares"
	insertCols := []string{"chain_id", "worker_id", "accepted_shares", "rejected_shares", "invalid_shares",
		"hashrate", "avg_hashrate", "pending", "count", "period", "start_time", "end_time"}
	updateCols := []string{"accepted_shares", "rejected_shares", "hashrate", "count"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateAdd(q, table, insertCols, updateCols, rawObjects)
}

func InsertFinalWorkerShares(q dbcl.Querier, objects ...*Share) error {
	const table = "worker_shares"
	insertCols := []string{"chain_id", "worker_id", "accepted_shares", "rejected_shares", "invalid_shares",
		"hashrate", "avg_hashrate", "pending", "count", "period", "start_time", "end_time"}
	updateCols := []string{"hashrate", "avg_hashrate", "pending"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateOverwrite(q, table, insertCols, updateCols, rawObjects)
}

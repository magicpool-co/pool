package pooldb

import (
	"database/sql"
	"math/big"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/magicpool-co/pool/pkg/dbcl"
)

/* nodes */

func GetNodeURLsByChain(q dbcl.Querier, chain string, mainnet bool) ([]string, error) {
	const query = `SELECT url 
	FROM nodes 
	WHERE
		chain_id = ?
	AND
		mainnet = ?
	AND
		enabled IS TRUE;`

	output := []string{}
	err := q.Select(&output, query, chain, mainnet)

	return output, err
}

func GetEnabledNodes(q dbcl.Querier, mainnet bool) ([]*Node, error) {
	const query = `SELECT * 
	FROM nodes 
	WHERE
		mainnet = ?
	AND
		enabled IS TRUE;`

	output := []*Node{}
	err := q.Select(&output, query, mainnet)

	return output, err
}

func GetBackupNodes(q dbcl.Querier, mainnet bool) ([]*Node, error) {
	const query = `SELECT * 
	FROM nodes 
	WHERE
		mainnet = ?
	AND
		backup IS TRUE
	AND
		enabled IS TRUE;`

	output := []*Node{}
	err := q.Select(&output, query, mainnet)

	return output, err
}

func GetPendingBackupNodes(q dbcl.Querier, mainnet bool) ([]*Node, error) {
	const query = `SELECT * 
	FROM nodes 
	WHERE
		mainnet = ?
	AND
		pending_backup IS TRUE
	AND
		enabled IS TRUE;`

	output := []*Node{}
	err := q.Select(&output, query, mainnet)

	return output, err
}

func GetPendingUpdateNodes(q dbcl.Querier, mainnet bool) ([]*Node, error) {
	const query = `SELECT * 
	FROM nodes 
	WHERE
		mainnet = ?
	AND
		pending_update IS TRUE
	AND
		enabled IS TRUE;`

	output := []*Node{}
	err := q.Select(&output, query, mainnet)

	return output, err
}

func GetPendingResizeNodes(q dbcl.Querier, mainnet bool) ([]*Node, error) {
	const query = `SELECT * 
	FROM nodes 
	WHERE
		mainnet = ?
	AND
		pending_resize IS TRUE
	AND
		enabled IS TRUE;`

	output := []*Node{}
	err := q.Select(&output, query, mainnet)

	return output, err
}

/* miners */

func GetMiner(q dbcl.Querier, id uint64) (*Miner, error) {
	const query = `SELECT *
	FROM miners
	WHERE
		id = ?;`

	output := new(Miner)
	err := q.Get(output, query, id)
	if err != nil && err != sql.ErrNoRows {
		return output, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, nil
}

func GetMinerID(q dbcl.Querier, address, chain string) (uint64, error) {
	const query = `SELECT id
	FROM miners
	WHERE
		address = ?
	AND
		chain_id = ?;`

	return dbcl.GetUint64(q, query, address, chain)
}

func GetMinerAddress(q dbcl.Querier, minerID uint64) (string, error) {
	const query = `SELECT address
	FROM miners
	WHERE
		id = ?;`

	return dbcl.GetString(q, query, minerID)
}

func GetMiners(q dbcl.Querier, minerIDs []uint64) ([]*Miner, error) {
	const rawQuery = `SELECT *
	FROM miners
	WHERE
		id IN (?);`

	if len(minerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs)
	if err != nil {
		return nil, err
	}

	output := []*Miner{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

/* recipients */

func GetRecipients(q dbcl.Querier) ([]*Miner, error) {
	const query = `SELECT *
	FROM miners
	WHERE
		recipient_fee_percent IS NOT NULL;`

	output := []*Miner{}
	err := q.Select(&output, query)

	return output, err
}

/* workers */

func GetWorkerID(q dbcl.Querier, minerID uint64, name string) (uint64, error) {
	const query = `SELECT id
	FROM workers
	WHERE
		miner_id = ?
	AND
		name = ?`

	return dbcl.GetUint64(q, query, minerID, name)
}

func GetWorkersByMiner(q dbcl.Querier, minerID uint64) ([]*Worker, error) {
	const query = `SELECT
		workers.name,
		ip_addresses.active,
		ip_addresses.last_share
	FROM workers
	JOIN ip_addresses ON workers.id = ip_addresses.worker_id
	WHERE
		workers.miner_id = ?;`

	output := []*Worker{}
	err := q.Select(&output, query, minerID)

	return output, err
}

/* ip addresses */

func GetActiveMinersCount(q dbcl.Querier) (uint64, error) {
	const query = `SELECT COUNT(DISTINCT miner_id)
	FROM ip_addresses
	WHERE
		active IS TRUE;`

	return dbcl.GetUint64(q, query)
}

func GetActiveWorkersCount(q dbcl.Querier) (uint64, error) {
	const query = `SELECT COUNT(DISTINCT worker_id)
	FROM ip_addresses
	WHERE
		worker_id != 0
	AND
		active IS TRUE;`

	return dbcl.GetUint64(q, query)
}

func GetActiveWorkersByMinerCount(q dbcl.Querier, minerID uint64) (uint64, error) {
	const query = `SELECT COUNT(DISTINCT worker_id)
	FROM ip_addresses
	WHERE
		miner_id = ?
	AND
		worker_id != 0
	AND
		active IS TRUE;`

	return dbcl.GetUint64(q, query, minerID)
}

func GetInactiveWorkersByMinerCount(q dbcl.Querier, minerID uint64) (uint64, error) {
	const query = `SELECT COUNT(DISTINCT worker_id)
	FROM ip_addresses
	WHERE
		miner_id = ?
	AND
		worker_id != 0
	AND
		active IS FALSE;`

	return dbcl.GetUint64(q, query, minerID)
}

func GetOldestActiveIPAddress(q dbcl.Querier, minerID uint64) (*IPAddress, error) {
	const query = `SELECT *
	FROM ip_addresses
	WHERE
		miner_id = ?
	AND
		active IS TRUE
	ORDER BY created_at
	LIMIT 1;`

	output := new(IPAddress)
	err := q.Get(output, query, minerID)
	if err != nil && err != sql.ErrNoRows {
		return output, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, nil
}

/* Round Queries */

func GetRound(q dbcl.Querier, id uint64) (*Round, error) {
	const query = `SELECT *
	FROM rounds
	WHERE
		id = ?`

	output := new(Round)
	err := q.Get(output, query, id)
	if err != nil && err != sql.ErrNoRows {
		return output, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, nil
}

func GetLastRoundBeforeTime(q dbcl.Querier, chain string, timestamp time.Time) (*Round, error) {
	const query = `SELECT *
	FROM rounds
	WHERE
		chain_id = ?
	AND
		created_at < ?
	ORDER BY id DESC
	LIMIT 1;`

	output := new(Round)
	err := q.Get(output, query, chain, timestamp)
	if err != nil && err != sql.ErrNoRows {
		return output, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, nil
}

func GetRoundMinTimestamp(q dbcl.Querier, chain string) (time.Time, error) {
	const query = `SELECT MIN(created_at)
	FROM rounds
	WHERE
		chain_id = ?
	AND
		pending IS FALSE;`

	return dbcl.GetTime(q, query, chain)
}

func GetRounds(q dbcl.Querier, page, size uint64) ([]*Round, error) {
	const query = `SELECT *
	FROM rounds
	ORDER BY id DESC
	LIMIT ? OFFSET ?`

	output := []*Round{}
	err := q.Select(&output, query, size, page*size)

	return output, err
}

func GetRoundsCount(q dbcl.Querier) (uint64, error) {
	const query = `SELECT count(id)
	FROM rounds`

	return dbcl.GetUint64(q, query)
}

func GetRoundsByMiner(q dbcl.Querier, minerID, page, size uint64) ([]*Round, error) {
	const query = `SELECT 
		rounds.*, 
		shares.count miner_accepted_shares,
		balance_inputs.value miner_value 
	FROM rounds
	JOIN shares ON rounds.id = shares.round_id
	JOIN balance_inputs ON rounds.id = balance_inputs.round_id AND balance_inputs.miner_id = shares.miner_id
	WHERE
		shares.miner_id = ?
	ORDER BY id DESC
	LIMIT ? OFFSET ?`

	output := []*Round{}
	err := q.Select(&output, query, minerID, size, page*size)

	return output, err
}

func GetRoundsByMinerCount(q dbcl.Querier, minerID uint64) (uint64, error) {
	const query = `SELECT count(rounds.id)
	FROM rounds
	JOIN shares ON rounds.id = shares.round_id
	WHERE
		shares.miner_id = ?;`

	return dbcl.GetUint64(q, query, minerID)
}

func GetRoundsBetweenTime(q dbcl.Querier, chain string, start, end time.Time) ([]*Round, error) {
	const query = `SELECT *
	FROM rounds
	WHERE
		chain_id = ?
	AND
		created_at > ?
	AND
		created_at <= ?`

	output := []*Round{}
	err := q.Select(&output, query, chain, start, end)

	return output, err
}

func GetPendingRoundsByChain(q dbcl.Querier, chain string, height uint64) ([]*Round, error) {
	const query = `SELECT *
	FROM rounds
	WHERE
		chain_id = ?
	AND
		height < ?
	AND
		pending IS TRUE;`

	output := []*Round{}
	err := q.Select(&output, query, chain, height)

	return output, err
}

func GetPendingRoundCountBetweenTime(q dbcl.Querier, chain string, start, end time.Time) (uint64, error) {
	const query = `SELECT COUNT(id)
	FROM rounds
	WHERE
		chain_id = ?
	AND
		created_at > ?
	AND
		created_at <= ?
	AND
		pending IS TRUE`

	return dbcl.GetUint64(q, query, chain, start, end)
}

func GetImmatureRoundsByChain(q dbcl.Querier, chain string, height uint64) ([]*Round, error) {
	const query = `SELECT *
	FROM rounds
	WHERE
		pending IS FALSE
	AND
		mature IS FALSE
	AND
		orphan IS FALSE
	AND
		chain_id = ?
	AND
		height < ?`

	output := []*Round{}
	err := q.Select(&output, query, chain, height)

	return output, err
}

func GetMatureUnspentRounds(q dbcl.Querier, chain string) ([]*Round, error) {
	const query = `SELECT *
	FROM rounds
	WHERE
		pending IS FALSE
	AND
		mature IS TRUE
	AND
		spent IS FALSE
	AND
		orphan IS FALSE
	AND
		chain_id = ?`

	output := []*Round{}
	err := q.Select(&output, query, chain)

	return output, err
}

func GetSumImmatureRoundValueByChain(q dbcl.Querier, chain string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM rounds
	WHERE
		chain_id = ?
	AND
		mature IS FALSE
	AND
		orphan IS FALSE;`

	return dbcl.GetBigInt(q, query, chain)
}

func GetSumUnspentRoundValueByChain(q dbcl.Querier, chain string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM rounds
	WHERE
		chain_id = ?
	AND
		spent IS FALSE
	AND
		orphan IS FALSE;`

	return dbcl.GetBigInt(q, query, chain)
}

/* Share Queries */

func GetSharesByRound(q dbcl.Querier, roundID uint64) ([]*Share, error) {
	const query = `SELECT *
	FROM shares
	WHERE
		round_id = ?`

	output := []*Share{}
	err := q.Select(&output, query, roundID)

	return output, err
}

/* utxos */

func GetUnspentUTXOsByChain(q dbcl.Querier, chainID string) ([]*UTXO, error) {
	const query = `SELECT *
	FROM utxos
	WHERE
		chain_id = ?
	AND
		spent = FALSE;`

	output := []*UTXO{}
	err := q.Select(&output, query, chainID)

	return output, err
}

func GetSumUnspentUTXOValueByChain(q dbcl.Querier, chainID string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM utxos
	WHERE
		chain_id = ?
	AND
		spent = FALSE;`

	return dbcl.GetBigInt(q, query, chainID)
}

/* batch queries */

func GetExchangeBatch(q dbcl.Querier, batchID uint64) (*ExchangeBatch, error) {
	const query = `SELECT *
	FROM exchange_batches
	WHERE
		id = ?;`

	output := new(ExchangeBatch)
	err := q.Get(output, query, batchID)
	if err != nil && err != sql.ErrNoRows {
		return output, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, nil
}

func GetActiveExchangeBatches(q dbcl.Querier) ([]*ExchangeBatch, error) {
	const query = `SELECT *
	FROM exchange_batches
	WHERE
		completed_at IS NULL`

	output := []*ExchangeBatch{}
	err := q.Select(&output, query)

	return output, err
}

func GetExchangeInputs(q dbcl.Querier, batchID uint64) ([]*ExchangeInput, error) {
	const query = `SELECT *
	FROM exchange_inputs
	WHERE
		batch_id = ?;`

	output := []*ExchangeInput{}
	err := q.Select(&output, query, batchID)

	return output, err
}

func GetExchangeDeposits(q dbcl.Querier, batchID uint64) ([]*ExchangeDeposit, error) {
	const query = `SELECT *
	FROM exchange_deposits
	WHERE
		batch_id = ?;`

	output := []*ExchangeDeposit{}
	err := q.Select(&output, query, batchID)

	return output, err
}

func GetExchangeTradesByStage(q dbcl.Querier, batchID uint64, stage int) ([]*ExchangeTrade, error) {
	const query = `SELECT *
	FROM exchange_trades
	WHERE
		batch_id = ?
	AND
		stage_id = ?;`

	output := []*ExchangeTrade{}
	err := q.Select(&output, query, batchID, stage)

	return output, err
}

func GetExchangeTradeByPathAndStage(q dbcl.Querier, batchID uint64, path, stage int) (*ExchangeTrade, error) {
	const query = `SELECT *
	FROM exchange_trades
	WHERE
		batch_id = ?
	AND
		path_id = ?
	AND
		stage_id = ?;`

	output := new(ExchangeTrade)
	err := q.Get(output, query, batchID, path, stage)
	if err != nil && err != sql.ErrNoRows {
		return output, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, nil
}

func GetFinalExchangeTrades(q dbcl.Querier, batchID uint64) ([]*ExchangeTrade, error) {
	const query = `WITH cte AS (
		SELECT path_id, max(stage_id) max_stage
		FROM exchange_trades
		WHERE
			batch_id = ?
		GROUP BY path_id
	)
	SELECT exchange_trades.*
	FROM exchange_trades
	JOIN cte ON exchange_trades.stage_id = cte.max_stage AND exchange_trades.path_id = cte.path_id
	WHERE
		batch_id = ?;`

	output := []*ExchangeTrade{}
	err := q.Select(&output, query, batchID, batchID)

	return output, err
}

func GetExchangeWithdrawals(q dbcl.Querier, batchID uint64) ([]*ExchangeWithdrawal, error) {
	const query = `SELECT *
	FROM exchange_withdrawals
	WHERE
		batch_id = ?;`

	output := []*ExchangeWithdrawal{}
	err := q.Select(&output, query, batchID)

	return output, err
}

/* balance queries */

func GetPendingBalanceInputsWithoutBatch(q dbcl.Querier) ([]*BalanceInput, error) {
	const query = `SELECT *
	FROM balance_inputs
	WHERE
		pending = TRUE
	AND
		batch_id IS NULL;`

	output := []*BalanceInput{}
	err := q.Select(&output, query)

	return output, err
}

func GetBalanceInputsByBatch(q dbcl.Querier, batchID uint64) ([]*BalanceInput, error) {
	const query = `SELECT *
	FROM balance_inputs
	WHERE
		batch_id = ?;`

	output := []*BalanceInput{}
	err := q.Select(&output, query, batchID)

	return output, err
}

func GetSumBalanceInputValueByChain(q dbcl.Querier, chain string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM balance_inputs
	WHERE
		chain_id = ?
	AND
		pending = TRUE;`

	return dbcl.GetBigInt(q, query, chain)
}

func GetBalanceOutputsByBatch(q dbcl.Querier, batchID uint64) ([]*BalanceOutput, error) {
	const query = `SELECT *
	FROM balance_outputs
	WHERE
		in_batch_id = ?;`

	output := []*BalanceOutput{}
	err := q.Select(&output, query, batchID)

	return output, err
}

func GetSumBalanceOutputValueByChain(q dbcl.Querier, chain string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM balance_outputs
	WHERE
		chain_id = ?
	AND
		out_payout_id IS NULL;`

	return dbcl.GetBigInt(q, query, chain)
}

func GetSumBalanceOutputValueByMiner(q dbcl.Querier, minerID uint64, chain string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM balance_outputs
	WHERE
		miner_id = ?
	AND
		chain_id = ?
	AND
		out_payout_id IS NULL;`

	return dbcl.GetBigInt(q, query, minerID, chain)
}

func GetSumBalanceOutputAboveThreshold(q dbcl.Querier, chain, threshold string) ([]*BalanceOutput, error) {
	const query = `WITH cte as (
		SELECT
			miner_id, 
			sum(value) value, 
			sum(pool_fees) pool_fees, 
			sum(exchange_fees) exchange_fees
		FROM balance_outputs
		WHERE
			chain_id = ?
		AND
			out_payout_id IS NULL
		GROUP BY miner_id
	)
	SELECT DISTINCT *
	FROM cte
	WHERE value >= ?;`

	output := []*BalanceOutput{}
	err := q.Select(&output, query, chain, threshold)

	return output, err
}

/* payout */

func GetUnconfirmedPayouts(q dbcl.Querier, chain string) ([]*Payout, error) {
	const query = `SELECT *
	FROM payouts
	WHERE
		chain_id = ?
	AND
		confirmed = FALSE
	and
		failed = FALSE;`

	output := []*Payout{}
	err := q.Select(&output, query, chain)

	return output, err
}

func GetPayouts(q dbcl.Querier, page, size uint64) ([]*Payout, error) {
	const query = `SELECT *
	FROM payouts
	ORDER BY id DESC
	LIMIT ? OFFSET ?`

	output := []*Payout{}
	err := q.Select(&output, query, size, page*size)

	return output, err
}

func GetPayoutsCount(q dbcl.Querier) (uint64, error) {
	const query = `SELECT COUNT(id)
	FROM payouts`

	return dbcl.GetUint64(q, query)
}

func GetPayoutsByMiner(q dbcl.Querier, minerID, page, size uint64) ([]*Payout, error) {
	const query = `SELECT *
	FROM payouts
	WHERE
		miner_id = ?
	ORDER BY id DESC
	LIMIT ? OFFSET ?`

	output := []*Payout{}
	err := q.Select(&output, query, minerID, size, page*size)

	return output, err
}

func GetPayoutsByMinerCount(q dbcl.Querier, minerID uint64) (uint64, error) {
	const query = `SELECT COUNT(id)
	FROM payouts
	WHERE
		miner_id = ?;`

	return dbcl.GetUint64(q, query, minerID)
}

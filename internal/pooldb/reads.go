package pooldb

import (
	"database/sql"
	"fmt"
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
		enabled = TRUE;`

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
		enabled = TRUE;`

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
		backup = TRUE
	AND
		enabled = TRUE;`

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
		pending_backup = TRUE
	AND
		enabled = TRUE;`

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
		pending_update = TRUE
	AND
		enabled = TRUE;`

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
		pending_resize = TRUE
	AND
		enabled = TRUE;`

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

func GetMinerID(q dbcl.Querier, chain, address string) (uint64, error) {
	const query = `SELECT id
	FROM miners
	WHERE
		chain_id = ?
	AND
		address = ?;`

	return dbcl.GetUint64(q, query, chain, address)
}

func GetMinerAddress(q dbcl.Querier, minerID uint64) (string, error) {
	const query = `SELECT address
	FROM miners
	WHERE
		id = ?;`

	return dbcl.GetString(q, query, minerID)
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

func GetWorker(q dbcl.Querier, id uint64) (*Worker, error) {
	const query = `SELECT *
	FROM workers
	WHERE
		id = ?`

	output := new(Worker)
	err := q.Get(output, query, id)
	if err != nil && err != sql.ErrNoRows {
		return output, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, nil
}

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
		workers.id,
		workers.name,
		MAX(ip_addresses.active) active,
		MIN(workers.created_at) created_at,
		MAX(ip_addresses.last_share) last_share,
		MAX(ip_addresses.last_difficulty) last_difficulty
	FROM workers
	JOIN ip_addresses ON workers.id = ip_addresses.worker_id
	WHERE
		workers.miner_id = ?
	AND
		ip_addresses.expired = FALSE
	GROUP BY workers.id;`

	output := []*Worker{}
	err := q.Select(&output, query, minerID)

	return output, err
}

/* ip addresses */

func GetMinersWithLastShares(q dbcl.Querier, minerIDs []uint64) ([]*Miner, error) {
	const rawQuery = `WITH cte AS (
		SELECT
			miner_id,
			MAX(last_share) last_share
		FROM ip_addresses
		GROUP BY miner_id
	) SELECT
		miners.id,
		miners.chain_id,
		miners.address,
		miners.created_at,
		cte.last_share
	FROM miners
	JOIN cte ON miners.id = cte.miner_id
	WHERE
		miners.id IN (?)`

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

func GetActiveWorkersByMinersCount(q dbcl.Querier, minerIDs []uint64) (uint64, error) {
	const rawQuery = `WITH cte AS (
	    SELECT
	        worker_id,
	        MAX(active) AS active
	    FROM
	        ip_addresses
	    WHERE
	        miner_id IN (?)
	    AND
	        worker_id != 0
	    AND
	        expired = FALSE
	    GROUP BY worker_id
	) SELECT COUNT(DISTINCT worker_id)
	FROM cte
	WHERE
	    active = TRUE;`

	if len(minerIDs) == 0 {
		return 0, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs)
	if err != nil {
		return 0, err
	}
	query = q.Rebind(query)

	return dbcl.GetUint64(q, query, args...)
}

func GetInactiveWorkersByMinersCount(q dbcl.Querier, minerIDs []uint64) (uint64, error) {
	const rawQuery = `WITH cte AS (
	    SELECT
	        worker_id,
	        MAX(active) AS active
	    FROM
	        ip_addresses
	    WHERE
	        miner_id IN (?)
	    AND
	        worker_id != 0
	    AND
	        expired = FALSE
	    GROUP BY worker_id
	) SELECT COUNT(DISTINCT worker_id)
	FROM cte
	WHERE
	    active = FALSE;`

	if len(minerIDs) == 0 {
		return 0, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs)
	if err != nil {
		return 0, err
	}
	query = q.Rebind(query)

	return dbcl.GetUint64(q, query, args...)
}

func GetWorkersWithLastShares(q dbcl.Querier, workerIDs []uint64) ([]*Worker, error) {
	const rawQuery = `SELECT
		workers.id,
		workers.miner_id,
		workers.name,
		MAX(ip_addresses.active) active,
		workers.notified,
		MIN(workers.created_at) created_at,
		MAX(ip_addresses.last_share) last_share
	FROM workers
	JOIN ip_addresses ON workers.id = ip_addresses.worker_id
	WHERE
		workers.id IN (?)
	AND
		ip_addresses.expired = FALSE
	GROUP BY workers.id;`

	if len(workerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, workerIDs)
	if err != nil {
		return nil, err
	}

	output := []*Worker{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

func GetOldestActiveIPAddress(q dbcl.Querier, minerID uint64) (*IPAddress, error) {
	const query = `SELECT *
	FROM ip_addresses
	WHERE
		miner_id = ?
	AND
		active = TRUE
	AND
		expired = FALSE
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

func GetNewestInactiveIPAddress(q dbcl.Querier, minerID uint64) (*IPAddress, error) {
	const query = `SELECT *
	FROM ip_addresses
	WHERE
		miner_id = ?
	AND
		active = FALSE
	ORDER BY last_share DESC
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
		pending = FALSE;`

	return dbcl.GetTime(q, query, chain)
}

func GetRounds(q dbcl.Querier, page, size uint64) ([]*Round, error) {
	const query = `SELECT
		rounds.*,
		CONCAT(miners.chain_id, ":", miners.address) miner
	FROM rounds
	JOIN miners ON rounds.miner_id = miners.id
	ORDER BY rounds.created_at DESC
	LIMIT ? OFFSET ?`

	output := []*Round{}
	err := q.Select(&output, query, size, page*size)

	return output, err
}

func GetRoundsByChain(q dbcl.Querier, chain string, page, size uint64) ([]*Round, error) {
	const query = `SELECT
		rounds.*,
		CONCAT(miners.chain_id, ":", miners.address) miner
	FROM rounds
	JOIN miners ON rounds.miner_id = miners.id
	WHERE
		rounds.chain_id = ?
	ORDER BY rounds.created_at DESC
	LIMIT ? OFFSET ?`

	output := []*Round{}
	err := q.Select(&output, query, chain, size, page*size)

	return output, err
}

func GetRoundsCount(q dbcl.Querier) (uint64, error) {
	const query = `SELECT count(id)
	FROM rounds`

	return dbcl.GetUint64(q, query)
}

func GetRoundsByChainCount(q dbcl.Querier, chain string) (uint64, error) {
	const query = `SELECT count(id)
	FROM rounds
	WHERE
		chain_id = ?`

	return dbcl.GetUint64(q, query, chain)
}

func GetRoundsByMiners(q dbcl.Querier, minerIDs []uint64, page, size uint64) ([]*Round, error) {
	const rawQuery = `SELECT 
		rounds.*, 
		balance_inputs.value miner_value 
	FROM rounds
	JOIN balance_inputs ON rounds.id = balance_inputs.round_id
	WHERE
		balance_inputs.miner_id IN (?)
	ORDER BY rounds.created_at DESC
	LIMIT ? OFFSET ?`

	if len(minerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs, size, page*size)
	if err != nil {
		return nil, err
	}

	output := []*Round{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

func GetRoundsByMinersCount(q dbcl.Querier, minerIDs []uint64) (uint64, error) {
	const rawQuery = `SELECT count(rounds.id)
	FROM rounds
	JOIN balance_inputs ON rounds.id = balance_inputs.round_id
	WHERE
		balance_inputs.miner_id IN (?);`

	if len(minerIDs) == 0 {
		return 0, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs)
	if err != nil {
		return 0, err
	}
	query = q.Rebind(query)

	return dbcl.GetUint64(q, query, args...)
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
		pending = TRUE;`

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
		pending = TRUE`

	return dbcl.GetUint64(q, query, chain, start, end)
}

func GetImmatureRoundsByChain(q dbcl.Querier, chain string, height uint64) ([]*Round, error) {
	const query = `SELECT *
	FROM rounds
	WHERE
		pending = FALSE
	AND
		mature = FALSE
	AND
		orphan = FALSE
	AND
		chain_id = ?
	AND
		height < ?`

	output := []*Round{}
	err := q.Select(&output, query, chain, height)

	return output, err
}

func GetUnspentRounds(q dbcl.Querier, chain string) ([]*Round, error) {
	const query = `SELECT *
	FROM rounds
	WHERE
		pending = FALSE
	AND
		spent = FALSE
	AND
		orphan = FALSE
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
		mature = FALSE
	AND
		orphan = FALSE;`

	return dbcl.GetBigInt(q, query, chain)
}

func GetSumUnspentRoundValueByChain(q dbcl.Querier, chain string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM rounds
	WHERE
		chain_id = ?
	AND
		spent = FALSE
	AND
		orphan = FALSE;`

	return dbcl.GetBigInt(q, query, chain)
}

func GetRoundLuckByChain(q dbcl.Querier, chain string, solo bool, duration time.Duration) (float64, error) {
	var query = fmt.Sprintf(`SELECT SUM(difficulty) / IFNULL(SUM(accepted_shares), 1)
	FROM rounds
	WHERE
		chain_id = ?
	AND
		solo = ?
	AND
		created_at > DATE_SUB(CURRENT_TIMESTAMP, %s);`, dbcl.ConvertDurationToInterval(duration))

	return dbcl.GetFloat64(q, query, chain, solo)
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
		transaction_id IS NULL
	AND
		active = TRUE
	AND
		spent = FALSE;`

	output := []*UTXO{}
	err := q.Select(&output, query, chainID)

	return output, err
}

func GetUTXOsByTransactionID(q dbcl.Querier, transactionID uint64) ([]*UTXO, error) {
	const query = `SELECT *
	FROM utxos
	WHERE
		transaction_id = ?;`

	output := []*UTXO{}
	err := q.Select(&output, query, transactionID)

	return output, err
}

func GetSumUnspentUTXOValueByChain(q dbcl.Querier, chainID string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM utxos
	WHERE
		chain_id = ?
	AND
		active = TRUE
	AND
		spent = FALSE;`

	return dbcl.GetBigInt(q, query, chainID)
}

/* transactions */

func GetTransaction(q dbcl.Querier, id uint64) (*Transaction, error) {
	const query = `SELECT *
	FROM transactions
	WHERE
		id = ?;`

	output := new(Transaction)
	err := q.Get(output, query, id)
	if err != nil && err != sql.ErrNoRows {
		return output, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, nil
}

func GetTransactionByTxID(q dbcl.Querier, txid string) (*Transaction, error) {
	const query = `SELECT *
	FROM transactions
	WHERE
		txid = ?;`

	output := new(Transaction)
	err := q.Get(output, query, txid)
	if err != nil && err != sql.ErrNoRows {
		return output, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, nil
}

func GetUnspentTransactions(q dbcl.Querier, chainID string) ([]*Transaction, error) {
	const query = `SELECT *
	FROM transactions
	WHERE
		chain_id = ?
	AND
		spent = FALSE;`

	output := []*Transaction{}
	err := q.Select(&output, query, chainID)

	return output, err
}

func GetUnspentTransactionCount(q dbcl.Querier, chainID string) (uint64, error) {
	const query = `SELECT COUNT(id)
	FROM transactions
	WHERE
		chain_id = ?
	AND
		spent = FALSE;`

	return dbcl.GetUint64(q, query, chainID)
}

func GetUnconfirmedTransactions(q dbcl.Querier, chainID string) ([]*Transaction, error) {
	const query = `SELECT *
	FROM transactions
	WHERE
		chain_id = ?
	AND
		spent = TRUE
	AND
		confirmed = FALSE;`

	output := []*Transaction{}
	err := q.Select(&output, query, chainID)

	return output, err
}

func GetUnconfirmedTransactionSum(q dbcl.Querier, chainID string) (*big.Int, error) {
	const query = `SELECT SUM(value) + SUM(fee) value
	FROM transactions
	WHERE
		chain_id = ?
	AND
		spent = TRUE
	AND
		confirmed = FALSE;`

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

func GetActiveExchangeBatches(q dbcl.Querier, exchangeID uint64) ([]*ExchangeBatch, error) {
	const query = `SELECT *
	FROM exchange_batches
	WHERE
		completed_at IS NULL
	AND
		exchange_id = ?;`

	output := []*ExchangeBatch{}
	err := q.Select(&output, query, exchangeID)

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

func GetUnregisteredExchangeDepositsByChain(q dbcl.Querier, chain string) ([]*ExchangeDeposit, error) {
	const query = `SELECT *
	FROM exchange_deposits
	WHERE
		chain_id = ?
	AND
		registered = FALSE;`

	output := []*ExchangeDeposit{}
	err := q.Select(&output, query, chain)

	return output, err
}

func GetExchangeTrades(q dbcl.Querier, batchID uint64) ([]*ExchangeTrade, error) {
	const query = `SELECT *
	FROM exchange_trades
	WHERE
		batch_id = ?;`

	output := []*ExchangeTrade{}
	err := q.Select(&output, query, batchID)

	return output, err
}

func GetExchangeTradesByStage(q dbcl.Querier, batchID uint64, stage int) ([]*ExchangeTrade, error) {
	const query = `WITH cte AS (
	    SELECT
	        id,
	        ROW_NUMBER() OVER (
	            PARTITION BY path_id
	            ORDER BY step_id DESC
	        ) AS rn
	    FROM exchange_trades
	    WHERE
	        batch_id = ?
	    AND
	        stage_id = ?
	) SELECT exchange_trades.*
	FROM exchange_trades
	JOIN cte on cte.id = exchange_trades.id
	WHERE cte.rn = 1;`

	output := []*ExchangeTrade{}
	err := q.Select(&output, query, batchID, stage)

	return output, err
}

func GetExchangeTradeByPathAndStage(q dbcl.Querier, batchID uint64, path, stage int) (*ExchangeTrade, error) {
	const query = `SELECT
	    MIN(exchange_trades.id) as id,
	    exchange_trades.batch_id,
	    exchange_trades.path_id,
	    exchange_trades.stage_id,
	    exchange_trades.initial_chain_id,
	    exchange_trades.from_chain_id,
	    exchange_trades.to_chain_id,
	    exchange_trades.direction,
	    SUM(exchange_trades.value) AS value,
	    SUM(exchange_trades.proceeds) AS proceeds,
	    SUM(exchange_trades.trade_fees) AS trade_fees,
	    SUM(exchange_trades.cumulative_deposit_fees) AS cumulative_deposit_fees,
	    SUM(exchange_trades.cumulative_trade_fees) AS cumulative_trade_fees,
	    SUM(order_price * proceeds) / SUM(proceeds) as order_price,
	    SUM(fill_price * proceeds) / SUM(proceeds) as fill_price,
	    SUM(cumulative_fill_price * proceeds) / SUM(proceeds) as cumulative_fill_price,
	    SUM(slippage * proceeds) / SUM(proceeds) as slippage,
	    MIN(initiated) as initiated,
	    MIN(confirmed) as confirmed,
	    MIN(created_at) as created_at,
	    MAX(updated_at) as updated_at
	FROM exchange_trades
	WHERE
		batch_id = ?
	AND
		path_id = ?
	AND
		stage_id = ?
	GROUP by batch_id, path_id, stage_id, initial_chain_id, from_chain_id, to_chain_id, direction;`

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
	SELECT
	    exchange_trades.batch_id,
	    exchange_trades.path_id,
	    exchange_trades.stage_id,
	    exchange_trades.initial_chain_id,
	    exchange_trades.from_chain_id,
	    exchange_trades.to_chain_id,
	    exchange_trades.direction,
	    SUM(exchange_trades.value) AS value,
	    SUM(exchange_trades.proceeds) AS proceeds,
	    SUM(exchange_trades.trade_fees) AS trade_fees,
	    SUM(exchange_trades.cumulative_deposit_fees) AS cumulative_deposit_fees,
	    SUM(exchange_trades.cumulative_trade_fees) AS cumulative_trade_fees,
	    SUM(order_price * proceeds) / SUM(proceeds) as order_price,
	    SUM(fill_price * proceeds) / SUM(proceeds) as fill_price,
	    SUM(cumulative_fill_price) as cumulative_fill_price,
	    SUM(slippage * proceeds) / SUM(proceeds) as slippage,
	    MIN(initiated) as initiated,
	    MIN(confirmed) as confirmed,
	    MIN(created_at) as created_at,
	    MAX(updated_at) as updated_at
	FROM exchange_trades
	JOIN cte ON exchange_trades.stage_id = cte.max_stage AND exchange_trades.path_id = cte.path_id
	WHERE
		batch_id = ?
	GROUP by batch_id, path_id, stage_id, initial_chain_id, from_chain_id, to_chain_id, direction;`

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
		mature = TRUE
	AND
		batch_id IS NULL;`

	output := []*BalanceInput{}
	err := q.Select(&output, query)

	return output, err
}

func GetPendingBalanceInputsSumWithoutBatch(q dbcl.Querier) ([]*BalanceInput, error) {
	const query = `SELECT
		chain_id,
		out_chain_id,
		SUM(value) as value
	FROM balance_inputs
	WHERE
		pending = TRUE
	AND
		mature = TRUE
	AND
		batch_id IS NULL
	GROUP BY chain_id, out_chain_id;`

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

func GetBalanceInputsByRound(q dbcl.Querier, roundID uint64) ([]*BalanceInput, error) {
	const query = `SELECT *
	FROM balance_inputs
	WHERE
		round_id = ?;`

	output := []*BalanceInput{}
	err := q.Select(&output, query, roundID)

	return output, err
}

func GetImmatureBalanceInputSumByChain(q dbcl.Querier, chain string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM balance_inputs
	WHERE
		chain_id = ?
	AND
		mature = FALSE;`

	return dbcl.GetBigInt(q, query, chain)
}

func GetPendingBalanceInputSumByChain(q dbcl.Querier, chain string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM balance_inputs
	WHERE
		chain_id = ?
	AND
		pending = TRUE
	AND
		mature = TRUE;`

	return dbcl.GetBigInt(q, query, chain)
}

func GetBalanceInputMinTimestamp(q dbcl.Querier, chain string) (time.Time, error) {
	const query = `SELECT MIN(created_at)
	FROM balance_inputs
	WHERE
		chain_id = ?;`

	return dbcl.GetTime(q, query, chain)
}

func GetBalanceInputSumFromRange(q dbcl.Querier, chain string, startTime, endTime time.Time) ([]*BalanceInput, error) {
	const query = `SELECT
		miner_id,
		sum(value) value
	FROM balance_inputs
	WHERE
		chain_id = ?
	AND
		created_at BETWEEN ? AND ?
	GROUP BY miner_id;`

	output := []*BalanceInput{}
	err := q.Select(&output, query, chain, startTime, endTime)

	return output, err
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

func GetBalanceOutputsByPayoutTransaction(q dbcl.Querier, transactionID uint64) ([]*BalanceOutput, error) {
	const query = `SELECT balance_outputs.*
	FROM balance_outputs
	JOIN payouts ON payouts.id = balance_outputs.out_payout_id
	WHERE
		payouts.transaction_id = ?
	AND
		out_merge_transaction_id IS NULL;`

	output := []*BalanceOutput{}
	err := q.Select(&output, query, transactionID)

	return output, err
}

func GetBalanceOutputsByPayout(q dbcl.Querier, payoutID uint64) ([]*BalanceOutput, error) {
	const query = `SELECT *
	FROM balance_outputs
	WHERE
		out_payout_id = ?;`

	output := []*BalanceOutput{}
	err := q.Select(&output, query, payoutID)

	return output, err
}

func GetRandomBalanceOutputAboveValue(q dbcl.Querier, chain, value string) (*BalanceOutput, error) {
	const query = `SELECT *
	FROM balance_outputs
	WHERE
		chain_id = ?
	AND
		value > ? * 10
	AND
		mature = TRUE
	AND
		spent = FALSE
	ORDER BY id DESC
	LIMIT 1;`

	output := new(BalanceOutput)
	err := q.Get(output, query, chain, value)
	if err != nil && err != sql.ErrNoRows {
		return output, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return output, nil
}

func GetUnpaidBalanceOutputsByMiner(q dbcl.Querier, minerID uint64, chain string) ([]*BalanceOutput, error) {
	const query = `SELECT *
	FROM balance_outputs
	WHERE
		miner_id = ?
	AND
		chain_id = ?
	AND
		mature = TRUE
	AND
		out_payout_id IS NULL;`

	output := []*BalanceOutput{}
	err := q.Select(&output, query, minerID, chain)

	return output, err
}

func GetUnpaidBalanceOutputSumByChain(q dbcl.Querier, chain string) (*big.Int, error) {
	const query = `SELECT sum(value)
	FROM balance_outputs
	WHERE
		chain_id = ?
	AND
		mature = TRUE
	AND
		spent = FALSE;`

	return dbcl.GetBigInt(q, query, chain)
}

func GetMinersWithBalanceAboveThresholdByChain(q dbcl.Querier, chain, threshold string) ([]*Miner, error) {
	const query = `SELECT DISTINCT miners.*
	FROM miners
	JOIN balance_sums ON
	        miners.id = balance_sums.miner_id
	    AND
	        miners.chain_id = balance_sums.chain_id
	LEFT OUTER JOIN payouts on
	        miners.id = payouts.miner_id
	    AND
	        payouts.confirmed = FALSE
	WHERE
		miners.chain_id = ?
	AND
		balance_sums.mature_value > ?
	AND
	    payouts.id IS NULL;`

	// AND
	//     balance_sums.mature_value >= IFNULL(miners.threshold, ?)

	output := []*Miner{}
	err := q.Select(&output, query, chain, threshold)

	return output, err
}

func GetBalanceSumsByMinerIDs(q dbcl.Querier, minerIDs []uint64) ([]*BalanceSum, error) {
	const rawQuery = `SELECT *
	FROM balance_sums
	WHERE
		miner_id IN (?);`

	if len(minerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs)
	if err != nil {
		return nil, err
	}

	output := []*BalanceSum{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

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

func GetUnconfirmedPayoutSum(q dbcl.Querier, chain string) (*big.Int, error) {
	const query = `SELECT SUM(payouts.value) + SUM(payouts.tx_fees) value
	FROM payouts
	JOIN transactions ON payouts.transaction_id = transactions.id
	WHERE
		payouts.chain_id = ?
	AND
		payouts.confirmed = FALSE
	AND
		transactions.spent = TRUE;`

	return dbcl.GetBigInt(q, query, chain)
}

func GetPayouts(q dbcl.Querier, page, size uint64) ([]*Payout, error) {
	const query = `SELECT *
	FROM payouts
	WHERE
		pending = FALSE
	ORDER BY id DESC
	LIMIT ? OFFSET ?`

	output := []*Payout{}
	err := q.Select(&output, query, size, page*size)

	return output, err
}

func GetPayoutsCount(q dbcl.Querier) (uint64, error) {
	const query = `SELECT COUNT(id)
	FROM payouts
	WHERE
		pending = FALSE`

	return dbcl.GetUint64(q, query)
}

func GetPayoutsByMiners(q dbcl.Querier, minerIDs []uint64, page, size uint64) ([]*Payout, error) {
	const rawQuery = `SELECT *
	FROM payouts
	WHERE
		miner_id IN (?)
	AND
		pending = FALSE
	ORDER BY id DESC
	LIMIT ? OFFSET ?`

	if len(minerIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs, size, page*size)
	if err != nil {
		return nil, err
	}

	output := []*Payout{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

func GetPayoutsByMinersCount(q dbcl.Querier, minerIDs []uint64) (uint64, error) {
	const rawQuery = `SELECT COUNT(id)
	FROM payouts
	WHERE
		miner_id IN (?)
	AND
		pending = FALSE;`

	if len(minerIDs) == 0 {
		return 0, nil
	}

	query, args, err := sqlx.In(rawQuery, minerIDs)
	if err != nil {
		return 0, err
	}
	query = q.Rebind(query)

	return dbcl.GetUint64(q, query, args...)
}

func GetPayoutBalanceInputSums(q dbcl.Querier, payoutIDs []uint64) ([]*BalanceInput, error) {
	const rawQuery = `SELECT
	    balance_inputs.chain_id,
	    balance_outputs.out_payout_id AS payout_id,
	    sum(balance_inputs.value) AS value
	FROM balance_inputs
	JOIN balance_outputs ON balance_inputs.balance_output_id = balance_outputs.id
	WHERE
	    balance_outputs.out_payout_id IN (?)
	GROUP BY balance_inputs.chain_id, balance_outputs.out_payout_id`

	if len(payoutIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(rawQuery, payoutIDs)
	if err != nil {
		return nil, err
	}

	output := []*BalanceInput{}
	query = q.Rebind(query)
	err = q.Select(&output, query, args...)

	return output, err
}

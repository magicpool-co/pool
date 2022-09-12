package pooldb

import (
	"database/sql"
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

func GetWorkersByMinerID(q dbcl.Querier, minerID uint64) ([]*Worker, error) {
	const query = `SELECT *
	FROM workers
	WHERE
		miner_id = ?;`

	output := []*Worker{}
	err := q.Select(&output, query, minerID)

	return output, err
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

/* ip addresses */

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

func GetRounds(q dbcl.Querier, page, size int) ([]*Round, error) {
	const query = `SELECT *
	FROM rounds
	ORDER BY id DESC
	LIMIT ? OFFSET ?`

	output := []*Round{}
	err := q.Select(&output, query, size+1, page*size)

	return output, err
}

func GetRoundsCount(q dbcl.Querier) (uint64, error) {
	const query = `SELECT count(id)
	FROM rounds`

	return dbcl.GetUint64(q, query)
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

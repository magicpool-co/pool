package pooldb

import (
	"fmt"
	"time"

	"github.com/magicpool-co/pool/pkg/dbcl"
)

func InsertNode(q dbcl.Querier, obj *Node) (uint64, error) {
	const table = "nodes"
	cols := []string{"chain_id", "region", "url", "version", "mainnet", "enabled", "backup",
		"active", "synced", "height", "down_at", "backup_at"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateNode(q dbcl.Querier, obj *Node, updateCols []string) error {
	const table = "nodes"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertMiner(q dbcl.Querier, obj *Miner) (uint64, error) {
	const table = "miners"
	cols := []string{"chain_id", "address", "active", "last_login", "last_share"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateMiner(q dbcl.Querier, obj *Miner, updateCols []string) error {
	const table = "miners"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertWorker(q dbcl.Querier, obj *Worker) (uint64, error) {
	const table = "workers"
	cols := []string{"miner_id", "name", "active", "last_login", "last_share"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateWorker(q dbcl.Querier, obj *Worker, updateCols []string) error {
	const table = "workers"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertIPAddresses(q dbcl.Querier, objects ...*IPAddress) error {
	const table = "ip_addresses"
	insertCols := []string{"miner_id", "ip_address", "active", "last_share"}
	updateCols := []string{"active", "last_share"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsertUpdateOverwrite(q, table, insertCols, updateCols, rawObjects)
}

func UpdateIPAddressesSetInactive(q dbcl.Querier, duration time.Duration) error {
	var query = fmt.Sprintf(`UPDATE ip_addresses
	SET active = FALSE
	WHERE
		last_share < DATE_SUB(CURRENT_TIMESTAMP, %s);`, dbcl.ConvertDurationToInterval(duration))

	_, err := q.Exec(query)

	return err
}

func InsertRound(q dbcl.Querier, obj *Round) (uint64, error) {
	const table = "rounds"
	cols := []string{"chain_id", "miner_id", "worker_id", "height", "epoch_height", "uncle_height", "hash",
		"nonce", "mix_digest", "coinbase_txid", "value", "difficulty", "luck", "accepted_shares", "rejected_shares",
		"invalid_shares", "mature", "pending", "uncle", "orphan", "spent"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateRound(q dbcl.Querier, obj *Round, updateCols []string) error {
	const table = "rounds"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertShare(q dbcl.Querier, obj *Share) (uint64, error) {
	const table = "shares"
	cols := []string{"round_id", "miner_id", "worker_id", "count"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func InsertShares(q dbcl.Querier, objects ...*Share) error {
	const table = "shares"
	cols := []string{"round_id", "miner_id", "worker_id", "count"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

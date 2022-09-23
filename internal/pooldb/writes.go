package pooldb

import (
	"fmt"
	"time"

	"github.com/magicpool-co/pool/pkg/dbcl"
)

func InsertNode(q dbcl.Querier, obj *Node) (uint64, error) {
	const table = "nodes"
	cols := []string{"url", "chain_id", "region", "version", "mainnet", "enabled", "backup",
		"active", "synced", "height", "down_at", "backup_at"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateNode(q dbcl.Querier, obj *Node, updateCols []string) error {
	const table = "nodes"
	whereCols := []string{"url"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertMiner(q dbcl.Querier, obj *Miner) (uint64, error) {
	const table = "miners"
	cols := []string{"chain_id", "address", "active"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateMiner(q dbcl.Querier, obj *Miner, updateCols []string) error {
	const table = "miners"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertWorker(q dbcl.Querier, obj *Worker) (uint64, error) {
	const table = "workers"
	cols := []string{"miner_id", "name", "active"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateWorker(q dbcl.Querier, obj *Worker, updateCols []string) error {
	const table = "workers"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertIPAddresses(q dbcl.Querier, objects ...*IPAddress) error {
	const table = "ip_addresses"
	insertCols := []string{"miner_id", "worker_id", "chain_id", "ip_address", "active", "last_share"}
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
	cols := []string{"chain_id", "miner_id", "height", "epoch_height", "uncle_height", "hash",
		"nonce", "mix_digest", "coinbase_txid", "value", "difficulty", "luck", "accepted_shares", "rejected_shares",
		"invalid_shares", "mature", "pending", "uncle", "orphan", "spent"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateRound(q dbcl.Querier, obj *Round, updateCols []string) error {
	const table = "rounds"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertShares(q dbcl.Querier, objects ...*Share) error {
	const table = "shares"
	cols := []string{"round_id", "miner_id", "count"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

/* exchange batches */

func InsertExchangeBatch(q dbcl.Querier, obj *ExchangeBatch) (uint64, error) {
	const table = "exchange_batches"
	cols := []string{"exchange_id", "status"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateExchangeBatch(q dbcl.Querier, obj *ExchangeBatch, updateCols []string) error {
	const table = "exchange_batches"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertExchangeInputs(q dbcl.Querier, objects ...*ExchangeInput) error {
	const table = "exchange_batches"
	cols := []string{"batch_id", "input_chain_id", "output_chain_id", "value"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

func InsertExchangeDeposit(q dbcl.Querier, obj *ExchangeDeposit) (uint64, error) {
	const table = "exchange_deposits"
	cols := []string{"batch_id", "chain_id", "network_id", "deposit_txid", "exchange_txid",
		"exchange_deposit_id", "value", "fees", "registered", "pending", "spent"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateExchangeDeposit(q dbcl.Querier, obj *ExchangeDeposit, updateCols []string) error {
	const table = "exchange_deposits"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertExchangeTrades(q dbcl.Querier, objects ...*ExchangeTrade) error {
	const table = "exchange_trades"
	cols := []string{"batch_id", "path_id", "stage", "exchange_trade_id", "next_trade_id",
		"from_chain_id", "to_chain_id", "market", "direction", "increment", "value", "remainder",
		"fees", "proceeds", "slippage", "initiated", "open", "filled"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

func UpdateExchangeTrade(q dbcl.Querier, obj *ExchangeTrade, updateCols []string) error {
	const table = "exchange_trades"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

func InsertExchangeWithdrawal(q dbcl.Querier, obj *ExchangeWithdrawal) (uint64, error) {
	const table = "exchange_withdrawals"
	cols := []string{"batch_id", "chain_id", "network_id", "exchange_txid", "exchange_withdrawal_id",
		"value", "trade_fees", "withdrawal_fees", "pending", "spent"}

	return dbcl.ExecInsert(q, table, cols, obj)
}

func UpdateExchangeWithdrawal(q dbcl.Querier, obj *ExchangeWithdrawal, updateCols []string) error {
	const table = "exchange_withdrawals"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

/* balances */

func InsertBalanceInputs(q dbcl.Querier, objects ...*BalanceInput) error {
	const table = "balance_inputs"
	cols := []string{"round_id", "chain_id", "miner_id", "exchange_batch_id", "output_balance_id", "value", "pending"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

/* utxos */

func InsertUTXOs(q dbcl.Querier, objects ...*UTXO) error {
	const table = "utxos"
	cols := []string{"chain_id", "value", "txid", "index", "spent"}

	rawObjects := make([]interface{}, len(objects))
	for i, object := range objects {
		rawObjects[i] = object
	}

	return dbcl.ExecBulkInsert(q, table, cols, rawObjects)
}

func UpdateUTXO(q dbcl.Querier, obj *UTXO, updateCols []string) error {
	const table = "utxos"
	whereCols := []string{"id"}

	return dbcl.ExecUpdate(q, table, updateCols, whereCols, true, obj)
}

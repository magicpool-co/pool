package dbcl

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math/big"

	"github.com/jmoiron/sqlx"
)

type Querier interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
}

type Tx struct {
	sqlx.Tx
	commited bool
}

func NewTx(tx *sqlx.Tx) *Tx {
	if tx == nil {
		return nil
	}

	return &Tx{Tx: *tx}
}

func (tx *Tx) SafeCommit() error {
	err := tx.Commit()
	if err == nil {
		tx.commited = true
	}

	return err
}

func (tx *Tx) SafeRollback() error {
	if tx.commited {
		return nil
	}

	return tx.Rollback()
}

type NullBigInt struct {
	BigInt *big.Int
	Valid  bool
}

// Scan implements the Scanner interface.
func (n *NullBigInt) Scan(src interface{}) error {
	if src == nil {
		n.BigInt, n.Valid = nil, false
		return nil
	}

	var source string
	switch v := src.(type) {
	case string:
		source = v
	case []byte:
		source = string(v)
	default:
		return fmt.Errorf("incompatible type for NullBigInt")
	}

	n.BigInt, n.Valid = new(big.Int).SetString(source, 10)

	return nil
}

// Value implements the driver Valuer interface.
func (n NullBigInt) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.BigInt.String(), nil
}

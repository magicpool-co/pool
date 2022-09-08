package dbcl

import (
	"database/sql"
	"math/big"
	"time"
)

func GetUint64(q Querier, query string, args ...interface{}) (uint64, error) {
	var output *uint64
	err := q.Get(&output, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	} else if output == nil {
		return 0, nil
	}

	return *output, nil
}

func GetFloat64(q Querier, query string, args ...interface{}) (float64, error) {
	var output *float64
	err := q.Get(&output, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	} else if output == nil {
		return 0, nil
	}

	return *output, nil
}

func GetTime(q Querier, query string, args ...interface{}) (time.Time, error) {
	var output *time.Time
	err := q.Get(&output, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return time.Time{}, err
	} else if output == nil {
		return time.Time{}, nil
	}

	return *output, nil
}

func GetBigInt(q Querier, query string, args ...interface{}) (*big.Int, error) {
	output := new(NullBigInt)
	err := q.Get(output, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if output.Valid {
		return output.BigInt, nil
	}

	return new(big.Int), nil
}

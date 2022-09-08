package tsdb

import (
	_ "embed"
	"fmt"
)

//go:embed migrations/000_init.sql
var sql000initUp string

//go:embed migrations/000_init.down.sql
var sql000initDown string

//go:embed migrations/001_tsdb_primary_key.sql
var sql001tsdbPrimaryKeyUp string

//go:embed migrations/001_tsdb_primary_key.down.sql
var sql001tsdbPrimaryKeyDown string

//go:embed migrations/002_invalid_shares.sql
var sql002invalidSharesUp string

//go:embed migrations/002_invalid_shares.down.sql
var sql002invalidSharesDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":                  sql000initUp,
		"000_init.down.sql":             sql000initDown,
		"001_tsdb_primary_key.sql":      sql001tsdbPrimaryKeyUp,
		"001_tsdb_primary_key.down.sql": sql001tsdbPrimaryKeyDown,
		"002_invalid_shares.sql":        sql002invalidSharesUp,
		"002_invalid_shares.down.sql":   sql002invalidSharesDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

package pooldb

import (
	_ "embed"
	"fmt"
)

//go:embed migrations/000_init.sql
var sql000initUp string

//go:embed migrations/000_init.down.sql
var sql000initDown string

//go:embed migrations/001_merge_tx.sql
var sql001mergeTxUp string

//go:embed migrations/001_merge_tx.down.sql
var sql001mergeTxDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":          sql000initUp,
		"000_init.down.sql":     sql000initDown,
		"001_merge_tx.sql":      sql001mergeTxUp,
		"001_merge_tx.down.sql": sql001mergeTxDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

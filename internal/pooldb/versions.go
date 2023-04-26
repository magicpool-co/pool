package pooldb

import (
	_ "embed"
	"fmt"
)

//go:embed migrations/000_init.sql
var sql000initUp string

//go:embed migrations/000_init.down.sql
var sql000initDown string

//go:embed migrations/004_balance_sums.sql
var sql004balanceSumsUp string

//go:embed migrations/004_balance_sums.down.sql
var sql004balanceSumsDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":              sql000initUp,
		"000_init.down.sql":         sql000initDown,
		"004_balance_sums.sql":      sql004balanceSumsUp,
		"004_balance_sums.down.sql": sql004balanceSumsDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

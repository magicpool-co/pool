package pooldb

import (
	_ "embed"
	"fmt"
)

//go:embed migrations/000_init.sql
var sql000initUp string

//go:embed migrations/000_init.down.sql
var sql000initDown string

//go:embed migrations/001_multi_exchange.sql
var sql001multiExchangeUp string

//go:embed migrations/001_multi_exchange.down.sql
var sql001multiExchangeDown string

//go:embed migrations/002_immature_balance.sql
var sql002immatureBalanceUp string

//go:embed migrations/002_immature_balance.down.sql
var sql002immatureBalanceDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":                  sql000initUp,
		"000_init.down.sql":             sql000initDown,
		"001_multi_exchange.sql":        sql001multiExchangeUp,
		"001_multi_exchange.down.sql":   sql001multiExchangeDown,
		"002_immature_balance.sql":      sql002immatureBalanceUp,
		"002_immature_balance.down.sql": sql002immatureBalanceDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

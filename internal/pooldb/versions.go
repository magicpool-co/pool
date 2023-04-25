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

//go:embed migrations/003_miner_email.sql
var sql003minerEmailUp string

//go:embed migrations/003_miner_email.down.sql
var sql003minerEmailDown string

//go:embed migrations/004_miner_notify.sql
var sql004minerNotifyUp string

//go:embed migrations/004_miner_notify.down.sql
var sql004minerNotifyDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":                  sql000initUp,
		"000_init.down.sql":             sql000initDown,
		"001_multi_exchange.sql":        sql001multiExchangeUp,
		"001_multi_exchange.down.sql":   sql001multiExchangeDown,
		"002_immature_balance.sql":      sql002immatureBalanceUp,
		"002_immature_balance.down.sql": sql002immatureBalanceDown,
		"003_miner_email.sql":           sql003minerEmailUp,
		"003_miner_email.down.sql":      sql003minerEmailDown,
		"004_miner_notify.sql":          sql004minerNotifyUp,
		"004_miner_notify.down.sql":     sql004minerNotifyDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

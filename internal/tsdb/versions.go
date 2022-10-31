package tsdb

import (
	_ "embed"
	"fmt"
)

//go:embed migrations/000_init.sql
var sql000initUp string

//go:embed migrations/000_init.down.sql
var sql000initDown string

//go:embed migrations/001_prices.sql
var sql001PricesUp string

//go:embed migrations/001_prices.down.sql
var sql001PricesDown string

//go:embed migrations/002_index.sql
var sql002IndexUp string

//go:embed migrations/002_index.down.sql
var sql002IndexDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":        sql000initUp,
		"000_init.down.sql":   sql000initDown,
		"001_prices.sql":      sql001PricesUp,
		"001_prices.down.sql": sql001PricesDown,
		"002_index.sql":       sql002IndexUp,
		"002_index.down.sql":  sql002IndexDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

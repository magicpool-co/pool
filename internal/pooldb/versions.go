package pooldb

import (
	_ "embed"
	"fmt"
)

//go:embed migrations/000_init.sql
var sql000initUp string

//go:embed migrations/000_init.down.sql
var sql000initDown string

//go:embed migrations/001_kaspa_bootnodes.sql
var sql001kaspaBootnodesUp string

//go:embed migrations/001_kaspa_bootnodes.down.sql
var sql001kaspaBootnodesDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":                 sql000initUp,
		"000_init.down.sql":            sql000initDown,
		"001_kaspa_bootnodes.sql":      sql001kaspaBootnodesUp,
		"001_kaspa_bootnodes.down.sql": sql001kaspaBootnodesDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

package tsdb

import (
	_ "embed"
	"fmt"
)

//go:embed migrations/000_init.sql
var sql000initUp string

//go:embed migrations/000_init.down.sql
var sql000initDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":      sql000initUp,
		"000_init.down.sql": sql000initDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

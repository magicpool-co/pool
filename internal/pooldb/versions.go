package pooldb

import (
	_ "embed"
	"fmt"
)

//go:embed migrations/000_init.sql
var sql000initUp string

//go:embed migrations/000_init.down.sql
var sql000initDown string

//go:embed migrations/001_unique_miners.sql
var sql001uniqueMinersUp string

//go:embed migrations/001_unique_miners.down.sql
var sql001uniqueMinersDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":               sql000initUp,
		"000_init.down.sql":          sql000initDown,
		"001_unique_miners.sql":      sql001uniqueMinersUp,
		"001_unique_miners.down.sql": sql001uniqueMinersDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

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
var sql001pricesUp string

//go:embed migrations/001_prices.down.sql
var sql001pricesDown string

//go:embed migrations/002_index.sql
var sql002indexUp string

//go:embed migrations/002_index.down.sql
var sql002indexDown string

//go:embed migrations/003_raw_blocks_hash.sql
var sql003rawBlocksHashUp string

//go:embed migrations/003_raw_blocks_hash.down.sql
var sql003rawBlocksHashDown string

//go:embed migrations/004_remove_reported_hashrate.sql
var sql004removeReportedHashrateUp string

//go:embed migrations/004_remove_reported_hashrate.down.sql
var sql004removeReportedHashrateDown string

//go:embed migrations/005_adjusted_shares.sql
var sql005adjustedSharesUp string

//go:embed migrations/005_adjusted_shares.down.sql
var sql005adjustedSharesDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":                          sql000initUp,
		"000_init.down.sql":                     sql000initDown,
		"001_prices.sql":                        sql001pricesUp,
		"001_prices.down.sql":                   sql001pricesDown,
		"002_index.sql":                         sql002indexUp,
		"002_index.down.sql":                    sql002indexDown,
		"003_raw_blocks_hash.sql":               sql003rawBlocksHashUp,
		"003_raw_blocks_hash.down.sql":          sql003rawBlocksHashDown,
		"004_remove_reported_hashrate.sql":      sql004removeReportedHashrateUp,
		"004_remove_reported_hashrate.down.sql": sql004removeReportedHashrateDown,
		"005_adjusted_shares.sql":               sql005adjustedSharesUp,
		"005_adjusted_shares.down.sql":          sql005adjustedSharesDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

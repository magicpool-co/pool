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

//go:embed migrations/002_exchange_batch.sql
var sql002ExchangeBatchUp string

//go:embed migrations/002_exchange_batch.down.sql
var sql002ExchangeBatchDown string

//go:embed migrations/003_miner_thresholds.sql
var sql003MinerThresholdsUp string

//go:embed migrations/003_miner_thresholds.down.sql
var sql003MinerThresholdsDown string

//go:embed migrations/004_expired_ip.sql
var sql004ExpiredIpUp string

//go:embed migrations/004_expired_ip.down.sql
var sql004ExpiredIpDown string

func getMigrationVersions() (map[string]string, error) {
	migrations := map[string]string{
		"000_init.sql":                  sql000initUp,
		"000_init.down.sql":             sql000initDown,
		"001_unique_miners.sql":         sql001uniqueMinersUp,
		"001_unique_miners.down.sql":    sql001uniqueMinersDown,
		"002_exchange_batch.sql":        sql002ExchangeBatchUp,
		"002_exchange_batch.down.sql":   sql002ExchangeBatchDown,
		"003_miner_thresholds.sql":      sql003MinerThresholdsUp,
		"003_miner_thresholds.down.sql": sql003MinerThresholdsDown,
		"004_expired_ip.sql":            sql004ExpiredIpUp,
		"004_expired_ip.down.sql":       sql004ExpiredIpDown,
	}

	for k, v := range migrations {
		if len(v) == 0 {
			return migrations, fmt.Errorf("%s is an empty migration", k)
		}
	}

	return migrations, nil
}

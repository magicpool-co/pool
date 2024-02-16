package tsdb

import (
	"embed"
	"fmt"

	"github.com/magicpool-co/pool/pkg/dbcl"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func New(args map[string]string) (*dbcl.Client, error) {
	var argKeys = []string{
		"TSDB_WRITE_HOST",
		"TSDB_READ_HOST",
		"TSDB_PORT",
		"TSDB_USER",
		"TSDB_PASS",
		"TSDB_NAME",
	}
	for _, k := range argKeys {
		if _, ok := args[k]; !ok {
			return nil, fmt.Errorf("%s is a required argument", k)
		}
	}

	writeHost := args["TSDB_WRITE_HOST"]
	readHost := args["TSDB_READ_HOST"]
	port := args["TSDB_PORT"]
	user := args["TSDB_USER"]
	pass := args["TSDB_PASS"]
	name := args["TSDB_NAME"]

	migrations, err := dbcl.FetchMigrations("migrations/*.sql", &migrationFS)
	if err != nil {
		return nil, err
	}

	client, err := dbcl.New(writeHost, readHost, port, name, user, pass, migrations)
	if err != nil {
		return nil, err
	}

	return client, nil
}

package pooldb

import (
	"embed"
	"fmt"

	"github.com/magicpool-co/pool/pkg/dbcl"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func New(args map[string]string) (*dbcl.Client, error) {
	var argKeys = []string{
		"POOLDB_WRITE_HOST",
		"POOLDB_READ_HOST",
		"POOLDB_PORT",
		"POOLDB_USER",
		"POOLDB_PASS",
		"POOLDB_NAME",
	}
	for _, k := range argKeys {
		if _, ok := args[k]; !ok {
			return nil, fmt.Errorf("%s is a required argument", k)
		}
	}

	writeHost := args["POOLDB_WRITE_HOST"]
	readHost := args["POOLDB_READ_HOST"]
	port := args["POOLDB_PORT"]
	user := args["POOLDB_USER"]
	pass := args["POOLDB_PASS"]
	name := args["POOLDB_NAME"]

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

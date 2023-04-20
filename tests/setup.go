//go:build integration

package tests

import (
	"fmt"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/dbcl"
)

const (
	mysqlHost = "localhost"
	mysqlPort = "3644"
	mysqlUser = "root"
	mysqlPass = "secret"

	redisHost = "localhost"
	redisPort = "3645"
	redisDB   = "0"
)

var (
	pooldbArgs = map[string]string{
		"POOLDB_WRITE_HOST": mysqlHost,
		"POOLDB_READ_HOST":  mysqlHost,
		"POOLDB_PORT":       mysqlPort,
		"POOLDB_USER":       mysqlUser,
		"POOLDB_PASS":       mysqlPass,
		"POOLDB_NAME":       "pooldb",
	}

	tsdbArgs = map[string]string{
		"TSDB_WRITE_HOST": mysqlHost,
		"TSDB_READ_HOST":  mysqlHost,
		"TSDB_PORT":       mysqlPort,
		"TSDB_USER":       mysqlUser,
		"TSDB_PASS":       mysqlPass,
		"TSDB_NAME":       "tsdb",
	}

	redisArgs = map[string]string{
		"REDIS_WRITE_HOST": redisHost,
		"REDIS_READ_HOST":  redisHost,
		"REDIS_PORT":       redisPort,
		"REDIS_DB":         redisDB,
	}
)

func newTestPooldb(retryInterval, timeoutInterval time.Duration) (*dbcl.Client, error) {
	var err error
	var client *dbcl.Client
	retry := time.NewTicker(retryInterval)
	quit := time.NewTimer(timeoutInterval)

	// wait for db to initialize
	for {
		select {
		case <-retry.C:
			client, err = pooldb.New(pooldbArgs)
			if err == nil {
				// needs to create tsdb since multiple databases on start is annoying for docker mysql
				client.Writer().Exec("CREATE DATABASE IF NOT EXISTS tsdb;")
				return client, nil
			}
		case <-quit.C:
			retry.Stop()
			return nil, fmt.Errorf("unable to connect to test pooldb")
		}
	}
}

func newTestTsdb(retryInterval, timeoutInterval time.Duration) (*dbcl.Client, error) {
	var err error
	var client *dbcl.Client
	retry := time.NewTicker(retryInterval)
	quit := time.NewTimer(timeoutInterval)

	// wait for db to initialize
	for {
		select {
		case <-retry.C:
			client, err = tsdb.New(tsdbArgs)
			if err == nil {
				return client, nil
			}
		case <-quit.C:
			retry.Stop()
			return nil, fmt.Errorf("unable to connect to tsdb")
		}
	}
}

func newTestRedis(retryInterval, timeoutInterval time.Duration) (*redis.Client, error) {
	var err error
	var client *redis.Client
	retry := time.NewTicker(retryInterval)
	quit := time.NewTimer(timeoutInterval)

	// wait for redis to initialize
	for {
		select {
		case <-retry.C:
			client, err = redis.New(redisArgs)
			if err == nil {
				return client, nil
			}
		case <-quit.C:
			retry.Stop()
			return nil, fmt.Errorf("unable to connect to test redis")
		}
	}
}

//go:build integration

package tests

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/pkg/dbcl"
)

var (
	pooldbClient *dbcl.Client
	tsdbClient   *dbcl.Client
	redisClient  *redis.Client
)

func TestMain(m *testing.M) {
	var err error

	pooldbClient, err = newTestPooldb(time.Second*5, time.Second*20)
	if err != nil {
		log.Fatalf("failed on setup: %v", err)
	}

	tsdbClient, err = newTestTsdb(time.Second*5, time.Second*20)
	if err != nil {
		log.Fatalf("failed on setup: %v", err)
	}

	redisClient, err = newTestRedis(time.Second*1, time.Second*10)
	if err != nil {
		log.Fatalf("failed on setup: %v", err)
	}

	code := m.Run()
	os.Exit(code)
}

/* internal */

func TestPoolDBReads(t *testing.T) {
	if err := pooldbClient.UpgradeMigrations(); err != nil {
		t.Errorf("TestPoolDBReads: failed on upgrade pooldb migrations: %v\n", err)
		return
	}

	suite.Run(t, new(PooldbReadsSuite))

	if err := pooldbClient.DowngradeMigrations(); err != nil {
		t.Errorf("TestPoolDBReads: failed on downgrade pooldb migrations: %v\n", err)
	}
}

func TestPoolDBWrites(t *testing.T) {
	if err := pooldbClient.UpgradeMigrations(); err != nil {
		t.Errorf("TestPoolDBWrites: failed on upgrade pooldb migrations: %v\n", err)
		return
	}

	suite.Run(t, new(PooldbWritesSuite))

	if err := pooldbClient.DowngradeMigrations(); err != nil {
		t.Errorf("TestPoolDBWrites: failed on downgrade pooldb migrations: %v\n", err)
	}
}

func TestTsDBReads(t *testing.T) {
	if err := tsdbClient.UpgradeMigrations(); err != nil {
		t.Errorf("TestTsDBReads: failed on upgrade tsdb migrations: %v\n", err)
		return
	}

	suite.Run(t, new(TsdbReadsSuite))

	if err := tsdbClient.DowngradeMigrations(); err != nil {
		t.Errorf("TestTsDBReads: failed on downgrade tsdb migrations: %v\n", err)
	}
}

func TestTsDBWrites(t *testing.T) {
	if err := tsdbClient.UpgradeMigrations(); err != nil {
		t.Errorf("TestTsDBWrites: failed on upgrade tsdb migrations: %v\n", err)
		return
	}

	suite.Run(t, new(TsdbWritesSuite))

	if err := tsdbClient.DowngradeMigrations(); err != nil {
		t.Errorf("TestTsDBWrites: failed on downgrade tsdb migrations: %v\n", err)
	}
}

func TestRedisReads(t *testing.T) {
	suite.Run(t, new(RedisReadsSuite))
}

func TestRedisWrites(t *testing.T) {
	suite.Run(t, new(RedisWritesSuite))
}

/* application */

func TestPool(t *testing.T) {
	if err := pooldbClient.UpgradeMigrations(); err != nil {
		t.Errorf("TestPool: failed on upgrade pooldb migrations: %v\n", err)
		return
	}

	suite.Run(t, new(PoolSuite))

	if err := pooldbClient.DowngradeMigrations(); err != nil {
		t.Errorf("TestPool: failed on downgrade pooldb migrations: %v\n", err)
	}
}

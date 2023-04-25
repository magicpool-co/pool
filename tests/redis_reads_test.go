//go:build integration

package tests

import (
	"github.com/stretchr/testify/suite"
)

type RedisReadsSuite struct {
	suite.Suite
}

func (suite *RedisReadsSuite) TestGetMiners() {
	var err error

	_, err = redisClient.GetMinerID("")
	if err != nil {
		suite.T().Errorf("failed: GetMinerID: %v", err)
	}

	_, err = redisClient.GetMinerIPAddresses("")
	if err != nil {
		suite.T().Errorf("failed: GetMinerIPAddresses: %v", err)
	}

	_, err = redisClient.GetWorkerID(0, "")
	if err != nil {
		suite.T().Errorf("failed: GetWorkerID: %v", err)
	}

	_, err = redisClient.GetTopMinerIDs("")
	if err != nil {
		suite.T().Errorf("failed: GetTopMinerIDs: %v", err)
	}
}

func (suite *RedisReadsSuite) TestGetShareIndexes() {
	var err error

	_, err = redisClient.GetShareIndexes("")
	if err != nil {
		suite.T().Errorf("failed: GetShareIndexes: %v", err)
	}
}

func (suite *RedisReadsSuite) TestGetRounds() {
	var err error

	_, err = redisClient.GetRoundShares("")
	if err != nil {
		suite.T().Errorf("failed: GetRoundShares: %v", err)
	}

	_, _, _, err = redisClient.GetRoundShareCounts("")
	if err != nil {
		suite.T().Errorf("failed: GetRoundShareCounts: %v", err)
	}
}

func (suite *RedisReadsSuite) TestGetIntervals() {
	var err error

	_, err = redisClient.GetIntervals("")
	if err != nil {
		suite.T().Errorf("failed: GetIntervals: %v", err)
	}

	_, err = redisClient.GetIntervalAcceptedShares("", "")
	if err != nil {
		suite.T().Errorf("failed: GetIntervalAcceptedShares: %v", err)
	}

	_, err = redisClient.GetIntervalRejectedShares("", "")
	if err != nil {
		suite.T().Errorf("failed: GetIntervalRejectedShares: %v", err)
	}

	_, err = redisClient.GetIntervalInvalidShares("", "")
	if err != nil {
		suite.T().Errorf("failed: GetIntervalInvalidShares: %v", err)
	}

	_, err = redisClient.GetIntervalReportedHashrates("", "")
	if err != nil {
		suite.T().Errorf("failed: GetIntervalReportedHashrates: %v", err)
	}
}

func (suite *RedisReadsSuite) TestGetCharts() {
	var err error

	_, err = redisClient.GetChartSharesLastTime("")
	if err != nil {
		suite.T().Errorf("failed: GetChartSharesLastTime: %v", err)
	}

	_, err = redisClient.GetChartBlocksLastTime("")
	if err != nil {
		suite.T().Errorf("failed: GetChartBlocksLastTime: %v", err)
	}

	_, err = redisClient.GetChartRoundsLastTime("")
	if err != nil {
		suite.T().Errorf("failed: GetChartRoundsLastTime: %v", err)
	}
}

func (suite *RedisReadsSuite) TestWriteCachedStats() {
	var err error

	_, err = redisClient.GetCachedGlobalLastShares()
	if err != nil {
		suite.T().Errorf("failed: GetCachedGlobalLastShares: %v", err)
	}

	_, err = redisClient.GetCachedGlobalLastProfits()
	if err != nil {
		suite.T().Errorf("failed: GetCachedGlobalLastProfits: %v", err)
	}

	_, err = redisClient.GetCachedLuckByChain("ETH")
	if err != nil {
		suite.T().Errorf("failed: GetCachedLuckByChain: %v", err)
	}

	_, err = redisClient.GetCachedMinersByChain("ETH")
	if err != nil {
		suite.T().Errorf("failed: GetCachedMinersByChain: %v", err)
	}

	_, err = redisClient.GetCachedWorkersByChain("ETH")
	if err != nil {
		suite.T().Errorf("failed: GetCachedWorkersByChain: %v", err)
	}
}

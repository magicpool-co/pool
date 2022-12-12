//go:build integration

package tests

import (
	"time"

	"github.com/stretchr/testify/suite"
)

type RedisWritesSuite struct {
	suite.Suite
}

func (suite *RedisReadsSuite) TestWriteMiners() {
	var err error

	err = redisClient.SetMinerID("", 0)
	if err != nil {
		suite.T().Errorf("failed: SetMinerID: %v", err)
	}

	err = redisClient.SetMinerIPAddressesBulk("", map[string]int64{"test": 3})
	if err != nil {
		suite.T().Errorf("failed: SetMinerIPAddressesBulk: %v", err)
	}

	err = redisClient.DeleteMinerIPAddresses("")
	if err != nil {
		suite.T().Errorf("failed: DeleteMinerIPAddresses: %v", err)
	}

	err = redisClient.SetWorkerID(0, "", 0)
	if err != nil {
		suite.T().Errorf("failed: SetWorkerID: %v", err)
	}

	err = redisClient.SetTopMinerIDs("", []uint64{0})
	if err != nil {
		suite.T().Errorf("failed: SetTopMinerIDs: %v", err)
	}
}

func (suite *RedisReadsSuite) TestWriteShareIndexes() {
	var err error

	err = redisClient.AddShareIndexHeight("", 0)
	if err != nil {
		suite.T().Errorf("failed: AddShareIndexHeight: %v", err)
	}

	err = redisClient.DeleteShareIndexHeight("", 0)
	if err != nil {
		suite.T().Errorf("failed: DeleteShareIndexHeight: %v", err)
	}

	_, err = redisClient.AddUniqueShare("", 0, "")
	if err != nil {
		suite.T().Errorf("failed: AddUniqueShare: %v", err)
	}
}

func (suite *RedisReadsSuite) TestWriteRounds() {
	var err error

	err = redisClient.AddAcceptedShare("", "", "", 1)
	if err != nil {
		suite.T().Errorf("failed: AddAcceptedShare: %v", err)
	}

	err = redisClient.AddRejectedShare("", "", "")
	if err != nil {
		suite.T().Errorf("failed: AddRejectedShare: %v", err)
	}

	err = redisClient.AddInvalidShare("", "", "")
	if err != nil {
		suite.T().Errorf("failed: AddInvalidShare: %v", err)
	}
}

func (suite *RedisReadsSuite) TestWriteIntervals() {
	var err error

	err = redisClient.AddInterval("", "")
	if err != nil {
		suite.T().Errorf("failed: AddInterval: %v", err)
	}

	err = redisClient.DeleteInterval("", "")
	if err != nil {
		suite.T().Errorf("failed: DeleteInterval: %v", err)
	}

	err = redisClient.SetIntervalReportedHashrateBatch("", "", map[string]float64{"test": 3})
	if err != nil {
		suite.T().Errorf("failed: SetIntervalReportedHashrateBatch: %v", err)
	}
}

func (suite *RedisReadsSuite) TestWriteCharts() {
	var err error

	err = redisClient.SetChartSharesLastTime("", time.Now())
	if err != nil {
		suite.T().Errorf("failed: SetChartSharesLastTime: %v", err)
	}

	err = redisClient.SetChartBlocksLastTime("", time.Now())
	if err != nil {
		suite.T().Errorf("failed: SetChartBlocksLastTime: %v", err)
	}

	err = redisClient.SetChartRoundsLastTime("", time.Now())
	if err != nil {
		suite.T().Errorf("failed: SetChartRoundsLastTime: %v", err)
	}
}

//go:build integration

package tests

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/magicpool-co/pool/internal/tsdb"
)

type RedisWritesSuite struct {
	suite.Suite
}

func (suite *RedisWritesSuite) TestWriteMiners() {
	var err error

	err = redisClient.SetMinerID("", 0)
	if err != nil {
		suite.T().Errorf("failed: SetMinerID: %v", err)
	}

	err = redisClient.SetMinerIPAddressesBulk("", map[string]int64{"test": 3})
	if err != nil {
		suite.T().Errorf("failed: SetMinerIPAddressesBulk: %v", err)
	}

	err = redisClient.SetMinerLatenciesBulk("", map[string]int64{"test": 3})
	if err != nil {
		suite.T().Errorf("failed: SetMinerLatenciesBulk: %v", err)
	}

	err = redisClient.RemoveMinerIPAddresses("", []string{"test"})
	if err != nil {
		suite.T().Errorf("failed: RemoveMinerIPAddresses: %v", err)
	}

	err = redisClient.AddMinerIPAddressesInactive("", []string{"test"})
	if err != nil {
		suite.T().Errorf("failed: AddMinerIPAddressesInactive: %v", err)
	}

	err = redisClient.RemoveMinerIPAddressesInactive("", []string{"test"})
	if err != nil {
		suite.T().Errorf("failed: RemoveMinerIPAddressesInactive: %v", err)
	}

	err = redisClient.SetWorkerID(0, "", 0)
	if err != nil {
		suite.T().Errorf("failed: SetWorkerID: %v", err)
	}

	err = redisClient.SetTopMinerIDsBulk("", map[uint64]float64{0: 2143}, false)
	if err != nil {
		suite.T().Errorf("failed: SetTopMinerIDsBulk: %v", err)
	}

	err = redisClient.SetTopMinerIDsBulk("", map[uint64]float64{0: 2143}, true)
	if err != nil {
		suite.T().Errorf("failed: SetTopMinerIDsBulk (increment): %v", err)
	}
}

func (suite *RedisWritesSuite) TestWriteShareIndexes() {
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

func (suite *RedisWritesSuite) TestWriteRounds() {
	var err error

	err = redisClient.AddAcceptedShare("", "", "", 0, 4, 1)
	if err != nil {
		suite.T().Errorf("failed: AddAcceptedShare: %v", err)
	}

	err = redisClient.AddAcceptedShare("", "", "", 1, 4, 1)
	if err != nil {
		suite.T().Errorf("failed: AddAcceptedShare (SOLO): %v", err)
	}

	err = redisClient.AddRejectedShare("", "", "", 0, 4)
	if err != nil {
		suite.T().Errorf("failed: AddRejectedShare: %v", err)
	}

	err = redisClient.AddRejectedShare("", "", "", 1, 4)
	if err != nil {
		suite.T().Errorf("failed: AddRejectedShare (SOLO): %v", err)
	}

	err = redisClient.AddInvalidShare("", "", "", 0, 4)
	if err != nil {
		suite.T().Errorf("failed: AddInvalidShare: %v", err)
	}

	err = redisClient.AddInvalidShare("", "", "", 1, 4)
	if err != nil {
		suite.T().Errorf("failed: AddInvalidShare (SOLO): %v", err)
	}
}

func (suite *RedisWritesSuite) TestWriteIntervals() {
	var err error

	err = redisClient.AddInterval("", "")
	if err != nil {
		suite.T().Errorf("failed: AddInterval: %v", err)
	}

	err = redisClient.DeleteInterval("", "")
	if err != nil {
		suite.T().Errorf("failed: DeleteInterval: %v", err)
	}
}

func (suite *RedisWritesSuite) TestWriteCharts() {
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

	err = redisClient.SetChartEarningsLastTime("", time.Now())
	if err != nil {
		suite.T().Errorf("failed: SetChartEarningsLastTime: %v", err)
	}
}

func (suite *RedisWritesSuite) TestWriteCachedStats() {
	var err error

	err = redisClient.SetCachedGlobalLastShares([]*tsdb.Share{&tsdb.Share{}}, 1)
	if err != nil {
		suite.T().Errorf("failed: SetCachedGlobalLastShares: %v", err)
	}

	err = redisClient.SetCachedGlobalLastProfits([]*tsdb.Block{&tsdb.Block{}}, 1)
	if err != nil {
		suite.T().Errorf("failed: SetCachedGlobalLastProfits: %v", err)
	}

	err = redisClient.SetCachedLuckByChain("ETH", 0.1, 1)
	if err != nil {
		suite.T().Errorf("failed: SetCachedLuckByChain: %v", err)
	}

	err = redisClient.SetCachedMinersByChain("ETH", 9, 1)
	if err != nil {
		suite.T().Errorf("failed: SetCachedMinersByChain: %v", err)
	}

	err = redisClient.SetCachedWorkersByChain("ETH", 9, 1)
	if err != nil {
		suite.T().Errorf("failed: SetCachedWorkersByChain: %v", err)
	}
}

//go:build integration

package tests

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/magicpool-co/pool/internal/pooldb"
)

type PooldbReadsSuite struct {
	suite.Suite
}

func (suite *PooldbReadsSuite) TestReadNode() {
	var err error

	_, err = pooldb.GetNodeURLsByChain(pooldbClient.Reader(), "ETH", true)
	if err != nil {
		suite.T().Errorf("failed: GetNodeURLsByChain: %v", err)
	}

	_, err = pooldb.GetEnabledNodes(pooldbClient.Reader(), true)
	if err != nil {
		suite.T().Errorf("failed: GetEnabledNodes: %v", err)
	}

	_, err = pooldb.GetBackupNodes(pooldbClient.Reader(), true)
	if err != nil {
		suite.T().Errorf("failed: GetBackupNodes: %v", err)
	}

	_, err = pooldb.GetPendingBackupNodes(pooldbClient.Reader(), true)
	if err != nil {
		suite.T().Errorf("failed: GetPendingBackupNodes: %v", err)
	}

	_, err = pooldb.GetPendingUpdateNodes(pooldbClient.Reader(), true)
	if err != nil {
		suite.T().Errorf("failed: GetPendingUpdateNodes: %v", err)
	}

	_, err = pooldb.GetPendingResizeNodes(pooldbClient.Reader(), true)
	if err != nil {
		suite.T().Errorf("failed: GetPendingResizeNodes: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadMiner() {
	var err error

	_, err = pooldb.GetMiner(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetMiner: %v", err)
	}

	_, err = pooldb.GetMinerID(pooldbClient.Reader(), "0x", "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetMinerID: %v", err)
	}

	_, err = pooldb.GetMiners(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetMiners: %v", err)
	}

	_, err = pooldb.GetMinerIDs(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetMinerIDs: %v", err)
	}

	_, err = pooldb.GetMinerIDsActive(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetMinerIDsActive: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadWorker() {
	var err error

	_, err = pooldb.GetWorkers(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetWorkers: %v", err)
	}

	_, err = pooldb.GetWorkerID(pooldbClient.Reader(), 0, "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetWorkerID: %v", err)
	}

	_, err = pooldb.GetWorkersActive(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetWorkersActive: %v", err)
	}

	_, err = pooldb.GetWorkerIDs(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetWorkerIDs: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadIPAddress() {
	var err error

	_, err = pooldb.GetIPAddressByMinerID(pooldbClient.Reader(), 0, "")
	if err != nil {
		suite.T().Errorf("failed: GetIPAddressByMinerID: %v", err)
	}

	_, err = pooldb.GetOldestActiveIPAddress(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetOldestActiveIPAddress: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadRound() {
	var err error

	_, err = pooldb.GetRound(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetRound: %v", err)
	}

	_, err = pooldb.GetLastRoundBeforeTime(pooldbClient.Reader(), "ETH", time.Now())
	if err != nil {
		suite.T().Errorf("failed: GetLastRoundBeforeTime: %v", err)
	}

	_, err = pooldb.GetRoundMinTimestamp(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetRoundMinTimestamp: %v", err)
	}

	_, err = pooldb.GetRounds(pooldbClient.Reader(), 0, 10)
	if err != nil {
		suite.T().Errorf("failed: GetRounds: %v", err)
	}

	_, err = pooldb.GetRoundsCount(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetRoundsCount: %v", err)
	}

	_, err = pooldb.GetRoundsBetweenTime(pooldbClient.Reader(), "ETH", time.Now(), time.Now())
	if err != nil {
		suite.T().Errorf("failed: GetRoundsBetweenTime: %v", err)
	}

	_, err = pooldb.GetPendingRoundsByChain(pooldbClient.Reader(), "ETH", 100000)
	if err != nil {
		suite.T().Errorf("failed: GetPendingRoundsByChain: %v", err)
	}

	_, err = pooldb.GetPendingRoundCountBetweenTime(pooldbClient.Reader(), "ETH", time.Now(), time.Now())
	if err != nil {
		suite.T().Errorf("failed: GetPendingRoundCountBetweenTime: %v", err)
	}

	_, err = pooldb.GetImmatureRoundsByChain(pooldbClient.Reader(), "ETH", 100000)
	if err != nil {
		suite.T().Errorf("failed: GetImmatureRoundsByChain: %v", err)
	}

	_, err = pooldb.GetMatureUnspentRounds(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetMatureUnspentRounds: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadShare() {
	var err error

	_, err = pooldb.GetSharesByRound(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetSharesByRound: %v", err)
	}
}

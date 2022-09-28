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

	_, err = pooldb.GetMinerAddress(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetMinerAddress: %v", err)
	}

	_, err = pooldb.GetMiners(pooldbClient.Reader(), []uint64{1, 2, 3})
	if err != nil {
		suite.T().Errorf("failed: GetMiners: %v", err)
	}

	_, err = pooldb.GetRecipients(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetRecipients: %v", err)
	}

	_, err = pooldb.GetActiveMinerCount(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetActiveMinerCount: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadWorker() {
	var err error

	_, err = pooldb.GetWorkerID(pooldbClient.Reader(), 0, "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetWorkerID: %v", err)
	}

	_, err = pooldb.GetWorkersByMiner(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetWorkersByMiner: %v", err)
	}

	_, err = pooldb.GetActiveWorkerCount(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetActiveWorkerCount: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadIPAddress() {
	var err error

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

	_, err = pooldb.GetRoundsByMiner(pooldbClient.Reader(), 0, 0, 10)
	if err != nil {
		suite.T().Errorf("failed: GetRoundsByMiner: %v", err)
	}

	_, err = pooldb.GetRoundsByMinerCount(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetRoundsByMinerCount: %v", err)
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

	_, err = pooldb.GetSumImmatureRoundValueByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetSumImmatureRoundValueByChain: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadShare() {
	var err error

	_, err = pooldb.GetSharesByRound(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetSharesByRound: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadUTXO() {
	var err error

	_, err = pooldb.GetUnspentUTXOsByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetUnspentUTXOsByChain: %v", err)
	}

	_, err = pooldb.GetSumUnspentUTXOValueByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetSumUnspentUTXOValueByChain: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadExchangeBatch() {
	var err error

	_, err = pooldb.GetExchangeBatch(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetExchangeBatch: %v", err)
	}

	_, err = pooldb.GetActiveExchangeBatches(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetActiveExchangeBatches: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadExchangeInput() {
	var err error

	_, err = pooldb.GetExchangeInputs(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetExchangeInputs: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadExchangeDeposit() {
	var err error

	_, err = pooldb.GetExchangeDeposits(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetExchangeDeposits: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadExchangeTrade() {
	var err error

	_, err = pooldb.GetExchangeTradesByStage(pooldbClient.Reader(), 1, 1)
	if err != nil {
		suite.T().Errorf("failed: GetExchangeTradesByStage: %v", err)
	}

	_, err = pooldb.GetExchangeTradeByPathAndStage(pooldbClient.Reader(), 1, 1, 1)
	if err != nil {
		suite.T().Errorf("failed: GetExchangeTradeByPathAndStage: %v", err)
	}

	_, err = pooldb.GetFinalExchangeTrades(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetFinalExchangeTrades: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadExchangeWithdrawal() {
	var err error

	_, err = pooldb.GetExchangeWithdrawals(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetExchangeWithdrawals: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadBalanceInput() {
	var err error

	_, err = pooldb.GetPendingBalanceInputsWithoutBatch(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetPendingBalanceInputsWithoutBatch: %v", err)
	}

	_, err = pooldb.GetBalanceInputsByBatch(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBalanceInputsByBatch: %v", err)
	}

	_, err = pooldb.GetSumBalanceInputValueByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetSumBalanceInputValueByChain: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadBalanceOutput() {
	var err error

	_, err = pooldb.GetBalanceOutputsByBatch(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBalanceOutputsByBatch: %v", err)
	}

	_, err = pooldb.GetSumBalanceOutputValueByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetSumBalanceOutputValueByChain: %v", err)
	}

	_, err = pooldb.GetSumBalanceOutputValueByMiner(pooldbClient.Reader(), 1, "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetSumBalanceOutputValueByMiner: %v", err)
	}

	_, err = pooldb.GetSumBalanceOutputAboveThreshold(pooldbClient.Reader(), "ETH", "10")
	if err != nil {
		suite.T().Errorf("failed: GetSumBalanceOutputAboveThreshold: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadPayout() {
	var err error

	_, err = pooldb.GetUnconfirmedPayouts(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetUnconfirmedPayouts: %v", err)
	}

	_, err = pooldb.GetPayouts(pooldbClient.Reader(), 10, 10)
	if err != nil {
		suite.T().Errorf("failed: GetPayouts: %v", err)
	}

	_, err = pooldb.GetPayoutsCount(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetPayouts: %v", err)
	}

	_, err = pooldb.GetPayoutsByMiner(pooldbClient.Reader(), 0, 10, 10)
	if err != nil {
		suite.T().Errorf("failed: GetPayoutsByMiner: %v", err)
	}

	_, err = pooldb.GetPayoutsByMinerCount(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetPayoutsByMinerCount: %v", err)
	}
}

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

	_, err = pooldb.GetMinersWithLastShares(pooldbClient.Reader(), []uint64{1, 2, 3})
	if err != nil {
		suite.T().Errorf("failed: GetMinersWithLastShares: %v", err)
	}

	_, err = pooldb.GetRecipients(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetRecipients: %v", err)
	}

	_, err = pooldb.GetActiveMiners(pooldbClient.Reader(), []uint64{1, 2, 3})
	if err != nil {
		suite.T().Errorf("failed: GetActiveMiners: %v", err)
	}

	_, err = pooldb.GetActiveMinersCount(pooldbClient.Reader(), "ETC")
	if err != nil {
		suite.T().Errorf("failed: GetActiveMinersCount: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadWorker() {
	var err error

	_, err = pooldb.GetWorker(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetWorker: %v", err)
	}

	_, err = pooldb.GetWorkerID(pooldbClient.Reader(), 0, "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetWorkerID: %v", err)
	}

	_, err = pooldb.GetWorkersByMiner(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetWorkersByMiner: %v", err)
	}

	_, err = pooldb.GetActiveWorkersCount(pooldbClient.Reader(), "ETC")
	if err != nil {
		suite.T().Errorf("failed: GetActiveWorkersCount: %v", err)
	}

	_, err = pooldb.GetActiveWorkersByMinersCount(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetActiveWorkersByMinersCount: %v", err)
	}

	_, err = pooldb.GetInactiveWorkersByMinersCount(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetInactiveWorkersByMinersCount: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadIPAddress() {
	var err error

	_, err = pooldb.GetOldestActiveIPAddress(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetOldestActiveIPAddress: %v", err)
	}

	_, err = pooldb.GetNewestInactiveIPAddress(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetNewestInactiveIPAddress: %v", err)
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

	_, err = pooldb.GetRoundsByMiners(pooldbClient.Reader(), []uint64{0, 1}, 0, 10)
	if err != nil {
		suite.T().Errorf("failed: GetRoundsByMiners: %v", err)
	}

	_, err = pooldb.GetRoundsByMinersCount(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetRoundsByMinersCount: %v", err)
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

	_, err = pooldb.GetSumUnspentRoundValueByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetSumUnspentRoundValueByChain: %v", err)
	}

	_, err = pooldb.GetRoundLuckByChain(pooldbClient.Reader(), "ETC", time.Hour*24*30)
	if err != nil {
		suite.T().Errorf("failed: GetRoundLuckByChain: %v", err)
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

	_, err = pooldb.GetUTXOsByTransactionID(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetUTXOsByTransactionID: %v", err)
	}

	_, err = pooldb.GetSumUnspentUTXOValueByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetSumUnspentUTXOValueByChain: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadTransaction() {
	var err error

	_, err = pooldb.GetTransaction(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetTransaction: %v", err)
	}

	_, err = pooldb.GetUnspentTransactions(pooldbClient.Reader(), "ETC")
	if err != nil {
		suite.T().Errorf("failed: GetUnspentTransactions: %v", err)
	}

	_, err = pooldb.GetUnspentTransactionCount(pooldbClient.Reader(), "ETC")
	if err != nil {
		suite.T().Errorf("failed: GetUnspentTransactionCount: %v", err)
	}

	_, err = pooldb.GetUnconfirmedTransactions(pooldbClient.Reader(), "ETC")
	if err != nil {
		suite.T().Errorf("failed: GetUnconfirmedTransactions: %v", err)
	}

	_, err = pooldb.GetUnconfirmedTransactionSum(pooldbClient.Reader(), "ETC")
	if err != nil {
		suite.T().Errorf("failed: GetUnconfirmedTransactionSum: %v", err)
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

	_, err = pooldb.GetUnregisteredExchangeDepositsByChain(pooldbClient.Reader(), "ETC")
	if err != nil {
		suite.T().Errorf("failed: GetUnregisteredExchangeDepositsByChain: %v", err)
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

	_, err = pooldb.GetPendingBalanceInputSumByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetPendingBalanceInputSumByChain: %v", err)
	}

	_, err = pooldb.GetPendingBalanceInputSumWithoutBatchByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetPendingBalanceInputSumWithoutBatchByChain: %v", err)
	}

	_, err = pooldb.GetPendingBalanceInputSumsByMiners(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetPendingBalanceInputSumsByMiners: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadBalanceOutput() {
	var err error

	_, err = pooldb.GetBalanceOutputsByBatch(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBalanceOutputsByBatch: %v", err)
	}

	_, err = pooldb.GetBalanceOutputsByPayoutTransaction(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBalanceOutputsByPayoutTransaction: %v", err)
	}

	_, err = pooldb.GetUnpaidBalanceOutputsByMiner(pooldbClient.Reader(), 1, "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetUnpaidBalanceOutputsByMiner: %v", err)
	}

	_, err = pooldb.GetUnpaidBalanceOutputSumByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetUnpaidBalanceOutputSumByChain: %v", err)
	}

	_, err = pooldb.GetUnpaidBalanceOutputSumByMiner(pooldbClient.Reader(), 1, "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetUnpaidBalanceOutputSumByMiner: %v", err)
	}

	_, err = pooldb.GetUnpaidBalanceOutputSumsByMiners(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetUnpaidBalanceOutputSumsByMiners: %v", err)
	}

	_, err = pooldb.GetUnpaidMinerIDsAbovePayoutThreshold(pooldbClient.Reader(), "ETH", "10")
	if err != nil {
		suite.T().Errorf("failed: GetUnpaidMinerIDsAbovePayoutThreshold: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadPayout() {
	var err error

	_, err = pooldb.GetUnconfirmedPayouts(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetUnconfirmedPayouts: %v", err)
	}

	_, err = pooldb.GetUnconfirmedPayoutSum(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetUnconfirmedPayoutSum: %v", err)
	}

	_, err = pooldb.GetPayouts(pooldbClient.Reader(), 10, 10)
	if err != nil {
		suite.T().Errorf("failed: GetPayouts: %v", err)
	}

	_, err = pooldb.GetPayoutsCount(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetPayouts: %v", err)
	}

	_, err = pooldb.GetPayoutsByMiners(pooldbClient.Reader(), []uint64{0, 1}, 10, 10)
	if err != nil {
		suite.T().Errorf("failed: GetPayoutsByMiners: %v", err)
	}

	_, err = pooldb.GetPayoutsByMinersCount(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetPayoutsByMinersCount: %v", err)
	}
}

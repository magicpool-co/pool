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

	_, err = pooldb.GetPendingBackupNodes(pooldbClient.Reader(), true)
	if err != nil {
		suite.T().Errorf("failed: GetPendingBackupNodes: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadMiner() {
	var err error

	_, err = pooldb.GetMiner(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetMiner: %v", err)
	}

	_, err = pooldb.GetMinerIDByChainAddress(pooldbClient.Reader(), "0x", "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetMinerIDByChainAddress: %v", err)
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
}

func (suite *PooldbReadsSuite) TestReadRecipient() {
	var err error

	_, err = pooldb.GetRecipients(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetRecipients: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadWorker() {
	var err error

	_, err = pooldb.GetWorker(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetWorker: %v", err)
	}

	_, err = pooldb.GetWorkerID(pooldbClient.Reader(), 0, "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetWorkerID: %v", err)
	}

	_, err = pooldb.GetWorkersByMinerID(pooldbClient.Reader(), 0)
	if err != nil {
		suite.T().Errorf("failed: GetWorkersByMinerID: %v", err)
	}

	_, err = pooldb.GetWorkersWithLastShares(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetWorkersWithLastShares: %v", err)
	}

	_, _, err = pooldb.GetWorkerCountByMinerIDs(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetWorkerCountByMinerIDs: %v", err)
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

	_, err = pooldb.GetRounds(pooldbClient.Reader(), 0, 10)
	if err != nil {
		suite.T().Errorf("failed: GetRounds: %v", err)
	}

	_, err = pooldb.GetRoundsByChain(pooldbClient.Reader(), "ETH", 0, 10)
	if err != nil {
		suite.T().Errorf("failed: GetRoundsByChain: %v", err)
	}

	_, err = pooldb.GetRoundsCount(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetRoundsCount: %v", err)
	}

	_, err = pooldb.GetRoundsByChainCount(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetRoundsByChainCount: %v", err)
	}

	_, err = pooldb.GetRoundsByMinerIDs(pooldbClient.Reader(), []uint64{0, 1}, 0, 10)
	if err != nil {
		suite.T().Errorf("failed: GetRoundsByMinerIDs: %v", err)
	}

	_, err = pooldb.GetRoundsByMinerIDsCount(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetRoundsByMinerIDsCount: %v", err)
	}

	_, err = pooldb.GetPendingRoundsByChain(pooldbClient.Reader(), "ETH", 100000)
	if err != nil {
		suite.T().Errorf("failed: GetPendingRoundsByChain: %v", err)
	}

	_, err = pooldb.GetImmatureRoundsByChain(pooldbClient.Reader(), "ETH", 100000)
	if err != nil {
		suite.T().Errorf("failed: GetImmatureRoundsByChain: %v", err)
	}

	_, err = pooldb.GetUnspentRoundsByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetUnspentRoundsByChain: %v", err)
	}

	_, err = pooldb.GetRoundLuckByChain(pooldbClient.Reader(), "ETC", true, time.Hour*24*30)
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

	_, err = pooldb.GetTransactionByTxID(pooldbClient.Reader(), "")
	if err != nil {
		suite.T().Errorf("failed: GetTransactionByTxID: %v", err)
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

	_, err = pooldb.GetActiveExchangeBatches(pooldbClient.Reader(), 0)
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

	_, err = pooldb.GetExchangeTrades(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetExchangeTrades: %v", err)
	}

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

	_, err = pooldb.GetPendingBalanceInputsSumWithoutBatch(pooldbClient.Reader())
	if err != nil {
		suite.T().Errorf("failed: GetPendingBalanceInputsSumWithoutBatch: %v", err)
	}

	_, err = pooldb.GetBalanceInputsByBatch(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBalanceInputsByBatch: %v", err)
	}

	_, err = pooldb.GetBalanceInputsByRound(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBalanceInputsByRound: %v", err)
	}

	_, err = pooldb.GetImmatureBalanceInputSumByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetImmatureBalanceInputSumByChain: %v", err)
	}

	_, err = pooldb.GetPendingBalanceInputSumByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetPendingBalanceInputSumByChain: %v", err)
	}

	_, err = pooldb.GetBalanceInputMinTimestamp(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetBalanceInputMinTimestamp: %v", err)
	}

	_, err = pooldb.GetBalanceInputSumFromRange(pooldbClient.Reader(), "ETH", time.Now(), time.Now())
	if err != nil {
		suite.T().Errorf("failed: GetBalanceInputSumFromRange: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadBalanceOutput() {
	var err error

	_, err = pooldb.GetBalanceOutputsByBatch(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBalanceOutputsByBatch: %v", err)
	}

	_, err = pooldb.GetBalanceOutputsByPayout(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBalanceOutputsByPayout: %v", err)
	}

	_, err = pooldb.GetBalanceOutputsByPayoutTransaction(pooldbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBalanceOutputsByPayoutTransaction: %v", err)
	}

	_, err = pooldb.GetRandomBalanceOutputAboveValue(pooldbClient.Reader(), "KAS", "1000")
	if err != nil {
		suite.T().Errorf("failed: GetRandomBalanceOutputAboveValue: %v", err)
	}

	_, err = pooldb.GetUnpaidBalanceOutputsByMiner(pooldbClient.Reader(), 1, "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetUnpaidBalanceOutputsByMiner: %v", err)
	}

	_, err = pooldb.GetUnpaidBalanceOutputSumByChain(pooldbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetUnpaidBalanceOutputSumByChain: %v", err)
	}

	_, err = pooldb.GetMinersWithBalanceAboveThresholdByChain(pooldbClient.Reader(), "ETH", "10")
	if err != nil {
		suite.T().Errorf("failed: GetMinersWithBalanceAboveThresholdByChain: %v", err)
	}
}

func (suite *PooldbReadsSuite) TestReadBalanceSum() {
	var err error

	_, err = pooldb.GetBalanceSumsByMinerIDs(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetBalanceSumsByMinerIDs: %v", err)
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

	_, err = pooldb.GetPayoutsByMinerIDs(pooldbClient.Reader(), []uint64{0, 1}, 10, 10)
	if err != nil {
		suite.T().Errorf("failed: GetPayoutsByMinerIDs: %v", err)
	}

	_, err = pooldb.GetPayoutsByMinerIDsCount(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetPayoutsByMinerIDsCount: %v", err)
	}

	_, err = pooldb.GetPayoutBalanceInputSums(pooldbClient.Reader(), []uint64{0, 1})
	if err != nil {
		suite.T().Errorf("failed: GetPayoutBalanceInputSums: %v", err)
	}
}

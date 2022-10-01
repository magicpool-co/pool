//go:build integration

package tests

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/magicpool-co/pool/internal/tsdb"
)

type TsdbReadsSuite struct {
	suite.Suite
}

func (suite *TsdbReadsSuite) TestReadRawBlocks() {
	var err error

	_, err = tsdb.GetRawBlockMaxHeight(tsdbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetRawBlockMaxHeight: %v", err)
	}

	_, err = tsdb.GetRawBlockMaxTimestampBeforeTime(tsdbClient.Reader(), "ETH", time.Now())
	if err != nil {
		suite.T().Errorf("failed: GetRawBlockMaxTimestampBeforeTime: %v", err)
	}

	_, err = tsdb.GetRawBlockRollup(tsdbClient.Reader(), "ETH", time.Now(), time.Now())
	if err != nil {
		suite.T().Errorf("failed: GetRawBlockRollup: %v", err)
	}
}

func (suite *TsdbReadsSuite) TestReadBlocks() {
	var err error

	_, err = tsdb.GetBlocks(tsdbClient.Reader(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetBlocks: %v", err)
	}

	_, err = tsdb.GetBlocksProfitability(tsdbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBlocksProfitability: %v", err)
	}

	_, err = tsdb.GetPendingBlocksAtEndTime(tsdbClient.Reader(), time.Now(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingBlocksAtEndTime: %v", err)
	}

	_, err = tsdb.GetBlockMaxEndTime(tsdbClient.Reader(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetBlockMaxEndTime: %v", err)
	}

	_, err = tsdb.GetBlocksAverageSlow(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Hour*24*30)
	if err != nil {
		suite.T().Errorf("failed: GetBlocksAverageSlow: %v", err)
	}
}

func (suite *TsdbReadsSuite) TestReadRounds() {
	var err error

	_, err = tsdb.GetRounds(tsdbClient.Reader(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetRounds: %v", err)
	}

	_, err = tsdb.GetPendingRoundsAtEndTime(tsdbClient.Reader(), time.Now(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingRoundsAtEndTime: %v", err)
	}

	_, err = tsdb.GetRoundsAverageLuckSlow(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Hour)
	if err != nil {
		suite.T().Errorf("failed: GetRoundsAverageLuckSlow: %v", err)
	}

	_, err = tsdb.GetRoundsAverageProfitabilitySlow(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Hour)
	if err != nil {
		suite.T().Errorf("failed: GetRoundsAverageProfitabilitySlow: %v", err)
	}

}

func (suite *TsdbReadsSuite) TestReadShares() {
	var err error

	_, err = tsdb.GetGlobalShares(tsdbClient.Reader(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetGlobalShares: %v", err)
	}

	_, err = tsdb.GetPendingGlobalSharesByEndTime(tsdbClient.Reader(), time.Now(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingGlobalSharesByEndTime: %v", err)
	}

	_, err = tsdb.GetMinerShares(tsdbClient.Reader(), []uint64{0, 1}, "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetMinerShares: %v", err)
	}

	_, err = tsdb.GetPendingMinerSharesByEndTime(tsdbClient.Reader(), time.Now(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingMinerSharesByEndTime: %v", err)
	}

	_, err = tsdb.GetWorkerShares(tsdbClient.Reader(), 1, "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetWorkerShares: %v", err)
	}

	_, err = tsdb.GetPendingWorkerSharesByEndTime(tsdbClient.Reader(), time.Now(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingWorkerSharesByEndTime: %v", err)
	}
}

func (suite *TsdbReadsSuite) TestReadSharesAverage() {
	var err error

	_, err = tsdb.GetGlobalSharesAverage(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Hour*24)
	if err != nil {
		suite.T().Errorf("failed: GetGlobalSharesAverage: %v", err)
	}

	_, err = tsdb.GetGlobalSharesAverageFast(tsdbClient.Reader(), time.Now(), "ETH", 1, 24, time.Hour*24)
	if err != nil {
		suite.T().Errorf("failed: GetGlobalSharesAverageFast: %v", err)
	}

	_, err = tsdb.GetGlobalSharesAverageSlow(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Hour)
	if err != nil {
		suite.T().Errorf("failed: GetGlobalSharesAverageSlow: %v", err)
	}

	_, err = tsdb.GetMinerSharesAverage(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Minute)
	if err != nil {
		suite.T().Errorf("failed: GetMinerSharesAverage: %v", err)
	}

	_, err = tsdb.GetMinerSharesAverageFast(tsdbClient.Reader(), time.Now(), "ETH", 1, 24, time.Minute)
	if err != nil {
		suite.T().Errorf("failed: GetMinerSharesAverageFast: %v", err)
	}

	_, err = tsdb.GetMinerSharesAverageSlow(tsdbClient.Reader(), 1, time.Now(), "ETH", 1, time.Second)
	if err != nil {
		suite.T().Errorf("failed: GetMinerSharesAverageSlow: %v", err)
	}

	_, err = tsdb.GetWorkerSharesAverage(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Second)
	if err != nil {
		suite.T().Errorf("failed: GetWorkerSharesAverage: %v", err)
	}

	_, err = tsdb.GetWorkerSharesAverageFast(tsdbClient.Reader(), time.Now(), "ETH", 1, 24, time.Second)
	if err != nil {
		suite.T().Errorf("failed: GetWorkerSharesAverageFast: %v", err)
	}

	_, err = tsdb.GetWorkerSharesAverageSlow(tsdbClient.Reader(), 1, time.Now(), "ETH", 1, time.Second)
	if err != nil {
		suite.T().Errorf("failed: GetWorkerSharesAverageSlow: %v", err)
	}
}

func (suite *TsdbReadsSuite) TestReadSharesSumLast() {
	var err error

	_, err = tsdb.GetGlobalSharesSum(tsdbClient.Reader(), 1, time.Hour*24)
	if err != nil {
		suite.T().Errorf("failed: GetGlobalSharesSum: %v", err)
	}

	_, err = tsdb.GetGlobalSharesLast(tsdbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetGlobalSharesLast: %v", err)
	}

	_, err = tsdb.GetMinersSharesSum(tsdbClient.Reader(), []uint64{0, 1}, 1, time.Hour*24)
	if err != nil {
		suite.T().Errorf("failed: GetMinersSharesSum: %v", err)
	}

	_, err = tsdb.GetMinersSharesLast(tsdbClient.Reader(), []uint64{0, 1}, 1)
	if err != nil {
		suite.T().Errorf("failed: GetMinersSharesLast: %v", err)
	}

	_, err = tsdb.GetWorkerSharesSum(tsdbClient.Reader(), 1, 1, time.Hour*24)
	if err != nil {
		suite.T().Errorf("failed: GetWorkerSharesSum: %v", err)
	}

	_, err = tsdb.GetWorkerSharesLast(tsdbClient.Reader(), 1, 1)
	if err != nil {
		suite.T().Errorf("failed: GetWorkerSharesLast: %v", err)
	}
}

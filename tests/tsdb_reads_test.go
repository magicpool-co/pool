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

func (suite *TsdbReadsSuite) TestReadPrices() {
	var err error

	_, err = tsdb.GetPriceMaxTimestamp(tsdbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetPriceMaxTimestamp: %v", err)
	}
}

func (suite *TsdbReadsSuite) TestReadRawBlocks() {
	var err error

	_, err = tsdb.GetRawBlockMaxTimestamp(tsdbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetRawBlockMaxTimestamp: %v", err)
	}

	_, err = tsdb.GetRawBlockMaxHeight(tsdbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetRawBlockMaxHeight: %v", err)
	}

	_, err = tsdb.GetRawBlockMaxHashByHeight(tsdbClient.Reader(), "ETH")
	if err != nil {
		suite.T().Errorf("failed: GetRawBlockMaxHashByHeight: %v", err)
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

	for _, metric := range []string{"value", "difficulty", "block_time", "hashrate"} {
		_, err = tsdb.GetBlocksSingleMetric(tsdbClient.Reader(), metric, 1)
		if err != nil {
			suite.T().Errorf("failed: GetBlocksSingleMetric: %s: %v", metric, err)
		}
	}

	_, err = tsdb.GetBlocksAdjustedValue(tsdbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBlocksAdjustedValue: %v", err)
	}

	_, err = tsdb.GetBlocksAdjustedEmission(tsdbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBlocksAdjustedEmission: %v", err)
	}

	_, err = tsdb.GetBlocksProfitability(tsdbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBlocksProfitability: %v", err)
	}

	_, err = tsdb.GetBlocksWithProfitabilityLast(tsdbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetBlocksWithProfitabilityLast: %v", err)
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

func (suite *TsdbReadsSuite) TestReadEarnings() {
	var err error
	metrics := []string{"value", "avg_value"}

	_, err = tsdb.GetEarningMaxEndTime(tsdbClient.Reader(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetEarningMaxEndTime: %v", err)
	}

	_, err = tsdb.GetGlobalEarningsAverage(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Hour)
	if err != nil {
		suite.T().Errorf("failed: GetGlobalEarningsAverage: %v", err)
	}

	_, err = tsdb.GetMinerEarningsAverage(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Hour)
	if err != nil {
		suite.T().Errorf("failed: GetMinerEarningsAverage: %v", err)
	}

	for _, metric := range metrics {
		_, err = tsdb.GetGlobalEarningsSingleMetric(tsdbClient.Reader(), metric, 3)
		if err != nil {
			suite.T().Errorf("failed: GetGlobalEarningsSingleMetric: %s: %v", metric, err)
		}

		_, err = tsdb.GetMinerEarningsSingleMetric(tsdbClient.Reader(), []uint64{0, 1}, metric, 3)
		if err != nil {
			suite.T().Errorf("failed: GetMinerEarningsSingleMetric: %s: %v", metric, err)
		}
	}
}

func (suite *TsdbReadsSuite) TestReadShares() {
	var err error
	metrics := []string{"hashrate", "avg_hashrate", "accepted_shares", "accepted_adjusted_shares"}

	_, err = tsdb.GetGlobalShares(tsdbClient.Reader(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingGlobalSharesByEndTime: %v", err)
	}

	for _, metric := range metrics {
		_, err = tsdb.GetGlobalSharesSingleMetric(tsdbClient.Reader(), metric, 1)
		if err != nil {
			suite.T().Errorf("failed: GetGlobalSharesSingleMetric: %s: %v", metric, err)
		}
	}

	_, err = tsdb.GetPendingGlobalSharesByEndTime(tsdbClient.Reader(), time.Now(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingGlobalSharesByEndTime: %v", err)
	}

	_, err = tsdb.GetMinerShares(tsdbClient.Reader(), []uint64{0, 1}, "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingGlobalSharesByEndTime: %v", err)
	}

	for _, metric := range metrics {
		_, err = tsdb.GetMinerSharesSingleMetric(tsdbClient.Reader(), []uint64{0, 1}, metric, 1)
		if err != nil {
			suite.T().Errorf("failed: GetMinerSharesSingleMetric: %s: %v", metric, err)
		}
	}

	_, err = tsdb.GetPendingMinerSharesByEndTime(tsdbClient.Reader(), time.Now(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingMinerSharesByEndTime: %v", err)
	}

	_, err = tsdb.GetMinerSharesByEndTime(tsdbClient.Reader(), time.Now(), []uint64{0, 1, 2}, "ETH", 0)
	if err != nil {
		suite.T().Errorf("failed: GetMinerSharesByEndTime: %v", err)
	}

	_, err = tsdb.GetWorkerShares(tsdbClient.Reader(), 1, "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingGlobalSharesByEndTime: %v", err)
	}

	for _, metric := range metrics {
		_, err = tsdb.GetWorkerSharesSingleMetric(tsdbClient.Reader(), 1, metric, 1)
		if err != nil {
			suite.T().Errorf("failed: GetWorkerSharesSingleMetric: %s: %v", metric, err)
		}
	}

	_, err = tsdb.GetPendingWorkerSharesByEndTime(tsdbClient.Reader(), time.Now(), "ETH", 1)
	if err != nil {
		suite.T().Errorf("failed: GetPendingWorkerSharesByEndTime: %v", err)
	}

	_, err = tsdb.GetWorkerSharesAllChainsByEndTime(tsdbClient.Reader(), time.Now(), []uint64{0, 1, 2}, 0)
	if err != nil {
		suite.T().Errorf("failed: GetWorkerSharesAllChainsByEndTime: %v", err)
	}
}

func (suite *TsdbReadsSuite) TestReadSharesAverage() {
	var err error

	_, err = tsdb.GetGlobalSharesAverage(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Hour*24)
	if err != nil {
		suite.T().Errorf("failed: GetGlobalSharesAverage: %v", err)
	}

	_, err = tsdb.GetMinerSharesAverage(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Minute)
	if err != nil {
		suite.T().Errorf("failed: GetMinerSharesAverage: %v", err)
	}

	_, err = tsdb.GetWorkerSharesAverage(tsdbClient.Reader(), time.Now(), "ETH", 1, time.Second)
	if err != nil {
		suite.T().Errorf("failed: GetWorkerSharesAverage: %v", err)
	}
}

func (suite *TsdbReadsSuite) TestReadSharesSumLast() {
	var err error

	_, err = tsdb.GetGlobalSharesLast(tsdbClient.Reader(), 1)
	if err != nil {
		suite.T().Errorf("failed: GetGlobalSharesLast: %v", err)
	}

	_, err = tsdb.GetMinersSharesSum(tsdbClient.Reader(), []uint64{0, 1, 2}, 1, time.Hour*24)
	if err != nil {
		suite.T().Errorf("failed: GetMinersSharesSum: %v", err)
	}

	_, err = tsdb.GetMinersSharesLast(tsdbClient.Reader(), []uint64{0, 1}, 1)
	if err != nil {
		suite.T().Errorf("failed: GetMinersSharesLast: %v", err)
	}

	_, err = tsdb.GetWorkerSharesSum(tsdbClient.Reader(), []uint64{0, 1, 2}, 1, time.Hour*24)
	if err != nil {
		suite.T().Errorf("failed: GetWorkerSharesSum: %v", err)
	}

	_, err = tsdb.GetWorkerSharesLast(tsdbClient.Reader(), 1, 1)
	if err != nil {
		suite.T().Errorf("failed: GetWorkerSharesLast: %v", err)
	}
}

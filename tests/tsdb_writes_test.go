//go:build integration

package tests

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/types"
)

type TsdbWritesSuite struct {
	suite.Suite
}

func (suite *TsdbWritesSuite) TestWritePrices() {
	tests := []struct {
		price *tsdb.Price
	}{
		{
			price: &tsdb.Price{
				ChainID: "ETH",

				PriceUSD: 0.1,
				PriceBTC: 0.1,
				PriceETH: 0.1,

				Timestamp: time.Now(),
			},
		},
	}

	var err error
	for i, tt := range tests {
		err = tsdb.InsertPrices(tsdbClient.Writer(), tt.price)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertPrices: %v", i, err)
		}
	}
}

func (suite *TsdbWritesSuite) TestWriteRawBlock() {
	tests := []struct {
		block *tsdb.RawBlock
	}{
		{
			block: &tsdb.RawBlock{
				Timestamp: time.Now(),
			},
		},
	}

	var err error
	for i, tt := range tests {
		err = tsdb.InsertRawBlocks(tsdbClient.Writer(), tt.block)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertRawBlocks: %v", i, err)
		}

		err = tsdb.DeleteRawBlocksBeforeTime(tsdbClient.Writer(), "ETH", time.Now())
		if err != nil {
			suite.T().Errorf("failed on %d: DeleteRawBlocksBeforeTime: %v", i, err)
		}
	}
}

func (suite *TsdbWritesSuite) TestWriteBlock() {
	tests := []struct {
		block *tsdb.Block
	}{
		{
			block: &tsdb.Block{
				StartTime: time.Now(),
				EndTime:   time.Now(),
			},
		},
	}

	var err error
	for i, tt := range tests {
		err = tsdb.InsertBlocks(tsdbClient.Writer(), tt.block)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertBlocks: %v", i, err)
		}

		err = tsdb.InsertPartialBlocks(tsdbClient.Writer(), tt.block)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertPartialBlocks: %v", i, err)
		}

		err = tsdb.InsertFinalBlocks(tsdbClient.Writer(), tt.block)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertFinalBlocks: %v", i, err)
		}

		err = tsdb.DeleteBlocksBeforeEndTime(tsdbClient.Writer(), time.Now(), "ETH", 1)
		if err != nil {
			suite.T().Errorf("failed on %d: DeleteBlocksBeforeEndTime: %v", i, err)
		}
	}
}

func (suite *TsdbWritesSuite) TestWriteEarning() {
	tests := []struct {
		earning *tsdb.Earning
	}{
		{
			earning: &tsdb.Earning{
				MinerID:   types.Uint64Ptr(1),
				StartTime: time.Now(),
				EndTime:   time.Now(),
			},
		},
	}

	var err error
	for i, tt := range tests {
		err = tsdb.InsertGlobalEarnings(tsdbClient.Writer(), tt.earning)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertGlobalEarnings: %v", i, err)
		}

		err = tsdb.InsertMinerEarnings(tsdbClient.Writer(), tt.earning)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertMinerEarnings: %v", i, err)
		}
	}
}

func (suite *TsdbWritesSuite) TestWriteShare() {
	tests := []struct {
		share *tsdb.Share
	}{
		{
			share: &tsdb.Share{
				MinerID:   types.Uint64Ptr(1),
				WorkerID:  types.Uint64Ptr(1),
				StartTime: time.Now(),
				EndTime:   time.Now(),
			},
		},
	}

	var err error
	for i, tt := range tests {
		err = tsdb.InsertGlobalShares(tsdbClient.Writer(), tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertGlobalShares: %v", i, err)
		}

		err = tsdb.InsertMinerShares(tsdbClient.Writer(), tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertMinerShares: %v", i, err)
		}

		err = tsdb.InsertWorkerShares(tsdbClient.Writer(), tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertWorkerShares: %v", i, err)
		}

		err = tsdb.InsertPartialGlobalShares(tsdbClient.Writer(), tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertPartialGlobalShares: %v", i, err)
		}

		err = tsdb.InsertPartialMinerShares(tsdbClient.Writer(), tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertPartialMinerShares: %v", i, err)
		}

		err = tsdb.InsertPartialWorkerShares(tsdbClient.Writer(), tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertPartialWorkerShares: %v", i, err)
		}

		err = tsdb.InsertFinalGlobalShares(tsdbClient.Writer(), tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertFinalGlobalShares: %v", i, err)
		}

		err = tsdb.InsertFinalMinerShares(tsdbClient.Writer(), tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertFinalMinerShares: %v", i, err)
		}

		err = tsdb.InsertFinalWorkerShares(tsdbClient.Writer(), tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: InsertFinalWorkerShares: %v", i, err)
		}

		err = tsdb.DeleteGlobalSharesBeforeEndTime(tsdbClient.Writer(), time.Now(), "ETH", 1)
		if err != nil {
			suite.T().Errorf("failed on %d: DeleteGlobalSharesBeforeEndTime: %v", i, err)
		}

		err = tsdb.DeleteMinerSharesBeforeEndTime(tsdbClient.Writer(), time.Now(), "ETH", 1)
		if err != nil {
			suite.T().Errorf("failed on %d: DeleteMinerSharesBeforeEndTime: %v", i, err)
		}

		err = tsdb.DeleteWorkerSharesBeforeEndTime(tsdbClient.Writer(), time.Now(), "ETH", 1)
		if err != nil {
			suite.T().Errorf("failed on %d: DeleteWorkerSharesBeforeEndTime: %v", i, err)
		}
	}
}

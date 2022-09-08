//go:build integration

package tests

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/magicpool-co/pool/internal/pooldb"
)

type PooldbWritesSuite struct {
	suite.Suite
}

func (suite *PooldbWritesSuite) TestWriteNode() {
	tests := []struct {
		node *pooldb.Node
	}{
		{
			&pooldb.Node{
				ChainID: "ETC",
			},
		},
	}

	var err error
	for i, tt := range tests {
		_, err = pooldb.InsertNode(pooldbClient.Writer(), tt.node)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteMiner() {
	tests := []struct {
		miner *pooldb.Miner
	}{
		{
			&pooldb.Miner{
				ChainID: "ETC",
			},
		},
	}

	var err error
	for i, tt := range tests {
		_, err = pooldb.InsertMiner(pooldbClient.Writer(), tt.miner)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		cols := []string{"active", "last_login", "last_share"}
		err = pooldb.UpdateMiner(pooldbClient.Writer(), tt.miner, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteWorker() {
	tests := []struct {
		worker *pooldb.Worker
	}{
		{
			&pooldb.Worker{},
		},
	}

	minerID, err := pooldb.InsertMiner(pooldbClient.Writer(), &pooldb.Miner{ChainID: "ETH"})
	if err != nil {
		suite.T().Errorf("failed on preliminary miner insert: %v", err)
	}

	for i, tt := range tests {
		tt.worker.MinerID = minerID
		_, err = pooldb.InsertWorker(pooldbClient.Writer(), tt.worker)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		cols := []string{"active", "last_login", "last_share"}
		err = pooldb.UpdateWorker(pooldbClient.Writer(), tt.worker, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteIPAddress() {
	tests := []struct {
		address *pooldb.IPAddress
	}{
		{
			&pooldb.IPAddress{
				IPAddress: "192.168.1.1",
				Active:    true,
				LastShare: time.Now(),
			},
		},
	}

	minerID, err := pooldb.InsertMiner(pooldbClient.Writer(), &pooldb.Miner{ChainID: "ETH"})
	if err != nil {
		suite.T().Errorf("failed on preliminary miner insert: %v", err)
	}

	for i, tt := range tests {
		tt.address.MinerID = minerID
		err = pooldb.InsertIPAddresses(pooldbClient.Writer(), tt.address)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		err = pooldb.UpdateIPAddressesSetInactive(pooldbClient.Writer(), time.Hour)
		if err != nil {
			suite.T().Errorf("failed on %d: update set inactive: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteRound() {
	tests := []struct {
		round *pooldb.Round
	}{
		{
			&pooldb.Round{
				ID:      1,
				ChainID: "ETC",
				MinerID: 1,
			},
		},
	}

	var err error
	for i, tt := range tests {
		_, err = pooldb.InsertRound(pooldbClient.Writer(), tt.round)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		cols := []string{"hash", "value", "pending", "mature", "uncle",
			"spent", "uncle_height", "orphan"}
		err = pooldb.UpdateRound(pooldbClient.Writer(), tt.round, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: updateStatus: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteShare() {
	tests := []struct {
		share *pooldb.Share
	}{
		{
			&pooldb.Share{
				RoundID: 1,
				MinerID: 1,
			},
		},
	}

	var err error
	for i, tt := range tests {
		_, err = pooldb.InsertShare(pooldbClient.Writer(), tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		err = pooldb.InsertShares(pooldbClient.Writer(), &pooldb.Share{RoundID: 1, MinerID: 1}, &pooldb.Share{RoundID: 1, MinerID: 1})
		if err != nil {
			suite.T().Errorf("failed on %d: bulk insert: %v", i, err)
		}

	}
}

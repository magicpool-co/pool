//go:build integration

package tests

import (
	"math/big"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/dbcl"
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

		cols := []string{"active", "synced", "height", "needs_backup", "pending_backup",
			"needs_update", "pending_update", "needs_resize", "pending_resize", "backup_at"}
		err = pooldb.UpdateNode(pooldbClient.Writer(), tt.node, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update: %v", i, err)
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

		cols := []string{"threshold", "active"}
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

	minerID, err := pooldb.InsertMiner(pooldbClient.Writer(), &pooldb.Miner{ChainID: "ETH", Address: "0"})
	if err != nil {
		suite.T().Errorf("failed on preliminary miner insert: %v", err)
	}

	for i, tt := range tests {
		tt.worker.MinerID = minerID
		_, err = pooldb.InsertWorker(pooldbClient.Writer(), tt.worker)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		cols := []string{"active"}
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
				ChainID:   "ETC",
				IPAddress: "192.168.1.1",
				Active:    true,
				LastShare: time.Now(),
			},
		},
	}

	minerID, err := pooldb.InsertMiner(pooldbClient.Writer(), &pooldb.Miner{ChainID: "ETH", Address: "1"})
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

		err = pooldb.UpdateIPAddressesSetExpired(pooldbClient.Writer(), time.Hour)
		if err != nil {
			suite.T().Errorf("failed on %d: update set expired: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteRound() {
	tests := []struct {
		round *pooldb.Round
	}{
		{
			&pooldb.Round{
				ChainID: "ETC",
				MinerID: 1,
				Value:   dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
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
			suite.T().Errorf("failed on %d: update: %v", i, err)
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
		err = pooldb.InsertShares(pooldbClient.Writer(), tt.share, tt.share)
		if err != nil {
			suite.T().Errorf("failed on %d: bulk insert: %v", i, err)
		}

	}
}

func (suite *PooldbWritesSuite) TestWriteUTXO() {
	tests := []struct {
		utxo *pooldb.UTXO
	}{
		{
			&pooldb.UTXO{
				ChainID: "ETC",
				Value:   dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			},
		},
	}

	var err error
	for i, tt := range tests {
		err = pooldb.InsertUTXOs(pooldbClient.Writer(), tt.utxo, tt.utxo)
		if err != nil {
			suite.T().Errorf("failed on %d: bulk insert: %v", i, err)
		}

		err = pooldb.UpdateUTXO(pooldbClient.Writer(), tt.utxo, []string{"active", "spent"})
		if err != nil {
			suite.T().Errorf("failed on %d: update: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteTransaction() {
	tests := []struct {
		tx *pooldb.Transaction
	}{
		{
			&pooldb.Transaction{
				ChainID:   "ETC",
				Value:     dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				Fee:       dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				Remainder: dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			},
		},
	}

	var err error
	for i, tt := range tests {
		_, err = pooldb.InsertTransaction(pooldbClient.Writer(), tt.tx)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		cols := []string{"height", "fee", "fee_balance", "spent", "confirmed", "failed"}
		err = pooldb.UpdateTransaction(pooldbClient.Writer(), tt.tx, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteExchangeBatch() {
	tests := []struct {
		batch *pooldb.ExchangeBatch
	}{
		{
			&pooldb.ExchangeBatch{},
		},
	}

	var err error
	for i, tt := range tests {
		_, err = pooldb.InsertExchangeBatch(pooldbClient.Writer(), tt.batch)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		cols := []string{"status", "completed_at"}
		err = pooldb.UpdateExchangeBatch(pooldbClient.Writer(), tt.batch, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteExchangeInput() {
	tests := []struct {
		input *pooldb.ExchangeInput
	}{
		{
			&pooldb.ExchangeInput{
				InChainID:  "ETC",
				OutChainID: "ETH",
				Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			},
		},
	}

	batchID, err := pooldb.InsertExchangeBatch(pooldbClient.Writer(), &pooldb.ExchangeBatch{})
	if err != nil {
		suite.T().Errorf("failed on preliminary batch insert: %v", err)
	}

	for i, tt := range tests {
		tt.input.BatchID = batchID
		err = pooldb.InsertExchangeInputs(pooldbClient.Writer(), tt.input, tt.input)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteExchangeDeposit() {
	tests := []struct {
		deposit *pooldb.ExchangeDeposit
	}{
		{
			&pooldb.ExchangeDeposit{
				ChainID:   "ETC",
				NetworkID: "ETC",
				Value:     dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			},
		},
	}

	batchID, err := pooldb.InsertExchangeBatch(pooldbClient.Writer(), &pooldb.ExchangeBatch{})
	if err != nil {
		suite.T().Errorf("failed on preliminary batch insert: %v", err)
	}

	for i, tt := range tests {
		tt.deposit.BatchID = batchID
		_, err = pooldb.InsertExchangeDeposit(pooldbClient.Writer(), tt.deposit)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		cols := []string{"exchange_txid", "exchange_deposit_id",
			"value", "fees", "confirmed", "registered"}
		err = pooldb.UpdateExchangeDeposit(pooldbClient.Writer(), tt.deposit, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteExchangeTrade() {
	tests := []struct {
		trade *pooldb.ExchangeTrade
	}{
		{
			&pooldb.ExchangeTrade{
				FromChainID: "ETC",
				ToChainID:   "ETH",
			},
		},
	}

	batchID, err := pooldb.InsertExchangeBatch(pooldbClient.Writer(), &pooldb.ExchangeBatch{})
	if err != nil {
		suite.T().Errorf("failed on preliminary batch insert: %v", err)
	}

	for i, tt := range tests {
		tt.trade.BatchID = batchID
		err = pooldb.InsertExchangeTrades(pooldbClient.Writer(), tt.trade, tt.trade)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		cols := []string{"exchange_trade_id", "value", "proceeds", "trade_fees",
			"cumulative_deposit_fees", "cumulative_trade_fees", "order_price",
			"fill_price", "cumulative_fill_price", "slippage", "initiated", "confirmed"}
		err = pooldb.UpdateExchangeTrade(pooldbClient.Writer(), tt.trade, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteExchangeWithdrawal() {
	tests := []struct {
		withdrawal *pooldb.ExchangeWithdrawal
	}{
		{
			&pooldb.ExchangeWithdrawal{
				ChainID:   "ETC",
				NetworkID: "ETC",
				Value:     dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			},
		},
	}

	batchID, err := pooldb.InsertExchangeBatch(pooldbClient.Writer(), &pooldb.ExchangeBatch{})
	if err != nil {
		suite.T().Errorf("failed on preliminary batch insert: %v", err)
	}

	for i, tt := range tests {
		tt.withdrawal.BatchID = batchID
		_, err = pooldb.InsertExchangeWithdrawal(pooldbClient.Writer(), tt.withdrawal)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		cols := []string{"value", "withdrawal_fees", "cumulative_fees", "confirmed"}
		err = pooldb.UpdateExchangeWithdrawal(pooldbClient.Writer(), tt.withdrawal, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteBalanceInput() {
	tests := []struct {
		input *pooldb.BalanceInput
	}{
		{
			&pooldb.BalanceInput{
				ChainID:    "ETC",
				OutChainID: "ETH",
				Value:      dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				PoolFees:   dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			},
		},
	}

	minerID, err := pooldb.InsertMiner(pooldbClient.Writer(), &pooldb.Miner{ChainID: "ETH", Address: "2"})
	if err != nil {
		suite.T().Errorf("failed on preliminary miner insert: %v", err)
	}

	roundID, err := pooldb.InsertRound(pooldbClient.Writer(), &pooldb.Round{ChainID: "ETH", MinerID: minerID})
	if err != nil {
		suite.T().Errorf("failed on preliminary round insert: %v", err)
	}

	for i, tt := range tests {
		tt.input.RoundID = roundID
		tt.input.MinerID = minerID
		err = pooldb.InsertBalanceInputs(pooldbClient.Writer(), tt.input, tt.input)
		if err != nil {
			suite.T().Errorf("failed on %d: bulk insert: %v", i, err)
		}

		cols := []string{"balance_output_id", "batch_id", "pending"}
		err = pooldb.UpdateBalanceInput(pooldbClient.Writer(), tt.input, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWriteBalanceOutput() {
	tests := []struct {
		output *pooldb.BalanceOutput
	}{
		{
			&pooldb.BalanceOutput{
				ChainID:      "ETC",
				Value:        dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				PoolFees:     dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				ExchangeFees: dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			},
		},
	}

	minerID, err := pooldb.InsertMiner(pooldbClient.Writer(), &pooldb.Miner{ChainID: "ETH", Address: "3"})
	if err != nil {
		suite.T().Errorf("failed on preliminary miner insert: %v", err)
	}

	for i, tt := range tests {
		tt.output.MinerID = minerID
		_, err = pooldb.InsertBalanceOutput(pooldbClient.Writer(), tt.output)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		err = pooldb.InsertBalanceOutputs(pooldbClient.Writer(), tt.output, tt.output)
		if err != nil {
			suite.T().Errorf("failed on %d: bulk insert: %v", i, err)
		}

		cols := []string{"out_payout_id"}
		err = pooldb.UpdateBalanceOutput(pooldbClient.Writer(), tt.output, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update: %v", i, err)
		}
	}
}

func (suite *PooldbWritesSuite) TestWritePayout() {
	tests := []struct {
		payout *pooldb.Payout
	}{
		{
			&pooldb.Payout{
				ChainID:      "ETC",
				Value:        dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				FeeBalance:   dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				PoolFees:     dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
				ExchangeFees: dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			},
		},
	}

	minerID, err := pooldb.InsertMiner(pooldbClient.Writer(), &pooldb.Miner{ChainID: "ETH", Address: "4"})
	if err != nil {
		suite.T().Errorf("failed on preliminary miner insert: %v", err)
	}

	for i, tt := range tests {
		tt.payout.MinerID = minerID
		_, err = pooldb.InsertPayout(pooldbClient.Writer(), tt.payout)
		if err != nil {
			suite.T().Errorf("failed on %d: insert: %v", i, err)
		}

		cols := []string{"height", "value", "tx_fees", "fee_balance", "confirmed"}
		err = pooldb.UpdatePayout(pooldbClient.Writer(), tt.payout, cols)
		if err != nil {
			suite.T().Errorf("failed on %d: update: %v", i, err)
		}
	}
}

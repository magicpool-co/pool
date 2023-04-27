package payout

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/core/bank"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

const (
	maxBatchSize = 10
)

type Client struct {
	pooldb   *dbcl.Client
	redis    *redis.Client
	telegram *telegram.Client
	bank     *bank.Client
}

func New(pooldbClient *dbcl.Client, redisClient *redis.Client, telegramClient *telegram.Client) *Client {
	client := &Client{
		pooldb:   pooldbClient,
		redis:    redisClient,
		telegram: telegramClient,
		bank:     bank.New(pooldbClient, redisClient, telegramClient),
	}

	return client
}

func (c *Client) InitiatePayouts(node types.PayoutNode) error {
	dbTx, err := c.pooldb.Begin()
	if err != nil {
		return err
	}
	defer dbTx.SafeRollback()

	payoutBound, err := common.GetDefaultPayoutBounds(node.Chain())
	if err != nil {
		return err
	}

	miners, err := pooldb.GetMinersWithBalanceAboveThresholdByChain(dbTx, node.Chain(), payoutBound.Default.String())
	if err != nil {
		return err
	} else if len(miners) == 0 {
		return nil
	}

	balanceOutputSums := make([]*pooldb.BalanceOutput, len(miners))
	balanceOutputIdx := make(map[uint64][]*pooldb.BalanceOutput, len(miners))
	for i, miner := range miners {
		balanceOutputs, err := pooldb.GetUnpaidBalanceOutputsByMiner(dbTx, miner.ID, node.Chain())
		if err != nil {
			return err
		} else if len(balanceOutputs) == 0 {
			return fmt.Errorf("no balance outputs found for miner %d", miner.ID)
		}

		valueSum, poolFeesSum, exchangeFeesSum := new(big.Int), new(big.Int), new(big.Int)
		balanceOutputIdx[miner.ID] = make([]*pooldb.BalanceOutput, 0)
		for _, balanceOutput := range balanceOutputs {
			if !balanceOutput.Value.Valid {
				return fmt.Errorf("no value for balance output %d", balanceOutput.ID)
			} else if !balanceOutput.PoolFees.Valid {
				return fmt.Errorf("no pool fees for balance output %d", balanceOutput.ID)
			} else if !balanceOutput.ExchangeFees.Valid {
				return fmt.Errorf("no exchange fees for balance output %d", balanceOutput.ID)
			}
			valueSum.Add(valueSum, balanceOutput.Value.BigInt)
			poolFeesSum.Add(poolFeesSum, balanceOutput.PoolFees.BigInt)
			exchangeFeesSum.Add(exchangeFeesSum, balanceOutput.ExchangeFees.BigInt)

			balanceOutputIdx[miner.ID] = append(balanceOutputIdx[miner.ID], balanceOutput)
		}

		threshold := payoutBound.Default
		if miner.Threshold.Valid && miner.Threshold.BigInt.Cmp(common.Big0) > 0 {
			threshold = miner.Threshold.BigInt
		}

		if valueSum.Cmp(threshold) < 0 {
			return fmt.Errorf("miner %d not actually above threshold: %s < %s", miner.ID, valueSum, threshold)
		}

		balanceOutputSums[i] = &pooldb.BalanceOutput{
			MinerID: miner.ID,
			ChainID: node.Chain(),

			Value:        dbcl.NullBigInt{Valid: true, BigInt: valueSum},
			PoolFees:     dbcl.NullBigInt{Valid: true, BigInt: poolFeesSum},
			ExchangeFees: dbcl.NullBigInt{Valid: true, BigInt: exchangeFeesSum},
		}
	}

	payouts := make([]*pooldb.Payout, len(balanceOutputSums))
	for i, balanceOutput := range balanceOutputSums {
		address, err := pooldb.GetMinerAddress(dbTx, balanceOutput.MinerID)
		if err != nil {
			return err
		}

		payouts[i] = &pooldb.Payout{
			ChainID: node.Chain(),
			MinerID: balanceOutput.MinerID,
			Address: address,

			Value:        balanceOutput.Value,
			FeeBalance:   dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			PoolFees:     balanceOutput.PoolFees,
			ExchangeFees: balanceOutput.ExchangeFees,

			Pending: true,
		}
	}

	switch node.GetAccountingType() {
	case types.AccountStructure:
		outputList := make([][]*types.TxOutput, len(payouts))
		for i, payout := range payouts {
			if !payout.Value.Valid {
				return fmt.Errorf("no value for payout %d", payout.ID)
			} else if !payout.FeeBalance.Valid {
				return fmt.Errorf("no fee balance for payout %d", payout.ID)
			}

			outputList[i] = []*types.TxOutput{
				&types.TxOutput{
					Address:    payout.Address,
					Value:      payout.Value.BigInt,
					FeeBalance: payout.FeeBalance.BigInt,
				},
			}
		}

		txs, err := c.bank.PrepareOutgoingTxs(dbTx, node, types.PayoutTx, outputList...)
		if err != nil {
			return err
		} else if len(txs) != len(payouts) {
			return fmt.Errorf("tx and payout count mismatch: %d and %d", len(txs), len(payouts))
		}

		for i, payout := range payouts {
			if txs[i] == nil {
				continue
			}

			payout.TransactionID = types.Uint64Ptr(txs[i].ID)
			payout.TxID = txs[i].TxID
			payout.TxFees = dbcl.NullBigInt{Valid: true, BigInt: outputList[i][0].Fee}
			payoutID, err := pooldb.InsertPayout(dbTx, payout)
			if err != nil {
				return err
			}

			for _, balanceOutput := range balanceOutputIdx[payout.MinerID] {
				balanceOutput.OutPayoutID = types.Uint64Ptr(payoutID)
				err = pooldb.UpdateBalanceOutput(dbTx, balanceOutput, []string{"out_payout_id"})
				if err != nil {
					return err
				}
			}

			explorerURL := node.GetAddressExplorerURL(payout.Address)
			floatValue := common.BigIntToFloat64(payout.Value.BigInt, node.GetUnits().Big())
			c.telegram.NotifyInitiatePayout(payout.ID, payout.ChainID, payout.Address, explorerURL, floatValue)
		}
	case types.UTXOStructure:
		if len(payouts) > maxBatchSize {
			payouts = payouts[:maxBatchSize]
		}

		outputs := make([]*types.TxOutput, len(payouts))
		for i, payout := range payouts {
			if !payout.Value.Valid {
				return fmt.Errorf("no value for payout %d", payout.ID)
			}

			outputs[i] = &types.TxOutput{
				Address:  payout.Address,
				Value:    payout.Value.BigInt,
				SplitFee: true,
			}
		}

		txs, err := c.bank.PrepareOutgoingTxs(dbTx, node, types.PayoutTx, outputs)
		if err != nil {
			return err
		} else if len(txs) != 1 {
			return fmt.Errorf("tx count not one: %d", len(txs))
		} else if txs[0] == nil {
			return nil
		}
		tx := txs[0]

		feeIdx := make(map[string]*big.Int)
		for _, output := range outputs {
			feeIdx[output.Address] = output.Fee
		}

		for _, payout := range payouts {
			payout.TransactionID = types.Uint64Ptr(tx.ID)
			payout.TxID = tx.TxID
			payout.TxFees = dbcl.NullBigInt{Valid: true, BigInt: feeIdx[payout.Address]}
			payoutID, err := pooldb.InsertPayout(dbTx, payout)
			if err != nil {
				return err
			}

			for _, balanceOutput := range balanceOutputIdx[payout.MinerID] {
				balanceOutput.OutPayoutID = types.Uint64Ptr(payoutID)
				err = pooldb.UpdateBalanceOutput(dbTx, balanceOutput, []string{"out_payout_id"})
				if err != nil {
					return err
				}
			}

			explorerURL := node.GetAddressExplorerURL(payout.Address)
			floatValue := common.BigIntToFloat64(payout.Value.BigInt, node.GetUnits().Big())
			c.telegram.NotifyInitiatePayout(payoutID, payout.ChainID, payout.Address, explorerURL, floatValue)
		}
	default:
		return nil
	}

	return dbTx.SafeCommit()
}

func (c *Client) finalizePayout(payout *pooldb.Payout) error {
	if payout.TransactionID == nil {
		return fmt.Errorf("no transaction id for payout %d", payout.ID)
	}

	tx, err := pooldb.GetTransaction(c.pooldb.Reader(), types.Uint64Value(payout.TransactionID))
	if err != nil {
		return err
	} else if !tx.Spent || !tx.Confirmed {
		if payout.Pending {
			payout.Pending = false
			err = pooldb.UpdatePayout(c.pooldb.Writer(), payout, []string{"pending"})
			if err != nil {
				return err
			}
		}

		return nil
	} else if !tx.Value.Valid {
		return fmt.Errorf("no value for tx %s", tx.TxID)
	} else if !tx.Fee.Valid {
		return fmt.Errorf("no fee for tx %s", tx.TxID)
	}

	balanceOutputs, err := pooldb.GetBalanceOutputsByPayout(c.pooldb.Reader(), payout.ID)
	if err != nil {
		return err
	}

	dbTx, err := c.pooldb.Begin()
	if err != nil {
		return err
	}
	defer dbTx.SafeRollback()

	sumBalanceToSubtract := new(big.Int)
	for _, balanceOutput := range balanceOutputs {
		if !balanceOutput.Value.Valid {
			return fmt.Errorf("no value for balance output %d", balanceOutput.ID)
		}

		balanceOutput.Spent = true
		err := pooldb.UpdateBalanceOutput(dbTx, balanceOutput, []string{"spent"})
		if err != nil {
			return err
		}

		sumBalanceToSubtract.Add(sumBalanceToSubtract, balanceOutput.Value.BigInt)
	}

	// subtract sum value for balance outputs spent in the payout
	err = pooldb.InsertSubtractBalanceSums(dbTx, &pooldb.BalanceSum{
		MinerID: payout.MinerID,
		ChainID: payout.ChainID,

		MatureValue: dbcl.NullBigInt{Valid: true, BigInt: sumBalanceToSubtract},
	})
	if err != nil {
		return err
	}

	if tx.FeeBalance.Valid && tx.FeeBalance.BigInt.Cmp(common.Big0) > 0 {
		payout.FeeBalance = tx.FeeBalance
		payout.TxFees.BigInt.Sub(payout.TxFees.BigInt, tx.FeeBalance.BigInt)

		balanceOutput := &pooldb.BalanceOutput{
			ChainID:    payout.ChainID,
			MinerID:    payout.MinerID,
			InPayoutID: types.Uint64Ptr(payout.ID),

			Value:        payout.FeeBalance,
			PoolFees:     dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			ExchangeFees: dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)},
			Mature:       true,
		}

		err = pooldb.InsertBalanceOutputs(dbTx, balanceOutput)
		if err != nil {
			return err
		}

		// add balance sum for fee balance
		err = pooldb.InsertAddBalanceSums(dbTx, &pooldb.BalanceSum{
			MinerID: payout.MinerID,
			ChainID: payout.ChainID,

			MatureValue: payout.FeeBalance,
		})
		if err != nil {
			return err
		}
	}

	payout.Height = tx.Height
	payout.Confirmed = true

	cols := []string{"height", "tx_fees", "fee_balance", "confirmed"}
	err = pooldb.UpdatePayout(dbTx, payout, cols)
	if err != nil {
		return err
	}

	err = dbTx.SafeCommit()
	if err != nil {
		return err
	}

	c.telegram.NotifyConfirmPayout(payout.ID)

	return nil
}

func (c *Client) FinalizePayouts(node types.PayoutNode) error {
	payouts, err := pooldb.GetUnconfirmedPayouts(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return err
	}

	for _, payout := range payouts {
		err = c.finalizePayout(payout)
		if err != nil {
			return err
		}
	}

	return nil
}

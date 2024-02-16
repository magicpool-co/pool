package payout

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/magicpool-co/pool/core/bank"
	"github.com/magicpool-co/pool/core/mailer"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/common"
	txCommon "github.com/magicpool-co/pool/pkg/crypto/tx"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

const (
	maxBatchSize = 100
)

type Client struct {
	pooldb   *dbcl.Client
	redis    *redis.Client
	telegram *telegram.Client
	bank     *bank.Client
	mailer   *mailer.Client
}

func New(
	pooldbClient *dbcl.Client,
	redisClient *redis.Client,
	telegramClient *telegram.Client,
	mailerClient *mailer.Client,
) *Client {
	client := &Client{
		pooldb:   pooldbClient,
		redis:    redisClient,
		telegram: telegramClient,
		bank:     bank.New(pooldbClient, redisClient, telegramClient),
		mailer:   mailerClient,
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

	payoutBoundStr := "1"
	if node.Chain() == "ETH" {
		payoutBoundStr = "2500000000000000"
	} else if node.Chain() == "BTC" {
		payoutBoundStr = "1000"
	}

	miners, err := pooldb.GetMinersWithBalanceAboveThresholdByChain(dbTx, node.Chain(), payoutBoundStr)
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

		valueSum, poolFeesSum := new(big.Int), new(big.Int)
		exchangeFeesSum, txFeesSum := new(big.Int), new(big.Int)
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
			if balanceOutput.TxFees.Valid {
				txFeesSum.Add(txFeesSum, balanceOutput.TxFees.BigInt)
			}

			balanceOutputIdx[miner.ID] = append(balanceOutputIdx[miner.ID], balanceOutput)
		}

		threshold := payoutBound.Default
		if miner.Threshold.Valid && miner.Threshold.BigInt.Cmp(common.Big0) > 0 {
			threshold = miner.Threshold.BigInt
		}

		if valueSum.Cmp(threshold) < 0 {
			return fmt.Errorf("miner %d not actually above threshold: %s < %s",
				miner.ID, valueSum, threshold)
		}

		balanceOutputSums[i] = &pooldb.BalanceOutput{
			MinerID: miner.ID,
			ChainID: node.Chain(),

			Value:        dbcl.NullBigInt{Valid: true, BigInt: valueSum},
			PoolFees:     dbcl.NullBigInt{Valid: true, BigInt: poolFeesSum},
			ExchangeFees: dbcl.NullBigInt{Valid: true, BigInt: exchangeFeesSum},
			TxFees:       dbcl.NullBigInt{Valid: true, BigInt: txFeesSum},
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
			TxFees:       balanceOutput.TxFees,

			Pending: true,
		}
	}

	bankLock, err := c.bank.FetchLock(node.Chain())
	if err != nil {
		return err
	}
	defer bankLock.Release(context.Background())

	switch node.GetAccountingType() {
	case types.AccountStructure:
		if len(payouts) > 15 {
			payouts = payouts[:15]
		}

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
			payout.TxFees.BigInt.Add(payout.TxFees.BigInt, outputList[i][0].Fee)
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
			c.telegram.NotifyInitiatePayout(payout.ID,
				payout.ChainID, payout.Address, explorerURL, floatValue)
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
		if err == txCommon.ErrTxTooBig && node.ShouldMergeUTXOs() {
			return c.bank.MergeUTXOs(node, 3)
		} else if err != nil {
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
			if len(payouts) != maxBatchSize && false {
				c.telegram.NotifyInitiatePayout(payoutID,
					payout.ChainID, payout.Address, explorerURL, floatValue)
			}
		}
	default:
		return nil
	}

	return dbTx.SafeCommit()
}

func (c *Client) finalizePayout(
	node types.PayoutNode,
	payout *pooldb.Payout,
	miner, emailAddress string,
) error {
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

	if c.mailer != nil && miner != "" && emailAddress != "" {
		units, err := common.GetDefaultUnits(payout.ChainID)
		if err != nil {
			return err
		}

		decimals := 4
		switch payout.ChainID {
		case "NEXA":
			decimals = 1
		case "KAS", "USDC", "USDT":
			decimals = 2
		case "BTC":
			decimals = 6
		}

		valueFloat := common.BigIntToFloat64(payout.Value.BigInt, units)
		valueStr := strconv.FormatFloat(valueFloat, 'f', decimals, 64) + " " + payout.ChainID

		explorerURL := node.GetTxExplorerURL(payout.TxID)
		err = c.mailer.SendEmailForPayout(emailAddress, miner,
			payout.TxID, explorerURL, valueStr, payout.CreatedAt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) FinalizePayouts(node types.PayoutNode) error {
	payouts, err := pooldb.GetUnconfirmedPayouts(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return err
	}

	minerIDIdx := make(map[uint64]bool)
	for _, payout := range payouts {
		minerIDIdx[payout.MinerID] = true
	}

	minerIDs := make([]uint64, 0)
	for minerID := range minerIDIdx {
		minerIDs = append(minerIDs, minerID)
	}

	miners, err := pooldb.GetMiners(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return err
	}

	minerIdx := make(map[uint64]*pooldb.Miner, 0)
	for _, miner := range miners {
		minerIdx[miner.ID] = miner
	}

	for _, payout := range payouts {
		var address, emailAddress string
		if miner, ok := minerIdx[payout.MinerID]; ok {
			if miner.EnabledPayoutNotifications && miner.Email != nil {
				address = miner.Address
				if parts := strings.Split(address, ":"); len(parts) == 1 {
					address = strings.ToLower(miner.ChainID) + ":" + address
				}

				emailAddress = types.StringValue(miner.Email)
			}
		}

		err = c.finalizePayout(node, payout, address, emailAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

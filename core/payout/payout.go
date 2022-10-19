package payout

import (
	"fmt"
	"math/big"
	"time"

	"github.com/magicpool-co/pool/core/bank"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

const (
	maxBatchSize = 15
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
		bank:     bank.New(pooldbClient, redisClient),
	}

	return client
}

func (c *Client) InitiatePayouts(node types.PayoutNode) error {
	defaultThreshold, err := common.GetDefaultPayoutThreshold(node.Chain())
	if err != nil {
		return err
	}

	balanceOutputs, err := pooldb.GetUnpaidBalanceOutputsAboveThreshold(c.pooldb.Reader(), node.Chain(), defaultThreshold.String())
	if err != nil {
		return err
	} else if len(balanceOutputs) == 0 {
		return nil
	}

	payouts := make([]*pooldb.Payout, len(balanceOutputs))
	for i, balanceOutput := range balanceOutputs {
		address, err := pooldb.GetMinerAddress(c.pooldb.Reader(), balanceOutput.MinerID)
		if err != nil {
			return err
		}

		feeBalance := new(big.Int)
		if balanceOutput.ChainID == "USDC" {
			feeBalance, err = pooldb.GetUnpaidBalanceOutputSumByMiner(c.pooldb.Reader(), balanceOutput.MinerID, "ETH")
			if err != nil {
				return err
			}
		}

		payouts[i] = &pooldb.Payout{
			ChainID: balanceOutput.ChainID,
			MinerID: balanceOutput.MinerID,
			Address: address,

			Value:        balanceOutput.Value,
			FeeBalance:   dbcl.NullBigInt{Valid: true, BigInt: feeBalance},
			PoolFees:     balanceOutput.PoolFees,
			ExchangeFees: balanceOutput.ExchangeFees,
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

		txs, err := c.bank.PrepareOutgoingTxs(node, outputList...)
		if err != nil {
			return err
		} else if len(txs) != len(payouts) {
			return fmt.Errorf("tx and payout count mismatch: %d and %d", len(txs), len(payouts))
		}

		for i, payout := range payouts {
			if txs[i] == nil {
				continue
			}

			payouts[i].TransactionID = txs[i].ID
			payouts[i].TxID = txs[i].TxID
			payoutID, err := pooldb.InsertPayout(c.pooldb.Writer(), payouts[i])
			if err != nil {
				return err
			}

			err = pooldb.UpdateBalanceOutputsSetOutPayoutID(c.pooldb.Writer(), payoutID, payouts[i].MinerID, payouts[i].ChainID)
			if err != nil {
				return err
			}
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

		txs, err := c.bank.PrepareOutgoingTxs(node, outputs)
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

		for i, payout := range payouts {
			payouts[i].TransactionID = tx.ID
			payouts[i].TxID = tx.TxID
			payouts[i].TxFees = dbcl.NullBigInt{Valid: true, BigInt: feeIdx[payouts[i].Address]}
			payoutID, err := pooldb.InsertPayout(c.pooldb.Writer(), payout)
			if err != nil {
				return err
			}

			err = pooldb.UpdateBalanceOutputsSetOutPayoutID(c.pooldb.Writer(), payoutID, payout.MinerID, payout.ChainID)
			if err != nil {
				return err
			}
		}
	default:
		return nil
	}

	return nil
}

func (c *Client) FinalizePayouts(node types.PayoutNode) error {
	payouts, err := pooldb.GetUnconfirmedPayouts(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return err
	}

	for _, payout := range payouts {
		tx, err := node.GetTx(payout.TxID)
		if err != nil {
			return err
		} else if !tx.Confirmed {
			if time.Since(payout.CreatedAt) > time.Hour*24 {
				// @TODO: manage failed transactions
			}
			continue
		} else if tx.Value == nil {
			return fmt.Errorf("no value for tx %s", tx.Hash)
		} else if tx.Fee == nil {
			return fmt.Errorf("no fee for tx %s", tx.Hash)
		}

		var feeBalance dbcl.NullBigInt
		if tx.FeeBalance != nil && tx.FeeBalance.Cmp(common.Big0) > 0 {
			feeBalance = dbcl.NullBigInt{Valid: true, BigInt: tx.FeeBalance}
			payout.TxFees.BigInt.Sub(payout.TxFees.BigInt, tx.FeeBalance)

			utxo := &pooldb.UTXO{
				ChainID: node.Chain(),
				TxID:    payout.TxID,
				Index:   0,
				Value:   feeBalance,
				Spent:   false,
			}

			err = pooldb.InsertUTXOs(c.pooldb.Writer(), utxo)
			if err != nil {
				return err
			}

			balanceOutput := &pooldb.BalanceOutput{
				ChainID:    payout.ChainID,
				MinerID:    payout.MinerID,
				InPayoutID: types.Uint64Ptr(payout.ID),

				Value: feeBalance,
			}

			err = pooldb.InsertBalanceOutputs(c.pooldb.Writer(), balanceOutput)
			if err != nil {
				return err
			}
		}

		payout.Height = types.Uint64Ptr(tx.BlockNumber)
		payout.FeeBalance = feeBalance
		payout.Confirmed = true

		cols := []string{"height", "tx_fees", "fee_balance", "confirmed"}
		err = pooldb.UpdatePayout(c.pooldb.Writer(), payout, cols)
		if err != nil {
			return err
		}
	}

	return nil
}

package payout

import (
	"fmt"
	"math/big"
	"sort"

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
	dbTx, err := c.pooldb.Begin()
	if err != nil {
		return err
	}
	defer dbTx.SafeRollback()

	defaultThreshold, err := common.GetDefaultPayoutThreshold(node.Chain())
	if err != nil {
		return err
	}

	balanceOutputs, err := pooldb.GetUnpaidBalanceOutputsAboveThreshold(dbTx, node.Chain(), defaultThreshold.String())
	if err != nil {
		return err
	} else if len(balanceOutputs) == 0 {
		return nil
	}

	balanceOutputIdx := make(map[uint64][]*pooldb.BalanceOutput)
	for _, balanceOutput := range balanceOutputs {
		if _, ok := balanceOutputIdx[balanceOutput.MinerID]; !ok {
			balanceOutputIdx[balanceOutput.MinerID] = make([]*pooldb.BalanceOutput, 0)
		}
		balanceOutputIdx[balanceOutput.MinerID] = append(balanceOutputIdx[balanceOutput.MinerID], balanceOutput)
	}

	balanceOutputSums := make([]*pooldb.BalanceOutput, 0)
	for minerID, minerBalanceOutputs := range balanceOutputIdx {
		if len(minerBalanceOutputs) == 0 {
			return fmt.Errorf("no balance outputs for miner %d", minerID)
		} else if len(minerBalanceOutputs) > 1 {
			for _, minerBalanceOutput := range minerBalanceOutputs[1:] {
				if !minerBalanceOutput.Value.Valid {
					return fmt.Errorf("no value for balance output %d", minerBalanceOutput.ID)
				} else if !minerBalanceOutput.PoolFees.Valid {
					return fmt.Errorf("no pool fees for balance output %d", minerBalanceOutput.ID)
				} else if !minerBalanceOutput.ExchangeFees.Valid {
					return fmt.Errorf("no exchange fees for balance output %d", minerBalanceOutput.ID)
				}
				minerBalanceOutputs[0].Value.BigInt.Add(minerBalanceOutputs[0].Value.BigInt, minerBalanceOutput.Value.BigInt)
				minerBalanceOutputs[0].PoolFees.BigInt.Add(minerBalanceOutputs[0].PoolFees.BigInt, minerBalanceOutput.PoolFees.BigInt)
				minerBalanceOutputs[0].ExchangeFees.BigInt.Add(minerBalanceOutputs[0].ExchangeFees.BigInt, minerBalanceOutput.ExchangeFees.BigInt)
			}
		}
		balanceOutputSums = append(balanceOutputSums, minerBalanceOutputs[0])
	}

	sort.Slice(balanceOutputSums, func(i, j int) bool {
		return balanceOutputSums[i].ID < balanceOutputSums[j].ID
	})

	payouts := make([]*pooldb.Payout, len(balanceOutputSums))
	for i, balanceOutput := range balanceOutputSums {
		address, err := pooldb.GetMinerAddress(dbTx, balanceOutput.MinerID)
		if err != nil {
			return err
		}

		feeBalance := new(big.Int)
		if balanceOutput.ChainID == "USDC" {
			feeBalance, err = pooldb.GetUnpaidBalanceOutputSumByMiner(dbTx, balanceOutput.MinerID, "ETH")
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

		txs, err := c.bank.PrepareOutgoingTxs(dbTx, node, outputList...)
		if err != nil {
			return err
		} else if len(txs) != len(payouts) {
			return fmt.Errorf("tx and payout count mismatch: %d and %d", len(txs), len(payouts))
		}

		for i, payout := range payouts {
			if txs[i] == nil {
				continue
			}

			payout.TransactionID = txs[i].ID
			payout.TxID = txs[i].TxID
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

		txs, err := c.bank.PrepareOutgoingTxs(dbTx, node, outputs)
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
			payoutID, err := pooldb.InsertPayout(dbTx, payout)
			if err != nil {
				return err
			}

			for _, balanceOutput := range balanceOutputIdx[payouts[i].MinerID] {
				balanceOutput.OutPayoutID = types.Uint64Ptr(payoutID)
				err = pooldb.UpdateBalanceOutput(dbTx, balanceOutput, []string{"out_payout_id"})
				if err != nil {
					return err
				}
			}
		}
	default:
		return nil
	}

	return dbTx.SafeCommit()
}

func (c *Client) FinalizePayouts(node types.PayoutNode) error {
	payouts, err := pooldb.GetUnconfirmedPayouts(c.pooldb.Reader(), node.Chain())
	if err != nil {
		return err
	}

	for _, payout := range payouts {
		tx, err := pooldb.GetTransaction(c.pooldb.Reader(), payout.TransactionID)
		if err != nil {
			return err
		} else if !tx.Spent || !tx.Confirmed {
			continue
		} else if !tx.Value.Valid {
			return fmt.Errorf("no value for tx %s", tx.TxID)
		} else if !tx.Fee.Valid {
			return fmt.Errorf("no fee for tx %s", tx.TxID)
		}

		if tx.FeeBalance.Valid && tx.FeeBalance.BigInt.Cmp(common.Big0) > 0 {
			payout.FeeBalance = tx.FeeBalance
			payout.TxFees.BigInt.Sub(payout.TxFees.BigInt, tx.FeeBalance.BigInt)

			utxo := &pooldb.UTXO{
				ChainID: node.Chain(),
				TxID:    payout.TxID,
				Index:   0,
				Value:   payout.FeeBalance,
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

				Value: payout.FeeBalance,
			}

			err = pooldb.InsertBalanceOutputs(c.pooldb.Writer(), balanceOutput)
			if err != nil {
				return err
			}
		}

		payout.Height = tx.Height
		payout.Confirmed = true

		cols := []string{"height", "tx_fees", "fee_balance", "confirmed"}
		err = pooldb.UpdatePayout(c.pooldb.Writer(), payout, cols)
		if err != nil {
			return err
		}
	}

	return nil
}

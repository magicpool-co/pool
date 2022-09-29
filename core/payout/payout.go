package payout

import (
	"fmt"
	"math/big"
	"time"

	"github.com/magicpool-co/pool/core/bank"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type Client struct {
	pooldb *dbcl.Client
	nodes  map[string]types.PayoutNode
}

func New(pooldbClient *dbcl.Client, nodes map[string]types.PayoutNode) (*Client, error) {
	client := &Client{
		pooldb: pooldbClient,
		nodes:  nodes,
	}

	return client, nil
}

func (c *Client) InitiatePayouts() error {
	for _, node := range c.nodes {
		defaultThreshold, err := common.GetDefaultPayoutThreshold(node.Chain())
		if err != nil {
			return err
		}

		balanceOutputs, err := pooldb.GetUnpaidBalanceOutputsAboveThreshold(c.pooldb.Reader(), node.Chain(), defaultThreshold.String())
		if err != nil {
			return err
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
			for _, payout := range payouts {
				if !payout.Value.Valid {
					return fmt.Errorf("no value for payout %d", payout.ID)
				} else if !payout.FeeBalance.Valid {
					return fmt.Errorf("no fee balance for payout %d", payout.ID)
				}

				outputs := []*types.TxOutput{
					&types.TxOutput{
						Address:    payout.Address,
						Value:      payout.Value.BigInt,
						FeeBalance: payout.FeeBalance.BigInt,
					},
				}

				payout.TxID, err = bank.SendOutgoingTx(node, c.pooldb, outputs)
				if err != nil {
					return err
				}

				payoutID, err := pooldb.InsertPayout(c.pooldb.Writer(), payout)
				if err != nil {
					return err
				}

				err = pooldb.UpdateBalanceOutputsSetOutPayoutID(c.pooldb.Writer(), payoutID, payout.MinerID, payout.ChainID)
				if err != nil {
					return err
				}
			}
		case types.UTXOStructure:
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

			txid, err := bank.SendOutgoingTx(node, c.pooldb, outputs)
			if err != nil {
				return err
			}

			for i, payout := range payouts {
				payouts[i].TxID = txid

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
	}
	return nil
}

func (c *Client) FinalizePayouts() error {
	for _, node := range c.nodes {
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
			if tx.FeeBalance != nil {
				feeBalance = dbcl.NullBigInt{Valid: true, BigInt: tx.FeeBalance}
				// @TODO: insert UTXO for remaining fee balance, add a new balance
				// output for the fee balance back to the miner
			}

			payout.Height = types.Uint64Ptr(tx.BlockNumber)
			payout.Value = dbcl.NullBigInt{Valid: true, BigInt: tx.Value}
			payout.TxFees = dbcl.NullBigInt{Valid: true, BigInt: tx.Fee}
			payout.FeeBalance = feeBalance
			payout.Confirmed = true

			cols := []string{"height", "value", "tx_fees", "fee_balance", "confirmed"}
			err = pooldb.UpdatePayout(c.pooldb.Writer(), payout, cols)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

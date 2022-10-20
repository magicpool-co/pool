package audit

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func CheckWallet(pooldbClient *dbcl.Client, node types.PayoutNode) error {
	chain := node.Chain()
	walletBalance, err := node.GetBalance()
	if err != nil {
		return err
	}

	utxoBalance, err := pooldb.GetSumUnspentUTXOValueByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	// add immature round sum to UTXOs since they're only added at the point of maturation
	// (ERGO is excluded since blocks are not shown in the wallet balance until they're mature)
	if chain != "ERGO" {
		immatureRoundSum, err := pooldb.GetSumImmatureRoundValueByChain(pooldbClient.Reader(), chain)
		if err != nil {
			return err
		}
		utxoBalance.Add(utxoBalance, immatureRoundSum)

		unconfirmedTxValue, err := pooldb.GetUnconfirmedTransactionSum(pooldbClient.Reader(), chain)
		if err != nil {
			return err
		}
		utxoBalance.Add(utxoBalance, unconfirmedTxValue)
	}

	if walletBalance.Cmp(utxoBalance) != 0 {
		return fmt.Errorf("mismatch for utxo and wallet: have %s, want %s", utxoBalance, walletBalance)
	}

	inputBalance, err := pooldb.GetPendingBalanceInputSumWithoutBatchByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	outputBalance, err := pooldb.GetUnpaidBalanceOutputSumByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	sumMinerBalance := new(big.Int).Add(inputBalance, outputBalance)

	// add unspent round sum to sum miner balance since they're only added at the point
	// of maturation (ERGO is excluded the immature round sum is excluded for UTXOs)
	if chain != "ERGO" {
		unspentRoundSum, err := pooldb.GetSumUnspentRoundValueByChain(pooldbClient.Reader(), chain)
		if err != nil {
			return err
		}
		sumMinerBalance.Add(sumMinerBalance, unspentRoundSum)

		unconfirmedPayoutValue, err := pooldb.GetUnconfirmedPayoutSum(pooldbClient.Reader(), chain)
		if err != nil {
			return err
		}
		sumMinerBalance.Sub(sumMinerBalance, unconfirmedPayoutValue)
	}

	if utxoBalance.Cmp(sumMinerBalance) != 0 {
		return fmt.Errorf("mismatch for miner sum and utxo: have %s, want %s", sumMinerBalance, utxoBalance)
	}

	return nil
}

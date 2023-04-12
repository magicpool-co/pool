package audit

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func chainIncludesImmature(chain string) bool {
	switch chain {
	case "ERG":
		return false
	default:
		return true
	}
}

func CheckWallet(pooldbClient *dbcl.Client, node types.PayoutNode) error {
	chain := node.Chain()
	walletBalance, err := node.GetBalance()
	if err != nil {
		return err
	} else if chain == "ERG" {
		// subtract the 1 ERG that is sitting as excess in the wallet
		walletBalance.Sub(walletBalance, new(big.Int).SetUint64(1_000_000_000))
	} else if chain == "KAS" {
		// add the extra KAS for blocks we missed
		walletBalance.Sub(walletBalance, new(big.Int).SetUint64(114382086903+46725396150))
	}

	utxoBalance, err := pooldb.GetSumUnspentUTXOValueByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	// add immature round sum to UTXOs since they're only added at the point of maturation
	// (if the chain shows blocks in the wallet balance before they're mature)
	if chainIncludesImmature(chain) {
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

	// add immature round sum to sum miner balance since they're only added at the point
	// of maturation (if the immature round sum is included beforehand too)
	if chainIncludesImmature(chain) {
		unspentRoundSum, err := pooldb.GetSumImmatureRoundValueByChain(pooldbClient.Reader(), chain)
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

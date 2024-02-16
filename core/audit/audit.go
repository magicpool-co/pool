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
	}

	utxoBalance, err := pooldb.GetSumUnspentUTXOValueByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	immatureBalance, err := pooldb.GetImmatureBalanceInputSumByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	pendingBalance, err := pooldb.GetPendingBalanceInputSumByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	unpaidBalance, err := pooldb.GetUnpaidBalanceOutputSumByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	// add immature round sum to UTXOs since they're only added at the point of maturation
	// (if the chain shows blocks in the wallet balance before they're mature)
	if chainIncludesImmature(chain) {
		unconfirmedTxValue, err := pooldb.GetUnconfirmedTransactionSum(pooldbClient.Reader(), chain)
		if err != nil {
			return err
		}

		utxoBalance.Add(utxoBalance, immatureBalance)
		utxoBalance.Add(utxoBalance, unconfirmedTxValue)
	}

	if walletBalance.Cmp(utxoBalance) != 0 {
		return fmt.Errorf("mismatch for utxo and wallet: have %s, want %s", utxoBalance, walletBalance)
	}

	sumMinerBalance := new(big.Int).Add(pendingBalance, unpaidBalance)

	// add immature round sum to sum miner balance since they're only added at the point
	// of maturation (if the immature round sum is included beforehand too)
	if chainIncludesImmature(chain) {
		unconfirmedPayoutValue, err := pooldb.GetUnconfirmedPayoutSum(pooldbClient.Reader(), chain)
		if err != nil {
			return err
		}

		sumMinerBalance.Add(sumMinerBalance, immatureBalance)
		sumMinerBalance.Sub(sumMinerBalance, unconfirmedPayoutValue)
	}

	if utxoBalance.Cmp(sumMinerBalance) != 0 {
		return fmt.Errorf("mismatch for miner sum and utxo: have %s, want %s", sumMinerBalance, utxoBalance)
	}

	return nil
}

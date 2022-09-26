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

	if walletBalance.Cmp(utxoBalance) != 0 {
		return fmt.Errorf("mismatch for utxo and wallet: have %s, want %s", utxoBalance, walletBalance)
	}

	immatureBalance, err := pooldb.GetSumImmatureRoundValueByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	inputBalance, err := pooldb.GetSumBalanceInputValueByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	outputBalance, err := pooldb.GetSumBalanceOutputValueByChain(pooldbClient.Reader(), chain)
	if err != nil {
		return err
	}

	sumMinerBalance := new(big.Int).Add(immatureBalance, inputBalance)
	sumMinerBalance.Add(sumMinerBalance, outputBalance)

	if utxoBalance.Cmp(sumMinerBalance) != 0 {
		return fmt.Errorf("mismatch for miner sum and utxo: have %s, want %s", sumMinerBalance, utxoBalance)
	}

	return nil
}

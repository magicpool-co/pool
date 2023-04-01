package credit

import (
	"fmt"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func matureRound(node types.MiningNode, pooldbClient *dbcl.Client, round *pooldb.Round) error {
	tx, err := pooldbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.SafeRollback()

	utxos, err := node.MatureRound(round)
	if err != nil {
		return err
	}

	cols := []string{"uncle", "orphan", "pending", "mature", "spent", "height",
		"epoch_height", "uncle_height", "hash", "coinbase_txid", "value", "created_at"}
	err = pooldb.UpdateRound(tx, round, cols)
	if err != nil {
		return err
	}

	for _, utxo := range utxos {
		err = pooldb.InsertUTXOs(tx, utxo)
		if err != nil {
			return fmt.Errorf("failed on round: %d: %s: %v", round.ID, utxo.TxID, err)
		}
	}

	if err := pooldb.UpdateBalanceInputsSetMatureByRound(tx, round.ID); err != nil {
		return err
	} else if err := pooldb.UpdateBalanceOutputsSetMatureByRound(tx, round.ID); err != nil {
		return err
	}

	return tx.SafeCommit()
}

func UnlockRounds(node types.MiningNode, pooldbClient *dbcl.Client) error {
	height, _, err := node.GetStatus()
	if err != nil {
		return err
	}

	pendingRounds, err := pooldb.GetPendingRoundsByChain(pooldbClient.Reader(), node.Chain(), height-node.GetImmatureDepth())
	if err != nil {
		return err
	}

	for _, round := range pendingRounds {
		err := node.UnlockRound(round)
		if err != nil {
			return err
		}

		cols := []string{"uncle", "orphan", "pending", "mature", "spent", "height",
			"epoch_height", "uncle_height", "hash", "coinbase_txid", "value", "created_at"}
		err = pooldb.UpdateRound(pooldbClient.Writer(), round, cols)
		if err != nil {
			return err
		}
	}

	immatureRounds, err := pooldb.GetImmatureRoundsByChain(pooldbClient.Reader(), node.Chain(), height-node.GetMatureDepth())
	if err != nil {
		return err
	}

	for _, round := range immatureRounds {
		err := matureRound(node, pooldbClient, round)
		if err != nil {
			return err
		}
	}

	return nil
}

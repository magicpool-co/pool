package credit

import (
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func UnlockRounds(node types.MiningNode, pooldbClient *dbcl.Client) error {
	height, syncing, err := node.GetStatus()
	if err != nil {
		return err
	} else if syncing {
		return nil
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
			"epoch_height", "uncle_height", "hash", "value", "created_at"}
		err = pooldb.UpdateRound(pooldbClient.Writer(), round, cols)
		if err != nil {
			return err
		}
	}

	immatureRounds, err := pooldb.GetImmatureRoundsByChain(pooldbClient.Reader(), node.Chain(), height-node.GetMatureDepth())
	if err != nil {
		return err
	}

	tx, err := pooldbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.SafeRollback()

	utxos := make([]*pooldb.UTXO, len(immatureRounds))
	for i, round := range immatureRounds {
		round.Mature = true
		err = pooldb.UpdateRound(tx, round, []string{"mature"})
		if err != nil {
			return err
		}

		var utxoHash string
		if round.CoinbaseTxID != nil {
			utxoHash = types.StringValue(round.CoinbaseTxID)
		} else {
			utxoHash = round.Hash
		}

		utxos[i] = &pooldb.UTXO{
			ChainID: round.ChainID,

			Value: round.Value,
			TxID:  utxoHash,
			Index: 0,
			Spent: false,
		}
	}

	err = pooldb.InsertUTXOs(tx, utxos...)
	if err != nil {
		return err
	}

	return tx.SafeCommit()
}
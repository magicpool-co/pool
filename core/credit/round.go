package credit

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/accounting"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

func CreditRound(pooldbClient *dbcl.Client, round *pooldb.Round, shares []*pooldb.Share) error {
	// create a miner index of proportional share values
	minerIdx := make(map[uint64]uint64)
	for _, share := range shares {
		minerIdx[share.MinerID] += share.Count
	}

	// fetch the recipients and create a recipient index of proportional fee values
	recipients, err := pooldb.GetRecipients(pooldbClient.Reader())
	if err != nil {
		return err
	}

	recipientIdx := make(map[uint64]uint64)
	for _, recipient := range recipients {
		if recipient.RecipientFeePercent == nil {
			return fmt.Errorf("no recipient fee set for %d", recipient.ID)
		}
		recipientIdx[recipient.ID] += types.Uint64Value(recipient.RecipientFeePercent)
	}

	// distribute the proceeds to miners and recipients
	compoundValues, err := accounting.CreditRound(round.Value.BigInt, minerIdx, recipientIdx)
	if err != nil {
		return err
	}

	// fetch miners and recipients to check their payout chain
	compoundIDs := make([]uint64, 0)
	for compoundID := range compoundValues {
		compoundIDs = append(compoundIDs, compoundID)
	}

	miners, err := pooldb.GetMiners(pooldbClient.Reader(), compoundIDs)
	if err != nil {
		return err
	}

	// create compound index for miners and recipients
	compoundIdx := make(map[uint64]*pooldb.Miner)
	for _, miner := range miners {
		compoundIdx[miner.ID] = miner
	}

	// process the balance inputs for all round distributions
	usedValue := new(big.Int)
	inputs := make([]*pooldb.BalanceInput, 0)
	for minerID, value := range compoundValues {
		miner, ok := compoundIdx[minerID]
		if !ok {
			return fmt.Errorf("no miner found for %d", minerID)
		} else if value == nil || value.Cmp(common.Big0) == 0 {
			continue
		}
		usedValue.Add(usedValue, value)

		input := &pooldb.BalanceInput{
			RoundID: round.ID,
			ChainID: round.ChainID,
			MinerID: miner.ID,

			Value:   dbcl.NullBigInt{Valid: true, BigInt: value},
			Pending: miner.ChainID != round.ChainID,
		}
		inputs = append(inputs, input)
		delete(compoundIdx, miner.ID)
		delete(compoundValues, miner.ID)
	}

	if len(compoundIdx) > 0 {
		return fmt.Errorf("unable to find %d miners in idx", len(compoundIdx))
	} else if len(compoundValues) > 0 {
		return fmt.Errorf("unable to find %d miners in values", len(compoundValues))
	} else if usedValue.Cmp(round.Value.BigInt) != 0 {
		return fmt.Errorf("crediting mismatch: have %s, want %s", usedValue, round.Value.BigInt)
	}

	tx, err := pooldbClient.Begin()
	if err != nil {
		return err
	}
	defer tx.SafeRollback()

	round.Spent = true
	if err := pooldb.InsertBalanceInputs(tx, inputs...); err != nil {
		return err
	} else if pooldb.UpdateRound(tx, round, []string{"spent"}); err != nil {
		return err
	}

	return tx.SafeCommit()
}

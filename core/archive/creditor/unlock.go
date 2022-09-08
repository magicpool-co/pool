package creditor

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/magicpool-co/pool/core/banker"
	"github.com/magicpool-co/pool/pkg/config"
)

func UnlockBlocks(conf *config.Config) error {
	for ticker, node := range conf.MiningNodes {
		start := time.Now()

		currentHeight, err := node.GetStatus()
		if err != nil {
			return err
		}

		pendingHeight := currentHeight - uint64(node.GetImmatureDepth())
		pendingRounds, err := conf.DB.GetPendingRounds(ticker, pendingHeight)
		if err != nil {
			return err
		}

		for _, round := range pendingRounds {
			if err := node.UnlockRound(round); err != nil {
				return err
			}
		}

		usdPrice, err := conf.Bittrex.GetRate(ticker, "USD")
		if err != nil {
			return err
		}

		for _, round := range pendingRounds {
			round.CostBasisPrice = sql.NullFloat64{Valid: true, Float64: usdPrice}
			cols := []string{"hash", "value", "mev_value", "cost_basis_price", "pending",
				"mature", "uncle", "spent", "uncle_height", "orphan"}
			if err := conf.DB.UpdateRound(round, cols); err != nil {
				return err
			}
		}

		end := time.Now().Sub(start)
		conf.Logger.Debug(fmt.Sprintf("unlocked %d blocks in %s\n", len(pendingRounds), end))

		matureHeight := currentHeight - uint64(node.GetMatureDepth())
		immatureRounds, err := conf.DB.GetImmatureRounds(ticker, matureHeight)
		if err != nil {
			return err
		}

		for _, round := range immatureRounds {
			round.Mature = true

			if err := banker.AddRound(conf, round); err != nil {
				return err
			}

			if err := conf.DB.UpdateRound(round, []string{"mature"}); err != nil {
				return err
			}
		}
	}

	return nil
}

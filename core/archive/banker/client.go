package banker

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/magicpool-co/pool/pkg/config"
	"github.com/magicpool-co/pool/pkg/db"
	"github.com/magicpool-co/pool/pkg/types"
)

func validateConfig(conf *config.Config) error {
	if conf.DB == nil {
		return fmt.Errorf("DB is nil for banker")
	} else if conf.PayoutNodes == nil {
		return fmt.Errorf("PayoutNodes is nil for banker")
	}

	return nil
}

func AddRound(conf *config.Config, round *db.Round) error {
	if err := validateConfig(conf); err != nil {
		return err
	}

	var inTxID string
	if round.CoinbaseTxID.Valid {
		inTxID = round.CoinbaseTxID.String
	} else if round.Hash.Valid {
		inTxID = round.Hash.String
	} else {
		return fmt.Errorf("no valid coinbase txid or hash in round %d", round.ID)
	}

	value := new(big.Int)
	if round.Value.Valid {
		value.Add(value, round.Value.BigInt)
	}

	if round.MEVValue.Valid {
		value.Add(value, round.MEVValue.BigInt)
	}

	utxo := &db.UTXO{
		CoinID:    round.CoinID,
		Value:     db.NullBigInt{Valid: true, BigInt: value},
		Spent:     false,
		InTxID:    inTxID,
		InIndex:   0,
		InRoundID: sql.NullInt64{int64(round.ID), true},
	}

	if _, err := conf.DB.InsertUTXO(utxo); err != nil {
		return err
	}

	return nil
}

func AddDepositChange(dbClient *db.DBClient, deposit *db.SwitchDeposit, change *big.Int) error {
	utxo := &db.UTXO{
		CoinID:      deposit.CoinID,
		Value:       db.NullBigInt{change, true},
		Spent:       false,
		InTxID:      deposit.TxID,
		InIndex:     0,
		InDepositID: sql.NullInt64{int64(deposit.ID), true},
	}

	if _, err := dbClient.InsertUTXO(utxo); err != nil {
		return err
	}

	return nil
}

func AddWithdrawal(dbClient *db.DBClient, withdrawal *db.SwitchWithdrawal, tx *types.RawTx, address string) error {
	utxos := make([]*db.UTXO, 0)
	if withdrawal.CoinID == "BTC" {
		for _, output := range tx.Outputs {
			if strings.ToLower(output.Address) == strings.ToLower(address) {
				utxo := &db.UTXO{
					CoinID:         withdrawal.CoinID,
					Value:          db.NullBigInt{new(big.Int).SetUint64(output.Value), true},
					Spent:          false,
					InTxID:         output.Hash,
					InIndex:        output.Index,
					InWithdrawalID: sql.NullInt64{int64(withdrawal.ID), true},
				}

				utxos = append(utxos, utxo)
			}
		}
	} else {
		utxo := &db.UTXO{
			CoinID:         withdrawal.CoinID,
			Value:          withdrawal.Value,
			Spent:          false,
			InTxID:         withdrawal.TxID.String,
			InIndex:        0,
			InWithdrawalID: sql.NullInt64{int64(withdrawal.ID), true},
		}

		utxos = append(utxos, utxo)
	}

	for _, utxo := range utxos {
		if _, err := dbClient.InsertUTXO(utxo); err != nil {
			return err
		}
	}

	return nil
}

func AddPayoutChange(dbClient *db.DBClient, payout *db.Payout, change *big.Int) error {
	utxo := &db.UTXO{
		CoinID:     payout.FeeBalanceCoin,
		Value:      db.NullBigInt{change, true},
		Spent:      false,
		InTxID:     payout.TxID.String,
		InIndex:    0,
		InPayoutID: sql.NullInt64{int64(payout.ID), true},
	}

	if _, err := dbClient.InsertUTXO(utxo); err != nil {
		return err
	}

	return nil
}

func SpendDeposit(conf *config.Config, chain string, value *big.Int, address string) (*types.TxResponse, error) {
	err := validateConfig(conf)
	if err != nil {
		return nil, err
	}

	node, ok := conf.PayoutNodes[chain]
	if !ok {
		return nil, fmt.Errorf("unable to find node for %s", chain)
	}

	outputs := []*types.TxOutput{
		&types.TxOutput{
			Address:  address,
			Value:    value,
			SplitFee: true,
		},
	}

	tx, err := spend(conf.DB, node, chain, outputs)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func SpendPayouts(conf *config.Config, payouts []*db.Payout) (*types.TxResponse, error) {
	err := validateConfig(conf)
	if err != nil {
		return nil, err
	}

	var chain string
	for _, payout := range payouts {
		if len(chain) == 0 {
			chain = payout.CoinID
		} else if chain != payout.CoinID {
			return nil, fmt.Errorf("mismatch on payout chains: %s and %s", chain, payout.CoinID)
		}
	}

	// create the TxOutputs depending on the chain
	txOutputs := make([]*types.TxOutput, 0)
	switch chain {
	case "RVN", "BTC":
		for _, payout := range payouts {
			output := &types.TxOutput{
				Address:  payout.Address,
				Value:    payout.Value.BigInt,
				SplitFee: true,
			}

			txOutputs = append(txOutputs, output)
		}
	case "ETH", "ETC", "USDC":
		output := &types.TxOutput{
			Address:    payouts[0].Address,
			Value:      payouts[0].Value.BigInt,
			FeeBalance: payouts[0].InFeeBalance.BigInt,
		}

		txOutputs = append(txOutputs, output)
	}

	node, ok := conf.PayoutNodes[chain]
	if !ok {
		return nil, fmt.Errorf("unable to find node for %s", chain)
	}

	tx, err := spend(conf.DB, node, chain, txOutputs)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func spend(dbClient *db.DBClient, node types.PayoutNode, chain string, txOutputs []*types.TxOutput) (*types.TxResponse, error) {
	txInputs := make([]*types.TxInput, 0)
	utxoInputs := make([]*db.UTXO, 0)
	utxoOutputs := make([]*db.UTXO, 0)

	// validate all outputs
	totalValue := new(big.Int)
	for i, output := range txOutputs {
		if output.Value == nil {
			return nil, fmt.Errorf("value for output %d is nil", i)
		}

		totalValue.Add(totalValue, output.Value)
		if output.FeeBalance != nil && chain == "ETH" {
			totalValue.Add(totalValue, output.FeeBalance)
		}
	}

	// aggregate all known utxo's
	utxoInputs, err := dbClient.GetUnspentUTXOsByCoin(chain)
	if err != nil {
		return nil, err
	}

	totalUtxoValue := new(big.Int)
	for _, utxo := range utxoInputs {
		if !utxo.Value.Valid {
			return nil, fmt.Errorf("invalid value for UTXO %d", utxo.ID)
		}

		totalUtxoValue.Add(totalUtxoValue, utxo.Value.BigInt)
	}

	// create the TxInputs depending on the coin
	switch chain {
	case "BTC", "RVN":
		for _, utxo := range utxoInputs {
			input := &types.TxInput{
				Hash:  utxo.InTxID,
				Index: utxo.InIndex,
				Value: utxo.Value.BigInt,
			}

			txInputs = append(txInputs, input)
		}
	case "ETH", "ETC", "USDC":
		if len(txOutputs) != 1 {
			return nil, fmt.Errorf("invalid number of outputs for %s: %d", chain, len(txOutputs))
		}

		input := &types.TxInput{
			Value:      txOutputs[0].Value,
			FeeBalance: txOutputs[0].FeeBalance,
		}

		txInputs = append(txInputs, input)

		// handle fee balance UTXOs
		if chain == "USDC" {
			feeBalance := txOutputs[0].FeeBalance
			ethUTXOs, err := dbClient.GetUnspentUTXOsByCoin("ETH")
			if err != nil {
				return nil, err
			}

			// sum existing ETH balances
			ethBalance := new(big.Int)
			for _, utxo := range ethUTXOs {
				ethBalance.Add(ethBalance, utxo.Value.BigInt)
			}

			// check for overspend on fee balance
			remainder := new(big.Int).Sub(ethBalance, feeBalance)
			if diff := remainder.Cmp(new(big.Int)); diff < 0 {
				return nil, fmt.Errorf("ETH fee balance overspend: have %s, want %s", feeBalance, ethBalance)
			}

			// add remainder UTXO for fee balance
			utxo := &db.UTXO{
				CoinID:  "ETH",
				Value:   db.NullBigInt{remainder, true},
				Spent:   false,
				InIndex: 0,
			}

			// finalize utxo sets
			utxoInputs = append(utxoInputs, ethUTXOs...)
			utxoOutputs = append(utxoOutputs, utxo)
		}
	}

	// verify that there are inputs to send
	if len(txInputs) == 0 {
		return nil, fmt.Errorf("cannot send transaction for %s with 0 inputs", chain)
	}

	// verify that the UTXO value is greater than or equal to the value being sent
	remainder := new(big.Int).Sub(totalUtxoValue, totalValue)
	if diff := remainder.Cmp(new(big.Int)); diff < 0 {
		return nil, fmt.Errorf("%s overspend: have %s, want %s", chain, totalValue, totalUtxoValue)
	} else if diff > 0 {
		// add a change UTXO for UTXO-based chains
		switch chain {
		case "BTC", "RVN":
			changeOutput := &types.TxOutput{
				Address:  node.GetWallet().Address,
				Value:    remainder,
				SplitFee: false,
			}

			txOutputs = append(txOutputs, changeOutput)
		}

		// add remainder UTXO
		utxo := &db.UTXO{
			CoinID:  chain,
			Value:   db.NullBigInt{remainder, true},
			Spent:   false,
			InIndex: uint32(len(txOutputs) - 1),
		}

		utxoOutputs = append(utxoOutputs, utxo)
	}

	// send the transaction
	tx, err := node.SendTX(txInputs, txOutputs)
	if err != nil {
		return nil, err
	}

	// remove the remainder on the tx value for UTXO-based chains
	switch chain {
	case "BTC", "RVN":
		tx.Value.Sub(tx.Value, remainder)
	}

	// mark all used UTXO's as spent
	for _, utxo := range utxoInputs {
		utxo.OutTxID = sql.NullString{tx.TxID, true}
		utxo.Spent = true
		cols := []string{"spent", "out_txid"}
		if err := dbClient.UpdateUTXO(utxo, cols); err != nil {
			return nil, err
		}
	}

	// insert the newly created UTXO's
	for _, utxo := range utxoOutputs {
		utxo.InTxID = tx.TxID
		if _, err := dbClient.InsertUTXO(utxo); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

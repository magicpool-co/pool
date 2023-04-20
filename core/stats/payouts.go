package stats

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
)

func getTxExplorerURL(chain, hash string) (string, error) {
	var explorerURL string
	var err error
	switch chain {
	case "AE":
		explorerURL = "https://explorer.aeternity.io/transactions/" + hash
	case "BTC":
		explorerURL = "https://blockchair.com/bitcoin/transaction/" + hash
	case "CFX":
		explorerURL = "https://www.confluxscan.io/tx/" + hash
	case "CTXC":
		explorerURL = "https://cerebro.cortexlabs.ai/#/tx/" + hash
	case "ERG":
		explorerURL = "https://explorer.ergoplatform.com/en/transactions/" + hash
	case "ETC":
		explorerURL = "https://blockscout.com/etc/mainnet/tx/" + hash
	case "ETH":
		explorerURL = "https://etherscan.io/tx/" + hash
	case "FIRO":
		explorerURL = "https://explorer.firo.org/tx/" + hash
	case "FLUX":
		explorerURL = "https://explorer.runonflux.io/tx/" + hash
	case "KAS":
		explorerURL = "https://katnip.kaspad.net/tx/" + hash
	case "NEXA":
		explorerURL = "https://explorer.nexa.org/tx/" + hash
	case "RVN":
		explorerURL = "https://ravencoin.network/tx/" + hash
	case "USDC":
		explorerURL = "https://etherscan.io/tx/" + hash
	default:
		err = fmt.Errorf("no tx explorer found for chain")
	}

	return explorerURL, err
}

func newPayout(dbPayout *pooldb.Payout, dbBalanceInputSums []*pooldb.BalanceInput) (*Payout, error) {
	if !dbPayout.Value.Valid {
		return nil, fmt.Errorf("no value for payout %d", dbPayout.ID)
	} else if !dbPayout.PoolFees.Valid {
		return nil, fmt.Errorf("no pool fees for payout %d", dbPayout.ID)
	} else if !dbPayout.ExchangeFees.Valid {
		return nil, fmt.Errorf("no exchange fees for payout %d", dbPayout.ID)
	} else if !dbPayout.TxFees.Valid {
		dbPayout.TxFees = dbcl.NullBigInt{Valid: true, BigInt: new(big.Int)}
	}

	totalFees := new(big.Int)
	totalFees.Add(totalFees, dbPayout.PoolFees.BigInt)
	totalFees.Add(totalFees, dbPayout.ExchangeFees.BigInt)
	totalFees.Add(totalFees, dbPayout.TxFees.BigInt)

	// aggregate balance input sums into an index
	var err error
	inputIdxFormatted := make(map[string]Number)
	for _, dbBalanceInputSum := range dbBalanceInputSums {
		chainID := dbBalanceInputSum.ChainID
		value := dbBalanceInputSum.Value.BigInt
		if !dbBalanceInputSum.Value.Valid || value.Cmp(common.Big0) == 0 {
			continue
		}

		inputIdxFormatted[chainID], err = newNumberFromBigInt(value, chainID)
		if err != nil {
			return nil, err
		}
	}

	// handle the native payout case to avoid extra db calls when we
	// can just calculate the number from value + tx fees
	if len(inputIdxFormatted) == 0 {
		inputValue := new(big.Int).Add(dbPayout.Value.BigInt, dbPayout.TxFees.BigInt)
		inputIdxFormatted[dbPayout.ChainID], err = newNumberFromBigInt(inputValue, dbPayout.ChainID)
		if err != nil {
			return nil, err
		}
	}

	valueFormatted, err := newNumberFromBigInt(dbPayout.Value.BigInt, dbPayout.ChainID)
	if err != nil {
		return nil, err
	}

	poolFeesFormatted, err := newNumberFromBigInt(dbPayout.PoolFees.BigInt, dbPayout.ChainID)
	if err != nil {
		return nil, err
	}

	exchangeFeesFormatted, err := newNumberFromBigInt(dbPayout.ExchangeFees.BigInt, dbPayout.ChainID)
	if err != nil {
		return nil, err
	}

	txFeesFormatted, err := newNumberFromBigInt(dbPayout.TxFees.BigInt, dbPayout.ChainID)
	if err != nil {
		return nil, err
	}

	totalFeesFormatted, err := newNumberFromBigInt(totalFees, dbPayout.ChainID)
	if err != nil {
		return nil, err
	}

	explorerURL, err := getTxExplorerURL(dbPayout.ChainID, dbPayout.TxID)
	if err != nil {
		return nil, err
	}

	payout := &Payout{
		Chain:        dbPayout.ChainID,
		Address:      dbPayout.Address,
		TxID:         dbPayout.TxID,
		ExplorerURL:  explorerURL,
		Confirmed:    dbPayout.Confirmed,
		Inputs:       inputIdxFormatted,
		Value:        valueFormatted,
		PoolFees:     poolFeesFormatted,
		ExchangeFees: exchangeFeesFormatted,
		TxFees:       txFeesFormatted,
		TotalFees:    totalFeesFormatted,
		Timestamp:    dbPayout.CreatedAt.Unix(),
	}

	return payout, nil
}

func (c *Client) GetGlobalPayouts(page, size uint64) ([]*Payout, uint64, error) {
	count, err := pooldb.GetPayoutsCount(c.pooldb.Reader())
	if err != nil {
		return nil, 0, err
	}

	dbPayouts, err := pooldb.GetPayouts(c.pooldb.Reader(), page, size)
	if err != nil {
		return nil, 0, err
	}

	payouts := make([]*Payout, len(dbPayouts))
	for i, dbPayout := range dbPayouts {
		payouts[i], err = newPayout(dbPayout, nil)
		if err != nil {
			return nil, 0, err
		}
	}

	return payouts, count, nil
}

func (c *Client) GetMinerPayouts(minerIDs []uint64, page, size uint64) ([]*Payout, uint64, error) {
	if len(minerIDs) == 0 {
		return nil, 0, nil
	}

	count, err := pooldb.GetPayoutsByMinersCount(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return nil, 0, err
	}

	dbPayouts, err := pooldb.GetPayoutsByMiners(c.pooldb.Reader(), minerIDs, page, size)
	if err != nil {
		return nil, 0, err
	}

	switchPayoutIDs := make([]uint64, 0)
	for _, dbPayout := range dbPayouts {
		if dbPayout.ChainID == "BTC" || dbPayout.ChainID == "ETH" {
			switchPayoutIDs = append(switchPayoutIDs, dbPayout.ID)
		}
	}

	balanceInputSums, err := pooldb.GetPayoutBalanceInputSums(c.pooldb.Reader(), switchPayoutIDs)
	if err != nil {
		return nil, 0, err
	}

	balanceInputSumIdx := make(map[uint64][]*pooldb.BalanceInput)
	for _, balanceInputSum := range balanceInputSums {
		payoutID := balanceInputSum.PayoutID
		if payoutID == 0 {
			continue
		} else if _, ok := balanceInputSumIdx[payoutID]; !ok {
			balanceInputSumIdx[payoutID] = make([]*pooldb.BalanceInput, 0)
		}

		balanceInputSumIdx[payoutID] = append(balanceInputSumIdx[payoutID], balanceInputSum)
	}

	payouts := make([]*Payout, len(dbPayouts))
	for i, dbPayout := range dbPayouts {
		payouts[i], err = newPayout(dbPayout, balanceInputSumIdx[dbPayout.ID])
		if err != nil {
			return nil, 0, err
		}
	}

	return payouts, count, nil
}

func (c *Client) GetAllMinerPayouts(minerIDs uint64) ([]*Payout, error) {
	payouts, _, err := c.GetMinerPayouts([]uint64{minerIDs}, 0, 99999999999)
	return payouts, err
}

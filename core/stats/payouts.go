package stats

import (
	"fmt"
	"math/big"

	"github.com/magicpool-co/pool/internal/pooldb"
)

func getTxExplorerURL(chain, hash string) (string, error) {
	var explorerURL string
	var err error
	switch chain {
	case "AE":
		explorerURL = "https://explorer.aeternity.io/transactions/" + hash
	case "CFX":
		explorerURL = "https://www.confluxscan.io/tx/" + hash
	case "CTXC":
		explorerURL = "https://cerebro.cortexlabs.ai/#/tx/" + hash
	case "ERGO":
		explorerURL = "https://explorer.ergoplatform.com/en/transactions/" + hash
	case "ETC":
		explorerURL = "https://blockscout.com/etc/mainnet/tx/" + hash
	case "FIRO":
		explorerURL = "https://explorer.firo.org/tx/" + hash
	case "FLUX":
		explorerURL = "https://explorer.runonflux.io/tx/" + hash
	case "RVN":
		explorerURL = "https://ravencoin.network/tx/" + hash
	default:
		err = fmt.Errorf("no tx explorer found for chain")
	}

	return explorerURL, err
}

func newPayout(dbPayout *pooldb.Payout) (*Payout, error) {
	if !dbPayout.Value.Valid {
		return nil, fmt.Errorf("no value for payout %d", dbPayout.ID)
	} else if !dbPayout.PoolFees.Valid {
		return nil, fmt.Errorf("no pool fees for payout %d", dbPayout.ID)
	} else if !dbPayout.ExchangeFees.Valid {
		return nil, fmt.Errorf("no exchange fees for payout %d", dbPayout.ID)
	} else if !dbPayout.TxFees.Valid {
		return nil, fmt.Errorf("no tx fees for payout %d", dbPayout.ID)
	}

	totalFees := new(big.Int)
	totalFees.Add(totalFees, dbPayout.Value.BigInt)
	totalFees.Add(totalFees, dbPayout.PoolFees.BigInt)
	totalFees.Add(totalFees, dbPayout.ExchangeFees.BigInt)
	totalFees.Add(totalFees, dbPayout.TxFees.BigInt)

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
		payouts[i], err = newPayout(dbPayout)
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

	count, err := pooldb.GetPayoutsByMinerCount(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return nil, 0, err
	}

	dbPayouts, err := pooldb.GetPayoutsByMiner(c.pooldb.Reader(), minerIDs, page, size)
	if err != nil {
		return nil, 0, err
	}

	payouts := make([]*Payout, len(dbPayouts))
	for i, dbPayout := range dbPayouts {
		payouts[i], err = newPayout(dbPayout)
		if err != nil {
			return nil, 0, err
		}
	}

	return payouts, count, nil
}

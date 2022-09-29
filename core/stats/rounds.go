package stats

import (
	"fmt"
	"strconv"

	"github.com/magicpool-co/pool/internal/pooldb"
)

func getBlockExplorerURL(chain, hash string, height uint64) (string, error) {
	heightStr := strconv.FormatUint(height, 10)

	var explorerURL string
	var err error
	switch chain {
	case "AE":
		explorerURL = "https://explorer.aeternity.io/generations/" + heightStr
	case "CFX":
		explorerURL = "https://www.confluxscan.io/block/" + hash
	case "CTXC":
		explorerURL = "https://cerebro.cortexlabs.ai/#/block/" + heightStr
	case "ERGO":
		explorerURL = "https://explorer.ergoplatform.com/en/blocks/" + hash
	case "ETC":
		explorerURL = "https://blockscout.com/etc/mainnet/block/" + heightStr
	case "FIRO":
		explorerURL = "https://explorer.firo.org/block/" + hash
	case "FLUX":
		explorerURL = "https://explorer.runonflux.io/block/" + hash
	case "RVN":
		explorerURL = "https://ravencoin.network/block/" + hash
	default:
		err = fmt.Errorf("no block explorer found for chain")
	}

	return explorerURL, err
}

func newRound(dbRound *pooldb.Round) (*Round, error) {
	if !dbRound.Value.Valid {
		return nil, fmt.Errorf("no value for round %d", dbRound.ID)
	}

	var roundType string
	if dbRound.Pending {
		roundType = "pending"
	} else if dbRound.Orphan {
		roundType = "orphan"
	} else if dbRound.Uncle {
		roundType = "uncle"
	} else if dbRound.Mature {
		roundType = "block"
	} else {
		return nil, fmt.Errorf("unknown block status for round %d", dbRound.ID)
	}

	parsedValue, err := newNumberFromBigInt(dbRound.Value.BigInt, dbRound.ChainID)
	if err != nil {
		return nil, err
	}

	var parsedMinerValue, parsedMinerPercentage *Number
	if dbRound.MinerValue.Valid && dbRound.MinerAcceptedShares != 0 {
		parsedMinerValue, err = newNumberFromBigIntPtr(dbRound.MinerValue.BigInt, dbRound.ChainID)
		if err != nil {
			return nil, err
		}

		var minerPercentage float64
		if dbRound.AcceptedShares != 0 {
			minerPercentage = 100 * (float64(dbRound.MinerAcceptedShares) / float64(dbRound.AcceptedShares))
		}
		parsedMinerPercentage = newNumberFromFloat64Ptr(minerPercentage, "%", false)
	}

	explorerURL, err := getBlockExplorerURL(dbRound.ChainID, dbRound.Hash, dbRound.Height)
	if err != nil {
		return nil, err
	}

	round := &Round{
		Chain:           dbRound.ChainID,
		Type:            roundType,
		Pending:         dbRound.Pending,
		Mature:          dbRound.Mature,
		Hash:            dbRound.Hash,
		Height:          dbRound.Height,
		ExplorerURL:     explorerURL,
		Difficulty:      newNumberFromFloat64(float64(dbRound.Difficulty), "", true),
		Hashrate:        Number{},
		Luck:            newNumberFromFloat64(dbRound.Luck, "%", false),
		Value:           parsedValue,
		MinerValue:      parsedMinerValue,
		MinerPercentage: parsedMinerPercentage,
		Timestamp:       dbRound.CreatedAt.Unix(),
	}

	return round, nil
}

func (c *Client) GetGlobalRounds(page, size uint64) ([]*Round, uint64, error) {
	count, err := pooldb.GetRoundsCount(c.pooldb.Reader())
	if err != nil {
		return nil, 0, err
	}

	dbRounds, err := pooldb.GetRounds(c.pooldb.Reader(), page, size)
	if err != nil {
		return nil, 0, err
	}

	rounds := make([]*Round, len(dbRounds))
	for i, dbRound := range dbRounds {
		rounds[i], err = newRound(dbRound)
		if err != nil {
			return nil, 0, err
		}
	}

	return rounds, count, nil
}

func (c *Client) GetMinerRounds(minerIDs []uint64, page, size uint64) ([]*Round, uint64, error) {
	if len(minerIDs) == 0 {
		return nil, 0, nil
	}

	count, err := pooldb.GetRoundsByMinersCount(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return nil, 0, err
	}

	dbRounds, err := pooldb.GetRoundsByMiners(c.pooldb.Reader(), minerIDs, page, size)
	if err != nil {
		return nil, 0, err
	}

	rounds := make([]*Round, len(dbRounds))
	for i, dbRound := range dbRounds {
		rounds[i], err = newRound(dbRound)
		if err != nil {
			return nil, 0, err
		}
	}

	return rounds, count, nil
}

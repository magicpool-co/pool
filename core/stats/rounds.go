package stats

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/types"
)

func getBlockExplorerURL(chain, hash string, height uint64) (string, error) {
	heightStr := strconv.FormatUint(height, 10)

	var explorerURL string
	var err error
	switch chain {
	case "CFX":
		explorerURL = "https://www.confluxscan.io/block/" + hash
	case "ERG":
		explorerURL = "https://explorer.ergoplatform.com/en/blocks/" + hash
	case "ETC":
		explorerURL = "https://blockscout.com/etc/mainnet/block/" + heightStr
	case "FIRO":
		explorerURL = "https://explorer.firo.org/block/" + hash
	case "FLUX":
		explorerURL = "https://explorer.runonflux.io/block/" + hash
	case "KAS":
		explorerURL = "https://explorer.kaspa.org/blocks/" + hash
	case "NEXA":
		explorerURL = "https://explorer.nexa.org/block/" + hash
	case "RVN":
		explorerURL = "https://ravencoin.network/block/" + hash
	default:
		err = fmt.Errorf("no block explorer found for chain")
	}

	return explorerURL, err
}

func newRound(dbRound *pooldb.Round) (*Round, error) {
	value := new(big.Int)
	if dbRound.Value.Valid {
		value = value.Set(dbRound.Value.BigInt)
	}

	poolType := "pplns"
	if dbRound.Solo {
		poolType = "solo"
	}

	roundType := "block"
	if dbRound.Pending {
		roundType = "pending"
	} else if dbRound.Orphan {
		roundType = "orphan"
	} else if dbRound.Uncle {
		roundType = "uncle"
	}

	parsedValue, err := newNumberFromBigInt(value, dbRound.ChainID)
	if err != nil {
		return nil, err
	}

	var parsedMinerValue, parsedMinerPercentage *Number
	if dbRound.MinerValue.Valid {
		parsedMinerValue, err = newNumberFromBigIntPtr(dbRound.MinerValue.BigInt, dbRound.ChainID)
		if err != nil {
			return nil, err
		}

		var minerPercentage float64
		if value.Cmp(common.Big0) > 0 {
			numerator := common.BigIntToFloat64(dbRound.MinerValue.BigInt, common.Big10)
			denominator := common.BigIntToFloat64(value, common.Big10)
			minerPercentage = 100 * (numerator / denominator)
		}

		parsedMinerPercentage = newNumberFromFloat64Ptr(minerPercentage, "%", false)
	}

	explorerURL, err := getBlockExplorerURL(dbRound.ChainID, dbRound.Hash, dbRound.Height)
	if err != nil {
		return nil, err
	}

	if dbRound.Miner != nil {
		miner := types.StringValue(dbRound.Miner)
		if parts := strings.Split(miner, ":"); len(parts) == 3 {
			dbRound.Miner = types.StringPtr(strings.Join(parts[1:], ":"))
		}
	}

	round := &Round{
		Chain:           dbRound.ChainID,
		Type:            roundType,
		Pool:            poolType,
		Pending:         dbRound.Pending,
		Mature:          dbRound.Mature,
		Hash:            dbRound.Hash,
		Height:          dbRound.Height,
		ExplorerURL:     explorerURL,
		Miner:           dbRound.Miner,
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

func (c *Client) GetGlobalRoundsByChain(
	chain string,
	page, size uint64,
) ([]*Round, uint64, error) {
	count, err := pooldb.GetRoundsByChainCount(c.pooldb.Reader(), chain)
	if err != nil {
		return nil, 0, err
	}

	dbRounds, err := pooldb.GetRoundsByChain(c.pooldb.Reader(), chain, page, size)
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

func (c *Client) GetMinerRounds(
	minerIDs []uint64,
	page, size uint64,
) ([]*Round, uint64, error) {
	if len(minerIDs) == 0 {
		return nil, 0, nil
	}

	count, err := pooldb.GetRoundsByMinerIDsCount(c.pooldb.Reader(), minerIDs)
	if err != nil {
		return nil, 0, err
	}

	dbRounds, err := pooldb.GetRoundsByMinerIDs(c.pooldb.Reader(), minerIDs, page, size)
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

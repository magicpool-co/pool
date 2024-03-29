package etc

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

func stringsEqualInsensitive(a, b string) bool {
	return strings.ToLower(a) == strings.ToLower(b)
}

func (node Node) GetBlockExplorerURL(round *pooldb.Round) string {
	if node.mainnet {
		return fmt.Sprintf("https://blockscout.com/etc/mainnet/block/%d", round.Height)
	}
	return fmt.Sprintf("https://blockscout.com/etc/mordor/block/%d", round.Height)
}

func (node Node) getStatusByHost(hostID string) (uint64, bool, error) {
	height, err := node.getBlockNumber(hostID)
	if err != nil {
		return 0, false, err
	}

	syncing, err := node.getSyncing(hostID)
	if err != nil {
		return 0, false, err
	}
	node.rpcHost.SetHostSyncStatus(hostID, !syncing)

	return height, syncing, nil
}

func (node Node) GetStatus() (uint64, bool, error) {
	return node.getStatusByHost("")
}

func (node Node) PingHosts() ([]string, []uint64, []bool, []error) {
	hostIDs := node.rpcHost.GetAllHosts()
	heights := make([]uint64, len(hostIDs))
	statuses := make([]bool, len(hostIDs))
	errs := make([]error, len(hostIDs))

	for i, hostID := range hostIDs {
		heights[i], statuses[i], errs[i] = node.getStatusByHost(hostID)
	}

	return hostIDs, heights, statuses, errs
}

func (node Node) GetBlocks(start, end uint64) ([]*tsdb.RawBlock, error) {
	const batchSize = 5
	if start > end {
		return nil, fmt.Errorf("invalid range")
	}

	blocks := make([]*tsdb.RawBlock, end-start+1)
	for i := 0; i < len(blocks); i += batchSize {
		limit := i + batchSize
		if len(blocks) < limit {
			limit = len(blocks)
		}

		heights := make([]uint64, limit-i)
		for j := range heights {
			heights[j] = start + uint64(i+j)
		}

		rawBlocks, err := node.getBlockByNumberMany(heights)
		if err != nil {
			return nil, err
		}

		for j, block := range rawBlocks {
			height, err := common.HexToUint64(block.Number)
			if err != nil {
				return nil, err
			}

			rawTimestamp, err := common.HexToUint64(block.Timestamp)
			if err != nil {
				return nil, err
			}
			timestamp := time.Unix(int64(rawTimestamp), 0)

			difficulty, err := common.HexToUint64(block.Difficulty)
			if err != nil {
				return nil, err
			}

			blockReward, err := node.calculateBlockReward(height, block)
			if err != nil {
				return nil, err
			}

			blocks[i+j] = &tsdb.RawBlock{
				ChainID:    node.Chain(),
				Hash:       block.Hash,
				Height:     height,
				Value:      common.BigIntToFloat64(blockReward, node.GetUnits().Big()),
				Difficulty: float64(difficulty),
				UncleCount: uint64(len(block.Uncles)),
				TxCount:    uint64(len(block.Transactions)),
				Timestamp:  timestamp,
			}
		}
	}

	return blocks, nil
}

func (node Node) GetBlocksByHash(startHash string, limit uint64) ([]*tsdb.RawBlock, error) {
	return nil, fmt.Errorf("GetBlocks: not implemented")
}

func (node Node) getBlockTemplate() (*types.StratumJob, error) {
	hostID, result, err := node.getWork()
	if err != nil {
		return nil, err
	}

	header, err := new(types.Hash).SetFromHex(result[0])
	if err != nil {
		return nil, err
	}

	seed, err := new(types.Hash).SetFromHex(result[1])
	if err != nil {
		return nil, err
	}

	target, ok := new(big.Int).SetString(strings.ReplaceAll(result[2], "0x", ""), 16)
	if !ok {
		return nil, fmt.Errorf("unable to cast target as big.Int")
	}

	height, err := new(types.Number).SetFromHex(result[3])
	if err != nil {
		return nil, err
	}

	template := &types.StratumJob{
		HostID:     hostID,
		ID:         header.PrefixedHex(),
		Header:     header,
		HeaderHash: header,
		Seed:       seed,
		Height:     height,
		Difficulty: new(types.Difficulty).SetFromBig(target, node.GetMaxDifficulty()),
	}

	return template, nil
}

func (node Node) JobNotify(ctx context.Context, interval time.Duration) chan *types.StratumJob {
	jobCh := make(chan *types.StratumJob)
	ticker := time.NewTicker(interval)
	staticInterval := time.Second * 15

	go func() {
		defer node.logger.RecoverPanic()

		var lastHeight uint64
		var lastJob time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				now := time.Now()
				job, err := node.getBlockTemplate()
				if err != nil {
					node.logger.Error(err)
				} else if lastHeight != job.Height.Value() || now.After(lastJob.Add(staticInterval)) {
					lastHeight = job.Height.Value()
					lastJob = now
					jobCh <- job
				}
			}
		}
	}()

	return jobCh
}

func (node Node) SubmitWork(
	job *types.StratumJob,
	work *types.StratumWork,
	diffFactor int,
) (types.ShareStatus, *types.Hash, *pooldb.Round, error) {
	mixDigest, digest, err := node.pow.Compute(work.Hash.Bytes(), job.Height.Value(), work.Nonce.Value())
	if err != nil {
		return types.InvalidShare, nil, nil, err
	} else if job.Header.Hex() != work.Hash.Hex() {
		return types.InvalidShare, nil, nil, nil
	} else if bytes.Compare(mixDigest, work.MixDigest.Bytes()) != 0 {
		return types.InvalidShare, nil, nil, nil
	}

	hash := new(types.Hash).SetFromBytes(digest)
	if !hash.MeetsDifficulty(node.GetShareDifficulty(diffFactor)) {
		return types.RejectedShare, hash, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, hash, nil, nil
	}

	accepted, err := node.sendSubmitWork(job.HostID, work.Nonce.PrefixedHex(),
		work.Hash.PrefixedHex(), work.MixDigest.PrefixedHex())
	if err != nil {
		return types.AcceptedShare, hash, nil, err
	} else if !accepted {
		return types.AcceptedShare, hash, nil, fmt.Errorf("block not accepted")
	}

	round := &pooldb.Round{
		ChainID:    node.Chain(),
		Height:     job.Height.Value(),
		Hash:       hash.PrefixedHex(),
		MixDigest:  types.StringPtr(work.MixDigest.PrefixedHex()),
		Nonce:      types.Uint64Ptr(work.Nonce.Value()),
		Difficulty: job.Difficulty.Value(),
		Pending:    true,
		Mature:     false,
		Uncle:      false,
		Orphan:     false,
	}

	return types.AcceptedShare, hash, round, nil
}

func (node Node) ParseWork(data []json.RawMessage, extraNonce string) (*types.StratumWork, error) {
	if len(data) != 3 {
		return nil, fmt.Errorf("incorrect work array length")
	}

	var nonce, hash, mixdigest string
	if err := json.Unmarshal(data[0], &nonce); err != nil || len(nonce) != (16+2) {
		return nil, fmt.Errorf("invalid nonce parameter")
	} else if err := json.Unmarshal(data[1], &hash); err != nil || len(hash) != (64+2) {
		return nil, fmt.Errorf("invalid hash parameter")
	} else if err := json.Unmarshal(data[2], &mixdigest); err != nil || len(mixdigest) != (64+2) {
		return nil, fmt.Errorf("invalid mixdigest parameter")
	}

	var err error
	var nonceVal *types.Number
	var hashVal, mixVal *types.Hash
	if nonceVal, err = new(types.Number).SetFromHex(nonce); err != nil {
		return nil, err
	} else if hashVal, err = new(types.Hash).SetFromHex(hash); err != nil {
		return nil, err
	} else if mixVal, err = new(types.Hash).SetFromHex(mixdigest); err != nil {
		return nil, err
	}

	work := &types.StratumWork{
		JobID:     hash,
		Nonce:     nonceVal,
		Hash:      hashVal,
		MixDigest: mixVal,
	}

	return work, nil
}

func (node Node) MarshalJob(
	rawID interface{},
	job *types.StratumJob,
	cleanJobs bool,
	clientType, diffFactor int,
) (interface{}, error) {
	id, err := json.Marshal(rawID)
	if err != nil {
		return nil, err
	}

	result := []interface{}{
		job.Header.PrefixedHex(),
		job.Seed.PrefixedHex(),
		node.GetShareDifficulty(diffFactor).TargetPrefixedHex(),
		job.Height.PrefixedHex(),
	}

	return rpc.NewResponse(id, result)
}

func (node Node) GetClientType(minerClient string) int {
	return 0
}

func (node Node) GetSubscribeResponses(id []byte, clientID, extraNonce string) ([]interface{}, error) {
	return nil, nil
}

func (node Node) GetAuthorizeResponses(diffFactor int) ([]interface{}, error) {
	return nil, nil
}

func (node Node) GetSetDifficultyResponse(diffFactor int) (interface{}, error) {
	return nil, nil
}

func (node Node) calculateBlockReward(height uint64, block *Block) (*big.Int, error) {
	var blockReward *big.Int
	switch node.ethType {
	case ETC:
		blockReward = getBlockRewardETC(height, uint64(len(block.Uncles)))
	case ETHW:
		blockReward = getBlockRewardETHW(height, uint64(len(block.Uncles)))
	}

	txids := make([]string, len(block.Transactions))
	for i, tx := range block.Transactions {
		txids[i] = tx.Hash
	}

	receipts, err := node.getTransactionReceiptMany(txids)
	if err != nil {
		return nil, err
	}

	txFees := new(big.Int)
	for i, tx := range block.Transactions {
		if receipts[i] == nil {
			continue
		}

		gasUsed, err := common.HexToBig(receipts[i].GasUsed)
		if err != nil {
			return nil, err
		}

		var gasPrice *big.Int
		if receipts[i].EffectiveGasPrice != "" {
			gasPrice, err = common.HexToBig(receipts[i].EffectiveGasPrice)
			if err != nil {
				return nil, err
			}
		} else {
			var ok bool
			// @TODO: this makes no sense to me, shouldn't it also be common.HexToBig?
			// @NOTE: it actually works, lets figure out why SetString works w/ leading 0x
			gasPrice, ok = new(big.Int).SetString(tx.GasPrice, 0)
			if !ok {
				return nil, fmt.Errorf("unable to parse gasPrice for tx %s", tx.Hash)
			}
		}

		txFees.Add(txFees, new(big.Int).Mul(gasUsed, gasPrice))
	}

	// EIP1559 fee burning
	if block.BaseFee != "" {
		baseFeePerGas, err := common.HexToBig(block.BaseFee)
		if err != nil {
			return nil, err
		}

		gasUsed, err := common.HexToBig(block.GasUsed)
		if err != nil {
			return nil, err
		}

		txFees.Sub(txFees, new(big.Int).Mul(baseFeePerGas, gasUsed))
	}

	blockReward.Add(blockReward, txFees)

	return blockReward, nil
}

func (node Node) UnlockRound(round *pooldb.Round) error {
	const unlockWindow = 7

	if round.Nonce == nil {
		return fmt.Errorf("block %d has no nonce", round.Height)
	}

	round.Uncle = false
	round.Orphan = true
	round.Pending = false
	round.Mature = false
	round.Spent = false

	for checkHeight := round.Height; checkHeight < round.Height+unlockWindow; checkHeight++ {
		block, err := node.getBlockByNumber(checkHeight)
		if err != nil {
			return err
		} else if stringsEqualInsensitive(block.Miner, node.address) {
			nonce, err := common.HexToUint64(block.Nonce)
			if err != nil {
				return err
			} else if nonce == types.Uint64Value(round.Nonce) {
				blockReward, err := node.calculateBlockReward(checkHeight, block)
				if err != nil {
					return err
				}

				rawTimestamp, err := common.HexToUint64(block.Timestamp)
				if err != nil {
					return err
				}

				round.Height = checkHeight
				round.Value = dbcl.NullBigInt{Valid: true, BigInt: blockReward}
				round.Hash = block.Hash
				round.Orphan = false
				round.CreatedAt = time.Unix(int64(rawTimestamp), 0)

				return nil
			}
		}

		// check for orphans at the same height
		for uncleIndex := range block.Uncles {
			uncle, err := node.getUncleByNumberAndIndex(checkHeight, uint64(uncleIndex))
			if err != nil {
				return err
			} else if stringsEqualInsensitive(uncle.Miner, node.address) {
				nonce, err := common.HexToUint64(uncle.Nonce)
				if err != nil {
					return err
				} else if nonce == types.Uint64Value(round.Nonce) {
					var uncleReward *big.Int
					switch node.ethType {
					case ETC:
						uncleReward = getUncleRewardETC(checkHeight, round.Height)
					case ETHW:
						uncleReward = getUncleRewardETHW(checkHeight, round.Height)
					}

					rawTimestamp, err := common.HexToUint64(uncle.Timestamp)
					if err != nil {
						return err
					}

					round.Value = dbcl.NullBigInt{Valid: true, BigInt: uncleReward}
					round.Hash = uncle.Hash
					round.UncleHeight = types.Uint64Ptr(checkHeight)
					round.Orphan = false
					round.Uncle = true
					round.CreatedAt = time.Unix(int64(rawTimestamp), 0)

					return nil
				}
			}
		}
	}

	return nil
}

func (node Node) MatureRound(round *pooldb.Round) ([]*pooldb.UTXO, error) {
	if round.Pending || round.Mature || round.Orphan {
		return nil, nil
	} else if !round.Value.Valid {
		return nil, fmt.Errorf("no value for round %d", round.ID)
	}

	height := round.Height
	if round.Uncle {
		if round.UncleHeight == nil {
			return nil, fmt.Errorf("no uncle height for round %d", round.ID)
		}
		height = types.Uint64Value(round.UncleHeight)
	}

	block, err := node.getBlockByNumber(height)
	if err != nil {
		return nil, err
	}

	if round.Uncle {
		var foundUncle bool
		for uncleIndex := range block.Uncles {
			uncle, err := node.getUncleByNumberAndIndex(height, uint64(uncleIndex))
			if err != nil {
				return nil, err
			} else if !stringsEqualInsensitive(uncle.Miner, node.address) {
				continue
			}

			nonce, err := common.HexToUint64(uncle.Nonce)
			if err != nil {
				return nil, err
			} else if nonce != types.Uint64Value(round.Nonce) {
				continue
			}

			foundUncle = true
			break
		}

		if !foundUncle {
			round.Uncle = false
			round.Orphan = true
			return nil, nil
		}
	} else {
		blockHeight, err := common.HexToUint64(block.Number)
		if err != nil {
			return nil, err
		} else if round.Hash != block.Hash || height != blockHeight {
			round.Orphan = true
			return nil, nil
		}
	}

	round.Mature = true

	utxos := []*pooldb.UTXO{
		&pooldb.UTXO{
			ChainID: round.ChainID,
			Value:   round.Value,
			TxID:    round.Hash,
			Index:   0,
			Active:  true,
			Spent:   false,
		},
	}

	return utxos, nil
}

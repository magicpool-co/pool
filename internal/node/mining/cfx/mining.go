package cfx

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

func (node Node) GetBlockExplorerURL(round *pooldb.Round) string {
	if node.mainnet {
		return fmt.Sprintf("https://www.confluxscan.io/epoch/%d", types.Uint64Value(round.EpochHeight))
	}
	return fmt.Sprintf("https://testnet.confluxscan.io/epoch/%d", types.Uint64Value(round.EpochHeight))
}

func parseBlockReward(blockReward *BlockRewardInfo) (*big.Int, error) {
	totalReward, err := common.HexToBig(blockReward.TotalReward)
	if err != nil {
		return nil, err
	}

	return totalReward, nil
}

func (node Node) getStatusByHost(hostID string) (uint64, bool, error) {
	now := time.Now()
	block, err := node.getLatestBlock(hostID)
	if err != nil {
		return 0, false, err
	}

	epochNumber, err := common.HexToUint64(block.EpochNumber)
	if err != nil {
		return 0, false, err
	}

	rawTimestamp, err := common.HexToUint64(block.Timestamp)
	if err != nil {
		return 0, false, err
	}
	timestamp := time.Unix(int64(rawTimestamp), 0)
	syncing := now.Sub(timestamp) > time.Minute*5

	return epochNumber, syncing, nil
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
	const batchSize = 75
	if start > end {
		return nil, fmt.Errorf("invalid range")
	}

	heights := make([]uint64, end-start+1)
	for i := range heights {
		heights[i] = start + uint64(i)
	}

	hashes := make([]string, 0)
	rewardIndex := make(map[string]float64)
	for i := 0; i < len(heights); i += batchSize {
		limit := i + batchSize
		if len(heights) < limit {
			limit = len(heights)
		}

		epochHashesList, err := node.getBlocksByEpochMany(heights[i:limit])
		if err != nil {
			return nil, err
		}

		for _, epochHashes := range epochHashesList {
			for _, epochHash := range epochHashes {
				hashes = append(hashes, epochHash)
			}
		}

		blockRewardsList, err := node.getBlockRewardInfoMany(heights[i:limit])
		if err != nil {
			return nil, err
		}

		for _, blockRewards := range blockRewardsList {
			for _, blockReward := range blockRewards {
				reward, err := parseBlockReward(blockReward)
				if err != nil {
					return nil, err
				}

				rewardIndex[blockReward.BlockHash] = common.BigIntToFloat64(reward, node.GetUnits().Big())
			}
		}
	}

	blocks := make([]*tsdb.RawBlock, len(hashes))
	for i := 0; i < len(hashes); i += batchSize {
		limit := i + batchSize
		if len(hashes) < limit {
			limit = len(hashes)
		}

		rawBlocks, err := node.getBlockByHashMany(hashes[i:limit])
		if err != nil {
			return nil, err
		}

		for j, block := range rawBlocks {
			if _, ok := rewardIndex[block.Hash]; !ok {
				blockRewards := make([]*BlockRewardInfo, 0)
				req, err := rpc.NewRequest("cfx_getBlockRewardInfo", block.EpochNumber)
				if err != nil {
					return nil, err
				}

				err = node.execRPCfromFallback(req, &blockRewards)
				if err != nil {
					return nil, err
				}

				for _, blockReward := range blockRewards {
					reward, err := parseBlockReward(blockReward)
					if err != nil {
						return nil, err
					}

					rewardIndex[block.Hash] = common.BigIntToFloat64(reward, node.GetUnits().Big())
				}
			}

			epochNumber, err := common.HexToUint64(block.EpochNumber)
			if err != nil {
				return nil, err
			}

			difficulty, err := common.HexToUint64(block.Difficulty)
			if err != nil {
				return nil, err
			}

			rawTimestamp, err := common.HexToUint64(block.Timestamp)
			if err != nil {
				return nil, err
			}
			timestamp := time.Unix(int64(rawTimestamp), 0)

			blocks[i+j] = &tsdb.RawBlock{
				ChainID:    node.Chain(),
				Height:     epochNumber,
				Value:      rewardIndex[block.Hash],
				Difficulty: float64(difficulty),
				TxCount:    uint64(len(block.Transactions)),
				Timestamp:  timestamp,
			}
		}
	}

	return blocks, nil
}

func (node Node) JobNotify(ctx context.Context, interval time.Duration, jobCh chan *types.StratumJob, errCh chan error) {
	go func() {
		defer node.logger.RecoverPanic()

		notifyCh := node.miningSubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case req := <-notifyCh:
				data := req.Params
				if len(data) < 3 {
					errCh <- fmt.Errorf("invalid job recieved: %v", data)
					continue
				}

				var ok bool
				var rawEpoch, rawHash, rawBoundary string
				if err := json.Unmarshal(data[1], &rawEpoch); err != nil {
					errCh <- fmt.Errorf("failed to parse epoch %v", data[1])
					continue
				} else if err := json.Unmarshal(data[2], &rawHash); err != nil {
					errCh <- fmt.Errorf("failed to parse hash %v", data[2])
					continue
				} else if err := json.Unmarshal(data[3], &rawBoundary); err != nil {
					errCh <- fmt.Errorf("failed to parse boundary %v", data[3])
					continue
				}

				epochVal, err := new(types.Number).SetFromString(rawEpoch)
				if err != nil {
					errCh <- fmt.Errorf("failed to set epoch %s", rawEpoch)
					continue
				}

				hashVal, err := new(types.Hash).SetFromHex(rawHash)
				if err != nil {
					errCh <- fmt.Errorf("failed to set hash %s", rawHash)
					continue
				}

				rawBoundary = strings.ReplaceAll(rawBoundary, "0x", "")
				boundaryBig, ok := new(big.Int).SetString(rawBoundary, 16)
				if !ok {
					errCh <- fmt.Errorf("failed to set boundary %s", rawBoundary)
					continue
				}

				job := &types.StratumJob{
					HostID:     req.HostID,
					Height:     epochVal,
					HeaderHash: hashVal,
					Difficulty: new(types.Difficulty).SetFromBig(boundaryBig, maxDiffBig),
				}

				jobCh <- job
			}
		}
	}()
}

func (node Node) SubmitWork(job *types.StratumJob, work *types.StratumWork) (types.ShareStatus, *pooldb.Round, error) {
	digest, err := node.pow.Compute(work.Hash.Bytes(), job.Height.Value(), work.Nonce.Value())
	if err != nil {
		return types.RejectedShare, nil, err
	} else if bytes.Compare(job.HeaderHash.Bytes(), work.Hash.Bytes()) != 0 {
		return types.InvalidShare, nil, nil
	}

	hash := new(types.Hash).SetFromBytes(digest)
	if !hash.MeetsDifficulty(node.GetShareDifficulty()) {
		return types.RejectedShare, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, nil, nil
	}

	validBlock, err := node.submitBlock(job.HostID, work.Nonce.PrefixedHex(), work.Hash.PrefixedHex())
	if err != nil {
		return types.AcceptedShare, nil, err
	} else if !validBlock {
		return types.AcceptedShare, nil, nil
	}

	round := &pooldb.Round{
		ChainID:     node.Chain(),
		Height:      job.Height.Value(),
		EpochHeight: types.Uint64Ptr(job.Height.Value()),
		Hash:        hash.PrefixedHex(),
		Nonce:       types.Uint64Ptr(work.Nonce.Value()),
		Difficulty:  job.Difficulty.Value(),
		Pending:     true,
		Mature:      false,
		Uncle:       false,
		Orphan:      false,
	}

	return types.AcceptedShare, round, nil
}

func (node Node) ParseWork(data []json.RawMessage, extraNonce string) (*types.StratumWork, error) {
	if len(data) != 4 {
		return nil, fmt.Errorf("incorrect work array length")
	}

	var worker, jobID string
	if err := json.Unmarshal(data[0], &worker); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data[1], &jobID); err != nil {
		return nil, err
	}

	var nonce, hash string
	if err := json.Unmarshal(data[2], &nonce); err != nil || len(nonce) != (16+2) {
		return nil, fmt.Errorf("invalid nonce parameter")
	} else if err := json.Unmarshal(data[3], &hash); err != nil || len(hash) != (64+2) {
		return nil, fmt.Errorf("invalid hash parameter")
	}

	var err error
	var nonceVal *types.Number
	var hashVal *types.Hash
	if nonceVal, err = new(types.Number).SetFromHex(nonce); err != nil {
		return nil, err
	} else if hashVal, err = new(types.Hash).SetFromHex(hash); err != nil {
		return nil, err
	}

	work := &types.StratumWork{
		WorkerID: worker,
		JobID:    jobID,
		Nonce:    nonceVal,
		Hash:     hashVal,
	}

	return work, nil
}

func (node Node) MarshalJob(id interface{}, job *types.StratumJob, cleanJobs bool) (interface{}, error) {
	result := []interface{}{
		job.ID,
		job.Height.Value(),
		job.HeaderHash.PrefixedHex(),
		node.GetShareDifficulty().TargetPrefixedHex(),
	}

	return rpc.NewRequestWithID(id, "mining.notify", result...)
}

func (node Node) GetSubscribeResponse(id []byte, clientID, extraNonce string) (interface{}, error) {
	return nil, nil
}

func (node Node) GetDifficultyRequest() (interface{}, error) {
	return nil, nil
}

func (node Node) UnlockRound(round *pooldb.Round) error {
	const unlockWindow = 32

	if round.Nonce == nil {
		return fmt.Errorf("block %d has no nonce", round.Height)
	} else if round.EpochHeight == nil {
		return fmt.Errorf("block %d has no epoch height", round.Height)
	}

	round.Uncle = false
	round.Orphan = true
	round.Pending = false
	round.Mature = false
	round.Spent = false

	epochHeight := types.Uint64Value(round.EpochHeight)
	for checkHeight := epochHeight; checkHeight < epochHeight+unlockWindow; checkHeight++ {
		blockRewards, err := node.getBlockRewardInfo(checkHeight)
		if err != nil {
			return err
		}

		for _, blockReward := range blockRewards {
			author := strings.ReplaceAll(blockReward.Author, ":TYPE.USER", "")
			if strings.ToLower(author) != strings.ToLower(node.address) {
				continue
			}

			blockHash := blockReward.BlockHash
			reward, err := parseBlockReward(blockReward)
			if err != nil {
				return err
			}

			block, err := node.getBlockByHash(blockHash)
			if err != nil {
				return err
			}

			nonce, err := common.HexToUint64(block.Nonce)
			if err != nil {
				return err
			} else if nonce != types.Uint64Value(round.Nonce) {
				continue
			}

			blockHeight, err := common.HexToUint64(block.Height)
			if err != nil {
				return err
			}

			rawTimestamp, err := common.HexToUint64(block.Timestamp)
			if err != nil {
				return err
			}

			round.Value = dbcl.NullBigInt{Valid: true, BigInt: reward}
			round.Hash = blockHash
			round.EpochHeight = types.Uint64Ptr(checkHeight)
			round.Height = blockHeight
			round.Orphan = false
			round.Uncle = false
			round.CreatedAt = time.Unix(int64(rawTimestamp), 0)

			return nil
		}
	}

	return nil
}

package kas

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/crypto/blkbuilder"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

const (
	coinbaseRecursionLimit = 50
	chainTipBufferSize     = 72
)

func (node Node) GetBlockExplorerURL(round *pooldb.Round) string {
	if node.mainnet {
		return fmt.Sprintf("https://katnip.kaspad.net/blocks/%s", round.Hash)
	}
	return fmt.Sprintf("https://katnip-testnet.kaspad.net/blocks/%s", round.Hash)
}

func (node Node) getStatusByHost(hostID string) (uint64, bool, error) {
	tipHash, err := node.getSelectedTipHash(hostID)
	if err != nil {
		return 0, false, err
	}

	tip, err := node.getBlock(hostID, tipHash, false)
	if err != nil {
		return 0, false, err
	}

	syncing, err := node.getInfo(hostID)
	if err != nil {
		return 0, false, err
	}

	return tip.BlueScore, syncing, nil
}

func (node Node) GetStatus() (uint64, bool, error) {
	return node.getStatusByHost("")
}

func (node Node) PingHosts() ([]string, []uint64, []bool, []error) {
	hostIDs := node.grpcHost.GetAllHosts()
	heights := make([]uint64, len(hostIDs))
	statuses := make([]bool, len(hostIDs))
	errs := make([]error, len(hostIDs))

	for i, hostID := range hostIDs {
		heights[i], statuses[i], errs[i] = node.getStatusByHost(hostID)
	}

	return hostIDs, heights, statuses, errs
}

func (node Node) GetBlocks(start, end uint64) ([]*tsdb.RawBlock, error) {
	return nil, fmt.Errorf("GetBlocks: not implemented")
}

func (node Node) fetchBlockWithCache(hash string, cache *blockCache) (*Block, error) {
	if cache != nil {
		block := cache.get(hash)
		if block != nil {
			return block, nil
		}
	}

	block, err := node.getBlock("", hash, true)
	if err != nil {
		return nil, err
	} else if len(block.Transactions) == 0 {
		return nil, fmt.Errorf("no transactions in block: %s", hash)
	} else if len(block.Transactions[0].Outputs) == 0 {
		return nil, fmt.Errorf("no transaction outputs found in block: %s", hash)
	} else if len(block.MergeSetBluesHashes) > len(block.Transactions[0].Outputs) {
		return nil, fmt.Errorf("more blue hashes than tx outputs in block: %s", hash)
	}

	if cache != nil {
		if block.BlueScore > cache.minScore && block.BlueScore < cache.maxScore {
			cache.add(block)
		}
	}

	return block, nil
}

func (node Node) getRewardsFromBlock(block *Block, cache *blockCache) (uint64, error) {
	const maxRecursionDepth = 50

	var reward uint64
	if block.IsChainBlock {
		tx := block.Transactions[0]
		if len(block.MergeSetBluesHashes) == len(tx.Outputs)+1 {
			reward += tx.Outputs[len(tx.Outputs)-1].Amount
		}
	}

	var depth int
	childrenHashes := block.ChildrenHashes
	for len(childrenHashes) > 0 {
		newChildrenHashes := make([]string, 0)
		for _, childHash := range childrenHashes {
			child, err := node.fetchBlockWithCache(childHash, cache)
			if err != nil {
				return 0, err
			} else if !child.IsChainBlock {
				if len(newChildrenHashes) == 0 {
					depth++
				}

				newChildrenHashes = append(newChildrenHashes, child.ChildrenHashes...)
				continue
			}

			for i, mergeSetHash := range child.MergeSetBluesHashes {
				if mergeSetHash == block.Hash {
					reward += child.Transactions[0].Outputs[i].Amount
				}
			}

			return reward, nil
		}

		childrenHashes = newChildrenHashes
		if depth > maxRecursionDepth {
			return 0, fmt.Errorf("past max recursion depth for block: %s", block.Hash)
		}
	}

	return 0, nil
}

func (node Node) GetBlocksByHash(startHash string, limit uint64) ([]*tsdb.RawBlock, error) {
	// fetch the tip hash and the tip block to find the max blue score
	endHash, err := node.getSelectedTipHash("")
	if err != nil {
		return nil, err
	}

	endBlock, err := node.getBlock("", endHash, false)
	if err != nil {
		return nil, err
	}

	startBlock, err := node.getBlock("", startHash, true)
	if err != nil {
		return nil, err
	} else if len(startBlock.Parents) == 0 {
		return nil, fmt.Errorf("no parents in start block: %s", startHash)
	}

	minPrimaryBlueScore := startBlock.BlueScore + 1
	minSecondaryBlueScore := startBlock.BlueScore + 1 - chainTipBufferSize

	maxPrimaryBlueScore := startBlock.BlueScore + limit
	maxSecondaryBlueScore := startBlock.BlueScore + limit + chainTipBufferSize
	if maxSecondaryBlueScore > endBlock.BlueScore-chainTipBufferSize {
		maxPrimaryBlueScore = endBlock.BlueScore - chainTipBufferSize*2
		maxSecondaryBlueScore = endBlock.BlueScore - chainTipBufferSize
	}

	// there are no blocks within the acceptable range
	if maxPrimaryBlueScore < minPrimaryBlueScore {
		return nil, nil
	}

	idx := make(map[string]bool)
	blocks := make([]*tsdb.RawBlock, 0)
	cache := newBlockCache(minSecondaryBlueScore, maxSecondaryBlueScore, startBlock)
	for cache.size() > 0 {
		pendingHashes := cache.list()
		for _, hash := range pendingHashes {
			block, err := node.fetchBlockWithCache(hash, cache)
			if err != nil {
				return nil, err
			} else if idx[hash] {
				continue
			} else if block.BlueScore < minPrimaryBlueScore || block.BlueScore > maxPrimaryBlueScore {
				continue
			}

			value, err := node.getRewardsFromBlock(block, cache)
			if err != nil {
				return nil, err
			}

			idx[hash] = true
			blocks = append(blocks, &tsdb.RawBlock{
				ChainID:    node.Chain(),
				Hash:       block.Hash,
				Height:     block.BlueScore,
				Difficulty: block.Difficulty / 2,
				Value:      float64(value) / 1e8,
				TxCount:    uint64(len(block.Transactions)),
				Timestamp:  time.Unix(block.Timestamp/1000, 0),
			})
		}
	}

	sort.Slice(blocks, func(i, j int) bool {
		if blocks[i].Height == blocks[j].Height {
			return blocks[i].Timestamp.Before(blocks[j].Timestamp)
		}
		return blocks[i].Height < blocks[j].Height
	})

	return blocks, nil
}

func (node Node) parseBlockTemplate(template *Block) (*types.StratumJob, error) {
	header, err := blkbuilder.SerializeKaspaBlockHeader(uint16(template.Version), template.Parents,
		template.HashMerkleRoot, template.AcceptedIdMerkleRoot, template.UtxoCommitment, 0,
		template.Bits, 0, template.DaaScore, template.BlueScore, template.BlueWork, template.PruningPoint)
	if err != nil {
		return nil, err
	}

	job := &types.StratumJob{
		Header:     new(types.Hash).SetFromBytes(header),
		HeaderHash: new(types.Hash).SetFromBytes(header),
		Height:     new(types.Number).SetFromValue(template.BlueScore),
		Difficulty: new(types.Difficulty).SetFromBits(template.Bits, node.GetMaxDifficulty()),
		Data:       template,
	}

	return job, nil
}

func (node Node) JobNotify(ctx context.Context, interval time.Duration, jobCh chan *types.StratumJob, errCh chan error) {
	refreshTimer := time.NewTimer(interval)
	staticInterval := time.Second * 3

	go func() {
		defer node.logger.RecoverPanic()

		var lastHash string
		var lastJob time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-refreshTimer.C:
				now := time.Now()
				template, hostID, err := node.getBlockTemplate("")
				if err != nil {
					errCh <- err
				}

				job, err := node.parseBlockTemplate(template)
				if err != nil {
					errCh <- err
				} else if job.Header.Hex() != lastHash || now.After(lastJob.Add(staticInterval)) {
					job.HostID = hostID
					lastHash = job.Header.Hex()
					lastJob = now
					jobCh <- job
				}
			}
		}
	}()
}

func (node Node) SubmitWork(job *types.StratumJob, work *types.StratumWork) (types.ShareStatus, *pooldb.Round, error) {
	template, ok := job.Data.(*Block)
	if !ok {
		return types.RejectedShare, nil, fmt.Errorf("unable to cast job data as block")
	}

	digest, err := node.pow.Compute(job.Header.Bytes(), template.Timestamp, work.Nonce.Value())
	if err != nil {
		return types.RejectedShare, nil, err
	}

	hash := new(types.Hash).SetFromBytes(digest)
	if !hash.MeetsDifficulty(node.GetShareDifficulty()) {
		return types.RejectedShare, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, nil, nil
	}

	template.Nonce = work.Nonce.Value()
	blockHash, err := blkbuilder.SerializeKaspaBlockHeader(uint16(template.Version), template.Parents,
		template.HashMerkleRoot, template.AcceptedIdMerkleRoot, template.UtxoCommitment, template.Timestamp,
		template.Bits, template.Nonce, template.DaaScore, template.BlueScore, template.BlueWork, template.PruningPoint)
	if err != nil {
		return types.RejectedShare, nil, err
	}

	err = node.submitBlock(job.HostID, template)
	if err != nil {
		return types.AcceptedShare, nil, err
	}

	round := &pooldb.Round{
		ChainID:    node.Chain(),
		Height:     job.Height.Value(),
		Hash:       hex.EncodeToString(blockHash),
		Nonce:      types.Uint64Ptr(work.Nonce.Value()),
		Difficulty: job.Difficulty.Value() / 2,
		Pending:    true,
		Mature:     false,
		Uncle:      false,
		Orphan:     false,
	}

	return types.AcceptedShare, round, nil
}

func (node Node) ParseWork(data []json.RawMessage, extraNonce string) (*types.StratumWork, error) {
	if len(data) != 3 {
		return nil, fmt.Errorf("incorrect work array length")
	}

	var worker, jobID string
	if err := json.Unmarshal(data[0], &worker); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data[1], &jobID); err != nil {
		return nil, err
	}

	workerParts := strings.Split(worker, ".")
	if len(workerParts) == 2 {
		worker = workerParts[1]
	} else {
		worker = ""
	}

	var nonce string
	if err := json.Unmarshal(data[2], &nonce); err != nil || len(nonce) != 16 { // no 0x prefix
		return nil, fmt.Errorf("invalid nonce parameter")
	}

	nonceVal, err := new(types.Number).SetFromHex(nonce)
	if err != nil {
		return nil, err
	}

	work := &types.StratumWork{
		WorkerID: worker,
		JobID:    jobID,
		Nonce:    nonceVal,
	}

	return work, nil
}

func (node Node) MarshalJob(id interface{}, job *types.StratumJob, cleanJobs bool) (interface{}, error) {
	// @TODO: need to manage multiple versions of bzminer/lolminer responses
	// https://github.com/onemorebsmith/kaspa-pool/blob/4d061af0f354ff5f17637060e78f478650624e07/src/kaspastratum/hasher.go#L83-L119
	header := job.Header.Bytes()
	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(job.Timestamp.Unix()))

	parts := make([]interface{}, 5)
	for i := 0; i < 4; i++ {
		parts[i] = binary.BigEndian.Uint64(header[i*8 : (i+1)*8])
	}
	parts[4] = uint64(binary.LittleEndian.Uint64(timestamp))

	result := []interface{}{job.ID}
	if true {
		result = append(result, fmt.Sprintf("%016x%016x%016x%016x%016x", parts...))
	} else {
		result = append(result, parts...)
	}

	return rpc.NewRequestWithID(id, "mining.notify", result...)
}

func (node Node) GetSubscribeResponse(id []byte, clientID, extraNonce string) (interface{}, error) {
	return rpc.NewResponse(id, []interface{}{true})
}

func (node Node) GetAuthorizeResponses(extraNonce string) ([]interface{}, error) {
	extraNonceRes, err := rpc.NewRequest("mining.set_extranonce", extraNonce, 6)
	if err != nil {
		return nil, err
	}

	diffRes, err := rpc.NewRequest("mining.set_difficulty", 10)
	if err != nil {
		return nil, err
	}

	return []interface{}{extraNonceRes, diffRes}, nil
}

func (node Node) UnlockRound(round *pooldb.Round) error {
	block, err := node.getBlock("", round.Hash, true)
	if err != nil {
		return err
	} else if round.Nonce == nil || block.Nonce != types.Uint64Value(round.Nonce) {
		return fmt.Errorf("round %s has a nonce mismatch", round.Hash)
	}

	blockReward, err := node.getRewardsFromBlock(block, nil)
	if err != nil {
		return err
	}

	round.Value = dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(blockReward)}
	round.Uncle = false
	round.Orphan = false
	round.Pending = false
	round.Mature = false
	round.Spent = false
	round.CreatedAt = time.Unix(block.Timestamp/1000, 0)

	return nil
}

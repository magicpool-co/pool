package kas

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"google.golang.org/protobuf/proto"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/crypto/blkbuilder"
	"github.com/magicpool-co/pool/pkg/crypto/tx/kastx"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

const (
	coinbaseRecursionLimit = 50
	chainTipBufferSize     = 72

	standardMinerClientID = 0
	bzMinerClientID       = 1
)

var isBzminer = regexp.MustCompile(".*BzMiner.*")

func (node Node) GetBlockExplorerURL(round *pooldb.Round) string {
	if node.mainnet {
		return fmt.Sprintf("https://katnip.kaspad.net/block/%s", round.Hash)
	}
	return fmt.Sprintf("https://katnip-testnet.kaspad.net/block/%s", round.Hash)
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
	node.grpcHost.SetHostSyncStatus(hostID, syncing)

	return tip.BlueScore, !syncing, nil
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
		return nil, nil
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

func (node Node) getRewardsFromBlock(block *Block, cache *blockCache) ([]*Transaction, []uint32, []uint64, error) {
	const maxRecursionDepth = 50

	// handle rewards in case the block takes the reward of red blocks that it merges
	var mergeTx *Transaction
	var mergeIndex uint32
	var mergeValue uint64
	if block.IsChainBlock {
		tx := block.Transactions[0]
		if len(block.MergeSetBluesHashes) == len(tx.Outputs)+1 {
			mergeTx = tx
			mergeIndex = 0
			mergeValue = tx.Outputs[len(tx.Outputs)-1].Amount
		}
	}

	var depth int
	childrenHashes := block.ChildrenHashes
	for len(childrenHashes) > 0 {
		newChildrenHashes := make([]string, 0)
		for _, childHash := range childrenHashes {
			child, err := node.fetchBlockWithCache(childHash, cache)
			if err != nil {
				return nil, nil, nil, err
			} else if child == nil {
				continue
			} else if !child.IsChainBlock {
				if len(newChildrenHashes) == 0 {
					depth++
				}

				newChildrenHashes = append(newChildrenHashes, child.ChildrenHashes...)
				continue
			}

			var coinbaseTx *Transaction
			var coinbaseIndex uint32
			var coinbaseValue uint64
			for i, mergeSetHash := range child.MergeSetBluesHashes {
				if mergeSetHash == block.Hash {
					coinbaseTx = child.Transactions[0]
					coinbaseIndex = uint32(i)
					coinbaseValue = child.Transactions[0].Outputs[i].Amount
					break
				}
			}

			// means the block is red
			if coinbaseTx == nil {
				return nil, nil, nil, nil
			}

			txs := []*Transaction{coinbaseTx}
			indexes := []uint32{coinbaseIndex}
			values := []uint64{coinbaseValue}
			if mergeTx != nil {
				txs = append(txs, mergeTx)
				indexes = append(indexes, mergeIndex)
				values = append(values, mergeValue)
			}

			return txs, indexes, values, nil
		}

		childrenHashes = newChildrenHashes
		if depth > maxRecursionDepth {
			return nil, nil, nil, fmt.Errorf("past max recursion depth for block: %s", block.Hash)
		}
	}

	return nil, nil, nil, nil
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
			} else if block == nil || idx[hash] {
				continue
			} else if block.BlueScore < minPrimaryBlueScore || block.BlueScore > maxPrimaryBlueScore {
				continue
			}

			_, _, coinbaseRewards, err := node.getRewardsFromBlock(block, cache)
			if err != nil {
				return nil, err
			}

			var blockReward uint64
			for _, coinbaseReward := range coinbaseRewards {
				blockReward += coinbaseReward
			}

			idx[hash] = true
			blocks = append(blocks, &tsdb.RawBlock{
				ChainID:    node.Chain(),
				Hash:       block.Hash,
				Height:     block.BlueScore,
				Difficulty: block.Difficulty / 2,
				Value:      float64(blockReward) / 1e8,
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
		Timestamp:  time.Unix(template.Timestamp, 0),
		Data:       template,
	}

	return job, nil
}

func (node Node) JobNotify(ctx context.Context, interval time.Duration) chan *types.StratumJob {
	jobCh := make(chan *types.StratumJob)
	ticker := time.NewTicker(interval)
	staticInterval := time.Second * 5

	go func() {
		defer node.logger.RecoverPanic()

		var lastHash string
		var lastJob time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				now := time.Now()
				template, hostID, err := node.getBlockTemplate("")
				if err != nil {
					node.logger.Error(fmt.Errorf("kas: %v", err))
				} else {
					job, err := node.parseBlockTemplate(template)
					if err != nil {
						node.logger.Error(fmt.Errorf("kas: %v", err))
					} else if job.Header.Hex() != lastHash || now.After(lastJob.Add(staticInterval)) {
						job.HostID = hostID
						lastHash = job.Header.Hex()
						lastJob = now
						jobCh <- job
					}
				}
			}
		}
	}()

	return jobCh
}

func (node Node) SubmitWork(job *types.StratumJob, work *types.StratumWork) (types.ShareStatus, *types.Hash, *pooldb.Round, error) {
	template, ok := job.Data.(*Block)
	if !ok {
		return types.RejectedShare, nil, nil, fmt.Errorf("unable to cast job data as block")
	}

	digest, err := node.pow.Compute(job.Header.Bytes(), template.Timestamp, work.Nonce.Value())
	if err != nil {
		return types.RejectedShare, nil, nil, err
	}

	hash := new(types.Hash).SetFromBytes(digest)
	if !hash.MeetsDifficulty(node.GetShareDifficulty()) {
		return types.RejectedShare, nil, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, hash, nil, nil
	}

	template.Nonce = work.Nonce.Value()
	blockHash, err := blkbuilder.SerializeKaspaBlockHeader(uint16(template.Version), template.Parents,
		template.HashMerkleRoot, template.AcceptedIdMerkleRoot, template.UtxoCommitment, template.Timestamp,
		template.Bits, template.Nonce, template.DaaScore, template.BlueScore, template.BlueWork, template.PruningPoint)
	if err != nil {
		return types.AcceptedShare, hash, nil, err
	}

	err = node.submitBlock(job.HostID, template)
	if err != nil {
		return types.AcceptedShare, hash, nil, err
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

	return types.AcceptedShare, hash, round, nil
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
	if err := json.Unmarshal(data[2], &nonce); err != nil || (len(nonce) != 16 && len(nonce) != 18) {
		return nil, fmt.Errorf("invalid nonce parameter: %s, %v", data[2], err)
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

func (node Node) MarshalJob(id interface{}, job *types.StratumJob, cleanJobs bool, clientType int) (interface{}, error) {
	result := []interface{}{job.ID}
	switch clientType {
	case standardMinerClientID:
		header := job.Header.Bytes()
		parts := make([]uint64, 4)
		for i := 0; i < 4; i++ {
			parts[i] = binary.LittleEndian.Uint64(header[i*8 : (i+1)*8])
		}

		result = append(result, parts)
		result = append(result, job.Timestamp.Unix())
	case bzMinerClientID:
		timestamp := make([]byte, 8)
		binary.LittleEndian.PutUint64(timestamp, uint64(job.Timestamp.Unix()))
		result = append(result, job.Header.Hex()+hex.EncodeToString(timestamp))
	}

	return rpc.NewRequestWithID(id, "mining.notify", result...)
}

func (node Node) GetClientType(minerClient string) int {
	if isBzminer.MatchString(minerClient) {
		return bzMinerClientID
	}

	return standardMinerClientID
}

func (node Node) GetSubscribeResponses(id []byte, clientID, extraNonce string) ([]interface{}, error) {
	res, err := rpc.NewResponse(id, []interface{}{true})
	if err != nil {
		return nil, err
	}

	extraNonceRes, err := rpc.NewRequest("mining.set_extranonce", extraNonce, 6)
	if err != nil {
		return nil, err
	}

	return []interface{}{res, extraNonceRes}, nil
}

func (node Node) GetAuthorizeResponses() ([]interface{}, error) {
	res, err := rpc.NewRequest("mining.set_difficulty", 10)
	if err != nil {
		return nil, err
	}

	return []interface{}{res}, nil
}

func (node Node) UnlockRound(round *pooldb.Round) error {
	block, err := node.getBlock("", round.Hash, true)
	if err != nil {
		return err
	} else if round.Nonce == nil || block.Nonce != types.Uint64Value(round.Nonce) {
		return fmt.Errorf("round %s has a nonce mismatch", round.Hash)
	}

	coinbaseTxs, _, coinbaseRewards, err := node.getRewardsFromBlock(block, nil)
	if err != nil {
		return err
	}

	orphan := len(coinbaseTxs) == 0
	var coinbaseTxID *string
	if !orphan {
		coinbaseTxBytes, err := proto.Marshal(transactionToProtowire(coinbaseTxs[0]))
		if err != nil {
			return err
		}
		coinbaseTxID = types.StringPtr(kastx.CalculateTxID(hex.EncodeToString(coinbaseTxBytes)))
	}

	var blockReward uint64
	for _, coinbaseReward := range coinbaseRewards {
		blockReward += coinbaseReward
	}

	round.Value = dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(blockReward)}
	round.CoinbaseTxID = coinbaseTxID
	round.Uncle = false
	round.Orphan = orphan
	round.Pending = false
	round.Mature = false
	round.Spent = false
	round.CreatedAt = time.Unix(block.Timestamp/1000, 0)

	return nil
}

func (node Node) MatureRound(round *pooldb.Round) ([]*pooldb.UTXO, error) {
	if round.Pending || round.Orphan || round.Mature {
		return nil, nil
	} else if round.Nonce == nil {
		return nil, fmt.Errorf("no nonce for round %d", round.ID)
	} else if !round.Value.Valid {
		return nil, fmt.Errorf("no value for round %d", round.ID)
	}

	block, err := node.getBlock("", round.Hash, true)
	if err != nil {
		return nil, err
	} else if block.Nonce != types.Uint64Value(round.Nonce) {
		round.Orphan = true
		return nil, nil
	}

	coinbaseTxs, coinbaseIndexes, coinbaseRewards, err := node.getRewardsFromBlock(block, nil)
	if err != nil {
		return nil, err
	} else if len(coinbaseTxs) == 0 {
		round.Orphan = true
		return nil, nil
	}

	round.Mature = true

	utxos := make([]*pooldb.UTXO, len(coinbaseTxs))
	for i, tx := range coinbaseTxs {
		txBytes, err := proto.Marshal(transactionToProtowire(tx))
		if err != nil {
			return nil, err
		}

		txid := kastx.CalculateTxID(hex.EncodeToString(txBytes))
		if txid == "" {
			return nil, fmt.Errorf("calculated empty coinbase txid")
		}

		utxos[i] = &pooldb.UTXO{
			ChainID: round.ChainID,
			Value:   dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(coinbaseRewards[i])},
			TxID:    txid,
			Index:   coinbaseIndexes[i],
			Active:  true,
			Spent:   false,
		}
	}

	return utxos, nil
}

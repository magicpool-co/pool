package erg

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strconv"
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
	// @TODO: we need to make sure we are properly handle multiple block scenarios (like 832628)
	hashes, err := node.getBlocksAtHeight(round.Height)
	if err == nil {
		if node.mainnet {
			return "https://explorer.ergoplatform.com/en/blocks/" + hashes[0]
		}
		return "https://testnet.ergoplatform.com/en/blocks/" + hashes[0]
	}
	return ""
}

func (node Node) GetBlocks(start, end uint64) ([]*tsdb.RawBlock, error) {
	if start > end {
		return nil, fmt.Errorf("invalid range")
	}

	blocks := make([]*tsdb.RawBlock, 0)
	for height := start; height <= end; height++ {
		hashes, err := node.getBlocksAtHeight(height)
		if err != nil {
			return nil, err
		} else if len(hashes) == 0 {
			continue
		}

		for _, hash := range hashes[:1] {
			rawBlock, err := node.getBlock(hash)
			if err != nil {
				return nil, err
			} else if rawBlock == nil || rawBlock.Header == nil {
				return nil, fmt.Errorf("empty block")
			}

			difficulty, err := strconv.ParseFloat(rawBlock.Header.Difficulty, 64)
			if err != nil {
				return nil, err
			}

			unitsBigFloat := new(big.Float).SetInt(node.GetUnits().Big())
			rewardBigInt := getBlockReward(height)
			_, _, txFees, err := node.getRewardsFromBlock(rawBlock)
			if err != nil {
				return nil, err
			}

			rewardBigInt.Add(rewardBigInt, new(big.Int).SetUint64(txFees))
			rewardsBigFloat := new(big.Float).SetInt(rewardBigInt)
			rewardFloat64, _ := new(big.Float).Quo(rewardsBigFloat, unitsBigFloat).Float64()

			block := &tsdb.RawBlock{
				ChainID:    node.Chain(),
				Hash:       rawBlock.Header.ID,
				Height:     height,
				Value:      rewardFloat64,
				Difficulty: difficulty,
				TxCount:    uint64(len(rawBlock.BlockTransactions.Transactions)),
				Timestamp:  time.Unix(rawBlock.Header.Timestamp/1000, 0),
			}
			blocks = append(blocks, block)
		}
	}

	return blocks, nil
}

func (node Node) GetBlocksByHash(startHash string, limit uint64) ([]*tsdb.RawBlock, error) {
	return nil, fmt.Errorf("GetBlocks: not implemented")
}

func (node Node) getStatusByHost(hostID string) (uint64, bool, error) {
	info, err := node.getInfo(hostID)
	if err != nil {
		return 0, false, err
	}

	height := info.FullHeight
	syncing := info.MaxPeerHeight > info.FullHeight
	node.httpHost.SetHostSyncStatus(hostID, !syncing)

	return height, syncing, nil
}

func (node Node) GetStatus() (uint64, bool, error) {
	return node.getStatusByHost("")
}

func (node Node) PingHosts() ([]string, []uint64, []bool, []error) {
	hostIDs := node.httpHost.GetAllHosts()
	heights := make([]uint64, len(hostIDs))
	statuses := make([]bool, len(hostIDs))
	errs := make([]error, len(hostIDs))

	for i, hostID := range hostIDs {
		heights[i], statuses[i], errs[i] = node.getStatusByHost(hostID)
	}

	return hostIDs, heights, statuses, errs
}

func (node Node) getBlockTemplate() (*types.StratumJob, error) {
	info, err := node.getInfo("")
	if err != nil {
		return nil, err
	}

	hostID, candidate, err := node.getMiningCandidate()
	if err != nil {
		return nil, err
	}

	msg, err := new(types.Hash).SetFromHex(candidate.Msg)
	if err != nil {
		return nil, err
	}

	if math.IsNaN(candidate.B) {
		return nil, fmt.Errorf("NaN target")
	}

	rawTarget := new(big.Float).SetFloat64(candidate.B)
	target, _ := rawTarget.Int(nil)

	template := &types.StratumJob{
		HostID:     hostID,
		Header:     msg,
		HeaderHash: msg,
		Height:     new(types.Number).SetFromValue(candidate.Height),
		Difficulty: new(types.Difficulty).SetFromBig(target, maxDiffBig),
		Version:    new(types.Number).SetFromValue(info.Parameters.BlockVersion),
	}

	return template, nil
}

func (node Node) getRewardsFromBlock(block *Block) (string, []string, uint64, error) {
	if block.BlockTransactions == nil {
		return "", nil, 0, fmt.Errorf("block transactions is nil")
	} else if len(block.BlockTransactions.Transactions) == 0 {
		return "", nil, 0, fmt.Errorf("no transactions in block")
	}

	coinbaseTx := block.BlockTransactions.Transactions[0]
	if len(coinbaseTx.Outputs) < 2 {
		return "", nil, 0, fmt.Errorf("invalid coinbase tx")
	}

	txids := []string{coinbaseTx.ID}
	var address string
	for _, output := range coinbaseTx.Outputs[1:] {
		var err error
		address, err = node.getAddressFromErgoTree(output.ErgoTree)
		if err != nil {
			return "", nil, 0, err
		}
		break
	}

	var feeValue uint64
	if txCount := len(block.BlockTransactions.Transactions); txCount != 1 {
		// if there are transactions in the block, the final
		// transaction will be the tx for transaction fees
		feeTx := block.BlockTransactions.Transactions[txCount-1]
		if len(feeTx.Outputs) != 1 {
			return "", nil, 0, fmt.Errorf("invalid fee tx")
		}

		txids = append(txids, feeTx.ID)
		feeValue = feeTx.Outputs[0].Value
	}

	return address, txids, feeValue, nil
}

func (node Node) JobNotify(ctx context.Context, interval time.Duration) chan *types.StratumJob {
	jobCh := make(chan *types.StratumJob)
	ticker := time.NewTicker(interval)
	staticInterval := time.Minute * 2

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
	digest, err := node.pow.Compute(job.Header.Bytes(), job.Height.Value(), work.Nonce.Value())
	if err != nil {
		return types.InvalidShare, nil, nil, err
	}

	hash := new(types.Hash).SetFromBytes(digest)
	if !hash.MeetsDifficulty(node.GetShareDifficulty(diffFactor)) {
		return types.RejectedShare, hash, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, hash, nil, nil
	}

	err = node.postMiningSolution(job.HostID, work.Nonce.Hex())
	if err != nil {
		return types.AcceptedShare, hash, nil, err
	}

	round := &pooldb.Round{
		ChainID:    node.Chain(),
		Height:     job.Height.Value(),
		Hash:       hash.Hex(),
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
	if len(data) != 5 {
		return nil, fmt.Errorf("incorrect work array length")
	}

	var worker, jobID string
	if err := json.Unmarshal(data[0], &worker); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data[1], &jobID); err != nil {
		return nil, err
	}

	var nonce string
	if err := json.Unmarshal(data[4], &nonce); err != nil || len(nonce) != 16 { // no 0x prefix
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

func (node Node) MarshalJob(
	id interface{},
	job *types.StratumJob,
	cleanJobs bool,
	clientType, diffFactor int,
) (interface{}, error) {
	result := []interface{}{
		job.ID,
		job.Height.Value(),
		job.Header.Hex(), // no 0x prefix
		"",
		"",
		job.Version.Hex(), // no 0x prefix
		node.GetShareDifficulty(diffFactor).TargetBig().String(),
		"",
		cleanJobs,
	}

	return rpc.NewRequestWithID(id, "mining.notify", result...)
}

func (node Node) GetClientType(minerClient string) int {
	return 0
}

func (node Node) GetSubscribeResponses(id []byte, clientID, extraNonce string) ([]interface{}, error) {
	res, err := rpc.NewResponseForced(id, []interface{}{nil, extraNonce, 6})
	if err != nil {
		return nil, err
	}

	return []interface{}{res}, nil
}

func (node Node) GetAuthorizeResponses(diffFactor int) ([]interface{}, error) {
	res, err := node.GetSetDifficultyResponse(diffFactor)
	if err != nil {
		return nil, err
	}

	return []interface{}{res}, nil
}

func (node Node) GetSetDifficultyResponse(diffFactor int) (interface{}, error) {
	return rpc.NewRequest("mining.set_difficulty", diffFactor)
}

func (node Node) UnlockRound(round *pooldb.Round) error {
	if round.Nonce == nil {
		return fmt.Errorf("block %d has no nonce", round.Height)
	}

	hashes, err := node.getBlocksAtHeight(round.Height)
	if err != nil {
		return err
	} else if len(hashes) > 1 {
		hashes = hashes[:1]
	}

	round.Uncle = false
	round.Orphan = true
	round.Pending = false
	round.Mature = false
	round.Spent = false

	for _, hash := range hashes {
		block, err := node.getBlock(hash)
		if err != nil {
			return err
		} else if block == nil || block.Header == nil {
			return fmt.Errorf("empty block")
		}

		_, txids, txFees, err := node.getRewardsFromBlock(block)
		if err != nil {
			return err
		}

		nonce, err := common.HexToUint64(block.Header.PowSolutions.N)
		if err != nil {
			return err
		} else if nonce == types.Uint64Value(round.Nonce) {
			if len(txids) == 0 {
				return fmt.Errorf("no txids for round %d", round.ID)
			}

			blockReward := getBlockReward(round.Height)
			blockReward.Add(blockReward, new(big.Int).SetUint64(txFees))

			round.Value = dbcl.NullBigInt{Valid: true, BigInt: blockReward}
			round.CoinbaseTxID = types.StringPtr(txids[0])
			round.Hash = hash
			round.Orphan = false
			round.CreatedAt = time.Unix(block.Header.Timestamp/1000, 0)
		}
	}

	return nil
}

func (node Node) MatureRound(round *pooldb.Round) ([]*pooldb.UTXO, error) {
	if round.Pending || round.Orphan || round.Mature {
		return nil, nil
	} else if round.Nonce == nil {
		return nil, fmt.Errorf("no nonce for round %d", round.ID)
	}

	block, err := node.getBlock(round.Hash)
	if err != nil {
		return nil, err
	} else if block == nil || block.Header == nil {
		return nil, fmt.Errorf("empty block for round %d", round.ID)
	}

	_, txids, feeValue, err := node.getRewardsFromBlock(block)
	if err != nil {
		return nil, err
	} else if (feeValue > 0 && len(txids) != 2) || (feeValue == 0 && len(txids) != 1) {
		return nil, fmt.Errorf("bad txid count for round %d", round.ID)
	}

	nonce, err := common.HexToUint64(block.Header.PowSolutions.N)
	if err != nil {
		return nil, err
	} else if nonce != types.Uint64Value(round.Nonce) {
		round.Orphan = true
		return nil, nil
	}

	round.Mature = true

	utxos := []*pooldb.UTXO{
		&pooldb.UTXO{
			ChainID: round.ChainID,
			Value:   dbcl.NullBigInt{Valid: true, BigInt: getBlockReward(round.Height)},
			TxID:    txids[0],
			Index:   0,
			Active:  true,
			Spent:   false,
		},
	}

	if len(txids) == 2 {
		utxos = append(utxos, &pooldb.UTXO{
			ChainID: round.ChainID,
			Value:   dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(feeValue)},
			TxID:    txids[1],
			Index:   uint32(len(block.BlockTransactions.Transactions) - 1),
			Active:  true,
			Spent:   false,
		})
	}

	return utxos, nil
}

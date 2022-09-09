package ergo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

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

// @TODO: this needs to be in a more reasonable place
// since a) on node reboots this gets reset (i think), b)
// it will only execute on the top node, not all of them (which could be
// solved by spread RPC call), c) if a new node comes into the mix it
// wont be initialized until the pool reboots
func (node Node) InitMining() error {
	status, err := node.getWalletStatus()
	if err != nil {
		return err
	} else if !status.IsInitialized {
		err = node.postWalletRestore()
		if err != nil {
			return err
		}
	}

	if !status.IsUnlocked {
		err = node.postWalletUnlock()
		if err != nil {
			return err
		}
	}

	return nil
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
			_, txFees, err := node.getRewardsFromBlock(rawBlock)
			if err != nil {
				return nil, err
			}

			rewardBigInt.Add(rewardBigInt, new(big.Int).SetUint64(txFees))
			rewardsBigFloat := new(big.Float).SetInt(rewardBigInt)
			rewardFloat64, _ := new(big.Float).Quo(rewardsBigFloat, unitsBigFloat).Float64()

			block := &tsdb.RawBlock{
				ChainID:    node.Chain(),
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

func (node Node) GetStatus() (uint64, bool, error) {
	info, err := node.getInfo()
	if err != nil {
		return 0, false, err
	}

	height := info.FullHeight
	syncing := info.MaxPeerHeight != info.FullHeight

	return height, syncing, nil
}

func (node Node) PingHosts() ([]uint64, []uint64, []bool, []error) {
	return nil, nil, nil, nil
}

func (node Node) getBlockTemplate() (*types.StratumJob, error) {
	info, err := node.getInfo()
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

func (node Node) getRewardsFromBlock(block *Block) (string, uint64, error) {
	var ergoTree string
	var value uint64
	if block.BlockTransactions == nil {
		return "", 0, fmt.Errorf("block transactions is nil")
	}

	txLen := len(block.BlockTransactions.Transactions)
	if txLen == 0 {
		return "", 0, fmt.Errorf("no transactions in block")
	} else if txLen == 1 {
		// means theres no transactions in the block
		coinbaseTx := block.BlockTransactions.Transactions[0]
		if len(coinbaseTx.Outputs) < 2 {
			return "", 0, fmt.Errorf("invalid coinbase tx")
		}

		for _, output := range coinbaseTx.Outputs[1:] {
			if ergoTree == "" {
				ergoTree = output.ErgoTree
			}
		}
	} else {
		// if there are transactions in the block, the final
		// transaction will be the tx for transaction fees
		feeTx := block.BlockTransactions.Transactions[txLen-1]
		if len(feeTx.Outputs) != 1 {
			return "", 0, fmt.Errorf("invalid fee tx")
		}

		value += feeTx.Outputs[0].Value
		ergoTree = feeTx.Outputs[0].ErgoTree
	}

	address, err := node.getAddressFromErgoTree(ergoTree)
	if err != nil {
		return "", 0, err
	}

	return address, value, nil
}

func (node Node) JobNotify(ctx context.Context, interval time.Duration, jobCh chan *types.StratumJob, errCh chan error) {
	refreshTimer := time.NewTimer(interval)
	staticInterval := time.Minute * 2

	go func() {
		var lastHeight uint64
		var lastJob time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-refreshTimer.C:
				now := time.Now()
				job, err := node.getBlockTemplate()
				if err != nil {
					errCh <- err
				} else if lastHeight != job.Height.Value() || now.After(lastJob.Add(staticInterval)) {
					lastHeight = job.Height.Value()
					lastJob = now
					jobCh <- job
				}

				refreshTimer.Reset(interval)
			}
		}
	}()
}

func (node Node) SubmitWork(job *types.StratumJob, work *types.StratumWork) (types.ShareStatus, *pooldb.Round, error) {
	digest, err := node.pow.Compute(job.Header.Bytes(), job.Height.Value(), work.Nonce.Value())
	if err != nil {
		return types.RejectedShare, nil, err
	}

	hash := new(types.Hash).SetFromBytes(digest)
	if !hash.MeetsDifficulty(node.GetShareDifficulty()) {
		return types.RejectedShare, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, nil, nil
	}

	err = node.postMiningSolution(job.HostID, work.Nonce.Hex())
	if err != nil {
		return types.AcceptedShare, nil, err
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

	return types.AcceptedShare, round, nil
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

func (node Node) MarshalJob(id interface{}, job *types.StratumJob, cleanJobs bool) (interface{}, error) {
	result := []interface{}{
		job.ID,
		job.Height.Value(),
		job.Header.Hex(), // no 0x prefix
		"",
		"",
		job.Version.Hex(), // no 0x prefix
		node.GetShareDifficulty().TargetBig().String(),
		"",
		cleanJobs,
	}

	return rpc.NewRequestWithID(id, "mining.notify", result...)
}

func (node Node) GetSubscribeResponse(id []byte, clientID, extraNonce string) (interface{}, error) {
	return rpc.NewResponseForced(id, []interface{}{nil, extraNonce, 6})
}

func (node Node) GetDifficultyRequest() (interface{}, error) {
	return rpc.NewRequest("mining.set_difficulty", 1)
}

func (node Node) UnlockRound(round *pooldb.Round) error {
	if round.Nonce == nil {
		return fmt.Errorf("block %d has no nonce", round.Height)
	}

	hashes, err := node.getBlocksAtHeight(round.Height)
	if err != nil {
		return err
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

		address, txFees, err := node.getRewardsFromBlock(block)
		if err != nil {
			return err
		} else if address == node.address {
			nonce, err := common.HexToUint64(block.Header.PowSolutions.N)
			if err != nil {
				return err
			} else if nonce == types.Uint64Value(round.Nonce) {
				blockReward := getBlockReward(round.Height)
				blockReward.Add(blockReward, new(big.Int).SetUint64(txFees))

				round.Value = dbcl.NullBigInt{Valid: true, BigInt: blockReward}
				round.Hash = hash
				round.Orphan = false
				round.CreatedAt = time.Unix(block.Header.Timestamp/1000, 0)
			}
		}
	}

	return nil
}

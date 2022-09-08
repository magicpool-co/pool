package ctxc

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetBlockExplorerURL(round *pooldb.Round) string {
	if node.mainnet {
		return fmt.Sprintf("https://cerebro.cortexlabs.ai/#/block/%d", round.Height)
	}
	return ""
}

func (node Node) GetStatus() (uint64, bool, error) {
	height, err := node.getBlockNumber()
	if err != nil {
		return 0, false, err
	}

	syncing, err := node.getSyncing()
	if err != nil {
		return 0, false, err
	}

	return height, syncing, nil
}

func (node Node) PingHosts() ([]uint64, []uint64, []bool, []error) {
	return nil, nil, nil, nil
}

func (node Node) GetBlocks(start, end uint64) ([]*tsdb.RawBlock, error) {
	if start > end {
		return nil, fmt.Errorf("invalid range")
	}

	heights := make([]uint64, end-start+1)
	for i := range heights {
		heights[i] = start + uint64(i)
	}

	rawBlocks, err := node.getBlockByNumberMany(heights)
	if err != nil {
		return nil, err
	}

	blocks := make([]*tsdb.RawBlock, len(rawBlocks))
	for i, block := range rawBlocks {
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

		blocks[i] = &tsdb.RawBlock{
			ChainID:    node.Chain(),
			Height:     height,
			Value:      common.BigIntToFloat64(blockReward, node.GetUnits().Big()),
			Difficulty: float64(difficulty),
			UncleCount: uint64(len(block.Uncles)),
			TxCount:    uint64(len(block.Transactions)),
			Timestamp:  timestamp,
		}
	}

	return blocks, nil
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

func (node Node) JobNotify(ctx context.Context, interval time.Duration, jobCh chan *types.StratumJob, errCh chan error) {
	refreshTimer := time.NewTimer(interval)
	staticInterval := time.Second * 15

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

func hashSolution(sol []uint64) []byte {
	buf := make([]byte, len(sol)*4)
	for i, v := range sol {
		binary.BigEndian.PutUint32(buf[i*4:], uint32(v))
	}

	return crypto.Keccak256(buf)
}

func generateHeader(hash []byte, nonce uint64) []byte {
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, nonce)
	header := append(hash, nonceBytes...)

	return header
}

func (node Node) SubmitWork(job *types.StratumJob, work *types.StratumWork) (types.ShareStatus, *pooldb.Round, error) {
	header := generateHeader(work.Hash.Bytes(), work.Nonce.Value())
	validSolution, err := node.pow.Verify(header, work.CuckooSolution.Data())
	if err != nil {
		return types.RejectedShare, nil, err
	} else if !validSolution {
		return types.InvalidShare, nil, nil
	} else if bytes.Compare(job.HeaderHash.Bytes(), work.Hash.Bytes()) != 0 {
		return types.InvalidShare, nil, nil
	}

	hash := new(types.Hash).SetFromBytes(hashSolution(work.CuckooSolution.Data()))
	if !hash.MeetsDifficulty(node.GetShareDifficulty()) {
		return types.RejectedShare, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, nil, nil
	}

	accepted, err := node.sendSubmitWork(job.HostID, work.Nonce.PrefixedHex(),
		work.Hash.PrefixedHex(), work.CuckooSolution.PrefixedHex())
	if err != nil {
		return types.AcceptedShare, nil, err
	} else if !accepted {
		return types.AcceptedShare, nil, fmt.Errorf("block not accepted")
	}

	round := &pooldb.Round{
		ChainID:    node.Chain(),
		Height:     job.Height.Value(),
		Hash:       hash.PrefixedHex(),
		Nonce:      types.Uint64Ptr(work.Nonce.Value()),
		Solution:   types.StringPtr(work.CuckooSolution.Hex()),
		Difficulty: job.Difficulty.Value(),
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

	var nonce, hash, soln string
	if err := json.Unmarshal(data[0], &nonce); err != nil || len(nonce) != (16+2) {
		return nil, fmt.Errorf("invalid nonce parameter")
	} else if err := json.Unmarshal(data[1], &hash); err != nil || len(hash) != (64+2) {
		return nil, fmt.Errorf("invalid hash parameter")
	} else if err := json.Unmarshal(data[2], &soln); err != nil || len(soln) != (336+2) {
		return nil, fmt.Errorf("invalid soln parameter")
	}

	var err error
	var nonceVal *types.Number
	var hashVal *types.Hash
	var solnVal *types.Solution
	if nonceVal, err = new(types.Number).SetFromHex(nonce); err != nil {
		return nil, err
	} else if hashVal, err = new(types.Hash).SetFromHex(hash); err != nil {
		return nil, err
	} else if solnVal, err = new(types.Solution).SetFromHex(soln); err != nil {
		return nil, err
	}

	work := &types.StratumWork{
		JobID:          hash,
		Nonce:          nonceVal,
		Hash:           hashVal,
		CuckooSolution: solnVal,
	}

	return work, nil
}

func (node Node) MarshalJob(rawID interface{}, job *types.StratumJob, cleanJobs bool) (interface{}, error) {
	id, err := json.Marshal(rawID)
	if err != nil {
		return nil, err
	}

	result := []interface{}{
		job.Header.PrefixedHex(),
		job.Seed.PrefixedHex(),
		node.GetShareDifficulty().TargetPrefixedHex(),
	}

	return rpc.NewResponse(id, result)
}

func (node Node) GetSubscribeResponse(id []byte, clientID, extraNonce string) (interface{}, error) {
	return nil, nil
}

func (node Node) GetDifficultyRequest() (interface{}, error) {
	return nil, nil
}

func (node Node) calculateBlockReward(height uint64, block *Block) (*big.Int, error) {
	blockReward := getBlockReward(height, uint64(len(block.Uncles)))

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

		// @TODO: this makes no sense to me, shouldn't it also be common.HexToBig?
		// @NOTE: it actually works, lets figure out why SetString works w/ leading 0x
		gasPrice, ok := new(big.Int).SetString(tx.GasPrice, 0)
		if !ok {
			return nil, fmt.Errorf("unable to parse gasPrice for tx %s", tx.Hash)
		}

		txFees.Add(txFees, new(big.Int).Mul(gasUsed, gasPrice))
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
		} else if common.StringsEqualInsensitive(block.Miner, node.address) {
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
			} else if common.StringsEqualInsensitive(uncle.Miner, node.address) {
				nonce, err := common.HexToUint64(uncle.Nonce)
				if err != nil {
					return err
				} else if nonce == types.Uint64Value(round.Nonce) {
					uncleReward := getUncleReward(checkHeight, round.Height)

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

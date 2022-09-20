package firo

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/blkbuilder"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

var blacklistTxIDs = map[string]bool{
	"a8772dbbb0ad2e73b4bde3013d4dde472b1cc35563842d46ea5f187e5a774efc": true,
}

func (node Node) GetBlockExplorerURL(round *pooldb.Round) string {
	hash, err := node.getBlockHash(round.Height)
	if err == nil {
		if node.mainnet {
			return fmt.Sprintf("https://explorer.firo.org/block/%s", hash)
		}
		return fmt.Sprintf("https://testexplorer.firo.org/block/%s", hash)
	}
	return ""
}

func (node Node) getStatusByHost(hostID string) (uint64, bool, error) {
	info, err := node.getBlockchainInfo(hostID)
	if err != nil {
		return 0, false, err
	}

	height := info.Blocks
	syncing := info.VerificationProgress < 0.9999 || info.Blocks != info.Headers

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
	if start > end {
		return nil, fmt.Errorf("invalid range")
	}

	heights := make([]uint64, end-start+1)
	for i := range heights {
		heights[i] = start + uint64(i)
	}

	hashes, err := node.getBlockHashMany(heights)
	if err != nil {
		return nil, err
	}

	rawBlocks, err := node.getBlockMany(hashes)
	if err != nil {
		return nil, err
	}

	devRewards, err := node.getCurrentDevRewards()
	if err != nil {
		return nil, err
	}

	blocks := make([]*tsdb.RawBlock, len(rawBlocks))
	for i, block := range rawBlocks {
		coinbaseTx, err := node.getSpecialTxesCoinbase(block.Hash)
		if err != nil {
			return nil, err
		}

		value, err := node.getRewardsFromTX(coinbaseTx, devRewards)
		if err != nil {
			return nil, err
		}

		blocks[i] = &tsdb.RawBlock{
			ChainID:    node.Chain(),
			Height:     start + uint64(i),
			Value:      float64(value),
			Difficulty: block.Difficulty,
			TxCount:    uint64(len(block.Transactions)),
			Timestamp:  time.Unix(block.Time, 0),
		}
	}

	return blocks, nil
}

func (node Node) getCurrentDevRewards() ([]uint64, error) {
	devRewards := make([]uint64, len(node.devWalletAmounts))
	for i, reward := range node.devWalletAmounts {
		devRewards[i] = reward
	}
	_, template, err := node.getBlockTemplate()
	if err != nil {
		return nil, err
	}
	for _, znode := range template.ZNode {
		devRewards = append(devRewards, znode.Amount)
	}

	return devRewards, nil
}

func (node Node) getRewardsFromTX(tx *Transaction, devRewards []uint64) (uint64, error) {
	var amount uint64
	for _, input := range tx.Inputs {
		if len(input.Coinbase) > 0 {
			for _, out := range tx.Outputs {
				valBig, err := common.StringDecimalToBigint(out.Value.String(), node.GetUnits().Big())
				if err != nil {
					return amount, err
				}

				var isReward bool
				for i, devReward := range devRewards {
					if valBig.Uint64() == devReward {
						isReward = true
						devRewards = append(devRewards[:i], devRewards[i+1:]...)
						break
					}
				}

				if !isReward {
					amount += valBig.Uint64()
				}
			}
		}
	}

	return amount, nil
}

func (node Node) parseBlockTemplate(template *BlockTemplate) (*types.StratumJob, error) {
	addresses := append([]string{node.address}, node.devWalletAddresses...)
	amounts := append([]uint64{template.CoinbaseValue}, node.devWalletAmounts...)
	for _, znode := range template.ZNode {
		addresses = append(addresses, znode.Payee)
		amounts = append(amounts, znode.Amount)
	}

	coinbaseHex, coinbaseHash, err := GenerateCoinbase(addresses, amounts, template.Height,
		uint64(template.CurTime), nil, template.CoinbasePayload, node.prefixP2PKH)
	if err != nil {
		return nil, err
	}

	txHashes := [][]byte{coinbaseHash}
	txHexes := [][]byte{coinbaseHex}
	if node.mocked {
		for _, tx := range template.Transactions {
			txid := tx.TxID
			if txid == "" {
				txid = tx.Hash
			}

			if blacklistTxIDs[txid] {
				continue
			}

			byteHash, err := hex.DecodeString(txid)
			if err != nil {
				return nil, err
			}

			byteHex, err := hex.DecodeString(tx.Data)
			if err != nil {
				return nil, err
			}

			txHashes = append(txHashes, byteHash)
			txHexes = append(txHexes, byteHex)
		}
	}

	builder, err := blkbuilder.NewProgPowBuilder(template.Version, template.CurTime, uint32(template.Height),
		template.Bits, template.PreviousBlockHash, txHashes, txHexes)
	if err != nil {
		return nil, err
	}

	// @TODO: i dont like this very much
	header, headerHash, err := builder.SerializeHeader(nil)
	if err != nil {
		return nil, err
	}

	seedHash := crypto.EthashSeedHash(template.Height, epochLength)
	bits, err := strconv.ParseUint(template.Bits, 16, 64)
	if err != nil {
		return nil, err
	}

	job := &types.StratumJob{
		CoinbaseTxID: new(types.Hash).SetFromBytes(coinbaseHash),
		Header:       new(types.Hash).SetFromBytes(header),
		HeaderHash:   new(types.Hash).SetFromBytes(headerHash),
		Seed:         new(types.Hash).SetFromBytes(seedHash),
		Height:       new(types.Number).SetFromValue(template.Height),
		Difficulty:   new(types.Difficulty).SetFromBits(uint32(bits), node.GetMaxDifficulty()),
		BlockBuilder: builder,
	}

	return job, nil
}

func (node Node) JobNotify(ctx context.Context, interval time.Duration, jobCh chan *types.StratumJob, errCh chan error) {
	refreshTimer := time.NewTimer(interval)
	staticInterval := time.Minute * 5

	go func() {
		var lastHeight uint64
		var lastJob time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-refreshTimer.C:
				now := time.Now()
				hostID, template, err := node.getBlockTemplate()
				if err != nil {
					errCh <- err
				} else if lastHeight != template.Height || now.After(lastJob.Add(staticInterval)) {
					job, err := node.parseBlockTemplate(template)
					if err != nil {
						errCh <- err
					} else {
						job.HostID = hostID
						lastHeight = job.Height.Value()
						lastJob = now
						jobCh <- job
					}
				}

				refreshTimer.Reset(interval)
			}
		}
	}()
}

func (node Node) SubmitWork(job *types.StratumJob, work *types.StratumWork) (types.ShareStatus, *pooldb.Round, error) {
	mixDigest, digest, err := node.pow.Compute(work.Hash.Bytes(), job.Height.Value(), work.Nonce.Value())
	if err != nil {
		return types.RejectedShare, nil, err
	} else if bytes.Compare(job.HeaderHash.Bytes(), work.Hash.Bytes()) != 0 {
		return types.InvalidShare, nil, nil
	} else if bytes.Compare(mixDigest, work.MixDigest.Bytes()) != 0 {
		return types.InvalidShare, nil, nil
	}

	hash := new(types.Hash).SetFromBytes(digest)
	if !hash.MeetsDifficulty(node.GetShareDifficulty()) {
		return types.RejectedShare, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, nil, nil
	}

	serializedBlock, err := job.BlockBuilder.SerializeBlock(work)
	if err != nil {
		return types.AcceptedShare, nil, err
	}

	err = node.submitBlock(job.HostID, hex.EncodeToString(serializedBlock))
	if err != nil {
		return types.AcceptedShare, nil, err
	}

	round := &pooldb.Round{
		ChainID:      node.Chain(),
		Height:       job.Height.Value(),
		Hash:         hash.Hex(),
		MixDigest:    types.StringPtr(work.MixDigest.Hex()),
		Nonce:        types.Uint64Ptr(work.Nonce.Value()),
		CoinbaseTxID: types.StringPtr(job.CoinbaseTxID.Hex()),
		Difficulty:   job.Difficulty.Value(),
		Pending:      true,
		Mature:       false,
		Uncle:        false,
		Orphan:       false,
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

	var nonce, hash, mixdigest string
	if err := json.Unmarshal(data[2], &nonce); err != nil || len(nonce) != (16+2) {
		return nil, fmt.Errorf("invalid nonce parameter")
	} else if err := json.Unmarshal(data[3], &hash); err != nil || len(hash) != (64+2) {
		return nil, fmt.Errorf("invalid hash parameter")
	} else if err := json.Unmarshal(data[4], &mixdigest); err != nil || len(mixdigest) != (64+2) {
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
		WorkerID:  worker,
		JobID:     jobID,
		Nonce:     nonceVal,
		Hash:      hashVal,
		MixDigest: mixVal,
	}

	return work, nil
}

func (node Node) MarshalJob(id interface{}, job *types.StratumJob, cleanJobs bool) (interface{}, error) {
	result := []interface{}{
		job.ID,
		job.HeaderHash.Hex(),
		job.Seed.Hex(),
		node.GetShareDifficulty().TargetHex(),
		cleanJobs,
		job.Height.Value(),
		fmt.Sprintf("%08x", job.Difficulty.Bits()),
	}

	return rpc.NewRequestWithID(id, "mining.notify", result...)
}

func (node Node) GetSubscribeResponse(id []byte, clientID, extraNonce string) (interface{}, error) {
	return rpc.NewResponse(id, []interface{}{clientID, extraNonce})
}

func (node Node) GetDifficultyRequest() (interface{}, error) {
	return rpc.NewRequest("mining.set_target", node.GetShareDifficulty().TargetHex())
}

func (node Node) UnlockRound(round *pooldb.Round) error {
	if round.CoinbaseTxID == nil {
		return fmt.Errorf("block %d has no coinbase txid", round.Height)
	}

	blockHash, err := node.getBlockHash(round.Height)
	if err != nil {
		return err
	}

	block, err := node.getBlock(blockHash)
	if err != nil {
		return err
	}

	round.Uncle = false
	round.Orphan = true
	round.Pending = false
	round.Mature = false
	round.Spent = false

	coinbaseTxID := types.StringValue(round.CoinbaseTxID)
	if len(block.Transactions) > 0 && block.Transactions[0] == coinbaseTxID {
		coinbaseTx, err := node.getSpecialTxesCoinbase(block.Hash)
		if err != nil {
			return err
		}

		devRewards, err := node.getCurrentDevRewards()
		if err != nil {
			return err
		}

		value, err := node.getRewardsFromTX(coinbaseTx, devRewards)
		if err != nil {
			return err
		}

		round.Value = dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(value)}
		round.Hash = block.Hash
		round.Orphan = false
		round.CreatedAt = time.Unix(block.Time, 0)
	}

	return nil
}

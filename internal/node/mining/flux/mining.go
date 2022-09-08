package flux

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/blkbuilder"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetBlockExplorerURL(round *pooldb.Round) string {
	if node.mainnet {
		return fmt.Sprintf("https://explorer.runonflux.io/block/%s", round.Hash)
	}
	return fmt.Sprintf("https://testnet.runonflux.io/block/%s", round.Hash)
}

func (node Node) GetStatus() (uint64, bool, error) {
	info, err := node.getBlockchainInfo()
	if err != nil {
		return 0, false, err
	}

	height := info.Blocks
	syncing := info.VerificationProgress < 0.99999 || info.Blocks != info.Headers

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

	hashes, err := node.getBlockHashMany(heights)
	if err != nil {
		return nil, err
	}

	rawBlocks, err := node.getBlockMany(hashes)
	if err != nil {
		return nil, err
	}

	devRewards := []uint64{}
	_, template, err := node.getBlockTemplate()
	if err != nil {
		return nil, err
	}
	if template.CumulusFluxnodeAddress != "" && template.CumulusFluxnodePayout != 0 {
		devRewards = append(devRewards, template.CumulusFluxnodePayout)
	}
	if template.NimbusFluxnodeAddress != "" && template.NimbusFluxnodePayout != 0 {
		devRewards = append(devRewards, template.NimbusFluxnodePayout)
	}
	if template.StratusFluxnodeAddress != "" && template.StratusFluxnodePayout != 0 {
		devRewards = append(devRewards, template.StratusFluxnodePayout)
	}

	blocks := make([]*tsdb.RawBlock, len(rawBlocks))
	for i, block := range rawBlocks {
		if len(block.Transactions) == 0 {
			return nil, fmt.Errorf("no transactions in block")
		}

		value, err := node.getRewardsFromTX(block.Transactions[0], devRewards)
		if err != nil {
			return nil, err
		}

		blocks[i] = &tsdb.RawBlock{
			ChainID:    node.Chain(),
			Height:     start + uint64(i),
			Value:      value,
			Difficulty: block.Difficulty,
			TxCount:    uint64(len(block.Transactions)),
			Timestamp:  time.Unix(block.Time, 0),
		}
	}

	return blocks, nil
}

func (node Node) getRewardsFromTX(tx *Transaction, devRewards []uint64) (float64, error) {
	var amount float64
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
					valFloat64, err := out.Value.Float64()
					if err != nil {
						return amount, err
					}
					amount += valFloat64
				}
			}
		}
	}

	return amount, nil
}

func (node Node) parseBlockTemplate(template *BlockTemplate) (*types.StratumJob, error) {
	addresses := []string{node.address}
	amounts := []uint64{template.MinerReward}

	if template.CumulusFluxnodeAddress != "" && template.CumulusFluxnodePayout != 0 {
		addresses = append(addresses, template.CumulusFluxnodeAddress)
		amounts = append(amounts, template.CumulusFluxnodePayout)
	}

	if template.NimbusFluxnodeAddress != "" && template.NimbusFluxnodePayout != 0 {
		addresses = append(addresses, template.NimbusFluxnodeAddress)
		amounts = append(amounts, template.NimbusFluxnodePayout)
	}

	if template.StratusFluxnodeAddress != "" && template.StratusFluxnodePayout != 0 {
		addresses = append(addresses, template.StratusFluxnodeAddress)
		amounts = append(amounts, template.StratusFluxnodePayout)
	}

	coinbaseHex, coinbaseHash, err := GenerateCoinbase(addresses, amounts, template.Height,
		"", node.prefixP2PKH, node.prefixP2SH)
	if err != nil {
		return nil, err
	}

	txHashes := [][]byte{coinbaseHash}
	txHexes := [][]byte{coinbaseHex}
	for _, tx := range template.Transactions {
		txid := tx.TxID
		if txid == "" {
			txid = tx.Hash
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

	builder, err := blkbuilder.NewEquihashBuilder(template.Version, template.CurTime, template.Bits,
		template.PreviousBlockHash, template.FinalSaplingRootHash, txHashes, txHexes)
	if err != nil {
		return nil, err
	}

	bits, err := strconv.ParseUint(template.Bits, 16, 64)
	if err != nil {
		return nil, err
	}

	job := &types.StratumJob{
		CoinbaseTxID: new(types.Hash).SetFromBytes(coinbaseHash),
		Height:       new(types.Number).SetFromValue(template.Height),
		Difficulty:   new(types.Difficulty).SetFromBits(uint32(bits), node.GetMaxDifficulty()),
		BlockBuilder: builder,
	}

	return job, nil
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
	header, headerHash, err := job.BlockBuilder.SerializeHeader(work)
	if err != nil {
		return types.RejectedShare, nil, err
	}

	validSolution, err := node.pow.Verify(header, work.EquihashSolution[1:])
	if err != nil {
		return types.RejectedShare, nil, err
	} else if !validSolution {
		return types.InvalidShare, nil, nil
	}

	hash := new(types.Hash).SetFromBytes(headerHash)
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
		Nonce:        types.Uint64Ptr(work.Nonce.Value()),
		Solution:     types.StringPtr(hex.EncodeToString(work.EquihashSolution)),
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

	var nonce, soln string
	if err := json.Unmarshal(data[3], &nonce); err != nil || len(nonce) != 56 {
		return nil, fmt.Errorf("invalid nonce parameter")
	} else if err := json.Unmarshal(data[4], &soln); err != nil || len(soln) != 106 {
		return nil, fmt.Errorf("invalid hash parameter")
	}

	nonceBytes, err := hex.DecodeString(extraNonce + nonce)
	if err != nil {
		return nil, err
	}

	solnValue, err := hex.DecodeString(soln)
	if err != nil {
		return nil, err
	}

	work := &types.StratumWork{
		WorkerID:         worker,
		JobID:            jobID,
		Nonce:            new(types.Number).SetFromBytes(nonceBytes),
		EquihashSolution: solnValue,
	}

	return work, nil
}

func (node Node) MarshalJob(id interface{}, job *types.StratumJob, cleanJobs bool) (interface{}, error) {
	partialJob := job.BlockBuilder.PartialJob()
	result := append([]interface{}{job.ID}, partialJob...)
	result = append(result, cleanJobs, "125_4", "ZelProof")

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

	if block.Hash == round.Hash {
		coinbaseTxID := types.StringValue(round.CoinbaseTxID)
		if len(block.Transactions) == 0 {
			return nil
		} else if tx := block.Transactions[0]; tx.TxID != coinbaseTxID && tx.Hash != coinbaseTxID {
			return nil
		}

		round.Orphan = false
		round.CreatedAt = time.Unix(block.Time, 0)
	}

	return nil
}

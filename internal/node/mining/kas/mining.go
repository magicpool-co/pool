package kas

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/crypto/blkbuilder"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
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

	tip, err := node.getBlock(hostID, tipHash)
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
	return nil, nil
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
	staticInterval := time.Second * 15

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
				} else if lastHash != template.HashMerkleRoot || now.After(lastJob.Add(staticInterval)) {
					job, err := node.parseBlockTemplate(template)
					if err != nil {
						errCh <- err
					} else {
						job.HostID = hostID
						// @TODO: is using hashMerkleRoot actually an acceptable way to check?
						// the reality is that it should probably be done through always parsing the template
						lastHash = template.HashMerkleRoot
						lastJob = now
						jobCh <- job
					}
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
	// @TODO: this should technically happen after authorize, maybe its okay here?
	return rpc.NewRequest("mining.set_extranonce", extraNonce, 6)
}

func (node Node) GetDifficultyRequest() (interface{}, error) {
	return rpc.NewRequest("mining.set_difficulty", 10)
}

func (node Node) UnlockRound(round *pooldb.Round) error {
	return nil
}
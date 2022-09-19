package ae

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/base58"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetBlockExplorerURL(round *pooldb.Round) string {
	if node.mainnet {
		return fmt.Sprintf("https://explorer.aeternity.io/generations/%d", round.Height)
	}
	return fmt.Sprintf("https://explorer.testnet.aeternity.io/generations/%d", round.Height)
}

func (node Node) GetStatus() (uint64, bool, error) {
	status, err := node.getStatus()
	if err != nil {
		return 0, false, err
	} else if status == nil {
		return 0, false, fmt.Errorf("empty status")
	}

	return status.TopBlockHeight, status.Syncing, nil
}

func (node Node) PingHosts() ([]string, []uint64, []bool, []error) {
	return nil, nil, nil, nil
}

func (node Node) GetBlocks(start, end uint64) ([]*tsdb.RawBlock, error) {
	if start > end {
		return nil, fmt.Errorf("invalid range")
	}

	blocks := make([]*tsdb.RawBlock, end-start+1)
	for i := range blocks {
		height := start + uint64(i)
		block, err := node.getBlock(height)
		if err != nil {
			return nil, err
		} else if block == nil {
			return nil, fmt.Errorf("empty block")
		}

		unitsBigFloat := new(big.Float).SetInt(node.GetUnits().Big())
		rewardBigInt, err := node.getBlockReward(height)
		if err != nil {
			return nil, err
		}
		rewardsBigFloat := new(big.Float).SetInt(rewardBigInt)
		rewardFloat64, _ := new(big.Float).Quo(rewardsBigFloat, unitsBigFloat).Float64()
		difficulty := new(types.Difficulty).SetFromBits(block.Target, node.GetMaxDifficulty())

		blocks[i] = &tsdb.RawBlock{
			ChainID:    node.Chain(),
			Height:     height,
			Value:      rewardFloat64,
			Difficulty: float64(difficulty.Value()),
			TxCount:    0,
			Timestamp:  time.Unix(block.Time/1000, 0),
		}
	}

	return blocks, nil
}

func (node Node) getBlockTemplate() (*types.StratumJob, error) {
	hostID, template, err := node.getPendingBlock()
	if err != nil {
		return nil, err
	} else if template == nil || len(template.Hash) < 3 {
		return nil, fmt.Errorf("empty template")
	}

	hash, err := base58.Decode(template.Hash[3:])
	if err != nil {
		return nil, err
	}

	job := &types.StratumJob{
		HostID:     hostID,
		HeaderHash: new(types.Hash).SetFromBytes(hash[:len(hash)-4]),
		Height:     new(types.Number).SetFromValue(template.Height),
		Difficulty: new(types.Difficulty).SetFromBits(template.Target, node.GetMaxDifficulty()),
		Data:       template,
	}

	return job, nil
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

	return crypto.Blake2b256(buf)
}

func generateHeader(hash []byte, nonce uint64) []byte {
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, nonce)
	hashEncoded := []byte(base64.StdEncoding.EncodeToString(hash))
	nonceEncoded := []byte(base64.StdEncoding.EncodeToString(nonceBytes))
	header := append(hashEncoded, append(nonceEncoded, make([]byte, 24)...)...)

	return header
}

func (node Node) SubmitWork(job *types.StratumJob, work *types.StratumWork) (types.ShareStatus, *pooldb.Round, error) {
	header := generateHeader(job.HeaderHash.Bytes(), work.Nonce.Value())
	validSolution, err := node.pow.Verify(header, work.CuckooSolution.Data())
	if err != nil {
		return types.RejectedShare, nil, err
	} else if !validSolution {
		return types.InvalidShare, nil, nil
	}

	hash := new(types.Hash).SetFromBytes(hashSolution(work.CuckooSolution.Data()))
	if !hash.MeetsDifficulty(node.GetShareDifficulty()) {
		return types.RejectedShare, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, nil, nil
	}

	// @TODO: this is horrible
	template, ok := job.Data.(*Block)
	if !ok {
		return types.AcceptedShare, nil, fmt.Errorf("unable to cast pending block: %v", job.Data)
	}

	body := map[string]interface{}{
		"hash":          template.Hash,
		"height":        job.Height.Value(),
		"prev_hash":     template.PrevHash,
		"prev_key_hash": template.PrevKeyHash,
		"state_hash":    template.StateHash,
		"miner":         template.Miner,
		"beneficiary":   template.Beneficiary,
		"target":        template.Target,
		"pow":           work.CuckooSolution.Data(),
		"nonce":         work.Nonce.Value(),
		"time":          template.Time,
		"version":       template.Version,
		"info":          template.Info,
	}

	err = node.postBlock(job.HostID, body)
	if err != nil {
		return types.AcceptedShare, nil, err
	}

	round := &pooldb.Round{
		ChainID:    node.Chain(),
		Height:     job.Height.Value(),
		Hash:       hash.Hex(),
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
	if len(data) != 4 {
		return nil, fmt.Errorf("incorrect work array length")
	}

	var worker, jobID string
	if err := json.Unmarshal(data[0], &worker); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data[1], &jobID); err != nil {
		return nil, err
	}

	var nonce string
	var sol []interface{}
	if err := json.Unmarshal(data[2], &nonce); err != nil || len(nonce) != 8 {
		return nil, fmt.Errorf("invalid nonce parameter")
	} else if err := json.Unmarshal(data[3], &sol); err != nil || len(sol) != 42 {
		return nil, fmt.Errorf("invalid sol parameter")
	}

	nonceVal, err := new(types.Number).SetFromHex(extraNonce + nonce)
	if err != nil {
		return nil, err
	}

	soln := make([]uint64, len(sol))
	for i, raw := range sol {
		val, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("invalid sol parameter")
		}

		soln[i], err = strconv.ParseUint(val, 16, 64)
		if err != nil {
			return nil, err
		}
	}

	work := &types.StratumWork{
		WorkerID:       worker,
		JobID:          jobID,
		Nonce:          nonceVal,
		CuckooSolution: new(types.Solution).SetFromData(soln),
	}

	return work, nil
}

func (node Node) MarshalJob(id interface{}, job *types.StratumJob, cleanJobs bool) (interface{}, error) {
	result := []interface{}{
		job.ID,
		job.HeaderHash.Hex(),
		job.Height.Value(),
		fmt.Sprintf("%0x", job.Difficulty.Value()),
		cleanJobs,
	}

	return rpc.NewRequestWithID(id, "mining.notify", result...)
}

func (node Node) GetSubscribeResponse(id []byte, clientID, extraNonce string) (interface{}, error) {
	return rpc.NewResponse(id, []interface{}{nil, extraNonce, 4})
}

func (node Node) GetDifficultyRequest() (interface{}, error) {
	return rpc.NewRequest("mining.set_difficulty", node.GetShareDifficulty().Value())
}

func (node Node) getBlockReward(height uint64) (*big.Int, error) {
	fullURL := node.fallbackURL + "/mdw/v2/deltastats?scope=gen:" + strconv.FormatUint(height, 10)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Timeout: time.Second * 3}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	deltaStats := new(DeltaStats)
	err = json.NewDecoder(res.Body).Decode(&deltaStats)
	if err != nil {
		return nil, err
	} else if len(deltaStats.Data) != 1 {
		return nil, fmt.Errorf("mismatch data length on deltastats: %d: %d", height, len(deltaStats.Data))
	}

	rawBlockReward := string(deltaStats.Data[0].BlockReward)
	blockReward, ok := new(big.Int).SetString(rawBlockReward, 10)
	if !ok || blockReward.Cmp(new(big.Int)) != 1 {
		return nil, fmt.Errorf("unable to unmarshal block reward from deltastats: %s", rawBlockReward)
	}

	return blockReward, nil
}

func (node Node) UnlockRound(round *pooldb.Round) error {
	block, err := node.getBlock(round.Height)
	if err != nil {
		return err
	} else if block == nil {
		return fmt.Errorf("empty block")
	}

	round.Uncle = false
	round.Orphan = true
	round.Pending = false
	round.Mature = false
	round.Spent = false

	if block.Beneficiary == node.address {
		blockReward, err := node.getBlockReward(round.Height)
		if err != nil {
			return err
		}

		round.Value = dbcl.NullBigInt{Valid: true, BigInt: blockReward}
		round.Hash = block.Hash
		round.Orphan = false
		round.CreatedAt = time.Unix(block.Time/1000, 0)
	}

	return nil
}

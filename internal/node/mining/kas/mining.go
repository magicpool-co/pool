package kas

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
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

func (node Node) JobNotify(ctx context.Context, interval time.Duration, jobCh chan *types.StratumJob, errCh chan error) {

}

func (node Node) SubmitWork(job *types.StratumJob, work *types.StratumWork) (types.ShareStatus, *pooldb.Round, error) {
	return types.RejectedShare, nil, nil
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

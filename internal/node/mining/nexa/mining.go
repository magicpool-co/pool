package nexa

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/wire"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const (
	standardMinerClientID = 0
	bzMinerClientID       = 1
	wildRigClientID       = 2
)

var (
	isBzminer = regexp.MustCompile(".*BzMiner.*")
	isWildRig = regexp.MustCompile(".*WildRig.*")
)

func (node Node) GetBlockExplorerURL(round *pooldb.Round) string {
	if node.mainnet {
		return fmt.Sprintf("https://explorer.nexa.org/block-height/%d", round.Height)
	}
	return fmt.Sprintf("https://testnet-explorer.nexa.org/block-height/%d", round.Height)
}

func (node Node) getStatusByHost(hostID string) (uint64, bool, error) {
	info, err := node.getBlockchainInfo(hostID)
	if err != nil {
		return 0, false, err
	}

	height := info.Blocks
	syncing := info.VerificationProgress < 0.9999 || info.Blocks != info.Headers
	node.rpcHost.SetHostSyncStatus(hostID, !syncing)

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

	blocks := make([]*tsdb.RawBlock, len(hashes))
	for i := 0; i < len(hashes); i += 25 {
		limit := i + 25
		if len(hashes) < limit {
			limit = len(hashes)
		}

		rawBlocks, err := node.getBlockMany(hashes[i:limit])
		if err != nil {
			return nil, err
		}

		for j, block := range rawBlocks {
			if len(block.Transactions) == 0 {
				return nil, fmt.Errorf("no transactions in block")
			}

			_, value, err := node.getRewardsFromTX(block.Transactions[0])
			if err != nil {
				return nil, err
			}
			valueBig := new(big.Int).SetUint64(value)

			blocks[i+j] = &tsdb.RawBlock{
				ChainID:    node.Chain(),
				Hash:       block.Hash,
				Height:     start + uint64(i+j),
				Value:      common.BigIntToFloat64(valueBig, node.GetUnits().Big()),
				Difficulty: block.Difficulty,
				TxCount:    uint64(len(block.Transactions)),
				Timestamp:  time.Unix(block.Time, 0),
			}
		}
	}

	return blocks, nil
}

func (node Node) GetBlocksByHash(startHash string, limit uint64) ([]*tsdb.RawBlock, error) {
	return nil, fmt.Errorf("GetBlocks: not implemented")
}

func (node Node) getRewardsFromTX(tx *Transaction) (string, uint64, error) {
	var txid string
	var amount uint64
	if len(tx.Inputs) == 0 {
		txid = tx.TxIdem
		for _, out := range tx.Outputs {
			valBig, err := common.StringDecimalToBigint(out.Value.String(), node.GetUnits().Big())
			if err != nil {
				return txid, amount, err
			}
			amount += valBig.Uint64()
		}
	}

	return txid, amount, nil
}

func (node Node) getBlockTemplate() (*types.StratumJob, error) {
	hostID, candidate, err := node.getMiningCandidate()
	if err != nil {
		return nil, err
	}

	info, err := node.getBlockchainInfo(hostID)
	if err != nil {
		return nil, err
	}

	bits, err := strconv.ParseUint(candidate.NBits, 16, 64)
	if err != nil {
		return nil, err
	}

	headerCommitment, err := new(types.Hash).SetFromHex(candidate.HeaderCommitment)
	if err != nil {
		return nil, err
	}

	job := &types.StratumJob{
		ID:         new(types.Number).SetFromValue(candidate.ID).Hex(),
		HostID:     hostID,
		Header:     headerCommitment,
		HeaderHash: headerCommitment,
		Height:     new(types.Number).SetFromValue(info.Blocks),
		Difficulty: new(types.Difficulty).SetFromBits(uint32(bits), node.GetMaxDifficulty()),
		Timestamp:  time.Now(),
		Data:       candidate.NBits,
	}

	return job, nil
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

func nexapow(headerCommitment, nonce []byte) ([]byte, error) {
	var buf bytes.Buffer
	var order = binary.BigEndian
	if err := wire.WriteElement(&buf, order, crypto.ReverseBytes(headerCommitment)); err != nil {
		return nil, err
	} else if err := wire.WriteVarBytes(&buf, order, nonce); err != nil {
		return nil, err
	}
	miningHash := crypto.Sha256d(buf.Bytes())
	signMsg := crypto.Sha256(miningHash)

	key := secp256k1.PrivKeyFromBytes(miningHash)
	sig := crypto.SchnorrSignBCH(key, signMsg)
	final := crypto.Sha256(sig.Serialize())
	final = crypto.ReverseBytes(final)

	return final, nil
}

func (node Node) SubmitWork(job *types.StratumJob, work *types.StratumWork) (types.ShareStatus, *types.Hash, *pooldb.Round, error) {
	jobID, err := new(types.Number).SetFromHex(job.ID)
	if err != nil {
		return types.InvalidShare, nil, nil, nil
	}

	digest, err := nexapow(job.HeaderHash.Bytes(), work.Nonce.BytesLE())
	if err != nil {
		return types.InvalidShare, nil, nil, nil
	}

	hash := new(types.Hash).SetFromBytes(digest)
	if !hash.MeetsDifficulty(node.GetShareDifficulty()) {
		return types.RejectedShare, nil, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, hash, nil, nil
	}

	height, blockHash, err := node.submitMiningSolution(job.HostID, jobID.Value(), work.Nonce.Hex())
	if err != nil {
		return types.AcceptedShare, hash, nil, err
	}

	round := &pooldb.Round{
		ChainID:    node.Chain(),
		Height:     height,
		Hash:       blockHash,
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
	switch len(data) {
	case 4, 5:
	default:
		dataRaw := make([]string, len(data))
		for i, d := range data {
			dataRaw[i] = string(d)
		}

		return nil, fmt.Errorf("incorrect work array length: %v", dataRaw)
	}

	var worker, jobID string
	if err := json.Unmarshal(data[0], &worker); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data[1], &jobID); err != nil {
		return nil, err
	}

	var nonce, timestamp string
	switch len(data) {
	case 4: // standard
		if err := json.Unmarshal(data[2], &nonce); err != nil {
			return nil, fmt.Errorf("invalid nonce parameter")
		} else if err := json.Unmarshal(data[3], &timestamp); err != nil {
			return nil, fmt.Errorf("invalid timestamp parameter")
		}
	case 5: // wildrig
		var extraNonce2 string
		if err := json.Unmarshal(data[2], &extraNonce2); err != nil {
			return nil, fmt.Errorf("invalid extraNonce2 parameter")
		} else if err := json.Unmarshal(data[3], &timestamp); err != nil {
			return nil, fmt.Errorf("invalid timestamp parameter")
		} else if err := json.Unmarshal(data[4], &nonce); err != nil {
			return nil, fmt.Errorf("invalid nonce parameter")
		}
		nonce = extraNonce2 + nonce
	}

	nonceBytes, err := hex.DecodeString(nonce)
	if err != nil {
		return nil, fmt.Errorf("invalid nonce parameter: %v", err)
	}

	nonceVal := new(types.Number).SetFromBytes(nonceBytes)

	work := &types.StratumWork{
		WorkerID: worker,
		JobID:    jobID,
		Nonce:    nonceVal,
	}

	return work, nil
}

func (node Node) MarshalJob(id interface{}, job *types.StratumJob, cleanJobs bool, clientType int) (interface{}, error) {
	var result []interface{}
	switch clientType {
	case standardMinerClientID, bzMinerClientID:
		timestamp := make([]byte, 8)
		binary.BigEndian.PutUint64(timestamp, uint64(job.Timestamp.Unix()))

		result = []interface{}{
			job.ID,
			job.Header.Hex(),
			job.Data, // raw bits
			hex.EncodeToString(timestamp),
			cleanJobs,
		}
	case wildRigClientID:
		result = []interface{}{
			job.ID,
			hex.EncodeToString(crypto.ReverseBytes(job.Header.Bytes())),
			job.Height.Value(),
			job.Data, // raw bits
		}
	}

	return rpc.NewRequestWithID(id, "mining.notify", result...)
}

func (node Node) GetClientType(minerClient string) int {
	if isBzminer.MatchString(minerClient) {
		return bzMinerClientID
	} else if isWildRig.MatchString(minerClient) {
		return wildRigClientID
	}

	return standardMinerClientID
}

func (node Node) GetSubscribeResponses(id []byte, clientID, extraNonce string) ([]interface{}, error) {
	var subscriptions = []interface{}{
		[]interface{}{"mining.set_difficulty", "b4b6693b72a50c7116db18d6497cac52"},
		[]interface{}{"mining.notify", "ae6812eb4cd7735a302a8a9dd95cf71f"},
	}

	res, err := rpc.NewResponse(id, []interface{}{subscriptions, extraNonce, 4})
	if err != nil {
		return nil, err
	}

	return []interface{}{res}, nil
}

func (node Node) GetAuthorizeResponses() ([]interface{}, error) {
	res, err := rpc.NewRequest("mining.set_difficulty", shareFactor)
	if err != nil {
		return nil, err
	}

	return []interface{}{res}, nil
}

func (node Node) UnlockRound(round *pooldb.Round) error {
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

	if block.Confirmations == -1 {
		round.Orphan = true
		return nil
	} else if uint64(block.Confirmations) < node.GetImmatureDepth() {
		return nil
	} else if block.Hash == round.Hash {
		txid, value, err := node.getRewardsFromTX(block.Transactions[0])
		if err != nil {
			return err
		}

		round.CoinbaseTxID = types.StringPtr(txid)
		round.Value = dbcl.NullBigInt{Valid: true, BigInt: new(big.Int).SetUint64(value)}
		round.Orphan = false
		round.CreatedAt = time.Unix(block.Time, 0)
	}

	return nil
}

func (node Node) MatureRound(round *pooldb.Round) ([]*pooldb.UTXO, error) {
	if round.Pending || round.Orphan || round.Mature {
		return nil, nil
	} else if !round.Value.Valid {
		return nil, fmt.Errorf("no value for round %d", round.ID)
	} else if round.CoinbaseTxID == nil {
		return nil, fmt.Errorf("no coinbase txid for round %d", round.ID)
	}

	block, err := node.getBlock(round.Hash)
	if err != nil {
		return nil, err
	} else if block.Height != round.Height {
		return nil, fmt.Errorf("mismatch on round and block height for round %d", round.ID)
	} else if block.Confirmations == -1 {
		round.Orphan = true
		return nil, nil
	} else if uint64(block.Confirmations) < node.GetMatureDepth() {
		return nil, nil
	}

	round.Mature = true

	utxos := []*pooldb.UTXO{
		&pooldb.UTXO{
			ChainID: round.ChainID,
			Value:   round.Value,
			TxID:    types.StringValue(round.CoinbaseTxID),
			Index:   0,
			Active:  true,
			Spent:   false,
		},
	}

	return utxos, nil
}

package rvn

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/tsdb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto"
	"github.com/magicpool-co/pool/pkg/crypto/blkbuilder"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

func (node Node) GetBlockExplorerURL(round *pooldb.Round) string {
	if node.mainnet {
		return fmt.Sprintf("https://ravencoin.network/block/%s", round.Hash)
	}
	return fmt.Sprintf("https://rvnt.cryptoscope.io/block/?blockhash=%s", round.Hash)
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

			value, err := node.getRewardsFromTX(block.Transactions[0])
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

func (node Node) getRewardsFromTX(tx *Transaction) (uint64, error) {
	var amount uint64
	for _, input := range tx.Inputs {
		if len(input.Coinbase) > 0 {
			for _, out := range tx.Outputs {
				valBig, err := common.StringDecimalToBigint(out.Value.String(), node.GetUnits().Big())
				if err != nil {
					return amount, err
				}
				amount += valBig.Uint64()
			}
		}
	}

	return amount, nil
}

func (node Node) parseBlockTemplate(template *BlockTemplate) (*types.StratumJob, error) {
	coinbaseHex, coinbaseHash, err := GenerateCoinbase(node.address, template.CoinbaseValue,
		template.Height, "", template.DefaultWitnessCommitment, node.prefixP2PKH)
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

func (node Node) JobNotify(ctx context.Context, interval time.Duration) chan *types.StratumJob {
	jobCh := make(chan *types.StratumJob)
	ticker := time.NewTicker(interval)
	staticInterval := time.Minute

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
				hostID, template, err := node.getBlockTemplate()
				if err != nil {
					node.logger.Error(err)
				} else if lastHeight != template.Height || now.After(lastJob.Add(staticInterval)) {
					job, err := node.parseBlockTemplate(template)
					if err != nil {
						node.logger.Error(err)
					} else {
						job.HostID = hostID
						lastHeight = job.Height.Value()
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
	mixDigest, digest, err := node.pow.Compute(work.Hash.Bytes(), job.Height.Value(), work.Nonce.Value())
	if err != nil {
		return types.RejectedShare, nil, nil, err
	} else if bytes.Compare(job.HeaderHash.Bytes(), work.Hash.Bytes()) != 0 {
		return types.InvalidShare, nil, nil, nil
	} else if bytes.Compare(mixDigest, work.MixDigest.Bytes()) != 0 {
		return types.InvalidShare, nil, nil, nil
	}

	hash := new(types.Hash).SetFromBytes(digest)
	if !hash.MeetsDifficulty(node.GetShareDifficulty()) {
		return types.RejectedShare, nil, nil, nil
	} else if !hash.MeetsDifficulty(job.Difficulty) {
		return types.AcceptedShare, hash, nil, nil
	}

	serializedBlock, err := job.BlockBuilder.SerializeBlock(work)
	if err != nil {
		return types.AcceptedShare, hash, nil, err
	}

	err = node.submitBlock(job.HostID, hex.EncodeToString(serializedBlock))
	if err != nil {
		return types.AcceptedShare, hash, nil, err
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

func (node Node) MarshalJob(id interface{}, job *types.StratumJob, cleanJobs bool, clientType int) (interface{}, error) {
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

func (node Node) GetClientType(minerClient string) int {
	return 0
}

func (node Node) GetSubscribeResponses(id []byte, clientID, extraNonce string) ([]interface{}, error) {
	res, err := rpc.NewResponse(id, []interface{}{clientID, extraNonce})
	if err != nil {
		return nil, err
	}

	return []interface{}{res}, nil
}

func (node Node) GetAuthorizeResponses() ([]interface{}, error) {
	res, err := rpc.NewRequest("mining.set_target", node.GetShareDifficulty().TargetHex())
	if err != nil {
		return nil, err
	}

	return []interface{}{res}, nil
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

	if block.Confirmations == -1 {
		round.Orphan = true
		return nil
	} else if uint64(block.Confirmations) < node.GetImmatureDepth() {
		return nil
	} else if block.Hash == round.Hash {
		coinbaseTxID := types.StringValue(round.CoinbaseTxID)
		if len(block.Transactions) == 0 {
			return nil
		} else if tx := block.Transactions[0]; tx.TxID != coinbaseTxID && tx.Hash != coinbaseTxID {
			return nil
		}

		value, err := node.getRewardsFromTX(block.Transactions[0])
		if err != nil {
			return err
		}

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

package pool

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/crypto/tx/btctx"
	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
	"github.com/magicpool-co/pool/types"
)

func generateExtraNonce(size int, mocked bool) string {
	if mocked {
		var extraNonce1 string
		for i := 0; i < size; i++ {
			extraNonce1 += "ff"
		}
		return extraNonce1
	}

	rand.Seed(time.Now().UnixNano())
	extraNonce1 := make([]byte, size)
	rand.Read(extraNonce1)

	return hex.EncodeToString(extraNonce1)
}

func (p *Pool) validateAddress(chain, address string) (bool, bool) {
	var ethRegex = regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

	switch chain {
	case "BTC":
		_, err := btctx.AddressToScript(address, []byte{0x00}, []byte{0x05}, true)
		return true, err == nil
	case "ETH":
		return true, ethRegex.MatchString(address)
	case p.chain:
		return true, p.node.ValidateAddress(address)
	}

	return false, false
}

func (p *Pool) handleLogin(c *stratum.Conn, req *rpc.Request) []interface{} {
	if c.GetAuthorized() {
		return errInvalidAuthRequest(req.ID)
	}

	var username string
	if len(req.Params) < 1 {
		return errInvalidAuthRequest(req.ID)
	} else if err := json.Unmarshal(req.Params[0], &username); err != nil || len(username) == 0 {
		return errInvalidAuthRequest(req.ID)
	}

	var workerName string
	args := strings.Split(username, ".")
	if len(args) > 1 {
		workerName = args[1]
	} else if req.Worker != "" {
		workerName = req.Worker
	}

	if workerName == "" {
		workerName = "default"
	}

	// address formatting is chain:address for standard,
	// solo:chain:address for solo mining (if enabled)
	var isSolo bool
	addressChain := args[0]
	partial := strings.Split(addressChain, ":")
	switch len(partial) {
	case 2:
	case 3:
		if !p.soloEnabled || strings.ToLower(partial[0]) != "solo" {
			return errInvalidAddressFormatting(req.ID)
		}
		isSolo = true
		partial = partial[1:]
	default:
		return errInvalidAddressFormatting(req.ID)
	}
	chain := strings.ToUpper(partial[0])
	address := partial[1]

	// for certain native payout chains (cfx and kas), the addresses are prefixed, but no
	// non-native payout chains require this. so on top of checking for the native/non-native
	// payout chains, we also check for the prefix of the native chain and add it back to
	// the active address if it matches (this works even for kas, which has a different prefix
	// than the internal chain name, kaspa vs. kas).
	if prefix := p.node.GetAddressPrefix(); prefix != "" && strings.ToUpper(prefix) == chain {
		address = prefix + ":" + address
		chain = p.node.Chain()
	} else if chain == "ERGO" {
		chain = "ERG"
	}

	validChain, validAddress := p.validateAddress(chain, address)
	if !validChain {
		p.logger.Debug(fmt.Sprintf("invalid chain: %s", username))
		return errInvalidChain(req.ID)
	} else if !validAddress {
		p.logger.Debug(fmt.Sprintf("invalid address: %s", username))
		return errInvalidAddress(req.ID)
	} else if len(workerName) > 32 {
		return errWorkerNameTooLong(req.ID)
	}

	// fetch minerID from redis
	minerID, err := p.redis.GetMinerID(addressChain)
	if err != nil || minerID == 0 {
		if err != nil {
			p.logger.Error(err)
		}

		// check the writer db directly
		minerID, err = pooldb.GetMinerID(p.db.Writer(), chain, address)
		if err != nil || minerID == 0 {
			if err != nil {
				p.logger.Error(err)
			}

			miner := &pooldb.Miner{
				ChainID: chain,
				Address: address,
				Active:  false,
			}

			// attempt to insert the minerID
			minerID, err = pooldb.InsertMiner(p.db.Writer(), miner)
			if err != nil {
				p.logger.Error(err)
				return nil
			}
		}

		// set the minerID in redis
		if err := p.redis.SetMinerID(addressChain, minerID); err != nil {
			p.logger.Error(err)
		}
	}

	port := c.GetPort()
	diffFactor := p.portDiffIdx[port]
	if diffFactor < 1 {
		diffFactor = 1
	}

	c.SetMiner(addressChain)
	c.SetMinerID(minerID)
	c.SetSubscribed(true)
	c.SetAuthorized(true)
	c.SetDiffFactor(diffFactor)
	c.SetIsSolo(isSolo)
	c.SetReadDeadline(time.Time{})

	var workerID uint64
	workerID, err = p.redis.GetWorkerID(minerID, workerName)
	if err != nil || workerID == 0 {
		if err != nil {
			p.logger.Error(err, c.GetCompoundID())
		}

		// check the writer db directly
		workerID, err = pooldb.GetWorkerID(p.db.Writer(), minerID, workerName)
		if err != nil || workerID == 0 {
			if err != nil {
				p.logger.Error(err, c.GetCompoundID())
			}

			worker := &pooldb.Worker{
				MinerID: minerID,
				Name:    workerName,
				Active:  false,
			}

			// attempt to insert the workerID
			workerID, err = pooldb.InsertWorker(p.db.Writer(), worker)
			if err != nil {
				p.logger.Error(err, c.GetCompoundID())
				return nil
			}
		}

		// set the workerID in redis
		if err := p.redis.SetWorkerID(minerID, workerName, workerID); err != nil {
			p.logger.Error(err, c.GetCompoundID())
		}
	}
	c.SetWorker(workerName)
	c.SetWorkerID(workerID)

	// handle connect streaming
	if p.streamWriter != nil {
		p.streamWriter.WriteConnectEvent(c.GetMinerID(), c.GetWorker(),
			c.GetClient(), c.GetPort(), c.GetIsSolo())
	}

	var msgs []interface{}
	if p.forceErrorOnResponse {
		msgs = []interface{}{rpc.NewResponseForcedFromJSON(req.ID, common.JsonTrue)}
	} else {
		msgs = []interface{}{rpc.NewResponseFromJSON(req.ID, common.JsonTrue)}
	}

	authResponses, err := p.node.GetAuthorizeResponses(diffFactor)
	if err != nil {
		p.logger.Error(err, c.GetCompoundID())
		return msgs
	}
	msgs = append(msgs, authResponses...)

	job := p.jobManager.LatestJob()
	if job != nil {
		msg, err := p.node.MarshalJob(0, job, true, c.GetClientType(), c.GetDiffFactor())
		if err != nil {
			p.logger.Error(err, c.GetCompoundID())
			return msgs
		}
		msgs = append(msgs, msg)
	}

	go p.jobManager.AddConn(c)

	return msgs
}

func (p *Pool) handleSubmit(c *stratum.Conn, req *rpc.Request) (bool, error) {
	submitTime := time.Now()
	extraNonce := c.GetExtraNonce()
	work, err := p.node.ParseWork(req.Params, extraNonce)
	if err != nil {
		return false, err
	}

	// if len(extraNonce) > 0 {
	// 	nonce := work.Nonce.Hex()
	// 	if len(nonce) < len(extraNonce) || nonce[:len(extraNonce)] != extraNonce {
	// 		return false, fmt.Errorf("nonce %s does not match extranonce %s", nonce, extraNonce)
	// 	}
	// }

	// if solo mining, prefix the chain with "S" so
	// that we can differentiate in charts and whatnot
	chain := p.chain
	var soloMinerID uint64
	if c.GetIsSolo() {
		chain = p.soloChain
		soloMinerID = c.GetMinerID()
	}
	activeDiffFactor := c.GetDiffFactor()

	var shareStatus types.ShareStatus = types.RejectedShare
	var hash *types.Hash
	var round *pooldb.Round
	job, activeShare := p.jobManager.GetJob(work.JobID)
	if job != nil && activeShare {
		shareStatus, hash, round, err = p.node.SubmitWork(job, work, activeDiffFactor)
		if err != nil {
			return false, err
		}

		// if the share is rejected, check to see if the last difficulty factor
		// is less than the current one, in which case check to see if the share
		// meets the difficulty level of the prior difficulty
		if p.varDiffEnabled && shareStatus == types.RejectedShare && hash != nil {
			lastDiffFactor := c.GetLastDiffFactor()
			timeSince := time.Since(c.GetLastDiffFactorAt())
			if lastDiffFactor > 0 && lastDiffFactor < activeDiffFactor && timeSince < time.Second*30 {
				if hash.MeetsDifficulty(p.node.GetShareDifficulty(lastDiffFactor)) {
					shareStatus = types.AcceptedShare
					activeDiffFactor = lastDiffFactor
				}
			}
		}

		// special handing for IceRiver ASICs since sometimes they submit
		// a solution for the prior job instead of the job ID that is sent
		if p.chain == "KAS" && shareStatus == types.RejectedShare {
			jobID := work.JobID
			for i := 0; i < 2; i++ {
				job, activeShare = p.jobManager.GetPriorJob(jobID)
				if job == nil {
					break
				}

				jobID = job.ID
				if activeShare {
					shareStatus, hash, round, err = p.node.SubmitWork(job, work, activeDiffFactor)
					if err != nil {
						return false, err
					} else if shareStatus == types.AcceptedShare {
						break
					}
				}
			}
		}
	}

	// handle round
	if round != nil {
		go func() {
			defer p.recoverPanic()

			p.wg.Add(1)
			defer p.wg.Done()

			compoundID := c.GetCompoundID()
			round.Solo = c.GetIsSolo()
			if round.Solo {
				p.logger.Info("found valid solo block")
			} else {
				p.logger.Info("found valid block")
			}

			sharesIdx := make(map[uint64]uint64)
			var err error
			if soloMinerID == 0 {
				sharesIdx, err = p.redis.GetRoundShares(chain)
			} else {
				sharesIdx[c.GetMinerID()], err = p.redis.GetRoundSoloShares(chain, soloMinerID)
			}
			if err != nil {
				p.logger.Error(err, compoundID)
				return
			}

			// number of accepted, rejected, and invalid shares since last block (not PPLNS share number)
			round.AcceptedShares, round.RejectedShares, round.InvalidShares, err = p.redis.GetRoundShareCounts(chain, soloMinerID)
			if err != nil {
				p.logger.Error(err, compoundID)
				return
			}

			shareDiff := float64(p.node.GetShareDifficulty(1).Value())
			if p.chain == "NEXA" {
				shareDiff = 0.2
			} else if shareDiff == 0 {
				shareDiff = 1
			}

			roundDiff := round.Difficulty
			if roundDiff == 0 {
				roundDiff = 1
			}

			minedDiff := shareDiff * float64(round.AcceptedShares+1)
			round.Luck = 100 * (float64(roundDiff) / float64(minedDiff))
			round.MinerID = c.GetMinerID()
			roundID, err := pooldb.InsertRound(p.db.Writer(), round)
			if err != nil {
				p.logger.Error(err, compoundID)
				return
			}

			shares := make([]*pooldb.Share, 0)
			for minerID, shareCount := range sharesIdx {
				share := &pooldb.Share{
					RoundID: roundID,
					MinerID: minerID,
					Count:   shareCount,
				}

				shares = append(shares, share)
			}

			if err := pooldb.InsertShares(p.db.Writer(), shares...); err != nil {
				p.logger.Error(err, compoundID)
				return
			}

			if p.telegram != nil {
				explorerURL := p.node.GetBlockExplorerURL(round)
				p.telegram.NotifyNewBlockCandidate(p.chain, explorerURL, round.Height, round.Luck)
			}
		}()
	}

	if shareStatus == types.AcceptedShare {
		if hash == nil {
			p.logger.Error(fmt.Errorf("no hash returned for an accepted share"), c.GetCompoundID())
		} else {
			isUnique, err := p.redis.AddUniqueShare(chain, job.Height.Value(), hash.Hex())
			if err != nil {
				return false, err
			} else if !isUnique {
				shareStatus = types.RejectedShare
			}
		}
	}

	// handle share streaming
	if p.streamWriter != nil {
		targetDiff := p.node.GetShareDifficulty(activeDiffFactor).Value()
		go func() {
			status := shareStatus.String()
			var reason string
			if shareStatus == types.RejectedShare {
				if job == nil {
					reason = "job not found"
				} else if !activeShare {
					reason = "job too old"
				} else {
					reason = "difficulty too low"
				}
			}

			var shareDiff uint64
			if hash != nil {
				shareDiff = hash.Difficulty(p.node.GetMaxDifficulty())
			}

			p.streamWriter.WriteShareEvent(c.GetMinerID(), c.GetWorker(), c.GetClient(),
				c.GetPort(), c.GetIsSolo(), status, reason, shareDiff, targetDiff)
		}()
	}

	// handle vardiff
	if p.varDiffEnabled && shareStatus != types.InvalidShare {
		go func() {
			defer p.recoverPanic()

			newDiffFactor := c.SetLastShareAt(submitTime)
			if newDiffFactor == -1 {
				return
			}

			diffResponse, err := p.node.GetSetDifficultyResponse(newDiffFactor)
			if err != nil {
				p.logger.Error(err)
				return
			} else if diffResponse == nil {
				return
			}

			p.logger.Info(fmt.Sprintf("setting vardiff for miner %s: %d -> %d", c.GetCompoundID(), c.GetDiffFactor(), newDiffFactor))
			err = p.writeToConn(c, diffResponse)
			if err != nil {
				p.logger.Error(err)
				return
			}

			oldDiff := p.node.GetShareDifficulty(c.GetDiffFactor()).Value()
			newDiff := p.node.GetShareDifficulty(newDiffFactor).Value()
			c.SetDiffFactor(newDiffFactor)

			// handle retarget streaming
			if p.streamWriter != nil {
				p.streamWriter.WriteRetargetEvent(c.GetMinerID(), c.GetWorker(),
					c.GetClient(), c.GetPort(), c.GetIsSolo(), oldDiff, newDiff)
			}
		}()
	}

	// handle share
	go func() {
		defer p.recoverPanic()

		p.wg.Add(1)
		defer p.wg.Done()

		interval := p.getCurrentInterval(false)
		switch shareStatus {
		case types.AcceptedShare:
			err := p.redis.AddAcceptedShare(chain, interval, c.GetCompoundID(), soloMinerID, activeDiffFactor, p.windowSize)
			if err != nil {
				p.logger.Error(err, c.GetCompoundID())
				return
			}

			if p.metrics != nil {
				p.metrics.AddCounter("accepted_shares_total", float64(activeDiffFactor), chain)
			}

			// need to replace ":" with "|" for IPv6 compatibility
			ip := strings.ReplaceAll(c.GetIP(), ":", "|")
			id := c.GetCompoundID() + ":" + ip
			latency, _ := c.GetLatency()

			p.minerStatsMu.Lock()
			p.lastShareIndex[id] = submitTime.Unix()
			p.lastDiffIndex[id] = int64(activeDiffFactor)
			if latency > 0 {
				p.latencyValueIndex[id] += int64(latency)
				p.latencyCountIndex[id]++
			}
			p.minerStatsMu.Unlock()
		case types.RejectedShare:
			err := p.redis.AddRejectedShare(chain, interval, c.GetCompoundID(), soloMinerID, activeDiffFactor)
			if err != nil {
				p.logger.Error(err, c.GetCompoundID())
			} else if p.metrics != nil {
				p.metrics.AddCounter("rejected_shares_total", float64(activeDiffFactor), chain)
			}
		case types.InvalidShare:
			err := p.redis.AddInvalidShare(chain, interval, c.GetCompoundID(), soloMinerID, activeDiffFactor)
			if err != nil {
				p.logger.Error(err, c.GetCompoundID())
			} else if p.metrics != nil {
				p.metrics.AddCounter("invalid_shares_total", float64(activeDiffFactor), chain)
			}
		}
	}()

	return shareStatus == types.AcceptedShare, nil
}

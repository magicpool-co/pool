package pool

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/pkg/common"
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

func validateAddress(address, chain string) bool {
	var btcRegex = regexp.MustCompile("^(?:[13]{1}[a-km-zA-HJ-NP-Z1-9]{26,33}|bc1[a-z0-9]{39,59})$")
	var ethRegex = regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

	switch strings.ToUpper(chain) {
	case "BTC":
		return btcRegex.MatchString(address)
	case "ETH":
		return ethRegex.MatchString(address)
	}

	return false
}

func (p *Pool) handleLogin(c *stratum.Conn, req *rpc.Request) ([]interface{}, error) {
	if c.GetAuthorized() {
		return nil, nil
	}

	var username string
	if len(req.Params) < 1 {
		return nil, fmt.Errorf("invalid login parameters")
	} else if err := json.Unmarshal(req.Params[0], &username); err != nil || len(username) == 0 {
		return nil, fmt.Errorf("invalid username %s", req.Params[0])
	}

	var workerName string
	args := strings.Split(username, ".")
	if len(args) > 1 {
		workerName = args[1]
	} else if req.Worker != "" {
		workerName = req.Worker
	}

	// address formatting is {address:chain}
	addressChain := args[0]
	partial := strings.Split(addressChain, ":")
	if len(partial) != 2 {
		return nil, fmt.Errorf("invalid address:chain formatting: %v", username)
	} else if valid := validateAddress(partial[0], strings.ToUpper(partial[1])); !valid {
		return nil, fmt.Errorf("invalid address: %s", args[0])
	} else if len(workerName) > 32 {
		return nil, fmt.Errorf("invalid worker name: %s", username)
	}
	address := partial[0]
	chain := partial[1]

	// fetch minerID from redis
	minerID, err := p.redis.GetMinerID(addressChain)
	if minerID == 0 {
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

			// finally check the db directly, if no minerID exists then bail
			minerID, err = pooldb.GetMinerID(p.db.Reader(), address, chain)
			if err != nil {
				return nil, err
			}
		}

		// set the minerID in redis
		if err := p.redis.SetMinerID(addressChain, minerID); err != nil {
			p.logger.Error(err)
		}
	}

	var workerID uint64
	if workerName != "" {
		workerID, err = p.redis.GetWorkerID(minerID, workerName)
		if workerID == 0 {
			if err != nil {
				p.logger.Error(err)
			}

			worker := &pooldb.Worker{
				MinerID: minerID,
				Name:    workerName,
				Active:  false,
			}

			// attempt to insert the workerID
			workerID, err = pooldb.InsertWorker(p.db.Writer(), worker)
			if err != nil {
				p.logger.Error(err)

				// finally check the db directly, if no workerID exists then bail
				workerID, err = pooldb.GetWorkerID(p.db.Reader(), minerID, workerName)
				if err != nil {
					return nil, err
				}
			}

			// set the workerID in redis
			if err := p.redis.SetWorkerID(minerID, workerName, workerID); err != nil {
				p.logger.Error(err)
			}
		}
	}

	c.SetUsername(username)
	c.SetMinerID(minerID)
	c.SetWorkerID(workerID)
	c.SetReadDeadline(time.Time{})

	var msgs []interface{}
	if p.forceErrorOnResponse {
		msgs = []interface{}{rpc.NewResponseForcedFromJSON(req.ID, common.JsonTrue)}
	} else {
		msgs = []interface{}{rpc.NewResponseFromJSON(req.ID, common.JsonTrue)}
	}

	diffReq, err := p.node.GetDifficultyRequest()
	if err != nil {
		return msgs, err
	} else if diffReq != nil {
		msgs = append(msgs, diffReq)
	}

	job := p.jobManager.LatestJob()
	if job != nil {
		msg, err := p.node.MarshalJob(0, job, true)
		if err != nil {
			return msgs, err
		}
		msgs = append(msgs, msg)
	}

	go p.jobManager.AddConn(c)

	return msgs, nil
}

func (p *Pool) handleSubmit(c *stratum.Conn, req *rpc.Request) (bool, error) {
	submitTime := time.Now()
	extraNonce := c.GetExtraNonce()
	work, err := p.node.ParseWork(req.Params, extraNonce)
	if err != nil {
		return false, err
	}

	if len(extraNonce) > 0 {
		nonce := work.Nonce.Hex()
		if len(nonce) < len(extraNonce) || nonce[:len(extraNonce)] != extraNonce {
			return false, fmt.Errorf("nonce %s does not match extranonce %s", nonce, extraNonce)
		}
	}

	var shareStatus types.ShareStatus
	var round *pooldb.Round
	job, activeShare := p.jobManager.GetJob(work.JobID)
	if job != nil && activeShare {
		shareStatus, round, err = p.node.SubmitWork(job, work)
		if err != nil {
			return false, err
		}
	}

	// handle round
	if round != nil {
		go func() {
			defer p.recoverPanic()

			p.wg.Add(1)
			defer p.wg.Done()

			p.logger.Info("found valid block")
			sharesIdx, err := p.redis.GetRoundShares(p.chain)
			if err != nil {
				p.logger.Error(err)
				return
			}

			// number of accepted, rejected, and invalid shares since last block (not PPLNS share number)
			round.AcceptedShares, round.RejectedShares, round.InvalidShares, err = p.redis.GetRoundShareCounts(p.chain)
			if err != nil {
				p.logger.Error(err)
				return
			}

			shareDiff := p.node.GetShareDifficulty().Value()
			if shareDiff == 0 {
				shareDiff = 1
			}

			roundDiff := round.Difficulty
			if roundDiff == 0 {
				roundDiff = 1
			}

			minedDiff := shareDiff * (round.AcceptedShares + 1)
			round.Luck = 100 * (float32(roundDiff) / float32(minedDiff))
			round.MinerID = c.GetMinerID()
			roundID, err := pooldb.InsertRound(p.db.Writer(), round)
			if err != nil {
				p.logger.Error(err)
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
				p.logger.Error(err)
				return
			}

			if p.telegram != nil {
				explorerURL := p.node.GetBlockExplorerURL(round)
				p.telegram.NotifyNewBlockCandidate(p.chain, explorerURL, round.Height, round.Luck)
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
			err := p.redis.AddAcceptedShare(p.chain, interval, c.GetCompoundID(), p.windowSize)
			if err != nil {
				p.logger.Error(err)
				return
			}

			if p.metrics != nil {
				p.metrics.IncrementCounter("accepted_shares_total", p.chain)
			}

			p.lastShareMu.Lock()
			if _, ok := p.lastShareIndex[c.GetCompoundID()]; !ok {
				p.lastShareIndex[c.GetCompoundID()] = make(map[string]time.Time)
			}
			p.lastShareIndex[c.GetCompoundID()][c.GetIP()] = submitTime
			p.lastShareMu.Unlock()
		case types.RejectedShare:
			err := p.redis.AddRejectedShare(p.chain, interval, c.GetCompoundID())
			if err != nil {
				p.logger.Error(err)
			} else if p.metrics != nil {
				p.metrics.IncrementCounter("rejected_shares_total", p.chain)
			}
		case types.InvalidShare:
			err := p.redis.AddInvalidShare(p.chain, interval, c.GetCompoundID())
			if err != nil {
				p.logger.Error(err)
			} else if p.metrics != nil {
				p.metrics.IncrementCounter("invalid_shares_total", p.chain)
			}
		}
	}()

	return shareStatus == types.AcceptedShare, nil
}

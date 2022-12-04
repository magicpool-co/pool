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
	case "ETH", "USDC":
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

	// address formatting is chain:address
	addressChain := args[0]
	partial := strings.Split(addressChain, ":")
	if len(partial) != 2 {
		return errInvalidAddressFormatting(req.ID)
	}

	chain := strings.ToUpper(partial[0])
	address := partial[1]
	if chain == "CFX" {
		address = "cfx:" + address
	}

	validChain, validAddress := p.validateAddress(chain, address)
	if !validChain {
		p.logger.Info(fmt.Sprintf("invalid chain: %s", username))
		return errInvalidChain(req.ID)
	} else if !validAddress {
		p.logger.Info(fmt.Sprintf("invalid address: %s", username))
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

	var workerID uint64
	if workerName != "" {
		workerID, err = p.redis.GetWorkerID(minerID, workerName)
		if err != nil || workerID == 0 {
			if err != nil {
				p.logger.Error(err)
			}

			// check the writer db directly
			workerID, err = pooldb.GetWorkerID(p.db.Writer(), minerID, workerName)
			if err != nil || workerID == 0 {
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
					return nil
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
	c.SetSubscribed(true)
	c.SetAuthorized(true)
	c.SetReadDeadline(time.Time{})

	var msgs []interface{}
	if p.forceErrorOnResponse {
		msgs = []interface{}{rpc.NewResponseForcedFromJSON(req.ID, common.JsonTrue)}
	} else {
		msgs = []interface{}{rpc.NewResponseFromJSON(req.ID, common.JsonTrue)}
	}

	authResponses, err := p.node.GetAuthorizeResponses(c.GetExtraNonce())
	if err != nil {
		p.logger.Error(err)
		return msgs
	}
	msgs = append(msgs, authResponses...)

	job := p.jobManager.LatestJob()
	if job != nil {
		msg, err := p.node.MarshalJob(0, job, true, c.GetClientType())
		if err != nil {
			p.logger.Error(err)
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
			round.Luck = 100 * (float64(roundDiff) / float64(minedDiff))
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

			// need to replace ":" with "|" for IPv6 compatibility
			ip := strings.ReplaceAll(c.GetIP(), ":", "|")

			p.lastShareMu.Lock()
			p.lastShareIndex[c.GetCompoundID()+":"+ip] = submitTime.Unix()
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

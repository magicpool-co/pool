package pool

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

type ProtocolHandler func(*stratum.Conn, *rpc.Request) error

/* router */

func (p *Pool) routeRequest(req *rpc.Request) ProtocolHandler {
	switch p.chain {
	case "AE", "ERGO", "FIRO", "FLUX", "RVN":
		switch req.Method {
		case "mining.subscribe":
			return p.subscribe
		case "mining.authorize":
			return p.login
		case "mining.submit":
			return p.submit
		case "mining.extranonce.subscribe":
			return p.extraNonce
		case "eth_submitHashrate":
			return p.submitHashrate
		}
	case "CFX":
		switch req.Method {
		case "mining.subscribe":
			return p.login
		case "mining.submit":
			return p.submit
		}
	case "ETH", "ETC":
		switch req.Method {
		case "eth_submitLogin":
			return p.login
		case "eth_submitWork":
			return p.submit
		case "eth_submitHashrate":
			return p.submitHashrate
		case "eth_getWork":
			return p.getWork
		}
	case "CTXC":
		switch req.Method {
		case "ctxc_submitLogin", "eth_submitLogin":
			return p.login
		case "ctxc_submitWork", "eth_submitWork":
			return p.submit
		case "ctxc_submitHashrate", "eth_submitHashrate":
			return p.submitHashrate
		case "ctxc_getWork", "eth_getWork":
			return p.getWork
		}
	}

	return nil
}

/* protocol functions */

func (p *Pool) subscribe(c *stratum.Conn, req *rpc.Request) error {
	if !c.GetSubscribed() {
		c.SetExtraNonce(generateExtraNonce(p.extraNonce1Size, p.node.Mocked()))
		c.SetSubscribed(true)
	}

	subID := strconv.FormatUint(c.GetID(), 16)
	subRes, err := p.node.GetSubscribeResponse(req.ID, subID, c.GetExtraNonce())
	if err != nil {
		return err
	} else if subRes != nil {
		err = c.Write(subRes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pool) extraNonce(c *stratum.Conn, req *rpc.Request) error {
	var res interface{}
	if p.forceErrorOnResponse {
		res = rpc.NewResponseForcedFromJSON(req.ID, common.JsonTrue)
	} else {
		res = rpc.NewResponseFromJSON(req.ID, common.JsonTrue)
	}

	return c.Write(res)
}

func (p *Pool) submitHashrate(c *stratum.Conn, req *rpc.Request) error {
	var res interface{}
	if len(req.Params) != 2 {
		res = errInvalidRequest(req.ID)
	} else {
		// store reported hashrate locally in a mutex protected map to
		// avoid an unnecessary of Redis calls, periodically push (in server.Serve)
		var rawHashrate string
		if err := json.Unmarshal(req.Params[0], &rawHashrate); err == nil {
			p.reportedMu.Lock()
			p.reportedIndex[c.GetCompoundID()] = rawHashrate
			p.reportedMu.Unlock()
		}

		if p.forceErrorOnResponse {
			res = rpc.NewResponseForcedFromJSON(req.ID, common.JsonTrue)
		} else {
			res = rpc.NewResponseFromJSON(req.ID, common.JsonTrue)
		}
	}

	return c.Write(res)
}

func (p *Pool) getWork(c *stratum.Conn, req *rpc.Request) error {
	work := p.jobManager.LatestJob()
	if work == nil {
		return fmt.Errorf("no template for %s", p.chain)
	}

	res, err := p.node.MarshalJob(req.ID, work, true)
	if err != nil {
		return err
	}

	return c.Write(res)
}

func (p *Pool) login(c *stratum.Conn, req *rpc.Request) error {
	msgs := p.handleLogin(c, req)
	for _, msg := range msgs {
		err := c.Write(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pool) submit(c *stratum.Conn, req *rpc.Request) error {
	validShare, err := p.handleSubmit(c, req)
	if err != nil {
		p.logger.Error(err)
	}

	var res interface{}
	if p.forceErrorOnResponse {
		res, err = rpc.NewResponseForced(req.ID, validShare)
	} else {
		res, err = rpc.NewResponse(req.ID, validShare)
	}

	if err != nil {
		return err
	}

	return c.Write(res)
}

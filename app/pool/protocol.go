package pool

import (
	"fmt"
	"strconv"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

type ProtocolHandler func(*stratum.Conn, *rpc.Request) error

/* router */

func (p *Pool) routeRequest(req *rpc.Request) ProtocolHandler {
	switch p.chain {
	case "AE", "ERGO", "FIRO", "FLUX", "KAS", "RVN":
		switch req.Method {
		case "mining.subscribe":
			return p.subscribe
		case "mining.authorize":
			return p.login
		case "mining.submit":
			return p.submit
		case "mining.extranonce.subscribe":
			return p.subscribeExtraNonce
		case "eth_submitHashrate":
			return p.submitHashrate
		}
	case "CFX":
		switch req.Method {
		case "mining.subscribe":
			return p.login
		case "mining.submit":
			return p.submit
		case "mining.extranonce.subscribe":
			return p.subscribeExtraNonce
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

	return p.logMethod
}

/* protocol functions */

func (p *Pool) logMethod(c *stratum.Conn, req *rpc.Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	p.logger.Info(fmt.Sprintf("recieved unknown request: %s: %s", req.Method, data))

	return nil
}

func (p *Pool) subscribe(c *stratum.Conn, req *rpc.Request) error {
	if len(req.Params) > 0 {
		var minerClient string
		err := json.Unmarshal(req.Params[0], &minerClient)
		if err == nil {
			c.SetClientType(p.node.GetClientType(minerClient))
		}
	}

	if !c.GetSubscribed() {
		c.SetExtraNonce(generateExtraNonce(p.extraNonce1Size, p.node.Mocked()))
		c.SetSubscribed(true)
	}

	subID := strconv.FormatUint(c.GetID(), 16)
	subResponses, err := p.node.GetSubscribeResponses(req.ID, subID, c.GetExtraNonce())
	if err != nil {
		return err
	}

	for _, msg := range subResponses {
		err = c.Write(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pool) subscribeExtraNonce(c *stratum.Conn, req *rpc.Request) error {
	if !c.GetExtraNonceSubscribed() {
		c.SetExtraNonceSubscribed(true)
	}

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

	res, err := p.node.MarshalJob(req.ID, work, true, c.GetClientType())
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

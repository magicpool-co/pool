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
	case "ERG", "FIRO", "FLUX", "KAS", "NEXA", "RVN":
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
	}

	return nil
}

/* protocol functions */

func (p *Pool) subscribe(c *stratum.Conn, req *rpc.Request) error {
	if len(req.Params) > 0 {
		var minerClient string
		err := json.Unmarshal(req.Params[0], &minerClient)
		if err == nil {
			c.SetClient(minerClient)
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
		err = p.writeToConn(c, msg)
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

	return p.writeToConn(c, res)
}

func (p *Pool) submitHashrate(c *stratum.Conn, req *rpc.Request) error {
	var res interface{}
	if p.forceErrorOnResponse {
		res = rpc.NewResponseForcedFromJSON(req.ID, common.JsonTrue)
	} else {
		res = rpc.NewResponseFromJSON(req.ID, common.JsonTrue)
	}

	return p.writeToConn(c, res)
}

func (p *Pool) getWork(c *stratum.Conn, req *rpc.Request) error {
	work := p.jobManager.LatestJob()
	if work == nil {
		return fmt.Errorf("no template for %s", p.chain)
	}

	res, err := p.node.MarshalJob(req.ID, work, true, c.GetClientType(), c.GetDiffFactor())
	if err != nil {
		return err
	}

	return p.writeToConn(c, res)
}

func (p *Pool) login(c *stratum.Conn, req *rpc.Request) error {
	msgs := p.handleLogin(c, req)
	for _, msg := range msgs {
		err := p.writeToConn(c, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pool) submit(c *stratum.Conn, req *rpc.Request) error {
	var validShare bool
	var err error
	if c.GetAuthorized() {
		validShare, err = p.handleSubmit(c, req)
		if err != nil {
			p.logger.Error(err, c.GetCompoundID())
		}
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

	return p.writeToConn(c, res)
}

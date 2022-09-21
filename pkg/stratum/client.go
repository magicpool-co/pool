package stratum

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

var (
	ErrNoActiveConn    = fmt.Errorf("no active conn")
	ErrResponseTimeout = fmt.Errorf("response timeout")
	ErrClientClosed    = fmt.Errorf("client closed")
	ErrRequestIDInUse  = fmt.Errorf("request id in use")
)

func recoverPanic(errCh chan error) {
	if r := recover(); r != nil {
		errCh <- fmt.Errorf("panic: %v: %s", r, debug.Stack())
	}
}

type Client struct {
	ctx             context.Context
	quit            chan struct{}
	host            string
	timeout         time.Duration
	reconnect       time.Duration
	conn            net.Conn
	connMu          sync.RWMutex
	requests        map[string]chan *rpc.Response
	requestsCounter uint64
	requestsMu      sync.RWMutex
	waiting         bool
	waitingMu       sync.Mutex
	waitingErr      chan error
}

func NewClient(ctx context.Context, host string, timeout, reconnect time.Duration) *Client {
	client := &Client{
		ctx:             ctx,
		quit:            make(chan struct{}),
		host:            host,
		timeout:         timeout,
		reconnect:       reconnect,
		requests:        make(map[string]chan *rpc.Response),
		requestsCounter: 1,
		waitingErr:      make(chan error),
	}

	return client
}

func (c *Client) connect(handshakeReqs []*rpc.Request, reqCh chan *rpc.Request, resCh chan *rpc.Response, errCh chan error) {
	var err error
	c.connMu.Lock()
	c.conn, err = net.Dial("tcp", c.host)
	c.connMu.Unlock()
	if err != nil {
		errCh <- err
		return
	}

	defer c.conn.Close()

	go func() {
		defer recoverPanic(errCh)

		err := c.sendHandshake(handshakeReqs)
		if err != nil {
			errCh <- err
			c.quit <- struct{}{}
		}

		c.waitingMu.Lock()
		if c.waiting {
			c.waitingErr <- err
		}
		c.waitingMu.Unlock()
	}()

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		select {
		case <-c.ctx.Done():
			return
		case <-c.quit:
			return
		default:
			var msg *rpc.Message
			if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
				errCh <- err
				continue
			}

			if msg.Method != "" {
				reqCh <- msg.ToRequest()
			} else {
				go func() {
					defer recoverPanic(errCh)

					res := msg.ToResponse()

					c.requestsMu.RLock()
					priorityCh, ok := c.requests[hex.EncodeToString(res.ID)]
					c.requestsMu.RUnlock()
					if !ok {
						priorityCh = resCh
					}

					priorityCh <- res
				}()
			}
		}
	}
}

func (c *Client) sendHandshake(reqs []*rpc.Request) error {
	for i, req := range reqs {
		res, err := c.WriteRequest(req)
		if err != nil {
			return err
		} else if len(reqs) == 1 || req.Method != "mining.subscribe" {
			if bytes.Compare(res.Result, common.JsonTrue) != 0 {
				return fmt.Errorf("server did not accept handshake request %d", i)
			}
		}
	}

	return nil
}

func (c *Client) Start(handshakeReqs []*rpc.Request) (chan *rpc.Request, chan *rpc.Response, chan error) {
	reqCh := make(chan *rpc.Request)
	resCh := make(chan *rpc.Response)
	errCh := make(chan error)

	go func() {
		defer recoverPanic(errCh)

		c.connect(handshakeReqs, reqCh, resCh, errCh)
		ticker := time.NewTicker(c.reconnect)
		// @TODO: this is just reconnecting immediately since the server read deadline is 1m
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				c.connect(handshakeReqs, reqCh, resCh, errCh)
			}
		}
	}()

	return reqCh, resCh, errCh
}

func (c *Client) ForceReconnect() {
	c.quit <- struct{}{}
}

func (c *Client) WaitForHandshake(timeout time.Duration) error {
	c.waitingMu.Lock()
	if c.waiting {
		c.waitingMu.Unlock()
		return fmt.Errorf("wait already active")
	}
	c.waiting = true
	c.waitingMu.Unlock()

	timer := time.NewTimer(timeout)
	for {
		select {
		case <-timer.C:
			return fmt.Errorf("wait timeout for handshake")
		case err := <-c.waitingErr:
			return err
		}
	}
}

func (c *Client) registerRequest(req *rpc.Request) (string, chan *rpc.Response, error) {
	c.requestsMu.Lock()
	defer c.requestsMu.Unlock()

	c.requestsCounter++
	if c.requestsCounter == math.MaxUint64 {
		c.requestsCounter = 1
	}

	var err error
	req.ID, err = json.Marshal(c.requestsCounter)
	if err != nil {
		return "", nil, err
	}

	hexID := hex.EncodeToString(req.ID)
	if _, ok := c.requests[hexID]; ok {
		return "", nil, ErrRequestIDInUse
	}

	c.requests[hexID] = make(chan *rpc.Response)

	return hexID, c.requests[hexID], nil
}

func (c *Client) deregisterRequest(id string) {
	c.requestsMu.Lock()
	delete(c.requests, id)
	c.requestsMu.Unlock()
}

func (c *Client) writeJSON(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.connMu.RLock()
	defer c.connMu.RUnlock()
	if c.conn == nil {
		return ErrNoActiveConn
	}

	_, err = c.conn.Write(append(data, '\n'))

	return err
}

func (c *Client) WriteRequest(req *rpc.Request) (*rpc.Response, error) {
	return c.WriteRequestWithTimeout(req, c.timeout)
}

func (c *Client) WriteRequestWithTimeout(req *rpc.Request, timeout time.Duration) (*rpc.Response, error) {
	hexID, ch, err := c.registerRequest(req)
	if err != nil {
		return nil, err
	}
	defer c.deregisterRequest(hexID)

	err = c.writeJSON(req)
	if err != nil {
		return nil, err
	}

	timer := time.NewTimer(timeout)
	for {
		select {
		case res := <-ch:
			if res.Error != nil {
				return res, res.Error
			}
			return res, nil
		case <-c.ctx.Done():
			return nil, ErrClientClosed
		case <-timer.C:
			return nil, ErrResponseTimeout
		}
	}
}

func (c *Client) WriteResponse(res *rpc.Response) error {
	return c.writeJSON(res)
}

package protowire

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/magicpool-co/pool/internal/log"
)

type Client struct {
	url            string
	timeout        time.Duration
	mu             sync.RWMutex
	logger         *log.Logger
	conn           *conn
	router         *router
	isConnected    uint32
	isReconnecting uint32
	isClosed       uint32
}

func NewClient(url string, timeout time.Duration, logger *log.Logger) (*Client, error) {
	c := &Client{
		url:     url,
		timeout: timeout,
		logger:  logger,
	}

	err := c.connect()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) disconnectHandler() {
	atomic.StoreUint32(&c.isConnected, 0)
	if atomic.LoadUint32(&c.isClosed) == 0 {
		c.disconnect()
		err := c.Reconnect()
		if err != nil {
			c.logger.Error(fmt.Errorf("kaspad grpc: reconnect: %v", err))
		}
	}
}

func (c *Client) errorHandler(err error) {
	c.logger.Error(fmt.Errorf("kaspad grpc: err: %v", err))
	c.disconnectHandler()
}

func (c *Client) connect() error {
	rtr := newRouter()
	grpcConn, err := newConn(c.url, c.timeout, rtr, c.logger, c.disconnectHandler, c.errorHandler)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.router = rtr
	c.conn = grpcConn
	atomic.StoreUint32(&c.isConnected, 1)

	return nil
}

func (c *Client) disconnect() {
	c.conn.disconnect()
}

func (c *Client) Send(raw interface{}) (interface{}, error) {
	req, ok := raw.(*KaspadMessage)
	if !ok {
		return nil, ErrUnknownMessage
	} else if atomic.LoadUint32(&c.isConnected) != 1 {
		var err = ErrClientNotConnected
		err = c.Reconnect()
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	err := c.router.outgoing.enqueue(req)
	if err != nil {
		return nil, err
	}

	cmd := req.getCmd()
	rte, ok := c.router.incoming[cmd]
	if !ok {
		return nil, ErrRouteNotFound
	}

	timer := time.NewTimer(c.timeout)
	for {
		select {
		case <-timer.C:
			return nil, ErrRouteTimedOut
		case res, ok := <-rte.ch:
			if !ok {
				return nil, ErrRouteClosed
			}
			return res, nil
		}
	}
}

func (c *Client) Reconnect() error {
	if atomic.LoadUint32(&c.isClosed) == 1 {
		return fmt.Errorf("cannot reconnect from a closed client")
	} else if !atomic.CompareAndSwapUint32(&c.isReconnecting, 0, 1) {
		return nil
	}
	defer atomic.StoreUint32(&c.isReconnecting, 0)

	if atomic.LoadUint32(&c.isConnected) == 1 {
		c.disconnect()
	}

	var err error
	for i := 0; i < 3; i++ {
		err = c.connect()
		if err == nil {
			return nil
		}
	}

	return err
}

//go:generate protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative rpc.proto messages.proto

package protowire

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"

	"github.com/magicpool-co/pool/internal/log"
)

// GRPCClient is a gRPC-based RPC client
type conn struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn

	stream RPC_MessageStreamClient
	// streamLock protects concurrent access to stream.
	// Note that it's an RWMutex. Despite what the name
	// implies, we use it to RLock() send() and receive() because
	// they can work perfectly fine in parallel, and Lock()
	// closeSend() because it must run alone.
	streamLock sync.RWMutex

	router            *router
	logger            *log.Logger
	disconnectHandler func()
	errorHandler      func(error)
	isConnected       uint32
}

func newGRPCConn(url string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return grpc.DialContext(ctx, url, grpc.WithInsecure(), grpc.WithBlock())

}

func newConn(
	url string,
	timeout time.Duration,
	rtr *router,
	logger *log.Logger,
	disconnectHandler func(),
	errorHandler func(error),
) (*conn, error) {
	grpcConn, err := newGRPCConn(url, timeout)
	if err != nil {
		return nil, err
	}

	rpcClient := NewRPCClient(grpcConn)
	stream, err := rpcClient.MessageStream(context.Background(),
		grpc.UseCompressor(gzip.Name),
		grpc.MaxCallRecvMsgSize(1024*1024*1024),
		grpc.MaxCallSendMsgSize(1024*1024*1024),
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &conn{
		ctx:               ctx,
		cancel:            cancel,
		conn:              grpcConn,
		stream:            stream,
		router:            rtr,
		logger:            logger,
		disconnectHandler: disconnectHandler,
		errorHandler:      errorHandler,
		isConnected:       1,
	}

	go c.sendLoop()
	go c.receiveLoop()

	return c, nil
}

func (c *conn) disconnect() {
	if !atomic.CompareAndSwapUint32(&c.isConnected, 1, 0) {
		return
	}

	c.cancel()
	c.closeSend()
}

func (c *conn) handleError(err error) {
	if errors.Is(err, io.EOF) {
		c.disconnectHandler()
		return
	} else if errors.Is(err, ErrRouteClosed) {
		c.disconnect()
	}

	c.errorHandler(err)
}

func (c *conn) receive() (*KaspadMessage, error) {
	// We use RLock here and in send() because they can work
	// in parallel. closeSend(), however, must not have either
	// receive() nor send() running while it's running.
	c.streamLock.RLock()
	defer c.streamLock.RUnlock()

	return c.stream.Recv()
}

func (c *conn) send(message *KaspadMessage) error {
	// We use RLock here and in receive() because they can work
	// in parallel. closeSend(), however, must not have either
	// receive() nor send() running while it's running.
	c.streamLock.RLock()
	defer c.streamLock.RUnlock()

	return c.stream.Send(message)
}

func (c *conn) closeSend() {
	c.streamLock.Lock()
	defer c.streamLock.Unlock()

	// ignore error because we don't really know what's the status of the connection
	_ = c.stream.CloseSend()
	_ = c.conn.Close()
}

func (c *conn) sendLoop() {
	defer c.logger.RecoverPanic()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.router.outgoing.ch:
			if !ok {
				return
			}

			err := c.send(msg)
			if err != nil {
				c.handleError(err)
				return
			}
		}
	}
}

func (c *conn) receiveLoop() {
	defer c.logger.RecoverPanic()

	for {
		msg, err := c.receive()
		if err != nil {
			c.handleError(err)
			return
		}

		cmd := msg.getCmd()
		if cmd == CmdUnknown {
			continue
		}

		rtr, ok := c.router.incoming[cmd]
		if !ok {
			c.handleError(ErrRouteNotFound)
			return
		}

		err = rtr.enqueue(msg)
		if err != nil {
			c.handleError(err)
			return
		}
	}
}

//go:generate protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative rpc.proto messages.proto

package protowire

import (
	"context"
	"errors"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"

	"github.com/magicpool-co/pool/internal/log"
)

// GRPCClient is a gRPC-based RPC client
type conn struct {
	ctx               context.Context
	cancel            context.CancelFunc
	conn              *grpc.ClientConn
	stream            RPC_MessageStreamClient
	router            *router
	logger            *log.Logger
	disconnectHandler func()
	errorHandler      func(error)
}

func newGRPCConn(url string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return grpc.DialContext(ctx, url, grpc.WithInsecure(), grpc.WithBlock())

}

func newConn(url string, timeout time.Duration, rtr *router, logger *log.Logger, disconnectHandler func(), errorHandler func(error)) (*conn, error) {
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
	}

	go c.sendLoop()
	go c.receiveLoop()

	return c, nil
}

func (c *conn) disconnect() error {
	return c.stream.CloseSend()
}

func (c *conn) handleError(err error) {
	if errors.Is(err, io.EOF) {
		c.disconnectHandler()
		return
	} else if errors.Is(err, ErrRouteClosed) {
		err = c.disconnect()
	}

	c.errorHandler(err)
}

func (c *conn) sendLoop() {
	defer c.logger.RecoverPanic()

	for {
		msg, ok := <-c.router.outgoing.ch
		if !ok {
			return
		}

		err := c.stream.Send(msg)
		if err != nil {
			c.handleError(err)
			return
		}
	}
}

func (c *conn) receiveLoop() {
	defer c.logger.RecoverPanic()

	for {
		msg, err := c.stream.Recv()
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

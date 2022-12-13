//go:generate protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative rpc.proto messages.proto

package protowire

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// GRPCClient is a gRPC-based RPC client
type Client struct {
	url            string
	timeout        time.Duration
	isConnected    uint32
	isReconnecting uint32
	stream         RPC_MessageStreamClient
}

var opts = []grpc.CallOption{
	grpc.UseCompressor(gzip.Name),
	grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024 * 5),
	grpc.MaxCallSendMsgSize(1024 * 1024 * 1024 * 5),
}

func newGRPCClientConn(url string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return grpc.DialContext(ctx, url, grpc.WithInsecure(), grpc.WithBlock())
}

func NewClient(url string, timeout time.Duration) (*Client, error) {
	conn, err := newGRPCClientConn(url, timeout)
	if err != nil {
		return nil, err
	}

	rpcClient := NewRPCClient(conn)
	stream, err := rpcClient.MessageStream(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	client := &Client{
		url:         url,
		timeout:     timeout,
		isConnected: 1,
		stream:      stream,
	}

	return client, nil
}

// Send is a helper function that sends the given request to the
// RPC server, accepts the first response that arrives back, and
// returns the response
func (c *Client) Send(raw interface{}) (interface{}, error) {
	request, ok := raw.(*KaspadMessage)
	if !ok {
		return nil, fmt.Errorf("unable to cast raw object as *KaspadMessage")
	} else if c.stream == nil {
		err := c.Reconnect()
		if err != nil {
			return nil, err
		}
	}

	err := c.stream.Send(request)
	if err != nil {
		return nil, err
	}

	return c.stream.Recv()
}

// Disconnects and from the RPC server
func (c *Client) Reconnect() error {
	swapped := atomic.CompareAndSwapUint32(&c.isReconnecting, 0, 1)
	if !swapped {
		return fmt.Errorf("reconnecting client")
	}
	defer atomic.StoreUint32(&c.isReconnecting, 0)

	connected := atomic.CompareAndSwapUint32(&c.isConnected, 1, 0)
	if connected {
		err := c.stream.CloseSend()
		atomic.StoreUint32(&c.isConnected, 0)
		if err != nil {
			return err
		}
	}

	newClient, err := NewClient(c.url, c.timeout)
	if err != nil {
		return err
	}
	c.stream = newClient.stream
	atomic.StoreUint32(&c.isConnected, 1)

	return nil
}

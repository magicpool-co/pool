//go:generate protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative rpc.proto messages.proto

package protowire

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// GRPCClient is a gRPC-based RPC client
type Client struct {
	url     string
	timeout time.Duration
	stream  RPC_MessageStreamClient
}

var opts = []grpc.CallOption{
	grpc.UseCompressor(gzip.Name),
	grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024),
	grpc.MaxCallSendMsgSize(1024 * 1024 * 1024),
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
		url:     url,
		timeout: timeout,
		stream:  stream,
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
		return nil, fmt.Errorf("no active stream running")
	}

	err := c.stream.Send(request)
	if err != nil {
		return nil, err
	}

	return c.stream.Recv()
}

// Disconnects and from the RPC server
func (c *Client) Reconnect() error {
	if c.stream != nil {
		err := c.stream.CloseSend()
		c.stream = nil
		if err != nil {
			return err
		}
	}

	newClient, err := NewClient(c.url, c.timeout)
	if err != nil {
		return err
	}
	c.stream = newClient.stream

	return nil
}

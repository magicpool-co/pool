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

	conn   *grpc.ClientConn
	stream RPC_MessageStreamClient
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
		url:         url,
		timeout:     timeout,
		isConnected: 1,

		conn:   conn,
		stream: stream,
	}

	return client, nil
}

// withGRPCTimeout runs f and returns its error.  If the timeout elapses first,
// it returns a context deadline exceeded error instead.
func (c *Client) sendWithTimeout(request *KaspadMessage) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- c.stream.Send(request)
		close(errChan)
	}()

	timer := time.NewTimer(c.timeout)
	select {
	case <-timer.C:
		go c.Reconnect()
		return fmt.Errorf("context deadline exceeded")
	case err := <-errChan:
		if !timer.Stop() {
			<-timer.C
		}
		return err
	}
}

// Send is a helper function that sends the given request to the
// RPC server, accepts the first response that arrives back, and
// returns the response
func (c *Client) Send(raw interface{}) (interface{}, error) {
	request, ok := raw.(*KaspadMessage)
	if !ok {
		return nil, fmt.Errorf("unable to cast raw object as *KaspadMessage")
	} else if atomic.LoadUint32(&c.isConnected) != 1 {
		err := c.Reconnect()
		if err != nil {
			return nil, err
		}
	}

	err := c.sendWithTimeout(request)
	if err != nil {
		return nil, err
	}

	return c.stream.Recv()
}

// Disconnects and from the RPC server
func (c *Client) Reconnect() error {
	swapped := atomic.CompareAndSwapUint32(&c.isReconnecting, 0, 1)
	if !swapped {
		timer := time.NewTimer(time.Millisecond * 500)
		for {
			select {
			case <-timer.C:
				reconnecting := atomic.LoadUint32(&c.isReconnecting)
				connected := atomic.LoadUint32(&c.isConnected)

				if reconnecting == 0 && connected == 1 {
					return nil
				}
				return fmt.Errorf("reconnecting client")
			}
		}
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

	// try three times before disconnecting for good
	var err error
	for i := 0; i < 3; i++ {
		var newClient *Client
		newClient, err = NewClient(c.url, c.timeout)
		if err != nil {
			continue
		}

		c.conn.Close()
		c.conn = newClient.conn
		c.stream = newClient.stream
		atomic.StoreUint32(&c.isConnected, 1)
		break
	}

	return err
}

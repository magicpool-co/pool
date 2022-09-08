package hostpool

import (
	"context"
	"sync"
	"time"

	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

type tcpConn struct {
	id      string
	ctx     context.Context
	mu      sync.RWMutex
	healthy bool
	enabled bool

	client *stratum.Client
}

// Run a healthcheck on the given TCP connection.
func (tc *tcpConn) healthCheck(healthCheck *TCPHealthCheck) int {
	var unit = time.Nanosecond
	var maxLatency = int(healthCheck.Timeout/unit) * 2

	start := time.Now()

	var err error
	if healthCheck.RPCRequest != nil {
		_, err = tc.exec(healthCheck.RPCRequest, healthCheck.Timeout)
	} else {
		return maxLatency
	}

	if err != nil {
		return maxLatency
	}

	tc.markHealthy(true)
	latency := int(time.Since(start) / unit)

	return latency
}

// Change a host's healthiness.
func (tc *tcpConn) markHealthy(healthy bool) {
	if !healthy {
		tc.client.ForceReconnect()
	}

	tc.mu.Lock()
	tc.healthy = healthy
	tc.mu.Unlock()
}

// Execute a request with a given timeout.
func (tc *tcpConn) exec(req *rpc.Request, timeout time.Duration) (*rpc.Response, error) {
	res, err := tc.client.WriteRequestWithTimeout(req, tcpTimeout)
	if err != nil {
		tc.markHealthy(false)
		return nil, err
	}

	res.HostID = tc.id

	return res, nil
}

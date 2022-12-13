package hostpool

import (
	"fmt"
	"sync"
	"time"

	"github.com/magicpool-co/pool/internal/log"
)

type grpcConn struct {
	id      string
	mu      sync.RWMutex
	errors  uint
	enabled bool
	synced  bool

	client GRPCClient
}

func (gc *grpcConn) healthy() bool {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	return gc.errors < 3
}

func (gc *grpcConn) usable(needsSynced bool) bool {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	healthy := gc.enabled && gc.errors < 3
	if needsSynced {
		return healthy && gc.synced
	}

	return healthy
}

// Run a healthcheck on the given TCP connection.
func (gc *grpcConn) healthCheck(healthCheck *GRPCHealthCheck, logger *log.Logger) int {
	var unit = time.Nanosecond
	var maxLatency = int(healthCheck.Timeout/unit) * 2

	start := time.Now()

	_, _, err := gc.exec(healthCheck.Request)
	if err != nil {
		logger.Error(fmt.Errorf("grpcconn: healthcheck: %s: %v", gc.id, err))
		return maxLatency
	}

	gc.markHealthy(true)
	latency := int(time.Since(start) / unit)

	return latency
}

// Change a host's healthiness.
func (gc *grpcConn) markHealthy(healthy bool) {
	if !healthy {
		gc.client.Reconnect()
	}

	gc.mu.Lock()
	defer gc.mu.Unlock()

	if healthy {
		gc.errors = 0
	} else {
		gc.errors++
	}
}

// Change a host's sync status.
func (gc *grpcConn) markSynced(synced bool) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.synced = synced
}

// Execute a request.
func (gc *grpcConn) exec(req interface{}) (interface{}, string, error) {
	res, err := gc.client.Send(req)
	if err != nil {
		gc.markHealthy(false)
		return nil, gc.id, err
	}

	return res, gc.id, nil
}

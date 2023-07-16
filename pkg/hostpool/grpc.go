package hostpool

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
)

const (
	grpcTimeout     = time.Second * 3
	grpcDialTimeout = time.Second * 2
)

type GRPCClient interface {
	URL() string
	Send(interface{}) (interface{}, error)
	Reconnect() error
}

type GRPCClientFactory func(string, time.Duration) (GRPCClient, error)

// HTTPPool represents a pool of HTTP hosts with methods to make standard HTTP calls.
type GRPCPool struct {
	ctx         context.Context
	mu          sync.RWMutex
	index       map[string]*grpcConn
	order       []string
	latencyIdx  map[string]int
	factory     GRPCClientFactory
	healthCheck *GRPCHealthCheck
	tunnel      *sshtunnel.SSHTunnel
	logger      *log.Logger
}

// GRPCHealthCheck specifies the definition of connection health for all
// connections in the pool. Interval is the frequency that the health check runs,
// Timeout is the connection timeout.
//
// Either the HTTP or RPC fields are required, if both are defined then HTTP
// will take precedence.
type GRPCHealthCheck struct {
	Interval time.Duration
	Timeout  time.Duration
	Request  interface{}
}

// NewGRPCPool creates a GRPC pool that manages connection health and optimizes
// for latency (while maintaining a degree of "stickiness" to avoid excessive reordering).
//
// The GRPCHealthCheck is not required, but without it the pool has little purpose. The
// health check interval defaults to one minute, the timeout defaults to
func NewGRPCPool(ctx context.Context, factory GRPCClientFactory, logger *log.Logger, healthCheck *GRPCHealthCheck, tunnel *sshtunnel.SSHTunnel) *GRPCPool {
	if healthCheck.Interval == 0 {
		healthCheck.Interval = time.Second * 30
	}
	if healthCheck.Timeout == 0 {
		healthCheck.Timeout = time.Second * 2
	}

	pool := &GRPCPool{
		ctx:         ctx,
		index:       make(map[string]*grpcConn),
		order:       make([]string, 0),
		factory:     factory,
		healthCheck: healthCheck,
		tunnel:      tunnel,
		logger:      logger,
	}

	if pool.healthCheck != nil {
		// run the healthcheck according to the given interval
		timer := time.NewTimer(pool.healthCheck.Interval)
		go func() {
			defer pool.logger.RecoverPanic()

			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					pool.runHealthCheck()
					timer.Reset(pool.healthCheck.Interval)
				}
			}
		}()
	}

	return pool
}

func (p *GRPCPool) GetAllHosts() []string {
	if p == nil {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	hosts := make([]string, 0)
	for id, client := range p.index {
		if client.enabled && client.healthy() {
			hosts = append(hosts, id)
		}
	}

	return hosts
}

// Adds a host to the pool. If the host already exists, nothing happens.
func (p *GRPCPool) AddHost(url string, port int) error {
	finalURL, id, err := parseURL(url, port, p.tunnel)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// add the host to the end of the list, mark as healthy to start
	// to avoid having no healthy hosts until the first healthcheck
	if _, ok := p.index[id]; !ok {
		client, err := p.factory(finalURL, grpcDialTimeout)
		if err != nil {
			return err
		}

		p.order = append(p.order, id)
		p.index[id] = &grpcConn{
			id:      id,
			enabled: true,
			synced:  true,
			client:  client,
		}
	}

	return nil
}

// Disables a host from being used for requests, though the host is not deleted (and can be enabled again).
func (p *GRPCPool) DisableHost(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.index[id]; ok {
		p.index[id].enabled = false
	}
}

// Enables a host, returning it to the active pool to be used in requests.
func (p *GRPCPool) EnableHost(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.index[id]; ok {
		p.index[id].enabled = true
	}
}

// Sets the sync status of a given host.
func (p *GRPCPool) SetHostSyncStatus(id string, synced bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if _, ok := p.index[id]; ok {
		p.index[id].markSynced(synced)
	}
}

// Executes a HTTP call to a specific host. If the host is not healthy, ErrNoHealthyHosts is returned.
// If the host is healthy, the error is returned.
func (p *GRPCPool) exec(hostID string, req interface{}, needsSynced bool) (interface{}, string, error) {
	// iterate through all host connections until no healthy connections
	// are left or a valid response is returned
	var res interface{}
	var err error
	var failed bool
	var count int
	for {
		count++
		gc := p.getConn(hostID, count, needsSynced)
		if gc == nil {
			failed = true
			break
		}

		res, hostID, err = gc.exec(req)
		if err != nil {
			p.logger.Error(fmt.Errorf("grpcpool: grpc: %v", err))
			continue
		}

		break
	}

	if failed && err == nil {
		err = ErrNoHealthyHosts
	}

	return res, hostID, err
}

// Executes a GRPC call. All healthy hosts will be attempted, if there are no healthy
// hosts to begin with, ErrNoHealthyHosts is returned. If there are healthy hosts, though
// all are actively unhealthy, the last error is returned.
func (p *GRPCPool) Exec(req interface{}) (interface{}, error) {
	res, _, err := p.exec("", req, false)
	return res, err
}

func (p *GRPCPool) ExecSticky(hostID string, req interface{}) (interface{}, string, error) {
	return p.exec(hostID, req, false)
}

func (p *GRPCPool) ExecSynced(req interface{}) (interface{}, string, error) {
	return p.exec("", req, true)
}

// pop the fastest healthy connection
func (p *GRPCPool) getConn(hostID string, count int, needsSynced bool) *grpcConn {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if hostID != "" && hostID != onceHostID {
		gc, ok := p.index[hostID]
		if ok && gc.healthy() {
			return gc
		}

		return nil
	} else if hostID == onceHostID && count > 1 {
		return nil
	}

	for _, id := range p.order {
		gc := p.index[id]
		if gc.usable(needsSynced) {
			return gc
		}
	}

	return nil
}

// run a healthcheck to update healthiness and reorder based on latency
func (p *GRPCPool) runHealthCheck() {
	p.mu.RLock()
	if len(p.order) == 0 {
		p.mu.RUnlock()
		return
	}

	// find the current best connection and the latency of all connections
	var latencyWg sync.WaitGroup
	var latencyMu sync.Mutex
	latencies := make(map[string]int, len(p.index))
	for id, gc := range p.index {
		latencyWg.Add(1)
		go func(id string, gc *grpcConn) {
			defer p.logger.RecoverPanic()
			defer latencyWg.Done()

			latency := gc.healthCheck(p.healthCheck, p.logger)
			latencyMu.Lock()
			latencies[id] = latency
			latencyMu.Unlock()
		}(id, gc)
	}

	p.mu.RUnlock()
	latencyWg.Wait()

	p.mu.Lock()
	p.order = processHealthCheck(p.order[0], latencies, nil)
	p.latencyIdx = latencies
	p.mu.Unlock()
}

func (p *GRPCPool) HandleInfoRequest(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	hosts := make([]map[string]interface{}, len(p.order))
	for i, id := range p.order {
		gc := p.index[id]
		gc.mu.Lock()
		url, errCount, synced := gc.client.URL(), gc.errors, gc.synced
		gc.mu.Unlock()
		hosts[i] = map[string]interface{}{
			"id":      id,
			"url":     url,
			"index":   i,
			"synced":  synced,
			"latency": time.Duration(p.latencyIdx[id]) * time.Nanosecond,
			"errors":  errCount,
		}
	}

	res := map[string]interface{}{
		"status": 200,
		"data":   hosts,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}

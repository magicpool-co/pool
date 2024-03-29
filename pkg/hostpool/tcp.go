package hostpool

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

const (
	tcpTimeout    = time.Second * 3
	tcpStickiness = 1.2 // 20% stickiness
)

type TCPHealthCheck struct {
	Interval   time.Duration
	Timeout    time.Duration
	RPCRequest *rpc.Request
}

type TCPPool struct {
	ctx         context.Context
	mu          sync.RWMutex
	index       map[string]*tcpConn
	order       []string
	counts      map[string]int
	latencyIdx  map[string]int
	healthCheck *TCPHealthCheck
	tunnel      *sshtunnel.SSHTunnel
	logger      *log.Logger

	errCh  chan error
	resCh  chan *rpc.Response
	reqChs map[string]chan *rpc.Request
}

func NewTCPPool(
	ctx context.Context,
	logger *log.Logger,
	healthCheck *TCPHealthCheck,
	tunnel *sshtunnel.SSHTunnel,
) *TCPPool {
	if healthCheck.Interval == 0 {
		healthCheck.Interval = time.Minute
	}
	if healthCheck.Timeout == 0 {
		healthCheck.Timeout = time.Second * 3
	}

	pool := &TCPPool{
		ctx:         ctx,
		index:       make(map[string]*tcpConn),
		order:       make([]string, 0),
		counts:      make(map[string]int),
		healthCheck: healthCheck,
		tunnel:      tunnel,
		logger:      logger,

		errCh:  make(chan error),
		resCh:  make(chan *rpc.Response),
		reqChs: make(map[string]chan *rpc.Request),
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

// Adds a host to the pool. If the host already exists, nothing happens.
func (p *TCPPool) AddHost(url string, port int) error {
	finalURL, id, err := parseURL(url, port, p.tunnel)
	if err != nil {
		return err
	}
	finalURL = strings.ReplaceAll(finalURL, "http://", "")
	finalURL = strings.ReplaceAll(finalURL, "https://", "")

	p.mu.Lock()
	defer p.mu.Unlock()

	// add the host to the end of the list, mark as healthy to start
	// to avoid having no healthy hosts until the first healthcheck
	if _, ok := p.index[id]; !ok {
		ctx := context.Background()
		p.order = append(p.order, id)
		p.index[id] = &tcpConn{
			id:      id,
			ctx:     ctx,
			enabled: true,
			synced:  true,
			client:  stratum.NewClient(ctx, finalURL, time.Second*10, time.Second*30),
		}

		reqCh, resCh, errCh := p.index[id].client.Start([]*rpc.Request{p.healthCheck.RPCRequest})
		go func() {
			defer p.logger.RecoverPanic()

			for {
				select {
				case <-ctx.Done():
					return
				case req := <-reqCh:
					p.mu.Lock()
					p.counts[id]++
					ch, ok := p.reqChs[req.Method]
					topID := p.order[0]
					usable := p.index[id].usable(true)
					p.mu.Unlock()
					if ok && topID == id && usable {
						ch <- req
					}
				case res := <-resCh:
					p.logger.Error(fmt.Errorf("tcppool: recieved unbound response: %s, %s", id, res))
				case err := <-errCh:
					p.errCh <- err
				}
			}
		}()

		err = p.index[id].client.WaitForHandshake(time.Second * 5)
		if err != nil {
			return err
		}
	}

	return nil
}

// Disables a host from being used for requests, though the host is not deleted (and can be enabled again).
func (p *TCPPool) DisableHost(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.index[id]; ok {
		p.index[id].enabled = false
	}
}

// Enables a host, returning it to the active pool to be used in requests.
func (p *TCPPool) EnableHost(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.index[id]; ok {
		p.index[id].enabled = true
	}
}

// Sets the sync status of a given host.
func (p *TCPPool) SetHostSyncStatus(id string, synced bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if _, ok := p.index[id]; ok {
		p.index[id].markSynced(synced)
	}
}

func (p *TCPPool) Subscribe(method string) chan *rpc.Request {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.reqChs[method]; !ok {
		p.reqChs[method] = make(chan *rpc.Request)
	}

	return p.reqChs[method]
}

func (p *TCPPool) exec(req *rpc.Request, needsSynced bool) (*rpc.Response, error) {
	var res *rpc.Response
	var err error
	var failed bool
	for {
		tc := p.getConn(req.HostID, needsSynced)
		if tc == nil {
			failed = true
			break
		}

		res, err = tc.client.WriteRequestWithTimeout(req, tcpTimeout)
		if err != nil {
			tc.markHealthy(false)
			continue
		}

		break
	}

	if failed && err == nil {
		err = ErrNoHealthyHosts
	}

	return res, err
}

func (p *TCPPool) Exec(req *rpc.Request) (*rpc.Response, error) {
	return p.exec(req, false)
}

func (p *TCPPool) ExecSynced(req *rpc.Request) (*rpc.Response, error) {
	return p.exec(req, true)
}

// pop the fastest healthy connection
func (p *TCPPool) getConn(hostID string, needsSynced bool) *tcpConn {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if hostID != "" {
		tc, ok := p.index[hostID]
		if ok && tc.healthy() {
			return tc
		}

		return nil
	}

	for _, id := range p.order {
		tc := p.index[id]
		if tc.usable(needsSynced) {
			return tc
		}
	}

	return nil
}

// run a healthcheck to update healthiness and reorder based on latency
func (p *TCPPool) runHealthCheck() {
	p.mu.RLock()
	healthCheckIdx := make(map[string]HealthCheckFunc)
	for id, tc := range p.index {
		healthCheckIdx[id] = func() int {
			return tc.healthCheck(p.healthCheck, p.logger)
		}
	}
	p.mu.RUnlock()

	if len(healthCheckIdx) == 0 {
		return
	}

	// find the current best connection and the latency of all connections
	latencies := runHealthcheck(healthCheckIdx, p.logger)
	order := processHealthCheck(p.order[0], latencies, p.counts)

	p.mu.Lock()
	p.order = order
	p.counts = make(map[string]int)
	p.latencyIdx = latencies
	p.mu.Unlock()
}

func (p *TCPPool) HandleInfoRequest(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	hosts := make([]map[string]interface{}, len(p.order))
	for i, id := range p.order {
		tc := p.index[id]
		tc.mu.Lock()
		url, errCount, synced := tc.client.URL(), tc.errors, tc.synced
		tc.mu.Unlock()

		latency := float64(p.latencyIdx[id]) / float64(time.Millisecond)
		latency = math.Round(latency*100) / 100

		hosts[i] = map[string]interface{}{
			"id":      id,
			"url":     url,
			"index":   i,
			"synced":  synced,
			"latency": latency,
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

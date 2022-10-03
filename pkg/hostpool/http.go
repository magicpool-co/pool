package hostpool

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/sshtunnel"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

const (
	httpTimeout    = time.Second * 10
	httpStickiness = 1.2 // 20% stickiness
	onceHostID     = "once"
)

var (
	ErrNoHealthyHosts = fmt.Errorf("no healthy hosts")
)

// HTTPError is a convenient custom error type for HTTP errors.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
	Data       []byte
}

// String representation of an HTTPError.
func (err HTTPError) Error() string {
	if len(err.Body) == 0 {
		return err.Status
	}
	return fmt.Sprintf("%v: %s: %s", err.Status, err.Body, err.Data)
}

// HTTPPool represents a pool of HTTP hosts with methods to make standard HTTP calls.
type HTTPPool struct {
	ctx         context.Context
	mu          sync.RWMutex
	index       map[string]*httpConn
	order       []string
	healthCheck *HTTPHealthCheck
	tunnel      *sshtunnel.SSHTunnel
	logger      *log.Logger
}

// HTTPHealthCheck specifies the definition of connection health for all
// connections in the pool. Interval is the frequency that the health check runs,
// Timeout is the connection timeout.
//
// Either the HTTP or RPC fields are required, if both are defined then HTTP
// will take precedence.
type HTTPHealthCheck struct {
	Interval   time.Duration
	Timeout    time.Duration
	HTTPMethod string
	HTTPPath   string
	HTTPBody   interface{}
	RPCRequest *rpc.Request
}

// NewHTTPPool creates an HTTP pool that manages connection health and optimizes
// for latency (while maintaining a degree of "stickiness" to avoid excessive reordering).
//
// The HTTPHealthCheck is not required, but without it the pool has little purpose. The
// health check interval defaults to one minute, the timeout defaults to
func NewHTTPPool(ctx context.Context, logger *log.Logger, healthCheck *HTTPHealthCheck, tunnel *sshtunnel.SSHTunnel) *HTTPPool {
	if healthCheck.Interval == 0 {
		healthCheck.Interval = time.Minute
	}
	if healthCheck.Timeout == 0 {
		healthCheck.Timeout = time.Second * 3
	}

	pool := &HTTPPool{
		ctx:         ctx,
		index:       make(map[string]*httpConn),
		order:       make([]string, 0),
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

type HTTPHostOptions struct {
	Username string
	Password string
	Headers  map[string]string
}

func (p *HTTPPool) GetAllHosts() []string {
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
func (p *HTTPPool) AddHost(url string, port int, opt *HTTPHostOptions) error {
	finalURL, id, err := parseURL(url, port, p.tunnel)
	if err != nil {
		return err
	}

	connHeaders := make(http.Header, 2)
	connHeaders.Set("Accept", "application/json")
	connHeaders.Set("Content-Type", "application/json")
	if opt != nil {
		// set the headers
		for k, v := range opt.Headers {
			connHeaders.Set(k, v)
		}

		// add basic auth if required
		if len(opt.Username) > 0 && len(opt.Password) > 0 {
			auth := opt.Username + ":" + opt.Password
			basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))
			connHeaders.Set("Authorization", "Basic "+basicAuth)
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// add the host to the end of the list, mark as healthy to start
	// to avoid having no healthy hosts until the first healthcheck
	if _, ok := p.index[id]; !ok {
		p.order = append(p.order, id)
		p.index[id] = &httpConn{
			id:      id,
			enabled: true,
			client:  new(http.Client),
			headers: connHeaders,
			url:     finalURL,
		}
	}

	return nil
}

// Disables a host from being used for requests, though the host is not deleted (and can be enabled again).
func (p *HTTPPool) DisableHost(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.index[id]; ok {
		p.index[id].enabled = false
	}
}

// Enables a host, returning it to the active pool to be used in requests.
func (p *HTTPPool) EnableHost(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.index[id]; ok {
		p.index[id].enabled = true
	}
}

// Executes a HTTP call to a specific host. If the host is not healthy, ErrNoHealthyHosts is returned.
// If the host is healthy, the error is returned.
func (p *HTTPPool) ExecHTTPSticky(hostID, method, path string, body, target interface{}) (string, error) {
	// iterate through all host connections until no healthy connections
	// are left or a valid response is returned
	var res []byte
	var err error
	var failed bool
	var count int
	for {
		count++
		hc := p.getConn(hostID, count)
		if hc == nil {
			failed = true
			break
		}

		// enforce a request timeout
		ctx, cancelFunc := context.WithTimeout(context.Background(), httpTimeout)
		defer cancelFunc()

		res, hostID, err = hc.execHTTP(ctx, method, path, body)
		if err != nil {
			p.logger.Error(fmt.Errorf("httppool: http: %v", err))
			continue
		}

		err = json.Unmarshal(res, target)
		if err != nil {
			p.logger.Error(fmt.Errorf("httppool: json: %s: %v: %s", hostID, err, res))
			continue
		}

		break
	}

	if failed && err == nil {
		err = ErrNoHealthyHosts
	}

	return hostID, err
}

// Executes a HTTP call. All healthy hosts will be attempted, if there are no healthy
// hosts to begin with, ErrNoHealthyHosts is returned. If there are healthy hosts, though
// all are actively unhealthy, the last error is returned.
//
// Note that user error is a possibility for hosts being marked as unhealthy if the
// json decoder fails. Target should be a pointer to the response object.
func (p *HTTPPool) ExecHTTP(method, path string, body, target interface{}) error {
	_, err := p.ExecHTTPSticky("", method, path, body, target)
	return err
}

// Executes a HTTP call the same way as ExecHTTP, except it will only attempt
// the request once instead of rotating through all hosts.
func (p *HTTPPool) ExecHTTPOnce(method, path string, body, target interface{}) error {
	_, err := p.ExecHTTPSticky(onceHostID, method, path, body, target)
	return err
}

// Executes an RPC call to all healthy hosts, unless req.HostID is set, in which case it
// will only try the defined host. It is just a convenient wrapper for
// ExecHTTP that reduces the verbosity of RPC calls.
func (p *HTTPPool) ExecRPC(req *rpc.Request) (*rpc.Response, error) {
	if len(req.JSONRPC) == 0 {
		req.JSONRPC = "2.0"
	}

	var res rpc.Response
	hostID, err := p.ExecHTTPSticky(req.HostID, "POST", "", req, &res)
	if err != nil {
		return nil, err
	} else if res.Error != nil {
		return nil, HTTPError{
			Status:     res.Error.Message,
			StatusCode: res.Error.Code,
			Body:       []byte(res.Error.Data),
		}
	}

	res.HostID = hostID

	return &res, err
}

// Executes an RPC call and generates the *rpc.Request internally, requiring
// only the method and the parameters.
func (p *HTTPPool) ExecRPCFromArgs(method string, params ...interface{}) (*rpc.Response, error) {
	req, err := rpc.NewRequest(method, params...)
	if err != nil {
		return nil, err
	}

	return p.ExecRPC(req)
}

// Executes an RPC call the same way as ExecRPCFromArgs, except it will only attempt
// the request once instead of rotating through all hosts.
func (p *HTTPPool) ExecRPCFromArgsOnce(method string, params ...interface{}) (*rpc.Response, error) {
	req, err := rpc.NewRequest(method, params...)
	if err != nil {
		return nil, err
	}
	req.HostID = onceHostID

	return p.ExecRPC(req)
}

// Executes a bulk RPC call to all healthy hosts.
func (p *HTTPPool) ExecRPCBulk(reqs []*rpc.Request) ([]*rpc.Response, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	responses := make([]*rpc.Response, 0)
	err := p.ExecHTTP("POST", "", reqs, &responses)
	if err != nil {
		return nil, err
	}

	for _, res := range responses {
		if res.Error != nil {
			return nil, HTTPError{
				Status:     res.Error.Message,
				StatusCode: res.Error.Code,
				Body:       []byte(res.Error.Data),
			}
		}
	}

	return responses, nil
}

// pop the fastest healthy connection
func (p *HTTPPool) getConn(hostID string, count int) *httpConn {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if hostID != "" && hostID != onceHostID {
		hc, ok := p.index[hostID]
		if ok && hc.healthy() {
			return hc
		}

		return nil
	} else if hostID == onceHostID && count > 1 {
		return nil
	}

	for _, id := range p.order {
		hc := p.index[id]
		if hc.usable() {
			return hc
		}
	}

	return nil
}

// run a healthcheck to update healthiness and reorder based on latency
func (p *HTTPPool) runHealthCheck() {
	p.mu.RLock()
	if len(p.order) == 0 {
		p.mu.RUnlock()
		return
	}

	// find the current best connection and the latency of all connections
	var latencyWg sync.WaitGroup
	var latencyMu sync.Mutex
	latencies := make(map[string]int, len(p.index))
	for id, hc := range p.index {
		latencyWg.Add(1)
		go func(id string, hc *httpConn) {
			defer p.logger.RecoverPanic()
			defer latencyWg.Done()

			latency := hc.healthCheck(p.healthCheck, p.logger)
			latencyMu.Lock()
			latencies[id] = latency
			latencyMu.Unlock()
		}(id, hc)
	}

	p.mu.RUnlock()
	latencyWg.Wait()

	p.mu.Lock()
	p.order = processHealthCheck(p.order[0], latencies, nil)
	p.mu.Unlock()
}

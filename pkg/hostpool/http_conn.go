package hostpool

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/log"
)

// An internal representation of an individual http connection.
type httpConn struct {
	id      string
	mu      sync.RWMutex
	errors  uint
	enabled bool

	client  *http.Client
	headers http.Header
	url     string
}

func (hc *httpConn) healthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	return hc.errors < 3
}

func (hc *httpConn) usable() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	return hc.enabled && hc.errors < 3
}

// Executes the healthcheck and measures the latency for the connection.
// The host's healthiness is automatically updated according to the outcome.
func (hc *httpConn) healthCheck(healthCheck *HTTPHealthCheck, logger *log.Logger) int {
	var unit = time.Nanosecond
	var maxLatency = int(healthCheck.Timeout/unit) * 2

	start := time.Now()
	ctx, cancelFunc := context.WithTimeout(context.Background(), healthCheck.Timeout)
	defer cancelFunc()

	var err error
	if len(healthCheck.HTTPMethod) > 0 {
		_, _, err = hc.execHTTP(ctx, healthCheck.HTTPMethod, healthCheck.HTTPPath, healthCheck.HTTPBody)
	} else if healthCheck.RPCRequest != nil {
		_, _, err = hc.execHTTP(ctx, "POST", "", healthCheck.RPCRequest)
	} else {
		return maxLatency
	}

	if err != nil {
		logger.Error(fmt.Errorf("httpconn: healthcheck: %s: %v", hc.id, err))
		return maxLatency
	}

	hc.markHealthy(true)
	latency := int(time.Since(start) / unit)

	return latency
}

// Change a host's healthiness.
func (hc *httpConn) markHealthy(healthy bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if healthy {
		hc.errors = 0
	} else {
		hc.errors++
	}
}

// Base call to execute an HTTP call. If the request succeeeds, but the status code
// is non-2xx and not 300, a HTTPError is returned.
func (hc *httpConn) execHTTP(ctx context.Context, method, path string, msg interface{}) ([]byte, string, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequestWithContext(ctx, method, hc.url+path, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, "", err
	}
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }

	hc.mu.Lock()
	req.Header = hc.headers.Clone()
	hc.mu.Unlock()

	res, err := hc.client.Do(req)
	if err != nil {
		hc.markHealthy(false)

		return nil, "", err
	} else if res.StatusCode < 200 || res.StatusCode >= 300 {
		hc.markHealthy(false)

		var buf bytes.Buffer
		var body []byte
		if _, err := buf.ReadFrom(res.Body); err == nil {
			body = buf.Bytes()
		}

		return nil, "", HTTPError{
			Status:     res.Status,
			StatusCode: res.StatusCode,
			Body:       body,
		}
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)

	return data, hc.id, err
}

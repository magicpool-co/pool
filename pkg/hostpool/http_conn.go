package hostpool

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/goccy/go-json"
)

// An internal representation of an individual http connection.
type httpConn struct {
	id      string
	mu      sync.RWMutex
	healthy bool
	enabled bool

	client  *http.Client
	headers http.Header
	url     string
}

// Executes the healthcheck and measures the latency for the connection.
// The host's healthiness is automatically updated according to the outcome.
func (hc *httpConn) healthCheck(healthCheck *HTTPHealthCheck) int {
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
		return maxLatency
	}

	hc.markHealthy(true)
	latency := int(time.Since(start) / unit)

	return latency
}

// Change a host's healthiness.
func (hc *httpConn) markHealthy(healthy bool) {
	hc.mu.Lock()
	hc.healthy = healthy
	hc.mu.Unlock()
}

// Base call to execute an HTTP call. If the request succeeeds, but the status code
// is non-2xx and not 300, a HTTPError is returned.
func (hc *httpConn) execHTTP(ctx context.Context, method, path string, msg interface{}) (io.ReadCloser, string, error) {
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
	return res.Body, hc.id, nil
}

package metrics

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Client struct {
	mux        *http.ServeMux
	server     *http.Server
	histograms map[string]*prometheus.HistogramVec
	summaries  map[string]*prometheus.SummaryVec
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
}

func InitClient(port int, enableProfiling bool) *Client {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	if enableProfiling {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	client := &Client{
		mux:        mux,
		server:     server,
		histograms: make(map[string]*prometheus.HistogramVec),
		summaries:  make(map[string]*prometheus.SummaryVec),
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
	}

	return client
}

func (c *Client) Server() *http.Server {
	return c.server
}

func (c *Client) AddHandler(path string, handler http.HandlerFunc) {
	c.mux.HandleFunc(path, handler)
}

func (c *Client) NewHistogram(namespace, name, env, help string, labels ...string) error {
	histogram := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      name,
		Help:      help,
		Buckets: []float64{
			.01, .1, .5, 1, 2.5, 5, 10,
			15, 25, 50, 100, 500, 1000,
		},
		ConstLabels: prometheus.Labels{"env": env},
	}, labels)

	c.histograms[name] = histogram

	return nil
}

func (c *Client) ObserveHistogram(name string, value float64, labels ...string) {
	if histogram, ok := c.histograms[name]; ok {
		histogram.WithLabelValues(labels...).Observe(value)
	}
}

func (c *Client) NewSummary(namespace, name, env, help string, labels ...string) error {
	summary := promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   namespace,
		Name:        name,
		Help:        help,
		ConstLabels: prometheus.Labels{"env": env},
	}, labels)

	c.summaries[name] = summary

	return nil
}

func (c *Client) ObserveSummary(name string, value float64, labels ...string) {
	if summary, ok := c.summaries[name]; ok {
		summary.WithLabelValues(labels...).Observe(value)
	}
}

func (c *Client) NewCounter(namespace, name, env, help string, labels ...string) error {
	counter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   namespace,
		Name:        name,
		Help:        help,
		ConstLabels: prometheus.Labels{"env": env},
	}, labels)

	c.counters[name] = counter

	return nil
}

func (c *Client) IncrementCounter(name string, labels ...string) {
	if counter, ok := c.counters[name]; ok {
		counter.WithLabelValues(labels...).Inc()
	}
}

func (c *Client) AddCounter(name string, value float64, labels ...string) {
	if counter, ok := c.counters[name]; ok {
		counter.WithLabelValues(labels...).Add(value)
	}
}

func (c *Client) NewGauge(namespace, name, env, help string, labels ...string) error {
	gauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        name,
		Help:        help,
		ConstLabels: prometheus.Labels{"env": env},
	}, labels)

	c.gauges[name] = gauge

	return nil
}

func (c *Client) SetGauge(name string, value float64, labels ...string) {
	if gauge, ok := c.gauges[name]; ok {
		gauge.WithLabelValues(labels...).Set(value)
	}
}

func (c *Client) IncrementGauge(name string, labels ...string) {
	if gauge, ok := c.gauges[name]; ok {
		gauge.WithLabelValues(labels...).Inc()
	}
}

func (c *Client) DecrementGauge(name string, labels ...string) {
	if gauge, ok := c.gauges[name]; ok {
		gauge.WithLabelValues(labels...).Dec()
	}
}

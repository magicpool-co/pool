package main

import (
	"github.com/magicpool-co/pool/internal/metrics"
)

func initMetrics(env string, port int) (*metrics.Client, error) {
	metricsClient := metrics.InitClient(port, true)
	err := metricsClient.NewGauge("pool", "clients_active", env,
		"The number of active TCP clients", "chain")
	if err != nil {
		return nil, err
	}

	err = metricsClient.NewCounter("pool", "client_connects", env,
		"The number TCP client connection events", "chain")
	if err != nil {
		return nil, err
	}

	err = metricsClient.NewCounter("pool", "client_disconnects", env,
		"The number TCP client disconnection events", "chain")
	if err != nil {
		return nil, err
	}

	err = metricsClient.NewHistogram("pool", "request_duration_ms", env,
		"The duration of TCP client requests in millseconds", "chain", "handler")
	if err != nil {
		return nil, err
	}

	err = metricsClient.NewCounter("pool", "requests_total", env,
		"The number of TCP client requests", "chain", "handler")
	if err != nil {
		return nil, err
	}

	err = metricsClient.NewCounter("pool", "broadcasts_total", env,
		"The number of TCP broadcast requests", "chain")
	if err != nil {
		return nil, err
	}

	err = metricsClient.NewGauge("pool", "share_difficuly", env,
		"The active share difficulty", "chain")
	if err != nil {
		return nil, err
	}

	err = metricsClient.NewCounter("pool", "accepted_shares_total", env,
		"The number of accepted shares", "chain")
	if err != nil {
		return nil, err
	}

	err = metricsClient.NewCounter("pool", "rejected_shares_total", env,
		"The number of rejected shares", "chain")
	if err != nil {
		return nil, err
	}

	err = metricsClient.NewCounter("pool", "invalid_shares_total", env,
		"The number of invalid shares", "chain")
	if err != nil {
		return nil, err
	}

	return metricsClient, nil
}

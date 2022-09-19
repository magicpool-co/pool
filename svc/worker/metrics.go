package main

import (
	"github.com/magicpool-co/pool/internal/metrics"
)

func initMetrics(env string, port int) (*metrics.Client, error) {
	metricsClient := metrics.InitClient(port, true)
	err := metricsClient.NewGauge("worker", "node_height", env,
		"The last height for a given host", "node_url", "node_chain", "node_region")
	if err != nil {
		return nil, err
	}

	return metricsClient, nil
}

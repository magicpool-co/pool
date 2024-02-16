package hostpool

import (
	"sort"
	"sync"

	"github.com/magicpool-co/pool/internal/log"
)

type HealthCheckFunc func() int

func runHealthcheck(
	index map[string]HealthCheckFunc,
	logger *log.Logger,
) map[string]int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	latencies := make(map[string]int, len(index))
	for id, healthCheckFunc := range index {
		wg.Add(1)
		go func(id string, healthCheckFunc HealthCheckFunc) {
			defer logger.RecoverPanic()
			defer wg.Done()

			latency := healthCheckFunc()
			mu.Lock()
			latencies[id] = latency
			mu.Unlock()
		}(id, healthCheckFunc)
	}

	wg.Wait()

	return latencies
}

func processHealthCheck(
	currentBest string,
	latency, counts map[string]int,
) []string {
	if len(latency) == 0 {
		return nil
	}

	// create a reverse map to find all unique latencies
	reverseLatency := make(map[int][]string)
	for k, v := range latency {
		reverseLatency[v] = append(reverseLatency[v], k)
	}

	// sort the unique latencies
	uniqueLatency := make([]int, 0)
	for k := range reverseLatency {
		uniqueLatency = append(uniqueLatency, k)
	}
	sort.Sort(sort.IntSlice(uniqueLatency))

	// check for stickiness (to avoid constant reording of the fastest connection):
	//	- existing connection is (one of) the fastest
	//	- new fastest connection is less than the stickiness constant
	sticky := false
	for _, k := range reverseLatency[uniqueLatency[0]] {
		if k == currentBest {
			sticky = true
			break
		}
	}
	if !sticky {
		if float64(uniqueLatency[0])/float64(latency[currentBest]) < httpStickiness {
			sticky = true
		}
	}

	// add a blacklist for any connection that has zero requests,
	// provided that the counts index has some connections that
	// have at least one request
	blacklist := make([]string, 0)
	var blacklistActive bool
	for _, count := range counts {
		if count > 0 {
			blacklistActive = true
			break
		}
	}

	// create and set the new order, according to stickiness
	// (but making sure it isn't rightfully blacklisted)
	newOrder := make([]string, 0)
	if sticky && (!blacklistActive || counts[currentBest] > 0) {
		newOrder = append(newOrder, currentBest)
	}

	for _, v := range uniqueLatency {
		for _, vi := range reverseLatency[v] {
			if blacklistActive && counts[vi] == 0 {
				blacklist = append(blacklist, vi)
			} else if !sticky || vi != currentBest {
				newOrder = append(newOrder, vi)
			}
		}
	}
	newOrder = append(newOrder, blacklist...)

	return newOrder
}

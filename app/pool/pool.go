package pool

import (
	"context"
	"fmt"
	"math/big"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/metrics"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/types"
)

type Options struct {
	Chain                string
	StratumPort          int
	WindowSize           int
	ExtraNonceSize       int
	JobListSize          int
	JobListAgeLimit      int
	ForceErrorOnResponse bool
	Flush                bool
	PollingPeriod        time.Duration
	Metrics              *metrics.Client
}

type Pool struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	server     *stratum.Server
	wg         sync.WaitGroup

	chain                string
	windowSize           int64
	extraNonce1Size      int
	forceErrorOnResponse bool
	node                 types.MiningNode

	pollingPeriod time.Duration
	jobManager    *JobManager
	counter       uint64
	counterMu     sync.Mutex
	interval      string
	intervalMu    sync.Mutex
	intervalDone  uint32

	reportedMu    sync.Mutex
	reportedIndex map[string]string

	lastShareMu    sync.Mutex
	lastShareIndex map[string]map[string]time.Time

	db       *dbcl.Client
	redis    *redis.Client
	logger   *log.Logger
	telegram *telegram.Client
	metrics  *metrics.Client
}

func New(node types.MiningNode, dbClient *dbcl.Client, redisClient *redis.Client, logger *log.Logger, telegramClient *telegram.Client, metricsClient *metrics.Client, opt *Options) (*Pool, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	server, err := stratum.NewServer(ctx, opt.StratumPort)
	if err != nil {
		return nil, err
	}

	pool := &Pool{
		ctx:        ctx,
		cancelFunc: cancelFunc,
		server:     server,

		chain:                opt.Chain,
		windowSize:           int64(opt.WindowSize),
		extraNonce1Size:      opt.ExtraNonceSize,
		forceErrorOnResponse: opt.ForceErrorOnResponse,
		node:                 node,

		pollingPeriod: opt.PollingPeriod,
		jobManager:    newJobManager(ctx, node, logger, opt.JobListSize, opt.JobListAgeLimit),

		reportedIndex:  make(map[string]string),
		lastShareIndex: make(map[string]map[string]time.Time),

		db:       dbClient,
		redis:    redisClient,
		logger:   logger,
		telegram: telegramClient,
		metrics:  metricsClient,
	}

	return pool, nil
}

func (p *Pool) recoverPanic() {
	if r := recover(); r != nil {
		p.logger.Panic(r, string(debug.Stack()))
	}
}

func (p *Pool) getCurrentInterval(reset bool) string {
	normalizedDate := common.NormalizeDate(time.Now().UTC(), time.Minute*15, false)
	interval := strconv.FormatInt(normalizedDate.Unix(), 10)
	if interval != p.interval {
		if atomic.LoadUint32(&p.intervalDone) != 1 {
			p.intervalMu.Lock()
			defer p.intervalMu.Unlock()

			if err := p.redis.AddInterval(p.chain, interval); err != nil {
				p.logger.Error(err)
			} else {
				p.interval = interval
				atomic.StoreUint32(&p.intervalDone, 1)
			}
		}
	}

	if reset {
		atomic.StoreUint32(&p.intervalDone, 0)
	}

	return interval
}

func (p *Pool) startJobNotify() {
	defer p.recoverPanic()

	jobCh := make(chan *types.StratumJob)
	errCh := make(chan error)

	p.node.JobNotify(p.ctx, p.pollingPeriod, jobCh, errCh)
	for {
		select {
		case <-p.ctx.Done():
			return
		case job := <-jobCh:
			p.jobManager.update(job)
		case err := <-errCh:
			p.logger.Error(err)
		}
	}
}

func (p *Pool) startReportedHashratePusher() {
	defer p.recoverPanic()

	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			interval := p.getCurrentInterval(true)

			// copy and replace reported index
			p.reportedMu.Lock()
			reportedRaw := p.reportedIndex
			p.reportedIndex = make(map[string]string)
			p.reportedMu.Unlock()

			// process raw reported index into float index
			reported := make(map[string]float64)
			for id, rawHashrate := range reportedRaw {
				hashrate, err := common.HexToBig(rawHashrate)
				if err == nil {
					reported[id], _ = new(big.Float).SetInt(hashrate).Float64()
				}
			}

			// process set miner reported in bulk
			err := p.redis.SetIntervalReportedHashrateBatch(p.chain, interval, reported)
			if err != nil {
				p.logger.Error(err)
			}
		}
	}
}

func (p *Pool) startIPAddressPusher() {
	defer p.recoverPanic()

	ticker := time.NewTicker(time.Minute * 5)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			// copy and replace last share index
			p.lastShareMu.Lock()
			lastShareIndex := p.lastShareIndex
			p.lastShareIndex = make(map[string]map[string]time.Time)
			p.lastShareMu.Unlock()

			// process the index into a slice of addresses
			addresses := make([]*pooldb.IPAddress, 0)
			for compoundID, index := range lastShareIndex {
				parts := strings.Split(compoundID, ":")
				if len(parts) != 2 {
					p.logger.Error(fmt.Errorf("invalid compoundID: %s", compoundID))
					continue
				}

				minerID, err := strconv.ParseUint(parts[0], 10, 64)
				if err != nil || minerID == 0 {
					p.logger.Error(fmt.Errorf("invalid compoundID: %s", compoundID))
					continue
				}

				workerID, err := strconv.ParseUint(parts[1], 10, 64)
				if err != nil || workerID == 0 {
					p.logger.Error(fmt.Errorf("invalid compoundID: %s", compoundID))
					continue
				}

				for ip, lastShare := range index {
					address := &pooldb.IPAddress{
						MinerID:   minerID,
						WorkerID:  workerID,
						ChainID:   p.chain,
						IPAddress: ip,

						Active:    true,
						LastShare: lastShare,
					}
					addresses = append(addresses, address)
				}
			}

			// insert the ip addresses and set old addresses to inactive
			// @TODO: UpdateIPAddressesSetInactive should be run on the worker, where it will run once
			if err := pooldb.InsertIPAddresses(p.db.Writer(), addresses...); err != nil {
				p.logger.Error(err)
			} else if err := pooldb.UpdateIPAddressesSetInactive(p.db.Writer(), time.Hour); err != nil {
				p.logger.Error(err)
			}
		}
	}
}

func (p *Pool) startStratum() {
	defer p.recoverPanic()

	msgCh, connectCh, disconnectCh, errCh, err := p.server.Start(time.Minute)
	if err != nil {
		p.logger.Error(err)
		return
	}

	for {
		select {
		case <-p.ctx.Done():
			return
		case connID := <-connectCh:
			p.logger.Debug(fmt.Sprintf("conn %d connected", connID))
			if p.metrics != nil {
				p.metrics.IncrementGauge("clients_active", p.chain)
				p.metrics.IncrementCounter("client_connects", p.chain)
			}
		case connID := <-disconnectCh:
			go p.jobManager.RemoveConn(connID)
			p.logger.Debug(fmt.Sprintf("conn %d disconnected", connID))
			if p.metrics != nil {
				p.metrics.DecrementGauge("clients_active", p.chain)
				p.metrics.IncrementCounter("client_disconnects", p.chain)
			}
		case msg := <-msgCh:
			handler := p.routeRequest(msg.Req)
			if handler == nil {
				continue
			}

			go func() {
				defer p.recoverPanic()

				if p.metrics != nil {
					startTime := time.Now()
					defer func() {
						requestTime := float64(time.Since(startTime) / time.Millisecond)
						p.metrics.ObserveHistogram("request_duration_ms", requestTime, p.chain, msg.Req.Method)
						p.metrics.IncrementCounter("requests_total", p.chain, msg.Req.Method)
					}()
				}

				err := handler(msg.Conn, msg.Req)
				if err != nil {
					p.logger.Error(err)
				}
			}()
		case err := <-errCh:
			p.logger.Error(err)
		}
	}
}

func (p *Pool) Port() int {
	return p.server.Port()
}

func (p *Pool) Serve() {
	go p.startJobNotify()
	go p.startReportedHashratePusher()
	go p.startIPAddressPusher()
	go p.startStratum()

	if p.metrics != nil {
		p.metrics.SetGauge("share_difficuly", p.node.GetAdjustedShareDifficulty(), p.chain)
	}
}

func (p *Pool) Stop() {
	p.cancelFunc()
	p.wg.Wait()
}

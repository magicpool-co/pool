package pool

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/core/stream"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/metrics"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/internal/telegram"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/types"
)

type Options struct {
	Chain                string
	PortDiffIdx          map[int]int
	WindowSize           int
	ExtraNonceSize       int
	JobListSize          int
	JobListAgeLimit      int
	SoloEnabled          bool
	VarDiffEnabled       bool
	StreamEnabled        bool
	ForceErrorOnResponse bool
	Flush                bool
	PollingPeriod        time.Duration
	PingingPeriod        time.Duration
	Metrics              *metrics.Client
}

type Pool struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	server     *stratum.Server
	wg         sync.WaitGroup

	chain                string
	soloChain            string
	portDiffIdx          map[int]int
	windowSize           int64
	extraNonce1Size      int
	soloEnabled          bool
	varDiffEnabled       bool
	forceErrorOnResponse bool
	node                 types.MiningNode
	streamWriter         *stream.Writer

	pollingPeriod time.Duration
	pingingPeriod time.Duration

	jobManager   *JobManager
	counter      uint64
	counterMu    sync.Mutex
	interval     string
	intervalMu   sync.Mutex
	intervalDone uint32

	minerStatsMu      sync.Mutex
	lastShareIndex    map[string]int64
	lastDiffIndex     map[string]int64
	latencyValueIndex map[string]int64
	latencyCountIndex map[string]int64

	db       *dbcl.Client
	redis    *redis.Client
	logger   *log.Logger
	telegram *telegram.Client
	metrics  *metrics.Client
}

func New(node types.MiningNode, dbClient *dbcl.Client, redisClient *redis.Client, logger *log.Logger, telegramClient *telegram.Client, metricsClient *metrics.Client, opt *Options) (*Pool, error) {
	ports := make([]int, 0)
	for port := range opt.PortDiffIdx {
		ports = append(ports, port)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	server, err := stratum.NewServer(ctx, logger, opt.VarDiffEnabled, ports...)
	if err != nil {
		return nil, err
	}

	logger.LabelKeys = []string{"miner"}

	var streamWriter *stream.Writer
	if opt.StreamEnabled {
		streamWriter, err = stream.NewWriter(ctx, opt.Chain, "/stream", logger, redisClient)
		if err != nil {
			logger.Fatal(fmt.Errorf("failed to init stream writer: %v", err))
		}
	}

	pool := &Pool{
		ctx:        ctx,
		cancelFunc: cancelFunc,
		server:     server,

		chain:                strings.ToUpper(opt.Chain),
		soloChain:            "S" + strings.ToUpper(opt.Chain),
		portDiffIdx:          opt.PortDiffIdx,
		windowSize:           int64(opt.WindowSize),
		extraNonce1Size:      opt.ExtraNonceSize,
		soloEnabled:          opt.SoloEnabled,
		varDiffEnabled:       opt.VarDiffEnabled,
		forceErrorOnResponse: opt.ForceErrorOnResponse,
		node:                 node,
		streamWriter:         streamWriter,

		pollingPeriod: opt.PollingPeriod,
		pingingPeriod: opt.PingingPeriod,

		jobManager: newJobManager(ctx, node, logger, opt.JobListSize, opt.JobListAgeLimit),

		lastShareIndex:    make(map[string]int64),
		lastDiffIndex:     make(map[string]int64),
		latencyValueIndex: make(map[string]int64),
		latencyCountIndex: make(map[string]int64),

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

func (p *Pool) writeToConn(c *stratum.Conn, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	p.logger.Debug("sending stratum response: " + string(data))

	return c.Write(data)
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

			if p.soloEnabled {
				err := p.redis.AddInterval(p.soloChain, interval)
				if err != nil {
					p.logger.Error(err)
				}
			}
		}
	}

	if reset {
		atomic.StoreUint32(&p.intervalDone, 0)
	}

	return interval
}

func (p *Pool) startPingHosts() {
	if p.pingingPeriod == 0 {
		return
	}

	defer p.recoverPanic()

	ticker := time.NewTicker(p.pingingPeriod)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.node.PingHosts()
		}
	}
}

func (p *Pool) startJobNotify() {
	defer p.recoverPanic()

	timer := time.NewTimer(time.Minute * 10)
	jobCh := p.node.JobNotify(p.ctx, p.pollingPeriod)

	for {
		select {
		case <-p.ctx.Done():
			return
		case job := <-jobCh:
			isNew, err := p.jobManager.update(job)
			if err != nil {
				p.logger.Error(err)
			}

			if isNew {
				timer.Reset(time.Minute * 10)

				err = p.redis.AddShareIndexHeight(p.chain, job.Height.Value())
				if err != nil {
					p.logger.Error(err)
				}

				if p.soloEnabled {
					err = p.redis.AddShareIndexHeight(p.soloChain, job.Height.Value())
					if err != nil {
						p.logger.Error(err)
					}
				}
			}
		case <-timer.C:
			timer.Reset(time.Minute * 10)
			job := p.jobManager.LatestJob()
			if job != nil {
				var hash string
				if job.HeaderHash != nil {
					hash = job.HeaderHash.Hex()
				} else if job.CoinbaseTxID != nil {
					hash = job.CoinbaseTxID.Hex()
				}

				p.logger.Error(fmt.Errorf("have not recieved new job in past 10 minutes: %s %s", hash, job.HostID))
			} else {
				p.logger.Error(fmt.Errorf("have not recieved new job in past 10 minutes"))
			}
		}
	}
}

func (p *Pool) startShareIndexClearer() {
	defer p.recoverPanic()

	ticker := time.NewTicker(time.Minute * 5)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			chains := []string{p.chain}
			if p.soloEnabled {
				chains = append(chains, p.soloChain)
			}

			for _, chain := range chains {
				indexes, err := p.redis.GetShareIndexes(chain)
				if err != nil {
					p.logger.Error(err)
					continue
				}

				for _, index := range indexes {
					height, err := strconv.ParseUint(index, 10, 64)
					if err != nil {
						p.logger.Error(err)
						continue
					} else if !p.jobManager.isExpiredHeight(height) {
						continue
					}

					err = p.redis.DeleteShareIndexHeight(chain, height)
					if err != nil {
						p.logger.Error(err)
					}
				}
			}
		}
	}
}

func (p *Pool) startMinerStatsPusher() {
	defer p.recoverPanic()

	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			// force interval addition
			p.getCurrentInterval(true)

			// copy and replace last share and latency index
			p.minerStatsMu.Lock()

			lastShareIndex := p.lastShareIndex
			p.lastShareIndex = make(map[string]int64)

			lastDiffIndex := p.lastDiffIndex
			p.lastDiffIndex = make(map[string]int64)

			latencyValueIndex, latencyCountIndex := p.latencyValueIndex, p.latencyCountIndex
			p.latencyValueIndex, p.latencyCountIndex = make(map[string]int64), make(map[string]int64)

			p.minerStatsMu.Unlock()

			// process set ip address in bulk
			err := p.redis.SetMinerIPAddressesBulk(p.chain, lastShareIndex)
			if err != nil {
				p.logger.Error(err)
			}

			err = p.redis.SetMinerDifficultiesBulk(p.chain, lastDiffIndex)
			if err != nil {
				p.logger.Error(err)
			}

			for k, value := range latencyValueIndex {
				if value == 0 {
					delete(latencyValueIndex, k)
				} else if count, ok := latencyCountIndex[k]; !ok || count == 0 {
					delete(latencyValueIndex, k)
				}

				latencyValueIndex[k] /= latencyCountIndex[k]
			}

			// process set ip address in bulk
			err = p.redis.SetMinerLatenciesBulk(p.chain, latencyValueIndex)
			if err != nil {
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
		case <-connectCh:
			if p.metrics != nil {
				p.metrics.IncrementGauge("clients_active", p.chain)
				p.metrics.IncrementCounter("client_connects", p.chain)
			}
		case c := <-disconnectCh:
			go p.jobManager.RemoveConn(c.GetID())

			// handle disconnect streaming
			if p.streamWriter != nil && c != nil && c.GetAuthorized() {
				p.streamWriter.WriteDisconnectEvent(c.GetMinerID(), c.GetWorker(),
					c.GetClient(), c.GetPort(), c.GetIsSolo())
			}

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
					if msg.Conn.GetLastErrorAt().Before(time.Now().Add(time.Minute * -5)) {
						msg.Conn.SetLastErrorAt(time.Now())
						msg.Conn.SetErrorCount(1)
					} else if cnt := msg.Conn.GetErrorCount(); cnt < 5 {
						msg.Conn.SetErrorCount(cnt + 1)
					} else {
						msg.Conn.Close()
					}

					p.logger.Error(err, msg.Conn.GetCompoundID())
				}
			}()
		case err := <-errCh:
			p.logger.Error(err)
		}
	}
}

func (p *Pool) Port(idx int) int {
	return p.server.Port(idx)
}

func (p *Pool) Serve() {
	go p.startPingHosts()
	go p.startJobNotify()
	go p.startShareIndexClearer()
	go p.startMinerStatsPusher()
	go p.startStratum()

	if p.metrics != nil {
		p.metrics.SetGauge("share_difficulty", p.node.GetAdjustedShareDifficulty(), p.chain)
		p.metrics.SetGauge("share_difficulty", p.node.GetAdjustedShareDifficulty(), p.soloChain)
	}
}

func (p *Pool) Stop() {
	p.cancelFunc()
	p.wg.Wait()
	p.server.Wait()
}

package pool

import (
	"context"
	"strconv"
	"sync"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/core/stream"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/types"
)

type JobList struct {
	mu       sync.RWMutex
	size     int
	ageLimit int
	counter  uint64
	height   uint64
	order    []string
	index    map[string]*types.StratumJob
	hashes   map[string]string
}

func newJobList(size, ageLimit int) *JobList {
	list := &JobList{
		size:     size,
		ageLimit: ageLimit,
		order:    make([]string, 0),
		index:    make(map[string]*types.StratumJob),
	}

	return list
}

func (l *JobList) Size() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return len(l.order)
}

func (l *JobList) nextCounter() string {
	l.counter++
	if l.counter%0xffffffffff == 0 {
		l.counter = 1
	}

	id := strconv.FormatUint(l.counter, 16)
	for i := 0; i < 10-len(id); i++ {
		id = "0" + id
	}

	return id
}

func (l *JobList) Append(job *types.StratumJob) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if job.ID == "" {
		job.ID = l.nextCounter()
	} else if _, ok := l.index[job.ID]; ok {
		return false
	}

	cleanJobs := job.Height.Value() != l.height

	l.order = append([]string{job.ID}, l.order...)
	l.index[job.ID] = job
	l.height = job.Height.Value()

	if len(l.order) > l.size {
		for _, id := range l.order[l.size:] {
			// in case we're getting rid of old jobs that haven't
			// been overwritten with a "cleanJobs" flag, force the flag
			if oldJob, ok := l.index[id]; ok && !cleanJobs {
				cleanJobs = oldJob.Height.Value() == job.Height.Value()
			}
			delete(l.index, id)
		}
		l.order = l.order[:l.size]
	}

	return cleanJobs
}

func (l *JobList) Get(id string) (*types.StratumJob, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	job := l.index[id]
	var active bool
	if job != nil {
		height := job.Height.Value()
		if height == l.height || l.ageLimit == -1 {
			active = true
		} else if height+uint64(l.ageLimit) >= l.height {
			active = true
		}
	}

	return job, active
}

func (l *JobList) GetPrior(id string) (*types.StratumJob, bool) {
	l.mu.RLock()
	var previous string
	for i := len(l.order) - 1; i >= 0; i-- {
		if id == l.order[i] {
			break
		}
		previous = l.order[i]
	}
	l.mu.RUnlock()

	return l.Get(previous)
}

func (l *JobList) Latest() *types.StratumJob {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.order) == 0 {
		return nil
	}

	return l.index[l.order[0]]
}

func (l *JobList) Oldest() *types.StratumJob {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.order) == 0 {
		return nil
	}

	return l.index[l.order[len(l.order)-1]]
}

type JobManager struct {
	ctx             context.Context
	node            types.MiningNode
	logger          *log.Logger
	streamWriter    *stream.Writer
	subscriptions   map[int]map[int]map[uint64]chan []byte
	subscriptionsMu sync.RWMutex
	jobList         *JobList
}

func newJobManager(
	ctx context.Context,
	node types.MiningNode,
	logger *log.Logger,
	streamWriter *stream.Writer,
	size, ageLimit int,
) *JobManager {
	manager := &JobManager{
		ctx:           ctx,
		node:          node,
		logger:        logger,
		streamWriter:  streamWriter,
		subscriptions: make(map[int]map[int]map[uint64]chan []byte),
		jobList:       newJobList(size, ageLimit),
	}

	return manager
}

func (m *JobManager) update(job *types.StratumJob) (bool, error) {
	if job == nil {
		return false, nil
	}

	cleanJobs := m.jobList.Append(job)

	m.subscriptionsMu.Lock()
	defer m.subscriptionsMu.Unlock()

	for clientType, clientSubscriptionsIdx := range m.subscriptions {
		for diffFactor, clientSubscriptions := range clientSubscriptionsIdx {
			if len(clientSubscriptions) == 0 {
				continue
			}

			msg, err := m.node.MarshalJob(0, job, cleanJobs, clientType, diffFactor)
			if err != nil {
				return cleanJobs, err
			}

			data, err := json.Marshal(msg)
			if err != nil {
				return cleanJobs, err
			}

			// thanks FlexPool :P
			for _, ch := range clientSubscriptions {
				select {
				case <-ch:
				default:
					select {
					case ch <- data:
					default:
					}
				}
			}

			// garbage collect old subscriptions
			for id, ch := range clientSubscriptions {
				select {
				case <-ch:
					delete(clientSubscriptions, id)
				default:
				}
			}
		}
	}

	return cleanJobs, nil
}

func (m *JobManager) isExpiredHeight(height uint64) bool {
	indexDepth := 3
	if m.jobList.ageLimit > 0 {
		indexDepth = m.jobList.ageLimit
	}

	oldest := m.jobList.Oldest()
	if oldest == nil {
		return false
	}

	return int(oldest.Height.Value())-indexDepth > int(height)
}

func (m *JobManager) AddConn(c *stratum.Conn) {
	// runs as goroutine
	defer m.logger.RecoverPanic()

	jobs := make(chan []byte)
	defer func() {
		if _, ok := <-jobs; ok {
			close(jobs)
		}
	}()

	m.subscriptionsMu.Lock()
	clientType := c.GetClientType()
	diffFactor := c.GetDiffFactor()
	if _, ok := m.subscriptions[clientType]; !ok {
		m.subscriptions[clientType] = make(map[int]map[uint64]chan []byte)
	}
	if _, ok := m.subscriptions[clientType][diffFactor]; !ok {
		m.subscriptions[clientType][diffFactor] = make(map[uint64]chan []byte)
	}
	m.subscriptions[clientType][diffFactor][c.GetID()] = jobs
	m.subscriptionsMu.Unlock()

	for {
		select {
		case <-m.ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			err := c.Write(job)
			if err != nil {
				return
			}

			if m.streamWriter != nil {
				m.streamWriter.WriteDebugResponse(c.GetIP(), job)
			}
		}
	}
}

func (m *JobManager) RemoveConn(id uint64) {
	// runs as goroutine
	defer m.logger.RecoverPanic()

	m.subscriptionsMu.RLock()
	defer m.subscriptionsMu.RUnlock()

	for _, clientSubscriptionsIdx := range m.subscriptions {
		for _, clientSubscriptions := range clientSubscriptionsIdx {
			if ch, ok := clientSubscriptions[id]; ok {
				close(ch)
			}
			break
		}
	}
}

func (m *JobManager) GetJob(id string) (*types.StratumJob, bool) {
	return m.jobList.Get(id)
}

func (m *JobManager) GetPriorJob(id string) (*types.StratumJob, bool) {
	return m.jobList.GetPrior(id)
}

func (m *JobManager) LatestJob() *types.StratumJob {
	return m.jobList.Latest()
}

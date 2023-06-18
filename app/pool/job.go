package pool

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/types"
)

type JobList struct {
	mu       sync.Mutex
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
	l.mu.Lock()
	defer l.mu.Unlock()

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

func (l *JobList) Get(id string) (*types.StratumJob, bool, uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()

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

	return job, active, l.height
}

func (l *JobList) Latest() *types.StratumJob {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.order) == 0 {
		return nil
	}

	return l.index[l.order[0]]
}

func (l *JobList) Oldest() *types.StratumJob {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.order) == 0 {
		return nil
	}

	return l.index[l.order[len(l.order)-1]]
}

type JobManager struct {
	ctx             context.Context
	node            types.MiningNode
	logger          *log.Logger
	subscriptions   map[int]map[uint64]chan []byte
	subscriptionsMu sync.RWMutex
	jobList         *JobList
}

func newJobManager(ctx context.Context, node types.MiningNode, logger *log.Logger, size, ageLimit int) *JobManager {
	manager := &JobManager{
		ctx:           ctx,
		node:          node,
		logger:        logger,
		subscriptions: make(map[int]map[uint64]chan []byte),
		jobList:       newJobList(size, ageLimit),
	}

	return manager
}

func (m *JobManager) recoverPanic() {
	if r := recover(); r != nil {
		m.logger.Panic(r, string(debug.Stack()))
	}
}

func (m *JobManager) update(job *types.StratumJob) (bool, error) {
	if job == nil {
		return false, nil
	}

	cleanJobs := m.jobList.Append(job)

	m.subscriptionsMu.Lock()
	defer m.subscriptionsMu.Unlock()

	for clientType, clientSubscriptions := range m.subscriptions {
		if len(clientSubscriptions) == 0 {
			continue
		}

		msg, err := m.node.MarshalJob(0, job, cleanJobs, clientType)
		if err != nil {
			return cleanJobs, err
		}

		data, err := json.Marshal(msg)
		if err != nil {
			return cleanJobs, err
		}

		m.logger.Debug("broadcasting stratum job: " + string(data))

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
	defer m.recoverPanic()

	jobs := make(chan []byte)
	defer func() {
		if _, ok := <-jobs; ok {
			close(jobs)
		}
	}()

	m.subscriptionsMu.Lock()
	clientType := c.GetClientType()
	if _, ok := m.subscriptions[clientType]; !ok {
		m.subscriptions[clientType] = make(map[uint64]chan []byte)
	}
	m.subscriptions[clientType][c.GetID()] = jobs
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
		}
	}
}

func (m *JobManager) RemoveConn(id uint64) {
	defer m.recoverPanic()

	m.subscriptionsMu.RLock()
	defer m.subscriptionsMu.RUnlock()

	for _, clientSubscriptions := range m.subscriptions {
		if ch, ok := clientSubscriptions[id]; ok {
			close(ch)
		}
		break
	}
}

func (m *JobManager) GetJob(id string) (*types.StratumJob, bool) {
	job, active, activeHeight := m.jobList.Get(id)
	if !active && job != nil {
		m.logger.Info(fmt.Sprintf("share not active: %s (%d vs %d)", job.ID, job.Height.Value(), activeHeight))
	}

	return job, active
}

func (m *JobManager) LatestJob() *types.StratumJob {
	return m.jobList.Latest()
}

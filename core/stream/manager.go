package stream

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/maphash"
	"sync/atomic"
	"time"

	"github.com/puzpuzpuz/xsync/v2"

	"github.com/magicpool-co/pool/internal/redis"
)

func hashUint32(seed maphash.Seed, u uint32) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	binary.Write(&h, binary.LittleEndian, u)
	return h.Sum64()
}

type stream struct {
	minerID       uint64
	ctx           context.Context
	cancel        context.CancelFunc
	redis         *redis.Client
	pubsub        *redis.PubSub
	counter       atomic.Uint32
	subscriptions *xsync.MapOf[uint32, chan string]
}

func newStream(minerID uint64, redisClient *redis.Client) (*stream, error) {
	pubsub, err := redisClient.GetStreamMinerChannel(minerID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &stream{
		minerID:       minerID,
		ctx:           ctx,
		cancel:        cancel,
		redis:         redisClient,
		pubsub:        pubsub,
		subscriptions: xsync.NewTypedMapOf[uint32, chan string](hashUint32),
	}

	go s.listenAndBroadcast()

	return s, nil
}

func (s *stream) close() {
	s.cancel()
	s.pubsub.Close()
	s.subscriptions.Range(func(subID uint32, ch chan string) bool {
		close(ch)
		s.subscriptions.Delete(subID)
		return true
	})
}

func (s *stream) listenAndBroadcast() {
	go func() {
		ackMsg := fmt.Sprintf("ack:%d", s.minerID)
		s.redis.WriteToStreamIndexChannel(ackMsg)
		ticker := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.redis.WriteToStreamIndexChannel(ackMsg)
			}
		}
	}()

	pubsubCh := s.pubsub.Channel()
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-pubsubCh:
			s.subscriptions.Range(func(subID uint32, ch chan string) bool {
				select {
				case <-ch:
				default:
					select {
					case ch <- msg.Payload:
					default:
					}
				}
				return true
			})
		}
	}
}

type Manager struct {
	redis   *redis.Client
	streams map[uint64]*stream
	mu      *xsync.RBMutex
}

func NewManager(redisClient *redis.Client) *Manager {
	manager := &Manager{
		redis:   redisClient,
		streams: make(map[uint64]*stream),
		mu:      xsync.NewRBMutex(),
	}

	return manager
}

func (m *Manager) Subscribe(minerID uint64) (uint32, <-chan string, error) {
	t := m.mu.RLock()
	s, ok := m.streams[minerID]
	m.mu.RUnlock(t)
	if !ok {
		var err error
		m.mu.Lock()
		s, err = newStream(minerID, m.redis)
		m.streams[minerID] = s
		m.mu.Unlock()
		if err != nil {
			return 0, nil, err
		}
	}

	subID := s.counter.Add(1)
	ch := make(chan string)
	s.subscriptions.Store(subID, ch)

	return subID, ch, nil
}

func (m *Manager) Unsubscribe(minerID uint64, subID uint32) {
	t := m.mu.RLock()
	s, ok := m.streams[minerID]
	m.mu.RUnlock(t)
	if !ok {
		return
	}

	s.subscriptions.Delete(subID)

	m.mu.Lock()
	defer m.mu.Unlock()
	if s.subscriptions.Size() == 0 {
		s.close()
		delete(m.streams, minerID)
	}
}

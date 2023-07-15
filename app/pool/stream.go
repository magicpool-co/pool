package pool

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/redis"
)

type streamWriter struct {
	chain   string
	path    string
	ctx     context.Context
	logger  *log.Logger
	redis   *redis.Client
	pubsub  *redis.PubSub
	streams map[uint64]time.Time
	mu      sync.RWMutex
}

func newStreamWriter(ctx context.Context, chain, path string, logger *log.Logger, redisClient *redis.Client) (*streamWriter, error) {
	pubsub, err := redisClient.GetStreamIndexChannel()
	if err != nil {
		return nil, err
	}

	writer := &streamWriter{
		chain:   chain,
		path:    path,
		ctx:     ctx,
		logger:  logger,
		redis:   redisClient,
		pubsub:  pubsub,
		streams: make(map[uint64]time.Time),
	}

	go writer.listen()

	return writer, nil
}

func (w *streamWriter) getStream(minerID uint64) bool {
	w.mu.RLock()
	lastAck, ok := w.streams[minerID]
	w.mu.RUnlock()

	if !ok || time.Since(lastAck) < time.Second*15 {
		return true
	}

	w.mu.Lock()
	delete(w.streams, minerID)
	w.mu.Unlock()

	return false
}

func (w *streamWriter) listen() {
	defer w.logger.RecoverPanic()

	ch := w.pubsub.Channel()
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-w.ctx.Done():
			w.pubsub.Close()
			w.mu.Lock()
			w.streams = nil
			w.mu.Unlock()
			return
		case <-ticker.C:
			w.mu.Lock()
			for minerID, lastAck := range w.streams {
				if time.Since(lastAck) > time.Minute {
					delete(w.streams, minerID)
				}
			}
			w.mu.Unlock()
		case msg := <-ch:
			parts := strings.Split(msg.Payload, "|")
			if len(parts) != 2 {
				continue
			}

			switch parts[0] {
			case "ack":
				minerID, err := strconv.ParseUint(parts[1], 10, 64)
				if err != nil {
					w.logger.Error(fmt.Errorf("invalid ack: %s", msg.Payload))
					continue
				}

				w.mu.Lock()
				w.streams[minerID] = time.Now()
				w.mu.Unlock()
			}
		}
	}
}

func (w *streamWriter) write(minerID uint64, chain, eventType, worker, client string, data map[string]interface{}) {
	if !w.getStream(minerID) {
		return
	}

	event := map[string]interface{}{
		"chain":  chain,
		"type":   eventType,
		"worker": worker,
		"client": client,
		"data":   data,
	}

	msg, err := json.Marshal(event)
	if err != nil {
		w.logger.Error(fmt.Errorf("failed marshaling event: %v", err))
		return
	}

	err = w.redis.WriteToStreamMinerChannel(minerID, string(msg))
	if err != nil {
		w.logger.Error(fmt.Errorf("failed writing event: %v", err))
		return
	}
}

func (w *streamWriter) WriteConnectEvent(minerID uint64, worker, client string) {
	w.write(minerID, w.chain, "connect", worker, client, nil)
}

func (w *streamWriter) WriteDisconnectEvent(minerID uint64, worker, client string) {
	w.write(minerID, w.chain, "disconnect", worker, client, nil)
}

func (w *streamWriter) WriteShareEvent(minerID uint64, worker, client, status string, shareDiff, targetDiff uint64) {
	data := map[string]interface{}{"status": status, "share_diff": shareDiff, "target_diff": targetDiff}
	w.write(minerID, w.chain, "share", worker, client, data)
}

func (w *streamWriter) WriteRetargetEvent(minerID uint64, worker, client string, oldDiff, newDiff uint64) {
	data := map[string]interface{}{"old_diff": oldDiff, "new_diff": newDiff}
	w.write(minerID, w.chain, "retarget", worker, client, data)
}

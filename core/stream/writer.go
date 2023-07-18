package stream

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

type Writer struct {
	chain   string
	path    string
	ctx     context.Context
	logger  *log.Logger
	redis   *redis.Client
	pubsub  *redis.PubSub
	streams map[uint64]time.Time
	mu      sync.RWMutex
}

func NewWriter(ctx context.Context, chain, path string, logger *log.Logger, redisClient *redis.Client) (*Writer, error) {
	pubsub, err := redisClient.GetStreamIndexChannel()
	if err != nil {
		return nil, err
	}

	writer := &Writer{
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

func (w *Writer) listen() {
	defer w.logger.RecoverPanic()

	ch := w.pubsub.Channel()
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-w.ctx.Done():
			w.pubsub.Close()
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

func (w *Writer) getStream(minerID uint64) bool {
	w.mu.RLock()
	lastAck, ok := w.streams[minerID]
	w.mu.RUnlock()

	if !ok || time.Since(lastAck) < time.Second*15 {
		return ok
	}

	w.mu.Lock()
	delete(w.streams, minerID)
	w.mu.Unlock()

	return false
}

func (w *Writer) write(minerID uint64, chain, eventType, worker, client string, port int, solo bool, data map[string]interface{}) {
	if !w.getStream(minerID) {
		return
	}

	event := map[string]interface{}{
		"chain":     chain,
		"type":      eventType,
		"port":      port,
		"solo":      solo,
		"worker":    worker,
		"client":    client,
		"data":      data,
		"timestamp": time.Now().Unix(),
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

func (w *Writer) WriteConnectEvent(minerID uint64, worker, client string, port int, solo bool) {
	w.write(minerID, w.chain, "connect", worker, client, port, solo, nil)
}

func (w *Writer) WriteDisconnectEvent(minerID uint64, worker, client string, port int, solo bool) {
	w.write(minerID, w.chain, "disconnect", worker, client, port, solo, nil)
}

func (w *Writer) WriteErrorEvent(minerID uint64, worker, client string, port int, solo bool, error string) {
	data := map[string]interface{}{"error": error}
	w.write(minerID, w.chain, "error", worker, client, port, solo, data)
}

func (w *Writer) WriteShareEvent(minerID uint64, worker, client string, port int, solo bool, status, reason string, shareDiff, targetDiff uint64) {
	data := map[string]interface{}{"status": status, "share_diff": shareDiff, "target_diff": targetDiff}
	if len(reason) > 0 {
		data["reason"] = reason
	}
	w.write(minerID, w.chain, "share", worker, client, port, solo, data)
}

func (w *Writer) WriteRetargetEvent(minerID uint64, worker, client string, port int, solo bool, oldDiff, newDiff uint64) {
	data := map[string]interface{}{"old_diff": oldDiff, "new_diff": newDiff}
	w.write(minerID, w.chain, "retarget", worker, client, port, solo, data)
}

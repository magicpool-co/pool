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
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

type Writer struct {
	chain        string
	path         string
	ctx          context.Context
	logger       *log.Logger
	redis        *redis.Client
	eventPubsub  *redis.PubSub
	debugPubsub  *redis.PubSub
	eventStreams map[uint64]time.Time
	debugStreams map[string]time.Time
	mu           sync.RWMutex
}

func NewWriter(ctx context.Context, chain, path string, logger *log.Logger, redisClient *redis.Client) (*Writer, error) {
	eventPubsub, err := redisClient.GetStreamMinerIndexChannel()
	if err != nil {
		return nil, err
	}

	debugPubsub, err := redisClient.GetStreamDebugIndexChannel()
	if err != nil {
		return nil, err
	}

	writer := &Writer{
		chain:        chain,
		path:         path,
		ctx:          ctx,
		logger:       logger,
		redis:        redisClient,
		eventPubsub:  eventPubsub,
		debugPubsub:  debugPubsub,
		eventStreams: make(map[uint64]time.Time),
		debugStreams: make(map[string]time.Time),
	}

	go writer.listen()

	return writer, nil
}

func (w *Writer) listen() {
	defer w.logger.RecoverPanic()

	eventCh := w.eventPubsub.Channel()
	debugCh := w.eventPubsub.Channel()
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-w.ctx.Done():
			w.eventPubsub.Close()
			w.debugPubsub.Close()
			return
		case <-ticker.C:
			w.mu.Lock()
			for minerID, lastAck := range w.eventStreams {
				if time.Since(lastAck) > time.Minute {
					delete(w.eventStreams, minerID)
				}
			}

			for ip, lastAck := range w.debugStreams {
				if time.Since(lastAck) > time.Minute {
					delete(w.debugStreams, ip)
				}
			}
			w.mu.Unlock()
		case msg := <-eventCh:
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
				w.eventStreams[minerID] = time.Now()
				w.mu.Unlock()
			}
		case msg := <-debugCh:
			parts := strings.Split(msg.Payload, "|")
			if len(parts) != 2 {
				continue
			}

			switch parts[0] {
			case "ack":
				ip := parts[1]

				w.mu.Lock()
				w.debugStreams[ip] = time.Now()
				w.mu.Unlock()
			}
		}
	}
}

func (w *Writer) getEventStream(minerID uint64) bool {
	w.mu.RLock()
	lastAck, ok := w.eventStreams[minerID]
	w.mu.RUnlock()

	if !ok || time.Since(lastAck) < time.Second*15 {
		return ok
	}

	w.mu.Lock()
	delete(w.eventStreams, minerID)
	w.mu.Unlock()

	return false
}

func (w *Writer) getDebugStream(ip string) bool {
	w.mu.RLock()
	lastAck, ok := w.debugStreams[ip]
	w.mu.RUnlock()

	if !ok || time.Since(lastAck) < time.Second*15 {
		return ok
	}

	w.mu.Lock()
	delete(w.debugStreams, ip)
	w.mu.Unlock()

	return false
}

func (w *Writer) writeEvent(minerID uint64, chain, eventType, worker, client string, port int, solo bool, data map[string]interface{}) {
	if !w.getEventStream(minerID) {
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
	w.writeEvent(minerID, w.chain, "connect", worker, client, port, solo, nil)
}

func (w *Writer) WriteDisconnectEvent(minerID uint64, worker, client string, port int, solo bool) {
	w.writeEvent(minerID, w.chain, "disconnect", worker, client, port, solo, nil)
}

func (w *Writer) WriteShareEvent(minerID uint64, worker, client string, port int, solo bool, status, reason string, shareDiff, targetDiff uint64) {
	data := map[string]interface{}{"status": status, "share_diff": shareDiff, "target_diff": targetDiff}
	if len(reason) > 0 {
		data["reason"] = reason
	}
	w.writeEvent(minerID, w.chain, "share", worker, client, port, solo, data)
}

func (w *Writer) WriteRetargetEvent(minerID uint64, worker, client string, port int, solo bool, oldDiff, newDiff uint64) {
	data := map[string]interface{}{"old_diff": oldDiff, "new_diff": newDiff}
	w.writeEvent(minerID, w.chain, "retarget", worker, client, port, solo, data)
}

func (w *Writer) WriteDebugRequest(ip string, req *rpc.Request) {
	if !w.getDebugStream(ip) {
		return
	}

	data, err := json.Marshal(req)
	if err != nil {
		w.logger.Error(fmt.Errorf("failed marshaing debug request: %v", err))
		return
	}

	err = w.redis.WriteToStreamDebugChannel(ip, string(data))
	if err != nil {
		w.logger.Error(fmt.Errorf("failed writing debug: %v", err))
	}
}

func (w *Writer) WriteDebugResponse(ip string, data []byte) {
	if !w.getDebugStream(ip) {
		return
	}

	err := w.redis.WriteToStreamDebugChannel(ip, string(data))
	if err != nil {
		w.logger.Error(fmt.Errorf("failed writing debug: %v", err))
	}
}

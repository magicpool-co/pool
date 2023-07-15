package stream

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/maphash"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/puzpuzpuz/xsync/v2"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/redis"
)

func hashUint64(seed maphash.Seed, u uint64) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	binary.Write(&h, binary.LittleEndian, u)
	return h.Sum64()
}

type Writer struct {
	chain   string
	path    string
	ctx     context.Context
	logger  *log.Logger
	redis   *redis.Client
	pubsub  *redis.PubSub
	streams *xsync.MapOf[uint64, time.Time]
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
		streams: xsync.NewTypedMapOf[uint64, time.Time](hashUint64),
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
			w.streams.Range(func(minerID uint64, lastAck time.Time) bool {
				if time.Since(lastAck) > time.Minute {
					w.streams.Delete(minerID)
				}
				return true
			})
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
				w.streams.Store(minerID, time.Now())
			}
		}
	}
}

func (w *Writer) write(minerID uint64, chain, eventType, worker, client string, data map[string]interface{}) {
	lastAck, ok := w.streams.Load(minerID)
	if ok && time.Since(lastAck) > time.Second*15 {
		w.streams.Delete(minerID)
		return
	} else if !ok {
		return
	}

	event := map[string]interface{}{
		"chain":     chain,
		"type":      eventType,
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

func (w *Writer) WriteConnectEvent(minerID uint64, worker, client string) {
	w.write(minerID, w.chain, "connect", worker, client, nil)
}

func (w *Writer) WriteDisconnectEvent(minerID uint64, worker, client string) {
	w.write(minerID, w.chain, "disconnect", worker, client, nil)
}

func (w *Writer) WriteShareEvent(minerID uint64, worker, client, status string, shareDiff, targetDiff uint64) {
	data := map[string]interface{}{"status": status, "share_diff": shareDiff, "target_diff": targetDiff}
	w.write(minerID, w.chain, "share", worker, client, data)
}

func (w *Writer) WriteRetargetEvent(minerID uint64, worker, client string, oldDiff, newDiff uint64) {
	data := map[string]interface{}{"old_diff": oldDiff, "new_diff": newDiff}
	w.write(minerID, w.chain, "retarget", worker, client, data)
}

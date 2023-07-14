package pool

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/types"
)

type stream struct {
	miner   string
	logger  zerolog.Logger
	lastAck time.Time
}

func newStream(miner, path string, logger *log.Logger) (*stream, error) {
	parts := strings.Split(miner, ":")
	if len(parts) == 2 {
		parts[0] = strings.ToLower(parts[0])
		switch parts[0] {
		case "ergo":
			parts[0] = "erg"
		case "kaspa":
			parts[0] = "kas"
		}
		path += "/" + parts[0]
	}
	path += "/" + miner

	output, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	writer := diode.NewWriter(output, 100, 10*time.Millisecond, func(i int) {
		logger.Error(fmt.Errorf("missed %d logs for %s", i, miner))
	})

	s := &stream{
		miner:  miner,
		logger: zerolog.New(writer).With().Timestamp().Logger(),
	}

	return s, nil
}

func (s *stream) write(chain, eventType, worker, client string, strings map[string]string, numbers map[string]uint64) {
	event := s.logger.Log().
		Str("chain", chain).
		Str("type", eventType).
		Str("worker", worker).
		Str("client", client)

	if len(strings) > 0 || len(numbers) > 0 {
		dict := zerolog.Dict()
		for k, v := range strings {
			dict = dict.Str(k, v)
		}

		for k, v := range numbers {
			dict = dict.Uint64(k, v)
		}
		event = event.Dict("data", dict)
	}

	event.Msg("")
}

type streamWriter struct {
	chain   string
	path    string
	ctx     context.Context
	logger  *log.Logger
	pubsub  *redis.PubSub
	streams map[string]*stream
	mu      sync.RWMutex
}

func newStreamWriter(ctx context.Context, chain, path string, logger *log.Logger, redisClient *redis.Client) (*streamWriter, error) {
	pubsub, err := redisClient.GetStreamChannel()
	if err != nil {
		return nil, err
	}

	writer := &streamWriter{
		chain:   chain,
		path:    path,
		ctx:     ctx,
		logger:  logger,
		pubsub:  pubsub,
		streams: make(map[string]*stream),
	}

	go writer.listen()

	return writer, nil
}

func (w *streamWriter) getStream(miner string) *stream {
	w.mu.RLock()
	stream, ok := w.streams[miner]
	w.mu.RUnlock()

	if !ok || time.Since(stream.lastAck) < time.Second*10 {
		return stream
	}

	w.mu.Lock()
	delete(w.streams, miner)
	w.mu.Unlock()

	return nil
}

func (w *streamWriter) listen() {
	ch := w.pubsub.Channel()
	ticker := time.NewTicker(time.Minute)
	var err error
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
			for miner, stream := range w.streams {
				if time.Since(stream.lastAck) > time.Minute {
					delete(w.streams, miner)
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
				miner := parts[1]
				w.mu.Lock()
				if _, ok := w.streams[miner]; !ok {
					w.streams[miner], err = newStream(miner, w.path, w.logger)
					if err != nil {
						w.logger.Error(fmt.Errorf("writing to stream: %s: %v", miner, err))
						continue
					}
				}
				w.streams[miner].lastAck = time.Now()
				w.mu.Unlock()
			}
		}
	}
}

func (w *streamWriter) WriteConnectEvent(miner, worker, client string) {
	if s := w.getStream(miner); s != nil {
		s.write(w.chain, "connect", worker, client, nil, nil)
	}
}

func (w *streamWriter) WriteDisconnectEvent(miner, worker, client string) {
	if s := w.getStream(miner); s != nil {
		s.write(w.chain, "disconnect", worker, client, nil, nil)
	}
}

func (w *streamWriter) WriteShareEvent(miner, worker, client string, status types.ShareStatus, shareDiff, targetDiff uint64) {
	if s := w.getStream(miner); s != nil {
		strings := map[string]string{"status": status.String()}
		numbers := map[string]uint64{"share_diff": shareDiff, "target_diff": targetDiff}
		s.write(w.chain, "share", worker, client, strings, numbers)
	}
}

func (w *streamWriter) WriteRetargetEvent(miner, worker, client string, oldDiff, newDiff uint64) {
	if s := w.getStream(miner); s != nil {
		numbers := map[string]uint64{"old_diff": oldDiff, "new_diff": newDiff}
		s.write(w.chain, "retarget", worker, client, nil, numbers)
	}
}

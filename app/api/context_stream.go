package api

import (
	"fmt"
	"net/http"
	"time"
)

type getMinerStreamArgs struct {
	miner string
}

func (ctx *Context) getMinerStream(args getMinerStreamArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			ctx.writeErrorResponse(w, errStreamingNotSupported)
			return
		}

		minerID, _, err := ctx.getMinerID(args.miner)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ackMsg := fmt.Sprintf("ack:%d", minerID)
		err = ctx.redis.WriteToStreamIndexChannel(ackMsg)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		pubsub, err := ctx.redis.GetStreamMinerChannel(minerID)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for {
				select {
				case <-r.Context().Done():
					pubsub.Close()
					return
				case <-ticker.C:
					err = ctx.redis.WriteToStreamIndexChannel(ackMsg)
					if err != nil {
						pubsub.Close()
						ctx.writeErrorResponse(w, err)
						return
					}
				}
			}
		}()

		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		ch := pubsub.Channel()
		for msg := range ch {
			fmt.Fprintf(w, "data: %s\n\n", msg)
		}
	})
}

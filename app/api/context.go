package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/metrics"
)

type Context struct {
	logger  *log.Logger
	metrics *metrics.Client
}

func (ctx *Context) writeErrorResponse(w http.ResponseWriter, err error) {
	httpErr, ok := err.(httpResponse)
	if ok {
		ctx.logger.Error(err)
	} else {
		ctx.logger.Fatal(err)
		httpErr = errInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErr.Status)
	json.NewEncoder(w).Encode(httpErr)
}

func (ctx *Context) writeOkResponse(w http.ResponseWriter, body interface{}) {
	res := httpResponse{
		Status: 200,
		Data:   body,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	json.NewEncoder(w).Encode(res)
}

func NewContext(logger *log.Logger, metricsClient *metrics.Client) *Context {
	ctx := &Context{
		logger:  logger,
		metrics: metricsClient,
	}

	return ctx
}

func (ctx *Context) getBase() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.writeOkResponse(w, nil)
	})
}

type dashboardArgs struct {
	miner  *string
	worker *string
}

func (ctx *Context) getDashboard(args dashboardArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if args.miner == nil && args.worker != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if args.worker != nil {

		} else if args.miner != nil {

		} else {

		}

		ctx.writeOkResponse(w, nil)
	})
}

type chartArgs struct {
	period string
	miner  *string
	worker *string
}

func (ctx *Context) getCharts(args chartArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch args.period {
		case "15m", "4h", "1d":
		default:
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		}

		if args.miner == nil && args.worker != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if args.worker != nil {

		} else if args.miner != nil {

		} else {

		}

		ctx.writeOkResponse(w, nil)
	})
}

type payoutArgs struct {
	page   string
	size   string
	miner  *string
	worker *string
}

func (ctx *Context) getPayouts(args payoutArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := strconv.ParseUint(args.page, 10, 64)
		if err != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		}

		size, err := strconv.ParseUint(args.size, 10, 64)
		if err != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		}

		if args.miner == nil && args.worker != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if args.worker != nil {

		} else if args.miner != nil {

		} else {

		}

		ctx.writeOkResponse(w, paginatedResponse{Page: page, Size: size})
	})
}

type blockArgs struct {
	page   string
	size   string
	miner  *string
	worker *string
}

func (ctx *Context) getBlocks(args blockArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := strconv.ParseUint(args.page, 10, 64)
		if err != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		}

		size, err := strconv.ParseUint(args.size, 10, 64)
		if err != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		}

		if args.miner == nil && args.worker != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if args.worker != nil {

		} else if args.miner != nil {

		} else {

		}

		ctx.writeOkResponse(w, paginatedResponse{Page: page, Size: size})
	})
}

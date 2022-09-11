package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/magicpool-co/pool/internal/charter"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/metrics"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type Context struct {
	logger  *log.Logger
	metrics *metrics.Client
	pooldb  *dbcl.Client
	tsdb    *dbcl.Client
	redis   *redis.Client
}

func NewContext(logger *log.Logger, metricsClient *metrics.Client, pooldbClient, tsdbClient *dbcl.Client, redisClient *redis.Client) *Context {
	ctx := &Context{
		logger:  logger,
		metrics: metricsClient,
		pooldb:  pooldbClient,
		tsdb:    tsdbClient,
		redis:   redisClient,
	}

	return ctx
}

/* helpers */

func (ctx *Context) writeErrorResponse(w http.ResponseWriter, err error) {
	httpErr, ok := err.(httpResponse)
	if ok {
		if !httpErr.Equals(errRouteNotFound) {
			ctx.logger.Error(err)
		}
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

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (ctx *Context) getMinerID(miner string) (uint64, error) {
	parts := strings.Split(miner, ":")
	if len(parts) != 2 {
		return 0, errMinerNotFound
	}

	minerID, err := pooldb.GetMinerID(ctx.pooldb.Reader(), parts[0], parts[1])
	if err != nil {
		return 0, err
	} else if minerID == 0 {
		return 0, errMinerNotFound
	}

	return minerID, nil
}

func (ctx *Context) getWorkerID(minerID uint64, worker string) (uint64, error) {
	workerID, err := pooldb.GetWorkerID(ctx.pooldb.Reader(), minerID, worker)
	if err != nil {
		return 0, err
	} else if workerID == 0 {
		return 0, errWorkerNotFound
	}

	return workerID, nil
}

/* routes */

func (ctx *Context) getBase() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.writeOkResponse(w, nil)
	})
}

type dashboardArgs struct {
	miner  string
	worker string
}

func (ctx *Context) getDashboard(args dashboardArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if args.miner == "" && args.worker != "" {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if args.worker != "" {

		} else if args.miner != "" {

		} else {

		}

		ctx.writeOkResponse(w, nil)
	})
}

type blockChartArgs struct {
	period string
}

func (ctx *Context) getBlockCharts(args blockChartArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		period, err := types.ParsePeriodType(args.period)
		if err != nil {
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		}

		data, err := charter.FetchBlocks(ctx.tsdb, period)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, data)
	})
}

type shareChartArgs struct {
	period string
	miner  string
	worker string
}

func (ctx *Context) getShareCharts(args shareChartArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		period, err := types.ParsePeriodType(args.period)
		if err != nil {
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		}

		var data interface{}
		if args.miner == "" && args.worker != "" {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if args.worker != "" {
			minerID, err := ctx.getMinerID(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			workerID, err := ctx.getWorkerID(minerID, args.worker)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			data, err = charter.FetchWorkerShares(ctx.tsdb, workerID, period)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}
		} else if args.miner != "" {
			minerID, err := ctx.getMinerID(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			data, err = charter.FetchMinerShares(ctx.tsdb, minerID, period)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}
		} else {
			var err error
			data, err = charter.FetchGlobalShares(ctx.tsdb, period)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}
		}

		ctx.writeOkResponse(w, data)
	})
}

type payoutArgs struct {
	page   string
	size   string
	miner  string
	worker string
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

		if args.miner == "" && args.worker != "" {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if args.worker != "" {

		} else if args.miner != "" {

		} else {

		}

		ctx.writeOkResponse(w, paginatedResponse{Page: page, Size: size})
	})
}

type blockArgs struct {
	page   string
	size   string
	miner  string
	worker string
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

		if args.miner == "" && args.worker != "" {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if args.worker != "" {

		} else if args.miner != "" {

		} else {

		}

		ctx.writeOkResponse(w, paginatedResponse{Page: page, Size: size})
	})
}

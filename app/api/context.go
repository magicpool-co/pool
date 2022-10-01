package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/core/stats"
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
	stats   *stats.Client
}

func NewContext(logger *log.Logger, metricsClient *metrics.Client, pooldbClient, tsdbClient *dbcl.Client, redisClient *redis.Client) *Context {
	ctx := &Context{
		logger:  logger,
		metrics: metricsClient,
		pooldb:  pooldbClient,
		tsdb:    tsdbClient,
		redis:   redisClient,
		stats:   stats.New(pooldbClient, tsdbClient, redisClient),
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}

func (ctx *Context) parsePageSize(rawPage, rawSize string) (uint64, uint64, error) {
	page, err := strconv.ParseUint(rawPage, 10, 64)
	if err != nil {
		return 0, 0, errInvalidParameters
	}

	size, err := strconv.ParseUint(rawSize, 10, 64)
	if err != nil {
		return 0, 0, errInvalidParameters
	}

	return page, size, nil
}

func (ctx *Context) getMinerID(miner string) (uint64, error) {
	minerIDs, err := ctx.getMinerIDs(miner)
	if err != nil {
		return 0, err
	} else if len(minerIDs) != 1 {
		return 0, errMinerNotFound
	}

	return minerIDs[0], nil
}

func (ctx *Context) getMinerIDs(rawMiner string) ([]uint64, error) {
	miners := strings.Split(rawMiner, ",")
	if len(miners) > 10 {
		return nil, errTooManyMiners
	}

	minerIDs := make([]uint64, len(miners))
	for i, miner := range miners {
		var err error
		minerIDs[i], err = ctx.redis.GetMinerID(miner)
		if err != nil || minerIDs[i] == 0 {
			if err != nil {
				ctx.logger.Error(err)
			}

			parts := strings.Split(miner, ":")
			if len(parts) != 2 {
				return nil, errMinerNotFound
			}

			minerIDs[i], err = pooldb.GetMinerID(ctx.pooldb.Reader(), parts[0], parts[1])
			if err != nil {
				return nil, err
			} else if minerIDs[i] == 0 {
				return nil, errMinerNotFound
			}
		}
	}

	return minerIDs, nil
}

func (ctx *Context) getWorkerID(minerID uint64, worker string) (uint64, error) {
	workerID, err := ctx.redis.GetWorkerID(minerID, worker)
	if err != nil || workerID == 0 {
		if err != nil {
			ctx.logger.Error(err)
		}

		workerID, err = pooldb.GetWorkerID(ctx.pooldb.Reader(), minerID, worker)
		if err != nil {
			return 0, err
		} else if workerID == 0 {
			return 0, errWorkerNotFound
		}
	}

	return workerID, nil
}

/* routes */

func (ctx *Context) getBase() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.writeOkResponse(w, nil)
	})
}

type existsArgs struct {
	miner  string
	worker string
}

func (ctx *Context) getExists(args existsArgs) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exists := true
		if args.miner == "" {
			exists = false
		} else {
			minerID, err := ctx.getMinerID(args.miner)
			if err != nil {
				exists = false
			}

			if exists && args.worker != "" {
				_, err := ctx.getWorkerID(minerID, args.worker)
				if err != nil {
					exists = false
				}
			}
		}

		ctx.writeOkResponse(w, map[string]interface{}{"exists": exists})
	})
}

type minersArgs struct {
	page string
	size string
}

func (ctx *Context) getMiners(args minersArgs) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, size, err := ctx.parsePageSize(args.page, args.size)
		if err != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		}

		miners, count, err := ctx.stats.GetMiners(page, size)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		items := make([]interface{}, len(miners))
		for i, miner := range miners {
			items[i] = miner
		}

		ctx.writeOkResponse(w, paginatedResponse{Page: page, Size: size, Results: count, Items: items})
	})
}

type workersArgs struct {
	miner string
	page  string
	size  string
}

func (ctx *Context) getWorkers(args workersArgs) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, size, err := ctx.parsePageSize(args.page, args.size)
		if err != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		}

		minerID, err := ctx.getMinerID(args.miner)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		workers, err := ctx.stats.GetWorkers(minerID, page, size)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, workers)
	})
}

type dashboardArgs struct {
	miner  string
	worker string
}

func (ctx *Context) getDashboard(args dashboardArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var dashboard *stats.Dashboard
		var err error
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

			dashboard, err = ctx.stats.GetWorkerDashboard(workerID)
		} else if args.miner != "" {
			minerIDs, err := ctx.getMinerIDs(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			dashboard, err = ctx.stats.GetMinerDashboard(minerIDs)
		} else {
			dashboard, err = ctx.stats.GetGlobalDashboard()
		}

		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, dashboard)
	})
}

type blockChartArgs struct {
	chain  string
	period string
}

func (ctx *Context) getBlockChart(args blockChartArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		period, err := types.ParsePeriodType(args.period)
		if err != nil {
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		} else if !validateChain(args.chain) {
			ctx.writeErrorResponse(w, errChainNotFound)
			return
		}

		data, err := ctx.stats.GetBlockChart(args.chain, period)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, data)
	})
}

type blockProfitabilityChartArgs struct {
	period string
}

func (ctx *Context) getBlockProfitabilityChart(args blockProfitabilityChartArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		period, err := types.ParsePeriodType(args.period)
		if err != nil {
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		}

		data, err := ctx.stats.GetBlockProfitabilityChart(period)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, data)
	})
}

type roundChartArgs struct {
	chain  string
	period string
}

func (ctx *Context) getRoundChart(args roundChartArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		period, err := types.ParsePeriodType(args.period)
		if err != nil {
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		} else if !validateChain(args.chain) {
			ctx.writeErrorResponse(w, errChainNotFound)
			return
		}

		data, err := ctx.stats.GetRoundChart(args.chain, period)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, data)
	})
}

type shareChartArgs struct {
	chain  string
	period string
	miner  string
	worker string
}

func (ctx *Context) getShareChart(args shareChartArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		period, err := types.ParsePeriodType(args.period)
		if err != nil {
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		} else if !validateChain(args.chain) {
			ctx.writeErrorResponse(w, errChainNotFound)
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

			data, err = ctx.stats.GetWorkerShareChart(workerID, args.chain, period)
		} else if args.miner != "" {
			minerIDs, err := ctx.getMinerIDs(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			data, err = ctx.stats.GetMinerShareChart(minerIDs, args.chain, period)
		} else {
			data, err = ctx.stats.GetGlobalShareChart(args.chain, period)
		}

		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, data)
	})
}

type payoutArgs struct {
	page  string
	size  string
	miner string
}

func (ctx *Context) getPayouts(args payoutArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, size, err := ctx.parsePageSize(args.page, args.size)
		if err != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		}

		var payouts []*stats.Payout
		var count uint64
		if args.miner != "" {
			minerIDs, err := ctx.getMinerIDs(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			payouts, count, err = ctx.stats.GetMinerPayouts(minerIDs, page, size)
		} else {
			payouts, count, err = ctx.stats.GetGlobalPayouts(page, size)
		}

		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		items := make([]interface{}, len(payouts))
		for i, payout := range payouts {
			items[i] = payout
		}

		ctx.writeOkResponse(w, paginatedResponse{Page: page, Size: size, Results: count, Items: items})
	})
}

type roundArgs struct {
	page  string
	size  string
	miner string
}

func (ctx *Context) getRounds(args roundArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, size, err := ctx.parsePageSize(args.page, args.size)
		if err != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		}

		var rounds []*stats.Round
		var count uint64
		if args.miner != "" {
			minerIDs, err := ctx.getMinerIDs(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			rounds, count, err = ctx.stats.GetMinerRounds(minerIDs, page, size)
		} else {
			rounds, count, err = ctx.stats.GetGlobalRounds(page, size)
		}

		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		items := make([]interface{}, len(rounds))
		for i, round := range rounds {
			items[i] = round
		}

		ctx.writeOkResponse(w, paginatedResponse{Page: page, Size: size, Results: count, Items: items})
	})
}

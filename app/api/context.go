package api

import (
	"fmt"
	"math/big"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/core/export"
	"github.com/magicpool-co/pool/core/stats"
	"github.com/magicpool-co/pool/core/stream"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/metrics"
	"github.com/magicpool-co/pool/internal/pooldb"
	"github.com/magicpool-co/pool/internal/redis"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/dbcl"
	"github.com/magicpool-co/pool/types"
)

type Context struct {
	logger        *log.Logger
	metrics       *metrics.Client
	pooldb        *dbcl.Client
	tsdb          *dbcl.Client
	redis         *redis.Client
	stats         *stats.Client
	nodes         []types.MiningNode
	streamManager *stream.Manager
}

func NewContext(
	logger *log.Logger,
	metricsClient *metrics.Client,
	pooldbClient, tsdbClient *dbcl.Client,
	redisClient *redis.Client,
	nodes []types.MiningNode,
	cacheEnabled bool,
) *Context {
	statsChains := []string{
		"CFX",
		"ERG",
		"ETC",
		"FIRO",
		"FLUX",
		"KAS",
		"NEXA",
		"RVN",
	}

	ctx := &Context{
		logger:        logger,
		metrics:       metricsClient,
		pooldb:        pooldbClient,
		tsdb:          tsdbClient,
		redis:         redisClient,
		stats:         stats.New(pooldbClient, tsdbClient, redisClient, statsChains, cacheEnabled),
		nodes:         nodes,
		streamManager: stream.NewManager(logger, redisClient),
	}

	return ctx
}

/* helpers */

func (ctx *Context) writeErrorResponse(w http.ResponseWriter, err error) {
	httpErr, ok := err.(httpResponse)
	if ok {
		if httpErr.shouldLog {
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

func (ctx *Context) writeCSVResponse(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "text/csv")
	w.WriteHeader(200)
	w.Write(body)
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

func (ctx *Context) getMinerID(miner string) (uint64, string, error) {
	minerIDs, chains, err := ctx.getMinerIDs(miner)
	if err != nil {
		return 0, "", err
	} else if len(minerIDs) != 1 {
		return 0, "", errMinerNotFound
	}

	return minerIDs[0], chains[0], nil
}

func (ctx *Context) getMinerIDs(rawMiner string) ([]uint64, []string, error) {
	miners := strings.Split(rawMiner, ",")
	if len(miners) > 10 {
		return nil, nil, errTooManyMiners
	}

	minerIDs := make([]uint64, len(miners))
	chains := make([]string, len(miners))
	for i, miner := range miners {
		var address string
		var err error

		chains[i], address, err = parseMiner(miner)
		if err != nil {
			return nil, nil, err
		}

		minerIDs[i], err = ctx.redis.GetMinerID(miner)
		if err != nil || minerIDs[i] == 0 {
			if err != nil {
				ctx.logger.Error(err)
			}

			minerIDs[i], err = pooldb.GetMinerIDByChainAddress(ctx.pooldb.Reader(), chains[i], address)
			if err != nil {
				return nil, nil, err
			} else if minerIDs[i] == 0 {
				return nil, nil, errMinerNotFound
			}
		}
	}

	return minerIDs, chains, nil
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

func (ctx *Context) getMinerIPAddress(minerID uint64) (*pooldb.IPAddress, error) {
	lastIP, err := pooldb.GetOldestActiveIPAddress(ctx.pooldb.Reader(), minerID)
	if err != nil {
		return nil, err
	} else if lastIP == nil {
		lastIP, err = pooldb.GetNewestInactiveIPAddress(ctx.pooldb.Reader(), minerID)
		if err != nil {
			return nil, err
		}
	}

	return lastIP, nil
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
		var chain *string
		if args.miner == "" {
			exists = false
		} else {
			minerID, minerChain, err := ctx.getMinerID(args.miner)
			if err != nil {
				exists = false
			}
			chain = types.StringPtr(minerChain)

			if exists && args.worker != "" {
				_, err := ctx.getWorkerID(minerID, args.worker)
				if err != nil {
					exists = false
				}
			}
		}

		data := map[string]interface{}{
			"exists": exists,
			"chain":  chain,
		}

		ctx.writeOkResponse(w, data)
	})
}

type validateAddressArgs struct {
	miner string
}

func (ctx *Context) getValidateAddress(args validateAddressArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var valid bool
		chain, address, err := parseMiner(args.miner)
		if err == nil {
			valid = ValidateAddress(chain, address)
		}

		ctx.writeOkResponse(w, map[string]interface{}{"valid": valid})
	})
}

type minersArgs struct {
	chain string
	page  string
	size  string
}

func (ctx *Context) getMiners(args minersArgs) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chain := strings.ToUpper(args.chain)
		page, size, err := ctx.parsePageSize(args.page, args.size)
		if err != nil {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if !validateMiningChain(chain) {
			ctx.writeErrorResponse(w, errChainNotFound)
			return
		}

		miners, count, err := ctx.stats.GetMiners(chain, page, size)
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
}

func (ctx *Context) getWorkers(args workersArgs) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		minerID, _, err := ctx.getMinerID(args.miner)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		workers, err := ctx.stats.GetWorkers(minerID)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, workers)
	})
}

func (ctx *Context) getPools() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pools, err := ctx.stats.GetPoolSummary(ctx.nodes)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, pools)
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
		if args.miner == "" {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if args.worker != "" {
			minerID, _, err := ctx.getMinerID(args.miner)
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
			minerIDs, chains, err := ctx.getMinerIDs(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			minerIdx := make(map[uint64]string, len(minerIDs))
			for i, minerID := range minerIDs {
				minerIdx[minerID] = chains[i]
			}

			dashboard, err = ctx.stats.GetMinerDashboard(minerIdx)
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
		chain := strings.ToUpper(args.chain)
		period, err := types.ParsePeriodType(args.period)
		if err != nil {
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		} else if !validateMiningChain(chain) {
			ctx.writeErrorResponse(w, errChainNotFound)
			return
		}

		data, err := ctx.stats.GetBlockChart(chain, period)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, data)
	})
}

type blockMetricChartArgs struct {
	metric  string
	period  string
	average bool
}

func (ctx *Context) getBlockMetricChart(args blockMetricChartArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metric, err := types.ParseNetworkMetric(args.metric)
		if err != nil {
			ctx.writeErrorResponse(w, errMetricNotFound)
			return
		}

		period, err := types.ParsePeriodType(args.period)
		if err != nil {
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		}

		data, err := ctx.stats.GetBlockSingleMetricChart(metric, period, args.average)
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
		chain := strings.ToUpper(args.chain)
		period, err := types.ParsePeriodType(args.period)
		if err != nil {
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		} else if !validateMiningChain(chain) {
			ctx.writeErrorResponse(w, errChainNotFound)
			return
		}

		var data interface{}
		if args.miner == "" && args.worker != "" {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		} else if args.worker != "" {
			minerID, _, err := ctx.getMinerID(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			workerID, err := ctx.getWorkerID(minerID, args.worker)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			data, err = ctx.stats.GetWorkerShareChart(workerID, chain, period)
		} else if args.miner != "" {
			minerIDs, _, err := ctx.getMinerIDs(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			data, err = ctx.stats.GetMinerShareChart(minerIDs, chain, period)
		} else {
			data, err = ctx.stats.GetGlobalShareChart(chain, period)
		}

		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, data)
	})
}

type shareMetricChartArgs struct {
	metric string
	period string
	miner  string
	worker string
}

func (ctx *Context) getShareMetricChart(args shareMetricChartArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metric, err := types.ParseShareMetric(args.metric)
		if err != nil {
			ctx.writeErrorResponse(w, errMetricNotFound)
			return
		}

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
			minerID, _, err := ctx.getMinerID(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			workerID, err := ctx.getWorkerID(minerID, args.worker)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			data, err = ctx.stats.GetWorkerShareSingleMetricChart(workerID, metric, period)
		} else if args.miner != "" {
			minerIDs, _, err := ctx.getMinerIDs(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			data, err = ctx.stats.GetMinerShareSingleMetricChart(minerIDs, metric, period)
		} else {
			data, err = ctx.stats.GetGlobalShareSingleMetricChart(metric, period)
		}

		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, data)
	})
}

type earningMetricChartArgs struct {
	metric string
	period string
	miner  string
	worker string
}

func (ctx *Context) getEarningMetricChart(args earningMetricChartArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metric, err := types.ParseEarningMetric(args.metric)
		if err != nil {
			ctx.writeErrorResponse(w, errMetricNotFound)
			return
		}

		period, err := types.ParsePeriodType(args.period)
		if err != nil || period != types.Period1d {
			ctx.writeErrorResponse(w, errPeriodNotFound)
			return
		}

		minerIDs, _, err := ctx.getMinerIDs(args.miner)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		data, err := ctx.stats.GetMinerEarningChart(minerIDs, metric, period)
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
			minerIDs, _, err := ctx.getMinerIDs(args.miner)
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

type payoutExportArgs struct {
	exportType string
	miner      string
}

func (ctx *Context) getPayoutsExport(args payoutExportArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if args.exportType != "csv" || args.miner == "" {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		}

		minerIDs, _, err := ctx.getMinerIDs(args.miner)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		} else if len(minerIDs) != 1 {
			ctx.writeErrorResponse(w, errInvalidParameters)
			return
		}

		payouts, err := ctx.stats.GetAllMinerPayouts(minerIDs[0])
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		exported, err := export.ExportPayoutsAsCSV(payouts)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeCSVResponse(w, exported)
	})
}

type roundArgs struct {
	page  string
	size  string
	miner string
	chain string
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
			minerIDs, _, err := ctx.getMinerIDs(args.miner)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			rounds, count, err = ctx.stats.GetMinerRounds(minerIDs, page, size)
		} else if args.chain != "" {
			if !validateMiningChain(args.chain) {
				ctx.writeErrorResponse(w, errChainNotFound)
				return
			}

			rounds, count, err = ctx.stats.GetGlobalRoundsByChain(args.chain, page, size)
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

type thresholdBoundsArgs struct {
	chain string
}

func (ctx *Context) getThresholdBounds(args thresholdBoundsArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bounds, err := common.GetDefaultPayoutBounds(args.chain)
		if err != nil {
			ctx.writeErrorResponse(w, errChainNotFound)
			return
		}

		units, err := common.GetDefaultUnits(args.chain)
		if err != nil {
			ctx.writeErrorResponse(w, errChainNotFound)
			return
		}

		data := map[string]interface{}{
			"chain":     strings.ToUpper(args.chain),
			"min":       common.BigIntToFloat64(bounds.Min, units),
			"default":   common.BigIntToFloat64(bounds.Default, units),
			"max":       common.BigIntToFloat64(bounds.Max, units),
			"precision": bounds.Precision,
			"units":     bounds.Units,
		}

		ctx.writeOkResponse(w, data)
	})
}

type minerSettingsArgs struct {
	miner string
}

func (ctx *Context) getMinerSettings(args minerSettingsArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		minerID, _, err := ctx.getMinerID(args.miner)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		var ipHint, workerName *string
		lastIP, err := ctx.getMinerIPAddress(minerID)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		} else if lastIP != nil {
			obscuredIP, err := common.ObscureIP(lastIP.IPAddress)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}
			ipHint = types.StringPtr(obscuredIP)

			if lastIP.WorkerID != 0 {
				worker, err := pooldb.GetWorker(ctx.pooldb.Reader(), lastIP.WorkerID)
				if err != nil {
					ctx.writeErrorResponse(w, err)
					return
				} else if worker != nil {
					workerName = types.StringPtr(worker.Name)
				}
			}
		}

		miner, err := pooldb.GetMiner(ctx.pooldb.Reader(), minerID)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		units, err := common.GetDefaultUnits(miner.ChainID)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		var obscuredEmail *string
		if miner.Email != nil {
			parts := strings.Split(types.StringValue(miner.Email), "@")
			if len(parts) == 2 {
				if len(parts[0]) <= 3 {
					parts[0] = "***"
				} else {
					parts[0] = "***" + parts[0][3:]
				}
				obscuredEmail = types.StringPtr(parts[0] + "@" + parts[1])
			}
		}

		var threshold float64
		if miner.Threshold.Valid {
			threshold = common.BigIntToFloat64(miner.Threshold.BigInt, units)
		} else {
			payoutBound, err := common.GetDefaultPayoutBounds(miner.ChainID)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			threshold = common.BigIntToFloat64(payoutBound.Default, units)
		}

		data := map[string]interface{}{
			"chain":                      miner.ChainID,
			"worker":                     workerName,
			"email":                      obscuredEmail,
			"threshold":                  threshold,
			"enabledWorkerNotifications": miner.EnabledWorkerNotifications,
			"enabledPayoutNotifications": miner.EnabledPayoutNotifications,
			"ipHint":                     ipHint,
		}

		ctx.writeOkResponse(w, data)
	})
}

type updateMinerSettingsArgs struct {
	miner                     string
	IP                        string  `json:"ip"`
	Email                     *string `json:"email"`
	EnableWorkerNotifications *bool   `json:"enableWorkerNotifications"`
	EnablePayoutNotifications *bool   `json:"enablePayoutNotifications"`
	Threshold                 *string `json:"threshold"`
}

func (ctx *Context) updateMinerSettings(args updateMinerSettingsArgs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		minerID, _, err := ctx.getMinerID(args.miner)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ip, err := ctx.getMinerIPAddress(minerID)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		} else if ip.IPAddress != args.IP {
			ctx.writeErrorResponse(w, errIncorrectIPAddress)
			return
		}

		miner, err := pooldb.GetMiner(ctx.pooldb.Reader(), minerID)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		hasEmail := miner.Email != nil
		if args.Email != nil {
			_, err := mail.ParseAddress(types.StringValue(args.Email))
			if err != nil {
				ctx.writeErrorResponse(w, errInvalidEmail)
				return
			}

			miner.Email = args.Email
			hasEmail = true
		}

		if hasEmail {
			if args.EnableWorkerNotifications != nil {
				miner.EnabledWorkerNotifications = types.BoolValue(args.EnableWorkerNotifications)
			}

			if args.EnablePayoutNotifications != nil {
				miner.EnabledPayoutNotifications = types.BoolValue(args.EnablePayoutNotifications)
			}
		}

		if args.Threshold != nil {
			units, err := common.GetDefaultUnits(miner.ChainID)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			}

			rawThreshold := types.StringValue(args.Threshold)
			if len(strings.Split(rawThreshold, ".")) == 1 {
				rawThreshold += ".0"
			}

			threshold, err := common.StringDecimalToBigint(rawThreshold, units)
			if err != nil {
				ctx.writeErrorResponse(w, errInvalidThreshold)
				return
			}

			payoutBound, err := common.GetDefaultPayoutBounds(miner.ChainID)
			if err != nil {
				ctx.writeErrorResponse(w, err)
				return
			} else if threshold.Cmp(payoutBound.Min) < 0 {
				ctx.writeErrorResponse(w, errThresholdTooSmall)
				return
			} else if threshold.Cmp(payoutBound.Max) > 0 {
				ctx.writeErrorResponse(w, errThresholdTooBig)
				return
			} else if new(big.Int).Mod(threshold, payoutBound.PrecisionMask()).Cmp(common.Big0) > 0 {
				ctx.writeErrorResponse(w, errThresholdTooPrecise)
				return
			}

			miner.Threshold = dbcl.NullBigInt{Valid: true, BigInt: threshold}
		}

		cols := []string{"email", "threshold",
			"enabled_worker_notifications", "enabled_payout_notifications"}
		err = pooldb.UpdateMiner(ctx.pooldb.Writer(), miner, cols)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		ctx.writeOkResponse(w, nil)
	})
}

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

		subID, ch, err := ctx.streamManager.Subscribe(minerID)
		if err != nil {
			ctx.writeErrorResponse(w, err)
			return
		}

		go func() {
			<-r.Context().Done()
			ctx.streamManager.Unsubscribe(minerID, subID)
		}()

		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		for msg := range ch {
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	})
}

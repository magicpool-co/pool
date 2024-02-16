// heavily inspired by https://benhoyt.com/writings/go-routing/

package api

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type router struct {
	ctx   *Context
	cache map[string]*regexp.Regexp
	mu    sync.RWMutex
}

func newRouter(ctx *Context) http.Handler {
	r := router{
		ctx:   ctx,
		cache: make(map[string]*regexp.Regexp),
	}

	return r
}

func (rtr router) compileExpr(pattern string) *regexp.Regexp {
	rtr.mu.RLock()
	regex := rtr.cache[pattern]
	rtr.mu.RUnlock()

	if regex == nil {
		rtr.mu.Lock()
		regex = regexp.MustCompile("^" + pattern + "$")
		rtr.cache[pattern] = regex
		rtr.mu.Unlock()
	}

	return regex
}

func (rtr router) match(path, pattern string, vars ...interface{}) bool {
	expr := rtr.compileExpr(strings.ReplaceAll(pattern, "+", "([^/]+)"))
	matches := expr.FindStringSubmatch(path)
	if len(matches) == 0 {
		return false
	} else if len(matches) != len(vars)+1 {
		return false
	}

	for i, match := range matches[1:] {
		switch p := vars[i].(type) {
		case *string:
			*p = match
		}
	}

	return true
}

func (rtr router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler
	var method string
	var miner, worker, metric string

	path := r.URL.Path
	switch {
	case path == "/":
		method = "GET"
		handler = rtr.ctx.getBase()
	case rtr.match(path, "/healthcheck"):
		method = "GET"
		handler = rtr.ctx.getBase()
	case rtr.match(path, "/global/pools"):
		method = "GET"
		handler = rtr.ctx.getPools()
	case rtr.match(path, "/global/dashboard"):
		method = "GET"
		handler = rtr.ctx.getDashboard(dashboardArgs{})
	case rtr.match(path, "/global/charts/blocks"):
		method = "GET"
		chain := r.URL.Query().Get("chain")
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getBlockChart(blockChartArgs{
			chain:  chain,
			period: period,
		})
	case rtr.match(path, "/global/charts/blocks/+", &metric):
		method = "GET"
		period := r.URL.Query().Get("period")
		average := strings.ToLower(r.URL.Query().Get("average")) == "true"
		handler = rtr.ctx.getBlockMetricChart(blockMetricChartArgs{
			metric:  metric,
			period:  period,
			average: average,
		})
	case rtr.match(path, "/global/charts/shares"):
		method = "GET"
		chain := r.URL.Query().Get("chain")
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getShareChart(shareChartArgs{
			chain:  chain,
			period: period,
		})
	case rtr.match(path, "/global/charts/shares/+", &metric):
		method = "GET"
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getShareMetricChart(shareMetricChartArgs{
			metric: metric,
			period: period,
		})
	case rtr.match(path, "/global/rounds"):
		method = "GET"
		chain := r.URL.Query().Get("chain")
		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		handler = rtr.ctx.getRounds(roundArgs{
			chain: chain,
			page:  page,
			size:  size,
		})
	case rtr.match(path, "/global/payouts"):
		method = "GET"
		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		handler = rtr.ctx.getPayouts(payoutArgs{
			page: page,
			size: size,
		})
	case rtr.match(path, "/global/miners"):
		method = "GET"
		chain := r.URL.Query().Get("chain")
		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		handler = rtr.ctx.getMiners(minersArgs{
			chain: chain,
			page:  page,
			size:  size,
		})
	case rtr.match(path, "/config/threshold"):
		method = "GET"
		chain := r.URL.Query().Get("chain")
		handler = rtr.ctx.getThresholdBounds(thresholdBoundsArgs{
			chain: chain,
		})
	case rtr.match(path, "/miner/+", &miner):
		method = "GET"
		handler = rtr.ctx.getExists(existsArgs{
			miner: miner,
		})
	case rtr.match(path, "/miner/+/validate", &miner):
		method = "GET"
		handler = rtr.ctx.getValidateAddress(validateAddressArgs{
			miner: miner,
		})
	case rtr.match(path, "/miner/+/dashboard", &miner):
		method = "GET"
		handler = rtr.ctx.getDashboard(dashboardArgs{
			miner: miner,
		})
	case rtr.match(path, "/miner/+/charts/shares", &miner):
		method = "GET"
		chain := r.URL.Query().Get("chain")
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getShareChart(shareChartArgs{
			chain:  chain,
			period: period,
			miner:  miner,
		})
	case rtr.match(path, "/miner/+/charts/shares/+", &miner, &metric):
		method = "GET"
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getShareMetricChart(shareMetricChartArgs{
			metric: metric,
			period: period,
			miner:  miner,
		})
	case rtr.match(path, "/miner/+/charts/earnings/+", &miner, &metric):
		method = "GET"
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getEarningMetricChart(earningMetricChartArgs{
			metric: metric,
			period: period,
			miner:  miner,
		})
	case rtr.match(path, "/miner/+/rounds", &miner):
		method = "GET"
		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		handler = rtr.ctx.getRounds(roundArgs{
			page:  page,
			size:  size,
			miner: miner,
		})
	case rtr.match(path, "/miner/+/payouts", &miner):
		method = "GET"
		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		handler = rtr.ctx.getPayouts(payoutArgs{
			page:  page,
			size:  size,
			miner: miner,
		})
	case rtr.match(path, "/miner/+/payouts/export", &miner):
		method = "GET"
		exportType := r.URL.Query().Get("type")
		handler = rtr.ctx.getPayoutsExport(payoutExportArgs{
			exportType: exportType,
			miner:      miner,
		})
	case rtr.match(path, "/miner/+/workers", &miner):
		method = "GET"
		handler = rtr.ctx.getWorkers(workersArgs{
			miner: miner,
		})
	case rtr.match(path, "/miner/+/settings", &miner):
		switch r.Method {
		case "GET":
			method = "GET"
			handler = rtr.ctx.getMinerSettings(minerSettingsArgs{
				miner: miner,
			})
		case "POST":
			method = "POST"
			args := updateMinerSettingsArgs{
				miner: miner,
			}
			err := decodeJSONBody(w, r, &args)
			if err != nil {
				rtr.ctx.writeErrorResponse(w, errInvalidJSONBody)
				return
			}
			handler = rtr.ctx.updateMinerSettings(args)
		}
	case rtr.match(path, "/miner/+/stream", &miner):
		method = "GET"
		handler = rtr.ctx.getMinerStream(getMinerStreamArgs{
			miner: miner,
		})
	case rtr.match(path, "/worker/+/+", &miner, &worker):
		method = "GET"
		handler = rtr.ctx.getExists(existsArgs{
			miner:  miner,
			worker: worker,
		})
	case rtr.match(path, "/worker/+/+/dashboard", &miner, &worker):
		method = "GET"
		handler = rtr.ctx.getDashboard(dashboardArgs{
			miner:  miner,
			worker: worker,
		})
	case rtr.match(path, "/worker/+/+/charts/shares", &miner, &worker):
		method = "GET"
		chain := r.URL.Query().Get("chain")
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getShareChart(shareChartArgs{
			chain:  chain,
			period: period,
			miner:  miner,
			worker: worker,
		})
	case rtr.match(path, "/worker/+/+/charts/shares/+", &miner, &worker, &metric):
		method = "GET"
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getShareMetricChart(shareMetricChartArgs{
			metric: metric,
			period: period,
			miner:  miner,
			worker: worker,
		})
	default:
		rtr.ctx.writeErrorResponse(w, errRouteNotFound)
		return
	}

	if r.Method == "HEAD" {
		w.Header().Set("Allow", method)
		rtr.ctx.writeOkResponse(w, nil)
		return
	} else if r.Method != method {
		w.Header().Set("Allow", method)
		rtr.ctx.writeErrorResponse(w, errMethodNotAllowed)
		return
	}

	handler.ServeHTTP(w, r)
}

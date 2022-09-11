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
	var miner, worker string

	path := r.URL.Path
	switch {
	case path == "/":
		method = "GET"
		handler = rtr.ctx.getBase()

	case rtr.match(path, "/healthcheck"):
		method = "GET"
		handler = rtr.ctx.getBase()

	case rtr.match(path, "/global/dashboard"):
		method = "GET"
		handler = rtr.ctx.getDashboard(dashboardArgs{})

	case rtr.match(path, "/global/charts/shares"):
		method = "GET"
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getShareCharts(shareChartArgs{period: period})

	case rtr.match(path, "/global/charts/blocks"):
		method = "GET"
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getBlockCharts(blockChartArgs{period: period})

	case rtr.match(path, "/global/blocks"):
		method = "GET"
		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		handler = rtr.ctx.getBlocks(blockArgs{page: page, size: size})

	case rtr.match(path, "/global/payouts"):
		method = "GET"
		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		handler = rtr.ctx.getPayouts(payoutArgs{page: page, size: size})

	case rtr.match(path, "/miner/+/dashboard", &miner):
		method = "GET"
		handler = rtr.ctx.getDashboard(dashboardArgs{miner: miner})

	case rtr.match(path, "/miner/+/charts/shares", &miner):
		method = "GET"
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getShareCharts(shareChartArgs{period: period, miner: miner})

	case rtr.match(path, "/miner/+/blocks", &miner):
		method = "GET"
		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		handler = rtr.ctx.getBlocks(blockArgs{page: page, size: size, miner: miner})

	case rtr.match(path, "/miner/+/payouts", &miner):
		method = "GET"
		page, size := r.URL.Query().Get("page"), r.URL.Query().Get("size")
		handler = rtr.ctx.getPayouts(payoutArgs{page: page, size: size, miner: miner})

	case rtr.match(path, "/worker/+/+/dashboard", &miner, &worker):
		method = "GET"
		handler = rtr.ctx.getDashboard(dashboardArgs{miner: miner, worker: worker})

	case rtr.match(path, "/worker/+/+/charts/shares", &miner, &worker):
		method = "GET"
		period := r.URL.Query().Get("period")
		handler = rtr.ctx.getShareCharts(shareChartArgs{period: period, miner: miner, worker: worker})

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

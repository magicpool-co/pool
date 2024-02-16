package api

import (
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

var (
	corsAllowHeaders = strings.Join([]string{
		"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token",
		"Authorization", "accept", "origin", "Cache-Control", "X-Requested-With",
	}, ", ")
	corsAllowMethods = "POST, OPTIONS, GET, PUT, PATCH, DELETE"
)

func corsMiddleware(ctx *Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", corsAllowHeaders)
		w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)

		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(ctx *Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				ctx.logger.Panic(r, string(debug.Stack()))
				ctx.writeErrorResponse(w, errInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func rateLimiterMiddleware(ctx *Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func metricsMiddleware(ctx *Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		elapsed := float64(time.Since(start)) / float64(time.Millisecond)

		// @TODO: either remove status/resSize or reimplement ResponseWriter to store them
		// https://github.com/urfave/negroni/blob/master/response_writer.go
		// https://stackoverflow.com/questions/13979078/how-to-get-the-current-response-length-from-a-http-responsewriter
		method, path, status := r.Method, r.URL.Path, "200"
		reqSize := requestSize(r)
		resSize := float64(0)

		if ctx.metrics != nil {
			ctx.metrics.ObserveHistogram("request_duration_ms", elapsed, status, method, path)
			ctx.metrics.IncrementCounter("requests_total", status, method, path)
			ctx.metrics.ObserveSummary("request_size_bytes", reqSize, status, method, path)
			ctx.metrics.ObserveSummary("response_size_bytes", resSize, status, method, path)
		}
	})
}

func requestSize(req *http.Request) float64 {
	size := 0
	if req.URL != nil {
		size = len(req.URL.Path)
	}

	size += len(req.Method)
	size += len(req.Proto)
	for name, values := range req.Header {
		size += len(name)
		for _, value := range values {
			size += len(value)
		}
	}

	size += len(req.Host)
	if req.ContentLength != -1 {
		size += int(req.ContentLength)
	}

	return float64(size)
}

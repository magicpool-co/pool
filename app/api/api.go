package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func validateChain(chain string) bool {
	switch strings.ToUpper(chain) {
	case "CFX", "CTXC", "ERGO", "ETC", "FIRO", "FLUX", "RVN":
		return true
	default:
		return false
	}
}

func New(ctx *Context, port int) *http.Server {
	mw := []middleware{
		corsMiddleware,
		recoveryMiddleware,
		rateLimiterMiddleware,
		metricsMiddleware,
	}

	// apply middleware to the router
	router := newRouter(ctx)
	for _, f := range mw {
		router = f(ctx, router)
	}

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        router,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return server
}

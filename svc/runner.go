package svc

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/magicpool-co/pool/internal/log"
)

type HTTPServer interface {
	ListenAndServe() error
	RegisterOnShutdown(func())
	Shutdown(context.Context) error
}

type TCPServer interface {
	Serve()
	Stop()
}

type WorkerServer interface {
	Start()
	Stop()
}

type Runner struct {
	logger        *log.Logger
	done          chan struct{}
	httpServers   []HTTPServer
	tcpServers    []TCPServer
	workerServers []WorkerServer
}

func NewRunner(logger *log.Logger) *Runner {
	runner := &Runner{
		logger: logger,

		done: make(chan struct{}),

		httpServers:   make([]HTTPServer, 0),
		tcpServers:    make([]TCPServer, 0),
		workerServers: make([]WorkerServer, 0),
	}

	return runner
}

func (r *Runner) AddHTTPServer(server HTTPServer) {
	r.httpServers = append(r.httpServers, server)
}

func (r *Runner) AddTCPServer(server TCPServer) {
	r.tcpServers = append(r.tcpServers, server)
}

func (r *Runner) AddWorker(server WorkerServer) {
	r.workerServers = append(r.workerServers, server)
}

// https://husobee.github.io/golang/ecs/2016/05/19/ecs-graceful-go-shutdown.html
func (r *Runner) Run() {
	var returnCode = make(chan int)
	var finishUP = make(chan struct{})

	signal.Notify(r.logger.ExitChan, syscall.SIGTERM)
	signal.Notify(r.logger.ExitChan, syscall.SIGINT)

	go func() {
		<-r.logger.ExitChan
		r.logger.Debug("notified of graceful stop request")

		finishUP <- struct{}{}

		select {
		case <-time.After(time.Minute):
			r.logger.Error(fmt.Errorf("exiting with status 1"))
			returnCode <- 1
		case <-r.done:
			r.logger.Debug("exiting with status 0")
			returnCode <- 0
		}
	}()

	go r.start()

	<-finishUP

	go r.stop()

	os.Exit(<-returnCode)
}

func (r *Runner) start() {
	for _, httpServer := range r.httpServers {
		go httpServer.ListenAndServe()
		httpServer.RegisterOnShutdown(func() {
			signal.Notify(r.logger.ExitChan, syscall.SIGTERM)
		})
	}

	for _, tcpServer := range r.tcpServers {
		go tcpServer.Serve()
	}

	for _, workerServer := range r.workerServers {
		workerServer.Start()
	}
}

func (r *Runner) stop() {
	for _, httpServer := range r.httpServers {
		if err := httpServer.Shutdown(context.Background()); err != nil {
			os.Exit(1)
		}
	}

	for _, tcpServer := range r.tcpServers {
		tcpServer.Stop()
	}

	for _, workerServer := range r.workerServers {
		workerServer.Stop()
	}

	r.done <- struct{}{}
}

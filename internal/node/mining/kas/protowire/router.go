package protowire

import (
	"sync"
)

const (
	defaultRouterCapacity = 50
)

type route struct {
	name     string
	capacity int
	ch       chan *KaspadMessage
	mu       sync.Mutex
	closed   bool
}

func newRoute(name string, capacity int) *route {
	rte := &route{
		name:     name,
		capacity: capacity,
		ch:       make(chan *KaspadMessage, capacity),
		closed:   false,
	}

	return rte
}

func (rte *route) enqueue(msg *KaspadMessage) error {
	rte.mu.Lock()
	defer rte.mu.Unlock()

	if rte.closed {
		return ErrRouteClosed
	} else if len(rte.ch) == rte.capacity {
		return ErrRouteAtCapacity
	}

	rte.ch <- msg

	return nil
}

func (rte *route) close() {
	rte.mu.Lock()
	defer rte.mu.Unlock()

	if !rte.closed {
		rte.closed = true
		close(rte.ch)
	}
}

type router struct {
	incoming map[string]*route
	outgoing *route
}

func newRouter() *router {
	rtr := &router{
		incoming: make(map[string]*route),
		outgoing: newRoute("outgoing", defaultRouterCapacity),
	}

	for _, cmd := range Cmds {
		rtr.incoming[cmd] = newRoute(cmd, defaultRouterCapacity)
	}

	return rtr
}

func (rtr *router) closeAll() {
	for _, rte := range rtr.incoming {
		rte.close()
	}

	rtr.outgoing.close()
}

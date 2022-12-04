package main

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"sync"
)

type server struct {
	listener net.Listener
	counter  uint64
	mu       sync.RWMutex
	clients  map[uint64]net.Conn
	msgs     chan []byte
	errs     chan error
}

func newServer(port int) (*server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &server{
		listener: listener,
		clients:  make(map[uint64]net.Conn),
		msgs:     make(chan []byte),
		errs:     make(chan error),
	}

	return s, nil
}

func (s *server) incrementCounter() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	if s.counter == math.MaxUint64 {
		s.counter = 0
	}

	return s.counter
}

func (s *server) broadcast(msg []byte) []error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	errs := make([]error, 0)
	for _, conn := range s.clients {
		if _, err := conn.Write(append(msg, '\n')); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (s *server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.errs <- err
		}

		if conn == nil {
			continue
		}

		go func() {
			id := s.incrementCounter()
			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.clients, id)
				conn.Close()
			}()

			s.mu.Lock()
			s.clients[id] = conn
			s.mu.Unlock()

			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				s.msgs <- scanner.Bytes()
			}
		}()
	}
}

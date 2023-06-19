package stratum

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

var (
	ErrConnNotFound = fmt.Errorf("conn not found")
)

type Message struct {
	Conn *Conn
	Req  *rpc.Request
}

type Server struct {
	ctx       context.Context
	logger    *log.Logger
	addrs     []*net.TCPAddr
	listeners []net.Listener
	wg        sync.WaitGroup
	mu        sync.RWMutex
	counter   uint64
	conns     map[uint64]*Conn
}

func NewServer(ctx context.Context, logger *log.Logger, ports ...int) (*Server, error) {
	sort.Ints(ports)
	if len(ports) == 0 {
		return nil, fmt.Errorf("no ports defined")
	}

	addrs := make([]*net.TCPAddr, len(ports))
	for i, port := range ports {
		var err error
		addrs[i], err = net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}
		logger.Info(fmt.Sprintf("listening on %s", addrs[i]))
	}

	server := &Server{
		ctx:       ctx,
		logger:    logger,
		addrs:     addrs,
		listeners: make([]net.Listener, len(addrs)),
		conns:     make(map[uint64]*Conn),
	}

	return server, nil
}

func (s *Server) Port(idx int) int {
	if idx > len(s.addrs) {
		return -1
	}

	return s.addrs[idx].Port
}

func (s *Server) newConn(rawConn net.Conn, port int) *Conn {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	if s.counter == math.MaxUint64 {
		s.counter = 0
	}

	var ip string
	if addr, ok := rawConn.RemoteAddr().(*net.TCPAddr); ok {
		ip = addr.IP.String()
	}

	conn := NewConn(s.counter, port, ip, rawConn)
	s.conns[conn.id] = conn

	return conn
}

func (s *Server) GetConn(id uint64) (*Conn, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if conn, ok := s.conns[id]; ok {
		return conn, nil
	}

	return nil, ErrConnNotFound
}

func (s *Server) Start(connTimeout time.Duration) (chan Message, chan uint64, chan uint64, chan error, error) {
	messageCh := make(chan Message)
	connectCh := make(chan uint64)
	disconnectCh := make(chan uint64)
	errCh := make(chan error)

	go func() {
		defer recoverPanic(errCh)
		<-s.ctx.Done()

		go func() {
			defer recoverPanic(errCh)

			s.close()
		}()
	}()

	for i := range s.addrs {
		var err error
		s.listeners[i], err = net.ListenTCP("tcp", s.addrs[i])
		if err != nil {
			return nil, nil, nil, nil, err
		}
		s.addrs[i] = s.listeners[i].Addr().(*net.TCPAddr)

		go func(listener net.Listener, port int) {
			defer recoverPanic(errCh)

			for {
				rawConn, err := listener.Accept()
				if err != nil {
					select {
					case <-s.ctx.Done():
						return
					default:
						if !os.IsTimeout(err) {
							errCh <- err
						}
						continue
					}
				}

				go func() {
					defer recoverPanic(errCh)
					s.wg.Add(1)

					c := s.newConn(rawConn, port)
					defer c.SoftClose()

					c.SetReadDeadline(time.Now().Add(connTimeout))
					connectCh <- c.id

					go func() {
						<-c.quit

						c.Close()
						s.wg.Done()
						disconnectCh <- c.id

						s.mu.Lock()
						defer s.mu.Unlock()
						delete(s.conns, c.id)
					}()

					scanner := c.NewScanner()
					for scanner.Scan() {
						var req *rpc.Request
						msg := scanner.Bytes()
						s.logger.Debug("recieved stratum request: " + string(msg))
						if err := json.Unmarshal(msg, &req); err == nil {
							messageCh <- Message{Conn: c, Req: req}
						}
					}
				}()
			}
		}(s.listeners[i], s.addrs[i].Port)
	}

	return messageCh, connectCh, disconnectCh, errCh, nil
}

// graceful shutdown
func (s *Server) close() {
	const shutdownDuration = time.Second * 30
	const batchInterval = time.Millisecond * 500

	for _, listener := range s.listeners {
		listener.Close()
	}

	s.mu.Lock()
	conns := s.conns
	defer s.mu.Unlock()

	count := len(conns)
	batchSize := int(shutdownDuration / batchInterval)
	if count < batchSize {
		batchSize = count
	}

	ticker := time.NewTicker(batchInterval)
	for {
		select {
		case <-ticker.C:
			var killed int
			toKill := count / batchSize
			for id, conn := range conns {
				if killed >= toKill {
					break
				}

				conn.SoftClose()
				delete(conns, id)
				killed++
			}

			if len(conns) == 0 {
				return
			}
		}
	}
}

func (s *Server) Wait() {
	s.wg.Wait()
}

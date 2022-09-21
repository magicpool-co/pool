package stratum

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"github.com/goccy/go-json"

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
	ctx     context.Context
	addr    *net.TCPAddr
	mu      sync.RWMutex
	counter uint64
	conns   map[uint64]*Conn
}

func NewServer(ctx context.Context, port int) (*Server, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		ctx:   ctx,
		addr:  addr,
		conns: make(map[uint64]*Conn),
	}

	return server, nil
}

func (s *Server) Port() int {
	return s.addr.Port
}

func (s *Server) newConn(rawConn net.Conn) *Conn {
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

	conn := &Conn{
		id:   s.counter,
		ip:   ip,
		conn: rawConn,
	}
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
	listener, err := net.ListenTCP("tcp", s.addr)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	s.addr = listener.Addr().(*net.TCPAddr)

	go func() {
		defer recoverPanic(errCh)

		for {
			rawConn, err := listener.Accept()
			select {
			case <-s.ctx.Done():
				listener.Close()
				return
			default:
				if err != nil {
					if !os.IsTimeout(err) {
						errCh <- err
					}
					continue
				}

				go func() {
					defer recoverPanic(errCh)

					c := s.newConn(rawConn)
					c.SetReadDeadline(time.Now().Add(connTimeout))
					connectCh <- c.id
					defer func() {
						c.Close()
						disconnectCh <- c.id
					}()

					scanner := c.NewScanner()
					for scanner.Scan() {
						var req *rpc.Request
						if err := json.Unmarshal(scanner.Bytes(), &req); err == nil {
							messageCh <- Message{Conn: c, Req: req}
						}
					}
				}()
			}
		}
	}()

	return messageCh, connectCh, disconnectCh, errCh, nil
}

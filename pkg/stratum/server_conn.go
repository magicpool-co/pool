package stratum

import (
	"bufio"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

type Conn struct {
	id     uint64
	port   int
	ip     string
	conn   net.Conn
	quit   chan struct{}
	closed uint32

	minerID              uint64
	workerID             uint64
	compoundID           *atomic.Value
	username             *atomic.Value
	extraNonce           *atomic.Value
	extraNonceSubscribed uint32
	subscribed           uint32
	authorized           uint32
	clientType           int32
	diffFactor           int32
	lastErrorAt          int64
	errorCount           int32
}

func storeBool(ptr *uint32, val bool) {
	var num uint32
	if val {
		num = 1
	}

	atomic.StoreUint32(ptr, num)
}

func loadBool(ptr *uint32) bool {
	return atomic.LoadUint32(ptr) != 0
}

func loadString(val *atomic.Value) string {
	raw := val.Load()
	str, ok := raw.(string)
	if ok {
		return str
	}

	return ""
}

func NewConn(id uint64, port int, ip string, rawConn net.Conn) *Conn {
	conn := &Conn{
		id:   id,
		port: port,
		ip:   ip,
		conn: rawConn,
		quit: make(chan struct{}),

		compoundID: new(atomic.Value),
		username:   new(atomic.Value),
		extraNonce: new(atomic.Value),
	}

	return conn
}

func (c *Conn) GetID() uint64                 { return c.id }
func (c *Conn) GetIP() string                 { return c.ip }
func (c *Conn) GetPort() int                  { return c.port }
func (c *Conn) GetMinerID() uint64            { return atomic.LoadUint64(&(c.minerID)) }
func (c *Conn) GetWorkerID() uint64           { return atomic.LoadUint64(&(c.workerID)) }
func (c *Conn) GetCompoundID() string         { return loadString(c.compoundID) }
func (c *Conn) GetUsername() string           { return loadString(c.username) }
func (c *Conn) GetExtraNonce() string         { return loadString(c.extraNonce) }
func (c *Conn) GetExtraNonceSubscribed() bool { return loadBool(&(c.extraNonceSubscribed)) }
func (c *Conn) GetSubscribed() bool           { return loadBool(&(c.subscribed)) }
func (c *Conn) GetAuthorized() bool           { return loadBool(&(c.authorized)) }
func (c *Conn) GetClientType() int            { return int(atomic.LoadInt32(&(c.clientType))) }
func (c *Conn) GetDiffFactor() int            { return int(atomic.LoadInt32(&(c.diffFactor))) }
func (c *Conn) GetLastErrorAt() time.Time     { return time.Unix(atomic.LoadInt64(&c.lastErrorAt), 0) }
func (c *Conn) GetErrorCount() int            { return int(atomic.LoadInt32(&c.errorCount)) }

func (c *Conn) resetCompoundID() {
	minerID := strconv.FormatUint(atomic.LoadUint64(&(c.minerID)), 10)
	workerID := strconv.FormatUint(atomic.LoadUint64(&(c.workerID)), 10)
	c.compoundID.Store(minerID + ":" + workerID)
}

func (c *Conn) SetMinerID(minerID uint64) {
	atomic.StoreUint64(&(c.minerID), minerID)
	c.resetCompoundID()
}
func (c *Conn) SetWorkerID(workerID uint64) {
	atomic.StoreUint64(&(c.workerID), workerID)
	c.resetCompoundID()
}
func (c *Conn) SetUsername(username string)     { c.username.Store(username) }
func (c *Conn) SetExtraNonce(extraNonce string) { c.extraNonce.Store(extraNonce) }
func (c *Conn) SetExtraNonceSubscribed(extraNonceSubscribed bool) {
	storeBool(&(c.extraNonceSubscribed), extraNonceSubscribed)
}
func (c *Conn) SetSubscribed(subscribed bool) { storeBool(&(c.subscribed), subscribed) }
func (c *Conn) SetAuthorized(authorized bool) { storeBool(&(c.authorized), authorized) }
func (c *Conn) SetClientType(clientType int)  { atomic.StoreInt32(&(c.clientType), int32(clientType)) }
func (c *Conn) SetDiffFactor(diffFactor int)  { atomic.StoreInt32(&(c.diffFactor), int32(diffFactor)) }
func (c *Conn) SetLastErrorAt(ts time.Time)   { atomic.StoreInt64(&(c.lastErrorAt), ts.Unix()) }
func (c *Conn) SetErrorCount(count int)       { atomic.StoreInt32(&(c.errorCount), int32(count)) }

func (c *Conn) GetLatency() (time.Duration, error) {
	return getLatency(c.conn)
}

func (c *Conn) Write(data []byte) error {
	_, err := c.conn.Write(append(data, '\n'))
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) NewScanner() *bufio.Scanner {
	return bufio.NewScanner(c.conn)
}

func (c *Conn) SetReadDeadline(timestamp time.Time) {
	c.conn.SetReadDeadline(timestamp)
}

func (c *Conn) Close() {
	c.conn.Close()
}

func (c *Conn) SoftClose() {
	if atomic.CompareAndSwapUint32(&(c.closed), 0, 1) {
		close(c.quit)
	}
}

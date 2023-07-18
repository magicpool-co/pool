package stratum

import (
	"bufio"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

type Conn struct {
	id      uint64
	port    int
	ip      string
	conn    net.Conn
	varDiff *varDiffManager
	quit    chan struct{}
	closed  uint32

	miner                *atomic.Value
	minerID              uint64
	worker               *atomic.Value
	workerID             uint64
	compoundID           *atomic.Value
	extraNonce           *atomic.Value
	extraNonceSubscribed uint32
	subscribed           uint32
	authorized           uint32
	isSolo               uint32
	client               *atomic.Value
	clientType           int32
	diffFactor           int32
	lastDiffFactor       int32
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

func NewConn(id uint64, port int, ip string, enableVarDiff bool, rawConn net.Conn) *Conn {
	var varDiff *varDiffManager
	if enableVarDiff {
		varDiff = newVarDiffManager(0)
	}

	conn := &Conn{
		id:      id,
		port:    port,
		ip:      ip,
		conn:    rawConn,
		varDiff: varDiff,
		quit:    make(chan struct{}),

		miner:      new(atomic.Value),
		worker:     new(atomic.Value),
		compoundID: new(atomic.Value),
		extraNonce: new(atomic.Value),
		client:     new(atomic.Value),
	}

	return conn
}

func (c *Conn) GetID() uint64                      { return c.id }
func (c *Conn) GetIP() string                      { return c.ip }
func (c *Conn) GetPort() int                       { return c.port }
func (c *Conn) GetMiner() string                   { return loadString(c.miner) }
func (c *Conn) GetMinerID() uint64                 { return atomic.LoadUint64(&(c.minerID)) }
func (c *Conn) GetWorker() string                  { return loadString(c.worker) }
func (c *Conn) GetWorkerID() uint64                { return atomic.LoadUint64(&(c.workerID)) }
func (c *Conn) GetCompoundID() string              { return loadString(c.compoundID) }
func (c *Conn) GetExtraNonce() string              { return loadString(c.extraNonce) }
func (c *Conn) GetExtraNonceSubscribed() bool      { return loadBool(&(c.extraNonceSubscribed)) }
func (c *Conn) GetSubscribed() bool                { return loadBool(&(c.subscribed)) }
func (c *Conn) GetAuthorized() bool                { return loadBool(&(c.authorized)) }
func (c *Conn) GetIsSolo() bool                    { return loadBool(&(c.isSolo)) }
func (c *Conn) GetClient() string                  { return loadString(c.client) }
func (c *Conn) GetClientType() int                 { return int(atomic.LoadInt32(&(c.clientType))) }
func (c *Conn) GetDiffFactor() int                 { return int(atomic.LoadInt32(&(c.diffFactor))) }
func (c *Conn) GetLastDiffFactor() int             { return int(atomic.LoadInt32(&(c.lastDiffFactor))) }
func (c *Conn) GetLastErrorAt() time.Time          { return time.Unix(atomic.LoadInt64(&c.lastErrorAt), 0) }
func (c *Conn) GetErrorCount() int                 { return int(atomic.LoadInt32(&c.errorCount)) }
func (c *Conn) GetLatency() (time.Duration, error) { return getLatency(c.conn) }

func (c *Conn) resetCompoundID() {
	minerID := strconv.FormatUint(atomic.LoadUint64(&(c.minerID)), 10)
	workerID := strconv.FormatUint(atomic.LoadUint64(&(c.workerID)), 10)
	c.compoundID.Store(minerID + ":" + workerID)
}

func (c *Conn) SetMiner(miner string) { c.miner.Store(miner) }
func (c *Conn) SetMinerID(minerID uint64) {
	atomic.StoreUint64(&(c.minerID), minerID)
	c.resetCompoundID()
}
func (c *Conn) SetWorker(worker string) { c.worker.Store(worker) }
func (c *Conn) SetWorkerID(workerID uint64) {
	atomic.StoreUint64(&(c.workerID), workerID)
	c.resetCompoundID()
}
func (c *Conn) SetExtraNonce(extraNonce string) { c.extraNonce.Store(extraNonce) }
func (c *Conn) SetExtraNonceSubscribed(extraNonceSubscribed bool) {
	storeBool(&(c.extraNonceSubscribed), extraNonceSubscribed)
}
func (c *Conn) SetSubscribed(subscribed bool) { storeBool(&(c.subscribed), subscribed) }
func (c *Conn) SetAuthorized(authorized bool) { storeBool(&(c.authorized), authorized) }
func (c *Conn) SetIsSolo(isSolo bool)         { storeBool(&(c.isSolo), isSolo) }
func (c *Conn) SetClient(client string)       { c.client.Store(client) }
func (c *Conn) SetClientType(clientType int)  { atomic.StoreInt32(&(c.clientType), int32(clientType)) }
func (c *Conn) SetDiffFactor(diffFactor int) {
	if c.varDiff != nil {
		c.varDiff.SetCurrentDiff(diffFactor, c.GetDiffFactor() == 0)
	}
	lastDiffFactor := atomic.SwapInt32(&(c.diffFactor), int32(diffFactor))
	atomic.StoreInt32(&(c.lastDiffFactor), lastDiffFactor)

}
func (c *Conn) SetLastErrorAt(ts time.Time) { atomic.StoreInt64(&(c.lastErrorAt), ts.Unix()) }
func (c *Conn) SetErrorCount(count int)     { atomic.StoreInt32(&(c.errorCount), int32(count)) }
func (c *Conn) SetLastShareAt(ts time.Time) int {
	if c.varDiff != nil {
		newDiffFactor := c.varDiff.Retarget(ts)
		if newDiffFactor != c.GetDiffFactor() {
			return newDiffFactor
		}
	}

	return -1
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

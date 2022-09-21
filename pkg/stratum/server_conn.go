package stratum

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/goccy/go-json"
)

type Conn struct {
	id         uint64
	ip         string
	conn       net.Conn
	minerID    uint64
	workerID   uint64
	compoundID string
	username   string
	extraNonce string
	subscribed bool
	authorized bool
}

func (c *Conn) GetID() uint64         { return c.id }
func (c *Conn) GetIP() string         { return c.ip }
func (c *Conn) GetMinerID() uint64    { return c.minerID }
func (c *Conn) GetWorkerID() uint64   { return c.workerID }
func (c *Conn) GetCompoundID() string { return c.compoundID }
func (c *Conn) GetUsername() string   { return c.username }
func (c *Conn) GetExtraNonce() string { return c.extraNonce }
func (c *Conn) GetSubscribed() bool   { return c.subscribed }
func (c *Conn) GetAuthorized() bool   { return c.authorized }

func (c *Conn) resetCompoundID()                { c.compoundID = fmt.Sprintf("%d:%d", c.minerID, c.workerID) }
func (c *Conn) SetMinerID(minerID uint64)       { c.minerID = minerID; c.resetCompoundID() }
func (c *Conn) SetWorkerID(workerID uint64)     { c.workerID = workerID; c.resetCompoundID() }
func (c *Conn) SetUsername(username string)     { c.username = username }
func (c *Conn) SetExtraNonce(extraNonce string) { c.extraNonce = extraNonce }
func (c *Conn) SetSubscribed(subscribed bool)   { c.subscribed = subscribed }
func (c *Conn) SetAuthorized(authorized bool)   { c.authorized = authorized }

func (c *Conn) Write(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return c.WriteRaw(data)
}

func (c *Conn) WriteRaw(data []byte) error {
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

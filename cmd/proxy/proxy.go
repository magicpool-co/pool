package main

import (
	"bufio"
	"net"
)

type proxy struct {
	url  string
	conn net.Conn
}

func newProxy(url string) (*proxy, error) {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		return nil, err
	}

	p := &proxy{
		url:  url,
		conn: conn,
	}

	return p, nil
}

func (p *proxy) close() error {
	return p.conn.Close()
}

func (p *proxy) recieve() (chan []byte, chan error) {
	msgs := make(chan []byte)
	errs := make(chan error)

	go func() {
		scanner := bufio.NewScanner(p.conn)
		for scanner.Scan() {
			msgs <- scanner.Bytes()
		}
	}()

	return msgs, errs
}

func (p *proxy) send(msg []byte) error {
	msg = append(msg, byte('\n'))
	_, err := p.conn.Write(msg)

	return err
}

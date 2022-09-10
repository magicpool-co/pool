package sshtunnel

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"

	"golang.org/x/crypto/ssh"
)

/* helpers */

func PrivateKeyFile(file, password string) (ssh.AuthMethod, error) {
	var method ssh.AuthMethod
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return method, err
	}

	var signer ssh.Signer
	if password == "" {
		signer, err = ssh.ParsePrivateKey(buffer)
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(password))
	}
	if err != nil {
		return method, err
	}

	return ssh.PublicKeys(signer), nil
}

/* tunnel */

type SSHTunnel struct {
	conn *ssh.Client
}

func New(user, host string, auth ssh.AuthMethod) (*SSHTunnel, error) {
	cfg := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{auth},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	tunnelConn, err := ssh.Dial("tcp", host, cfg)
	if err != nil {
		return nil, err
	}

	tunnel := &SSHTunnel{
		conn: tunnelConn,
	}

	return tunnel, nil
}

func (tunnel *SSHTunnel) AddDestination(dest string) (string, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	port := listener.Addr().(*net.TCPAddr).Port
	host := fmt.Sprintf("http://localhost:%d", port)

	go func() {
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			err = tunnel.forward(conn, dest)
			if err != nil {
				return
			}
		}
	}()

	return host, nil
}

func (tunnel *SSHTunnel) forward(localConn net.Conn, dest string) error {
	remoteConn, err := tunnel.conn.Dial("tcp", dest)
	if err != nil {
		return err
	}

	go io.Copy(localConn, remoteConn)
	go io.Copy(remoteConn, localConn)

	return nil
}

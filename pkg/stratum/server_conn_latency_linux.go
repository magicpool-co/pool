//go:build linux

package stratum

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/sys/unix"
)

func getLatency(conn net.Conn) (time.Duration, error) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return 0, fmt.Errorf("unable to cast as *net.TCPConn")
	}

	raw, err := tcpConn.SyscallConn()
	if err != nil {
		return 0, err
	}

	var info *unix.TCPInfo
	ctrlErr := raw.Control(func(fd uintptr) {
		info, err = unix.GetsockoptTCPInfo(int(fd), unix.IPPROTO_TCP, unix.TCP_INFO)
	})

	switch {
	case ctrlErr != nil:
		return 0, ctrlErr
	case err != nil:
		return 0, err
	}

	return time.Duration(info.Rtt), nil
}

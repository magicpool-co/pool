//go:build !linux

package stratum

import (
	"fmt"
	"net"
	"time"
)

func getLatency(conn net.Conn) (time.Duration, error) {
	return 0, fmt.Errorf("platform is not linux")
}

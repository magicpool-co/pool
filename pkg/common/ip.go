package common

import (
	"fmt"
	"net"
	"strings"
)

var (
	ipv4Mask = net.IPv4Mask(0, 0, 0, 255)
	ipv6Mask = net.IPMask{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0xff, 0xff}

	ErrInvalidIP = fmt.Errorf("invalid ip address")
)

func obscureIPv4(ip net.IP) string {
	obscured := ip.Mask(ipv4Mask).String()
	return strings.ReplaceAll(obscured, "0.0.0.", "*.*.*.")
}

func obscureIPv6(ip net.IP) string {
	obscured := ip.Mask(ipv6Mask).String()
	parts := strings.Split(obscured, ":")
	return "*:*:*:*:*:" + parts[len(parts)-1]
}

func ObscureIP(rawIP string) (string, error) {
	ip := net.ParseIP(rawIP)
	if ip == nil {
		return "", ErrInvalidIP
	}

	for i := 0; i < len(rawIP); i++ {
		switch rawIP[i] {
		case '.':
			return obscureIPv4(ip), nil
		case ':':
			return obscureIPv6(ip), nil
		}
	}

	return "", ErrInvalidIP
}

package common

import (
	"testing"
)

func TestObscureIP(t *testing.T) {
	tests := []struct {
		rawIP      string
		obscuredIP string
	}{
		{
			rawIP:      "1.1.1.1",
			obscuredIP: "*.*.*.1",
		},
		{
			rawIP:      "192.168.3.129",
			obscuredIP: "*.*.*.129",
		},
		{
			rawIP:      "192.168.3.0",
			obscuredIP: "*.*.*.0",
		},
		{
			rawIP:      "0.0.0.0",
			obscuredIP: "*.*.*.0",
		},
		{
			rawIP:      "2a00:1450:4001:820::200e",
			obscuredIP: "*:*:*:*:*:200e",
		},
		{
			rawIP:      "2a00:1450:4001:820::",
			obscuredIP: "*:*:*:*:*:",
		},
		{
			rawIP:      "2a00::",
			obscuredIP: "*:*:*:*:*:",
		},
	}

	for i, tt := range tests {
		obscuredIP, err := ObscureIP(tt.rawIP)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if obscuredIP != tt.obscuredIP {
			t.Errorf("failed on %d: have %s, want %s", i, obscuredIP, tt.obscuredIP)
		}
	}
}

package dbcl

import (
	"testing"
	"time"
)

func TestConvertDurationToInterval(t *testing.T) {
	tests := []struct {
		duration time.Duration
		interval string
	}{
		{
			duration: time.Hour*3*24*30 + time.Second*32,
			interval: "INTERVAL 3 MONTH",
		},
		{
			duration: time.Hour*3*24 + time.Hour,
			interval: "INTERVAL 3 DAY",
		},
		{
			duration: time.Hour * 3,
			interval: "INTERVAL 3 HOUR",
		},
		{
			duration: time.Hour*3 + time.Second*9,
			interval: "INTERVAL 3 HOUR",
		},
		{
			duration: time.Minute*9 + time.Second*9,
			interval: "INTERVAL 9 MINUTE",
		},
		{
			duration: time.Second * 9,
			interval: "INTERVAL 9 SECOND",
		},
		{
			duration: 0,
			interval: "INTERVAL 1 SECOND",
		},
	}

	for i, tt := range tests {
		interval := ConvertDurationToInterval(tt.duration)
		if interval != tt.interval {
			t.Errorf("failed on %d: have %s, want %s", i, interval, tt.interval)
		}
	}
}

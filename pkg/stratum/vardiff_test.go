package stratum

import (
	"reflect"
	"testing"
	"time"
)

func TestRingBuffer(t *testing.T) {
	tests := []struct {
		size  int
		items []int64
		avgs  []int64
	}{
		{
			size:  5,
			items: []int64{1, 2, 3, 4, 5},
			avgs:  []int64{1, 1, 2, 2, 3},
		},
	}

	for i, tt := range tests {
		buf := newRingBuffer(tt.size)
		for j, item := range tt.items {
			buf.Append(item)
			if !reflect.DeepEqual(buf.Items(), tt.items[:j+1]) {
				t.Errorf("failed on %d: items: %d: have %v, want %v",
					i, j, buf.Items(), tt.items[:j+1])
			} else if buf.Average() != tt.avgs[j] {
				t.Errorf("failed on %d: avg: %d: have %d, want %d",
					i, j, buf.Average(), tt.avgs[j])
			}
		}
	}
}

func TestVarDiffManager(t *testing.T) {
	tests := []struct {
		startDiff     int
		lastShares    []time.Time
		lastRetargets []time.Time
		newDiffs      []int
	}{
		{
			startDiff: 4,
			lastShares: []time.Time{
				time.Now().Add(time.Second * -15),
				time.Now().Add(time.Second * -20),
				time.Now().Add(time.Second * -25),
				time.Now().Add(time.Second * -30),
				time.Now().Add(time.Second * -35),
				time.Now().Add(time.Second * -40),
				time.Now().Add(time.Second * -45),
				time.Now().Add(time.Second * -50),
				time.Now().Add(time.Second * -20),
				time.Now().Add(time.Second * -18),
				time.Now().Add(time.Second * -15),
				time.Now().Add(time.Second * -18),
			},
			lastRetargets: []time.Time{
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
			},
			newDiffs: []int{4, 4, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			startDiff: 32,
			lastShares: []time.Time{
				time.Now().Add(time.Second * -15),
				time.Now().Add(time.Second * -20),
				time.Now().Add(time.Second * -25),
				time.Now().Add(time.Second * -30),
				time.Now().Add(time.Second * -35),
				time.Now().Add(time.Second * -40),
				time.Now().Add(time.Second * -45),
				time.Now().Add(time.Second * -50),
				time.Now().Add(time.Second * -20),
				time.Now().Add(time.Second * -18),
				time.Now().Add(time.Second * -15),
				time.Now().Add(time.Second * -18),
			},
			lastRetargets: []time.Time{
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -60),
			},
			newDiffs: []int{32, 32, 16, 16, 16, 8, 8, 8, 8, 8, 8, 8},
		},
	}

	for i, tt := range tests {
		mgr := newVarDiffManager(tt.startDiff)
		for j, lastShare := range tt.lastShares {
			mgr.lastRetarget = tt.lastRetargets[j]
			newDiff := mgr.Retarget(lastShare)
			mgr.SetCurrentDiff(newDiff)
			if newDiff != tt.newDiffs[j] {
				t.Errorf("failed on %d: retarget: %d: have %d, want %d",
					i, j, newDiff, tt.newDiffs[j])
			}
		}
	}
}

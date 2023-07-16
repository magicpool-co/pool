package stratum

import (
	"testing"
	"time"
)

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
				time.Now().Add(time.Second * -112),
				time.Now().Add(time.Second * -109),
				time.Now().Add(time.Second * -106),
				time.Now().Add(time.Second * -103),
				time.Now().Add(time.Second * -100),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -80),
				time.Now().Add(time.Second * -70),
				time.Now().Add(time.Second * -60),
				time.Now().Add(time.Second * -30),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * 25),
				time.Now().Add(time.Second * 50),
				time.Now().Add(time.Second * 75),
				time.Now().Add(time.Second * 100),
			},
			lastRetargets: []time.Time{
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
			},
			newDiffs: []int{4, 4, 4, 4, 4, 8, 8, 8, 8, 8, 4, 2, 1, 1, 1},
		},
		{
			startDiff: 64,
			lastShares: []time.Time{
				time.Now().Add(time.Second * -200),
				time.Now().Add(time.Second * -125),
				time.Now().Add(time.Second * -100),
				time.Now().Add(time.Second * -75),
				time.Now().Add(time.Second * -50),
				time.Now().Add(time.Second * -40),
				time.Now().Add(time.Second * -30),
				time.Now().Add(time.Second * -20),
				time.Now().Add(time.Second * -10),
				time.Now().Add(time.Second * 10),
				time.Now().Add(time.Second * 30),
				time.Now().Add(time.Second * 50),
				time.Now().Add(time.Second * 70),
				time.Now().Add(time.Second * 90),
				time.Now().Add(time.Second * 110),
			},
			lastRetargets: []time.Time{
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * 0),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
				time.Now().Add(time.Second * -90),
			},
			newDiffs: []int{64, 32, 32, 32, 32, 16, 16, 16, 16, 16, 16, 16, 16, 8, 8},
		},
	}

	for i, tt := range tests {
		mgr := newVarDiffManager(tt.startDiff)
		for j, lastShare := range tt.lastShares {
			mgr.lastRetarget = tt.lastRetargets[j]
			newDiff := mgr.Retarget(lastShare)
			mgr.SetCurrentDiff(newDiff, false)
			if newDiff != tt.newDiffs[j] {
				t.Errorf("failed on %d: retarget: %d: have %d, want %d",
					i, j, newDiff, tt.newDiffs[j])
			}
		}
	}
}

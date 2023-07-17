package stratum

import (
	"sync"
	"time"
)

const (
	targetTime    = time.Second * 10
	retargetDelay = time.Second * 60
	bufferSize    = (int(retargetDelay) / int(targetTime)) * 4

	variance      = time.Duration(float64(targetTime) * 0.6)
	minTargetTime = targetTime - variance
	maxTargetTime = targetTime + variance

	globalMinDiff   = 1
	globalMaxDiff   = 8192
	diffBoundFactor = 4
)

type diffList struct {
	size  int
	len   int
	items []int64
}

func newDiffList(size int) *diffList {
	list := &diffList{
		size:  size,
		items: make([]int64, 0),
	}

	return list
}

func (l *diffList) Items() []int64 {
	return l.items
}

func (l *diffList) Append(item int64) {
	l.items = append([]int64{item}, l.items...)
	if len(l.items) > l.size {
		l.items = l.items[:l.size]
	} else {
		l.len++
	}
}

func (l *diffList) Clear() {
	l.items = make([]int64, 0)
}

func (l *diffList) Average() int64 {
	length := len(l.items)
	if length == 0 {
		return 0
	}

	var sum int64
	for _, item := range l.items {
		sum += item
	}

	return sum / int64(length)
}

type varDiffManager struct {
	diff          int
	minDiff       int
	maxDiff       int
	lastDiff      int
	lastShare     time.Time
	lastRetarget  time.Time
	retargetCount int

	diffList *diffList
	mu       sync.Mutex
}

func floorDiff(currentDiff int) int {
	diff := currentDiff / diffBoundFactor
	if diff < globalMinDiff {
		return globalMinDiff
	}
	return diff
}

func ceilDiff(currentDiff int) int {
	diff := currentDiff * diffBoundFactor
	if diff > globalMaxDiff {
		return globalMaxDiff
	}
	return diff
}

func newVarDiffManager(currentDiff int) *varDiffManager {
	manager := &varDiffManager{
		diff:         currentDiff,
		minDiff:      floorDiff(currentDiff),
		maxDiff:      ceilDiff(currentDiff),
		lastDiff:     currentDiff,
		diffList:     newDiffList(bufferSize),
		lastShare:    time.Now(),
		lastRetarget: time.Now(),
	}

	return manager
}

func (m *varDiffManager) SetCurrentDiff(currentDiff int, shiftBounds bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastDiff = m.diff
	m.diff = currentDiff
	if shiftBounds {
		m.minDiff = floorDiff(currentDiff)
		m.maxDiff = ceilDiff(currentDiff)
	}
}

func (m *varDiffManager) Retarget(shareAt time.Time) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	timeSinceLastShare := shareAt.Sub(m.lastShare)
	if timeSinceLastShare < 0 {
		timeSinceLastShare = 0
	}
	timeSinceLastRetarget := time.Now().Sub(m.lastRetarget)
	m.lastShare = shareAt

	// add time since last share to diff list
	m.diffList.Append(int64(timeSinceLastShare))

	if timeSinceLastRetarget < retargetDelay {
		// if time since last retarget is less than the
		// retarget wait period, don't do anything
		return m.diff
	} else if m.retargetCount > 3 && m.diffList.len < 10 {
		// if there have been more than 3 retargets,
		// require at least 10 shares before retargeting
		return m.diff
	} else if m.diffList.len < 5 && timeSinceLastShare < time.Minute {
		// if the share rate is reasonable (one per minute),
		// require at least 5 shares before retargeting
		return m.diff
	}

	// fetch the average share submit time
	avg := time.Duration(m.diffList.Average())
	var newDiff int

	if avg > maxTargetTime {
		// decrease the difficulty by a factor of 2
		newDiff = m.diff / 2
		if newDiff < m.minDiff {
			newDiff = m.minDiff
		}
	} else if avg < minTargetTime {
		// increase the difficulty by a factor of 2
		newDiff = m.diff * 2
		if newDiff > m.maxDiff {
			newDiff = m.maxDiff
		}
	} else {
		return m.diff
	}

	// if trying to set to the old diff,
	// require at least 5 shares before changing back
	if m.diffList.len < 5 && newDiff == m.lastDiff {
		return m.diff
	}

	// set the retargets, clear buffer of old submit times
	m.lastRetarget = time.Now()
	m.retargetCount++
	m.diffList.Clear()

	return newDiff
}

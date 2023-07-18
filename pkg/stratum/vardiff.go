package stratum

import (
	"sync"
	"time"
)

const (
	targetTime    = time.Second * 15
	retargetDelay = time.Second * 90
	bufferSize    = (int(retargetDelay) / int(targetTime)) * 4

	variance      = time.Duration(float64(targetTime) * 0.5)
	minTargetTime = targetTime - variance
	maxTargetTime = targetTime + variance

	globalMinDiff   = 1
	globalMaxDiff   = 256
	diffBoundFactor = 8
)

type ringBuffer struct {
	size   int
	len    int
	cursor int
	full   bool
	items  []int64
}

func newRingBuffer(size int) *ringBuffer {
	b := &ringBuffer{
		size:  size,
		items: make([]int64, 0),
	}

	return b
}

func (b *ringBuffer) Append(item int64) {
	if b.full {
		b.items[b.cursor] = item
		b.cursor = (b.cursor + 1) % b.size
	} else {
		b.items = append(b.items, item)
		b.cursor++
		b.len++
		if b.len == b.size {
			b.cursor = 0
			b.full = true
		}
	}
}

func (b *ringBuffer) Clear() {
	b.items = make([]int64, 0)
	b.len = 0
	b.cursor = 0
	b.full = false
}

func (b *ringBuffer) Average() int64 {
	if b.len == 0 {
		return 0
	}

	var sum int64
	for _, item := range b.items {
		sum += item
	}

	return sum / int64(b.len)
}

type varDiffManager struct {
	diff          int
	minDiff       int
	maxDiff       int
	lastDiff      int
	lastShare     time.Time
	lastRetarget  time.Time
	retargetCount int

	ringBuffer *ringBuffer
	mu         sync.Mutex
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
	now := time.Now()
	manager := &varDiffManager{
		diff:         currentDiff,
		minDiff:      floorDiff(currentDiff),
		maxDiff:      ceilDiff(currentDiff),
		lastDiff:     currentDiff,
		ringBuffer:   newRingBuffer(bufferSize),
		lastShare:    now,
		lastRetarget: now.Add(-retargetDelay / 2),
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

	now := time.Now()
	timeSinceLastShare := shareAt.Sub(m.lastShare)
	if timeSinceLastShare < 0 {
		timeSinceLastShare = 0
	}

	m.lastShare = shareAt
	m.ringBuffer.Append(int64(timeSinceLastShare))

	timeSinceLastRetarget := now.Sub(m.lastRetarget)
	if timeSinceLastRetarget < retargetDelay {
		// if time since last retarget is less than the
		// retarget wait period, don't do anything
		return m.diff
	}
	m.lastRetarget = now

	// fetch the average share submit time
	avg := time.Duration(m.ringBuffer.Average())
	var newDiff int
	if avg > maxTargetTime && m.diff > m.minDiff {
		// decrease the difficulty by a factor of 2
		newDiff = m.diff / 2
		if newDiff < m.minDiff {
			newDiff = m.minDiff
		}
	} else if avg < minTargetTime && m.diff < m.maxDiff {
		// increase the difficulty by a factor of 2
		newDiff = m.diff * 2
		if newDiff > m.maxDiff {
			newDiff = m.maxDiff
		}
	} else {
		return m.diff
	}

	// set the retargets, clear buffer of old submit times
	m.retargetCount++
	m.ringBuffer.Clear()

	return newDiff
}

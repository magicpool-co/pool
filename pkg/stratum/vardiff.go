package stratum

import (
	"container/ring"
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

	minDiff         = 1
	maxDiff         = 512
	diffBoundFactor = 4
)

type ringBuffer struct {
	size int
	len  int
	ring *ring.Ring
}

func newRingBuffer(size int) *ringBuffer {
	buf := &ringBuffer{
		size: size,
		ring: ring.New(size),
	}

	return buf
}

func (r *ringBuffer) Items() []int64 {
	count := r.len
	item := r.ring.Move(-count)
	items := make([]int64, count)
	for i := range items {
		items[i] = item.Value.(int64)
		item = item.Next()
	}

	return items
}

func (r *ringBuffer) Append(item int64) {
	r.ring.Value = item
	r.ring = r.ring.Next()
	if r.len < r.size {
		r.len++
	}
}

func (r *ringBuffer) Clear() {
	r.ring = ring.New(r.size)
	r.len = 0
}

func (r *ringBuffer) Average() int64 {
	items := r.Items()
	if len(items) == 0 {
		return 0
	}

	var sum int64
	for _, item := range items {
		sum += item
	}

	return sum / int64(len(items))
}

type varDiffManager struct {
	diff          int
	minDiff       int
	maxDiff       int
	lastShare     time.Time
	lastRetarget  time.Time
	retargetCount int

	buffer *ringBuffer
	mu     sync.Mutex
}

func floorDiff(currentDiff int) int {
	diff := currentDiff / diffBoundFactor
	if diff < 1 {
		return 1
	}
	return diff
}

func ceilDiff(currentDiff int) int {
	diff := currentDiff * diffBoundFactor
	if diff > 512 {
		return 512
	}
	return diff
}

func newVarDiffManager(currentDiff int) *varDiffManager {
	manager := &varDiffManager{
		diff:         currentDiff,
		minDiff:      floorDiff(currentDiff),
		maxDiff:      ceilDiff(currentDiff),
		buffer:       newRingBuffer(bufferSize),
		lastShare:    time.Now(),
		lastRetarget: time.Now(),
	}

	return manager
}

func (m *varDiffManager) SetCurrentDiff(currentDiff int, shiftBounds bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

	// add time since last share to ring buffer
	m.buffer.Append(int64(timeSinceLastShare))

	if timeSinceLastRetarget < retargetDelay {
		// if time since last retarget is less than the
		// retarget wait period, don't do anything
		return m.diff
	} else if m.retargetCount > 3 && m.buffer.len < 10 {
		// if there have been more than 3 retargets,
		// require at least 10 shares before retargeting
	} else if m.buffer.len < 5 && timeSinceLastShare < time.Minute {
		// if the share rate is reasonable (one per minute),
		// require at least 5 shares before retargeting
		return m.diff
	}

	// fetch the average share submit time
	avg := time.Duration(m.buffer.Average())
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

	// set the retargets, clear buffer of old submit times
	m.lastRetarget = time.Now()
	m.retargetCount++
	m.buffer.Clear()

	return newDiff
}

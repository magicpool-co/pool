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

	variance      = time.Duration(float64(targetTime) * 0.5)
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
	diff         int
	minDiff      int
	maxDiff      int
	lastRetarget time.Time

	buffer *ringBuffer
	mu     sync.Mutex
}

func newVarDiffManager(currentDiff int) *varDiffManager {
	// set the minimum difficulty to MIN(1, diff / diffBoundFactor)
	minDiff := currentDiff / diffBoundFactor
	if minDiff < 1 {
		minDiff = 1
	}

	// set the maximum difficulty to MAX(512, diff * diffBoundFactor)
	maxDiff := currentDiff * diffBoundFactor
	if maxDiff > 512 {
		maxDiff = 512
	}

	manager := &varDiffManager{
		diff:         currentDiff,
		minDiff:      minDiff,
		maxDiff:      maxDiff,
		buffer:       newRingBuffer(bufferSize),
		lastRetarget: time.Now(),
	}

	return manager
}

func (m *varDiffManager) SetCurrentDiff(currentDiff int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.diff = currentDiff
}

func (m *varDiffManager) Retarget(lastShare time.Time) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	timeSinceLastShare := now.Sub(lastShare)
	timeSinceLastRetarget := now.Sub(m.lastRetarget)

	// add time since last share to ring buffer
	m.buffer.Append(int64(timeSinceLastShare))

	// if time since last retarget is less than the
	// retarget wait period, don't do anything
	if timeSinceLastRetarget < retargetDelay {
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
		if newDiff < m.maxDiff {
			newDiff = m.maxDiff
		}
	} else {
		return m.diff
	}

	// clear buffer of old submit times
	m.buffer.Clear()

	return newDiff
}

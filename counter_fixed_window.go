package ratelimit

import (
	"math"
	"sync"
)

// FixedWindow implements Counter using a fixed window algorithm.
// Requests are counted within discrete time windows that reset at boundaries.
type FixedWindow struct {
	mu       sync.Mutex
	limit    int
	windowMs int64
	count    int
	winStart int64 // floor(now/windowMs)*windowMs
}

// NewFixedWindow creates a fixed window counter.
func NewFixedWindow(limit, windowSeconds int) *FixedWindow {
	return &FixedWindow{
		limit:    limit,
		windowMs: int64(windowSeconds) * 1000,
		winStart: -1,
	}
}

func (f *FixedWindow) Allow(nowMs int64) (bool, int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	win := (nowMs / f.windowMs) * f.windowMs
	if win != f.winStart {
		f.winStart = win
		f.count = 0
	}

	f.count++
	if f.count <= f.limit {
		return true, f.limit - f.count, 0
	}

	msToReset := (f.winStart + f.windowMs) - nowMs
	retry := int(math.Ceil(float64(msToReset) / 1000.0))
	if retry < 1 {
		retry = 1
	}
	return false, 0, retry
}

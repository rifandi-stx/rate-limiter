package counter

import (
	"math"
	"sync"
)

// TokenBucket implements Counter using a token bucket algorithm.
// Tokens refill over time up to a maximum capacity (burst).
type TokenBucket struct {
	mu       sync.Mutex
	capacity float64 // burst
	refillMs float64 // tokens per millisecond = (limit/windowSeconds)/1000
	tokens   float64
	lastMs   int64
}

// NewTokenBucket creates a token bucket counter.
// burst is the maximum tokens (bucket capacity).
// limit is the sustained rate (tokens per windowSeconds).
// windowSeconds is the time window for the sustained rate.
// refillRate = limit / windowSeconds tokens per second.
func NewTokenBucket(burst, limit, windowSeconds int) *TokenBucket {
	return &TokenBucket{
		capacity: float64(burst),
		refillMs: (float64(limit) / float64(windowSeconds)) / 1000.0,
		tokens:   float64(burst), // start full
		lastMs:   -1,
	}
}

func (t *TokenBucket) Allow(nowMs int64) (bool, int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.lastMs < 0 {
		t.lastMs = nowMs
	}

	elapsed := float64(nowMs - t.lastMs)
	if elapsed < 0 {
		elapsed = 0
	}

	t.tokens = math.Min(t.capacity, t.tokens+elapsed*t.refillMs)
	t.lastMs = nowMs

	if t.tokens >= 1 {
		t.tokens -= 1
		return true, int(math.Floor(t.tokens)), 0
	}

	deficit := 1 - t.tokens
	retryMs := deficit / t.refillMs
	retry := int(math.Ceil(retryMs / 1000.0))
	if retry < 1 {
		retry = 1
	}
	return false, 0, retry
}

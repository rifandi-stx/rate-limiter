package ratelimit

import "testing"

func TestTokenBucket_AllowsBurstUpToCapacity(t *testing.T) {
	// RFC vector: token_bucket_allows_burst_up_to_capacity
	// limit=10, windowSeconds=10, burst=5
	// refillRate = 10/10 = 1 token/second
	counter := NewTokenBucket(5, 10, 10)

	tests := []struct {
		nowMs           int64
		expectedAllowed bool
	}{
		{0, true},
		{0, true},
		{0, true},
		{0, true},
		{0, true},
		{0, false}, // burst exhausted
	}

	for i, tt := range tests {
		allowed, _, _ := counter.Allow(tt.nowMs)
		if allowed != tt.expectedAllowed {
			t.Errorf("request %d at %dms: allowed=%v, want %v", i, tt.nowMs, allowed, tt.expectedAllowed)
		}
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	// RFC vector: token_bucket_refills_over_time
	// limit=10, windowSeconds=10, burst=2
	// refillRate = 10/10 = 1 token/second
	counter := NewTokenBucket(2, 10, 10)

	tests := []struct {
		nowMs           int64
		expectedAllowed bool
	}{
		{0, true},
		{0, true},
		{0, false},        // burst exhausted
		{1000, true},      // 1 second passed, 1 token refilled
		{1000, false},     // no more tokens yet
	}

	for i, tt := range tests {
		allowed, _, _ := counter.Allow(tt.nowMs)
		if allowed != tt.expectedAllowed {
			t.Errorf("request %d at %dms: allowed=%v, want %v", i, tt.nowMs, allowed, tt.expectedAllowed)
		}
	}
}

func TestTokenBucket_RetryAfterSeconds(t *testing.T) {
	// burst=1, limit=1 per 10 seconds = 0.1 tokens/second
	// After consuming the token, need 10 seconds to refill
	counter := NewTokenBucket(1, 1, 10)

	// First request allowed
	allowed, _, retry := counter.Allow(0)
	if !allowed || retry != 0 {
		t.Errorf("first request: allowed=%v retry=%d, want allowed=true retry=0", allowed, retry)
	}

	// Second request denied
	allowed, _, retry = counter.Allow(0)
	if allowed {
		t.Error("second request should be denied")
	}
	if retry < 1 {
		t.Errorf("retry=%d, want >= 1", retry)
	}
}

func TestTokenBucket_RemainingTokens(t *testing.T) {
	counter := NewTokenBucket(5, 10, 10)

	// Start with 5 tokens
	allowed, remaining, _ := counter.Allow(0)
	if !allowed || remaining != 4 {
		t.Errorf("got allowed=%v remaining=%d, want allowed=true remaining=4", allowed, remaining)
	}

	allowed, remaining, _ = counter.Allow(0)
	if !allowed || remaining != 3 {
		t.Errorf("got allowed=%v remaining=%d, want allowed=true remaining=3", allowed, remaining)
	}
}

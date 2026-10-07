package ratelimit

import "testing"

func TestFixedWindow_AllowsUpToLimitThenDenies(t *testing.T) {
	// RFC vector: fixed_window_allows_up_to_limit_then_denies
	counter := NewFixedWindow(3, 1)

	tests := []struct {
		nowMs             int64
		expectedAllowed   bool
		expectedRemaining int
	}{
		{0, true, 2},
		{100, true, 1},
		{200, true, 0},
		{300, false, 0},
	}

	for i, tt := range tests {
		allowed, remaining, _ := counter.Allow(tt.nowMs)
		if allowed != tt.expectedAllowed {
			t.Errorf("request %d at %dms: allowed=%v, want %v", i, tt.nowMs, allowed, tt.expectedAllowed)
		}
		if remaining != tt.expectedRemaining {
			t.Errorf("request %d at %dms: remaining=%d, want %d", i, tt.nowMs, remaining, tt.expectedRemaining)
		}
	}
}

func TestFixedWindow_ResetsNextWindow(t *testing.T) {
	// RFC vector: fixed_window_resets_next_window
	counter := NewFixedWindow(2, 1)

	tests := []struct {
		nowMs           int64
		expectedAllowed bool
	}{
		{0, true},
		{500, true},
		{600, false},
		{1000, true}, // new window
	}

	for i, tt := range tests {
		allowed, _, _ := counter.Allow(tt.nowMs)
		if allowed != tt.expectedAllowed {
			t.Errorf("request %d at %dms: allowed=%v, want %v", i, tt.nowMs, allowed, tt.expectedAllowed)
		}
	}
}

func TestFixedWindow_RetryAfterSeconds(t *testing.T) {
	counter := NewFixedWindow(1, 10) // 1 request per 10 seconds

	// First request allowed
	allowed, _, retry := counter.Allow(0)
	if !allowed || retry != 0 {
		t.Errorf("first request: allowed=%v retry=%d, want allowed=true retry=0", allowed, retry)
	}

	// Second request denied, retry should be ~10 seconds
	allowed, _, retry = counter.Allow(100)
	if allowed {
		t.Error("second request should be denied")
	}
	if retry < 1 || retry > 10 {
		t.Errorf("retry=%d, want 1-10", retry)
	}
}

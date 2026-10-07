package ratelimit

import "testing"

func TestKeyedLimiter_NoRulesAllowsByDefault(t *testing.T) {
	provider := NewDeclarativeProvider(func() []RuleSpec { return nil })
	factory := NewLocalCounterFactory()
	limiter := NewKeyedLimiter(provider, factory, "LOCAL")

	ctx := RequestContext{UserID: "123"}
	decision := limiter.Allow(ctx)

	if !decision.Allowed {
		t.Error("expected allowed when no rules match")
	}
	if decision.Matched {
		t.Error("expected matched=false when no rules")
	}
}

func TestKeyedLimiter_SingleRule(t *testing.T) {
	rules := []RuleSpec{{
		Name:             "per-user",
		Scope:            "user",
		KeyField:         "userId",
		Match:            MatchSpec{Field: "userId", Exists: boolPtr(true)},
		Limit:            2,
		WindowSeconds:    1,
		Algorithm:        "fixed_window",
		DistributionMode: "LOCAL",
	}}
	provider := NewDeclarativeProvider(func() []RuleSpec { return rules })
	factory := NewLocalCounterFactory()

	nowMs := int64(0)
	limiter := NewKeyedLimiter(provider, factory, "LOCAL", WithClock(func() int64 { return nowMs }))

	ctx := RequestContext{UserID: "123"}

	// First two requests allowed
	d1 := limiter.Allow(ctx)
	if !d1.Allowed || d1.Remaining != 1 {
		t.Errorf("request 1: allowed=%v remaining=%d, want allowed=true remaining=1", d1.Allowed, d1.Remaining)
	}

	d2 := limiter.Allow(ctx)
	if !d2.Allowed || d2.Remaining != 0 {
		t.Errorf("request 2: allowed=%v remaining=%d, want allowed=true remaining=0", d2.Allowed, d2.Remaining)
	}

	// Third request denied
	d3 := limiter.Allow(ctx)
	if d3.Allowed {
		t.Error("request 3 should be denied")
	}
	if d3.RuleName != "per-user" {
		t.Errorf("ruleName=%s, want per-user", d3.RuleName)
	}
	if d3.RetryAfterSeconds < 1 {
		t.Errorf("retryAfterSeconds=%d, want >= 1", d3.RetryAfterSeconds)
	}
}

func TestKeyedLimiter_DifferentUsers(t *testing.T) {
	rules := []RuleSpec{{
		Name:             "per-user",
		Scope:            "user",
		KeyField:         "userId",
		Match:            MatchSpec{Field: "userId", Exists: boolPtr(true)},
		Limit:            1,
		WindowSeconds:    1,
		Algorithm:        "fixed_window",
		DistributionMode: "LOCAL",
	}}
	provider := NewDeclarativeProvider(func() []RuleSpec { return rules })
	factory := NewLocalCounterFactory()
	limiter := NewKeyedLimiter(provider, factory, "LOCAL")

	// User A
	d1 := limiter.Allow(RequestContext{UserID: "A"})
	if !d1.Allowed {
		t.Error("user A first request should be allowed")
	}

	// User B has separate bucket
	d2 := limiter.Allow(RequestContext{UserID: "B"})
	if !d2.Allowed {
		t.Error("user B first request should be allowed")
	}

	// User A exhausted
	d3 := limiter.Allow(RequestContext{UserID: "A"})
	if d3.Allowed {
		t.Error("user A second request should be denied")
	}
}

func TestKeyedLimiter_TokenBucket(t *testing.T) {
	rules := []RuleSpec{{
		Name:             "per-user-burst",
		Scope:            "user",
		KeyField:         "userId",
		Match:            MatchSpec{Field: "userId", Exists: boolPtr(true)},
		Limit:            10,
		WindowSeconds:    10,
		Algorithm:        "token_bucket",
		DistributionMode: "LOCAL",
		Params:           map[string]any{"burst": 3},
	}}
	provider := NewDeclarativeProvider(func() []RuleSpec { return rules })
	factory := NewLocalCounterFactory()
	limiter := NewKeyedLimiter(provider, factory, "LOCAL")

	ctx := RequestContext{UserID: "123"}

	// Burst of 3
	for i := 0; i < 3; i++ {
		d := limiter.Allow(ctx)
		if !d.Allowed {
			t.Errorf("request %d should be allowed (burst)", i+1)
		}
	}

	// 4th request denied
	d := limiter.Allow(ctx)
	if d.Allowed {
		t.Error("request 4 should be denied (burst exhausted)")
	}
}

func TestKeyedLimiter_MaxKeysEviction(t *testing.T) {
	rules := []RuleSpec{{
		Name:             "per-ip",
		Scope:            "ip",
		KeyField:         "ip",
		Match:            MatchSpec{Always: boolPtr(true)},
		Limit:            100,
		WindowSeconds:    60,
		Algorithm:        "fixed_window",
		DistributionMode: "LOCAL",
	}}
	provider := NewDeclarativeProvider(func() []RuleSpec { return rules })
	factory := NewLocalCounterFactory()
	limiter := NewKeyedLimiter(provider, factory, "LOCAL", WithMaxKeys(2))

	// Create 3 buckets, but maxKeys=2
	limiter.Allow(RequestContext{IP: "1.1.1.1"})
	limiter.Allow(RequestContext{IP: "2.2.2.2"})
	limiter.Allow(RequestContext{IP: "3.3.3.3"})

	if limiter.BucketCount() > 2 {
		t.Errorf("bucket count=%d, want <= 2", limiter.BucketCount())
	}
}

func boolPtr(b bool) *bool {
	return &b
}

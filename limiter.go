package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	counter  Counter
	lastSeen int64 // ms, for idle eviction
}

// KeyedLimiter is the orchestrator that holds the per-key Counter map,
// asks the RuleProvider which rules apply, lazily builds Counters via
// the CounterFactory, runs the multi-bucket AND, and assembles the Decision.
type KeyedLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*entry
	provider RuleProvider
	factory  CounterFactory
	mode     string // LOCAL | REMOTE (carried into the Decision)
	maxKeys  int    // cardinality cap; 0 = unlimited
	idleMs   int64  // evict buckets unused for longer than this
	clock    func() int64
}

// KeyedLimiterOption configures a KeyedLimiter.
type KeyedLimiterOption func(*KeyedLimiter)

// WithMaxKeys sets the cardinality cap for buckets.
func WithMaxKeys(n int) KeyedLimiterOption {
	return func(k *KeyedLimiter) {
		k.maxKeys = n
	}
}

// WithIdleTTL sets the idle eviction threshold.
func WithIdleTTL(d time.Duration) KeyedLimiterOption {
	return func(k *KeyedLimiter) {
		k.idleMs = d.Milliseconds()
	}
}

// WithClock injects a clock for testing. Default is time.Now().UnixMilli().
func WithClock(clock func() int64) KeyedLimiterOption {
	return func(k *KeyedLimiter) {
		k.clock = clock
	}
}

// NewKeyedLimiter creates a new KeyedLimiter.
func NewKeyedLimiter(provider RuleProvider, factory CounterFactory, mode string, opts ...KeyedLimiterOption) *KeyedLimiter {
	k := &KeyedLimiter{
		buckets:  make(map[string]*entry),
		provider: provider,
		factory:  factory,
		mode:     mode,
		clock:    func() int64 { return time.Now().UnixMilli() },
	}
	for _, opt := range opts {
		opt(k)
	}
	return k
}

// Allow evaluates the request against all applicable rules.
// All rules must pass for the request to be allowed (AND semantics).
func (k *KeyedLimiter) Allow(ctx RequestContext) Decision {
	nowMs := k.clock()
	rules := k.provider.RulesFor(ctx)

	if len(rules) == 0 {
		return Decision{
			Allowed:          true,
			Matched:          false,
			DistributionMode: k.mode,
		}
	}

	outcomes := make([]RuleOutcome, 0, len(rules))
	allowed := true
	var firstFail *Rule
	var firstFailRem, firstFailRetry int

	// track the binding (tightest) rule for the allow case
	tightestRem := int(^uint(0) >> 1) // max int
	var binding *Rule
	var bindingRetry int

	for i := range rules {
		r := &rules[i]
		c := k.counterFor(*r, nowMs)
		ok, rem, retry := c.Allow(nowMs)
		outcomes = append(outcomes, RuleOutcome{
			RuleName:  r.Name,
			Allowed:   ok,
			Remaining: rem,
		})

		if !ok {
			allowed = false
			if firstFail == nil {
				firstFail = r
				firstFailRem, firstFailRetry = rem, retry
			}
		} else if rem < tightestRem {
			tightestRem = rem
			binding = r
			bindingRetry = retry
		}
	}

	d := Decision{
		Allowed:          allowed,
		Matched:          true,
		DistributionMode: k.mode,
		Evaluated:        outcomes,
	}

	if !allowed && firstFail != nil {
		fillDecision(&d, firstFail, firstFailRem, firstFailRetry)
	} else if allowed && binding != nil {
		fillDecision(&d, binding, tightestRem, bindingRetry)
	}

	return d
}

func fillDecision(d *Decision, r *Rule, rem, retry int) {
	d.RuleName = r.Name
	d.Scope = r.Scope
	d.Key = r.BucketKey()
	d.Limit = r.Limit
	d.WindowSeconds = r.WindowSeconds
	d.Algorithm = r.Algorithm
	d.DistributionMode = r.DistributionMode
	d.Remaining = rem
	d.RetryAfterSeconds = retry
}

func (k *KeyedLimiter) counterFor(r Rule, nowMs int64) Counter {
	key := r.BucketKey()

	k.mu.Lock()
	defer k.mu.Unlock()

	e, ok := k.buckets[key]
	if !ok {
		if k.maxKeys > 0 && len(k.buckets) >= k.maxKeys {
			k.evictOldestLocked()
		}
		e = &entry{counter: k.factory.Build(r)}
		k.buckets[key] = e
	}
	e.lastSeen = nowMs
	return e.counter
}

// Sweep removes buckets that have been idle beyond the TTL.
// Call periodically from a background goroutine.
func (k *KeyedLimiter) Sweep(nowMs int64) {
	if k.idleMs <= 0 {
		return
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	for key, e := range k.buckets {
		if nowMs-e.lastSeen > k.idleMs {
			delete(k.buckets, key)
		}
	}
}

// BucketCount returns the current number of tracked buckets.
func (k *KeyedLimiter) BucketCount() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return len(k.buckets)
}

func (k *KeyedLimiter) evictOldestLocked() {
	var oldestKey string
	var oldest int64 = 1<<63 - 1
	for key, e := range k.buckets {
		if e.lastSeen < oldest {
			oldest = e.lastSeen
			oldestKey = key
		}
	}
	if oldestKey != "" {
		delete(k.buckets, oldestKey)
	}
}

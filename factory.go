package ratelimit

import "github.com/rifandi-stx/rate-limiter/counter"

// LocalCounterFactory builds in-memory (LOCAL) counters for rules.
// It implements CounterFactory.
type LocalCounterFactory struct{}

// NewLocalCounterFactory creates a factory for LOCAL counters.
func NewLocalCounterFactory() *LocalCounterFactory {
	return &LocalCounterFactory{}
}

func (f *LocalCounterFactory) Build(rule Rule) Counter {
	switch rule.Algorithm {
	case "token_bucket":
		burst := rule.Limit // default burst = limit
		if b, ok := rule.Params["burst"].(int); ok {
			burst = b
		} else if b, ok := rule.Params["burst"].(float64); ok {
			burst = int(b)
		}
		return counter.NewTokenBucket(burst, rule.Limit, rule.WindowSeconds)

	case "fixed_window":
		return counter.NewFixedWindow(rule.Limit, rule.WindowSeconds)

	default:
		// Unknown algorithm: fall back to fixed window as safe default
		return counter.NewFixedWindow(rule.Limit, rule.WindowSeconds)
	}
}

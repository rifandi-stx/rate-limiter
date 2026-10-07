package ratelimit

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
		return NewTokenBucket(burst, rule.Limit, rule.WindowSeconds)

	case "fixed_window":
		return NewFixedWindow(rule.Limit, rule.WindowSeconds)

	default:
		// Unknown algorithm: fall back to fixed window as safe default
		return NewFixedWindow(rule.Limit, rule.WindowSeconds)
	}
}

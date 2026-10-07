package ratelimit

// Rule is one resolved rule for a request: the key to count against,
// the threshold, the algorithm, and algorithm-specific params.
type Rule struct {
	Name             string
	Scope            string // "user", "org", "ip", ...
	KeyValue         string // the concrete value, e.g. "123"
	Limit            int
	WindowSeconds    int
	Algorithm        string // fixed_window | sliding_window_log | sliding_window_counter | token_bucket | leaky_bucket
	DistributionMode string // LOCAL | REMOTE
	Params           map[string]any
}

// BucketKey is the engine's map key: scope + ":" + value.
func (r Rule) BucketKey() string {
	return r.Scope + ":" + r.KeyValue
}

// RuleProvider is AXIS 1: what to count. Returns applicable rules in
// precedence order (empty = no match = allow-by-default + log).
type RuleProvider interface {
	RulesFor(ctx RequestContext) []Rule
}

// CounterFactory is AXIS 2: how to count. Builds a Counter for a rule,
// picking LOCAL vs REMOTE per rule.DistributionMode.
type CounterFactory interface {
	Build(rule Rule) Counter
}

// Limiter is the top-level entry point consumers call per request.
type Limiter interface {
	Allow(ctx RequestContext) Decision
}

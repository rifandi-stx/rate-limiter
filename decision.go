package ratelimit

// RuleOutcome is one rule's result within a decision (for observability).
type RuleOutcome struct {
	RuleName  string `json:"ruleName"`
	Allowed   bool   `json:"allowed"`
	Remaining int    `json:"remaining"`
}

// Decision is produced once per Allow() and drives caller, logs, metrics, traces.
type Decision struct {
	// Allowed is the final decision across all applicable rules.
	Allowed bool `json:"allowed"`

	// Matched indicates whether any rule matched at all.
	// false means allowed by default; always logged.
	Matched bool `json:"matched"`

	// RuleName is the single deciding rule (never a list).
	// On deny: the first rule that failed in evaluation order.
	// On allow: the matched rule with the least remaining headroom.
	// Empty when no rule matched.
	RuleName string `json:"ruleName"`

	// Scope of the deciding rule (e.g. user, org, ip).
	Scope string `json:"scope"`

	// Key is the bucket key of the deciding rule (e.g. user:123).
	// For logs/traces only, never a metric label.
	Key string `json:"key"`

	// Limit applied by the deciding rule.
	Limit int `json:"limit"`

	// WindowSeconds applied by the deciding rule.
	WindowSeconds int `json:"windowSeconds"`

	// Algorithm of the deciding rule.
	Algorithm string `json:"algorithm"`

	// DistributionMode is the enforcement mode used: LOCAL or REMOTE.
	DistributionMode string `json:"distributionMode"`

	// Remaining is the approximate remaining allowance in the deciding bucket.
	Remaining int `json:"remaining"`

	// RetryAfterSeconds is seconds until the deciding bucket would admit
	// one more request. Drives the Retry-After header. Zero when allowed.
	RetryAfterSeconds int `json:"retryAfterSeconds"`

	// Evaluated contains per-rule outcomes for every rule that applied.
	// This is the full list for a multi-rule (AND) request.
	Evaluated []RuleOutcome `json:"evaluated"`

	// Degraded is true when the decision came from a failsafe path
	// (e.g. local_fallback after Redis failure).
	Degraded bool `json:"degraded"`
}

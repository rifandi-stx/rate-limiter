package counter

// Counter is the one-method contract every algorithm implements.
// The engine holds map[string]Counter and is blind to the concrete type.
//
// Clock convention: nowMs is milliseconds since the Unix epoch, injected,
// never read from the system clock inside algorithms.
type Counter interface {
	// Allow attempts to consume one unit at time nowMs.
	// Returns whether it was allowed, the approximate remaining allowance,
	// and how many seconds until the next unit is available (0 when allowed).
	Allow(nowMs int64) (allowed bool, remaining int, retryAfterSeconds int)
}

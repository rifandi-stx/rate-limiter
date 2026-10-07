package counter

import (
	"encoding/json"
	"os"
	"testing"
)

type testVector struct {
	Name          string         `json:"name"`
	Algorithm     string         `json:"algorithm"`
	Limit         int            `json:"limit"`
	WindowSeconds int            `json:"window_seconds"`
	Params        map[string]any `json:"params,omitempty"`
	Requests      []testRequest  `json:"requests"`
}

type testRequest struct {
	NowMs             int64 `json:"now_ms"`
	ExpectedAllowed   *bool `json:"expected_allowed,omitempty"`
	ExpectedRemaining *int  `json:"expected_remaining,omitempty"`
}

type testVectors struct {
	Vectors []testVector `json:"vectors"`
}

func TestVectors(t *testing.T) {
	data, err := os.ReadFile("../testdata/test-vectors.json")
	if err != nil {
		t.Fatalf("failed to read test vectors: %v", err)
	}

	var vectors testVectors
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatalf("failed to parse test vectors: %v", err)
	}

	for _, v := range vectors.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			c := buildCounter(t, v)
			if c == nil {
				return // skipped
			}

			for i, req := range v.Requests {
				allowed, remaining, _ := c.Allow(req.NowMs)

				if req.ExpectedAllowed != nil && allowed != *req.ExpectedAllowed {
					t.Errorf("request %d at %dms: allowed=%v, want %v",
						i, req.NowMs, allowed, *req.ExpectedAllowed)
				}
				if req.ExpectedRemaining != nil && remaining != *req.ExpectedRemaining {
					t.Errorf("request %d at %dms: remaining=%d, want %d",
						i, req.NowMs, remaining, *req.ExpectedRemaining)
				}
			}
		})
	}
}

func buildCounter(t *testing.T, v testVector) Counter {
	switch v.Algorithm {
	case "fixed_window":
		return NewFixedWindow(v.Limit, v.WindowSeconds)

	case "token_bucket":
		burst := v.Limit
		if b, ok := v.Params["burst"].(float64); ok {
			burst = int(b)
		}
		return NewTokenBucket(burst, v.Limit, v.WindowSeconds)

	case "sliding_window_counter", "sliding_window_log", "leaky_bucket":
		t.Skipf("algorithm %s not implemented yet", v.Algorithm)
		return nil

	default:
		t.Fatalf("unknown algorithm: %s", v.Algorithm)
		return nil
	}
}

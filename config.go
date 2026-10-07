package ratelimit

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadConfigFile reads and validates a config from a JSON file.
// This is the v1 approach; AppConfig polling is wired later.
func LoadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// StaticProvider wraps a fixed Config for use as a RuleProvider source.
// Use with NewDeclarativeProvider for static file-based config.
func StaticProvider(cfg *Config) func() []RuleSpec {
	return func() []RuleSpec {
		return cfg.Rules
	}
}

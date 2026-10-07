package ratelimit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFile_Valid(t *testing.T) {
	cfg, err := LoadConfigFile("testdata/sample-config.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(cfg.Rules))
	}

	if cfg.Rules[0].Name != "default-per-user" {
		t.Errorf("expected first rule name 'default-per-user', got %q", cfg.Rules[0].Name)
	}

	if cfg.Rules[1].Algorithm != "fixed_window" {
		t.Errorf("expected second rule algorithm 'fixed_window', got %q", cfg.Rules[1].Algorithm)
	}
}

func TestLoadConfigFile_NotFound(t *testing.T) {
	_, err := LoadConfigFile("testdata/nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadConfigFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")
	os.WriteFile(path, []byte("not json"), 0644)

	_, err := LoadConfigFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadConfigFile_ValidationFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid-config.json")
	os.WriteFile(path, []byte(`{"rules": [{"name": "test"}]}`), 0644)

	_, err := LoadConfigFile(path)
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestStaticProvider(t *testing.T) {
	cfg, err := LoadConfigFile("testdata/sample-config.json")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	provider := NewDeclarativeProvider(StaticProvider(cfg))

	// Test with a user context - should match first rule
	ctx := RequestContext{UserID: "123", IP: "1.2.3.4"}
	rules := provider.RulesFor(ctx)

	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	if rules[0].Name != "default-per-user" {
		t.Errorf("expected rule 'default-per-user', got %q", rules[0].Name)
	}

	if rules[0].KeyValue != "123" {
		t.Errorf("expected keyValue '123', got %q", rules[0].KeyValue)
	}

	// Test with anonymous context - should fall through to second rule
	ctx = RequestContext{IP: "1.2.3.4"}
	rules = provider.RulesFor(ctx)

	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	if rules[0].Name != "anonymous-per-ip" {
		t.Errorf("expected rule 'anonymous-per-ip', got %q", rules[0].Name)
	}
}

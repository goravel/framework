package cache

import (
	"testing"
	"time"
)

// mockConfig implements the config.Config interface minimally for tests.
// Only GetString is used by the code under test; other methods return
// sensible zero values.
type mockConfig struct {
	prefix string
}

func (m *mockConfig) Env(envName string, defaultValue ...any) any             { return nil }
func (m *mockConfig) EnvString(envName string, defaultValue ...string) string { return "" }
func (m *mockConfig) EnvBool(envName string, defaultValue ...bool) bool       { return false }
func (m *mockConfig) Add(name string, configuration any)                      {}
func (m *mockConfig) Get(path string, defaultValue ...any) any                { return m.prefix }
func (m *mockConfig) GetString(path string, defaultValue ...string) string {
	if path == "cache.prefix" {
		return m.prefix
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}
func (m *mockConfig) GetInt(path string, defaultValue ...int) int    { return 0 }
func (m *mockConfig) GetBool(path string, defaultValue ...bool) bool { return false }
func (m *mockConfig) GetDuration(path string, defaultValue ...time.Duration) time.Duration {
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}
func TestPrefixWithValue(t *testing.T) {
	cfg := &mockConfig{prefix: "myprefix"}
	if got := prefix(cfg); got != "myprefix:" {
		t.Fatalf("prefix returned %q, want %q", got, "myprefix:")
	}
}
func TestPrefixEmpty(t *testing.T) {
	cfg := &mockConfig{prefix: ""}
	if got := prefix(cfg); got != ":" {
		t.Fatalf("prefix returned %q, want %q", got, ":")
	}
}

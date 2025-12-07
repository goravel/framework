package cache

import (
	"testing"

	mocksconfig "github.com/goravel/framework/mocks/config"
)

func TestPrefixWithValue(t *testing.T) {
	cfg := mocksconfig.NewConfig(t)
	cfg.EXPECT().GetString("cache.prefix").Return("myprefix").Once()
	if got := prefix(cfg); got != "myprefix:" {
		t.Fatalf("prefix returned %q, want %q", got, "myprefix:")
	}
}

func TestPrefixEmpty(t *testing.T) {
	cfg := mocksconfig.NewConfig(t)
	cfg.EXPECT().GetString("cache.prefix").Return("").Once()
	if got := prefix(cfg); got != ":" {
		t.Fatalf("prefix returned %q, want %q", got, ":")
	}
}

package config_test

import (
	"testing"

	"github.com/DarknessKiller/pingopher/internal/config"
)

func TestLoadDNSOptimisticCache(t *testing.T) {
	t.Setenv("ENV", "test")

	for _, tc := range []struct {
		value string
		want  bool
	}{
		{"", false},
		{"false", false},
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"0", false},
	} {
		t.Setenv("PINGOPHER_DNS_OPTIMISTIC_CACHE", tc.value)
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("%q: expected no error, got %v", tc.value, err)
		}
		if cfg.DNSOptimisticCache != tc.want {
			t.Fatalf("%q: expected DNSOptimisticCache %v, got %v", tc.value, tc.want, cfg.DNSOptimisticCache)
		}
	}

	t.Setenv("PINGOPHER_DNS_OPTIMISTIC_CACHE", "yes")
	if cfg, err := config.Load(); err == nil {
		t.Fatalf("expected PINGOPHER_DNS_OPTIMISTIC_CACHE=yes to fail the load, got %+v", cfg)
	}
}

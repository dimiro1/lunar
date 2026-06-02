package cmd

import (
	"testing"

	"github.com/caarlos0/env/v11"
)

// TestEnvConfig pins the LUNAR_SERVER/LUNAR_TOKEN -> envConfig field bindings so
// a future tag rename can't silently break environment-variable configuration.
func TestEnvConfig(t *testing.T) {
	t.Setenv("LUNAR_SERVER", "http://example:9000")
	t.Setenv("LUNAR_TOKEN", "tok-xyz")

	cfg, err := env.ParseAs[envConfig]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server != "http://example:9000" {
		t.Errorf("Server = %q, want http://example:9000", cfg.Server)
	}
	if cfg.Token != "tok-xyz" {
		t.Errorf("Token = %q, want tok-xyz", cfg.Token)
	}
}

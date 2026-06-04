// Package config loads runtime configuration from the environment.
//
// Configuration lives in its own package (rather than in cmd) so that the
// per-package fx modules can depend on it directly when they need settings such
// as the execution timeout, base URL, or API key.
//
// Simple fields are populated by github.com/caarlos0/env via struct tags. A few
// concerns can't be expressed as plain tags and are handled after parsing:
//   - EXECUTION_TIMEOUT is an integer number of seconds (not a Go duration), so
//     a custom parser is registered for time.Duration.
//   - BASE_URL defaults to http://localhost:<PORT> when unset.
//   - The data directory is created on load.
//   - The API key falls back from env var → file → freshly generated key.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/dimiro1/lunar/internal/token"
)

// apiKeyFile is the filename under DataDir where a generated API key is stored.
const apiKeyFile = "api_key.txt"

// Config holds the application's runtime configuration.
type Config struct {
	Port             string        `env:"PORT" envDefault:"3000"`
	DataDir          string        `env:"DATA_DIR" envDefault:"./data"`
	ExecutionTimeout time.Duration `env:"EXECUTION_TIMEOUT" envDefault:"300"` // seconds
	APIKey           string        `env:"API_KEY"`
	BaseURL          string        `env:"BASE_URL"`
	// MetricsRetentionDays is how long pre-aggregated metric buckets are kept.
	// Metrics outlive executions (default 7-day retention) so dashboards can show
	// long-range trends; housekeeping deletes buckets older than this.
	MetricsRetentionDays int `env:"METRICS_RETENTION_DAYS" envDefault:"365"`
}

// Load reads configuration from the process environment.
func Load() (Config, error) {
	return parse(nil)
}

// parse builds the Config from the given environment. A nil environment reads
// the process environment (os.Environ); a non-nil map overrides it entirely,
// which keeps the loader testable without mutating global state.
func parse(environment map[string]string) (Config, error) {
	cfg, err := env.ParseAsWithOptions[Config](env.Options{
		Environment: environment,
		FuncMap: map[reflect.Type]env.ParserFunc{
			// EXECUTION_TIMEOUT is expressed in seconds.
			reflect.TypeFor[time.Duration](): parseSeconds,
		},
	})
	if err != nil {
		return Config{}, fmt.Errorf("parse environment: %w", err)
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:" + cfg.Port
	}

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return Config{}, fmt.Errorf("create data directory: %w", err)
	}

	if cfg.APIKey == "" {
		apiKey, err := resolveAPIKey(cfg.DataDir)
		if err != nil {
			return Config{}, err
		}
		cfg.APIKey = apiKey
	}

	return cfg, nil
}

// parseSeconds interprets a string as an integer number of seconds. To preserve
// the historical behaviour, an unparseable value falls back to the 5-minute
// default rather than failing the load.
func parseSeconds(v string) (any, error) {
	seconds, err := strconv.Atoi(v)
	if err != nil {
		return 5 * time.Minute, nil
	}
	return time.Duration(seconds) * time.Second, nil
}

// resolveAPIKey returns the API key from the on-disk file, generating and
// persisting a new one (0600) if the file does not yet exist.
func resolveAPIKey(dataDir string) (string, error) {
	apiKeyPath := filepath.Join(dataDir, apiKeyFile)

	keyBytes, err := os.ReadFile(apiKeyPath)
	if err == nil {
		return string(keyBytes), nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	apiKey, err := token.Generate()
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(apiKeyPath, []byte(apiKey), 0o600); err != nil {
		return "", err
	}

	slog.Info("Generated new API key", "key", apiKey, "file", apiKeyPath)
	return apiKey, nil
}

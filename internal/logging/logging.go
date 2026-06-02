// Package logging configures the process-wide slog logger.
//
// This is distinct from internal/services/logger, which stores per-execution
// function logs. Here we only set up the global slog handler used for the
// application's own diagnostics.
package logging

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
)

// config holds the logging settings read from the environment. slog.Level
// implements encoding.TextUnmarshaler, so caarlos0/env parses LOG_LEVEL values
// such as "debug" or "ERROR" (case-insensitive) directly into it.
type config struct {
	Level slog.Level `env:"LOG_LEVEL" envDefault:"INFO"`
}

// Setup configures the global slog logger from the LOG_LEVEL environment
// variable (DEBUG, INFO, WARN, or ERROR — case-insensitive). It should be
// called once, before the fx app is built, so that all wiring and lifecycle
// events are captured by the configured handler. An unset or unrecognized value
// falls back to INFO.
func Setup() {
	level, err := load(nil)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
	if err != nil {
		slog.Warn("invalid LOG_LEVEL, defaulting to info", "error", err)
	}
}

// load resolves the log level from the given environment; a nil map reads the
// process environment. On an unrecognized LOG_LEVEL it returns LevelInfo along
// with the parse error, so the caller can surface the misconfiguration without
// failing startup.
func load(environment map[string]string) (slog.Level, error) {
	cfg, err := env.ParseAsWithOptions[config](env.Options{Environment: environment})
	if err != nil {
		return slog.LevelInfo, err
	}
	return cfg.Level, nil
}

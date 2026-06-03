package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/dimiro1/lunar/frontend"
	"github.com/dimiro1/lunar/internal/api"
	"github.com/dimiro1/lunar/internal/config"
	internalcron "github.com/dimiro1/lunar/internal/cron"
	"github.com/dimiro1/lunar/internal/engine"
	"github.com/dimiro1/lunar/internal/graph"
	"github.com/dimiro1/lunar/internal/housekeeping"
	"github.com/dimiro1/lunar/internal/migrate"
	"github.com/dimiro1/lunar/internal/runner"
	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/services/env"
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"github.com/dimiro1/lunar/internal/starlarkrt"
	"github.com/dimiro1/lunar/internal/store"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	_ "modernc.org/sqlite"
)

// newApp is the composition root. It wires together the per-package fx modules
// — each of which owns its own slice of the dependency graph — plus a few
// app-level providers (configuration, the database connection, and the embedded
// frontend) that don't belong to any single domain package.
//
// fx resolves the whole graph in topological order and drives the OnStart/OnStop
// lifecycle hooks declared inside the cron, housekeeping, and api modules.
func newApp() *fx.App {
	return fx.New(appOptions())
}

// appOptions returns the complete set of fx options that define the application
// graph. It is shared by newApp (which builds and runs the app) and the
// graph-validation test (fx.ValidateApp), which checks the graph without
// invoking any constructors.
func appOptions() fx.Option {
	return fx.Options(
		// Route fx's own diagnostics through slog. Routine graph/lifecycle
		// events are emitted at DEBUG (hidden at the default log level, revealed
		// by LOG_LEVEL=debug to inspect the full dependency graph); fx errors
		// stay at ERROR and are always surfaced.
		fx.WithLogger(func() fxevent.Logger {
			l := &fxevent.SlogLogger{Logger: slog.Default()}
			l.UseLogLevel(slog.LevelDebug)
			return l
		}),

		// Match the previous 30s graceful-shutdown budget. This deadline is
		// passed to every OnStop hook (notably the HTTP server's Shutdown).
		fx.StopTimeout(30*time.Second),

		// App-level providers that aren't owned by a domain package.
		fx.Provide(
			provideConfig,
			provideDB,
			provideFrontend,
		),

		// Each package contributes its own slice of the graph.
		store.Module,
		kv.Module,
		env.Module,
		logger.Module,
		internalhttp.Module,
		ai.Module,
		email.Module,
		runner.Module,
		starlarkrt.Module,
		engine.Module,
		internalcron.Module,
		housekeeping.Module,
		graph.Module,
		api.Module,
	)
}

// provideConfig loads configuration from the environment.
func provideConfig() (config.Config, error) {
	return config.Load()
}

// provideFrontend exposes the embedded SPA as an http.Handler.
func provideFrontend() http.Handler {
	return frontend.Handler()
}

// provideDB opens the SQLite database, applies pragmas, runs migrations, and
// registers a hook to close the connection on shutdown.
func provideDB(lc fx.Lifecycle, cfg config.Config) (*sql.DB, error) {
	dbPath := filepath.Join(cfg.DataDir, "lunar.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("set pragma %q: %w", pragma, err)
		}
	}

	if err := migrate.Run(db, migrate.FS); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run database migrations: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			slog.Info("Closing database")
			return db.Close()
		},
	})

	return db, nil
}

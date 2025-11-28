package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dimiro1/faas-go/frontend"
	"github.com/dimiro1/faas-go/internal/ai"
	"github.com/dimiro1/faas-go/internal/api"
	"github.com/dimiro1/faas-go/internal/env"
	"github.com/dimiro1/faas-go/internal/housekeeping"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
	"github.com/dimiro1/faas-go/internal/migrate"
	store "github.com/dimiro1/faas-go/internal/store"
	_ "modernc.org/sqlite"
)

func main() {
	dataDir, err := initDataDir(os.Getenv)
	if err != nil {
		slog.Error("Failed to create data directory", "error", err)
		os.Exit(1)
	}

	config, err := loadConfig(os.Getenv, dataDir)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(config.DataDir, "faas.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}()

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		slog.Error("Failed to enable foreign keys", "error", err)
		os.Exit(1)
	}

	// Run database migrations
	if err := migrate.Run(db, migrate.FS); err != nil {
		slog.Error("Failed to run database migrations", "error", err)
		os.Exit(1)
	}

	apiDB := store.NewSQLiteDB(db)
	kvStore := kv.NewSQLiteStore(db)
	envStore := env.NewSQLiteStore(db)
	appLogger := logger.NewSQLiteLogger(db)
	aiRequestTracker := ai.NewSQLiteTracker(db)
	httpClient := internalhttp.NewDefaultClient()

	// Initialize housekeeping scheduler
	housekeepingScheduler := housekeeping.NewScheduler(apiDB)
	if err := housekeepingScheduler.Start(); err != nil {
		slog.Error("Failed to start housekeeping scheduler", "error", err)
		os.Exit(1)
	}

	server := api.NewServer(api.ServerConfig{
		DB:               apiDB,
		Logger:           appLogger,
		KVStore:          kvStore,
		EnvStore:         envStore,
		HTTPClient:       httpClient,
		AITracker:        aiRequestTracker,
		ExecutionTimeout: config.ExecutionTimeout,
		FrontendHandler:  frontend.Handler(),
		APIKey:           config.APIKey,
		BaseURL:          config.BaseURL,
	})

	addr := ":" + config.Port
	slog.Info("Starting FaaS-Go server",
		"port", config.Port,
		"data_dir", config.DataDir,
		"execution_timeout", config.ExecutionTimeout)
	slog.Info("Frontend available", "url", "http://localhost:"+config.Port)
	slog.Info("API available", "url", "http://localhost:"+config.Port+"/api")

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(addr); err != nil {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-shutdown:
		slog.Info("Shutdown signal received", "signal", sig)

		// Stop housekeeping scheduler
		slog.Info("Stopping housekeeping scheduler...")
		housekeepingScheduler.Stop()

		// Give active connections 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		slog.Info("Shutting down server gracefully...")
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Error during shutdown", "error", err)
			os.Exit(1)
		}
		slog.Info("Server stopped gracefully")

	case err := <-serverErr:
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

package main

import (
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dimiro1/faas-go/frontend"
	"github.com/dimiro1/faas-go/internal/api"
	"github.com/dimiro1/faas-go/internal/env"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
	_ "modernc.org/sqlite"
)

type Config struct {
	Port             string
	DataDir          string
	ExecutionTimeout time.Duration
}

func loadConfig(getenv func(string) string) Config {
	port := getenv("PORT")
	if port == "" {
		port = "3000"
	}

	dataDir := getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	timeoutStr := getenv("EXECUTION_TIMEOUT")
	timeout := 5 * time.Minute
	if timeoutStr != "" {
		if seconds, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(seconds) * time.Second
		}
	}

	return Config{
		Port:             port,
		DataDir:          dataDir,
		ExecutionTimeout: timeout,
	}
}

func main() {
	config := loadConfig(os.Getenv)

	if err := os.MkdirAll(config.DataDir, 0o755); err != nil {
		slog.Error("Failed to create data directory", "error", err)
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

	slog.Info("Running database migrations")
	if err := kv.Migrate(db); err != nil {
		slog.Error("Failed to run KV migrations", "error", err)
		os.Exit(1)
	}
	if err := env.Migrate(db); err != nil {
		slog.Error("Failed to run env migrations", "error", err)
		os.Exit(1)
	}
	if err := logger.Migrate(db); err != nil {
		slog.Error("Failed to run logger migrations", "error", err)
		os.Exit(1)
	}
	if err := api.Migrate(db); err != nil {
		slog.Error("Failed to run API migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("Migrations completed successfully")

	apiDB := api.NewSQLiteDB(db)
	kvStore := kv.NewSQLiteStore(db)
	envStore := env.NewSQLiteStore(db)
	appLogger := logger.NewSQLiteLogger(db)
	httpClient := internalhttp.NewDefaultClient()

	server := api.NewServer(api.ServerConfig{
		DB:               apiDB,
		Logger:           appLogger,
		KVStore:          kvStore,
		EnvStore:         envStore,
		HTTPClient:       httpClient,
		ExecutionTimeout: config.ExecutionTimeout,
		FrontendHandler:  frontend.Handler(),
	})

	addr := ":" + config.Port
	slog.Info("Starting FaaS-Go server",
		"port", config.Port,
		"data_dir", config.DataDir,
		"execution_timeout", config.ExecutionTimeout)
	slog.Info("Frontend available", "url", "http://localhost:"+config.Port)
	slog.Info("API available", "url", "http://localhost:"+config.Port+"/api")
	if err := server.ListenAndServe(addr); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

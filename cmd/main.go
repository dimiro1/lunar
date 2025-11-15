package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/dimiro1/faas-go/frontend"
	"github.com/dimiro1/faas-go/internal/api"
	"github.com/dimiro1/faas-go/internal/env"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
	_ "modernc.org/sqlite"
)

func main() {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll("./data", 0o755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite", "./data/faas.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Run migrations
	log.Println("Running database migrations...")
	if err := kv.Migrate(db); err != nil {
		log.Fatalf("Failed to run KV migrations: %v", err)
	}
	if err := env.Migrate(db); err != nil {
		log.Fatalf("Failed to run env migrations: %v", err)
	}
	if err := logger.Migrate(db); err != nil {
		log.Fatalf("Failed to run logger migrations: %v", err)
	}
	if err := api.Migrate(db); err != nil {
		log.Fatalf("Failed to run API migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Create API database
	apiDB := api.NewSQLiteDB(db)

	// Create stores and services
	kvStore := kv.NewSQLiteStore(db)
	envStore := env.NewSQLiteStore(db)
	appLogger := logger.NewSQLiteLogger(db)
	httpClient := internalhttp.NewDefaultClient()

	// Create API server with full configuration
	server := api.NewServer(api.ServerConfig{
		DB:               apiDB,
		Logger:           appLogger,
		KVStore:          kvStore,
		EnvStore:         envStore,
		HTTPClient:       httpClient,
		ExecutionTimeout: 30 * time.Second,
		FrontendHandler:  frontend.Handler(),
	})

	// Start server
	addr := ":3000"
	log.Printf("Starting FaaS-Go server on %s", addr)
	log.Printf("Frontend available at http://localhost%s", addr)
	log.Printf("API available at http://localhost%s/api", addr)
	if err := server.ListenAndServe(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

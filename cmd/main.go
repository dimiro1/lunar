package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/dimiro1/faas-go/internal/kv"
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

	// Run KV store migrations
	if err := kv.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create KV store
	store := kv.NewSQLiteStore(db)

	// Example usage
	if err := store.Set("function-123", "counter", "0"); err != nil {
		log.Printf("Failed to set value: %v", err)
	}

	value, err := store.Get("function-123", "counter")
	if err != nil {
		log.Printf("Failed to get value: %v", err)
	} else {
		log.Printf("Counter value: %s", value)
	}

	log.Println("FaaS-Go started successfully")
}

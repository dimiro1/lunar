package env

import (
	"database/sql"
	"fmt"
	"maps"

	_ "modernc.org/sqlite"
)

// Error represents an env store error
type Error struct {
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Env error: %s", e.Message)
}

// Store is an interface for environment variable storage operations
// functionID is used to isolate env vars between functions
type Store interface {
	Get(functionID, key string) (string, error)
	Set(functionID, key, value string) error
	Delete(functionID, key string) error
	All(functionID string) (map[string]string, error)
}

// MemoryStore is an in-memory implementation of Store
type MemoryStore struct {
	data map[string]map[string]string // functionID -> key -> value
}

// NewMemoryStore creates a new in-memory env store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]map[string]string),
	}
}

// Get retrieves a value by functionID and key
func (m *MemoryStore) Get(functionID, key string) (string, error) {
	ns, exists := m.data[functionID]
	if !exists {
		return "", &Error{Message: fmt.Sprintf("key not found: %s", key)}
	}

	value, exists := ns[key]
	if !exists {
		return "", &Error{Message: fmt.Sprintf("key not found: %s", key)}
	}
	return value, nil
}

// Set stores a key-value pair for a functionID
func (m *MemoryStore) Set(functionID, key, value string) error {
	if _, exists := m.data[functionID]; !exists {
		m.data[functionID] = make(map[string]string)
	}
	m.data[functionID][key] = value
	return nil
}

// Delete removes a key-value pair for a functionID
func (m *MemoryStore) Delete(functionID, key string) error {
	if ns, exists := m.data[functionID]; exists {
		delete(ns, key)
	}
	return nil
}

// All returns all environment variables for a functionID
func (m *MemoryStore) All(functionID string) (map[string]string, error) {
	ns, exists := m.data[functionID]
	if !exists {
		return make(map[string]string), nil
	}

	// Return a copy to prevent modification
	result := make(map[string]string, len(ns))
	maps.Copy(result, ns)
	return result, nil
}

// SQLiteStore is a SQLite-backed implementation of Store
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite-backed env store
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// Migrate runs the database migration for the env store
func Migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS env_vars (
		function_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		PRIMARY KEY (function_id, key)
	);
	CREATE INDEX IF NOT EXISTS idx_env_function_id ON env_vars(function_id);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// Get retrieves a value by functionID and key
func (s *SQLiteStore) Get(functionID, key string) (string, error) {
	var value string
	err := s.db.QueryRow(
		"SELECT value FROM env_vars WHERE function_id = ? AND key = ?",
		functionID, key,
	).Scan(&value)

	if err == sql.ErrNoRows {
		return "", &Error{Message: fmt.Sprintf("key not found: %s", key)}
	}
	if err != nil {
		return "", fmt.Errorf("failed to get value: %w", err)
	}

	return value, nil
}

// Set stores a key-value pair for a functionID
func (s *SQLiteStore) Set(functionID, key, value string) error {
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO env_vars (function_id, key, value) VALUES (?, ?, ?)",
		functionID, key, value,
	)
	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	return nil
}

// Delete removes a key-value pair for a functionID
func (s *SQLiteStore) Delete(functionID, key string) error {
	_, err := s.db.Exec(
		"DELETE FROM env_vars WHERE function_id = ? AND key = ?",
		functionID, key,
	)
	if err != nil {
		return fmt.Errorf("failed to delete value: %w", err)
	}
	return nil
}

// All returns all environment variables for a functionID
func (s *SQLiteStore) All(functionID string) (map[string]string, error) {
	rows, err := s.db.Query(
		"SELECT key, value FROM env_vars WHERE function_id = ?",
		functionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query env vars: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		result[key] = value
	}

	return result, nil
}

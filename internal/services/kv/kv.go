package kv

import (
	"database/sql"
	"errors"
	"fmt"
	"maps"
)

// Error represents a KV store error
type Error struct {
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("KV error: %s", e.Message)
}

// Store is an interface for key-value storage operations
// functionID is used to isolate data between functions
type Store interface {
	Get(functionID, key string) (string, error)
	Set(functionID, key, value string) error
	Delete(functionID, key string) error
	ListKeys(functionID string) ([]string, error)
	GetGlobal(key string) (string, error)
	SetGlobal(key, value string) error
	DeleteGlobal(key string) error
	ListGlobalKeys() ([]string, error)
	All(functionID string) (map[string]string, error)
	AllGlobal() (map[string]string, error)
}

// MemoryStore is an in-memory implementation of Store
type MemoryStore struct {
	data map[string]map[string]string // functionID -> key -> value
}

// NewMemoryStore creates a new in-memory KV store
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

// ListKeys lists all keys for a given functionID
func (m *MemoryStore) ListKeys(functionID string) ([]string, error) {
	ns, exists := m.data[functionID]
	if !exists {
		return nil, &Error{Message: fmt.Sprintf("functionID not found: %s", functionID)}
	}

	keys := make([]string, 0, len(ns))
	for key := range ns {
		keys = append(keys, key)
	}
	return keys, nil
}

// GetGlobal retrieves a value from the global key-value store
func (m *MemoryStore) GetGlobal(key string) (string, error) {
	if key == "" {
		return "", &Error{Message: "key cannot be empty"}
	}

	return m.Get("", key)
}

// SetGlobal sets a value in the global key-value store
func (m *MemoryStore) SetGlobal(key, value string) error {
	if key == "" {
		return &Error{Message: "key cannot be empty"}
	}

	return m.Set("", key, value)
}

// DeleteGlobal removes a key-value pair from the global key-value store
func (m *MemoryStore) DeleteGlobal(key string) error {
	if key == "" {
		return &Error{Message: "key cannot be empty"}
	}

	return m.Delete("", key)
}

// ListGlobalKeys lists all keys in the global key-value store
func (m *MemoryStore) ListGlobalKeys() ([]string, error) {
	return m.ListKeys("")
}

// All returns all key-value pairs for a given functionID
func (m *MemoryStore) All(functionID string) (map[string]string, error) {
	ns := m.data[functionID]

	// Return a copy to prevent modification
	result := make(map[string]string, len(ns))
	maps.Copy(result, ns)
	return result, nil
}

// AllGlobal returns all key-value pairs in the global store
func (m *MemoryStore) AllGlobal() (map[string]string, error) {
	return m.All("")
}

// SQLiteStore is a SQLite-backed implementation of Store
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite-backed KV store
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// Get retrieves a value by functionID and key
func (s *SQLiteStore) Get(functionID, key string) (string, error) {
	var value string
	err := s.db.QueryRow(
		"SELECT value FROM kv_store WHERE function_id = ? AND key = ?",
		functionID, key,
	).Scan(&value)

	if errors.Is(err, sql.ErrNoRows) {
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
		"INSERT OR REPLACE INTO kv_store (function_id, key, value) VALUES (?, ?, ?)",
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
		"DELETE FROM kv_store WHERE function_id = ? AND key = ?",
		functionID, key,
	)
	if err != nil {
		return fmt.Errorf("failed to delete value: %w", err)
	}
	return nil
}

// ListKeys lists all keys for a given functionID
func (s *SQLiteStore) ListKeys(functionID string) ([]string, error) {
	rows, err := s.db.Query(
		"SELECT key FROM kv_store WHERE function_id = ?", functionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan key: %w", err)
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return keys, nil
}

// GetGlobal retrieves a value from the global key-value store
func (s *SQLiteStore) GetGlobal(key string) (string, error) {
	if key == "" {
		return "", &Error{Message: "key cannot be empty"}
	}

	return s.Get("", key)
}

// SetGlobal sets a value in the global key-value store
func (s *SQLiteStore) SetGlobal(key, value string) error {
	if key == "" {
		return &Error{Message: "key cannot be empty"}
	}

	return s.Set("", key, value)
}

// DeleteGlobal removes a key-value pair from the global key-value store
func (s *SQLiteStore) DeleteGlobal(key string) error {
	if key == "" {
		return &Error{Message: "key cannot be empty"}
	}

	return s.Delete("", key)
}

// ListGlobalKeys lists all keys in the global key-value store
func (s *SQLiteStore) ListGlobalKeys() ([]string, error) {
	return s.ListKeys("")
}

// All returns all key-value pairs for a given functionID
func (s *SQLiteStore) All(functionID string) (map[string]string, error) {
	rows, err := s.db.Query(
		"SELECT key, value FROM kv_store WHERE function_id = ? ORDER BY key", functionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query kv store: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}()
	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return result, nil
}

// AllGlobal returns all key-value pairs in the global store
func (s *SQLiteStore) AllGlobal() (map[string]string, error) {
	return s.All("")
}

package kv

import (
	"database/sql"
	"errors"
	"fmt"
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
	OpenNamed(storeName string) error
	CloseNamed()
}

// MemoryStore is an in-memory implementation of Store
type MemoryStore struct {
	data      map[string]map[string]string // functionID -> key -> value
	StoreName string
}

// NewMemoryStore creates a new in-memory KV store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]map[string]string),
	}
}

// Open opens a new in-memory KV store (for interface compatibility)
func (m *MemoryStore) Open(storeName string) (Store, error) {
	return NewMemoryStore(), nil
}

// Get retrieves a value by functionID and key
func (m *MemoryStore) Get(functionID, key string) (string, error) {
	storeName := functionID
	if m.StoreName != "" {
		storeName = m.StoreName
	}

	ns, exists := m.data[storeName]
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
	storeName := functionID
	if m.StoreName != "" {
		storeName = m.StoreName
	}

	if _, exists := m.data[storeName]; !exists {
		m.data[storeName] = make(map[string]string)
	}
	m.data[storeName][key] = value
	return nil
}

// Delete removes a key-value pair for a functionID
func (m *MemoryStore) Delete(functionID, key string) error {
	storeName := functionID
	if m.StoreName != "" {
		storeName = m.StoreName
	}

	if ns, exists := m.data[storeName]; exists {
		delete(ns, key)
	}
	return nil
}

// OpenNamed opens a new in-memory KV store with a given name (for interface compatibility)
func (m *MemoryStore) OpenNamed(storeName string) error {
	if storeName == "" {
		return &Error{Message: "storeName cannot be empty"}
	}

	m.StoreName = storeName
	return nil
}

// CloseNamed clears the StoreName for the MemoryStore instance (for interface compatibility)
func (m *MemoryStore) CloseNamed() {
	m.StoreName = ""
}

// SQLiteStore is a SQLite-backed implementation of Store
type SQLiteStore struct {
	db        *sql.DB
	StoreName string
}

// NewSQLiteStore creates a new SQLite-backed KV store
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// Get retrieves a value by functionID and key
func (s *SQLiteStore) Get(functionID, key string) (string, error) {
	storeName := functionID
	if s.StoreName != "" {
		storeName = s.StoreName
	}

	var value string
	err := s.db.QueryRow(
		"SELECT value FROM kv_store WHERE function_id = ? AND key = ?",
		storeName, key,
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
	storeName := functionID
	if s.StoreName != "" {
		storeName = s.StoreName
	}

	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO kv_store (function_id, key, value) VALUES (?, ?, ?)",
		storeName, key, value,
	)
	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	return nil
}

// Delete removes a key-value pair for a functionID
func (s *SQLiteStore) Delete(functionID, key string) error {
	storeName := functionID
	if s.StoreName != "" {
		storeName = s.StoreName
	}

	_, err := s.db.Exec(
		"DELETE FROM kv_store WHERE function_id = ? AND key = ?",
		storeName, key,
	)
	if err != nil {
		return fmt.Errorf("failed to delete value: %w", err)
	}
	return nil
}

// OpenNamed sets the StoreName for the SQLiteStore instance, allowing for namespacing of stores
func (s *SQLiteStore) OpenNamed(storeName string) error {
	if storeName == "" {
		return &Error{Message: "storeName cannot be empty"}
	}

	s.StoreName = storeName
	return nil
}

// CloseNamed clears the StoreName for the SQLiteStore instance
func (s *SQLiteStore) CloseNamed() {
	s.StoreName = ""
}

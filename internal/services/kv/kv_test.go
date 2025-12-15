package kv

import (
	"database/sql"
	"os"
	"testing"

	"github.com/dimiro1/lunar/internal/migrate"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create a temporary database file
	tmpfile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_ = tmpfile.Close()

	db, err := sql.Open("sqlite", tmpfile.Name())
	if err != nil {
		_ = os.Remove(tmpfile.Name())
		t.Fatalf("Failed to open database: %v", err)
	}

	// Run migrations
	migrate.RunTest(t, db)

	t.Cleanup(func() {
		_ = db.Close()
		_ = os.Remove(tmpfile.Name())
	})

	return db
}

func TestSQLiteStore_SetAndGet(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	// Set a value
	err := store.Set("func-123", "key1", "value1")
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Get the value
	value, err := store.Get("func-123", "key1")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if value != "value1" {
		t.Errorf("Expected value 'value1', got '%s'", value)
	}
}

func TestSQLiteStore_GetNotFound(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	// Try to get a non-existent key
	_, err := store.Get("func-123", "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent key, got nil")
	}

	kvErr, ok := err.(*Error)
	if !ok {
		t.Errorf("Expected *Error type, got %T", err)
	} else if kvErr.Message != "key not found: nonexistent" {
		t.Errorf("Expected 'key not found' error, got '%s'", kvErr.Message)
	}
}

func TestSQLiteStore_Update(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	// Set initial value
	err := store.Set("func-123", "counter", "0")
	if err != nil {
		t.Fatalf("Failed to set initial value: %v", err)
	}

	// Update the value
	err = store.Set("func-123", "counter", "42")
	if err != nil {
		t.Fatalf("Failed to update value: %v", err)
	}

	// Get the updated value
	value, err := store.Get("func-123", "counter")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if value != "42" {
		t.Errorf("Expected value '42', got '%s'", value)
	}
}

func TestSQLiteStore_Delete(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	// Set a value
	err := store.Set("func-123", "key1", "value1")
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Delete the value
	err = store.Delete("func-123", "key1")
	if err != nil {
		t.Fatalf("Failed to delete value: %v", err)
	}

	// Try to get the deleted value
	_, err = store.Get("func-123", "key1")
	if err == nil {
		t.Error("Expected error for deleted key, got nil")
	}
}

func TestSQLiteStore_DeleteNonExistent(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	// Delete a non-existent key (should not error)
	err := store.Delete("func-123", "nonexistent")
	if err != nil {
		t.Errorf("Expected no error for deleting non-existent key, got %v", err)
	}
}

func TestSQLiteStore_FunctionIsolation(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	// Set values for different functions
	err := store.Set("func-123", "key", "value-123")
	if err != nil {
		t.Fatalf("Failed to set value for function func-123: %v", err)
	}

	err = store.Set("func-456", "key", "value-456")
	if err != nil {
		t.Fatalf("Failed to set value for function func-456: %v", err)
	}

	// Get values from each function
	value1, err := store.Get("func-123", "key")
	if err != nil {
		t.Fatalf("Failed to get value from func-123: %v", err)
	}

	value2, err := store.Get("func-456", "key")
	if err != nil {
		t.Fatalf("Failed to get value from func-456: %v", err)
	}

	// Verify isolation
	if value1 != "value-123" {
		t.Errorf("Expected value 'value-123' from func-123, got '%s'", value1)
	}

	if value2 != "value-456" {
		t.Errorf("Expected value 'value-456' from func-456, got '%s'", value2)
	}

	// Delete from one function shouldn't affect the other
	err = store.Delete("func-123", "key")
	if err != nil {
		t.Fatalf("Failed to delete from func-123: %v", err)
	}

	// func-456 should still have its value
	value2, err = store.Get("func-456", "key")
	if err != nil {
		t.Fatalf("Failed to get value from func-456 after delete: %v", err)
	}

	if value2 != "value-456" {
		t.Errorf("Expected value 'value-456' from func-456, got '%s'", value2)
	}

	// func-123 should not have the value
	_, err = store.Get("func-123", "key")
	if err == nil {
		t.Error("Expected error for deleted key in func-123, got nil")
	}
}

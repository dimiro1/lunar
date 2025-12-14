package env

import (
	"database/sql"
	"os"
	"testing"

	"github.com/dimiro1/lunar/internal/migrate"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	tmpfile, err := os.CreateTemp("", "test-env-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_ = tmpfile.Close()

	db, err := sql.Open("sqlite", tmpfile.Name())
	if err != nil {
		_ = os.Remove(tmpfile.Name())
		t.Fatalf("Failed to open database: %v", err)
	}

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

	err := store.Set("func-123", "DATABASE_URL", "postgres://localhost")
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	value, err := store.Get("func-123", "DATABASE_URL")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if value != "postgres://localhost" {
		t.Errorf("Expected 'postgres://localhost', got '%s'", value)
	}
}

func TestSQLiteStore_GetNotFound(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	_, err := store.Get("func-123", "NONEXISTENT")
	if err == nil {
		t.Error("Expected error for non-existent key, got nil")
	}

	envErr, ok := err.(*Error)
	if !ok {
		t.Errorf("Expected *Error type, got %T", err)
	} else if envErr.Message != "key not found: NONEXISTENT" {
		t.Errorf("Expected 'key not found' error, got '%s'", envErr.Message)
	}
}

func TestSQLiteStore_Update(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	err := store.Set("func-123", "API_KEY", "old-key")
	if err != nil {
		t.Fatalf("Failed to set initial value: %v", err)
	}

	err = store.Set("func-123", "API_KEY", "new-key")
	if err != nil {
		t.Fatalf("Failed to update value: %v", err)
	}

	value, err := store.Get("func-123", "API_KEY")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if value != "new-key" {
		t.Errorf("Expected 'new-key', got '%s'", value)
	}
}

func TestSQLiteStore_Delete(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	err := store.Set("func-123", "TEMP_VAR", "temp-value")
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	err = store.Delete("func-123", "TEMP_VAR")
	if err != nil {
		t.Fatalf("Failed to delete value: %v", err)
	}

	_, err = store.Get("func-123", "TEMP_VAR")
	if err == nil {
		t.Error("Expected error for deleted key, got nil")
	}
}

func TestSQLiteStore_DeleteNonExistent(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	err := store.Delete("func-123", "NONEXISTENT")
	if err != nil {
		t.Errorf("Expected no error for deleting non-existent key, got %v", err)
	}
}

func TestSQLiteStore_FunctionIsolation(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	err := store.Set("func-123", "API_KEY", "key-123")
	if err != nil {
		t.Fatalf("Failed to set value for function func-123: %v", err)
	}

	err = store.Set("func-456", "API_KEY", "key-456")
	if err != nil {
		t.Fatalf("Failed to set value for function func-456: %v", err)
	}

	value1, err := store.Get("func-123", "API_KEY")
	if err != nil {
		t.Fatalf("Failed to get value from func-123: %v", err)
	}

	value2, err := store.Get("func-456", "API_KEY")
	if err != nil {
		t.Fatalf("Failed to get value from func-456: %v", err)
	}

	if value1 != "key-123" {
		t.Errorf("Expected 'key-123' from func-123, got '%s'", value1)
	}

	if value2 != "key-456" {
		t.Errorf("Expected 'key-456' from func-456, got '%s'", value2)
	}

	// Delete from one function shouldn't affect the other
	err = store.Delete("func-123", "API_KEY")
	if err != nil {
		t.Fatalf("Failed to delete from func-123: %v", err)
	}

	value2, err = store.Get("func-456", "API_KEY")
	if err != nil {
		t.Fatalf("Failed to get value from func-456 after delete: %v", err)
	}

	if value2 != "key-456" {
		t.Errorf("Expected 'key-456' from func-456, got '%s'", value2)
	}

	_, err = store.Get("func-123", "API_KEY")
	if err == nil {
		t.Error("Expected error for deleted key in func-123, got nil")
	}
}

func TestSQLiteStore_All(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	vars := map[string]string{
		"DATABASE_URL": "postgres://localhost",
		"API_KEY":      "secret-key",
		"PORT":         "8080",
	}

	for k, v := range vars {
		err := store.Set("func-123", k, v)
		if err != nil {
			t.Fatalf("Failed to set key '%s': %v", k, err)
		}
	}

	all, err := store.All("func-123")
	if err != nil {
		t.Fatalf("Failed to get all vars: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 env vars, got %d", len(all))
	}

	for k, expectedValue := range vars {
		value, exists := all[k]
		if !exists {
			t.Errorf("Key '%s' not found in All() result", k)
		}
		if value != expectedValue {
			t.Errorf("Key '%s': expected '%s', got '%s'", k, expectedValue, value)
		}
	}
}

func TestSQLiteStore_AllEmpty(t *testing.T) {
	db := setupTestDB(t)
	store := NewSQLiteStore(db)

	all, err := store.All("nonexistent-function")
	if err != nil {
		t.Fatalf("Failed to get all vars: %v", err)
	}

	if len(all) != 0 {
		t.Errorf("Expected 0 env vars for non-existent function, got %d", len(all))
	}
}

func TestMemoryStore_SetAndGet(t *testing.T) {
	store := NewMemoryStore()

	err := store.Set("func-123", "DATABASE_URL", "postgres://localhost")
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	value, err := store.Get("func-123", "DATABASE_URL")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if value != "postgres://localhost" {
		t.Errorf("Expected 'postgres://localhost', got '%s'", value)
	}
}

func TestMemoryStore_FunctionIsolation(t *testing.T) {
	store := NewMemoryStore()

	_ = store.Set("func-123", "API_KEY", "key-123")
	_ = store.Set("func-456", "API_KEY", "key-456")

	value1, _ := store.Get("func-123", "API_KEY")
	value2, _ := store.Get("func-456", "API_KEY")

	if value1 != "key-123" {
		t.Errorf("Expected 'key-123', got '%s'", value1)
	}

	if value2 != "key-456" {
		t.Errorf("Expected 'key-456', got '%s'", value2)
	}
}

func TestMemoryStore_All(t *testing.T) {
	store := NewMemoryStore()

	_ = store.Set("func-123", "VAR1", "value1")
	_ = store.Set("func-123", "VAR2", "value2")

	all, err := store.All("func-123")
	if err != nil {
		t.Fatalf("Failed to get all vars: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("Expected 2 env vars, got %d", len(all))
	}

	if all["VAR1"] != "value1" || all["VAR2"] != "value2" {
		t.Error("Unexpected values in All() result")
	}
}

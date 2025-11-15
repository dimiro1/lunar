package logger

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	tmpfile, err := os.CreateTemp("", "test-logger-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_ = tmpfile.Close()

	db, err := sql.Open("sqlite", tmpfile.Name())
	if err != nil {
		_ = os.Remove(tmpfile.Name())
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := Migrate(db); err != nil {
		_ = db.Close()
		_ = os.Remove(tmpfile.Name())
		t.Fatalf("Failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
		_ = os.Remove(tmpfile.Name())
	})

	return db
}

func TestSQLiteLogger_Log(t *testing.T) {
	db := setupTestDB(t)
	logger := NewSQLiteLogger(db)

	logger.Info("func-123", "Test message")

	entries := logger.Entries("func-123")
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", entries[0].Message)
	}

	if entries[0].Level != Info {
		t.Errorf("Expected level Info, got %v", entries[0].Level)
	}

	if entries[0].Namespace != "func-123" {
		t.Errorf("Expected namespace 'func-123', got '%s'", entries[0].Namespace)
	}
}

func TestSQLiteLogger_LogLevels(t *testing.T) {
	db := setupTestDB(t)
	logger := NewSQLiteLogger(db)

	logger.Info("func-123", "Info message")
	logger.Debug("func-123", "Debug message")
	logger.Warn("func-123", "Warn message")
	logger.Error("func-123", "Error message")

	entries := logger.Entries("func-123")
	if len(entries) != 4 {
		t.Fatalf("Expected 4 entries, got %d", len(entries))
	}

	expectedLevels := []LogLevel{Info, Debug, Warn, Error}
	for i, entry := range entries {
		if entry.Level != expectedLevels[i] {
			t.Errorf("Entry %d: expected level %v, got %v", i, expectedLevels[i], entry.Level)
		}
	}
}

func TestSQLiteLogger_NamespaceIsolation(t *testing.T) {
	db := setupTestDB(t)
	logger := NewSQLiteLogger(db)

	logger.Info("func-123", "Message from 123")
	logger.Info("func-456", "Message from 456")
	logger.Error("func-123", "Error from 123")

	entries123 := logger.Entries("func-123")
	if len(entries123) != 2 {
		t.Fatalf("Expected 2 entries for func-123, got %d", len(entries123))
	}

	entries456 := logger.Entries("func-456")
	if len(entries456) != 1 {
		t.Fatalf("Expected 1 entry for func-456, got %d", len(entries456))
	}

	if entries456[0].Message != "Message from 456" {
		t.Errorf("Expected 'Message from 456', got '%s'", entries456[0].Message)
	}
}

func TestSQLiteLogger_EntriesByLevel(t *testing.T) {
	db := setupTestDB(t)
	logger := NewSQLiteLogger(db)

	logger.Info("func-123", "Info 1")
	logger.Error("func-123", "Error 1")
	logger.Info("func-123", "Info 2")
	logger.Error("func-123", "Error 2")

	errors := logger.EntriesByLevel("func-123", Error)
	if len(errors) != 2 {
		t.Fatalf("Expected 2 error entries, got %d", len(errors))
	}

	for _, entry := range errors {
		if entry.Level != Error {
			t.Errorf("Expected Error level, got %v", entry.Level)
		}
	}

	infos := logger.EntriesByLevel("func-123", Info)
	if len(infos) != 2 {
		t.Fatalf("Expected 2 info entries, got %d", len(infos))
	}
}

func TestSQLiteLogger_EntriesByNamespace(t *testing.T) {
	db := setupTestDB(t)
	logger := NewSQLiteLogger(db)

	logger.Info("func-123", "Message 1")
	logger.Error("func-123", "Message 2")
	logger.Warn("func-456", "Message 3")

	entries := logger.EntriesByNamespace("func-123")
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	for _, entry := range entries {
		if entry.Namespace != "func-123" {
			t.Errorf("Expected namespace 'func-123', got '%s'", entry.Namespace)
		}
	}
}

func TestSQLiteLogger_Timestamp(t *testing.T) {
	db := setupTestDB(t)
	logger := NewSQLiteLogger(db)

	logger.Info("func-123", "First message")
	logger.Info("func-123", "Second message")

	entries := logger.Entries("func-123")
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// Entries should be ordered by timestamp
	if entries[0].Timestamp > entries[1].Timestamp {
		t.Error("Entries are not ordered by timestamp")
	}

	// Both should have valid timestamps
	if entries[0].Timestamp == 0 || entries[1].Timestamp == 0 {
		t.Error("Invalid timestamp")
	}
}

func TestMemoryLogger_Basic(t *testing.T) {
	logger := NewMemoryLogger()

	logger.Info("func-123", "Test message")

	entries := logger.Entries("func-123")
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", entries[0].Message)
	}
}

func TestMemoryLogger_Clear(t *testing.T) {
	logger := NewMemoryLogger()

	logger.Info("func-123", "Message 1")
	logger.Info("func-123", "Message 2")

	if logger.Count() != 2 {
		t.Fatalf("Expected count 2, got %d", logger.Count())
	}

	logger.Clear()

	if logger.Count() != 0 {
		t.Errorf("Expected count 0 after clear, got %d", logger.Count())
	}
}

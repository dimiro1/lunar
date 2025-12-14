package logger

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/dimiro1/lunar/internal/migrate"
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

	migrate.RunTest(t, db)

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

	if entries[0].ExecutionID != "func-123" {
		t.Errorf("Expected executionID 'func-123', got '%s'", entries[0].ExecutionID)
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

func TestSQLiteLogger_ExecutionIsolation(t *testing.T) {
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

func TestSQLiteLogger_EntriesByExecutionID(t *testing.T) {
	db := setupTestDB(t)
	logger := NewSQLiteLogger(db)

	logger.Info("func-123", "Message 1")
	logger.Error("func-123", "Message 2")
	logger.Warn("func-456", "Message 3")

	entries := logger.EntriesByExecutionID("func-123")
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	for _, entry := range entries {
		if entry.ExecutionID != "func-123" {
			t.Errorf("Expected executionID 'func-123', got '%s'", entry.ExecutionID)
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

func TestMemoryLogger_SensitiveDataMasking(t *testing.T) {
	logger := NewMemoryLogger()

	testCases := []struct {
		name             string
		message          string
		shouldContain    string
		shouldNotContain string
	}{
		{
			name:             "JWT token masked",
			message:          "User authenticated with JWT: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:             "Bearer token masked",
			message:          "Authorization: Bearer abc123def456ghi789",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "abc123def456ghi789",
		},
		{
			name:             "API key masked",
			message:          "Using API key: sk_live_51234567890abcdefghij",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "sk_live_51234567890abcdefghij",
		},
		{
			name:             "AWS key masked",
			message:          "AWS Access Key: AKIAIOSFODNN7EXAMPLE",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "AKIAIOSFODNN7EXAMPLE",
		},
		{
			name:             "Password masked",
			message:          "User password: my_secret_password_123",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "my_secret_password_123",
		},
		{
			name:             "Regular message unchanged",
			message:          "User logged in successfully",
			shouldContain:    "User logged in successfully",
			shouldNotContain: "[REDACTED]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executionID := "test-exec-" + tc.name
			logger.Info(executionID, tc.message)

			entries := logger.Entries(executionID)
			if len(entries) != 1 {
				t.Fatalf("Expected 1 entry, got %d", len(entries))
			}

			loggedMessage := entries[0].Message

			if tc.shouldContain != "" && !contains(loggedMessage, tc.shouldContain) {
				t.Errorf("Expected logged message to contain %q, got %q", tc.shouldContain, loggedMessage)
			}

			if tc.shouldNotContain != "" && contains(loggedMessage, tc.shouldNotContain) {
				t.Errorf("Expected logged message to NOT contain %q, but it does: %q", tc.shouldNotContain, loggedMessage)
			}

			// Clean up
			logger.Clear()
		})
	}
}

func TestSQLiteLogger_SensitiveDataMasking(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	logger := NewSQLiteLogger(db)

	testCases := []struct {
		name             string
		message          string
		shouldContain    string
		shouldNotContain string
	}{
		{
			name:             "JWT token masked",
			message:          "Authenticated with token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:             "Bearer token masked",
			message:          "Request with Authorization: Bearer secret_token_here",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "secret_token_here",
		},
		{
			name:             "API key masked",
			message:          "API Key provided: api_key=sk_test_1234567890abcdef",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "sk_test_1234567890abcdef",
		},
		{
			name:             "Multiple secrets masked",
			message:          "Credentials: password=mysecret123 and token=abc123def456ghi",
			shouldContain:    "[REDACTED]",
			shouldNotContain: "mysecret123",
		},
		{
			name:             "Regular message unchanged",
			message:          "Processing request from user ID 12345",
			shouldContain:    "Processing request from user ID 12345",
			shouldNotContain: "[REDACTED]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executionID := "test-exec-sqlite-" + tc.name
			logger.Info(executionID, tc.message)

			entries := logger.Entries(executionID)
			if len(entries) != 1 {
				t.Fatalf("Expected 1 entry, got %d", len(entries))
			}

			loggedMessage := entries[0].Message

			if tc.shouldContain != "" && !contains(loggedMessage, tc.shouldContain) {
				t.Errorf("Expected logged message to contain %q, got %q", tc.shouldContain, loggedMessage)
			}

			if tc.shouldNotContain != "" && contains(loggedMessage, tc.shouldNotContain) {
				t.Errorf("Expected logged message to NOT contain %q, but it does: %q", tc.shouldNotContain, loggedMessage)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

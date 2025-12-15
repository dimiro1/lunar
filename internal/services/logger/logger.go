package logger

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/dimiro1/lunar/internal/masking"
	"github.com/rs/xid"
	_ "modernc.org/sqlite"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	Info LogLevel = iota
	Debug
	Warn
	Error
)

// String returns the string representation of a LogLevel
func (l LogLevel) String() string {
	switch l {
	case Info:
		return "INFO"
	case Debug:
		return "DEBUG"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a single log entry
type LogEntry struct {
	ExecutionID string
	Level       LogLevel
	Message     string
	Timestamp   int64
}

// Logger is an interface for logging operations
// executionID is used to isolate logs for each function execution
type Logger interface {
	Log(executionID string, level LogLevel, message string)
	Info(executionID string, message string)
	Debug(executionID string, message string)
	Warn(executionID string, message string)
	Error(executionID string, message string)
	Entries(executionID string) []LogEntry
	EntriesPaginated(executionID string, limit, offset int) ([]LogEntry, int64)
}

// MemoryLogger is an in-memory implementation of Logger
type MemoryLogger struct {
	mu      sync.RWMutex
	entries []LogEntry
}

// NewMemoryLogger creates a new in-memory logger
func NewMemoryLogger() *MemoryLogger {
	return &MemoryLogger{
		entries: make([]LogEntry, 0),
	}
}

// Log records a log entry with the specified executionID, level and message
func (m *MemoryLogger) Log(executionID string, level LogLevel, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Mask sensitive data in log messages
	maskedMessage := masking.MaskLogMessage(message)

	entry := LogEntry{
		ExecutionID: executionID,
		Level:       level,
		Message:     maskedMessage,
		Timestamp:   time.Now().Unix(),
	}

	m.entries = append(m.entries, entry)
}

// Info logs an informational message
func (m *MemoryLogger) Info(executionID string, message string) {
	m.Log(executionID, Info, message)
}

// Debug logs a debug message
func (m *MemoryLogger) Debug(executionID string, message string) {
	m.Log(executionID, Debug, message)
}

// Warn logs a warning message
func (m *MemoryLogger) Warn(executionID string, message string) {
	m.Log(executionID, Warn, message)
}

// Error logs an error message
func (m *MemoryLogger) Error(executionID string, message string) {
	m.Log(executionID, Error, message)
}

// Entries returns all log entries for the specified executionID
func (m *MemoryLogger) Entries(executionID string) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]LogEntry, 0)
	for _, entry := range m.entries {
		if entry.ExecutionID == executionID {
			entries = append(entries, entry)
		}
	}
	return entries
}

// EntriesPaginated returns paginated log entries for the specified executionID
func (m *MemoryLogger) EntriesPaginated(executionID string, limit, offset int) ([]LogEntry, int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Filter entries by executionID
	filtered := make([]LogEntry, 0)
	for _, entry := range m.entries {
		if entry.ExecutionID == executionID {
			filtered = append(filtered, entry)
		}
	}

	total := int64(len(filtered))

	// Apply pagination
	if offset >= len(filtered) {
		return []LogEntry{}, total
	}

	end := min(offset+limit, len(filtered))

	return filtered[offset:end], total
}

// EntriesByLevel returns all log entries with the specified executionID and level
func (m *MemoryLogger) EntriesByLevel(executionID string, level LogLevel) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]LogEntry, 0)
	for _, entry := range m.entries {
		if entry.ExecutionID == executionID && entry.Level == level {
			entries = append(entries, entry)
		}
	}
	return entries
}

// Clear removes all log entries
func (m *MemoryLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = make([]LogEntry, 0)
}

// Count returns the total number of log entries
func (m *MemoryLogger) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.entries)
}

// String returns a formatted string representation of all log entries
func (m *MemoryLogger) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := ""
	for _, entry := range m.entries {
		timestamp := time.Unix(entry.Timestamp, 0).Format("2006-01-02 15:04:05")
		result += fmt.Sprintf("[%s] [%s] %s: %s\n", timestamp, entry.ExecutionID, entry.Level, entry.Message)
	}
	return result
}

// SQLiteLogger is a SQLite-backed implementation of Logger
type SQLiteLogger struct {
	db *sql.DB
}

// NewSQLiteLogger creates a new SQLite-backed logger
func NewSQLiteLogger(db *sql.DB) *SQLiteLogger {
	return &SQLiteLogger{db: db}
}

// Log records a log entry with the specified executionID, level and message
func (s *SQLiteLogger) Log(executionID string, level LogLevel, message string) {
	// Mask sensitive data in log messages
	maskedMessage := masking.MaskLogMessage(message)

	id := xid.New().String()
	_, err := s.db.Exec(
		"INSERT INTO logs (id, execution_id, level, message, timestamp) VALUES (?, ?, ?, ?, ?)",
		id, executionID, int(level), maskedMessage, time.Now().Unix(),
	)
	if err != nil {
		// For now, we silently ignore errors to match the Logger interface
		fmt.Printf("Failed to write log: %v\n", err)
	}
}

// Info logs an informational message
func (s *SQLiteLogger) Info(executionID string, message string) {
	s.Log(executionID, Info, message)
}

// Debug logs a debug message
func (s *SQLiteLogger) Debug(executionID string, message string) {
	s.Log(executionID, Debug, message)
}

// Warn logs a warning message
func (s *SQLiteLogger) Warn(executionID string, message string) {
	s.Log(executionID, Warn, message)
}

// Error logs an error message
func (s *SQLiteLogger) Error(executionID string, message string) {
	s.Log(executionID, Error, message)
}

// Entries returns all log entries for the specified executionID
func (s *SQLiteLogger) Entries(executionID string) []LogEntry {
	rows, err := s.db.Query(
		"SELECT execution_id, level, message, timestamp FROM logs WHERE execution_id = ? ORDER BY timestamp",
		executionID,
	)
	if err != nil {
		return []LogEntry{}
	}
	defer func() { _ = rows.Close() }()

	return s.scanEntries(rows)
}

// EntriesPaginated returns paginated log entries for the specified executionID
func (s *SQLiteLogger) EntriesPaginated(executionID string, limit, offset int) ([]LogEntry, int64) {
	// Get total count
	var total int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM logs WHERE execution_id = ?", executionID).Scan(&total)
	if err != nil {
		return []LogEntry{}, 0
	}

	// Get paginated entries
	rows, err := s.db.Query(
		"SELECT execution_id, level, message, timestamp FROM logs WHERE execution_id = ? ORDER BY timestamp LIMIT ? OFFSET ?",
		executionID, limit, offset,
	)
	if err != nil {
		return []LogEntry{}, total
	}
	defer func() { _ = rows.Close() }()

	return s.scanEntries(rows), total
}

// EntriesByLevel returns all log entries with the specified executionID and level
func (s *SQLiteLogger) EntriesByLevel(executionID string, level LogLevel) []LogEntry {
	rows, err := s.db.Query(
		"SELECT execution_id, level, message, timestamp FROM logs WHERE execution_id = ? AND level = ? ORDER BY timestamp",
		executionID, int(level),
	)
	if err != nil {
		return []LogEntry{}
	}
	defer func() { _ = rows.Close() }()

	return s.scanEntries(rows)
}

// EntriesByExecutionID returns all log entries for the specified executionID
func (s *SQLiteLogger) EntriesByExecutionID(executionID string) []LogEntry {
	return s.Entries(executionID)
}

// scanEntries is a helper to scan rows into LogEntry slice
func (s *SQLiteLogger) scanEntries(rows *sql.Rows) []LogEntry {
	entries := make([]LogEntry, 0)
	for rows.Next() {
		var entry LogEntry
		var level int
		if err := rows.Scan(&entry.ExecutionID, &level, &entry.Message, &entry.Timestamp); err != nil {
			continue
		}
		entry.Level = LogLevel(level)
		entries = append(entries, entry)
	}
	return entries
}

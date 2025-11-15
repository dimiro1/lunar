package logger

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

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
	Namespace string
	Level     LogLevel
	Message   string
	Timestamp int64
}

// Logger is an interface for logging operations
// namespace is typically the function ID to isolate logs between functions
type Logger interface {
	Log(namespace string, level LogLevel, message string)
	Info(namespace string, message string)
	Debug(namespace string, message string)
	Warn(namespace string, message string)
	Error(namespace string, message string)
	Entries(namespace string) []LogEntry
	EntriesPaginated(namespace string, limit, offset int) ([]LogEntry, int64)
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

// Log records a log entry with the specified namespace, level and message
func (m *MemoryLogger) Log(namespace string, level LogLevel, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := LogEntry{
		Namespace: namespace,
		Level:     level,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	m.entries = append(m.entries, entry)
}

// Info logs an informational message
func (m *MemoryLogger) Info(namespace string, message string) {
	m.Log(namespace, Info, message)
}

// Debug logs a debug message
func (m *MemoryLogger) Debug(namespace string, message string) {
	m.Log(namespace, Debug, message)
}

// Warn logs a warning message
func (m *MemoryLogger) Warn(namespace string, message string) {
	m.Log(namespace, Warn, message)
}

// Error logs an error message
func (m *MemoryLogger) Error(namespace string, message string) {
	m.Log(namespace, Error, message)
}

// Entries returns all log entries for the specified namespace
func (m *MemoryLogger) Entries(namespace string) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]LogEntry, 0)
	for _, entry := range m.entries {
		if entry.Namespace == namespace {
			entries = append(entries, entry)
		}
	}
	return entries
}

// EntriesPaginated returns paginated log entries for the specified namespace
func (m *MemoryLogger) EntriesPaginated(namespace string, limit, offset int) ([]LogEntry, int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Filter entries by namespace
	filtered := make([]LogEntry, 0)
	for _, entry := range m.entries {
		if entry.Namespace == namespace {
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

// EntriesByLevel returns all log entries with the specified namespace and level
func (m *MemoryLogger) EntriesByLevel(namespace string, level LogLevel) []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]LogEntry, 0)
	for _, entry := range m.entries {
		if entry.Namespace == namespace && entry.Level == level {
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
		result += fmt.Sprintf("[%s] [%s] %s: %s\n", timestamp, entry.Namespace, entry.Level, entry.Message)
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

// Migrate runs the database migration for the logger
func Migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		namespace TEXT NOT NULL,
		level INTEGER NOT NULL,
		message TEXT NOT NULL,
		timestamp INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_logs_namespace ON logs(namespace);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
	CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// Log records a log entry with the specified namespace, level and message
func (s *SQLiteLogger) Log(namespace string, level LogLevel, message string) {
	id := xid.New().String()
	_, err := s.db.Exec(
		"INSERT INTO logs (id, namespace, level, message, timestamp) VALUES (?, ?, ?, ?, ?)",
		id, namespace, int(level), message, time.Now().Unix(),
	)
	if err != nil {
		// For now, we silently ignore errors to match the Logger interface
		fmt.Printf("Failed to write log: %v\n", err)
	}
}

// Info logs an informational message
func (s *SQLiteLogger) Info(namespace string, message string) {
	s.Log(namespace, Info, message)
}

// Debug logs a debug message
func (s *SQLiteLogger) Debug(namespace string, message string) {
	s.Log(namespace, Debug, message)
}

// Warn logs a warning message
func (s *SQLiteLogger) Warn(namespace string, message string) {
	s.Log(namespace, Warn, message)
}

// Error logs an error message
func (s *SQLiteLogger) Error(namespace string, message string) {
	s.Log(namespace, Error, message)
}

// Entries returns all log entries for the specified namespace
func (s *SQLiteLogger) Entries(namespace string) []LogEntry {
	rows, err := s.db.Query(
		"SELECT namespace, level, message, timestamp FROM logs WHERE namespace = ? ORDER BY timestamp",
		namespace,
	)
	if err != nil {
		return []LogEntry{}
	}
	defer func() { _ = rows.Close() }()

	return s.scanEntries(rows)
}

// EntriesPaginated returns paginated log entries for the specified namespace
func (s *SQLiteLogger) EntriesPaginated(namespace string, limit, offset int) ([]LogEntry, int64) {
	// Get total count
	var total int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM logs WHERE namespace = ?", namespace).Scan(&total)
	if err != nil {
		return []LogEntry{}, 0
	}

	// Get paginated entries
	rows, err := s.db.Query(
		"SELECT namespace, level, message, timestamp FROM logs WHERE namespace = ? ORDER BY timestamp LIMIT ? OFFSET ?",
		namespace, limit, offset,
	)
	if err != nil {
		return []LogEntry{}, total
	}
	defer func() { _ = rows.Close() }()

	return s.scanEntries(rows), total
}

// EntriesByLevel returns all log entries with the specified namespace and level
func (s *SQLiteLogger) EntriesByLevel(namespace string, level LogLevel) []LogEntry {
	rows, err := s.db.Query(
		"SELECT namespace, level, message, timestamp FROM logs WHERE namespace = ? AND level = ? ORDER BY timestamp",
		namespace, int(level),
	)
	if err != nil {
		return []LogEntry{}
	}
	defer func() { _ = rows.Close() }()

	return s.scanEntries(rows)
}

// EntriesByNamespace returns all log entries for the specified namespace
func (s *SQLiteLogger) EntriesByNamespace(namespace string) []LogEntry {
	return s.Entries(namespace)
}

// scanEntries is a helper to scan rows into LogEntry slice
func (s *SQLiteLogger) scanEntries(rows *sql.Rows) []LogEntry {
	entries := make([]LogEntry, 0)
	for rows.Next() {
		var entry LogEntry
		var level int
		if err := rows.Scan(&entry.Namespace, &level, &entry.Message, &entry.Timestamp); err != nil {
			continue
		}
		entry.Level = LogLevel(level)
		entries = append(entries, entry)
	}
	return entries
}

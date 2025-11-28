package api

import "github.com/dimiro1/faas-go/internal/store"

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// DiffLineType represents the type of change in a diff line
type DiffLineType string

const (
	DiffLineUnchanged DiffLineType = "unchanged"
	DiffLineAdded     DiffLineType = "added"
	DiffLineRemoved   DiffLineType = "removed"
)

// LogEntry represents a log entry from function execution
type LogEntry struct {
	Level     LogLevel `json:"level"`
	Message   string   `json:"message"`
	CreatedAt int64    `json:"created_at"`
}

// DiffLine represents a line in a version diff
type DiffLine struct {
	LineType DiffLineType `json:"line_type"`
	OldLine  *int         `json:"old_line,omitempty"`
	NewLine  *int         `json:"new_line,omitempty"`
	Content  string       `json:"content"`
}

// CreateFunctionRequest is the request body for creating a function
type CreateFunctionRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Code        string  `json:"code"`
}

// UpdateEnvVarsRequest is the request body for updating environment variables
type UpdateEnvVarsRequest struct {
	EnvVars map[string]string `json:"env_vars"`
}

// ListFunctionsResponse is the response for listing functions
type ListFunctionsResponse struct {
	Functions []store.FunctionWithActiveVersion `json:"functions"`
}

// ListVersionsResponse is the response for listing versions
type ListVersionsResponse struct {
	Versions []store.FunctionVersion `json:"versions"`
}

// ListExecutionsResponse is the response for listing executions
type ListExecutionsResponse struct {
	Executions []store.Execution `json:"executions"`
}

// ExecutionWithLogs includes execution details and logs
type ExecutionWithLogs struct {
	store.Execution
	Logs []LogEntry `json:"logs"`
}

// VersionDiffResponse is the response for version diff
type VersionDiffResponse struct {
	OldVersion int        `json:"old_version"`
	NewVersion int        `json:"new_version"`
	Diff       []DiffLine `json:"diff"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Pagination types moved to internal/db package - re-exported in store.go for compatibility

// PaginatedFunctionsResponse is the paginated response for listing functions
type PaginatedFunctionsResponse struct {
	Functions  []store.FunctionWithActiveVersion `json:"functions"`
	Pagination store.PaginationInfo              `json:"pagination"`
}

// PaginatedVersionsResponse is the paginated response for listing versions
type PaginatedVersionsResponse struct {
	Versions   []store.FunctionVersion `json:"versions"`
	Pagination store.PaginationInfo    `json:"pagination"`
}

// PaginatedExecutionsResponse is the paginated response for listing executions
type PaginatedExecutionsResponse struct {
	Executions []store.Execution    `json:"executions"`
	Pagination store.PaginationInfo `json:"pagination"`
}

// PaginatedLogsResponse is the paginated response for listing logs
type PaginatedLogsResponse struct {
	Logs       []LogEntry           `json:"logs"`
	Pagination store.PaginationInfo `json:"pagination"`
}

// PaginatedExecutionWithLogs includes execution details with paginated logs
type PaginatedExecutionWithLogs struct {
	store.Execution
	Logs       []LogEntry           `json:"logs"`
	Pagination store.PaginationInfo `json:"pagination"`
}

// PaginatedAIRequestsResponse is the paginated response for AI requests
type PaginatedAIRequestsResponse struct {
	AIRequests []store.AIRequest    `json:"ai_requests"`
	Pagination store.PaginationInfo `json:"pagination"`
}

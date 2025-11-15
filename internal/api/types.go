package api

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

// ExecutionStatus represents the status of a function execution
type ExecutionStatus string

const (
	ExecutionStatusPending ExecutionStatus = "pending"
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusError   ExecutionStatus = "error"
)

// Function represents a serverless function
type Function struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	EnvVars     map[string]string `json:"env_vars"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
}

// FunctionVersion represents a specific version of a function
type FunctionVersion struct {
	ID         string  `json:"id"`
	FunctionID string  `json:"function_id"`
	Version    int     `json:"version"`
	Code       string  `json:"code"`
	CreatedAt  int64   `json:"created_at"`
	CreatedBy  *string `json:"created_by,omitempty"`
	IsActive   bool    `json:"is_active"`
}

// Execution represents a function execution record
type Execution struct {
	ID                string          `json:"id"`
	FunctionID        string          `json:"function_id"`
	FunctionVersionID string          `json:"function_version_id"`
	Status            ExecutionStatus `json:"status"`
	DurationMs        *int64          `json:"duration_ms,omitempty"`
	ErrorMessage      *string         `json:"error_message,omitempty"`
	CreatedAt         int64           `json:"created_at"`
}

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

// UpdateFunctionRequest is the request body for updating a function
type UpdateFunctionRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Code        *string `json:"code,omitempty"`
}

// UpdateEnvVarsRequest is the request body for updating environment variables
type UpdateEnvVarsRequest struct {
	EnvVars map[string]string `json:"env_vars"`
}

// FunctionWithActiveVersion includes the function and its active version
type FunctionWithActiveVersion struct {
	Function
	ActiveVersion FunctionVersion `json:"active_version"`
}

// ListFunctionsResponse is the response for listing functions
type ListFunctionsResponse struct {
	Functions []FunctionWithActiveVersion `json:"functions"`
}

// ListVersionsResponse is the response for listing versions
type ListVersionsResponse struct {
	Versions []FunctionVersion `json:"versions"`
}

// ListExecutionsResponse is the response for listing executions
type ListExecutionsResponse struct {
	Executions []Execution `json:"executions"`
}

// ExecutionWithLogs includes execution details and logs
type ExecutionWithLogs struct {
	Execution
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

// PaginationParams contains pagination parameters
type PaginationParams struct {
	Limit  int // Number of items per page (default: 20, max: 100)
	Offset int // Number of items to skip (default: 0)
}

// Normalize applies defaults and constraints to pagination parameters
func (p PaginationParams) Normalize() PaginationParams {
	if p.Limit <= 0 {
		p.Limit = 20 // Default
	}
	if p.Limit > 100 {
		p.Limit = 100 // Max
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Total  int64 `json:"total"`  // Total number of items
	Limit  int   `json:"limit"`  // Items per page
	Offset int   `json:"offset"` // Items skipped
}

// PaginatedFunctionsResponse is the paginated response for listing functions
type PaginatedFunctionsResponse struct {
	Functions  []FunctionWithActiveVersion `json:"functions"`
	Pagination PaginationInfo              `json:"pagination"`
}

// PaginatedVersionsResponse is the paginated response for listing versions
type PaginatedVersionsResponse struct {
	Versions   []FunctionVersion `json:"versions"`
	Pagination PaginationInfo    `json:"pagination"`
}

// PaginatedExecutionsResponse is the paginated response for listing executions
type PaginatedExecutionsResponse struct {
	Executions []Execution    `json:"executions"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginatedLogsResponse is the paginated response for listing logs
type PaginatedLogsResponse struct {
	Logs       []LogEntry     `json:"logs"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginatedExecutionWithLogs includes execution details with paginated logs
type PaginatedExecutionWithLogs struct {
	Execution
	Logs       []LogEntry     `json:"logs"`
	Pagination PaginationInfo `json:"pagination"`
}

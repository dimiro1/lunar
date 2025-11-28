package store

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// ExecutionStatus represents the status of a function execution
type ExecutionStatus string

const (
	ExecutionStatusPending ExecutionStatus = "pending"
	ExecutionStatusSuccess ExecutionStatus = "success"
	ExecutionStatusError   ExecutionStatus = "error"
)

// AIRequestStatus represents the status of an AI API request
type AIRequestStatus string

const (
	AIRequestStatusSuccess AIRequestStatus = "success"
	AIRequestStatusError   AIRequestStatus = "error"
)

// AIRequest represents a tracked AI API request
type AIRequest struct {
	ID           string          `json:"id"`
	ExecutionID  string          `json:"execution_id"`
	Provider     string          `json:"provider"`
	Model        string          `json:"model"`
	Endpoint     string          `json:"endpoint"`
	RequestJSON  string          `json:"request_json"`
	ResponseJSON *string         `json:"response_json,omitempty"`
	Status       AIRequestStatus `json:"status"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	InputTokens  *int            `json:"input_tokens,omitempty"`
	OutputTokens *int            `json:"output_tokens,omitempty"`
	DurationMs   int64           `json:"duration_ms"`
	CreatedAt    int64           `json:"created_at"`
}

// Function represents a serverless function
type Function struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Description   *string           `json:"description,omitempty"`
	EnvVars       map[string]string `json:"env_vars"`
	Disabled      bool              `json:"disabled"`
	RetentionDays *int              `json:"retention_days,omitempty"`
	CreatedAt     int64             `json:"created_at"`
	UpdatedAt     int64             `json:"updated_at"`
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
	EventJSON         *string         `json:"event_json,omitempty"`
	CreatedAt         int64           `json:"created_at"`
}

// FunctionWithActiveVersion includes the function and its active version
type FunctionWithActiveVersion struct {
	Function
	ActiveVersion FunctionVersion `json:"active_version"`
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

// UpdateFunctionRequest is the request body for updating a function
type UpdateFunctionRequest struct {
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	Code          *string `json:"code,omitempty"`
	Disabled      *bool   `json:"disabled,omitempty"`
	RetentionDays *int    `json:"retention_days,omitempty"`
}

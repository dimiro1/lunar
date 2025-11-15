package events

// ExecutionContext represents the execution context for a function invocation
type ExecutionContext struct {
	// Unique identifier for this execution
	ExecutionID string `json:"execution_id"`

	// Function ID (used for isolating KV and Env data)
	FunctionID string `json:"function_id"`

	// Timestamp when execution started (Unix epoch in seconds)
	StartedAt int64 `json:"started_at"`

	// Request ID from the incoming event (for correlation)
	RequestID string `json:"request_id,omitempty"`

	// Function version
	Version string `json:"version,omitempty"`

	// Function name or identifier
	FunctionName string `json:"function_name,omitempty"`
}

package events

import "time"

// ExecutionContext represents the execution context for a function invocation
type ExecutionContext struct {
	// Unique identifier for this execution
	ExecutionID string `json:"execution_id"`

	// Timestamp when execution started (Unix epoch in seconds)
	StartedAt int64 `json:"started_at"`

	// Optional request ID from the incoming event (for correlation)
	RequestID *string `json:"request_id,omitempty"`

	// Function version
	Version *string `json:"version,omitempty"`

	// Function name or identifier
	FunctionName *string `json:"function_name,omitempty"`
}

// NewExecutionContext creates a new execution context with default values
func NewExecutionContext(executionID, functionName string) *ExecutionContext {
	return &ExecutionContext{
		ExecutionID:  executionID,
		StartedAt:    time.Now().Unix(),
		FunctionName: &functionName,
	}
}

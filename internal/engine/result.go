package engine

import (
	"time"

	"github.com/dimiro1/lunar/internal/events"
	"github.com/dimiro1/lunar/internal/store"
)

// ExecutionResult contains the outcome of a function execution.
type ExecutionResult struct {
	// ExecutionID is the unique identifier for this execution
	ExecutionID string

	// FunctionVersionID is the ID of the function version that was executed
	FunctionVersionID string

	// Response is the HTTP response from the function (for HTTP events)
	Response *events.HTTPResponse

	// Duration is how long the execution took
	Duration time.Duration

	// Status indicates whether execution succeeded or failed
	Status store.ExecutionStatus

	// Error contains the error if execution failed
	Error error
}

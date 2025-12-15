package engine

import (
	"github.com/dimiro1/lunar/internal/events"
	"github.com/dimiro1/lunar/internal/store"
)

// ExecutionRequest contains all information needed to execute a function.
type ExecutionRequest struct {
	// FunctionID is the unique identifier of the function to execute
	FunctionID string

	// Event is the trigger event (HTTP request, cron trigger, etc.)
	Event events.Event

	// Trigger indicates how the execution was triggered (HTTP, cron, etc.)
	Trigger store.ExecutionTrigger

	// BaseURL is the base URL of the server for generating function URLs
	BaseURL string
}

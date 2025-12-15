package engine

import (
	"context"

	"github.com/dimiro1/lunar/internal/events"
)

// Runtime is the interface for language-specific code executors.
// Implementations handle the actual execution of function code in a specific
// language runtime (Lua, JavaScript, Python, etc.).
type Runtime interface {
	// Execute runs the provided code with the given context and event.
	// It returns the execution result or an error if execution failed.
	Execute(ctx context.Context, req RuntimeRequest) (*RuntimeResult, error)
}

// RuntimeRequest contains all information needed to execute function code.
type RuntimeRequest struct {
	// Code is the function source code to execute
	Code string

	// Context provides execution metadata (function ID, execution ID, etc.)
	Context *events.ExecutionContext

	// Event is the trigger event (HTTP request, cron trigger, etc.)
	Event events.Event
}

// RuntimeResult contains the output from executing function code.
type RuntimeResult struct {
	// Response is the HTTP response from the function (for HTTP events)
	Response *events.HTTPResponse
}

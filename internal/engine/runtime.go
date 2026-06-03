package engine

import (
	"context"

	"github.com/dimiro1/lunar/internal/events"
)

// Supported language identifiers. A version's Language selects which runtime
// executes its code; an empty language defaults to Lua for backward compatibility.
const (
	LanguageLua      = "lua"
	LanguageStarlark = "starlark"
)

// DefaultLanguage is used when a version does not specify a language.
const DefaultLanguage = LanguageLua

// Runtime is the interface for language-specific code executors.
// Implementations handle the actual execution of function code in a specific
// language runtime (Lua, Starlark, etc.).
type Runtime interface {
	// Execute runs the provided code with the given context and event.
	// It returns the execution result or an error if execution failed.
	Execute(ctx context.Context, req RuntimeRequest) (*RuntimeResult, error)
}

// RuntimeEntry associates a language identifier with its runtime implementation.
// Each language runtime package contributes one entry; the engine selects among
// them by the executing version's Language.
type RuntimeEntry struct {
	Language string
	Runtime  Runtime
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

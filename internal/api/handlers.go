package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/dimiro1/lunar/internal/engine"
	"github.com/dimiro1/lunar/internal/events"
	"github.com/dimiro1/lunar/internal/store"
)

// ExecuteFunctionDeps holds dependencies for executing functions
type ExecuteFunctionDeps struct {
	Engine  engine.Engine
	BaseURL string
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// ExecuteFunctionHandler returns the handler for the public /fn/* execution
// passthrough. Unlike the management API (now served over GraphQL), this stays
// REST: it forwards an arbitrary HTTP request to a function and relays the
// function's status code, headers, and body back to the caller verbatim.
func ExecuteFunctionHandler(deps ExecuteFunctionDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		functionID := r.PathValue("function_id")

		// Parse HTTP event from request
		httpEvent, err := parseHTTPEvent(r, functionID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Determine trigger (from X-Trigger header or default to HTTP)
		trigger := store.ExecutionTriggerHTTP
		if r.Header.Get("X-Trigger") == "cron" {
			trigger = store.ExecutionTriggerCron
		}

		// Execute via engine
		result, err := deps.Engine.Execute(r.Context(), engine.ExecutionRequest{
			FunctionID: functionID,
			Event:      httpEvent,
			Trigger:    trigger,
			BaseURL:    deps.BaseURL,
		})
		// Handle engine errors
		if err != nil {
			handleEngineError(w, err)
			return
		}

		// Set execution metadata headers
		w.Header().Set("X-Function-Id", functionID)
		w.Header().Set("X-Function-Version-Id", result.FunctionVersionID)
		w.Header().Set("X-Execution-Id", result.ExecutionID)
		w.Header().Set("X-Execution-Duration-Ms", strconv.FormatInt(result.Duration.Milliseconds(), 10))

		// Handle execution errors
		if result.Error != nil {
			slog.Error("Function execution failed",
				"execution_id", result.ExecutionID,
				"function_id", functionID,
				"error", result.Error)
			writeError(w, http.StatusInternalServerError, "Function execution failed")
			return
		}

		// Write HTTP response
		writeExecutionResponse(w, result)
	}
}

// parseHTTPEvent creates an HTTPEvent from an HTTP request
func parseHTTPEvent(r *http.Request, functionID string) (events.HTTPEvent, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return events.HTTPEvent{}, err
	}

	// Compute relativePath by stripping /fn/{function_id} prefix
	prefix := "/fn/" + functionID
	relativePath := strings.TrimPrefix(r.URL.Path, prefix)
	if relativePath == "" {
		relativePath = "/"
	}

	httpEvent := events.HTTPEvent{
		Method:       r.Method,
		Path:         r.URL.Path,
		RelativePath: relativePath,
		Headers:      make(map[string]string),
		Body:         string(body),
		Query:        make(map[string]string),
	}

	// Copy headers
	for key, values := range r.Header {
		if len(values) > 0 {
			httpEvent.Headers[key] = values[0]
		}
	}

	// Copy query parameters
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			httpEvent.Query[key] = values[0]
		}
	}

	return httpEvent, nil
}

// handleEngineError writes the appropriate HTTP error for engine errors
func handleEngineError(w http.ResponseWriter, err error) {
	var fnNotFound *engine.FunctionNotFoundError
	var fnDisabled *engine.FunctionDisabledError
	var noVersion *engine.NoActiveVersionError

	switch {
	case errors.As(err, &fnNotFound):
		writeError(w, http.StatusNotFound, "Function not found")
	case errors.As(err, &fnDisabled):
		writeError(w, http.StatusForbidden, "Function is disabled")
	case errors.As(err, &noVersion):
		writeError(w, http.StatusInternalServerError, "No active version found")
	default:
		slog.Error("Unexpected engine error", "error", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// writeExecutionResponse writes the function's HTTP response to the client
func writeExecutionResponse(w http.ResponseWriter, result *engine.ExecutionResult) {
	if result.Response == nil {
		writeError(w, http.StatusInternalServerError, "Function did not return HTTP response")
		return
	}

	// Set custom headers from function response
	for key, value := range result.Response.Headers {
		w.Header().Set(key, value)
	}

	// Set the status code
	statusCode := result.Response.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	// Only set default Content-Type if the function didn't provide one
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(result.Response.Body))
}

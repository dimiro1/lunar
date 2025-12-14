package engine

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/dimiro1/lunar/internal/events"
	"github.com/dimiro1/lunar/internal/masking"
	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/services/env"
	"github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"github.com/dimiro1/lunar/internal/store"
)

// Engine orchestrates function execution with full lifecycle management.
type Engine interface {
	// Execute runs a function with the given request and returns the result.
	Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error)
}

// Config holds all dependencies needed to create an engine.
type Config struct {
	DB               store.DB
	Runtime          Runtime
	Logger           logger.Logger
	KVStore          kv.Store
	EnvStore         env.Store
	HTTPClient       http.Client
	AIClient         ai.Client
	AITracker        ai.Tracker
	EmailClient      email.Client
	EmailTracker     email.Tracker
	ExecutionTimeout time.Duration
	IDGenerator      func() string
}

// DefaultEngine is the default implementation of the Engine interface.
type DefaultEngine struct {
	db               store.DB
	runtime          Runtime
	logger           logger.Logger
	kvStore          kv.Store
	envStore         env.Store
	httpClient       http.Client
	aiClient         ai.Client
	aiTracker        ai.Tracker
	emailClient      email.Client
	emailTracker     email.Tracker
	executionTimeout time.Duration
	idGenerator      func() string
}

// New creates a new DefaultEngine with the given configuration.
func New(cfg Config) *DefaultEngine {
	return &DefaultEngine{
		db:               cfg.DB,
		runtime:          cfg.Runtime,
		logger:           cfg.Logger,
		kvStore:          cfg.KVStore,
		envStore:         cfg.EnvStore,
		httpClient:       cfg.HTTPClient,
		aiClient:         cfg.AIClient,
		aiTracker:        cfg.AITracker,
		emailClient:      cfg.EmailClient,
		emailTracker:     cfg.EmailTracker,
		executionTimeout: cfg.ExecutionTimeout,
		idGenerator:      cfg.IDGenerator,
	}
}

// Execute runs a function with full lifecycle management.
func (e *DefaultEngine) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
	startTime := time.Now()
	executionID := e.idGenerator()

	// Get the function
	fn, err := e.db.GetFunction(ctx, req.FunctionID)
	if err != nil {
		return nil, &FunctionNotFoundError{FunctionID: req.FunctionID}
	}

	// Check if function is disabled
	if fn.Disabled {
		return nil, &FunctionDisabledError{FunctionID: req.FunctionID}
	}

	// Get the active version
	version, err := e.db.GetActiveVersion(ctx, req.FunctionID)
	if err != nil {
		return nil, &NoActiveVersionError{FunctionID: req.FunctionID}
	}

	// Create execution context
	execContext := &events.ExecutionContext{
		ExecutionID: executionID,
		FunctionID:  req.FunctionID,
		StartedAt:   time.Now().Unix(),
		Version:     strconv.Itoa(version.Version),
		BaseURL:     req.BaseURL,
	}

	// Mask and serialize the event for storage
	eventJSONStr, err := e.serializeEvent(req.Event)
	if err != nil {
		return nil, err
	}

	// Create execution record
	execution := store.Execution{
		ID:                executionID,
		FunctionID:        req.FunctionID,
		FunctionVersionID: version.ID,
		Status:            store.ExecutionStatusPending,
		EventJSON:         &eventJSONStr,
		Trigger:           req.Trigger,
	}

	if _, err := e.db.CreateExecution(ctx, execution); err != nil {
		return nil, &ExecutionRecordError{Err: err}
	}

	// Execute via runtime
	runtimeReq := RuntimeRequest{
		Code:    version.Code,
		Context: execContext,
		Event:   req.Event,
	}

	runtimeResult, runErr := e.runtime.Execute(ctx, runtimeReq)

	// Calculate duration
	duration := time.Since(startTime)
	durationMs := duration.Milliseconds()

	// Determine execution status
	var errorMsg *string
	status := store.ExecutionStatusSuccess

	if runErr != nil {
		status = store.ExecutionStatusError
		errStr := runErr.Error()
		errorMsg = &errStr
	} else if runtimeResult != nil && runtimeResult.Response != nil && runtimeResult.Response.StatusCode >= 400 {
		status = store.ExecutionStatusError
	}

	// Save response JSON if function has SaveResponse enabled
	var responseJSON *string
	if fn.SaveResponse && runtimeResult != nil && runtimeResult.Response != nil {
		responseJSONStr := serializeHTTPResponse(runtimeResult.Response)
		responseJSON = &responseJSONStr
	}

	// Update execution record
	if err := e.db.UpdateExecution(ctx, executionID, status, &durationMs, errorMsg, responseJSON); err != nil {
		slog.Error("Failed to update execution status", "execution_id", executionID, "error", err)
	}

	// Log error if execution failed
	if runErr != nil {
		e.logger.Error(req.FunctionID, runErr.Error())
		slog.Error("Function execution failed",
			"execution_id", executionID,
			"function_id", req.FunctionID,
			"error", runErr)
	}

	// Build result
	result := &ExecutionResult{
		ExecutionID:       executionID,
		FunctionVersionID: version.ID,
		Duration:          duration,
		Status:            status,
		Error:             runErr,
	}

	if runtimeResult != nil {
		result.Response = runtimeResult.Response
	}

	return result, nil
}

// serializeEvent masks sensitive data and serializes the event to JSON.
func (e *DefaultEngine) serializeEvent(event events.Event) (string, error) {
	switch ev := event.(type) {
	case events.HTTPEvent:
		maskedEvent := masking.MaskHTTPEvent(ev)
		eventJSONBytes, err := json.Marshal(maskedEvent)
		if err != nil {
			return "", err
		}
		return string(eventJSONBytes), nil
	default:
		// For other event types, serialize directly
		eventJSONBytes, err := json.Marshal(event)
		if err != nil {
			return "", err
		}
		return string(eventJSONBytes), nil
	}
}

// MaxResponseBodySize is the maximum size of response body to store (1MB)
const MaxResponseBodySize = 1024 * 1024

// serializeHTTPResponse converts an HTTPResponse to a JSON string for storage.
// If the response body exceeds MaxResponseBodySize, it is truncated.
func serializeHTTPResponse(resp *events.HTTPResponse) string {
	// Create a copy to avoid modifying the original response
	respToStore := *resp

	// Truncate the body if it exceeds the maximum size
	if len(respToStore.Body) > MaxResponseBodySize {
		respToStore.Body = respToStore.Body[:MaxResponseBodySize] + "\n[TRUNCATED - Response exceeded 1MB]"
	}

	jsonBytes, err := json.Marshal(respToStore)
	if err != nil {
		slog.Error("Failed to serialize HTTP response", "error", err)
		return "{}"
	}
	return string(jsonBytes)
}

package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/dimiro1/faas-go/internal/ai"
	"github.com/dimiro1/faas-go/internal/diff"
	"github.com/dimiro1/faas-go/internal/env"
	"github.com/dimiro1/faas-go/internal/events"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
	"github.com/dimiro1/faas-go/internal/masking"
	"github.com/dimiro1/faas-go/internal/runner"
	"github.com/dimiro1/faas-go/internal/store"
	"github.com/rs/xid"
)

// ExecuteFunctionDeps holds dependencies for executing functions
type ExecuteFunctionDeps struct {
	DB               store.DB
	Logger           logger.Logger
	KVStore          kv.Store
	EnvStore         env.Store
	HTTPClient       internalhttp.Client
	AIClient         ai.Client
	AITracker        ai.Tracker
	ExecutionTimeout time.Duration
	BaseURL          string
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

func parsePaginationParams(r *http.Request) store.PaginationParams {
	params := store.PaginationParams{
		Limit:  20, // Default
		Offset: 0,  // Default
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			// Enforce maximum page size
			params.Limit = min(limit, MaxPageSize)
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			params.Offset = offset
		}
	}

	return params
}

func generateID() string {
	return xid.New().String()
}

func generateDiff(oldCode, newCode string, oldVersion, newVersion int) VersionDiffResponse {
	// Use the diff package to generate the diff
	result := diff.Compare(oldCode, newCode)

	// Convert from diff.Line to DiffLine (API type)
	apiDiffLines := make([]DiffLine, len(result.Lines))
	for i, line := range result.Lines {
		apiDiffLines[i] = DiffLine{
			LineType: DiffLineType(line.Type),
			OldLine:  line.OldLine,
			NewLine:  line.NewLine,
			Content:  line.Content,
		}
	}

	return VersionDiffResponse{
		OldVersion: oldVersion,
		NewVersion: newVersion,
		Diff:       apiDiffLines,
	}
}

// Functional handler factories - each handler explicitly declares its dependencies

// CreateFunctionHandler returns a handler for creating functions
func CreateFunctionHandler(database store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateFunctionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate request
		if err := ValidateCreateFunctionRequest(&req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Generate unique ID for the function
		functionID := generateID()

		// Create the function
		fn := store.Function{
			ID:          functionID,
			Name:        req.Name,
			Description: req.Description,
			EnvVars:     make(map[string]string),
		}

		createdFn, err := database.CreateFunction(r.Context(), fn)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create function")
			return
		}

		// Create the first version
		version, err := database.CreateVersion(r.Context(), createdFn.ID, req.Code, nil)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create initial version")
			return
		}

		// Return function with active version
		resp := store.FunctionWithActiveVersion{
			Function:      createdFn,
			ActiveVersion: version,
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// ListFunctionsHandler returns a handler for listing functions
func ListFunctionsHandler(database store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := parsePaginationParams(r)

		functions, total, err := database.ListFunctions(r.Context(), params)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list functions")
			return
		}

		params = params.Normalize()
		resp := PaginatedFunctionsResponse{
			Functions: functions,
			Pagination: store.PaginationInfo{
				Total:  total,
				Limit:  params.Limit,
				Offset: params.Offset,
			},
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// GetFunctionHandler returns a handler for getting a specific function
func GetFunctionHandler(database store.DB, envStore env.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		fn, err := database.GetFunction(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Function not found")
			return
		}

		activeVersion, err := database.GetActiveVersion(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "No active version found")
			return
		}

		// Get env vars from env store
		envVars, err := envStore.All(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get env vars")
			return
		}
		fn.EnvVars = envVars

		resp := store.FunctionWithActiveVersion{
			Function:      fn,
			ActiveVersion: activeVersion,
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// UpdateFunctionHandler returns a handler for updating functions
func UpdateFunctionHandler(database store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var req store.UpdateFunctionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate request
		if err := ValidateUpdateFunctionRequest(&req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// If code is provided, create a new version
		if req.Code != nil {
			_, err := database.CreateVersion(r.Context(), id, *req.Code, nil)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to create new version")
				return
			}
		}

		// If metadata is provided, update the function
		if req.Name != nil || req.Description != nil || req.Disabled != nil || req.RetentionDays != nil {
			err := database.UpdateFunction(r.Context(), id, req)
			if err != nil {
				writeError(w, http.StatusNotFound, "Function not found")
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

// DeleteFunctionHandler returns a handler for deleting functions
func DeleteFunctionHandler(database store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := database.DeleteFunction(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete function")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateEnvVarsHandler returns a handler for updating environment variables
func UpdateEnvVarsHandler(database store.DB, envStore env.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var req UpdateEnvVarsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate request
		if err := ValidateUpdateEnvVarsRequest(&req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Verify function exists
		_, err := database.GetFunction(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Function not found")
			return
		}

		// Get current env vars from env store
		currentEnvVars, err := envStore.All(id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get current env vars")
			return
		}

		// Delete removed env vars
		for key := range currentEnvVars {
			if _, exists := req.EnvVars[key]; !exists {
				if err := envStore.Delete(id, key); err != nil {
					writeError(w, http.StatusInternalServerError, "Failed to delete env var")
					return
				}
			}
		}

		// Set new/updated env vars
		for key, value := range req.EnvVars {
			if err := envStore.Set(id, key, value); err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to set env var")
				return
			}
		}

		// Get the active version to return
		activeVersion, err := database.GetActiveVersion(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get active version")
			return
		}

		writeJSON(w, http.StatusOK, activeVersion)
	}
}

// ListVersionsHandler returns a handler for listing function versions
func ListVersionsHandler(database store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		params := parsePaginationParams(r)

		// Verify function exists
		if _, err := database.GetFunction(r.Context(), id); err != nil {
			writeError(w, http.StatusNotFound, "Function not found")
			return
		}

		versions, total, err := database.ListVersions(r.Context(), id, params)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list versions")
			return
		}

		params = params.Normalize()
		resp := PaginatedVersionsResponse{
			Versions: versions,
			Pagination: store.PaginationInfo{
				Total:  total,
				Limit:  params.Limit,
				Offset: params.Offset,
			},
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// GetVersionHandler returns a handler for getting a specific version
func GetVersionHandler(database store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		versionStr := r.PathValue("version")

		// Parse version number
		versionNum, err := strconv.Atoi(versionStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid version number")
			return
		}

		version, err := database.GetVersion(r.Context(), id, versionNum)
		if err != nil {
			writeError(w, http.StatusNotFound, "Version not found")
			return
		}

		writeJSON(w, http.StatusOK, version)
	}
}

// ActivateVersionHandler returns a handler for activating a version
func ActivateVersionHandler(database store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		versionStr := r.PathValue("version")

		// Parse version number
		versionNum, err := strconv.Atoi(versionStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid version number")
			return
		}

		// Activate the version
		if err := database.ActivateVersion(r.Context(), id, versionNum); err != nil {
			writeError(w, http.StatusNotFound, "Version not found")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// GetVersionDiffHandler returns a handler for getting diff between versions
func GetVersionDiffHandler(database store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		v1Str := r.PathValue("v1")
		v2Str := r.PathValue("v2")

		// Parse version numbers
		v1, err := strconv.Atoi(v1Str)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid version number v1")
			return
		}

		v2, err := strconv.Atoi(v2Str)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid version number v2")
			return
		}

		// Get both versions from the database
		version1, err := database.GetVersion(r.Context(), id, v1)
		if err != nil {
			writeError(w, http.StatusNotFound, "Version v1 not found")
			return
		}

		version2, err := database.GetVersion(r.Context(), id, v2)
		if err != nil {
			writeError(w, http.StatusNotFound, "Version v2 not found")
			return
		}

		// Generate the diff using our utility function
		diffResult := generateDiff(version1.Code, version2.Code, v1, v2)

		writeJSON(w, http.StatusOK, diffResult)
	}
}

// ListExecutionsHandler returns a handler for listing executions
func ListExecutionsHandler(database store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		params := parsePaginationParams(r)

		// Verify function exists
		if _, err := database.GetFunction(r.Context(), id); err != nil {
			writeError(w, http.StatusNotFound, "Function not found")
			return
		}

		executions, total, err := database.ListExecutions(r.Context(), id, params)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list executions")
			return
		}

		params = params.Normalize()
		resp := PaginatedExecutionsResponse{
			Executions: executions,
			Pagination: store.PaginationInfo{
				Total:  total,
				Limit:  params.Limit,
				Offset: params.Offset,
			},
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// GetExecutionHandler returns a handler for getting a specific execution
func GetExecutionHandler(database store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		execution, err := database.GetExecution(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Execution not found")
			return
		}

		writeJSON(w, http.StatusOK, execution)
	}
}

// GetExecutionLogsHandler returns a handler for getting execution logs
func GetExecutionLogsHandler(database store.DB, appLogger logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		params := parsePaginationParams(r)

		// Get the execution
		execution, err := database.GetExecution(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Execution not found")
			return
		}

		// Get the logs for this execution from the logger
		params = params.Normalize()
		logEntries, total := appLogger.EntriesPaginated(id, params.Limit, params.Offset)

		// Convert logger.LogEntry to API LogEntry format
		apiLogs := make([]LogEntry, len(logEntries))
		for i, entry := range logEntries {
			// Map logger.LogLevel (int) to API LogLevel (string)
			var level LogLevel
			switch entry.Level {
			case logger.Debug:
				level = LogLevelDebug
			case logger.Info:
				level = LogLevelInfo
			case logger.Warn:
				level = LogLevelWarn
			case logger.Error:
				level = LogLevelError
			default:
				level = LogLevelInfo
			}

			apiLogs[i] = LogEntry{
				Level:     level,
				Message:   entry.Message,
				CreatedAt: entry.Timestamp,
			}
		}

		resp := PaginatedExecutionWithLogs{
			Execution: execution,
			Logs:      apiLogs,
			Pagination: store.PaginationInfo{
				Total:  total,
				Limit:  params.Limit,
				Offset: params.Offset,
			},
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// GetExecutionAIRequestsHandler returns a handler for getting AI requests for an execution
func GetExecutionAIRequestsHandler(database store.DB, aiTracker ai.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		params := parsePaginationParams(r)

		// Verify execution exists
		_, err := database.GetExecution(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Execution not found")
			return
		}

		// Get AI requests for this execution
		params = params.Normalize()
		aiRequests, total := aiTracker.RequestsPaginated(id, params.Limit, params.Offset)

		resp := PaginatedAIRequestsResponse{
			AIRequests: aiRequests,
			Pagination: store.PaginationInfo{
				Total:  total,
				Limit:  params.Limit,
				Offset: params.Offset,
			},
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// ExecuteFunctionHandler returns a handler for executing functions
func ExecuteFunctionHandler(deps ExecuteFunctionDeps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		functionID := r.PathValue("function_id")
		executionID := generateID()

		// Get the function
		fn, err := deps.DB.GetFunction(r.Context(), functionID)
		if err != nil {
			writeError(w, http.StatusNotFound, "Function not found")
			return
		}

		// Check if function is disabled
		if fn.Disabled {
			writeError(w, http.StatusForbidden, "Function is disabled")
			return
		}

		// Get the active version
		version, err := deps.DB.GetActiveVersion(r.Context(), functionID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "No active version found")
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Create HTTP event from the request
		httpEvent := events.HTTPEvent{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: make(map[string]string),
			Body:    string(body),
			Query:   make(map[string]string),
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

		// Create execution context
		execContext := &events.ExecutionContext{
			ExecutionID: executionID,
			FunctionID:  functionID,
			StartedAt:   time.Now().Unix(),
			Version:     strconv.Itoa(version.Version),
			BaseURL:     deps.BaseURL,
		}

		// Mask sensitive data in the event before storing
		maskedEvent := masking.MaskHTTPEvent(httpEvent)

		// Serialize the masked event to JSON for storage
		eventJSONBytes, err := json.Marshal(maskedEvent)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to serialize event")
			return
		}
		eventJSONStr := string(eventJSONBytes)

		// Create execution record
		execution := store.Execution{
			ID:                executionID,
			FunctionID:        functionID,
			FunctionVersionID: version.ID,
			Status:            store.ExecutionStatusPending,
			EventJSON:         &eventJSONStr,
		}

		_, err = deps.DB.CreateExecution(r.Context(), execution)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create execution record")
			return
		}

		// Prepare runner dependencies
		runnerDeps := runner.Dependencies{
			Logger:    deps.Logger,
			KV:        deps.KVStore,
			Env:       deps.EnvStore,
			HTTP:      deps.HTTPClient,
			AI:        deps.AIClient,
			AITracker: deps.AITracker,
			Timeout:   deps.ExecutionTimeout,
		}

		// Execute the function
		req := runner.Request{
			Context: execContext,
			Event:   httpEvent,
			Code:    version.Code,
		}

		resp, runErr := runner.Run(r.Context(), runnerDeps, req)

		// Calculate duration
		duration := time.Since(startTime).Milliseconds()

		// Determine execution status
		var errorMsg *string
		status := store.ExecutionStatusSuccess

		if runErr != nil {
			status = store.ExecutionStatusError
			errStr := runErr.Error()
			errorMsg = &errStr
		} else if resp.HTTP != nil && resp.HTTP.StatusCode >= 400 {
			// Mark as error if the function returns an error status code
			status = store.ExecutionStatusError
		}

		if err := deps.DB.UpdateExecution(r.Context(), executionID, status, &duration, errorMsg); err != nil {
			slog.Error("Failed to update execution status", "execution_id", executionID, "error", err)
		}

		// Set custom headers
		w.Header().Set("X-Function-Id", functionID)
		w.Header().Set("X-Function-Version-Id", version.ID)
		w.Header().Set("X-Execution-Id", executionID)
		w.Header().Set("X-Execution-Duration-Ms", strconv.FormatInt(duration, 10))

		// If execution failed, log details and return generic error
		if runErr != nil {
			deps.Logger.Error(functionID, runErr.Error())
			slog.Error("Function execution failed",
				"execution_id", executionID,
				"function_id", functionID,
				"error", runErr)
			writeError(w, http.StatusInternalServerError, "Function execution failed")
			return
		}

		// Return HTTP response
		if resp.HTTP != nil {
			// Set custom headers from function response
			for key, value := range resp.HTTP.Headers {
				w.Header().Set(key, value)
			}

			// Set status code
			statusCode := resp.HTTP.StatusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			// Write response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			_, _ = w.Write([]byte(resp.HTTP.Body))
		} else {
			// No HTTP response, return 500
			writeError(w, http.StatusInternalServerError, "Function did not return HTTP response")
		}
	}
}

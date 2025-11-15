package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/dimiro1/faas-go/internal/diff"
	"github.com/dimiro1/faas-go/internal/env"
	"github.com/dimiro1/faas-go/internal/events"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
	"github.com/dimiro1/faas-go/internal/runner"
	"github.com/rs/xid"
)

// ExecuteFunctionDeps holds dependencies for executing functions
type ExecuteFunctionDeps struct {
	DB               DB
	Logger           logger.Logger
	KVStore          kv.Store
	EnvStore         env.Store
	HTTPClient       internalhttp.Client
	ExecutionTimeout time.Duration
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parsePaginationParams(r *http.Request) PaginationParams {
	params := PaginationParams{
		Limit:  20, // Default
		Offset: 0,  // Default
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
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
func CreateFunctionHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateFunctionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Generate unique ID for the function
		functionID := generateID()

		// Create the function
		fn := Function{
			ID:          functionID,
			Name:        req.Name,
			Description: req.Description,
			EnvVars:     make(map[string]string),
		}

		createdFn, err := db.CreateFunction(r.Context(), fn)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create function")
			return
		}

		// Create the first version
		version, err := db.CreateVersion(r.Context(), createdFn.ID, req.Code, nil)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create initial version")
			return
		}

		// Return function with active version
		resp := FunctionWithActiveVersion{
			Function:      createdFn,
			ActiveVersion: version,
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// ListFunctionsHandler returns a handler for listing functions
func ListFunctionsHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := parsePaginationParams(r)

		functions, total, err := db.ListFunctions(r.Context(), params)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list functions")
			return
		}

		// Get active versions for each function
		functionsWithVersions := make([]FunctionWithActiveVersion, 0, len(functions))
		for _, fn := range functions {
			activeVersion, err := db.GetActiveVersion(r.Context(), fn.ID)
			if err != nil {
				// Skip functions without active versions
				continue
			}
			functionsWithVersions = append(functionsWithVersions, FunctionWithActiveVersion{
				Function:      fn,
				ActiveVersion: activeVersion,
			})
		}

		params = params.Normalize()
		resp := PaginatedFunctionsResponse{
			Functions: functionsWithVersions,
			Pagination: PaginationInfo{
				Total:  total,
				Limit:  params.Limit,
				Offset: params.Offset,
			},
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// GetFunctionHandler returns a handler for getting a specific function
func GetFunctionHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		fn, err := db.GetFunction(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Function not found")
			return
		}

		activeVersion, err := db.GetActiveVersion(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "No active version found")
			return
		}

		resp := FunctionWithActiveVersion{
			Function:      fn,
			ActiveVersion: activeVersion,
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// UpdateFunctionHandler returns a handler for updating functions
func UpdateFunctionHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var req UpdateFunctionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// If code is provided, create a new version
		if req.Code != nil {
			_, err := db.CreateVersion(r.Context(), id, *req.Code, nil)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to create new version")
				return
			}
		}

		// If metadata is provided, update the function
		if req.Name != nil || req.Description != nil {
			err := db.UpdateFunction(r.Context(), id, req)
			if err != nil {
				writeError(w, http.StatusNotFound, "Function not found")
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

// DeleteFunctionHandler returns a handler for deleting functions
func DeleteFunctionHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := db.DeleteFunction(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to delete function")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateEnvVarsHandler returns a handler for updating environment variables
func UpdateEnvVarsHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var req UpdateEnvVarsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Get the current function
		_, err := db.GetFunction(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Function not found")
			return
		}

		// Update the environment variables
		if err := db.UpdateFunctionEnvVars(r.Context(), id, req.EnvVars); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update environment variables")
			return
		}

		// Get the active version to return
		activeVersion, err := db.GetActiveVersion(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get active version")
			return
		}

		writeJSON(w, http.StatusOK, activeVersion)
	}
}

// ListVersionsHandler returns a handler for listing function versions
func ListVersionsHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		params := parsePaginationParams(r)

		// Verify function exists
		if _, err := db.GetFunction(r.Context(), id); err != nil {
			writeError(w, http.StatusNotFound, "Function not found")
			return
		}

		versions, total, err := db.ListVersions(r.Context(), id, params)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list versions")
			return
		}

		params = params.Normalize()
		resp := PaginatedVersionsResponse{
			Versions: versions,
			Pagination: PaginationInfo{
				Total:  total,
				Limit:  params.Limit,
				Offset: params.Offset,
			},
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// GetVersionHandler returns a handler for getting a specific version
func GetVersionHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		versionStr := r.PathValue("version")

		// Parse version number
		versionNum, err := strconv.Atoi(versionStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid version number")
			return
		}

		version, err := db.GetVersion(r.Context(), id, versionNum)
		if err != nil {
			writeError(w, http.StatusNotFound, "Version not found")
			return
		}

		writeJSON(w, http.StatusOK, version)
	}
}

// ActivateVersionHandler returns a handler for activating a version
func ActivateVersionHandler(db DB) http.HandlerFunc {
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
		if err := db.ActivateVersion(r.Context(), id, versionNum); err != nil {
			writeError(w, http.StatusNotFound, "Version not found")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// GetVersionDiffHandler returns a handler for getting diff between versions
func GetVersionDiffHandler(db DB) http.HandlerFunc {
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
		version1, err := db.GetVersion(r.Context(), id, v1)
		if err != nil {
			writeError(w, http.StatusNotFound, "Version v1 not found")
			return
		}

		version2, err := db.GetVersion(r.Context(), id, v2)
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
func ListExecutionsHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		params := parsePaginationParams(r)

		// Verify function exists
		if _, err := db.GetFunction(r.Context(), id); err != nil {
			writeError(w, http.StatusNotFound, "Function not found")
			return
		}

		executions, total, err := db.ListExecutions(r.Context(), id, params)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to list executions")
			return
		}

		params = params.Normalize()
		resp := PaginatedExecutionsResponse{
			Executions: executions,
			Pagination: PaginationInfo{
				Total:  total,
				Limit:  params.Limit,
				Offset: params.Offset,
			},
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// GetExecutionHandler returns a handler for getting a specific execution
func GetExecutionHandler(db DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		execution, err := db.GetExecution(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Execution not found")
			return
		}

		writeJSON(w, http.StatusOK, execution)
	}
}

// GetExecutionLogsHandler returns a handler for getting execution logs
func GetExecutionLogsHandler(db DB, appLogger logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		params := parsePaginationParams(r)

		// Get the execution
		execution, err := db.GetExecution(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "Execution not found")
			return
		}

		// Get the logs for this execution from the logger (using execution_id as namespace)
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
			Pagination: PaginationInfo{
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
		_, err := deps.DB.GetFunction(r.Context(), functionID)
		if err != nil {
			writeError(w, http.StatusNotFound, "Function not found")
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
		}

		// Create execution record
		execution := Execution{
			ID:                executionID,
			FunctionID:        functionID,
			FunctionVersionID: version.ID,
			Status:            ExecutionStatusPending,
		}

		_, err = deps.DB.CreateExecution(r.Context(), execution)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create execution record")
			return
		}

		// Prepare runner dependencies
		runnerDeps := runner.Dependencies{
			Logger:  deps.Logger,
			KV:      deps.KVStore,
			Env:     deps.EnvStore,
			HTTP:    deps.HTTPClient,
			Timeout: deps.ExecutionTimeout,
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

		// Update execution status
		var errorMsg *string
		status := ExecutionStatusSuccess
		if runErr != nil {
			status = ExecutionStatusError
			errStr := runErr.Error()
			errorMsg = &errStr
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

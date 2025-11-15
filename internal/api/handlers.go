package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dimiro1/faas-go/internal/diff"
	"github.com/rs/xid"
)

// Handler holds dependencies for API handlers
type Handler struct {
	db DB
}

// NewHandler creates a new Handler instance
func NewHandler(db DB) *Handler {
	return &Handler{
		db: db,
	}
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but don't change response since headers are already written
		_ = err
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
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

// generateDiff creates a diff between two code strings and returns it in our DiffLine format
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

// Function Management Handlers

// CreateFunction handles POST /api/functions
func (h *Handler) CreateFunction(w http.ResponseWriter, r *http.Request) {
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

	createdFn, err := h.db.CreateFunction(r.Context(), fn)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create function")
		return
	}

	// Create the first version with the code
	version, err := h.db.CreateVersion(r.Context(), createdFn.ID, req.Code, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create function version")
		return
	}

	resp := FunctionWithActiveVersion{
		Function:      createdFn,
		ActiveVersion: version,
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListFunctions handles GET /api/functions
func (h *Handler) ListFunctions(w http.ResponseWriter, r *http.Request) {
	params := parsePaginationParams(r)

	functions, total, err := h.db.ListFunctions(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list functions")
		return
	}

	// Build response with active versions
	functionsWithVersions := make([]FunctionWithActiveVersion, 0, len(functions))
	for _, fn := range functions {
		activeVersion, err := h.db.GetActiveVersion(r.Context(), fn.ID)
		if err != nil {
			// If no active version exists, skip this function or handle gracefully
			continue
		}

		functionsWithVersions = append(functionsWithVersions, FunctionWithActiveVersion{
			Function:      fn,
			ActiveVersion: activeVersion,
		})
	}

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

// GetFunction handles GET /api/functions/{id}
func (h *Handler) GetFunction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	fn, err := h.db.GetFunction(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Function not found")
		return
	}

	activeVersion, err := h.db.GetActiveVersion(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get active version")
		return
	}

	resp := FunctionWithActiveVersion{
		Function:      fn,
		ActiveVersion: activeVersion,
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateFunction handles PUT /api/functions/{id}
func (h *Handler) UpdateFunction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req UpdateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If code is provided, create a new version
	if req.Code != nil {
		_, err := h.db.CreateVersion(r.Context(), id, *req.Code, nil)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create new version")
			return
		}
	}

	// Update function metadata if provided
	if req.Name != nil || req.Description != nil {
		if err := h.db.UpdateFunction(r.Context(), id, req); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update function")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteFunction handles DELETE /api/functions/{id}
func (h *Handler) DeleteFunction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.db.DeleteFunction(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete function")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateEnvVars handles PUT /api/functions/{id}/env
func (h *Handler) UpdateEnvVars(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req UpdateEnvVarsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get the current function
	fn, err := h.db.GetFunction(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Function not found")
		return
	}

	// Update the environment variables
	if err := h.db.UpdateFunctionEnvVars(r.Context(), id, req.EnvVars); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update environment variables")
		return
	}

	// Get the active version to return
	activeVersion, err := h.db.GetActiveVersion(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get active version")
		return
	}

	// Return the active version (env vars are stored on the function, not version)
	_ = fn // Used for validation above
	writeJSON(w, http.StatusOK, activeVersion)
}

// Version Management Handlers

// ListVersions handles GET /api/functions/{id}/versions
func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	params := parsePaginationParams(r)

	// Verify function exists
	if _, err := h.db.GetFunction(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "Function not found")
		return
	}

	versions, total, err := h.db.ListVersions(r.Context(), id, params)
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

// GetVersion handles GET /api/functions/{id}/versions/{version}
func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	versionStr := r.PathValue("version")

	// Parse version number
	versionNum, err := strconv.Atoi(versionStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid version number")
		return
	}

	version, err := h.db.GetVersion(r.Context(), id, versionNum)
	if err != nil {
		writeError(w, http.StatusNotFound, "Version not found")
		return
	}

	writeJSON(w, http.StatusOK, version)
}

// ActivateVersion handles POST /api/functions/{id}/versions/{version}/activate
func (h *Handler) ActivateVersion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	versionStr := r.PathValue("version")

	// Parse version number
	versionNum, err := strconv.Atoi(versionStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid version number")
		return
	}

	// Activate the version
	if err := h.db.ActivateVersion(r.Context(), id, versionNum); err != nil {
		writeError(w, http.StatusNotFound, "Version not found")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetVersionDiff handles GET /api/functions/{id}/diff/{v1}/{v2}
func (h *Handler) GetVersionDiff(w http.ResponseWriter, r *http.Request) {
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
	version1, err := h.db.GetVersion(r.Context(), id, v1)
	if err != nil {
		writeError(w, http.StatusNotFound, "Version v1 not found")
		return
	}

	version2, err := h.db.GetVersion(r.Context(), id, v2)
	if err != nil {
		writeError(w, http.StatusNotFound, "Version v2 not found")
		return
	}

	// Generate the diff using our utility function
	diff := generateDiff(version1.Code, version2.Code, v1, v2)

	writeJSON(w, http.StatusOK, diff)
}

// Execution Handlers

// ListExecutions handles GET /api/functions/{id}/executions
func (h *Handler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	params := parsePaginationParams(r)

	// Verify function exists
	if _, err := h.db.GetFunction(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "Function not found")
		return
	}

	executions, total, err := h.db.ListExecutions(r.Context(), id, params)
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

// GetExecution handles GET /api/executions/{id}
func (h *Handler) GetExecution(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	execution, err := h.db.GetExecution(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Execution not found")
		return
	}

	writeJSON(w, http.StatusOK, execution)
}

// GetExecutionLogs handles GET /api/executions/{id}/logs
func (h *Handler) GetExecutionLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	params := parsePaginationParams(r)

	// Get the execution
	execution, err := h.db.GetExecution(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Execution not found")
		return
	}

	// Get the logs for this execution
	logs, total, err := h.db.GetExecutionLogs(r.Context(), id, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get execution logs")
		return
	}

	params = params.Normalize()
	resp := PaginatedExecutionWithLogs{
		Execution: execution,
		Logs:      logs,
		Pagination: PaginationInfo{
			Total:  total,
			Limit:  params.Limit,
			Offset: params.Offset,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

// Runtime Execution Handlers

// ExecuteFunction handles GET/POST/PUT/DELETE /fn/{function_id}
func (h *Handler) ExecuteFunction(w http.ResponseWriter, r *http.Request) {
	functionID := r.PathValue("function_id")

	// Add custom headers
	w.Header().Set("X-Function-Id", functionID)
	w.Header().Set("X-Function-Version-Id", "ver_123")
	w.Header().Set("X-Execution-Id", generateID())
	w.Header().Set("X-Execution-Duration-Ms", "42")

	// Dummy response based on method
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message": "Function executed successfully", "method": "` + r.Method + `"}`))
}

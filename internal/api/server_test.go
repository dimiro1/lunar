package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Helper function to create a test function in the database with an initial version
func createTestFunction(t *testing.T, db DB) Function {
	t.Helper()
	desc := "Test function"
	fn := Function{
		ID:          "func_test_123",
		Name:        "test-function",
		Description: &desc,
		EnvVars:     map[string]string{"KEY": "value"},
	}
	created, err := db.CreateFunction(context.Background(), fn)
	if err != nil {
		t.Fatalf("failed to create test function: %v", err)
	}

	// Create an initial version for the function
	_, err = db.CreateVersion(context.Background(), created.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend", nil)
	if err != nil {
		t.Fatalf("failed to create initial version: %v", err)
	}

	return created
}

// Helper function to create a test version
func createTestVersion(t *testing.T, db DB, functionID string, code string) FunctionVersion {
	t.Helper()
	version, err := db.CreateVersion(context.Background(), functionID, code, nil)
	if err != nil {
		t.Fatalf("failed to create test version: %v", err)
	}
	return version
}

// Helper function to create a test execution
func createTestExecution(t *testing.T, db DB, functionID, versionID string) Execution {
	t.Helper()
	exec := Execution{
		ID:                "exec_test_123",
		FunctionID:        functionID,
		FunctionVersionID: versionID,
		Status:            ExecutionStatusSuccess,
	}
	created, err := db.CreateExecution(context.Background(), exec)
	if err != nil {
		t.Fatalf("failed to create test execution: %v", err)
	}
	return created
}

func TestCreateFunction(t *testing.T) {
	server := NewServer(NewMemoryDB())

	reqBody := CreateFunctionRequest{
		Name: "test-function",
		Code: "function handler(ctx, event)\n  return {statusCode = 200}\nend",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/functions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp FunctionWithActiveVersion
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Name != reqBody.Name {
		t.Errorf("expected name %q, got %q", reqBody.Name, resp.Name)
	}

	if resp.ActiveVersion.Version != 1 {
		t.Errorf("expected version 1, got %d", resp.ActiveVersion.Version)
	}
}

func TestListFunctions(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function first
	createTestFunction(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/functions", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp PaginatedFunctionsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Functions) == 0 {
		t.Error("expected at least one function")
	}
}

func TestGetFunction(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function first
	fn := createTestFunction(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/functions/"+fn.ID, nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp FunctionWithActiveVersion
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != fn.ID {
		t.Errorf("expected ID %s, got %q", fn.ID, resp.ID)
	}
}

func TestUpdateFunction(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function first
	fn := createTestFunction(t, db)

	name := "updated-name"
	reqBody := UpdateFunctionRequest{
		Name: &name,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/functions/"+fn.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestDeleteFunction(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function first
	fn := createTestFunction(t, db)

	req := httptest.NewRequest(http.MethodDelete, "/api/functions/"+fn.ID, nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestListVersions(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function and version
	fn := createTestFunction(t, db)
	createTestVersion(t, db, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")

	req := httptest.NewRequest(http.MethodGet, "/api/functions/"+fn.ID+"/versions", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp PaginatedVersionsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Versions) == 0 {
		t.Error("expected at least one version")
	}
}

func TestGetVersion(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function (which creates version 1) and another version (version 2)
	fn := createTestFunction(t, db)
	ver := createTestVersion(t, db, fn.ID, "function handler(ctx, event)\n  return {statusCode = 201}\nend")

	// Request version 2 which we just created
	req := httptest.NewRequest(http.MethodGet, "/api/functions/"+fn.ID+"/versions/2", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp FunctionVersion
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != ver.ID {
		t.Errorf("expected version ID %s, got %s", ver.ID, resp.ID)
	}

	if resp.Version != 2 {
		t.Errorf("expected version number 2, got %d", resp.Version)
	}
}

func TestActivateVersion(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function and two versions
	fn := createTestFunction(t, db)
	createTestVersion(t, db, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	createTestVersion(t, db, fn.ID, "function handler(ctx, event)\n  return {statusCode = 201}\nend")

	req := httptest.NewRequest(http.MethodPost, "/api/functions/"+fn.ID+"/versions/1/activate", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetVersionDiff(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function and two versions with different code
	fn := createTestFunction(t, db)
	createTestVersion(t, db, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	createTestVersion(t, db, fn.ID, "function handler(ctx, event)\n  return {statusCode = 201}\nend")

	req := httptest.NewRequest(http.MethodGet, "/api/functions/"+fn.ID+"/diff/1/2", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp VersionDiffResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Diff) == 0 {
		t.Error("expected at least one diff line")
	}
}

func TestUpdateEnvVars(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function first
	fn := createTestFunction(t, db)

	reqBody := UpdateEnvVarsRequest{
		EnvVars: map[string]string{
			"API_KEY": "secret-123",
			"DEBUG":   "true",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/functions/"+fn.ID+"/env", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestListExecutions(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function and execution
	fn := createTestFunction(t, db)
	ver := createTestVersion(t, db, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	createTestExecution(t, db, fn.ID, ver.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/functions/"+fn.ID+"/executions", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp PaginatedExecutionsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestGetExecution(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function, version and execution
	fn := createTestFunction(t, db)
	ver := createTestVersion(t, db, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	exec := createTestExecution(t, db, fn.ID, ver.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/executions/"+exec.ID, nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp Execution
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestGetExecutionLogs(t *testing.T) {
	db := NewMemoryDB()
	server := NewServer(db)

	// Create a test function, version, execution and log
	fn := createTestFunction(t, db)
	ver := createTestVersion(t, db, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	exec := createTestExecution(t, db, fn.ID, ver.ID)

	// Create a log entry for the execution
	log := LogEntry{
		ID:          "log_test_123",
		ExecutionID: exec.ID,
		Level:       LogLevelInfo,
		Message:     "Test log message",
	}
	if err := db.CreateLog(context.Background(), log); err != nil {
		t.Fatalf("failed to create test log: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/executions/"+exec.ID+"/logs", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp PaginatedExecutionWithLogs
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Logs) == 0 {
		t.Error("expected at least one log entry")
	}
}

func TestExecuteFunction(t *testing.T) {
	server := NewServer(NewMemoryDB())

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/fn/func_123"},
		{http.MethodPost, "/fn/func_123"},
		{http.MethodPut, "/fn/func_123"},
		{http.MethodDelete, "/fn/func_123"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			server.Handler().ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			// Check custom headers
			if w.Header().Get("X-Function-Id") == "" {
				t.Error("expected X-Function-Id header")
			}
			if w.Header().Get("X-Execution-Id") == "" {
				t.Error("expected X-Execution-Id header")
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	server := NewServer(NewMemoryDB())

	req := httptest.NewRequest(http.MethodOptions, "/api/functions", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS headers")
	}
}

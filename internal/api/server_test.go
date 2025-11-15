package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimiro1/faas-go/internal/env"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
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

// Helper function to create a test server with full configuration
func createTestServer(db DB) *Server {
	return NewServer(ServerConfig{
		DB:         db,
		Logger:     logger.NewMemoryLogger(),
		KVStore:    kv.NewMemoryStore(),
		EnvStore:   env.NewMemoryStore(),
		HTTPClient: internalhttp.NewDefaultClient(),
	})
}

func TestCreateFunction(t *testing.T) {
	server := createTestServer(NewMemoryDB())

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
	server := createTestServer(db)

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
	server := createTestServer(db)

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
	server := createTestServer(db)

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
	server := createTestServer(db)

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
	server := createTestServer(db)

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
	server := createTestServer(db)

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
	server := createTestServer(db)

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
	server := createTestServer(db)

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
	server := createTestServer(db)

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
	server := createTestServer(db)

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
	server := createTestServer(db)

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
	memLogger := logger.NewMemoryLogger()

	server := NewServer(ServerConfig{
		DB:         db,
		Logger:     memLogger,
		KVStore:    kv.NewMemoryStore(),
		EnvStore:   env.NewMemoryStore(),
		HTTPClient: internalhttp.NewDefaultClient(),
	})

	// Create a test function, version, execution
	fn := createTestFunction(t, db)
	ver := createTestVersion(t, db, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	exec := createTestExecution(t, db, fn.ID, ver.ID)

	// Create a log entry for the execution using the logger
	memLogger.Info(exec.ID, "Test log message")

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
	t.Run("success with simple response", func(t *testing.T) {
		db := NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         db,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
		})

		fn := createTestFunction(t, db)
		_, err := db.CreateVersion(context.Background(), fn.ID, `
function handler(ctx, event)
  return {
    statusCode = 200,
    body = '{"message": "success"}'
  }
end
`, nil)
		if err != nil {
			t.Fatalf("Failed to create version: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID, nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		if w.Header().Get("X-Function-Id") != fn.ID {
			t.Errorf("expected X-Function-Id %s, got %s", fn.ID, w.Header().Get("X-Function-Id"))
		}
		if w.Header().Get("X-Execution-Id") == "" {
			t.Error("expected X-Execution-Id header")
		}
		if w.Header().Get("X-Execution-Duration-Ms") == "" {
			t.Error("expected X-Execution-Duration-Ms header")
		}
	})

	t.Run("success with request body", func(t *testing.T) {
		db := NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         db,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
		})

		fn := createTestFunction(t, db)
		_, err := db.CreateVersion(context.Background(), fn.ID, `
function handler(ctx, event)
  return {
    statusCode = 200,
    body = event.body
  }
end
`, nil)
		if err != nil {
			t.Fatalf("Failed to create version: %v", err)
		}

		requestBody := `{"name": "test"}`
		req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID, bytes.NewReader([]byte(requestBody)))
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if w.Body.String() != requestBody {
			t.Errorf("expected body %s, got %s", requestBody, w.Body.String())
		}
	})

	t.Run("success with custom status code", func(t *testing.T) {
		db := NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         db,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
		})

		fn := createTestFunction(t, db)
		_, err := db.CreateVersion(context.Background(), fn.ID, `
function handler(ctx, event)
  return {
    statusCode = 201,
    body = '{"created": true}'
  }
end
`, nil)
		if err != nil {
			t.Fatalf("Failed to create version: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID, nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}
	})

	t.Run("success with custom headers", func(t *testing.T) {
		db := NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         db,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
		})

		fn := createTestFunction(t, db)
		_, err := db.CreateVersion(context.Background(), fn.ID, `
function handler(ctx, event)
  return {
    statusCode = 200,
    headers = {
      ["X-Custom-Header"] = "custom-value",
      ["Content-Type"] = "text/plain"
    },
    body = 'hello'
  }
end
`, nil)
		if err != nil {
			t.Fatalf("Failed to create version: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/fn/"+fn.ID, nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if w.Header().Get("X-Custom-Header") != "custom-value" {
			t.Errorf("expected X-Custom-Header 'custom-value', got %s", w.Header().Get("X-Custom-Header"))
		}
	})

	t.Run("error with syntax error in lua code", func(t *testing.T) {
		db := NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         db,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
		})

		fn := createTestFunction(t, db)
		_, err := db.CreateVersion(context.Background(), fn.ID, `
function handler(ctx, event)
  return {
    statusCode = 200
    -- missing comma
    body = 'test'
  }
end
`, nil)
		if err != nil {
			t.Fatalf("Failed to create version: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID, nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("error with runtime error in lua code", func(t *testing.T) {
		db := NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         db,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
		})

		fn := createTestFunction(t, db)
		_, err := db.CreateVersion(context.Background(), fn.ID, `
function handler(ctx, event)
  error("Something went wrong!")
end
`, nil)
		if err != nil {
			t.Fatalf("Failed to create version: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID, nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}

		var resp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["error"] != "Function execution failed" {
			t.Errorf("expected generic error message, got %q", resp["error"])
		}

		if w.Header().Get("X-Execution-Id") == "" {
			t.Error("expected X-Execution-Id header even on error")
		}
	})

	t.Run("error with function not found", func(t *testing.T) {
		db := NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         db,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
		})

		req := httptest.NewRequest(http.MethodPost, "/fn/nonexistent", nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("error with no active version", func(t *testing.T) {
		db := NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         db,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
		})

		fn := Function{
			ID:          "test-no-version",
			Name:        "test",
			Description: nil,
			EnvVars:     map[string]string{},
		}
		_, err := db.CreateFunction(context.Background(), fn)
		if err != nil {
			t.Fatalf("Failed to create function: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID, nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("different HTTP methods", func(t *testing.T) {
		db := NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         db,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
		})

		fn := createTestFunction(t, db)
		_, err := db.CreateVersion(context.Background(), fn.ID, `
function handler(ctx, event)
  return {
    statusCode = 200,
    body = '{"method": "' .. event.method .. '"}'
  }
end
`, nil)
		if err != nil {
			t.Fatalf("Failed to create version: %v", err)
		}

		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				req := httptest.NewRequest(method, "/fn/"+fn.ID, nil)
				w := httptest.NewRecorder()

				server.Handler().ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("expected status 200 for %s, got %d", method, w.Code)
				}
			})
		}
	})
}

func TestCORSMiddleware(t *testing.T) {
	server := createTestServer(NewMemoryDB())

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

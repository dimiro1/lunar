package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dimiro1/lunar/internal/env"
	internalhttp "github.com/dimiro1/lunar/internal/http"
	"github.com/dimiro1/lunar/internal/kv"
	"github.com/dimiro1/lunar/internal/logger"
	"github.com/dimiro1/lunar/internal/store"
)

// Helper function to create a test function in the database with an initial version
func createTestFunction(t *testing.T, database store.DB) store.Function {
	t.Helper()
	desc := "Test function"
	fn := store.Function{
		ID:          "func_test_123",
		Name:        "test-function",
		Description: &desc,
		EnvVars:     map[string]string{"KEY": "value"},
	}
	created, err := database.CreateFunction(context.Background(), fn)
	if err != nil {
		t.Fatalf("failed to create test function: %v", err)
	}

	// Create an initial version for the function
	_, err = database.CreateVersion(context.Background(), created.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend", nil)
	if err != nil {
		t.Fatalf("failed to create initial version: %v", err)
	}

	return created
}

// Helper function to create a test version
func createTestVersion(t *testing.T, database store.DB, functionID string, code string) store.FunctionVersion {
	t.Helper()
	version, err := database.CreateVersion(context.Background(), functionID, code, nil)
	if err != nil {
		t.Fatalf("failed to create test version: %v", err)
	}
	return version
}

// Helper function to create a test execution
func createTestExecution(t *testing.T, database store.DB, functionID, versionID string) store.Execution {
	t.Helper()
	exec := store.Execution{
		ID:                "exec_test_123",
		FunctionID:        functionID,
		FunctionVersionID: versionID,
		Status:            store.ExecutionStatusSuccess,
	}
	created, err := database.CreateExecution(context.Background(), exec)
	if err != nil {
		t.Fatalf("failed to create test execution: %v", err)
	}
	return created
}

// Helper function to create a test server with full configuration
func createTestServer(database store.DB) *Server {
	return NewServer(ServerConfig{
		DB:         database,
		Logger:     logger.NewMemoryLogger(),
		KVStore:    kv.NewMemoryStore(),
		EnvStore:   env.NewMemoryStore(),
		HTTPClient: internalhttp.NewDefaultClient(),
		APIKey:     "test-api-key",
		BaseURL:    "http://localhost:8080",
	})
}

// Helper function to make authenticated API requests
func makeAuthRequest(method, path string, body []byte) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Authorization", "Bearer test-api-key")
	return req
}

func TestDocsPage(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("expected Content-Type text/html; charset=utf-8, got %q", ct)
	}

	if w.Body.Len() == 0 {
		t.Fatal("expected non-empty response body")
	}
}

func TestOpenAPISpecEndpoint(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	req := httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Fatalf("expected Content-Type application/yaml, got %q", ct)
	}

	if !bytes.Equal(w.Body.Bytes(), openAPISpec) {
		t.Fatal("expected response body to match embedded OpenAPI spec")
	}
}

func TestCreateFunction(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	reqBody := CreateFunctionRequest{
		Name: "test-function",
		Code: "function handler(ctx, event)\n  return {statusCode = 200}\nend",
	}

	body, _ := json.Marshal(reqBody)
	req := makeAuthRequest(http.MethodPost, "/api/functions", body)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp store.FunctionWithActiveVersion
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
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function first
	createTestFunction(t, database)

	req := makeAuthRequest(http.MethodGet, "/api/functions", nil)
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
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function first
	fn := createTestFunction(t, database)

	req := makeAuthRequest(http.MethodGet, "/api/functions/"+fn.ID, nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp store.FunctionWithActiveVersion
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != fn.ID {
		t.Errorf("expected ID %s, got %q", fn.ID, resp.ID)
	}
}

func TestUpdateFunction(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function first
	fn := createTestFunction(t, database)

	name := "updated-name"
	reqBody := store.UpdateFunctionRequest{
		Name: &name,
	}

	body, _ := json.Marshal(reqBody)
	req := makeAuthRequest(http.MethodPut, "/api/functions/"+fn.ID, body)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestDeleteFunction(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function first
	fn := createTestFunction(t, database)

	req := makeAuthRequest(http.MethodDelete, "/api/functions/"+fn.ID, nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestListVersions(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function and version
	fn := createTestFunction(t, database)
	createTestVersion(t, database, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")

	req := makeAuthRequest(http.MethodGet, "/api/functions/"+fn.ID+"/versions", nil)
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
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function (which creates version 1) and another version (version 2)
	fn := createTestFunction(t, database)
	ver := createTestVersion(t, database, fn.ID, "function handler(ctx, event)\n  return {statusCode = 201}\nend")

	// Request version 2 which we just created
	req := makeAuthRequest(http.MethodGet, "/api/functions/"+fn.ID+"/versions/2", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp store.FunctionVersion
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
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function and two versions
	fn := createTestFunction(t, database)
	createTestVersion(t, database, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	createTestVersion(t, database, fn.ID, "function handler(ctx, event)\n  return {statusCode = 201}\nend")

	req := makeAuthRequest(http.MethodPost, "/api/functions/"+fn.ID+"/versions/1/activate", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetVersionDiff(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function and two versions with different code
	fn := createTestFunction(t, database)
	createTestVersion(t, database, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	createTestVersion(t, database, fn.ID, "function handler(ctx, event)\n  return {statusCode = 201}\nend")

	req := makeAuthRequest(http.MethodGet, "/api/functions/"+fn.ID+"/diff/1/2", nil)
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
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function first
	fn := createTestFunction(t, database)

	reqBody := UpdateEnvVarsRequest{
		EnvVars: map[string]string{
			"API_KEY": "secret-123",
			"DEBUG":   "true",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := makeAuthRequest(http.MethodPut, "/api/functions/"+fn.ID+"/env", body)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestListExecutions(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function and execution
	fn := createTestFunction(t, database)
	ver := createTestVersion(t, database, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	createTestExecution(t, database, fn.ID, ver.ID)

	req := makeAuthRequest(http.MethodGet, "/api/functions/"+fn.ID+"/executions", nil)
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
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function, version and execution
	fn := createTestFunction(t, database)
	ver := createTestVersion(t, database, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	exec := createTestExecution(t, database, fn.ID, ver.ID)

	req := makeAuthRequest(http.MethodGet, "/api/executions/"+exec.ID, nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp store.Execution
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestGetExecutionLogs(t *testing.T) {
	database := store.NewMemoryDB()
	memLogger := logger.NewMemoryLogger()

	server := NewServer(ServerConfig{
		DB:         database,
		Logger:     memLogger,
		KVStore:    kv.NewMemoryStore(),
		EnvStore:   env.NewMemoryStore(),
		HTTPClient: internalhttp.NewDefaultClient(),
		APIKey:     "test-api-key",
	})

	// Create a test function, version, execution
	fn := createTestFunction(t, database)
	ver := createTestVersion(t, database, fn.ID, "function handler(ctx, event)\n  return {statusCode = 200}\nend")
	exec := createTestExecution(t, database, fn.ID, ver.ID)

	// Create a log entry for the execution using the logger
	memLogger.Info(exec.ID, "Test log message")

	req := makeAuthRequest(http.MethodGet, "/api/executions/"+exec.ID+"/logs", nil)
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
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		fn := createTestFunction(t, database)
		_, err := database.CreateVersion(context.Background(), fn.ID, `
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
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		fn := createTestFunction(t, database)
		_, err := database.CreateVersion(context.Background(), fn.ID, `
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
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		fn := createTestFunction(t, database)
		_, err := database.CreateVersion(context.Background(), fn.ID, `
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
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		fn := createTestFunction(t, database)
		_, err := database.CreateVersion(context.Background(), fn.ID, `
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

		if w.Header().Get("Content-Type") != "text/plain" {
			t.Errorf("expected Content-Type 'text/plain', got %s", w.Header().Get("Content-Type"))
		}
	})

	t.Run("success with html content type", func(t *testing.T) {
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		fn := createTestFunction(t, database)
		_, err := database.CreateVersion(context.Background(), fn.ID, `
function handler(ctx, event)
  return {
    statusCode = 200,
    headers = {
      ["Content-Type"] = "text/html"
    },
    body = '<html><body><h1>Hello World</h1></body></html>'
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

		if w.Header().Get("Content-Type") != "text/html" {
			t.Errorf("expected Content-Type 'text/html', got %s", w.Header().Get("Content-Type"))
		}

		expectedBody := "<html><body><h1>Hello World</h1></body></html>"
		if w.Body.String() != expectedBody {
			t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
		}
	})

	t.Run("default content type when not specified", func(t *testing.T) {
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		fn := createTestFunction(t, database)
		_, err := database.CreateVersion(context.Background(), fn.ID, `
function handler(ctx, event)
  return {
    statusCode = 200,
    body = '{"message": "hello"}'
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

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %s", w.Header().Get("Content-Type"))
		}
	})

	t.Run("error with syntax error in lua code", func(t *testing.T) {
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		fn := createTestFunction(t, database)
		_, err := database.CreateVersion(context.Background(), fn.ID, `
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
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		fn := createTestFunction(t, database)
		_, err := database.CreateVersion(context.Background(), fn.ID, `
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
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		req := httptest.NewRequest(http.MethodPost, "/fn/nonexistent", nil)
		w := httptest.NewRecorder()

		server.Handler().ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("error with no active version", func(t *testing.T) {
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		fn := store.Function{
			ID:          "test-no-version",
			Name:        "test",
			Description: nil,
			EnvVars:     map[string]string{},
		}
		_, err := database.CreateFunction(context.Background(), fn)
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
		database := store.NewMemoryDB()
		server := NewServer(ServerConfig{
			DB:         database,
			Logger:     logger.NewMemoryLogger(),
			KVStore:    kv.NewMemoryStore(),
			EnvStore:   env.NewMemoryStore(),
			HTTPClient: internalhttp.NewDefaultClient(),
			APIKey:     "test-api-key",
		})

		fn := createTestFunction(t, database)
		_, err := database.CreateVersion(context.Background(), fn.ID, `
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

func TestUpdateFunction_ToggleDisabled(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function first
	fn := createTestFunction(t, database)

	// Disable the function
	disabled := true
	reqBody := store.UpdateFunctionRequest{
		Disabled: &disabled,
	}

	body, _ := json.Marshal(reqBody)
	req := makeAuthRequest(http.MethodPut, "/api/functions/"+fn.ID, body)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify the function is disabled
	updated, err := database.GetFunction(context.Background(), fn.ID)
	if err != nil {
		t.Fatalf("failed to get updated function: %v", err)
	}

	if !updated.Disabled {
		t.Error("expected function to be disabled")
	}

	// Enable the function again
	enabled := false
	reqBody2 := store.UpdateFunctionRequest{
		Disabled: &enabled,
	}

	body2, _ := json.Marshal(reqBody2)
	req2 := makeAuthRequest(http.MethodPut, "/api/functions/"+fn.ID, body2)
	w2 := httptest.NewRecorder()

	server.Handler().ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}

	// Verify the function is enabled
	reenabled, err := database.GetFunction(context.Background(), fn.ID)
	if err != nil {
		t.Fatalf("failed to get re-enabled function: %v", err)
	}

	if reenabled.Disabled {
		t.Error("expected function to be enabled")
	}
}

func TestUpdateFunction_RetentionDays(t *testing.T) {
	database := store.NewMemoryDB()
	server := createTestServer(database)

	// Create a test function first
	fn := createTestFunction(t, database)

	// Update retention days to 30
	retentionDays := 30
	reqBody := store.UpdateFunctionRequest{
		RetentionDays: &retentionDays,
	}

	body, _ := json.Marshal(reqBody)
	req := makeAuthRequest(http.MethodPut, "/api/functions/"+fn.ID, body)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify the retention days was updated
	updated, err := database.GetFunction(context.Background(), fn.ID)
	if err != nil {
		t.Fatalf("failed to get updated function: %v", err)
	}

	if updated.RetentionDays == nil {
		t.Fatal("expected retention_days to be set")
	}

	if *updated.RetentionDays != 30 {
		t.Errorf("expected retention_days to be 30, got %d", *updated.RetentionDays)
	}
}

func TestExecuteFunction_DisabledFunction(t *testing.T) {
	database := store.NewMemoryDB()
	server := NewServer(ServerConfig{
		DB:         database,
		Logger:     logger.NewMemoryLogger(),
		KVStore:    kv.NewMemoryStore(),
		EnvStore:   env.NewMemoryStore(),
		HTTPClient: internalhttp.NewDefaultClient(),
		APIKey:     "test-api-key",
	})

	// Create a test function
	fn := createTestFunction(t, database)
	_, err := database.CreateVersion(context.Background(), fn.ID, `
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

	// Disable the function
	disabled := true
	updates := store.UpdateFunctionRequest{
		Disabled: &disabled,
	}
	if err := database.UpdateFunction(context.Background(), fn.ID, updates); err != nil {
		t.Fatalf("Failed to disable function: %v", err)
	}

	// Try to execute the disabled function
	req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID, nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	// Should return 403 Forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["error"] != "Function is disabled" {
		t.Errorf("expected error 'Function is disabled', got %q", resp["error"])
	}
}

func TestCORSMiddleware(t *testing.T) {
	server := createTestServer(store.NewMemoryDB())

	req := makeAuthRequest(http.MethodOptions, "/api/functions", nil)
	w := httptest.NewRecorder()

	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS headers")
	}
}

func TestExecuteFunction_EventJSONStorage(t *testing.T) {
	database := store.NewMemoryDB()
	server := NewServer(ServerConfig{
		DB:         database,
		Logger:     logger.NewMemoryLogger(),
		KVStore:    kv.NewMemoryStore(),
		EnvStore:   env.NewMemoryStore(),
		HTTPClient: internalhttp.NewDefaultClient(),
		APIKey:     "test-api-key",
	})

	fn := createTestFunction(t, database)
	_, err := database.CreateVersion(context.Background(), fn.ID, `
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

	// Create a request with specific headers, query params, and body
	requestBody := `{"test": "data", "number": 42}`
	req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID+"?param1=value1&param2=value2", bytes.NewReader([]byte(requestBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom-Header", "custom-value")
	req.Header.Set("Authorization", "Bearer test-token")

	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Get the execution ID from the response header
	executionID := w.Header().Get("X-Execution-Id")
	if executionID == "" {
		t.Fatal("expected X-Execution-Id header")
	}

	// Retrieve the execution from the database
	execution, err := database.GetExecution(context.Background(), executionID)
	if err != nil {
		t.Fatalf("Failed to get execution: %v", err)
	}

	// Verify event JSON was stored
	if execution.EventJSON == nil {
		t.Fatal("Expected EventJSON to be stored")
	}

	// Parse and verify the event JSON content
	var eventData map[string]any
	if err := json.Unmarshal([]byte(*execution.EventJSON), &eventData); err != nil {
		t.Fatalf("Failed to parse event JSON: %v", err)
	}

	// Verify method
	if method, ok := eventData["method"].(string); !ok || method != "POST" {
		t.Errorf("Expected method POST, got %v", eventData["method"])
	}

	// Verify path
	if path, ok := eventData["path"].(string); !ok || path != "/fn/"+fn.ID {
		t.Errorf("Expected path /fn/%s, got %v", fn.ID, eventData["path"])
	}

	// Verify body is present (JSON order may vary, so just check it's not empty)
	if body, ok := eventData["body"].(string); !ok || body == "" {
		t.Errorf("Expected body to be present, got %v", eventData["body"])
	}

	// Verify headers are present
	headers, ok := eventData["headers"].(map[string]any)
	if !ok {
		t.Fatal("Expected headers to be present")
	}

	if contentType, ok := headers["Content-Type"].(string); !ok || contentType != "application/json" {
		t.Errorf("Expected Content-Type header, got %v", headers["Content-Type"])
	}

	if customHeader, ok := headers["X-Custom-Header"].(string); !ok || customHeader != "custom-value" {
		t.Errorf("Expected X-Custom-Header, got %v", headers["X-Custom-Header"])
	}

	// Authorization header should be masked now
	if authHeader, ok := headers["Authorization"].(string); !ok || authHeader != "[REDACTED]" {
		t.Errorf("Expected Authorization header to be [REDACTED], got %v", headers["Authorization"])
	}

	// Verify query parameters
	query, ok := eventData["query"].(map[string]any)
	if !ok {
		t.Fatal("Expected query to be present")
	}

	if param1, ok := query["param1"].(string); !ok || param1 != "value1" {
		t.Errorf("Expected param1=value1, got %v", query["param1"])
	}

	if param2, ok := query["param2"].(string); !ok || param2 != "value2" {
		t.Errorf("Expected param2=value2, got %v", query["param2"])
	}
}

func TestExecuteFunction_EventJSONWithDifferentMethods(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			database := store.NewMemoryDB()
			server := NewServer(ServerConfig{
				DB:         database,
				Logger:     logger.NewMemoryLogger(),
				KVStore:    kv.NewMemoryStore(),
				EnvStore:   env.NewMemoryStore(),
				HTTPClient: internalhttp.NewDefaultClient(),
				APIKey:     "test-api-key",
			})

			fn := createTestFunction(t, database)
			_, err := database.CreateVersion(context.Background(), fn.ID, `
function handler(ctx, event)
  return {
    statusCode = 200,
    body = '{"ok": true}'
  }
end
`, nil)
			if err != nil {
				t.Fatalf("Failed to create version: %v", err)
			}

			req := httptest.NewRequest(method, "/fn/"+fn.ID, nil)
			w := httptest.NewRecorder()
			server.Handler().ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}

			executionID := w.Header().Get("X-Execution-Id")
			if executionID == "" {
				t.Fatal("expected X-Execution-Id header")
			}

			execution, err := database.GetExecution(context.Background(), executionID)
			if err != nil {
				t.Fatalf("Failed to get execution: %v", err)
			}

			if execution.EventJSON == nil {
				t.Fatal("Expected EventJSON to be stored")
			}

			var eventData map[string]interface{}
			if err := json.Unmarshal([]byte(*execution.EventJSON), &eventData); err != nil {
				t.Fatalf("Failed to parse event JSON: %v", err)
			}

			if eventMethod, ok := eventData["method"].(string); !ok || eventMethod != method {
				t.Errorf("Expected method %s, got %v", method, eventData["method"])
			}
		})
	}
}

func TestExecuteFunction_SensitiveDataMasking(t *testing.T) {
	database := store.NewMemoryDB()
	server := NewServer(ServerConfig{
		DB:         database,
		Logger:     logger.NewMemoryLogger(),
		KVStore:    kv.NewMemoryStore(),
		EnvStore:   env.NewMemoryStore(),
		HTTPClient: internalhttp.NewDefaultClient(),
		APIKey:     "test-api-key",
	})

	fn := createTestFunction(t, database)
	_, err := database.CreateVersion(context.Background(), fn.ID, `
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

	// Create a request with sensitive headers and body
	requestBody := `{"username":"john","password":"secret123","api_key":"my-secret-api-key"}`
	req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID+"?api_key=secret-query-key&limit=10", bytes.NewReader([]byte(requestBody)))
	req.Header.Set("Authorization", "Bearer secret_token_12345")
	req.Header.Set("Cookie", "auth_token=f150e53a96f53affce140b818440d8aef5e499038cdc2860ff07b3e6f036d6f1")
	req.Header.Set("X-API-Key", "my-api-key-123")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	executionID := w.Header().Get("X-Execution-Id")
	if executionID == "" {
		t.Fatal("expected X-Execution-Id header")
	}

	// Retrieve the execution from the database
	execution, err := database.GetExecution(context.Background(), executionID)
	if err != nil {
		t.Fatalf("Failed to get execution: %v", err)
	}

	if execution.EventJSON == nil {
		t.Fatal("Expected EventJSON to be stored")
	}

	// Parse and verify the event JSON has masked sensitive data
	var eventData map[string]any
	if err := json.Unmarshal([]byte(*execution.EventJSON), &eventData); err != nil {
		t.Fatalf("Failed to parse event JSON: %v", err)
	}

	// Verify sensitive headers are masked
	headers, ok := eventData["headers"].(map[string]any)
	if !ok {
		t.Fatal("Expected headers to be present")
	}

	if auth, ok := headers["Authorization"].(string); !ok || auth != "[REDACTED]" {
		t.Errorf("Expected Authorization header to be [REDACTED], got %v", headers["Authorization"])
	}

	if cookie, ok := headers["Cookie"].(string); !ok || cookie != "[REDACTED]" {
		t.Errorf("Expected Cookie header to be [REDACTED], got %v", headers["Cookie"])
	}

	// X-API-Key header might be stored with different casing
	apiKeyFound := false
	for key, value := range headers {
		if strings.ToLower(key) == "x-api-key" {
			apiKeyFound = true
			if strValue, ok := value.(string); !ok || strValue != "[REDACTED]" {
				t.Errorf("Expected X-API-Key header to be [REDACTED], got %v", value)
			}
			break
		}
	}
	if !apiKeyFound {
		t.Error("Expected X-API-Key header to be present in headers")
	}

	// Verify non-sensitive headers are not masked
	if contentType, ok := headers["Content-Type"].(string); !ok || contentType != "application/json" {
		t.Errorf("Expected Content-Type header to be unchanged, got %v", headers["Content-Type"])
	}

	// Verify sensitive query params are masked
	query, ok := eventData["query"].(map[string]any)
	if !ok {
		t.Fatal("Expected query to be present")
	}

	if apiKeyQuery, ok := query["api_key"].(string); !ok || apiKeyQuery != "[REDACTED]" {
		t.Errorf("Expected api_key query param to be [REDACTED], got %v", query["api_key"])
	}

	// Verify non-sensitive query params are not masked
	if limit, ok := query["limit"].(string); !ok || limit != "10" {
		t.Errorf("Expected limit query param to be unchanged, got %v", query["limit"])
	}

	// Verify sensitive body fields are masked
	body, ok := eventData["body"].(string)
	if !ok {
		t.Fatal("Expected body to be present")
	}

	// Parse the body JSON
	var bodyData map[string]any
	if err := json.Unmarshal([]byte(body), &bodyData); err != nil {
		t.Fatalf("Failed to parse body JSON: %v", err)
	}

	if password, ok := bodyData["password"].(string); !ok || password != "[REDACTED]" {
		t.Errorf("Expected password field to be [REDACTED], got %v", bodyData["password"])
	}

	if apiKeyBody, ok := bodyData["api_key"].(string); !ok || apiKeyBody != "[REDACTED]" {
		t.Errorf("Expected api_key field to be [REDACTED], got %v", bodyData["api_key"])
	}

	// Verify non-sensitive body fields are not masked
	if username, ok := bodyData["username"].(string); !ok || username != "john" {
		t.Errorf("Expected username field to be unchanged, got %v", bodyData["username"])
	}
}

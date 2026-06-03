package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dimiro1/lunar/internal/services/env"
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
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

// newGraphQLProbe builds an unauthenticated POST /graphql request running a
// trivial query. It is used by the auth-middleware tests to assert that a
// caller is (or isn't) allowed through to a protected route — /graphql being
// the canonical auth-protected endpoint now that the REST management API is
// gone. Callers add the credential (bearer token or cookie) under test.
func newGraphQLProbe() *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader([]byte(`{"query":"{ __typename }"}`)))
	req.Header.Set("Content-Type", "application/json")
	return req
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

	req := makeAuthRequest(http.MethodOptions, "/graphql", nil)
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

			var eventData map[string]any
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

func TestExecuteFunction_SaveResponse(t *testing.T) {
	t.Run("saves response when enabled", func(t *testing.T) {
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
      ["Content-Type"] = "application/json"
    },
    body = '{"message": "success"}'
  }
end
`, nil)
		if err != nil {
			t.Fatalf("Failed to create version: %v", err)
		}

		// Enable save_response
		saveResponse := true
		if err := database.UpdateFunction(context.Background(), fn.ID, store.UpdateFunctionRequest{
			SaveResponse: &saveResponse,
		}); err != nil {
			t.Fatalf("Failed to enable save_response: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID, nil)
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

		// Verify response JSON was stored
		if execution.ResponseJSON == nil {
			t.Fatal("Expected ResponseJSON to be stored when save_response is enabled")
		}

		// Parse and verify the response JSON content
		var responseData map[string]any
		if err := json.Unmarshal([]byte(*execution.ResponseJSON), &responseData); err != nil {
			t.Fatalf("Failed to parse response JSON: %v", err)
		}

		// Verify status code
		if statusCode, ok := responseData["statusCode"].(float64); !ok || int(statusCode) != 200 {
			t.Errorf("Expected statusCode 200, got %v", responseData["statusCode"])
		}

		// Verify body
		if body, ok := responseData["body"].(string); !ok || body != `{"message": "success"}` {
			t.Errorf("Expected body to be stored, got %v", responseData["body"])
		}

		// Verify headers
		if headers, ok := responseData["headers"].(map[string]any); ok {
			if contentType, ok := headers["Content-Type"].(string); !ok || contentType != "application/json" {
				t.Errorf("Expected Content-Type header, got %v", headers["Content-Type"])
			}
		} else {
			t.Error("Expected headers to be present in response JSON")
		}
	})

	t.Run("does not save response when disabled", func(t *testing.T) {
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

		// save_response is disabled by default

		req := httptest.NewRequest(http.MethodPost, "/fn/"+fn.ID, nil)
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

		// Verify response JSON was NOT stored
		if execution.ResponseJSON != nil {
			t.Error("Expected ResponseJSON to be nil when save_response is disabled")
		}
	})
}

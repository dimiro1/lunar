package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultClient_Get(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	client := NewDefaultClient()
	resp, err := client.Get(Request{
		URL: server.URL,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if resp.Body != "Hello, World!" {
		t.Errorf("Expected body 'Hello, World!', got '%s'", resp.Body)
	}

	if !resp.IsSuccess() {
		t.Error("Expected IsSuccess() to return true")
	}
}

func TestDefaultClient_Post(t *testing.T) {
	expectedBody := `{"name":"test"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Read and verify request body
		buf := make([]byte, len(expectedBody))
		_, _ = r.Body.Read(buf)
		if string(buf) != expectedBody {
			t.Errorf("Expected body '%s', got '%s'", expectedBody, string(buf))
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("Created"))
	}))
	defer server.Close()

	client := NewDefaultClient()
	resp, err := client.Post(Request{
		URL:  server.URL,
		Body: expectedBody,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	if resp.Body != "Created" {
		t.Errorf("Expected body 'Created', got '%s'", resp.Body)
	}
}

func TestDefaultClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefaultClient()
	resp, err := client.Put(Request{
		URL:  server.URL,
		Body: "update data",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestDefaultClient_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefaultClient()
	resp, err := client.Patch(Request{
		URL:  server.URL,
		Body: "partial update",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestDefaultClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewDefaultClient()
	resp, err := client.Delete(Request{
		URL: server.URL,
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestDefaultClient_QueryParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("Expected query param 'page=1', got '%s'", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("Expected query param 'limit=10', got '%s'", r.URL.Query().Get("limit"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefaultClient()
	resp, err := client.Get(Request{
		URL: server.URL,
		Query: Query{
			"page":  "1",
			"limit": "10",
		},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestDefaultClient_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Authorization") != "Bearer token123" {
			t.Errorf("Expected Authorization header 'Bearer token123', got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		// Send response headers
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewDefaultClient()
	resp, err := client.Get(Request{
		URL: server.URL,
		Headers: Headers{
			"Authorization": "Bearer token123",
			"Content-Type":  "application/json",
		},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify response headers
	if resp.Headers["X-Custom-Header"] != "test-value" {
		t.Errorf("Expected X-Custom-Header 'test-value', got '%s'", resp.Headers["X-Custom-Header"])
	}
}

func TestDefaultClient_ErrorStatusCodes(t *testing.T) {
	testCases := []struct {
		name            string
		statusCode      int
		expectedError   bool
		expectedSuccess bool
	}{
		{"Success 200", http.StatusOK, false, true},
		{"Success 201", http.StatusCreated, false, true},
		{"Client Error 400", http.StatusBadRequest, true, false},
		{"Client Error 404", http.StatusNotFound, true, false},
		{"Server Error 500", http.StatusInternalServerError, true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			client := NewDefaultClient()
			resp, err := client.Get(Request{
				URL: server.URL,
			})
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if resp.StatusCode != tc.statusCode {
				t.Errorf("Expected status code %d, got %d", tc.statusCode, resp.StatusCode)
			}

			if resp.IsError() != tc.expectedError {
				t.Errorf("Expected IsError() to return %v, got %v", tc.expectedError, resp.IsError())
			}

			if resp.IsSuccess() != tc.expectedSuccess {
				t.Errorf("Expected IsSuccess() to return %v, got %v", tc.expectedSuccess, resp.IsSuccess())
			}
		})
	}
}

func TestDefaultClient_InvalidURL(t *testing.T) {
	client := NewDefaultClient()
	_, err := client.Get(Request{
		URL: "://invalid-url",
	})

	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestDefaultClient_NetworkError(t *testing.T) {
	client := NewDefaultClient()
	// Use a URL that will fail (non-existent server)
	_, err := client.Get(Request{
		URL: "http://localhost:99999",
	})

	if err == nil {
		t.Error("Expected network error, got nil")
	}
}

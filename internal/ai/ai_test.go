package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimiro1/faas-go/internal/env"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
)

func TestNewDefaultClient(t *testing.T) {
	httpClient := internalhttp.NewDefaultClient()
	envStore := env.NewMemoryStore()

	client := NewDefaultClient(httpClient, envStore)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}
	if client.envStore == nil {
		t.Error("expected non-nil envStore")
	}
}

func TestChat_UnsupportedProvider(t *testing.T) {
	envStore := env.NewMemoryStore()
	client := NewDefaultClient(internalhttp.NewDefaultClient(), envStore)

	req := ChatRequest{
		Provider: "unsupported",
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "Hello"}},
	}

	_, err := client.Chat("func-1", req)
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	if err.Error() != "unsupported provider: unsupported (use openai or anthropic)" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChat_MissingOpenAIAPIKey(t *testing.T) {
	envStore := env.NewMemoryStore()
	client := NewDefaultClient(internalhttp.NewDefaultClient(), envStore)

	req := ChatRequest{
		Provider: "openai",
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hello"}},
	}

	_, err := client.Chat("func-1", req)
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if err.Error() != "OPENAI_API_KEY not set in function environment" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChat_MissingAnthropicAPIKey(t *testing.T) {
	envStore := env.NewMemoryStore()
	client := NewDefaultClient(internalhttp.NewDefaultClient(), envStore)

	req := ChatRequest{
		Provider: "anthropic",
		Model:    "claude-3-haiku",
		Messages: []Message{{Role: "user", Content: "Hello"}},
	}

	_, err := client.Chat("func-1", req)
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if err.Error() != "ANTHROPIC_API_KEY not set in function environment" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChat_OpenAI_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Authorization header")
		}

		resp := map[string]any{
			"model": "gpt-4",
			"choices": []map[string]any{
				{"message": map[string]any{"content": "Hello there!"}},
			},
			"usage": map[string]any{
				"prompt_tokens":     10,
				"completion_tokens": 5,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("func-1", "OPENAI_API_KEY", "test-api-key")

	client := NewDefaultClient(internalhttp.NewDefaultClient(), envStore)

	req := ChatRequest{
		Provider:  "openai",
		Model:     "gpt-4",
		Messages:  []Message{{Role: "user", Content: "Hello"}},
		MaxTokens: 100,
		Endpoint:  server.URL,
	}

	resp, err := client.Chat("func-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "Hello there!" {
		t.Errorf("expected content 'Hello there!', got '%s'", resp.Content)
	}
	if resp.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", resp.Model)
	}
	if resp.Usage.InputTokens != 10 {
		t.Errorf("expected input tokens 10, got %d", resp.Usage.InputTokens)
	}
	if resp.Usage.OutputTokens != 5 {
		t.Errorf("expected output tokens 5, got %d", resp.Usage.OutputTokens)
	}
	if resp.Endpoint == "" {
		t.Error("expected endpoint to be set")
	}
	if resp.RequestJSON == "" {
		t.Error("expected request JSON to be set")
	}
	if resp.ResponseJSON == "" {
		t.Error("expected response JSON to be set")
	}
}

func TestChat_Anthropic_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("expected x-api-key header")
		}
		if r.Header.Get("anthropic-version") != anthropicVersion {
			t.Errorf("expected anthropic-version header")
		}

		resp := map[string]any{
			"model": "claude-3-haiku",
			"content": []map[string]any{
				{"type": "text", "text": "Hello from Claude!"},
			},
			"usage": map[string]any{
				"input_tokens":  15,
				"output_tokens": 8,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("func-1", "ANTHROPIC_API_KEY", "test-api-key")

	client := NewDefaultClient(internalhttp.NewDefaultClient(), envStore)

	req := ChatRequest{
		Provider:  "anthropic",
		Model:     "claude-3-haiku",
		Messages:  []Message{{Role: "user", Content: "Hello"}},
		MaxTokens: 100,
		Endpoint:  server.URL,
	}

	resp, err := client.Chat("func-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "Hello from Claude!" {
		t.Errorf("expected content 'Hello from Claude!', got '%s'", resp.Content)
	}
	if resp.Usage.InputTokens != 15 {
		t.Errorf("expected input tokens 15, got %d", resp.Usage.InputTokens)
	}
}

func TestChat_OpenAI_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"error": map[string]any{
				"message": "Rate limit exceeded",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("func-1", "OPENAI_API_KEY", "test-api-key")

	client := NewDefaultClient(internalhttp.NewDefaultClient(), envStore)

	req := ChatRequest{
		Provider: "openai",
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hello"}},
		Endpoint: server.URL,
	}

	_, err := client.Chat("func-1", req)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "OpenAI API error: Rate limit exceeded" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChat_Anthropic_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"error": map[string]any{
				"message": "Invalid API key",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("func-1", "ANTHROPIC_API_KEY", "invalid-key")

	client := NewDefaultClient(internalhttp.NewDefaultClient(), envStore)

	req := ChatRequest{
		Provider: "anthropic",
		Model:    "claude-3-haiku",
		Messages: []Message{{Role: "user", Content: "Hello"}},
		Endpoint: server.URL,
	}

	_, err := client.Chat("func-1", req)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "anthropic API error: Invalid API key" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChat_UsesCustomEndpoint(t *testing.T) {
	var receivedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.Path
		resp := map[string]any{
			"model":   "gpt-4",
			"choices": []map[string]any{{"message": map[string]any{"content": "Hi"}}},
			"usage":   map[string]any{"prompt_tokens": 1, "completion_tokens": 1},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("func-1", "OPENAI_API_KEY", "test-key")
	_ = envStore.Set("func-1", "OPENAI_ENDPOINT", "http://should-be-overridden")

	client := NewDefaultClient(internalhttp.NewDefaultClient(), envStore)

	req := ChatRequest{
		Provider: "openai",
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hello"}},
		Endpoint: server.URL, // This should override the env value
	}

	_, err := client.Chat("func-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedURL != "/chat/completions" {
		t.Errorf("expected /chat/completions, got %s", receivedURL)
	}
}

func TestChat_UsesEnvEndpoint(t *testing.T) {
	var receivedHost string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHost = r.Host
		resp := map[string]any{
			"model":   "gpt-4",
			"choices": []map[string]any{{"message": map[string]any{"content": "Hi"}}},
			"usage":   map[string]any{"prompt_tokens": 1, "completion_tokens": 1},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("func-1", "OPENAI_API_KEY", "test-key")
	_ = envStore.Set("func-1", "OPENAI_ENDPOINT", server.URL)

	client := NewDefaultClient(internalhttp.NewDefaultClient(), envStore)

	req := ChatRequest{
		Provider: "openai",
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hello"}},
		// No Endpoint override - should use env
	}

	_, err := client.Chat("func-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedHost == "" {
		t.Error("expected request to be received")
	}
}

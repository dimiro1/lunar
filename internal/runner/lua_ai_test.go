package runner

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dimiro1/faas-go/internal/ai"
	"github.com/dimiro1/faas-go/internal/env"
	"github.com/dimiro1/faas-go/internal/events"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
)

func TestRun_AI_OpenAI_Success(t *testing.T) {
	// Create mock OpenAI server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Errorf("expected path ending with /chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("expected Authorization header 'Bearer test-api-key', got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]any
		_ = json.Unmarshal(body, &reqBody)

		if reqBody["model"] != "gpt-4o-mini" {
			t.Errorf("expected model 'gpt-4o-mini', got %v", reqBody["model"])
		}

		// Return mock response
		resp := map[string]any{
			"id":    "chatcmpl-123",
			"model": "gpt-4o-mini",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": "Hello! How can I help you today?",
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     10,
				"completion_tokens": 8,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "OPENAI_API_KEY", "test-api-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "openai",
		model = "gpt-4o-mini",
		messages = {
			{role = "user", content = "Hello!"}
		},
		endpoint = "` + server.URL + `"
	})

	if err then
		return {
			statusCode = 500,
			body = "Error: " .. err
		}
	end

	return {
		statusCode = 200,
		headers = { ["Content-Type"] = "application/json" },
		body = json.encode({
			content = response.content,
			model = response.model,
			input_tokens = response.usage.input_tokens,
			output_tokens = response.usage.output_tokens
		})
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 200 {
		t.Errorf("expected status 200, got %d: %s", resp.HTTP.StatusCode, resp.HTTP.Body)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(resp.HTTP.Body), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["content"] != "Hello! How can I help you today?" {
		t.Errorf("expected content 'Hello! How can I help you today?', got %v", result["content"])
	}
	if result["model"] != "gpt-4o-mini" {
		t.Errorf("expected model 'gpt-4o-mini', got %v", result["model"])
	}
	if result["input_tokens"].(float64) != 10 {
		t.Errorf("expected input_tokens 10, got %v", result["input_tokens"])
	}
	if result["output_tokens"].(float64) != 8 {
		t.Errorf("expected output_tokens 8, got %v", result["output_tokens"])
	}
}

func TestRun_AI_Anthropic_Success(t *testing.T) {
	// Create mock Anthropic server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/v1/messages") {
			t.Errorf("expected path ending with /v1/messages, got %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-anthropic-key" {
			t.Errorf("expected x-api-key header 'test-anthropic-key', got %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected anthropic-version header '2023-06-01', got %s", r.Header.Get("anthropic-version"))
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]any
		_ = json.Unmarshal(body, &reqBody)

		if reqBody["model"] != "claude-3-haiku-20240307" {
			t.Errorf("expected model 'claude-3-haiku-20240307', got %v", reqBody["model"])
		}

		// Verify system prompt was extracted
		if reqBody["system"] != "You are helpful" {
			t.Errorf("expected system 'You are helpful', got %v", reqBody["system"])
		}

		// Return mock response
		resp := map[string]any{
			"id":    "msg_123",
			"model": "claude-3-haiku-20240307",
			"content": []map[string]any{
				{
					"type": "text",
					"text": "Hello! I'm Claude, happy to help!",
				},
			},
			"usage": map[string]any{
				"input_tokens":  15,
				"output_tokens": 12,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "ANTHROPIC_API_KEY", "test-anthropic-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "anthropic",
		model = "claude-3-haiku-20240307",
		messages = {
			{role = "system", content = "You are helpful"},
			{role = "user", content = "Hello!"}
		},
		endpoint = "` + server.URL + `"
	})

	if err then
		return {
			statusCode = 500,
			body = "Error: " .. err
		}
	end

	return {
		statusCode = 200,
		headers = { ["Content-Type"] = "application/json" },
		body = json.encode({
			content = response.content,
			model = response.model,
			input_tokens = response.usage.input_tokens,
			output_tokens = response.usage.output_tokens
		})
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 200 {
		t.Errorf("expected status 200, got %d: %s", resp.HTTP.StatusCode, resp.HTTP.Body)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(resp.HTTP.Body), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["content"] != "Hello! I'm Claude, happy to help!" {
		t.Errorf("expected content 'Hello! I'm Claude, happy to help!', got %v", result["content"])
	}
	if result["model"] != "claude-3-haiku-20240307" {
		t.Errorf("expected model 'claude-3-haiku-20240307', got %v", result["model"])
	}
}

func TestRun_AI_MissingProvider(t *testing.T) {
	envStore := env.NewMemoryStore()
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		model = "gpt-4o-mini",
		messages = {
			{role = "user", content = "Hello!"}
		}
	})

	if err then
		return {
			statusCode = 400,
			body = err
		}
	end

	return { statusCode = 200 }
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.HTTP.StatusCode)
	}

	if !strings.Contains(resp.HTTP.Body, "provider is required") {
		t.Errorf("expected error about missing provider, got: %s", resp.HTTP.Body)
	}
}

func TestRun_AI_MissingModel(t *testing.T) {
	envStore := env.NewMemoryStore()
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "openai",
		messages = {
			{role = "user", content = "Hello!"}
		}
	})

	if err then
		return {
			statusCode = 400,
			body = err
		}
	end

	return { statusCode = 200 }
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.HTTP.StatusCode)
	}

	if !strings.Contains(resp.HTTP.Body, "model is required") {
		t.Errorf("expected error about missing model, got: %s", resp.HTTP.Body)
	}
}

func TestRun_AI_MissingMessages(t *testing.T) {
	envStore := env.NewMemoryStore()
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "openai",
		model = "gpt-4o-mini"
	})

	if err then
		return {
			statusCode = 400,
			body = err
		}
	end

	return { statusCode = 200 }
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.HTTP.StatusCode)
	}

	if !strings.Contains(resp.HTTP.Body, "messages is required") {
		t.Errorf("expected error about missing messages, got: %s", resp.HTTP.Body)
	}
}

func TestRun_AI_MissingAPIKey(t *testing.T) {
	// Create mock server that should not be called
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called when API key is missing")
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	// Not setting OPENAI_API_KEY

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "openai",
		model = "gpt-4o-mini",
		messages = {
			{role = "user", content = "Hello!"}
		},
		endpoint = "` + server.URL + `"
	})

	if err then
		return {
			statusCode = 400,
			body = err
		}
	end

	return { statusCode = 200 }
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.HTTP.StatusCode)
	}

	if !strings.Contains(resp.HTTP.Body, "OPENAI_API_KEY not set") {
		t.Errorf("expected error about missing API key, got: %s", resp.HTTP.Body)
	}
}

func TestRun_AI_UnsupportedProvider(t *testing.T) {
	envStore := env.NewMemoryStore()
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "unsupported",
		model = "some-model",
		messages = {
			{role = "user", content = "Hello!"}
		}
	})

	if err then
		return {
			statusCode = 400,
			body = err
		}
	end

	return { statusCode = 200 }
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.HTTP.StatusCode)
	}

	if !strings.Contains(resp.HTTP.Body, "unsupported provider") {
		t.Errorf("expected error about unsupported provider, got: %s", resp.HTTP.Body)
	}
}

func TestRun_AI_OpenAI_APIError(t *testing.T) {
	// Create mock OpenAI server that returns an error
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
	_ = envStore.Set("test-function", "OPENAI_API_KEY", "test-api-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "openai",
		model = "gpt-4o-mini",
		messages = {
			{role = "user", content = "Hello!"}
		},
		endpoint = "` + server.URL + `"
	})

	if err then
		return {
			statusCode = 500,
			body = err
		}
	end

	return { statusCode = 200 }
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", resp.HTTP.StatusCode)
	}

	if !strings.Contains(resp.HTTP.Body, "Rate limit exceeded") {
		t.Errorf("expected error about rate limit, got: %s", resp.HTTP.Body)
	}
}

func TestRun_AI_Anthropic_APIError(t *testing.T) {
	// Create mock Anthropic server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"error": map[string]any{
				"type":    "invalid_request_error",
				"message": "Invalid API key",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "ANTHROPIC_API_KEY", "invalid-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "anthropic",
		model = "claude-3-haiku-20240307",
		messages = {
			{role = "user", content = "Hello!"}
		},
		endpoint = "` + server.URL + `"
	})

	if err then
		return {
			statusCode = 500,
			body = err
		}
	end

	return { statusCode = 200 }
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", resp.HTTP.StatusCode)
	}

	if !strings.Contains(resp.HTTP.Body, "Invalid API key") {
		t.Errorf("expected error about invalid API key, got: %s", resp.HTTP.Body)
	}
}

func TestRun_AI_OpenAI_WithTemperature(t *testing.T) {
	var receivedTemp float64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]any
		_ = json.Unmarshal(body, &reqBody)

		if temp, ok := reqBody["temperature"].(float64); ok {
			receivedTemp = temp
		}

		resp := map[string]any{
			"id":    "chatcmpl-123",
			"model": "gpt-4o-mini",
			"choices": []map[string]any{
				{"message": map[string]any{"content": "Response"}},
			},
			"usage": map[string]any{
				"prompt_tokens":     5,
				"completion_tokens": 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "OPENAI_API_KEY", "test-api-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "openai",
		model = "gpt-4o-mini",
		messages = {
			{role = "user", content = "Hello!"}
		},
		temperature = 0.8,
		endpoint = "` + server.URL + `"
	})

	if err then
		return { statusCode = 500, body = err }
	end

	return { statusCode = 200 }
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 200 {
		t.Errorf("expected status 200, got %d: %s", resp.HTTP.StatusCode, resp.HTTP.Body)
	}

	if receivedTemp != 0.8 {
		t.Errorf("expected temperature 0.8, got %f", receivedTemp)
	}
}

func TestRun_AI_OpenAI_WithMaxTokens(t *testing.T) {
	var receivedMaxTokens float64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]any
		_ = json.Unmarshal(body, &reqBody)

		if maxTokens, ok := reqBody["max_tokens"].(float64); ok {
			receivedMaxTokens = maxTokens
		}

		resp := map[string]any{
			"id":    "chatcmpl-123",
			"model": "gpt-4o-mini",
			"choices": []map[string]any{
				{"message": map[string]any{"content": "Response"}},
			},
			"usage": map[string]any{
				"prompt_tokens":     5,
				"completion_tokens": 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "OPENAI_API_KEY", "test-api-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "openai",
		model = "gpt-4o-mini",
		messages = {
			{role = "user", content = "Hello!"}
		},
		max_tokens = 2000,
		endpoint = "` + server.URL + `"
	})

	if err then
		return { statusCode = 500, body = err }
	end

	return { statusCode = 200 }
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 200 {
		t.Errorf("expected status 200, got %d: %s", resp.HTTP.StatusCode, resp.HTTP.Body)
	}

	if receivedMaxTokens != 2000 {
		t.Errorf("expected max_tokens 2000, got %f", receivedMaxTokens)
	}
}

func TestRun_AI_Anthropic_MultipleContentBlocks(t *testing.T) {
	// Create mock Anthropic server that returns multiple content blocks
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":    "msg_123",
			"model": "claude-3-haiku-20240307",
			"content": []map[string]any{
				{
					"type": "text",
					"text": "First part. ",
				},
				{
					"type": "text",
					"text": "Second part.",
				},
			},
			"usage": map[string]any{
				"input_tokens":  10,
				"output_tokens": 5,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "ANTHROPIC_API_KEY", "test-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "anthropic",
		model = "claude-3-haiku-20240307",
		messages = {
			{role = "user", content = "Hello!"}
		},
		endpoint = "` + server.URL + `"
	})

	if err then
		return { statusCode = 500, body = err }
	end

	return {
		statusCode = 200,
		body = response.content
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 200 {
		t.Errorf("expected status 200, got %d: %s", resp.HTTP.StatusCode, resp.HTTP.Body)
	}

	expectedContent := "First part. Second part."
	if resp.HTTP.Body != expectedContent {
		t.Errorf("expected content '%s', got '%s'", expectedContent, resp.HTTP.Body)
	}
}

func TestRun_AI_EmptyMessages(t *testing.T) {
	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "OPENAI_API_KEY", "test-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		AI:     ai.NewDefaultClient(internalhttp.NewDefaultClient(), envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/chat",
	}

	luaCode := `
function handler(ctx, event)
	local response, err = ai.chat({
		provider = "openai",
		model = "gpt-4o-mini",
		messages = {}
	})

	if err then
		return {
			statusCode = 400,
			body = err
		}
	end

	return { statusCode = 200 }
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.HTTP.StatusCode)
	}

	if !strings.Contains(resp.HTTP.Body, "messages cannot be empty") {
		t.Errorf("expected error about empty messages, got: %s", resp.HTTP.Body)
	}
}

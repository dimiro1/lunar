package runner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dimiro1/faas-go/internal/email"
	"github.com/dimiro1/faas-go/internal/env"
	"github.com/dimiro1/faas-go/internal/events"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
)

func TestRun_Email_MissingFrom(t *testing.T) {
	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "RESEND_API_KEY", "test-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/send",
	}

	luaCode := `
function handler(ctx, event)
	local result, err = email.send({
		to = "recipient@example.com",
		subject = "Test",
		text = "Hello"
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

	if !strings.Contains(resp.HTTP.Body, "from is required") {
		t.Errorf("expected error about missing from, got: %s", resp.HTTP.Body)
	}
}

func TestRun_Email_MissingTo(t *testing.T) {
	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "RESEND_API_KEY", "test-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/send",
	}

	luaCode := `
function handler(ctx, event)
	local result, err = email.send({
		from = "sender@example.com",
		subject = "Test",
		text = "Hello"
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

	if !strings.Contains(resp.HTTP.Body, "to is required") {
		t.Errorf("expected error about missing to, got: %s", resp.HTTP.Body)
	}
}

func TestRun_Email_MissingSubject(t *testing.T) {
	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "RESEND_API_KEY", "test-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/send",
	}

	luaCode := `
function handler(ctx, event)
	local result, err = email.send({
		from = "sender@example.com",
		to = "recipient@example.com",
		text = "Hello"
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

	if !strings.Contains(resp.HTTP.Body, "subject is required") {
		t.Errorf("expected error about missing subject, got: %s", resp.HTTP.Body)
	}
}

func TestRun_Email_MissingContent(t *testing.T) {
	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "RESEND_API_KEY", "test-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/send",
	}

	luaCode := `
function handler(ctx, event)
	local result, err = email.send({
		from = "sender@example.com",
		to = "recipient@example.com",
		subject = "Test"
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

	if !strings.Contains(resp.HTTP.Body, "either text or html content is required") {
		t.Errorf("expected error about missing content, got: %s", resp.HTTP.Body)
	}
}

func TestRun_Email_MissingAPIKey(t *testing.T) {
	envStore := env.NewMemoryStore()
	// Not setting RESEND_API_KEY

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		Email:  email.NewDefaultClient(envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/send",
	}

	luaCode := `
function handler(ctx, event)
	local result, err = email.send({
		from = "sender@example.com",
		to = "recipient@example.com",
		subject = "Test",
		text = "Hello"
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

	if !strings.Contains(resp.HTTP.Body, "RESEND_API_KEY not set") {
		t.Errorf("expected error about missing API key, got: %s", resp.HTTP.Body)
	}
}

func TestRun_Email_EmptyToArray(t *testing.T) {
	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "RESEND_API_KEY", "test-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/send",
	}

	luaCode := `
function handler(ctx, event)
	local result, err = email.send({
		from = "sender@example.com",
		to = {},
		subject = "Test",
		text = "Hello"
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

	if !strings.Contains(resp.HTTP.Body, "to cannot be empty") {
		t.Errorf("expected error about empty to, got: %s", resp.HTTP.Body)
	}
}

func TestRun_Email_InvalidToType(t *testing.T) {
	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "RESEND_API_KEY", "test-key")

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/send",
	}

	luaCode := `
function handler(ctx, event)
	local result, err = email.send({
		from = "sender@example.com",
		to = 12345,
		subject = "Test",
		text = "Hello"
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

	if !strings.Contains(resp.HTTP.Body, "to must be a string or table") {
		t.Errorf("expected error about invalid to type, got: %s", resp.HTTP.Body)
	}
}

// TestRun_Email_Success tests a successful email send with a mock server
func TestRun_Email_Success_MockServer(t *testing.T) {
	// Create mock Resend server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/emails" {
			t.Errorf("expected path /emails, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-resend-key" {
			t.Errorf("expected Authorization header 'Bearer test-resend-key', got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// Verify request body
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if reqBody["from"] != "sender@example.com" {
			t.Errorf("expected from 'sender@example.com', got %v", reqBody["from"])
		}
		if reqBody["subject"] != "Test Subject" {
			t.Errorf("expected subject 'Test Subject', got %v", reqBody["subject"])
		}

		// Return mock response
		resp := map[string]string{
			"id": "email_123456",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "RESEND_API_KEY", "test-resend-key")
	_ = envStore.Set("test-function", "RESEND_BASE_URL", server.URL)

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		Email:  email.NewDefaultClient(envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/send",
	}

	luaCode := `
function handler(ctx, event)
	local result, err = email.send({
		from = "sender@example.com",
		to = {"recipient@example.com"},
		subject = "Test Subject",
		text = "Hello, World!"
	})

	if err then
		return {
			statusCode = 500,
			body = err
		}
	end

	return {
		statusCode = 200,
		headers = { ["Content-Type"] = "application/json" },
		body = json.encode({ id = result.id })
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

	if result["id"] != "email_123456" {
		t.Errorf("expected id 'email_123456', got %v", result["id"])
	}
}

// TestRun_Email_Success_WithScheduledAt tests email scheduling with Unix timestamp
func TestRun_Email_Success_WithScheduledAt(t *testing.T) {
	var receivedScheduledAt string

	// Create mock Resend server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		// Capture the scheduled_at value
		if sa, ok := reqBody["scheduled_at"].(string); ok {
			receivedScheduledAt = sa
		}

		resp := map[string]string{"id": "email_scheduled"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	envStore := env.NewMemoryStore()
	_ = envStore.Set("test-function", "RESEND_API_KEY", "test-resend-key")
	_ = envStore.Set("test-function", "RESEND_BASE_URL", server.URL)

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   internalhttp.NewDefaultClient(),
		Email:  email.NewDefaultClient(envStore),
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/send",
	}

	// Use a fixed timestamp for testing: 2025-01-01 12:00:00 UTC
	luaCode := `
function handler(ctx, event)
	local result, err = email.send({
		from = "sender@example.com",
		to = "recipient@example.com",
		subject = "Scheduled Email",
		text = "This is scheduled",
		scheduled_at = 1735732800
	})

	if err then
		return { statusCode = 500, body = err }
	end

	return {
		statusCode = 200,
		body = result.id
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

	// Verify the scheduled_at was converted to ISO 8601 format
	expectedScheduledAt := "2025-01-01T12:00:00Z"
	if receivedScheduledAt != expectedScheduledAt {
		t.Errorf("expected scheduled_at '%s', got '%s'", expectedScheduledAt, receivedScheduledAt)
	}
}

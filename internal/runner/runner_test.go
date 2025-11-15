package runner

import (
	"context"
	"testing"
	"time"

	"github.com/dimiro1/faas-go/internal/env"
	"github.com/dimiro1/faas-go/internal/events"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
)

func TestRun_HTTPEvent_Success(t *testing.T) {
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    env.NewMemoryStore(),
		HTTP:   &internalhttp.FakeClient{},
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "GET",
		Path:   "/test",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:  `{"test": "data"}`,
		Query: map[string]string{"foo": "bar"},
	}

	luaCode := `
function handler(ctx, event)
	return {
		statusCode = 200,
		headers = {
			["Content-Type"] = "application/json"
		},
		body = "Hello, World!"
	}
end
`

	req := Request{
		Context: execCtx,
		Event:   event,
		Code:    luaCode,
	}

	resp, err := Run(context.Background(), deps, req)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.Type != events.EventTypeHTTP {
		t.Errorf("expected response type %s, got %s", events.EventTypeHTTP, resp.Type)
	}

	if resp.HTTP == nil {
		t.Fatal("expected HTTP response, got nil")
	}

	if resp.HTTP.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", resp.HTTP.StatusCode)
	}

	if resp.HTTP.Body != "Hello, World!" {
		t.Errorf("expected body 'Hello, World!', got %s", resp.HTTP.Body)
	}
}

func TestRun_HTTPEvent_AccessContext(t *testing.T) {
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    env.NewMemoryStore(),
		HTTP:   &internalhttp.FakeClient{},
	}

	execCtx := &events.ExecutionContext{
		ExecutionID:  "exec-456",
		FunctionID:   "test-function",
		FunctionName: "MyFunction",
		Version:      "v1.0.0",
		RequestID:    "req-789",
		StartedAt:    time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "GET",
		Path:   "/",
	}

	luaCode := `
function handler(ctx, event)
	return {
		statusCode = 200,
		body = "ExecutionID: " .. ctx.executionId ..
		       ", FunctionID: " .. ctx.functionId ..
		       ", FunctionName: " .. ctx.functionName ..
		       ", Version: " .. ctx.version ..
		       ", RequestID: " .. ctx.requestId
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expectedBody := "ExecutionID: exec-456, FunctionID: test-function, FunctionName: MyFunction, Version: v1.0.0, RequestID: req-789"
	if resp.HTTP.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, resp.HTTP.Body)
	}
}

func TestRun_HTTPEvent_AccessEvent(t *testing.T) {
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    env.NewMemoryStore(),
		HTTP:   &internalhttp.FakeClient{},
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "POST",
		Path:   "/api/users",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"name": "John"}`,
		Query: map[string]string{
			"action": "create",
		},
	}

	luaCode := `
function handler(ctx, event)
	return {
		statusCode = 200,
		body = "Method: " .. event.method ..
		       ", Path: " .. event.path ..
		       ", Query: " .. event.query.action ..
		       ", Body: " .. event.body
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expectedBody := "Method: POST, Path: /api/users, Query: create, Body: {\"name\": \"John\"}"
	if resp.HTTP.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, resp.HTTP.Body)
	}
}

func TestRun_Logger(t *testing.T) {
	memLogger := logger.NewMemoryLogger()
	deps := Dependencies{
		Logger: memLogger,
		KV:     kv.NewMemoryStore(),
		Env:    env.NewMemoryStore(),
		HTTP:   &internalhttp.FakeClient{},
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "GET",
		Path:   "/",
	}

	luaCode := `
function handler(ctx, event)
	log.info("Info message")
	log.debug("Debug message")
	log.warn("Warning message")
	log.error("Error message")

	return {
		statusCode = 200,
		body = "OK"
	}
end
`

	_, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	entries := memLogger.Entries("test-function")
	if len(entries) != 4 {
		t.Fatalf("expected 4 log entries, got %d", len(entries))
	}

	expectedLogs := []struct {
		level   logger.LogLevel
		message string
	}{
		{logger.Info, "Info message"},
		{logger.Debug, "Debug message"},
		{logger.Warn, "Warning message"},
		{logger.Error, "Error message"},
	}

	for i, expected := range expectedLogs {
		if entries[i].Level != expected.level {
			t.Errorf("entry %d: expected level %v, got %v", i, expected.level, entries[i].Level)
		}
		if entries[i].Message != expected.message {
			t.Errorf("entry %d: expected message %q, got %q", i, expected.message, entries[i].Message)
		}
	}
}

func TestRun_KV(t *testing.T) {
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    env.NewMemoryStore(),
		HTTP:   &internalhttp.FakeClient{},
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "GET",
		Path:   "/",
	}

	luaCode := `
function handler(ctx, event)
	kv.set("key1", "value1")
	local val = kv.get("key1")

	return {
		statusCode = 200,
		body = "Retrieved: " .. val
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.Body != "Retrieved: value1" {
		t.Errorf("expected body 'Retrieved: value1', got %q", resp.HTTP.Body)
	}

	// Verify the value is stored in KV
	val, err := deps.KV.Get("test-function", "key1")
	if err != nil {
		t.Fatalf("failed to get key from KV: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected value 'value1', got %q", val)
	}
}

func TestRun_Env(t *testing.T) {
	envStore := env.NewMemoryStore()
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    envStore,
		HTTP:   &internalhttp.FakeClient{},
	}

	// Pre-populate env variable
	_ = envStore.Set("test-function", "API_KEY", "secret-123")

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "GET",
		Path:   "/",
	}

	luaCode := `
function handler(ctx, event)
	local apiKey = env.get("API_KEY")

	return {
		statusCode = 200,
		body = "API Key: " .. apiKey
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.Body != "API Key: secret-123" {
		t.Errorf("expected body 'API Key: secret-123', got %q", resp.HTTP.Body)
	}
}

func TestRun_HTTP(t *testing.T) {
	fakeClient := internalhttp.NewFakeClient()
	fakeClient.SetResponse("GET", "https://api.example.com/status", internalhttp.Response{
		StatusCode: 200,
		Body:       `{"status": "ok"}`,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})

	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    env.NewMemoryStore(),
		HTTP:   fakeClient,
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "GET",
		Path:   "/",
	}

	luaCode := `
function handler(ctx, event)
	local resp, err = http.get("https://api.example.com/status")

	if err then
		return {
			statusCode = 500,
			body = "Error: " .. err
		}
	end

	return {
		statusCode = 200,
		body = "Remote status: " .. resp.body
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expectedBody := "Remote status: {\"status\": \"ok\"}"
	if resp.HTTP.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, resp.HTTP.Body)
	}
}

func TestRun_NoHandler(t *testing.T) {
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    env.NewMemoryStore(),
		HTTP:   &internalhttp.FakeClient{},
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "GET",
		Path:   "/",
	}

	luaCode := `
-- No handler function defined
local x = 1 + 1
`

	_, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err == nil {
		t.Fatal("expected error for missing handler, got nil")
	}

	expectedError := "handler function not found in Lua code"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

func TestRun_InvalidLuaCode(t *testing.T) {
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kv.NewMemoryStore(),
		Env:    env.NewMemoryStore(),
		HTTP:   &internalhttp.FakeClient{},
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "GET",
		Path:   "/",
	}

	luaCode := `
function handler(ctx, event)
	this is invalid lua syntax
end
`

	_, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err == nil {
		t.Fatal("expected error for invalid Lua code, got nil")
	}
}

func TestRun_Timeout(t *testing.T) {
	deps := Dependencies{
		Logger:  logger.NewMemoryLogger(),
		KV:      kv.NewMemoryStore(),
		Env:     env.NewMemoryStore(),
		HTTP:    &internalhttp.FakeClient{},
		Timeout: 100 * time.Millisecond, // Very short timeout
	}

	execCtx := &events.ExecutionContext{
		ExecutionID: "exec-123",
		FunctionID:  "test-function",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "GET",
		Path:   "/",
	}

	luaCode := `
function handler(ctx, event)
	-- Infinite loop to trigger timeout
	while true do
		local x = 1 + 1
	end

	return {
		statusCode = 200,
		body = "This should never be reached"
	}
end
`

	_, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestRun_NamespaceIsolation(t *testing.T) {
	kvStore := kv.NewMemoryStore()
	deps := Dependencies{
		Logger: logger.NewMemoryLogger(),
		KV:     kvStore,
		Env:    env.NewMemoryStore(),
		HTTP:   &internalhttp.FakeClient{},
	}

	// Function 1 sets a value
	execCtx1 := &events.ExecutionContext{
		ExecutionID: "exec-1",
		FunctionID:  "function-1",
		StartedAt:   time.Now().Unix(),
	}

	event := events.HTTPEvent{
		Method: "GET",
		Path:   "/",
	}

	luaCode1 := `
function handler(ctx, event)
	kv.set("shared-key", "value-from-function-1")
	return { statusCode = 200, body = "OK" }
end
`

	_, err := Run(context.Background(), deps, Request{Context: execCtx1, Event: event, Code: luaCode1})
	if err != nil {
		t.Fatalf("Run failed for function-1: %v", err)
	}

	// Function 2 sets the same key (different namespace)
	execCtx2 := &events.ExecutionContext{
		ExecutionID: "exec-2",
		FunctionID:  "function-2",
		StartedAt:   time.Now().Unix(),
	}

	luaCode2 := `
function handler(ctx, event)
	kv.set("shared-key", "value-from-function-2")
	return { statusCode = 200, body = "OK" }
end
`

	_, err = Run(context.Background(), deps, Request{Context: execCtx2, Event: event, Code: luaCode2})
	if err != nil {
		t.Fatalf("Run failed for function-2: %v", err)
	}

	// Verify namespace isolation
	val1, _ := kvStore.Get("function-1", "shared-key")
	val2, _ := kvStore.Get("function-2", "shared-key")

	if val1 != "value-from-function-1" {
		t.Errorf("function-1: expected 'value-from-function-1', got %q", val1)
	}

	if val2 != "value-from-function-2" {
		t.Errorf("function-2: expected 'value-from-function-2', got %q", val2)
	}
}

package runner

import (
	"context"
	"encoding/json"
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

	entries := memLogger.Entries("exec-123")
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

func TestRun_JSON_Decode(t *testing.T) {
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
		Path:   "/api/data",
		Body:   `{"name": "Alice", "age": 30, "active": true}`,
	}

	luaCode := `
function handler(ctx, event)
	local data = json.decode(event.body)

	return {
		statusCode = 200,
		body = "Name: " .. data.name .. ", Age: " .. tostring(data.age) .. ", Active: " .. tostring(data.active)
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expectedBody := "Name: Alice, Age: 30, Active: true"
	if resp.HTTP.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, resp.HTTP.Body)
	}
}

func TestRun_JSON_Encode(t *testing.T) {
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
		Path:   "/api/user",
	}

	luaCode := `
function handler(ctx, event)
	local user = {
		id = 123,
		name = "Bob",
		email = "bob@example.com",
		admin = false
	}

	local jsonStr = json.encode(user)

	return {
		statusCode = 200,
		headers = {
			["Content-Type"] = "application/json"
		},
		body = jsonStr
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", resp.HTTP.StatusCode)
	}

	// Verify the body is valid JSON
	var result map[string]any
	if err := json.Unmarshal([]byte(resp.HTTP.Body), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if result["name"] != "Bob" {
		t.Errorf("expected name 'Bob', got %v", result["name"])
	}

	if result["id"].(float64) != 123 {
		t.Errorf("expected id 123, got %v", result["id"])
	}

	if result["admin"] != false {
		t.Errorf("expected admin false, got %v", result["admin"])
	}
}

func TestRun_JSON_Array(t *testing.T) {
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
		Path:   "/api/items",
		Body:   `["apple", "banana", "cherry"]`,
	}

	luaCode := `
function handler(ctx, event)
	local items = json.decode(event.body)

	-- Create a new array with modified items
	local result = {}
	for i, item in ipairs(items) do
		table.insert(result, "fruit: " .. item)
	end

	local jsonStr = json.encode(result)

	return {
		statusCode = 200,
		headers = {
			["Content-Type"] = "application/json"
		},
		body = jsonStr
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify the body is valid JSON array
	var result []string
	if err := json.Unmarshal([]byte(resp.HTTP.Body), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	expected := []string{"fruit: apple", "fruit: banana", "fruit: cherry"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), len(result))
	}

	for i, item := range expected {
		if result[i] != item {
			t.Errorf("item %d: expected %q, got %q", i, item, result[i])
		}
	}
}

func TestRun_JSON_NestedObjects(t *testing.T) {
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
		Path:   "/api/data",
		Body:   `{"user": {"name": "Charlie", "age": 25}, "items": [1, 2, 3]}`,
	}

	luaCode := `
function handler(ctx, event)
	local data = json.decode(event.body)

	local response = {
		userName = data.user.name,
		userAge = data.user.age,
		itemCount = #data.items
	}

	return {
		statusCode = 200,
		headers = {
			["Content-Type"] = "application/json"
		},
		body = json.encode(response)
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify the response
	var result map[string]any
	if err := json.Unmarshal([]byte(resp.HTTP.Body), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if result["userName"] != "Charlie" {
		t.Errorf("expected userName 'Charlie', got %v", result["userName"])
	}

	if result["userAge"].(float64) != 25 {
		t.Errorf("expected userAge 25, got %v", result["userAge"])
	}

	if result["itemCount"].(float64) != 3 {
		t.Errorf("expected itemCount 3, got %v", result["itemCount"])
	}
}

func TestRun_JSON_DecodeError(t *testing.T) {
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
		Path:   "/api/data",
		Body:   `{"invalid json`,
	}

	luaCode := `
function handler(ctx, event)
	local data, err = json.decode(event.body)

	if err then
		return {
			statusCode = 400,
			body = "JSON Error: " .. err
		}
	end

	return {
		statusCode = 200,
		body = "OK"
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if resp.HTTP.StatusCode != 400 {
		t.Errorf("expected status code 400, got %d", resp.HTTP.StatusCode)
	}

	if !contains(resp.HTTP.Body, "JSON Error") {
		t.Errorf("expected error message in body, got %q", resp.HTTP.Body)
	}
}

// Helper function for substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRun_Base64(t *testing.T) {
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
	local original = "Hello, World!"
	local encoded = base64.encode(original)
	local decoded, err = base64.decode(encoded)

	if err then
		return {
			statusCode = 500,
			body = "Error: " .. err
		}
	end

	return {
		statusCode = 200,
		body = "Original: " .. original .. ", Encoded: " .. encoded .. ", Decoded: " .. decoded
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expectedBody := "Original: Hello, World!, Encoded: SGVsbG8sIFdvcmxkIQ==, Decoded: Hello, World!"
	if resp.HTTP.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, resp.HTTP.Body)
	}
}

func TestRun_Crypto_SHA256(t *testing.T) {
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
	local hash = crypto.sha256("password")

	return {
		statusCode = 200,
		body = hash
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// SHA256 of "password"
	expectedHash := "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"
	if resp.HTTP.Body != expectedHash {
		t.Errorf("expected hash %q, got %q", expectedHash, resp.HTTP.Body)
	}
}

func TestRun_Crypto_HMAC(t *testing.T) {
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
		Path:   "/webhook",
		Body:   "test message",
	}

	luaCode := `
function handler(ctx, event)
	local signature = crypto.hmac_sha256(event.body, "secret-key")

	return {
		statusCode = 200,
		body = signature
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify it's a valid hex string (64 chars for SHA256)
	if len(resp.HTTP.Body) != 64 {
		t.Errorf("expected HMAC length 64, got %d", len(resp.HTTP.Body))
	}
}

func TestRun_Crypto_UUID(t *testing.T) {
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
	local id = crypto.uuid()

	return {
		statusCode = 200,
		body = id
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Check UUID format (36 chars with dashes)
	if len(resp.HTTP.Body) != 36 {
		t.Errorf("expected UUID length 36, got %d: %s", len(resp.HTTP.Body), resp.HTTP.Body)
	}
}

func TestRun_Time(t *testing.T) {
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
	local now = time.now()
	local formatted = time.format(now, "2006-01-02")

	return {
		statusCode = 200,
		body = "Now: " .. tostring(now) .. ", Formatted: " .. formatted
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Just verify it contains expected parts
	if !contains(resp.HTTP.Body, "Now:") || !contains(resp.HTTP.Body, "Formatted:") {
		t.Errorf("unexpected body: %q", resp.HTTP.Body)
	}
}

func TestRun_Time_Parse(t *testing.T) {
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
	local timestamp, err = time.parse("2024-01-15", "2006-01-02")

	if err then
		return {
			statusCode = 500,
			body = "Error: " .. err
		}
	end

	return {
		statusCode = 200,
		body = tostring(timestamp)
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify we got a timestamp
	if resp.HTTP.Body == "" {
		t.Errorf("expected timestamp, got empty string")
	}
}

func TestRun_URL_Parse(t *testing.T) {
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
	local parsed, err = url.parse("https://example.com:8080/path?foo=bar&baz=qux#section")

	if err then
		return {
			statusCode = 500,
			body = "Error: " .. err
		}
	end

	return {
		statusCode = 200,
		body = "Scheme: " .. parsed.scheme .. ", Host: " .. parsed.host .. ", Path: " .. parsed.path .. ", Query.foo: " .. parsed.query.foo
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expectedBody := "Scheme: https, Host: example.com:8080, Path: /path, Query.foo: bar"
	if resp.HTTP.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, resp.HTTP.Body)
	}
}

func TestRun_URL_Encode_Decode(t *testing.T) {
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
	local original = "hello world & test"
	local encoded = url.encode(original)
	local decoded, err = url.decode(encoded)

	if err then
		return {
			statusCode = 500,
			body = "Error: " .. err
		}
	end

	return {
		statusCode = 200,
		body = "Original: " .. original .. ", Encoded: " .. encoded .. ", Decoded: " .. decoded
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expectedBody := "Original: hello world & test, Encoded: hello+world+%26+test, Decoded: hello world & test"
	if resp.HTTP.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, resp.HTTP.Body)
	}
}

func TestRun_Strings(t *testing.T) {
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
	local trimmed = strings.trim("  hello  ")
	local upper = strings.toUpper(trimmed)
	local lower = strings.toLower(upper)
	local parts = strings.split("a,b,c", ",")
	local joined = strings.join(parts, "-")
	local hasPrefix = strings.hasPrefix("hello", "hel")

	return {
		statusCode = 200,
		body = "Trimmed: " .. trimmed .. ", Upper: " .. upper .. ", Joined: " .. joined .. ", HasPrefix: " .. tostring(hasPrefix)
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	expectedBody := "Trimmed: hello, Upper: HELLO, Joined: a-b-c, HasPrefix: true"
	if resp.HTTP.Body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, resp.HTTP.Body)
	}
}

func TestRun_Random(t *testing.T) {
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
	local num = random.int(1, 100)
	local float = random.float()
	local str = random.string(16)
	local hex = random.hex(8)

	return {
		statusCode = 200,
		body = "Int: " .. tostring(num) .. ", Float: " .. tostring(float) .. ", StrLen: " .. #str .. ", HexLen: " .. #hex
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Just verify the expected structure
	if !contains(resp.HTTP.Body, "Int:") || !contains(resp.HTTP.Body, "StrLen: 16") || !contains(resp.HTTP.Body, "HexLen: 16") {
		t.Errorf("unexpected body: %q", resp.HTTP.Body)
	}
}

func TestRun_Random_ID(t *testing.T) {
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
	local id1 = random.id()
	local id2 = random.id()

	return {
		statusCode = 200,
		body = "ID1: " .. id1 .. ", ID2: " .. id2 .. ", Same: " .. tostring(id1 == id2)
	}
end
`

	resp, err := Run(context.Background(), deps, Request{Context: execCtx, Event: event, Code: luaCode})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify IDs are different
	if !contains(resp.HTTP.Body, "Same: false") {
		t.Errorf("expected different IDs, got: %q", resp.HTTP.Body)
	}

	// Verify the format (xid is 20 characters)
	if !contains(resp.HTTP.Body, "ID1:") || !contains(resp.HTTP.Body, "ID2:") {
		t.Errorf("unexpected body: %q", resp.HTTP.Body)
	}
}

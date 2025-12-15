package ai

import (
	"testing"

	"github.com/dimiro1/lunar/internal/store"
)

func TestNewMemoryTracker(t *testing.T) {
	tracker := NewMemoryTracker()

	if tracker == nil {
		t.Fatal("expected non-nil tracker")
	}
	if tracker.requests == nil {
		t.Error("expected non-nil requests slice")
	}
}

func TestMemoryTracker_Track(t *testing.T) {
	tracker := NewMemoryTracker()

	inputTokens := 10
	outputTokens := 5
	responseJSON := `{"content":"Hello"}`

	req := TrackRequest{
		Provider:     "openai",
		Model:        "gpt-4",
		Endpoint:     "https://api.openai.com/v1/chat/completions",
		RequestJSON:  `{"model":"gpt-4"}`,
		ResponseJSON: &responseJSON,
		Status:       store.AIRequestStatusSuccess,
		InputTokens:  &inputTokens,
		OutputTokens: &outputTokens,
		DurationMs:   1500,
	}

	tracker.Track("exec-1", req)

	requests := tracker.Requests("exec-1")
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	tracked := requests[0]
	if tracked.Provider != "openai" {
		t.Errorf("expected provider 'openai', got '%s'", tracked.Provider)
	}
	if tracked.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", tracked.Model)
	}
	if tracked.ExecutionID != "exec-1" {
		t.Errorf("expected executionID 'exec-1', got '%s'", tracked.ExecutionID)
	}
	if tracked.Status != store.AIRequestStatusSuccess {
		t.Errorf("expected status 'success', got '%s'", tracked.Status)
	}
	if tracked.DurationMs != 1500 {
		t.Errorf("expected duration 1500, got %d", tracked.DurationMs)
	}
	if tracked.ID == "" {
		t.Error("expected non-empty ID")
	}
	if tracked.CreatedAt == 0 {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestMemoryTracker_Track_ErrorRequest(t *testing.T) {
	tracker := NewMemoryTracker()

	errMsg := "API rate limit exceeded"
	req := TrackRequest{
		Provider:     "openai",
		Model:        "gpt-4",
		Endpoint:     "https://api.openai.com/v1/chat/completions",
		RequestJSON:  `{"model":"gpt-4"}`,
		Status:       store.AIRequestStatusError,
		ErrorMessage: &errMsg,
		DurationMs:   500,
	}

	tracker.Track("exec-1", req)

	requests := tracker.Requests("exec-1")
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	tracked := requests[0]
	if tracked.Status != store.AIRequestStatusError {
		t.Errorf("expected status 'error', got '%s'", tracked.Status)
	}
	if tracked.ErrorMessage == nil || *tracked.ErrorMessage != "API rate limit exceeded" {
		t.Errorf("expected error message 'API rate limit exceeded'")
	}
	if tracked.InputTokens != nil {
		t.Error("expected nil input tokens for error request")
	}
	if tracked.OutputTokens != nil {
		t.Error("expected nil output tokens for error request")
	}
}

func TestMemoryTracker_Requests_FiltersByExecutionID(t *testing.T) {
	tracker := NewMemoryTracker()

	req1 := TrackRequest{Provider: "openai", Model: "gpt-4", Status: store.AIRequestStatusSuccess}
	req2 := TrackRequest{Provider: "anthropic", Model: "claude-3", Status: store.AIRequestStatusSuccess}
	req3 := TrackRequest{Provider: "openai", Model: "gpt-3.5", Status: store.AIRequestStatusSuccess}

	tracker.Track("exec-1", req1)
	tracker.Track("exec-2", req2)
	tracker.Track("exec-1", req3)

	exec1Requests := tracker.Requests("exec-1")
	if len(exec1Requests) != 2 {
		t.Errorf("expected 2 requests for exec-1, got %d", len(exec1Requests))
	}

	exec2Requests := tracker.Requests("exec-2")
	if len(exec2Requests) != 1 {
		t.Errorf("expected 1 request for exec-2, got %d", len(exec2Requests))
	}

	exec3Requests := tracker.Requests("exec-3")
	if len(exec3Requests) != 0 {
		t.Errorf("expected 0 requests for exec-3, got %d", len(exec3Requests))
	}
}

func TestMemoryTracker_RequestsPaginated(t *testing.T) {
	tracker := NewMemoryTracker()

	// Add 5 requests
	for i := 0; i < 5; i++ {
		req := TrackRequest{Provider: "openai", Model: "gpt-4", Status: store.AIRequestStatusSuccess}
		tracker.Track("exec-1", req)
	}

	// Test first page
	requests, total := tracker.RequestsPaginated("exec-1", 2, 0)
	if len(requests) != 2 {
		t.Errorf("expected 2 requests, got %d", len(requests))
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}

	// Test second page
	requests, _ = tracker.RequestsPaginated("exec-1", 2, 2)
	if len(requests) != 2 {
		t.Errorf("expected 2 requests, got %d", len(requests))
	}

	// Test last page
	requests, _ = tracker.RequestsPaginated("exec-1", 2, 4)
	if len(requests) != 1 {
		t.Errorf("expected 1 request, got %d", len(requests))
	}

	// Test offset beyond range
	requests, total = tracker.RequestsPaginated("exec-1", 2, 10)
	if len(requests) != 0 {
		t.Errorf("expected 0 requests, got %d", len(requests))
	}
	if total != 5 {
		t.Errorf("expected total 5 even with offset beyond range, got %d", total)
	}
}

func TestMemoryTracker_Clear(t *testing.T) {
	tracker := NewMemoryTracker()

	req := TrackRequest{Provider: "openai", Model: "gpt-4", Status: store.AIRequestStatusSuccess}
	tracker.Track("exec-1", req)
	tracker.Track("exec-2", req)

	if len(tracker.requests) != 2 {
		t.Fatalf("expected 2 requests before clear, got %d", len(tracker.requests))
	}

	tracker.Clear()

	if len(tracker.requests) != 0 {
		t.Errorf("expected 0 requests after clear, got %d", len(tracker.requests))
	}
}

func TestMemoryTracker_MasksSensitiveData(t *testing.T) {
	tracker := NewMemoryTracker()

	responseJSON := `{"api_key":"secret123","content":"Hello"}`
	req := TrackRequest{
		Provider:     "openai",
		Model:        "gpt-4",
		RequestJSON:  `{"api_key":"mysecret","messages":[]}`,
		ResponseJSON: &responseJSON,
		Status:       store.AIRequestStatusSuccess,
	}

	tracker.Track("exec-1", req)

	requests := tracker.Requests("exec-1")
	tracked := requests[0]

	// The masking package should mask sensitive fields
	// Check that the original values are not present
	if tracked.RequestJSON == `{"api_key":"mysecret","messages":[]}` {
		t.Error("expected request JSON to be masked")
	}
	if tracked.ResponseJSON != nil && *tracked.ResponseJSON == `{"api_key":"secret123","content":"Hello"}` {
		t.Error("expected response JSON to be masked")
	}
}

func TestMemoryTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewMemoryTracker()

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func() {
			req := TrackRequest{Provider: "openai", Model: "gpt-4", Status: store.AIRequestStatusSuccess}
			tracker.Track("exec-1", req)
			done <- true
		}()
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = tracker.Requests("exec-1")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify all writes completed
	requests := tracker.Requests("exec-1")
	if len(requests) != 10 {
		t.Errorf("expected 10 requests, got %d", len(requests))
	}
}

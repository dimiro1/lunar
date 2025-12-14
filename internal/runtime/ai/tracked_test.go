package ai

import (
	"errors"
	"testing"

	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/store"
)

// mockClient implements ai.Client for testing
type mockClient struct {
	response *ai.ChatResponse
	err      error
}

func (m *mockClient) Chat(functionID string, req ai.ChatRequest) (*ai.ChatResponse, error) {
	return m.response, m.err
}

// mockTracker implements ai.Tracker for testing
type mockTracker struct {
	tracked []ai.TrackRequest
}

func (m *mockTracker) Track(executionID string, req ai.TrackRequest) {
	m.tracked = append(m.tracked, req)
}

func (m *mockTracker) Requests(executionID string) []store.AIRequest {
	return nil
}

func (m *mockTracker) RequestsPaginated(executionID string, limit, offset int) ([]store.AIRequest, int64) {
	return nil, 0
}

func TestNewTrackedClient(t *testing.T) {
	client := &mockClient{}
	tracker := &mockTracker{}

	tc := NewTrackedClient(client, tracker, "exec-123")

	if tc == nil {
		t.Fatal("NewTrackedClient returned nil")
	}
	if tc.client != client {
		t.Error("client not set correctly")
	}
	if tc.tracker != tracker {
		t.Error("tracker not set correctly")
	}
	if tc.executionID != "exec-123" {
		t.Errorf("executionID = %q, want %q", tc.executionID, "exec-123")
	}
}

func TestChatWithTracking_Success(t *testing.T) {
	client := &mockClient{
		response: &ai.ChatResponse{
			Content:      "Hello!",
			Model:        "gpt-4",
			Endpoint:     "https://api.openai.com/v1/chat/completions",
			RequestJSON:  `{"model":"gpt-4"}`,
			ResponseJSON: `{"content":"Hello!"}`,
			Usage: ai.Usage{
				InputTokens:  10,
				OutputTokens: 5,
			},
		},
	}
	tracker := &mockTracker{}

	tc := NewTrackedClient(client, tracker, "exec-123")
	result := tc.ChatWithTracking("func-1", ai.ChatRequest{
		Provider: "openai",
		Model:    "gpt-4",
		Messages: []ai.Message{{Role: "user", Content: "Hi"}},
	})

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Response == nil {
		t.Fatal("response is nil")
	}
	if result.Response.Content != "Hello!" {
		t.Errorf("response.Content = %q, want %q", result.Response.Content, "Hello!")
	}

	// Check tracking
	if len(tracker.tracked) != 1 {
		t.Fatalf("expected 1 tracked request, got %d", len(tracker.tracked))
	}
	tracked := tracker.tracked[0]
	if tracked.Status != store.AIRequestStatusSuccess {
		t.Errorf("tracked.Status = %q, want %q", tracked.Status, store.AIRequestStatusSuccess)
	}
	if tracked.Provider != "openai" {
		t.Errorf("tracked.Provider = %q, want %q", tracked.Provider, "openai")
	}
	if tracked.Model != "gpt-4" {
		t.Errorf("tracked.Model = %q, want %q", tracked.Model, "gpt-4")
	}
	if tracked.InputTokens == nil || *tracked.InputTokens != 10 {
		t.Error("tracked.InputTokens not set correctly")
	}
	if tracked.OutputTokens == nil || *tracked.OutputTokens != 5 {
		t.Error("tracked.OutputTokens not set correctly")
	}
	if tracked.DurationMs < 0 {
		t.Error("tracked.DurationMs should be non-negative")
	}
}

func TestChatWithTracking_Error(t *testing.T) {
	client := &mockClient{
		response: &ai.ChatResponse{
			Endpoint:    "https://api.openai.com/v1/chat/completions",
			RequestJSON: `{"model":"gpt-4"}`,
		},
		err: errors.New("API error"),
	}
	tracker := &mockTracker{}

	tc := NewTrackedClient(client, tracker, "exec-123")
	result := tc.ChatWithTracking("func-1", ai.ChatRequest{
		Provider: "openai",
		Model:    "gpt-4",
		Messages: []ai.Message{{Role: "user", Content: "Hi"}},
	})

	if result.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if result.Response != nil {
		t.Error("response should be nil on error")
	}

	// Check tracking
	if len(tracker.tracked) != 1 {
		t.Fatalf("expected 1 tracked request, got %d", len(tracker.tracked))
	}
	tracked := tracker.tracked[0]
	if tracked.Status != store.AIRequestStatusError {
		t.Errorf("tracked.Status = %q, want %q", tracked.Status, store.AIRequestStatusError)
	}
	if tracked.ErrorMessage == nil || *tracked.ErrorMessage != "API error" {
		t.Error("tracked.ErrorMessage not set correctly")
	}
}

func TestChat_Wrapper(t *testing.T) {
	client := &mockClient{
		response: &ai.ChatResponse{
			Content: "Hello!",
			Model:   "gpt-4",
		},
	}
	tracker := &mockTracker{}

	tc := NewTrackedClient(client, tracker, "exec-123")
	response, err := tc.Chat("func-1", ai.ChatRequest{
		Provider: "openai",
		Model:    "gpt-4",
		Messages: []ai.Message{{Role: "user", Content: "Hi"}},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response == nil {
		t.Fatal("response is nil")
	}
	if response.Content != "Hello!" {
		t.Errorf("response.Content = %q, want %q", response.Content, "Hello!")
	}
}

func TestChatWithTracking_NilTracker(t *testing.T) {
	client := &mockClient{
		response: &ai.ChatResponse{
			Content: "Hello!",
			Model:   "gpt-4",
		},
	}

	tc := NewTrackedClient(client, nil, "exec-123")
	result := tc.ChatWithTracking("func-1", ai.ChatRequest{
		Provider: "openai",
		Model:    "gpt-4",
		Messages: []ai.Message{{Role: "user", Content: "Hi"}},
	})

	// Should not panic with nil tracker
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Response == nil {
		t.Fatal("response is nil")
	}
}

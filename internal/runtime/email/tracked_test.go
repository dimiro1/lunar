package email

import (
	"errors"
	"testing"

	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/store"
)

var _ email.Tracker = (*mockTracker)(nil) // Compile-time check

// mockClient implements email.Client for testing
type mockClient struct {
	response *email.SendResponse
	err      error
}

func (m *mockClient) Send(functionID string, req email.SendRequest) (*email.SendResponse, error) {
	return m.response, m.err
}

// mockTracker implements email.Tracker for testing
type mockTracker struct {
	tracked []email.TrackRequest
}

func (m *mockTracker) Track(executionID string, req email.TrackRequest) {
	m.tracked = append(m.tracked, req)
}

func (m *mockTracker) Requests(executionID string) []store.EmailRequest {
	return nil
}

func (m *mockTracker) RequestsPaginated(executionID string, limit, offset int) ([]store.EmailRequest, int64) {
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

func TestSendWithTracking_Success(t *testing.T) {
	client := &mockClient{
		response: &email.SendResponse{
			ID:          "msg-123",
			RequestJSON: `{"to":["test@example.com"]}`,
		},
	}
	tracker := &mockTracker{}

	tc := NewTrackedClient(client, tracker, "exec-123")
	result := tc.SendWithTracking("func-1", email.SendRequest{
		From:    "sender@example.com",
		To:      []string{"test@example.com"},
		Subject: "Test",
		Text:    "Hello",
	})

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Response == nil {
		t.Fatal("response is nil")
	}
	if result.Response.ID != "msg-123" {
		t.Errorf("response.ID = %q, want %q", result.Response.ID, "msg-123")
	}

	// Check tracking
	if len(tracker.tracked) != 1 {
		t.Fatalf("expected 1 tracked request, got %d", len(tracker.tracked))
	}
	tracked := tracker.tracked[0]
	if tracked.Status != store.EmailRequestStatusSuccess {
		t.Errorf("tracked.Status = %q, want %q", tracked.Status, store.EmailRequestStatusSuccess)
	}
	if tracked.From != "sender@example.com" {
		t.Errorf("tracked.From = %q, want %q", tracked.From, "sender@example.com")
	}
	if len(tracked.To) != 1 || tracked.To[0] != "test@example.com" {
		t.Errorf("tracked.To = %v, want [test@example.com]", tracked.To)
	}
	if tracked.Subject != "Test" {
		t.Errorf("tracked.Subject = %q, want %q", tracked.Subject, "Test")
	}
	if !tracked.HasText {
		t.Error("tracked.HasText should be true")
	}
	if tracked.HasHTML {
		t.Error("tracked.HasHTML should be false")
	}
	if tracked.EmailID == nil || *tracked.EmailID != "msg-123" {
		t.Error("tracked.EmailID not set correctly")
	}
	if tracked.DurationMs < 0 {
		t.Error("tracked.DurationMs should be non-negative")
	}
}

func TestSendWithTracking_Error(t *testing.T) {
	client := &mockClient{
		err: errors.New("send failed"),
	}
	tracker := &mockTracker{}

	tc := NewTrackedClient(client, tracker, "exec-123")
	result := tc.SendWithTracking("func-1", email.SendRequest{
		From:    "sender@example.com",
		To:      []string{"test@example.com"},
		Subject: "Test",
		Text:    "Hello",
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
	if tracked.Status != store.EmailRequestStatusError {
		t.Errorf("tracked.Status = %q, want %q", tracked.Status, store.EmailRequestStatusError)
	}
	if tracked.ErrorMessage == nil || *tracked.ErrorMessage != "send failed" {
		t.Error("tracked.ErrorMessage not set correctly")
	}
}

func TestSend_Wrapper(t *testing.T) {
	client := &mockClient{
		response: &email.SendResponse{
			ID: "msg-123",
		},
	}
	tracker := &mockTracker{}

	tc := NewTrackedClient(client, tracker, "exec-123")
	response, err := tc.Send("func-1", email.SendRequest{
		From:    "sender@example.com",
		To:      []string{"test@example.com"},
		Subject: "Test",
		Text:    "Hello",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response == nil {
		t.Fatal("response is nil")
	}
	if response.ID != "msg-123" {
		t.Errorf("response.ID = %q, want %q", response.ID, "msg-123")
	}
}

func TestSendWithTracking_NilTracker(t *testing.T) {
	client := &mockClient{
		response: &email.SendResponse{
			ID: "msg-123",
		},
	}

	tc := NewTrackedClient(client, nil, "exec-123")
	result := tc.SendWithTracking("func-1", email.SendRequest{
		From:    "sender@example.com",
		To:      []string{"test@example.com"},
		Subject: "Test",
		Text:    "Hello",
	})

	// Should not panic with nil tracker
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Response == nil {
		t.Fatal("response is nil")
	}
}

func TestSendWithTracking_HTMLContent(t *testing.T) {
	client := &mockClient{
		response: &email.SendResponse{
			ID: "msg-123",
		},
	}
	tracker := &mockTracker{}

	tc := NewTrackedClient(client, tracker, "exec-123")
	tc.SendWithTracking("func-1", email.SendRequest{
		From:    "sender@example.com",
		To:      []string{"test@example.com"},
		Subject: "Test",
		HTML:    "<p>Hello</p>",
	})

	tracked := tracker.tracked[0]
	if tracked.HasText {
		t.Error("tracked.HasText should be false")
	}
	if !tracked.HasHTML {
		t.Error("tracked.HasHTML should be true")
	}
}

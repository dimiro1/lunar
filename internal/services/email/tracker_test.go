package email

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

	responseJSON := `{"id":"email_abc123"}`
	emailID := "email_abc123"

	req := TrackRequest{
		From:         "sender@example.com",
		To:           []string{"recipient@example.com"},
		Subject:      "Test Email",
		HasText:      true,
		HasHTML:      false,
		RequestJSON:  `{"from":"sender@example.com","to":["recipient@example.com"],"subject":"Test Email"}`,
		ResponseJSON: &responseJSON,
		Status:       store.EmailRequestStatusSuccess,
		EmailID:      &emailID,
		DurationMs:   250,
	}

	tracker.Track("exec-1", req)

	requests := tracker.Requests("exec-1")
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	tracked := requests[0]
	if tracked.From != "sender@example.com" {
		t.Errorf("expected from 'sender@example.com', got '%s'", tracked.From)
	}
	if len(tracked.To) != 1 || tracked.To[0] != "recipient@example.com" {
		t.Errorf("expected to ['recipient@example.com'], got %v", tracked.To)
	}
	if tracked.Subject != "Test Email" {
		t.Errorf("expected subject 'Test Email', got '%s'", tracked.Subject)
	}
	if tracked.ExecutionID != "exec-1" {
		t.Errorf("expected executionID 'exec-1', got '%s'", tracked.ExecutionID)
	}
	if tracked.Status != store.EmailRequestStatusSuccess {
		t.Errorf("expected status 'success', got '%s'", tracked.Status)
	}
	if tracked.DurationMs != 250 {
		t.Errorf("expected duration 250, got %d", tracked.DurationMs)
	}
	if !tracked.HasText {
		t.Error("expected HasText to be true")
	}
	if tracked.HasHTML {
		t.Error("expected HasHTML to be false")
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

	errMsg := "Invalid API key"
	req := TrackRequest{
		From:         "sender@example.com",
		To:           []string{"recipient@example.com"},
		Subject:      "Test Email",
		HasText:      true,
		HasHTML:      false,
		RequestJSON:  `{"from":"sender@example.com"}`,
		Status:       store.EmailRequestStatusError,
		ErrorMessage: &errMsg,
		DurationMs:   50,
	}

	tracker.Track("exec-1", req)

	requests := tracker.Requests("exec-1")
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	tracked := requests[0]
	if tracked.Status != store.EmailRequestStatusError {
		t.Errorf("expected status 'error', got '%s'", tracked.Status)
	}
	if tracked.ErrorMessage == nil || *tracked.ErrorMessage != "Invalid API key" {
		t.Errorf("expected error message 'Invalid API key'")
	}
	if tracked.EmailID != nil {
		t.Error("expected nil email ID for error request")
	}
	if tracked.ResponseJSON != nil {
		t.Error("expected nil response JSON for error request")
	}
}

func TestMemoryTracker_Track_MultipleRecipients(t *testing.T) {
	tracker := NewMemoryTracker()

	req := TrackRequest{
		From:        "sender@example.com",
		To:          []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		Subject:     "Broadcast Email",
		HasText:     true,
		HasHTML:     true,
		RequestJSON: `{}`,
		Status:      store.EmailRequestStatusSuccess,
	}

	tracker.Track("exec-1", req)

	requests := tracker.Requests("exec-1")
	tracked := requests[0]

	if len(tracked.To) != 3 {
		t.Errorf("expected 3 recipients, got %d", len(tracked.To))
	}
	if tracked.To[0] != "user1@example.com" {
		t.Errorf("expected first recipient 'user1@example.com', got '%s'", tracked.To[0])
	}
	if !tracked.HasText || !tracked.HasHTML {
		t.Error("expected both HasText and HasHTML to be true")
	}
}

func TestMemoryTracker_Requests_FiltersByExecutionID(t *testing.T) {
	tracker := NewMemoryTracker()

	req1 := TrackRequest{From: "a@example.com", To: []string{"b@example.com"}, Subject: "Email 1", Status: store.EmailRequestStatusSuccess}
	req2 := TrackRequest{From: "c@example.com", To: []string{"d@example.com"}, Subject: "Email 2", Status: store.EmailRequestStatusSuccess}
	req3 := TrackRequest{From: "e@example.com", To: []string{"f@example.com"}, Subject: "Email 3", Status: store.EmailRequestStatusSuccess}

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
	for range 5 {
		req := TrackRequest{From: "sender@example.com", To: []string{"recipient@example.com"}, Subject: "Test", Status: store.EmailRequestStatusSuccess}
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

	req := TrackRequest{From: "sender@example.com", To: []string{"recipient@example.com"}, Subject: "Test", Status: store.EmailRequestStatusSuccess}
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

	responseJSON := `{"api_key":"secret123","id":"email_abc"}`
	req := TrackRequest{
		From:         "sender@example.com",
		To:           []string{"recipient@example.com"},
		Subject:      "Test",
		RequestJSON:  `{"api_key":"mysecret","from":"sender@example.com"}`,
		ResponseJSON: &responseJSON,
		Status:       store.EmailRequestStatusSuccess,
	}

	tracker.Track("exec-1", req)

	requests := tracker.Requests("exec-1")
	tracked := requests[0]

	// The masking package should mask sensitive fields
	// Check that the original values are not present
	if tracked.RequestJSON == `{"api_key":"mysecret","from":"sender@example.com"}` {
		t.Error("expected request JSON to be masked")
	}
	if tracked.ResponseJSON != nil && *tracked.ResponseJSON == `{"api_key":"secret123","id":"email_abc"}` {
		t.Error("expected response JSON to be masked")
	}
}

func TestMemoryTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewMemoryTracker()

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func() {
			req := TrackRequest{From: "sender@example.com", To: []string{"recipient@example.com"}, Subject: "Test", Status: store.EmailRequestStatusSuccess}
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

func TestEmailParamsToJSON(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		to          []string
		subject     string
		text        string
		html        string
		replyTo     string
		cc          []string
		bcc         []string
		scheduledAt string
		headers     map[string]string
		tags        []map[string]string
		wantContain []string
	}{
		{
			name:        "basic email",
			from:        "sender@example.com",
			to:          []string{"recipient@example.com"},
			subject:     "Test Subject",
			text:        "Hello",
			wantContain: []string{`"from":"sender@example.com"`, `"to":["recipient@example.com"]`, `"subject":"Test Subject"`, `"text":"Hello"`},
		},
		{
			name:        "with html",
			from:        "sender@example.com",
			to:          []string{"recipient@example.com"},
			subject:     "Test",
			html:        "<h1>Hello</h1>",
			wantContain: []string{`"html":"`}, // JSON escapes < and > so we just check the key exists
		},
		{
			name:        "with reply_to",
			from:        "sender@example.com",
			to:          []string{"recipient@example.com"},
			subject:     "Test",
			text:        "Hello",
			replyTo:     "reply@example.com",
			wantContain: []string{`"reply_to":"reply@example.com"`},
		},
		{
			name:        "with cc and bcc",
			from:        "sender@example.com",
			to:          []string{"recipient@example.com"},
			subject:     "Test",
			text:        "Hello",
			cc:          []string{"cc@example.com"},
			bcc:         []string{"bcc@example.com"},
			wantContain: []string{`"cc":["cc@example.com"]`, `"bcc":["bcc@example.com"]`},
		},
		{
			name:        "with scheduled_at",
			from:        "sender@example.com",
			to:          []string{"recipient@example.com"},
			subject:     "Test",
			text:        "Hello",
			scheduledAt: "2024-01-01T10:00:00Z",
			wantContain: []string{`"scheduled_at":"2024-01-01T10:00:00Z"`},
		},
		{
			name:        "with headers",
			from:        "sender@example.com",
			to:          []string{"recipient@example.com"},
			subject:     "Test",
			text:        "Hello",
			headers:     map[string]string{"X-Custom": "value"},
			wantContain: []string{`"headers":`},
		},
		{
			name:        "with tags",
			from:        "sender@example.com",
			to:          []string{"recipient@example.com"},
			subject:     "Test",
			text:        "Hello",
			tags:        []map[string]string{{"name": "campaign", "value": "test"}},
			wantContain: []string{`"tags":`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EmailParamsToJSON(tt.from, tt.to, tt.subject, tt.text, tt.html, tt.replyTo, tt.cc, tt.bcc, tt.scheduledAt, tt.headers, tt.tags)

			for _, want := range tt.wantContain {
				if !contains(result, want) {
					t.Errorf("expected result to contain %q, got %s", want, result)
				}
			}
		})
	}
}

func TestEmailParamsToJSON_EmptyOptionalFields(t *testing.T) {
	result := EmailParamsToJSON("sender@example.com", []string{"recipient@example.com"}, "Subject", "", "", "", nil, nil, "", nil, nil)

	// Should only contain from, to, subject
	if !contains(result, `"from"`) || !contains(result, `"to"`) || !contains(result, `"subject"`) {
		t.Error("expected result to contain from, to, subject")
	}

	// Should not contain optional fields
	if contains(result, `"text"`) || contains(result, `"html"`) || contains(result, `"reply_to"`) {
		t.Error("expected result not to contain empty optional fields")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package email

import (
	"testing"
)

func TestNewDefaultClient(t *testing.T) {
	mockEnvStore := &mockEnvStore{
		values: map[string]map[string]string{},
	}

	client := NewDefaultClient(mockEnvStore)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.envStore == nil {
		t.Error("expected non-nil envStore")
	}
}

func TestDefaultClient_Send_MissingAPIKey(t *testing.T) {
	mockEnvStore := &mockEnvStore{
		values: map[string]map[string]string{
			"func-1": {}, // No API key
		},
	}

	client := NewDefaultClient(mockEnvStore)

	req := SendRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Text:    "Hello",
	}

	_, err := client.Send("func-1", req)

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	configErr, ok := err.(*ConfigError)
	if !ok {
		t.Fatalf("expected ConfigError, got %T", err)
	}
	if configErr.Field != ResendAPIKeyEnv {
		t.Errorf("expected field %s, got %s", ResendAPIKeyEnv, configErr.Field)
	}
}

func TestConfigError_Error(t *testing.T) {
	err := &ConfigError{Field: "RESEND_API_KEY"}

	expected := "RESEND_API_KEY not set in function environment"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestSendRequest_Fields(t *testing.T) {
	req := SendRequest{
		From:        "sender@example.com",
		To:          []string{"recipient@example.com"},
		Subject:     "Test Subject",
		Text:        "Plain text content",
		HTML:        "<h1>HTML content</h1>",
		ReplyTo:     "reply@example.com",
		Cc:          []string{"cc@example.com"},
		Bcc:         []string{"bcc@example.com"},
		Headers:     map[string]string{"X-Custom": "value"},
		Tags:        []Tag{{Name: "campaign", Value: "test"}},
		ScheduledAt: "2024-01-01T10:00:00Z",
	}

	if req.From != "sender@example.com" {
		t.Errorf("expected From 'sender@example.com', got '%s'", req.From)
	}
	if len(req.To) != 1 || req.To[0] != "recipient@example.com" {
		t.Errorf("expected To ['recipient@example.com'], got %v", req.To)
	}
	if req.Subject != "Test Subject" {
		t.Errorf("expected Subject 'Test Subject', got '%s'", req.Subject)
	}
	if req.Text != "Plain text content" {
		t.Errorf("expected Text 'Plain text content', got '%s'", req.Text)
	}
	if req.HTML != "<h1>HTML content</h1>" {
		t.Errorf("expected HTML '<h1>HTML content</h1>', got '%s'", req.HTML)
	}
	if req.ReplyTo != "reply@example.com" {
		t.Errorf("expected ReplyTo 'reply@example.com', got '%s'", req.ReplyTo)
	}
	if len(req.Cc) != 1 || req.Cc[0] != "cc@example.com" {
		t.Errorf("expected Cc ['cc@example.com'], got %v", req.Cc)
	}
	if len(req.Bcc) != 1 || req.Bcc[0] != "bcc@example.com" {
		t.Errorf("expected Bcc ['bcc@example.com'], got %v", req.Bcc)
	}
	if req.Headers["X-Custom"] != "value" {
		t.Errorf("expected Headers['X-Custom'] 'value', got '%s'", req.Headers["X-Custom"])
	}
	if len(req.Tags) != 1 || req.Tags[0].Name != "campaign" || req.Tags[0].Value != "test" {
		t.Errorf("expected Tags [{campaign, test}], got %v", req.Tags)
	}
	if req.ScheduledAt != "2024-01-01T10:00:00Z" {
		t.Errorf("expected ScheduledAt '2024-01-01T10:00:00Z', got '%s'", req.ScheduledAt)
	}
}

func TestTag_Fields(t *testing.T) {
	tag := Tag{
		Name:  "category",
		Value: "newsletter",
	}

	if tag.Name != "category" {
		t.Errorf("expected Name 'category', got '%s'", tag.Name)
	}
	if tag.Value != "newsletter" {
		t.Errorf("expected Value 'newsletter', got '%s'", tag.Value)
	}
}

func TestSendResponse_Fields(t *testing.T) {
	resp := SendResponse{
		ID:          "email_abc123",
		RequestJSON: `{"from":"sender@example.com"}`,
	}

	if resp.ID != "email_abc123" {
		t.Errorf("expected ID 'email_abc123', got '%s'", resp.ID)
	}
	if resp.RequestJSON != `{"from":"sender@example.com"}` {
		t.Errorf("expected RequestJSON '{\"from\":\"sender@example.com\"}', got '%s'", resp.RequestJSON)
	}
}

// mockEnvStore is a mock implementation of env.Store for testing
type mockEnvStore struct {
	values map[string]map[string]string
}

func (m *mockEnvStore) Get(functionID, key string) (string, error) {
	if funcVars, ok := m.values[functionID]; ok {
		if val, ok := funcVars[key]; ok {
			return val, nil
		}
	}
	return "", nil
}

func (m *mockEnvStore) Set(functionID, key, value string) error {
	if m.values[functionID] == nil {
		m.values[functionID] = make(map[string]string)
	}
	m.values[functionID][key] = value
	return nil
}

func (m *mockEnvStore) Delete(functionID, key string) error {
	if funcVars, ok := m.values[functionID]; ok {
		delete(funcVars, key)
	}
	return nil
}

func (m *mockEnvStore) All(functionID string) (map[string]string, error) {
	if funcVars, ok := m.values[functionID]; ok {
		return funcVars, nil
	}
	return map[string]string{}, nil
}

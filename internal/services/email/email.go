package email

import (
	"net/url"

	"github.com/dimiro1/lunar/internal/services/env"
	"github.com/resend/resend-go/v3"
)

// Environment variable names for Resend configuration
const (
	ResendAPIKeyEnv  = "RESEND_API_KEY"
	ResendBaseURLEnv = "RESEND_BASE_URL"
)

// Tag represents an email tag for categorization
type Tag struct {
	Name  string
	Value string
}

// SendRequest represents a request to send an email
type SendRequest struct {
	From        string
	To          []string
	Subject     string
	Text        string
	HTML        string
	ReplyTo     string
	Cc          []string
	Bcc         []string
	Headers     map[string]string
	Tags        []Tag
	ScheduledAt string
}

// SendResponse represents the response from sending an email
type SendResponse struct {
	ID          string
	RequestJSON string
}

// Client is an interface for sending emails
type Client interface {
	Send(functionID string, req SendRequest) (*SendResponse, error)
}

// DefaultClient is the default implementation of Client using Resend
type DefaultClient struct {
	envStore env.Store
}

// NewDefaultClient creates a new email client
func NewDefaultClient(envStore env.Store) *DefaultClient {
	return &DefaultClient{
		envStore: envStore,
	}
}

// Send sends an email using Resend
func (c *DefaultClient) Send(functionID string, req SendRequest) (*SendResponse, error) {
	// Get API key from environment
	apiKey, err := c.envStore.Get(functionID, ResendAPIKeyEnv)
	if err != nil || apiKey == "" {
		return nil, &ConfigError{Field: ResendAPIKeyEnv}
	}

	// Create Resend client
	client := resend.NewClient(apiKey)

	// Allow custom base URL for testing (read from function env)
	if baseURL, err := c.envStore.Get(functionID, ResendBaseURLEnv); err == nil && baseURL != "" {
		if parsedURL, err := url.Parse(baseURL); err == nil {
			client.BaseURL = parsedURL
		}
	}

	// Build Resend request params
	params := &resend.SendEmailRequest{
		From:    req.From,
		To:      req.To,
		Subject: req.Subject,
		Text:    req.Text,
		Html:    req.HTML,
		ReplyTo: req.ReplyTo,
	}

	if len(req.Cc) > 0 {
		params.Cc = req.Cc
	}
	if len(req.Bcc) > 0 {
		params.Bcc = req.Bcc
	}
	if len(req.Headers) > 0 {
		params.Headers = req.Headers
	}
	if len(req.Tags) > 0 {
		var tags []resend.Tag
		for _, t := range req.Tags {
			tags = append(tags, resend.Tag{Name: t.Name, Value: t.Value})
		}
		params.Tags = tags
	}
	if req.ScheduledAt != "" {
		params.ScheduledAt = req.ScheduledAt
	}

	// Build request JSON for tracking
	var tagsForJSON []map[string]string
	for _, tag := range req.Tags {
		tagsForJSON = append(tagsForJSON, map[string]string{
			"name":  tag.Name,
			"value": tag.Value,
		})
	}
	requestJSON := EmailParamsToJSON(req.From, req.To, req.Subject, req.Text, req.HTML, req.ReplyTo, req.Cc, req.Bcc, req.ScheduledAt, req.Headers, tagsForJSON)

	// Send email
	sent, err := client.Emails.Send(params)
	if err != nil {
		return &SendResponse{RequestJSON: requestJSON}, err
	}

	return &SendResponse{
		ID:          sent.Id,
		RequestJSON: requestJSON,
	}, nil
}

// ConfigError is returned when a required configuration is missing
type ConfigError struct {
	Field string
}

func (e *ConfigError) Error() string {
	return e.Field + " not set in function environment"
}

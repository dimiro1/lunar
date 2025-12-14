package email

import (
	"testing"

	"github.com/dimiro1/lunar/internal/services/email"
)

func TestValidateSendRequest(t *testing.T) {
	tests := []struct {
		name      string
		req       email.SendRequest
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid request with text",
			req: email.SendRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Text:    "Hello, World!",
			},
			expectErr: false,
		},
		{
			name: "valid request with HTML",
			req: email.SendRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				HTML:    "<p>Hello, World!</p>",
			},
			expectErr: false,
		},
		{
			name: "valid request with both text and HTML",
			req: email.SendRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Text:    "Hello, World!",
				HTML:    "<p>Hello, World!</p>",
			},
			expectErr: false,
		},
		{
			name: "multiple recipients",
			req: email.SendRequest{
				From:    "sender@example.com",
				To:      []string{"a@example.com", "b@example.com"},
				Subject: "Test Subject",
				Text:    "Hello!",
			},
			expectErr: false,
		},
		{
			name: "missing from",
			req: email.SendRequest{
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Text:    "Hello!",
			},
			expectErr: true,
			errMsg:    "from is required",
		},
		{
			name: "missing to",
			req: email.SendRequest{
				From:    "sender@example.com",
				Subject: "Test Subject",
				Text:    "Hello!",
			},
			expectErr: true,
			errMsg:    "to is required",
		},
		{
			name: "empty to array",
			req: email.SendRequest{
				From:    "sender@example.com",
				To:      []string{},
				Subject: "Test Subject",
				Text:    "Hello!",
			},
			expectErr: true,
			errMsg:    "to is required",
		},
		{
			name: "missing subject",
			req: email.SendRequest{
				From: "sender@example.com",
				To:   []string{"recipient@example.com"},
				Text: "Hello!",
			},
			expectErr: true,
			errMsg:    "subject is required",
		},
		{
			name: "missing content",
			req: email.SendRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
			},
			expectErr: true,
			errMsg:    "either text or html content is required",
		},
		{
			name: "empty text and HTML",
			req: email.SendRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Text:    "",
				HTML:    "",
			},
			expectErr: true,
			errMsg:    "either text or html content is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSendRequest(tt.req)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

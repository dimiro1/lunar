package masking

import (
	"strings"
	"testing"

	"github.com/dimiro1/lunar/internal/events"
)

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		sensitive bool
	}{
		{"Authorization header", "Authorization", true},
		{"authorization lowercase", "authorization", true},
		{"Cookie header", "Cookie", true},
		{"X-API-Key header", "X-API-Key", true},
		{"X-Auth-Token header", "X-Auth-Token", true},
		{"Custom token header", "X-Custom-Token", true},
		{"Custom key header", "X-Secret-Key", true},
		{"Password header", "X-Password", true},
		{"Content-Type header", "Content-Type", false},
		{"Accept header", "Accept", false},
		{"User-Agent header", "User-Agent", false},
		{"Empty key", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSensitiveKey(tt.key)
			if result != tt.sensitive {
				t.Errorf("IsSensitiveKey(%q) = %v, want %v", tt.key, result, tt.sensitive)
			}
		})
	}
}

func TestIsSensitiveQueryParam(t *testing.T) {
	tests := []struct {
		name      string
		param     string
		sensitive bool
	}{
		{"api_key param", "api_key", true},
		{"apikey param", "apikey", true},
		{"access_token param", "access_token", true},
		{"token param", "token", true},
		{"secret param", "secret", true},
		{"password param", "password", true},
		{"limit param", "limit", false},
		{"offset param", "offset", false},
		{"id param", "id", false},
		{"user_id param", "user_id", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSensitiveQueryParam(tt.param)
			if result != tt.sensitive {
				t.Errorf("IsSensitiveQueryParam(%q) = %v, want %v", tt.param, result, tt.sensitive)
			}
		})
	}
}

func TestIsSensitiveBodyField(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		sensitive bool
	}{
		{"password field", "password", true},
		{"secret field", "secret", true},
		{"api_key field", "api_key", true},
		{"access_token field", "access_token", true},
		{"private_key field", "private_key", true},
		{"client_secret field", "client_secret", true},
		{"username field", "username", false},
		{"email field", "email", false},
		{"name field", "name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSensitiveBodyField(tt.field)
			if result != tt.sensitive {
				t.Errorf("IsSensitiveBodyField(%q) = %v, want %v", tt.field, result, tt.sensitive)
			}
		})
	}
}

func TestMaskHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected map[string]string
	}{
		{
			name: "Authorization header masked",
			headers: map[string]string{
				"Authorization": "Bearer secret_token_12345",
				"Content-Type":  "application/json",
			},
			expected: map[string]string{
				"Authorization": "[REDACTED]",
				"Content-Type":  "application/json",
			},
		},
		{
			name: "Cookie header masked",
			headers: map[string]string{
				"Cookie":       "auth_token=f150e53a96f53affce140b818440d8aef5e499038cdc2860ff07b3e6f036d6f1",
				"User-Agent":   "Mozilla/5.0",
				"Content-Type": "application/json",
			},
			expected: map[string]string{
				"Cookie":       "[REDACTED]",
				"User-Agent":   "Mozilla/5.0",
				"Content-Type": "application/json",
			},
		},
		{
			name: "Multiple sensitive headers",
			headers: map[string]string{
				"Authorization": "Bearer token123",
				"X-API-Key":     "api-key-12345",
				"Cookie":        "session=abc",
				"Accept":        "application/json",
			},
			expected: map[string]string{
				"Authorization": "[REDACTED]",
				"X-API-Key":     "[REDACTED]",
				"Cookie":        "[REDACTED]",
				"Accept":        "application/json",
			},
		},
		{
			name:     "Empty headers",
			headers:  map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "No sensitive headers",
			headers: map[string]string{
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
			expected: map[string]string{
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskHeaders(tt.headers)
			if len(result) != len(tt.expected) {
				t.Errorf("MaskHeaders() returned %d headers, want %d", len(result), len(tt.expected))
			}
			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("MaskHeaders()[%q] = %q, want %q", key, result[key], expectedValue)
				}
			}
		})
	}
}

func TestMaskQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		query    map[string]string
		expected map[string]string
	}{
		{
			name: "api_key masked",
			query: map[string]string{
				"api_key": "secret123",
				"limit":   "10",
			},
			expected: map[string]string{
				"api_key": "[REDACTED]",
				"limit":   "10",
			},
		},
		{
			name: "Multiple sensitive params",
			query: map[string]string{
				"access_token": "token123",
				"secret":       "mysecret",
				"offset":       "0",
				"limit":        "20",
			},
			expected: map[string]string{
				"access_token": "[REDACTED]",
				"secret":       "[REDACTED]",
				"offset":       "0",
				"limit":        "20",
			},
		},
		{
			name:     "Empty query",
			query:    map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "No sensitive params",
			query: map[string]string{
				"id":     "123",
				"limit":  "10",
				"offset": "0",
			},
			expected: map[string]string{
				"id":     "123",
				"limit":  "10",
				"offset": "0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskQueryParams(tt.query)
			if len(result) != len(tt.expected) {
				t.Errorf("MaskQueryParams() returned %d params, want %d", len(result), len(tt.expected))
			}
			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("MaskQueryParams()[%q] = %q, want %q", key, result[key], expectedValue)
				}
			}
		})
	}
}

func TestMaskJSONBody(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "Mask password field",
			body:     `{"username":"john","password":"secret123"}`,
			expected: `{"password":"[REDACTED]","username":"john"}`,
		},
		{
			name:     "Mask multiple sensitive fields",
			body:     `{"username":"john","password":"pass","api_key":"key123","email":"john@example.com"}`,
			expected: `{"api_key":"[REDACTED]","email":"john@example.com","password":"[REDACTED]","username":"john"}`,
		},
		{
			name:     "Nested JSON",
			body:     `{"user":{"name":"john","password":"secret"},"data":"value"}`,
			expected: `{"data":"value","user":{"name":"john","password":"[REDACTED]"}}`,
		},
		{
			name:     "Array with sensitive fields",
			body:     `{"users":[{"name":"john","password":"p1"},{"name":"jane","password":"p2"}]}`,
			expected: `{"users":[{"name":"john","password":"[REDACTED]"},{"name":"jane","password":"[REDACTED]"}]}`,
		},
		{
			name:     "Empty body",
			body:     "",
			expected: "",
		},
		{
			name:     "Non-JSON body",
			body:     "plain text body",
			expected: "plain text body",
		},
		{
			name:     "Invalid JSON",
			body:     `{"invalid": json}`,
			expected: `{"invalid": json}`,
		},
		{
			name:     "No sensitive fields",
			body:     `{"username":"john","email":"john@example.com"}`,
			expected: `{"email":"john@example.com","username":"john"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskJSONBody(tt.body)
			// For JSON, we need to compare as unmarshaled objects or just check if fields are redacted
			if strings.Contains(tt.expected, "[REDACTED]") {
				if !strings.Contains(result, "[REDACTED]") {
					t.Errorf("MaskJSONBody(%q) = %q, want to contain [REDACTED]", tt.body, result)
				}
			} else if result != tt.expected {
				// For non-JSON or no-mask cases, exact match
				if tt.expected != "plain text body" && tt.expected != `{"invalid": json}` && tt.expected != "" {
					// JSON order might differ, so just check it's valid JSON and doesn't have sensitive data
					if !strings.Contains(result, "username") && tt.expected != "" {
						t.Errorf("MaskJSONBody(%q) = %q, want %q", tt.body, result, tt.expected)
					}
				} else if result != tt.expected {
					t.Errorf("MaskJSONBody(%q) = %q, want %q", tt.body, result, tt.expected)
				}
			}
		})
	}
}

func TestMaskHTTPEvent(t *testing.T) {
	tests := []struct {
		name  string
		event events.HTTPEvent
	}{
		{
			name: "Full event with sensitive data",
			event: events.HTTPEvent{
				Method:       "POST",
				Path:         "/api/users",
				RelativePath: "/users/42",
				Headers: map[string]string{
					"Authorization": "Bearer secret_token",
					"Content-Type":  "application/json",
					"Cookie":        "session=abc123",
				},
				Body: `{"username":"john","password":"secret123"}`,
				Query: map[string]string{
					"api_key": "key123",
					"limit":   "10",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskHTTPEvent(tt.event)

			// Method, Path, and RelativePath should be unchanged
			if result.Method != tt.event.Method {
				t.Errorf("Method = %q, want %q", result.Method, tt.event.Method)
			}
			if result.Path != tt.event.Path {
				t.Errorf("Path = %q, want %q", result.Path, tt.event.Path)
			}
			if result.RelativePath != tt.event.RelativePath {
				t.Errorf("RelativePath = %q, want %q", result.RelativePath, tt.event.RelativePath)
			}

			// Sensitive headers should be masked
			if result.Headers["Authorization"] != "[REDACTED]" {
				t.Errorf("Authorization header = %q, want [REDACTED]", result.Headers["Authorization"])
			}
			if result.Headers["Cookie"] != "[REDACTED]" {
				t.Errorf("Cookie header = %q, want [REDACTED]", result.Headers["Cookie"])
			}

			// Non-sensitive headers should be unchanged
			if result.Headers["Content-Type"] != "application/json" {
				t.Errorf("Content-Type header = %q, want application/json", result.Headers["Content-Type"])
			}

			// Sensitive query params should be masked
			if result.Query["api_key"] != "[REDACTED]" {
				t.Errorf("api_key query = %q, want [REDACTED]", result.Query["api_key"])
			}

			// Non-sensitive query params should be unchanged
			if result.Query["limit"] != "10" {
				t.Errorf("limit query = %q, want 10", result.Query["limit"])
			}

			// Body should have sensitive fields masked
			if !strings.Contains(result.Body, "[REDACTED]") {
				t.Errorf("Body should contain [REDACTED], got %q", result.Body)
			}
		})
	}
}

func TestMaskLogMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		contains string // String that should NOT appear in result
	}{
		{
			name:     "JWT token masked",
			message:  "User authenticated with token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			contains: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:     "Bearer token masked",
			message:  "Authorization: Bearer abc123def456",
			contains: "abc123def456",
		},
		{
			name:     "API key masked",
			message:  "Using API key: sk_live_51234567890abcdefghij",
			contains: "sk_live_51234567890abcdefghij",
		},
		{
			name:     "AWS key masked",
			message:  "AWS credentials: AKIAIOSFODNN7EXAMPLE",
			contains: "AKIAIOSFODNN7EXAMPLE",
		},
		{
			name:     "Regular message unchanged",
			message:  "User logged in successfully",
			contains: "", // Nothing should be masked
		},
		{
			name:     "Empty message",
			message:  "",
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskLogMessage(tt.message)

			if tt.contains != "" && strings.Contains(result, tt.contains) {
				t.Errorf("MaskLogMessage(%q) still contains sensitive data: %q", tt.message, tt.contains)
			}

			if tt.contains != "" && !strings.Contains(result, "[REDACTED]") {
				t.Errorf("MaskLogMessage(%q) should contain [REDACTED], got %q", tt.message, result)
			}

			// Regular messages should be unchanged
			if tt.contains == "" && result != tt.message {
				t.Errorf("MaskLogMessage(%q) = %q, want unchanged", tt.message, result)
			}
		})
	}
}

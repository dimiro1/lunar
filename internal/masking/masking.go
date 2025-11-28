package masking

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/dimiro1/faas-go/internal/events"
)

const redactedValue = "[REDACTED]"

// Sensitive header patterns (case-insensitive)
var sensitiveHeaderPatterns = []string{
	"authorization",
	"cookie",
	"x-api-key",
	"x-auth-token",
	"proxy-authorization",
	"token",
	"key",
	"secret",
	"password",
	"auth",
}

// Sensitive query parameter patterns (case-insensitive)
var sensitiveQueryParams = []string{
	"api_key",
	"apikey",
	"access_token",
	"token",
	"secret",
	"password",
	"auth",
	"key",
}

// Sensitive JSON body field patterns (case-insensitive)
// These are exact matches or prefix matches (e.g., "token" matches "token" but not "completion_tokens")
var sensitiveBodyFields = []string{
	"password",
	"secret",
	"api_key",
	"apikey",
	"access_token",
	"private_key",
	"client_secret",
	"token",
	"auth",
	"authorization",
	"bearer",
	"credential",
	"credentials",
}

// Regex patterns for detecting sensitive data in log messages
var sensitiveLogPatterns = []*regexp.Regexp{
	// JWT tokens
	regexp.MustCompile(`eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*`),
	// Bearer tokens
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9_\-.]+`),
	// API keys (various formats including sk_live, sk_test, etc.)
	regexp.MustCompile(`(?i)(api[_-]?key|apikey|key)[\s:=]+[A-Za-z0-9_\-]{15,}`),
	// Generic tokens
	regexp.MustCompile(`(?i)(token|secret|password)[\s:=]+[A-Za-z0-9_\-]{10,}`),
	// AWS keys
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	// Base64 encoded secrets (40+ chars)
	regexp.MustCompile(`(?i)(secret|token|password)[\s:=]+[A-Za-z0-9+/]{40,}={0,2}`),
}

// IsSensitiveKey checks if a key name suggests it contains sensitive data
func IsSensitiveKey(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, pattern := range sensitiveHeaderPatterns {
		if strings.Contains(lowerKey, pattern) {
			return true
		}
	}
	return false
}

// IsSensitiveQueryParam checks if a query parameter name suggests it contains sensitive data
func IsSensitiveQueryParam(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, pattern := range sensitiveQueryParams {
		if strings.Contains(lowerKey, pattern) {
			return true
		}
	}
	return false
}

// IsSensitiveBodyField checks if a JSON body field name suggests it contains sensitive data
// Uses exact match or prefix match to avoid false positives (e.g., "completion_tokens" should not match "token")
func IsSensitiveBodyField(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, pattern := range sensitiveBodyFields {
		// Exact match
		if lowerKey == pattern {
			return true
		}
		// Prefix match with underscore or dash separator (e.g., "auth_token", "api-key")
		if strings.HasPrefix(lowerKey, pattern+"_") || strings.HasPrefix(lowerKey, pattern+"-") {
			return true
		}
	}
	return false
}

// MaskHeaders masks sensitive headers in a map
func MaskHeaders(headers map[string]string) map[string]string {
	masked := make(map[string]string, len(headers))
	for key, value := range headers {
		if IsSensitiveKey(key) {
			masked[key] = redactedValue
		} else {
			masked[key] = value
		}
	}
	return masked
}

// MaskQueryParams masks sensitive query parameters in a map
func MaskQueryParams(query map[string]string) map[string]string {
	masked := make(map[string]string, len(query))
	for key, value := range query {
		if IsSensitiveQueryParam(key) {
			masked[key] = redactedValue
		} else {
			masked[key] = value
		}
	}
	return masked
}

// MaskJSONBody attempts to parse the body as JSON and mask sensitive fields
// If parsing fails, returns the original body unchanged
func MaskJSONBody(body string) string {
	if body == "" {
		return body
	}

	// Try to parse as JSON
	var data any
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// Not JSON or invalid JSON, return as-is
		return body
	}

	// Mask sensitive fields
	masked := maskJSONValue(data)

	// Marshal back to JSON
	maskedBytes, err := json.Marshal(masked)
	if err != nil {
		// If marshaling fails, return original
		return body
	}

	return string(maskedBytes)
}

// maskJSONValue recursively masks sensitive fields in JSON structures
func maskJSONValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		masked := make(map[string]any, len(v))
		for key, val := range v {
			if IsSensitiveBodyField(key) {
				masked[key] = redactedValue
			} else {
				masked[key] = maskJSONValue(val)
			}
		}
		return masked
	case []any:
		masked := make([]any, len(v))
		for i, val := range v {
			masked[i] = maskJSONValue(val)
		}
		return masked
	default:
		return v
	}
}

// MaskHTTPEvent creates a copy of the HTTPEvent with sensitive data masked
func MaskHTTPEvent(event events.HTTPEvent) events.HTTPEvent {
	return events.HTTPEvent{
		Method:  event.Method,
		Path:    event.Path,
		Headers: MaskHeaders(event.Headers),
		Body:    MaskJSONBody(event.Body),
		Query:   MaskQueryParams(event.Query),
	}
}

// MaskLogMessage masks sensitive patterns in log messages
func MaskLogMessage(message string) string {
	masked := message
	for _, pattern := range sensitiveLogPatterns {
		masked = pattern.ReplaceAllString(masked, redactedValue)
	}
	return masked
}

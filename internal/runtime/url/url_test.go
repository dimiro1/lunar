package url

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		rawURL    string
		expected  *ParsedURL
		expectErr bool
	}{
		{
			name:   "simple URL",
			rawURL: "https://example.com/path",
			expected: &ParsedURL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/path",
				Query:  map[string][]string{},
			},
		},
		{
			name:   "URL with query params",
			rawURL: "https://example.com/search?q=hello&page=1",
			expected: &ParsedURL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/search",
				Query: map[string][]string{
					"q":    {"hello"},
					"page": {"1"},
				},
			},
		},
		{
			name:   "URL with fragment",
			rawURL: "https://example.com/page#section",
			expected: &ParsedURL{
				Scheme:   "https",
				Host:     "example.com",
				Path:     "/page",
				Fragment: "section",
				Query:    map[string][]string{},
			},
		},
		{
			name:   "URL with user info",
			rawURL: "https://user:pass@example.com/path",
			expected: &ParsedURL{
				Scheme:   "https",
				Host:     "example.com",
				Path:     "/path",
				Username: "user",
				Password: "pass",
				Query:    map[string][]string{},
			},
		},
		{
			name:   "URL with port",
			rawURL: "https://example.com:8080/api",
			expected: &ParsedURL{
				Scheme: "https",
				Host:   "example.com:8080",
				Path:   "/api",
				Query:  map[string][]string{},
			},
		},
		{
			name:   "URL with multiple query values",
			rawURL: "https://example.com?tag=a&tag=b&tag=c",
			expected: &ParsedURL{
				Scheme: "https",
				Host:   "example.com",
				Query: map[string][]string{
					"tag": {"a", "b", "c"},
				},
			},
		},
		{
			name:      "invalid URL",
			rawURL:    "://invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.rawURL)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got nil", tt.rawURL)
				}
				return
			}
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.rawURL, err)
				return
			}
			if result.Scheme != tt.expected.Scheme {
				t.Errorf("Parse(%q).Scheme = %q, want %q", tt.rawURL, result.Scheme, tt.expected.Scheme)
			}
			if result.Host != tt.expected.Host {
				t.Errorf("Parse(%q).Host = %q, want %q", tt.rawURL, result.Host, tt.expected.Host)
			}
			if result.Path != tt.expected.Path {
				t.Errorf("Parse(%q).Path = %q, want %q", tt.rawURL, result.Path, tt.expected.Path)
			}
			if result.Fragment != tt.expected.Fragment {
				t.Errorf("Parse(%q).Fragment = %q, want %q", tt.rawURL, result.Fragment, tt.expected.Fragment)
			}
			if result.Username != tt.expected.Username {
				t.Errorf("Parse(%q).Username = %q, want %q", tt.rawURL, result.Username, tt.expected.Username)
			}
			if result.Password != tt.expected.Password {
				t.Errorf("Parse(%q).Password = %q, want %q", tt.rawURL, result.Password, tt.expected.Password)
			}
			if !reflect.DeepEqual(result.Query, tt.expected.Query) {
				t.Errorf("Parse(%q).Query = %v, want %v", tt.rawURL, result.Query, tt.expected.Query)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello+world"},
		{"hello+world", "hello%2Bworld"},
		{"a=b&c=d", "a%3Db%26c%3Dd"},
		{"special!@#$%", "special%21%40%23%24%25"},
		{"unicode 你好", "unicode+%E4%BD%A0%E5%A5%BD"},
		{"", ""},
		{"no-encoding-needed", "no-encoding-needed"},
	}

	for _, tt := range tests {
		result := Encode(tt.input)
		if result != tt.expected {
			t.Errorf("Encode(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		input     string
		expected  string
		expectErr bool
	}{
		{"hello+world", "hello world", false},
		{"hello%20world", "hello world", false},
		{"hello%2Bworld", "hello+world", false},
		{"a%3Db%26c%3Dd", "a=b&c=d", false},
		{"unicode+%E4%BD%A0%E5%A5%BD", "unicode 你好", false},
		{"", "", false},
		{"no-encoding", "no-encoding", false},
		{"%ZZ", "", true}, // Invalid hex
	}

	for _, tt := range tests {
		result, err := Decode(tt.input)
		if tt.expectErr {
			if err == nil {
				t.Errorf("Decode(%q) expected error, got nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("Decode(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Decode(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	inputs := []string{
		"hello world",
		"special chars: !@#$%^&*()",
		"unicode: 你好",
		"path/to/resource",
		"query=value&other=123",
	}

	for _, input := range inputs {
		encoded := Encode(input)
		decoded, err := Decode(encoded)
		if err != nil {
			t.Errorf("Round trip failed for %q: %v", input, err)
			continue
		}
		if decoded != input {
			t.Errorf("Round trip mismatch: %q -> %q -> %q", input, encoded, decoded)
		}
	}
}

package router

import (
	"reflect"
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		pattern     string
		wantMatched bool
		wantParams  map[string]string
	}{
		// Exact matches
		{"exact root", "/", "/", true, map[string]string{}},
		{"exact path", "/users", "/users", true, map[string]string{}},
		{"exact nested", "/users/list", "/users/list", true, map[string]string{}},

		// No match
		{"no match different", "/users", "/posts", false, nil},
		{"no match extra segment", "/users/123", "/users", false, nil},
		{"no match missing segment", "/users", "/users/123", false, nil},

		// Parameter extraction
		{"single param", "/users/123", "/users/:id", true, map[string]string{"id": "123"}},
		{"multiple params", "/users/123/posts/456", "/users/:userId/posts/:postId", true, map[string]string{"userId": "123", "postId": "456"}},
		{"param at start", "/123/profile", "/:id/profile", true, map[string]string{"id": "123"}},

		// Wildcard
		{"wildcard", "/api/v1/users", "/api/*", true, map[string]string{}},
		{"wildcard nested", "/static/css/main.css", "/static/*", true, map[string]string{}},

		// Trailing slashes
		{"trailing slash path", "/users/", "/users", true, map[string]string{}},
		{"trailing slash pattern", "/users", "/users/", true, map[string]string{}},

		// Edge cases
		{"empty vs root", "", "/", true, map[string]string{}},
		{"param with hyphen", "/users/john-doe", "/users/:name", true, map[string]string{"name": "john-doe"}},
		{"param with underscore", "/users/john_doe", "/users/:name", true, map[string]string{"name": "john_doe"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Match(tt.path, tt.pattern)
			if result.Matched != tt.wantMatched {
				t.Errorf("Match(%q, %q).Matched = %v, want %v", tt.path, tt.pattern, result.Matched, tt.wantMatched)
			}
			if tt.wantMatched && !reflect.DeepEqual(result.Params, tt.wantParams) {
				t.Errorf("Match(%q, %q).Params = %v, want %v", tt.path, tt.pattern, result.Params, tt.wantParams)
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"/", []string{}},
		{"", []string{}},
		{"/users", []string{"users"}},
		{"/users/123", []string{"users", "123"}},
		{"/users/123/posts", []string{"users", "123", "posts"}},
		{"users/123", []string{"users", "123"}},
		{"/users/", []string{"users"}},
		{"///multiple///slashes///", []string{"multiple", "slashes"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := SplitPath(tt.path)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SplitPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestBuildPath(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		params   map[string]string
		expected string
	}{
		{"no params", "/users", nil, "/users"},
		{"no params empty map", "/users", map[string]string{}, "/users"},
		{"single param", "/users/:id", map[string]string{"id": "123"}, "/users/123"},
		{"multiple params", "/users/:userId/posts/:postId", map[string]string{"userId": "42", "postId": "99"}, "/users/42/posts/99"},
		{"unused param", "/users/:id", map[string]string{"id": "123", "extra": "ignored"}, "/users/123"},
		{"missing param", "/users/:id", map[string]string{"other": "value"}, "/users/:id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPath(tt.pattern, tt.params)
			if result != tt.expected {
				t.Errorf("BuildPath(%q, %v) = %q, want %q", tt.pattern, tt.params, result, tt.expected)
			}
		})
	}
}

func TestFunctionPath(t *testing.T) {
	tests := []struct {
		name       string
		functionID string
		pattern    string
		params     map[string]string
		expected   string
	}{
		{"simple", "my-func", "/", nil, "/fn/my-func/"},
		{"with path", "my-func", "/users", nil, "/fn/my-func/users"},
		{"with params", "my-func", "/users/:id", map[string]string{"id": "123"}, "/fn/my-func/users/123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FunctionPath(tt.functionID, tt.pattern, tt.params)
			if result != tt.expected {
				t.Errorf("FunctionPath(%q, %q, %v) = %q, want %q", tt.functionID, tt.pattern, tt.params, result, tt.expected)
			}
		})
	}
}

func TestFunctionURL(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		functionID string
		pattern    string
		params     map[string]string
		expected   string
	}{
		{"simple", "https://api.example.com", "my-func", "/", nil, "https://api.example.com/fn/my-func/"},
		{"with trailing slash", "https://api.example.com/", "my-func", "/users", nil, "https://api.example.com/fn/my-func/users"},
		{"with params", "https://api.example.com", "my-func", "/users/:id", map[string]string{"id": "123"}, "https://api.example.com/fn/my-func/users/123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FunctionURL(tt.baseURL, tt.functionID, tt.pattern, tt.params)
			if result != tt.expected {
				t.Errorf("FunctionURL(%q, %q, %q, %v) = %q, want %q", tt.baseURL, tt.functionID, tt.pattern, tt.params, result, tt.expected)
			}
		})
	}
}

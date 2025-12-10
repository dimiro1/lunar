package runner

import (
	"testing"

	"github.com/dimiro1/lunar/internal/events"
	lua "github.com/yuin/gopher-lua"
)

func TestRouterMatch(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		// Basic matches
		{"exact match root", "/", "/", true},
		{"exact match simple", "/users", "/users", true},
		{"exact match nested", "/users/list", "/users/list", true},
		{"no match different path", "/users", "/posts", false},
		{"no match extra segment", "/users/42", "/users", false},
		{"no match missing segment", "/users", "/users/42", false},

		// Parameter matches
		{"param single", "/users/42", "/users/:id", true},
		{"param nested", "/users/42/posts", "/users/:id/posts", true},
		{"param multiple", "/users/42/posts/5", "/users/:userId/posts/:postId", true},
		{"param at start", "/42/profile", "/:id/profile", true},
		{"param no match extra", "/users/42/extra", "/users/:id", false},

		// Wildcard matches
		{"wildcard simple", "/files/a", "/files/*", true},
		{"wildcard nested", "/files/a/b/c", "/files/*", true},
		{"wildcard with prefix", "/api/v1/users/list", "/api/*", true},
		{"wildcard only", "/anything/here", "/*", true},
		{"wildcard no match empty", "/files", "/files/*", false},

		// Edge cases - trailing slashes
		{"trailing slash path", "/users/", "/users", true},
		{"trailing slash pattern", "/users", "/users/", true},
		{"both trailing slashes", "/users/", "/users/", true},
		{"empty vs root", "", "/", true},

		// Edge cases - special characters in path segments
		{"url-encoded space", "/users/john%20doe", "/users/:name", true},
		{"hyphenated name", "/users/john-doe", "/users/:name", true},
		{"underscore name", "/users/john_doe", "/users/:name", true},
		{"dots in name", "/files/report.pdf", "/files/:filename", true},
		{"unicode path", "/users/日本語", "/users/:name", true},
		{"numeric zero", "/users/0", "/users/:id", true},
		{"negative number", "/users/-1", "/users/:id", true},

		// Edge cases - pattern variations
		{"only param", "/:id", "/123", false},  // pattern has :id, path doesn't
		{"path only param", "/123", "/:id", true},
		{"multiple consecutive params", "/a/b", "/:first/:second", true},
		{"param followed by literal", "/42/profile", "/:id/profile", true},
		{"literal followed by param", "/profile/42", "/profile/:id", true},

		// Edge cases - wildcard behavior
		{"wildcard with multiple segments", "/api/v1/users/123/posts", "/api/*", true},
		{"wildcard deep nested", "/a/b/c/d/e/f", "/a/*", true},
		{"wildcard single segment after prefix", "/files/readme", "/files/*", true},

		// Edge cases - no match scenarios
		{"longer path no wildcard", "/a/b/c", "/a/b", false},
		{"shorter path", "/a", "/a/b", false},
		{"completely different", "/foo/bar", "/baz/qux", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			ctx := &events.ExecutionContext{
				FunctionID: "test-func",
				BaseURL:    "http://localhost:8080",
			}
			registerRouter(L, ctx)

			// Call router.match(path, pattern)
			err := L.DoString(`result = router.match("` + tt.path + `", "` + tt.pattern + `")`)
			if err != nil {
				t.Fatalf("Lua error: %v", err)
			}

			result := L.GetGlobal("result")
			if result.Type() != lua.LTBool {
				t.Fatalf("Expected bool, got %s", result.Type())
			}

			got := lua.LVAsBool(result)
			if got != tt.expected {
				t.Errorf("router.match(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.expected)
			}
		})
	}
}

func TestRouterParams(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pattern  string
		expected map[string]string
	}{
		{
			name:     "single param",
			path:     "/users/42",
			pattern:  "/users/:id",
			expected: map[string]string{"id": "42"},
		},
		{
			name:     "multiple params",
			path:     "/users/42/posts/5",
			pattern:  "/users/:userId/posts/:postId",
			expected: map[string]string{"userId": "42", "postId": "5"},
		},
		{
			name:     "param with string value",
			path:     "/users/john-doe",
			pattern:  "/users/:username",
			expected: map[string]string{"username": "john-doe"},
		},
		{
			name:     "no match returns empty",
			path:     "/users/42",
			pattern:  "/posts/:id",
			expected: map[string]string{},
		},
		{
			name:     "no params in pattern",
			path:     "/users",
			pattern:  "/users",
			expected: map[string]string{},
		},
		// Edge cases
		{
			name:     "url-encoded value",
			path:     "/users/john%20doe",
			pattern:  "/users/:name",
			expected: map[string]string{"name": "john%20doe"},
		},
		{
			name:     "unicode param value",
			path:     "/users/日本語",
			pattern:  "/users/:name",
			expected: map[string]string{"name": "日本語"},
		},
		{
			name:     "numeric param value",
			path:     "/items/0",
			pattern:  "/items/:id",
			expected: map[string]string{"id": "0"},
		},
		{
			name:     "dots in param value",
			path:     "/files/report.pdf",
			pattern:  "/files/:filename",
			expected: map[string]string{"filename": "report.pdf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			ctx := &events.ExecutionContext{
				FunctionID: "test-func",
				BaseURL:    "http://localhost:8080",
			}
			registerRouter(L, ctx)

			// Call router.params(path, pattern)
			err := L.DoString(`result = router.params("` + tt.path + `", "` + tt.pattern + `")`)
			if err != nil {
				t.Fatalf("Lua error: %v", err)
			}

			result := L.GetGlobal("result")
			if result.Type() != lua.LTTable {
				t.Fatalf("Expected table, got %s", result.Type())
			}

			tbl := result.(*lua.LTable)

			// Check all expected params are present
			for key, expectedValue := range tt.expected {
				value := tbl.RawGetString(key)
				if value == lua.LNil {
					t.Errorf("Missing param %q", key)
					continue
				}
				got := lua.LVAsString(value)
				if got != expectedValue {
					t.Errorf("Param %q = %q, want %q", key, got, expectedValue)
				}
			}

			// Check no extra params
			count := 0
			tbl.ForEach(func(_, _ lua.LValue) {
				count++
			})
			if count != len(tt.expected) {
				t.Errorf("Got %d params, want %d", count, len(tt.expected))
			}
		})
	}
}

func TestRouterPath(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "simple path",
			code:     `result = router.path("/users")`,
			expected: "/fn/test-func/users",
		},
		{
			name:     "root path",
			code:     `result = router.path("/")`,
			expected: "/fn/test-func/",
		},
		{
			name:     "path with param substitution",
			code:     `result = router.path("/users/:id", {id = "42"})`,
			expected: "/fn/test-func/users/42",
		},
		{
			name:     "path with multiple params",
			code:     `result = router.path("/users/:userId/posts/:postId", {userId = "123", postId = "456"})`,
			expected: "/fn/test-func/users/123/posts/456",
		},
		{
			name:     "path without params table",
			code:     `result = router.path("/users/:id")`,
			expected: "/fn/test-func/users/:id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			ctx := &events.ExecutionContext{
				FunctionID: "test-func",
				BaseURL:    "http://localhost:8080",
			}
			registerRouter(L, ctx)

			err := L.DoString(tt.code)
			if err != nil {
				t.Fatalf("Lua error: %v", err)
			}

			result := L.GetGlobal("result")
			got := lua.LVAsString(result)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestRouterURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		code     string
		expected string
	}{
		{
			name:     "simple url",
			baseURL:  "http://localhost:8080",
			code:     `result = router.url("/users")`,
			expected: "http://localhost:8080/fn/test-func/users",
		},
		{
			name:     "url with trailing slash in baseURL",
			baseURL:  "http://localhost:8080/",
			code:     `result = router.url("/users")`,
			expected: "http://localhost:8080/fn/test-func/users",
		},
		{
			name:     "url with param substitution",
			baseURL:  "https://api.example.com",
			code:     `result = router.url("/users/:id", {id = "42"})`,
			expected: "https://api.example.com/fn/test-func/users/42",
		},
		{
			name:     "url with multiple params",
			baseURL:  "http://localhost:8080",
			code:     `result = router.url("/users/:userId/posts/:postId", {userId = "123", postId = "456"})`,
			expected: "http://localhost:8080/fn/test-func/users/123/posts/456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			L := lua.NewState()
			defer L.Close()

			ctx := &events.ExecutionContext{
				FunctionID: "test-func",
				BaseURL:    tt.baseURL,
			}
			registerRouter(L, ctx)

			err := L.DoString(tt.code)
			if err != nil {
				t.Fatalf("Lua error: %v", err)
			}

			result := L.GetGlobal("result")
			got := lua.LVAsString(result)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		pattern        string
		expectMatch    bool
		expectedParams map[string]string
	}{
		{
			name:           "simple param extraction",
			path:           "/api/users/123",
			pattern:        "/api/users/:id",
			expectMatch:    true,
			expectedParams: map[string]string{"id": "123"},
		},
		{
			name:           "multiple params",
			path:           "/api/users/123/posts/456/comments/789",
			pattern:        "/api/users/:userId/posts/:postId/comments/:commentId",
			expectMatch:    true,
			expectedParams: map[string]string{"userId": "123", "postId": "456", "commentId": "789"},
		},
		{
			name:        "wildcard captures rest",
			path:        "/static/css/style.css",
			pattern:     "/static/*",
			expectMatch: true,
			// Wildcard doesn't capture named params
			expectedParams: map[string]string{},
		},
		{
			name:        "no match wrong literal",
			path:        "/api/users/123",
			pattern:     "/api/posts/:id",
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, params := matchPath(tt.path, tt.pattern)

			if matched != tt.expectMatch {
				t.Errorf("matchPath(%q, %q) matched = %v, want %v", tt.path, tt.pattern, matched, tt.expectMatch)
			}

			if tt.expectMatch && tt.expectedParams != nil {
				for key, expectedValue := range tt.expectedParams {
					if got, ok := params[key]; !ok {
						t.Errorf("Missing param %q", key)
					} else if got != expectedValue {
						t.Errorf("Param %q = %q, want %q", key, got, expectedValue)
					}
				}
			}
		})
	}
}

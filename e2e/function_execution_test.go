package e2e

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/dimiro1/lunar/internal/store"
)

// seedFunction creates a function and an initial active version directly in the
// store, bypassing the UI. Used by execution tests that exercise the runtime
// over HTTP rather than the browser.
func seedFunction(t *testing.T, env *testEnv, id, language, code string) {
	t.Helper()
	ctx := context.Background()
	if _, err := env.Store.CreateFunction(ctx, store.Function{
		ID:      id,
		Name:    id,
		EnvVars: map[string]string{},
	}); err != nil {
		t.Fatalf("CreateFunction(%q): %v", id, err)
	}
	if _, err := env.Store.CreateVersion(ctx, id, code, language, nil); err != nil {
		t.Fatalf("CreateVersion(%q): %v", id, err)
	}
}

// invoke sends an HTTP request to a function's execution endpoint and returns the
// response and its body. No auth is required for /fn execution.
func invoke(t *testing.T, env *testEnv, method, path, body string, headers map[string]string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, env.Server.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	return resp, string(b)
}

func TestExecute_Lua_HelloWorld(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "lua_hello", "lua", `
function handler(ctx, event)
  return {
    statusCode = 200,
    headers = { ["Content-Type"] = "text/plain" },
    body = "hello-lua"
  }
end`)

	resp, body := invoke(t, env, "GET", "/fn/lua_hello", "", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200 (body: %s)", resp.StatusCode, body)
	}
	if body != "hello-lua" {
		t.Errorf("body = %q, want hello-lua", body)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/plain" {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
	if resp.Header.Get("X-Execution-Id") == "" {
		t.Error("expected X-Execution-Id header to be set")
	}
}

func TestExecute_Starlark_HelloWorld(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "star_hello", "starlark", `
def handler(ctx, event):
    return {
        "statusCode": 200,
        "headers": {"Content-Type": "text/plain"},
        "body": "hello-starlark",
    }`)

	resp, body := invoke(t, env, "GET", "/fn/star_hello", "", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200 (body: %s)", resp.StatusCode, body)
	}
	if body != "hello-starlark" {
		t.Errorf("body = %q, want hello-starlark", body)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/plain" {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
}

func TestExecute_EventMapping(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "lua_echo", "lua", `
function handler(ctx, event)
  return {
    statusCode = 200,
    body = json.encode({
      method = event.method,
      rel = event.relativePath,
      q = event.query.foo,
      h = event.headers["X-Test"],
      body = event.body,
    })
  }
end`)
	seedFunction(t, env, "star_echo", "starlark", `
def handler(ctx, event):
    body, _ = json.encode({
        "method": event.method,
        "rel": event.relativePath,
        "q": event.query.get("foo", ""),
        "h": event.headers.get("X-Test", ""),
        "body": event.body,
    })
    return {"statusCode": 200, "body": body}`)

	for _, id := range []string{"lua_echo", "star_echo"} {
		resp, body := invoke(t, env, "POST", "/fn/"+id+"/sub/path?foo=bar",
			"the-body", map[string]string{"X-Test": "hi"})
		if resp.StatusCode != 200 {
			t.Fatalf("%s: status = %d (body: %s)", id, resp.StatusCode, body)
		}
		for _, want := range []string{
			`"method":"POST"`,
			`"rel":"/sub/path"`,
			`"q":"bar"`,
			`"h":"hi"`,
			`"body":"the-body"`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("%s: body %q missing %q", id, body, want)
			}
		}
	}
}

func TestExecute_KVPersistsAcrossInvocations(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "lua_counter", "lua", `
function handler(ctx, event)
  local n = tonumber(kv.get("c") or "0") + 1
  kv.set("c", tostring(n))
  return { statusCode = 200, body = tostring(n) }
end`)
	seedFunction(t, env, "star_counter", "starlark", `
def handler(ctx, event):
    n = int(kv.get("c") or "0") + 1
    kv.set("c", str(n))
    return {"statusCode": 200, "body": str(n)}`)

	for _, id := range []string{"lua_counter", "star_counter"} {
		_, first := invoke(t, env, "GET", "/fn/"+id, "", nil)
		_, second := invoke(t, env, "GET", "/fn/"+id, "", nil)
		if first != "1" || second != "2" {
			t.Errorf("%s: counter = %q,%q want 1,2", id, first, second)
		}
	}
}

func TestExecute_Starlark_JSONRoundTrip(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "star_json", "starlark", `
def handler(ctx, event):
    data, err = json.decode(event.body)
    if err != None:
        return {"statusCode": 400, "body": err}
    out, _ = json.encode({"next": data["n"] + 1})
    return {"statusCode": 200, "body": out}`)

	resp, body := invoke(t, env, "POST", "/fn/star_json", `{"n": 41}`, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"next":42`) {
		t.Errorf("body = %q, want next:42", body)
	}
}

func TestExecute_CustomStatusFromHandler(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "star_404", "starlark", `
def handler(ctx, event):
    return {"statusCode": 404, "body": "nope"}`)

	resp, body := invoke(t, env, "GET", "/fn/star_404", "", nil)
	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	if body != "nope" {
		t.Errorf("body = %q, want nope", body)
	}
}

func TestExecute_HandlerError(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "lua_boom", "lua",
		`function handler(ctx, event) error("boom") end`)
	seedFunction(t, env, "star_boom", "starlark",
		`def handler(ctx, event):
    fail("boom")`)

	for _, id := range []string{"lua_boom", "star_boom"} {
		resp, _ := invoke(t, env, "GET", "/fn/"+id, "", nil)
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("%s: status = %d, want 500", id, resp.StatusCode)
		}
	}
}

func TestExecute_NotFound(t *testing.T) {
	env := startTestServer(t)
	resp, _ := invoke(t, env, "GET", "/fn/does-not-exist", "", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestExecute_DisabledFunction(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "lua_disabled", "lua",
		`function handler(ctx, event) return { statusCode = 200 } end`)

	disabled := true
	if err := env.Store.UpdateFunction(context.Background(), "lua_disabled",
		store.UpdateFunctionRequest{Disabled: &disabled}); err != nil {
		t.Fatalf("UpdateFunction: %v", err)
	}

	resp, _ := invoke(t, env, "GET", "/fn/lua_disabled", "", nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
}

func TestExecute_EnvVar(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "lua_env", "lua",
		`function handler(ctx, event) return { statusCode = 200, body = env.get("GREETING") or "unset" } end`)
	seedFunction(t, env, "star_env", "starlark",
		`def handler(ctx, event):
    return {"statusCode": 200, "body": env.get("GREETING") or "unset"}`)

	if err := env.EnvStore.Set("lua_env", "GREETING", "hola"); err != nil {
		t.Fatalf("env set: %v", err)
	}
	if err := env.EnvStore.Set("star_env", "GREETING", "hej"); err != nil {
		t.Fatalf("env set: %v", err)
	}

	if _, body := invoke(t, env, "GET", "/fn/lua_env", "", nil); body != "hola" {
		t.Errorf("lua env body = %q, want hola", body)
	}
	if _, body := invoke(t, env, "GET", "/fn/star_env", "", nil); body != "hej" {
		t.Errorf("starlark env body = %q, want hej", body)
	}
}

func TestExecute_Router(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "star_router", "starlark", `
def handler(ctx, event):
    if router.match(event.relativePath, "/users/:id"):
        params = router.params(event.relativePath, "/users/:id")
        return {"statusCode": 200, "body": params["id"]}
    return {"statusCode": 404, "body": "no match"}`)

	resp, body := invoke(t, env, "GET", "/fn/star_router/users/99", "", nil)
	if resp.StatusCode != 200 || body != "99" {
		t.Errorf("got %d %q, want 200 99", resp.StatusCode, body)
	}

	resp2, _ := invoke(t, env, "GET", "/fn/star_router/posts/1", "", nil)
	if resp2.StatusCode != 404 {
		t.Errorf("unmatched route status = %d, want 404", resp2.StatusCode)
	}
}

// TestExecute_StarlarkKeywordArgs verifies the Starlark host functions accept
// keyword arguments (a Starlark-only ergonomic the Lua runtime cannot offer).
func TestExecute_StarlarkKeywordArgs(t *testing.T) {
	env := startTestServer(t)
	seedFunction(t, env, "star_kwargs", "starlark", `
def handler(ctx, event):
    kv.set(key="k", value="v")
    return {"statusCode": 200, "body": kv.get(key="k")}`)

	resp, body := invoke(t, env, "GET", "/fn/star_kwargs", "", nil)
	if resp.StatusCode != 200 || body != "v" {
		t.Errorf("got %d %q, want 200 v", resp.StatusCode, body)
	}
}

// TestExecute_LanguageStickyAcrossVersions verifies that a new version created
// without an explicit language keeps running on the function's language: a
// second Starlark version (language "") must still execute as Starlark.
func TestExecute_LanguageStickyAcrossVersions(t *testing.T) {
	env := startTestServer(t)
	ctx := context.Background()
	seedFunction(t, env, "sticky", "starlark",
		`def handler(ctx, event):
    return {"statusCode": 200, "body": "v1"}`)

	// New version, no language specified — must inherit Starlark.
	if _, err := env.Store.CreateVersion(ctx, "sticky",
		`def handler(ctx, event):
    return {"statusCode": 200, "body": "v2"}`, "", nil); err != nil {
		t.Fatalf("CreateVersion v2: %v", err)
	}

	resp, body := invoke(t, env, "GET", "/fn/sticky", "", nil)
	if resp.StatusCode != 200 || body != "v2" {
		t.Errorf("got %d %q, want 200 v2 (a Lua runtime would have failed to parse)", resp.StatusCode, body)
	}
}

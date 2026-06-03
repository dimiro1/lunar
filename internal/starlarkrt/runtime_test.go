package starlarkrt

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/dimiro1/lunar/internal/engine"
	"github.com/dimiro1/lunar/internal/events"
	"github.com/dimiro1/lunar/internal/services/env"
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
)

// newTestRuntime builds a StarlarkRuntime backed by in-memory/fake collaborators.
// The returned http fake lets tests configure responses.
func newTestRuntime() (*StarlarkRuntime, *internalhttp.FakeClient, kv.Store) {
	httpClient := internalhttp.NewFakeClient()
	kvStore := kv.NewMemoryStore()
	rt := New(Config{
		Logger: logger.NewMemoryLogger(),
		KV:     kvStore,
		Env:    env.NewMemoryStore(),
		HTTP:   httpClient,
	})
	return rt, httpClient, kvStore
}

func run(t *testing.T, rt *StarlarkRuntime, code string, event events.HTTPEvent) (*events.HTTPResponse, error) {
	t.Helper()
	execCtx := &events.ExecutionContext{
		ExecutionID:  "exec-1",
		FunctionID:   "fn-1",
		FunctionName: "MyFn",
		Version:      "v1",
		RequestID:    "req-1",
		BaseURL:      "https://example.com",
		StartedAt:    time.Now().Unix(),
	}
	res, err := rt.Execute(context.Background(), engine.RuntimeRequest{
		Code:    code,
		Context: execCtx,
		Event:   event,
	})
	if err != nil {
		return nil, err
	}
	return res.Response, nil
}

func TestExecute_Success(t *testing.T) {
	rt, _, _ := newTestRuntime()
	code := `
def handler(ctx, event):
    return {
        "statusCode": 201,
        "headers": {"Content-Type": "application/json"},
        "body": "Hello, World!",
    }
`
	resp, err := run(t, rt, code, events.HTTPEvent{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
	if resp.Body != "Hello, World!" {
		t.Errorf("body = %q", resp.Body)
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("headers = %v", resp.Headers)
	}
}

func TestExecute_DefaultStatus(t *testing.T) {
	rt, _, _ := newTestRuntime()
	code := `
def handler(ctx, event):
    return {"body": "ok"}
`
	resp, err := run(t, rt, code, events.HTTPEvent{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want default 200", resp.StatusCode)
	}
}

func TestExecute_AccessContextAndEvent(t *testing.T) {
	rt, _, _ := newTestRuntime()
	code := `
def handler(ctx, event):
    return {
        "statusCode": 200,
        "body": ctx.executionId + " " + ctx.functionName + " " + event.method + " " + event.path + " " + event.query["foo"],
    }
`
	event := events.HTTPEvent{Method: "POST", Path: "/x", Query: map[string]string{"foo": "bar"}}
	resp, err := run(t, rt, code, event)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	want := "exec-1 MyFn POST /x bar"
	if resp.Body != want {
		t.Errorf("body = %q, want %q", resp.Body, want)
	}
}

func TestExecute_JSONRoundTrip(t *testing.T) {
	rt, _, _ := newTestRuntime()
	code := `
def handler(ctx, event):
    decoded, err = json.decode(event.body)
    if err != None:
        return {"statusCode": 400, "body": err}
    encoded, err = json.encode({"name": decoded["name"], "n": decoded["n"] + 1})
    return {"statusCode": 200, "body": encoded}
`
	resp, err := run(t, rt, code, events.HTTPEvent{Method: "POST", Path: "/", Body: `{"name":"lunar","n":41}`})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !strings.Contains(resp.Body, `"name":"lunar"`) || !strings.Contains(resp.Body, `"n":42`) {
		t.Errorf("body = %q", resp.Body)
	}
}

func TestExecute_KVStore(t *testing.T) {
	rt, _, store := newTestRuntime()
	code := `
def handler(ctx, event):
    kv.set("greeting", "hi")
    return {"statusCode": 200, "body": kv.get("greeting")}
`
	resp, err := run(t, rt, code, events.HTTPEvent{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.Body != "hi" {
		t.Errorf("body = %q, want hi", resp.Body)
	}
	if got, _ := store.Get("fn-1", "greeting"); got != "hi" {
		t.Errorf("kv persisted = %q, want hi", got)
	}
}

func TestExecute_HTTPTuple(t *testing.T) {
	rt, httpClient, _ := newTestRuntime()
	httpClient.SetResponse("GET", "https://api.test/data", internalhttp.Response{
		StatusCode: 200,
		Headers:    internalhttp.Headers{},
		Body:       "payload",
	})
	code := `
def handler(ctx, event):
    resp, err = http.get("https://api.test/data")
    if err != None:
        return {"statusCode": 502, "body": err}
    return {"statusCode": resp["statusCode"], "body": resp["body"]}
`
	resp, err := run(t, rt, code, events.HTTPEvent{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.StatusCode != 200 || resp.Body != "payload" {
		t.Errorf("resp = %d %q", resp.StatusCode, resp.Body)
	}
}

func TestExecute_NoHandler(t *testing.T) {
	rt, _, _ := newTestRuntime()
	code := `x = 1`
	_, err := run(t, rt, code, events.HTTPEvent{Method: "GET", Path: "/"})
	if err == nil {
		t.Fatal("expected error for missing handler")
	}
	if !strings.Contains(err.Error(), "handler") {
		t.Errorf("error = %v", err)
	}
}

func TestExecute_SyntaxError(t *testing.T) {
	rt, _, _ := newTestRuntime()
	code := "def handler(ctx, event)\n    return {}\n" // missing colon
	_, err := run(t, rt, code, events.HTTPEvent{Method: "GET", Path: "/"})
	if err == nil {
		t.Fatal("expected syntax error")
	}
	// The enhanced error reports the offending line and a syntax tip.
	if !strings.Contains(err.Error(), "Error at line") {
		t.Errorf("error did not reference a line: %v", err)
	}
	if !strings.Contains(err.Error(), "Python dialect") {
		t.Errorf("expected a syntax tip: %v", err)
	}
}

func TestExecute_RuntimeErrorEnhanced(t *testing.T) {
	rt, _, _ := newTestRuntime()
	// Reference an undefined name inside the handler.
	code := `
def handler(ctx, event):
    return {"body": undefined_name}
`
	_, err := run(t, rt, code, events.HTTPEvent{Method: "GET", Path: "/"})
	if err == nil {
		t.Fatal("expected runtime error")
	}
	if !strings.Contains(err.Error(), "[TIP]") {
		t.Errorf("expected a tip in the enhanced error: %v", err)
	}
}

func TestExecute_Timeout(t *testing.T) {
	rt, _, _ := newTestRuntime()
	rt.timeout = 50 * time.Millisecond
	// An unbounded loop must be cancelled by the deadline watcher.
	code := `
def handler(ctx, event):
    n = 0
    for _ in range(1000000000):
        n = n + 1
    return {"body": str(n)}
`
	_, err := run(t, rt, code, events.HTTPEvent{Method: "GET", Path: "/"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestExecute_RouterParams(t *testing.T) {
	rt, _, _ := newTestRuntime()
	code := `
def handler(ctx, event):
    params = router.params("/users/42", "/users/:id")
    return {"statusCode": 200, "body": params["id"]}
`
	resp, err := run(t, rt, code, events.HTTPEvent{Method: "GET", Path: "/users/42"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if resp.Body != "42" {
		t.Errorf("body = %q, want 42", resp.Body)
	}
}

// compile-time assurance the runtime satisfies the engine contract.
var _ engine.Runtime = (*StarlarkRuntime)(nil)

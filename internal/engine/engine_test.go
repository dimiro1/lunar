package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/dimiro1/lunar/internal/events"
	"github.com/dimiro1/lunar/internal/services/logger"
	"github.com/dimiro1/lunar/internal/store"
)

// mockRuntime implements Runtime for testing
type mockRuntime struct {
	result *RuntimeResult
	err    error
}

func (m *mockRuntime) Execute(ctx context.Context, req RuntimeRequest) (*RuntimeResult, error) {
	return m.result, m.err
}

// labelRuntime records that it ran, so tests can assert which language was selected.
type labelRuntime struct {
	label string
	ran   *string
}

func (r *labelRuntime) Execute(ctx context.Context, req RuntimeRequest) (*RuntimeResult, error) {
	*r.ran = r.label
	return &RuntimeResult{Response: &events.HTTPResponse{StatusCode: 200}}, nil
}

func TestEngine_Execute_SelectsRuntimeByLanguage(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	fn, _ := db.CreateFunction(ctx, store.Function{ID: "fn", Name: "fn"})
	// Active version is Starlark.
	if _, err := db.CreateVersion(ctx, fn.ID, "def handler(ctx, event): return {}", string(store.LanguageStarlark), nil); err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}

	var ran string
	eng := New(Config{
		DB: db,
		Runtimes: []RuntimeEntry{
			{Language: LanguageLua, Runtime: &labelRuntime{label: "lua", ran: &ran}},
			{Language: LanguageStarlark, Runtime: &labelRuntime{label: "starlark", ran: &ran}},
		},
		Logger:      logger.NewMemoryLogger(),
		IDGenerator: func() string { return "exec-1" },
	})

	if _, err := eng.Execute(ctx, ExecutionRequest{
		FunctionID: fn.ID,
		Event:      events.HTTPEvent{Method: "GET", Path: "/"},
		Trigger:    store.ExecutionTriggerHTTP,
	}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if ran != "starlark" {
		t.Errorf("selected runtime = %q, want starlark", ran)
	}
}

func TestEngine_Execute_UnsupportedLanguage(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	fn, _ := db.CreateFunction(ctx, store.Function{ID: "fn", Name: "fn"})
	if _, err := db.CreateVersion(ctx, fn.ID, "code", string(store.LanguageStarlark), nil); err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}

	// Only Lua is registered; the Starlark version cannot be executed.
	eng := New(Config{
		DB:          db,
		Runtimes:    []RuntimeEntry{{Language: LanguageLua, Runtime: &mockRuntime{}}},
		Logger:      logger.NewMemoryLogger(),
		IDGenerator: func() string { return "exec-1" },
	})

	_, err := eng.Execute(ctx, ExecutionRequest{
		FunctionID: fn.ID,
		Event:      events.HTTPEvent{Method: "GET", Path: "/"},
		Trigger:    store.ExecutionTriggerHTTP,
	})
	var unsupported *UnsupportedLanguageError
	if !errors.As(err, &unsupported) {
		t.Fatalf("error = %v, want UnsupportedLanguageError", err)
	}
}

func TestEngine_Execute_Success(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	// Create a function with a version
	fn, _ := db.CreateFunction(ctx, store.Function{
		ID:   "test-func",
		Name: "Test Function",
	})
	_, _ = db.CreateVersion(ctx, fn.ID, "return {}", "", nil)

	runtime := &mockRuntime{
		result: &RuntimeResult{
			Response: &events.HTTPResponse{
				StatusCode: 200,
				Body:       `{"status":"ok"}`,
			},
		},
	}

	eng := New(Config{
		DB:          db,
		Runtimes:    []RuntimeEntry{{Language: LanguageLua, Runtime: runtime}},
		Logger:      logger.NewMemoryLogger(),
		IDGenerator: func() string { return "exec-123" },
	})

	result, err := eng.Execute(ctx, ExecutionRequest{
		FunctionID: fn.ID,
		Event: events.HTTPEvent{
			Method: "GET",
			Path:   "/fn/test-func",
		},
		Trigger: store.ExecutionTriggerHTTP,
		BaseURL: "http://localhost",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExecutionID != "exec-123" {
		t.Errorf("ExecutionID = %q, want %q", result.ExecutionID, "exec-123")
	}

	if result.Status != store.ExecutionStatusSuccess {
		t.Errorf("Status = %v, want %v", result.Status, store.ExecutionStatusSuccess)
	}

	if result.Response == nil {
		t.Fatal("Response is nil")
	}

	if result.Response.StatusCode != 200 {
		t.Errorf("Response.StatusCode = %d, want %d", result.Response.StatusCode, 200)
	}
}

func TestEngine_Execute_FunctionNotFound(t *testing.T) {
	db := store.NewMemoryDB()

	eng := New(Config{
		DB:          db,
		Runtimes:    []RuntimeEntry{{Language: LanguageLua, Runtime: &mockRuntime{}}},
		Logger:      logger.NewMemoryLogger(),
		IDGenerator: func() string { return "exec-123" },
	})

	_, err := eng.Execute(context.Background(), ExecutionRequest{
		FunctionID: "nonexistent",
	})

	var fnNotFound *FunctionNotFoundError
	if !errors.As(err, &fnNotFound) {
		t.Errorf("expected FunctionNotFoundError, got %T: %v", err, err)
	}
}

func TestEngine_Execute_FunctionDisabled(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	fn, _ := db.CreateFunction(ctx, store.Function{
		ID:       "test-func",
		Name:     "Test Function",
		Disabled: true,
	})

	eng := New(Config{
		DB:          db,
		Runtimes:    []RuntimeEntry{{Language: LanguageLua, Runtime: &mockRuntime{}}},
		Logger:      logger.NewMemoryLogger(),
		IDGenerator: func() string { return "exec-123" },
	})

	_, err := eng.Execute(ctx, ExecutionRequest{
		FunctionID: fn.ID,
	})

	var fnDisabled *FunctionDisabledError
	if !errors.As(err, &fnDisabled) {
		t.Errorf("expected FunctionDisabledError, got %T: %v", err, err)
	}
}

func TestEngine_Execute_NoActiveVersion(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	// Create function without any version
	fn, _ := db.CreateFunction(ctx, store.Function{
		ID:   "test-func",
		Name: "Test Function",
	})

	eng := New(Config{
		DB:          db,
		Runtimes:    []RuntimeEntry{{Language: LanguageLua, Runtime: &mockRuntime{}}},
		Logger:      logger.NewMemoryLogger(),
		IDGenerator: func() string { return "exec-123" },
	})

	_, err := eng.Execute(ctx, ExecutionRequest{
		FunctionID: fn.ID,
	})

	var noVersion *NoActiveVersionError
	if !errors.As(err, &noVersion) {
		t.Errorf("expected NoActiveVersionError, got %T: %v", err, err)
	}
}

func TestEngine_Execute_RuntimeError(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	fn, _ := db.CreateFunction(ctx, store.Function{
		ID:   "test-func",
		Name: "Test Function",
	})
	_, _ = db.CreateVersion(ctx, fn.ID, "invalid code", "", nil)

	runtime := &mockRuntime{
		err: errors.New("runtime error: syntax error"),
	}

	eng := New(Config{
		DB:          db,
		Runtimes:    []RuntimeEntry{{Language: LanguageLua, Runtime: runtime}},
		Logger:      logger.NewMemoryLogger(),
		IDGenerator: func() string { return "exec-123" },
	})

	result, err := eng.Execute(ctx, ExecutionRequest{
		FunctionID: fn.ID,
		Event: events.HTTPEvent{
			Method: "GET",
			Path:   "/fn/test-func",
		},
		Trigger: store.ExecutionTriggerHTTP,
	})

	// Engine returns result even with runtime error
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Error == nil {
		t.Error("expected result.Error to be set")
	}

	if result.Status != store.ExecutionStatusError {
		t.Errorf("Status = %v, want %v", result.Status, store.ExecutionStatusError)
	}
}

func TestEngine_Execute_ErrorStatusCode(t *testing.T) {
	db := store.NewMemoryDB()
	ctx := context.Background()

	fn, _ := db.CreateFunction(ctx, store.Function{
		ID:   "test-func",
		Name: "Test Function",
	})
	_, _ = db.CreateVersion(ctx, fn.ID, "return {statusCode=500}", "", nil)

	runtime := &mockRuntime{
		result: &RuntimeResult{
			Response: &events.HTTPResponse{
				StatusCode: 500,
				Body:       `{"error":"internal error"}`,
			},
		},
	}

	eng := New(Config{
		DB:          db,
		Runtimes:    []RuntimeEntry{{Language: LanguageLua, Runtime: runtime}},
		Logger:      logger.NewMemoryLogger(),
		IDGenerator: func() string { return "exec-123" },
	})

	result, err := eng.Execute(ctx, ExecutionRequest{
		FunctionID: fn.ID,
		Event: events.HTTPEvent{
			Method: "GET",
			Path:   "/fn/test-func",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Status should be error when response has >= 400 status code
	if result.Status != store.ExecutionStatusError {
		t.Errorf("Status = %v, want %v", result.Status, store.ExecutionStatusError)
	}
}

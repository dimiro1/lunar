package starlarkrt

import (
	"context"
	"fmt"
	"time"

	"github.com/dimiro1/lunar/internal/engine"
	"github.com/dimiro1/lunar/internal/events"
	"github.com/dimiro1/lunar/internal/services/ai"
	"github.com/dimiro1/lunar/internal/services/email"
	"github.com/dimiro1/lunar/internal/services/env"
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	"github.com/dimiro1/lunar/internal/services/kv"
	"github.com/dimiro1/lunar/internal/services/logger"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Compile-time check that StarlarkRuntime implements engine.Runtime.
var _ engine.Runtime = (*StarlarkRuntime)(nil)

// scriptName is the filename reported in Starlark error messages.
const scriptName = "handler.star"

// StarlarkRuntime implements the engine.Runtime interface for Starlark code
// execution. It mirrors runner.LuaRuntime: same collaborators, same handler
// contract, different language.
type StarlarkRuntime struct {
	logger       logger.Logger
	kv           kv.Store
	env          env.Store
	http         internalhttp.Client
	ai           ai.Client
	aiTracker    ai.Tracker
	email        email.Client
	emailTracker email.Tracker
	timeout      time.Duration
}

// Config holds the configuration for creating a StarlarkRuntime.
type Config struct {
	Logger       logger.Logger
	KV           kv.Store
	Env          env.Store
	HTTP         internalhttp.Client
	AI           ai.Client
	AITracker    ai.Tracker
	Email        email.Client
	EmailTracker email.Tracker
	Timeout      time.Duration
}

// New creates a new StarlarkRuntime with the given configuration.
func New(cfg Config) *StarlarkRuntime {
	return &StarlarkRuntime{
		logger:       cfg.Logger,
		kv:           cfg.KV,
		env:          cfg.Env,
		http:         cfg.HTTP,
		ai:           cfg.AI,
		aiTracker:    cfg.AITracker,
		email:        cfg.Email,
		emailTracker: cfg.EmailTracker,
		timeout:      cfg.Timeout,
	}
}

// Execute implements the engine.Runtime interface.
func (r *StarlarkRuntime) Execute(ctx context.Context, req engine.RuntimeRequest) (*engine.RuntimeResult, error) {
	timeout := r.timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	thread := &starlark.Thread{
		Name: req.Context.ExecutionID,
		Print: func(_ *starlark.Thread, msg string) {
			r.logger.Info(req.Context.ExecutionID, msg)
		},
	}

	// Starlark has no context hook (unlike gopher-lua's SetContext), so cancel
	// the thread from a watcher when the deadline fires or the request is
	// canceled. The watcher exits as soon as Execute returns.
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			thread.Cancel(ctx.Err().Error())
		case <-stop:
		}
	}()

	predeclared := r.buildPredeclared(ctx, req.Context)

	globals, err := starlark.ExecFileOptions(&syntax.FileOptions{}, thread, scriptName, req.Code, predeclared)
	if err != nil {
		return nil, EnhanceError(fmt.Errorf("failed to load Starlark code: %w", err), req.Code)
	}

	handler, ok := globals["handler"].(starlark.Callable)
	if !ok {
		return nil, EnhanceError(fmt.Errorf("handler function not found in Starlark code"), req.Code)
	}

	switch req.Event.Type() {
	case events.EventTypeHTTP:
		return r.runHTTPEvent(thread, handler, req.Context, req.Event.(events.HTTPEvent), req.Code)
	default:
		return nil, fmt.Errorf("unsupported event type: %s", req.Event.Type())
	}
}

// runHTTPEvent calls handler(ctx, event) for an HTTP event and converts its
// return value to an HTTPResponse.
func (r *StarlarkRuntime) runHTTPEvent(
	thread *starlark.Thread,
	handler starlark.Callable,
	execCtx *events.ExecutionContext,
	event events.HTTPEvent,
	sourceCode string,
) (*engine.RuntimeResult, error) {
	ctxVal := contextToStarlark(execCtx)
	eventVal := httpEventToStarlark(event)

	ret, err := starlark.Call(thread, handler, starlark.Tuple{ctxVal, eventVal}, nil)
	if err != nil {
		return nil, EnhanceError(fmt.Errorf("failed to execute handler: %w", err), sourceCode)
	}

	if ret == starlark.None {
		return nil, EnhanceError(fmt.Errorf("handler did not return a response"), sourceCode)
	}

	httpResp := starlarkToHTTPResponse(ret)
	return &engine.RuntimeResult{Response: &httpResp}, nil
}

// buildPredeclared assembles the global environment exposed to handler code.
// Every module mirrors its gopher-lua counterpart in internal/runner.
func (r *StarlarkRuntime) buildPredeclared(ctx context.Context, ec *events.ExecutionContext) starlark.StringDict {
	return starlark.StringDict{
		"log":     logModule(r.logger, ec.ExecutionID),
		"kv":      kvModule(r.kv, ec.FunctionID),
		"env":     envModule(r.env, ec.FunctionID),
		"http":    httpModule(r.http),
		"json":    jsonModule(),
		"base64":  base64Module(),
		"crypto":  cryptoModule(),
		"time":    timeModule(ctx),
		"url":     urlModule(),
		"strings": stringsModule(),
		"random":  randomModule(),
		"router":  routerModule(ec),
		"ai":      aiModule(r.ai, ec.FunctionID, r.aiTracker, ec.ExecutionID),
		"email":   emailModule(r.email, ec.FunctionID, r.emailTracker, ec.ExecutionID),
	}
}

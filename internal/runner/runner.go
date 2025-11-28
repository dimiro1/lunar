package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/dimiro1/faas-go/internal/ai"
	"github.com/dimiro1/faas-go/internal/env"
	"github.com/dimiro1/faas-go/internal/events"
	internalhttp "github.com/dimiro1/faas-go/internal/http"
	"github.com/dimiro1/faas-go/internal/kv"
	"github.com/dimiro1/faas-go/internal/logger"
	lua "github.com/yuin/gopher-lua"
)

// Response represents the response from executing a function
// The actual response data depends on the event type
type Response struct {
	Type events.EventType
	HTTP *events.HTTPResponse
}

// Dependencies holds all the dependencies needed to run a Lua function
type Dependencies struct {
	Logger    logger.Logger
	KV        kv.Store
	Env       env.Store
	HTTP      internalhttp.Client
	AI        ai.Client
	AITracker ai.Tracker
	Timeout   time.Duration // Execution timeout (defaults to 5 minutes if not set)
}

// Request represents a function execution request
type Request struct {
	Context *events.ExecutionContext
	Event   events.Event
	Code    string
}

// Run executes a Lua function with the given event
func Run(ctx context.Context, deps Dependencies, req Request) (Response, error) {
	// Use provided timeout or default to 5 minutes
	timeout := deps.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	L := lua.NewState()
	defer L.Close()

	// Set the context to enable timeout
	L.SetContext(ctx)

	// Register global modules
	registerLogger(L, deps.Logger, req.Context.ExecutionID)
	registerKV(L, deps.KV, req.Context.FunctionID)
	registerEnv(L, deps.Env, req.Context.FunctionID)
	registerHTTP(L, deps.HTTP)

	// Register utility modules
	registerJSON(L)
	registerBase64(L)
	registerCrypto(L)
	registerTime(L)
	registerURL(L)
	registerStrings(L)
	registerRandom(L)

	// Register AI module
	registerAI(L, deps.AI, req.Context.FunctionID, deps.AITracker, req.Context.ExecutionID)

	// Register Email module
	registerEmail(L, deps.Env, req.Context.FunctionID)

	// Load and execute the Lua code
	if err := L.DoString(req.Code); err != nil {
		enhancedErr := EnhanceError(fmt.Errorf("failed to load Lua code: %w", err), req.Code)
		return Response{}, enhancedErr
	}

	// Get the handler function
	handlerFn := L.GetGlobal("handler")
	if handlerFn.Type() != lua.LTFunction {
		enhancedErr := EnhanceError(fmt.Errorf("handler function not found in Lua code"), req.Code)
		return Response{}, enhancedErr
	}

	// Handle different event types
	switch req.Event.Type() {
	case events.EventTypeHTTP:
		return runHTTPEvent(L, req.Context, req.Event.(events.HTTPEvent), req.Code)
	default:
		return Response{}, fmt.Errorf("unsupported event type: %s", req.Event.Type())
	}
}

// runHTTPEvent executes the handler for an HTTP event
func runHTTPEvent(L *lua.LState, execCtx *events.ExecutionContext, event events.HTTPEvent, sourceCode string) (Response, error) {
	// Create context and event Lua tables
	ctxTable := contextToLuaTable(L, execCtx)
	eventTable := httpEventToLuaTable(L, event)

	// Call handler(ctx, event)
	handlerFn := L.GetGlobal("handler")
	if err := L.CallByParam(lua.P{
		Fn:      handlerFn,
		NRet:    1,
		Protect: true,
	}, ctxTable, eventTable); err != nil {
		enhancedErr := EnhanceError(fmt.Errorf("failed to execute handler: %w", err), sourceCode)
		return Response{}, enhancedErr
	}

	// Get the response from the stack
	ret := L.Get(-1)
	L.Pop(1)

	// Convert response table to HTTPResponse
	if tbl, ok := ret.(*lua.LTable); ok {
		httpResp := luaTableToHTTPResponse(L, tbl)
		return Response{
			Type: events.EventTypeHTTP,
			HTTP: &httpResp,
		}, nil
	}

	enhancedErr := EnhanceError(fmt.Errorf("handler did not return a table"), sourceCode)
	return Response{}, enhancedErr
}

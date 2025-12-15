// Package engine provides the execution orchestration layer for function execution.
//
// The engine package separates concerns between:
//   - HTTP handling (in the api package)
//   - Execution orchestration (this package)
//   - Language-specific execution (runtime implementations)
//
// # Architecture
//
// The engine package defines two main interfaces:
//
// Runtime: Implemented by language-specific executors (Lua, JavaScript, etc.)
// that handle the actual code execution.
//
// Engine: Orchestrates the complete execution lifecycle including:
//   - Function and version retrieval
//   - Execution record management
//   - Event masking for storage
//   - Runtime invocation
//   - Status tracking and duration measurement
//
// # Usage
//
// Create an engine with all required dependencies:
//
//	eng := engine.New(engine.Config{
//	    DB:           db,
//	    Runtime:      luaRuntime,
//	    Logger:       logger,
//	    // ... other dependencies
//	})
//
// Execute a function:
//
//	result, err := eng.Execute(ctx, engine.ExecutionRequest{
//	    FunctionID: "my-function",
//	    Event:      httpEvent,
//	    Trigger:    store.ExecutionTriggerHTTP,
//	})
package engine

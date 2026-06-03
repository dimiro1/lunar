// Package starlarkrt implements the engine.Runtime interface for Starlark, a
// deterministic, sandboxed Python dialect (google/starlark-go). It is the
// sibling of the gopher-lua runtime in internal/runner: both expose the same
// host capabilities (log, kv, env, http, json, ...) backed by the shared,
// language-agnostic services in internal/services and internal/runtime, and
// both call a user-defined handler(ctx, event) entry point.
//
// A Starlark function looks like:
//
//	def handler(ctx, event):
//	    log.info("hit: " + event.method)
//	    return {"statusCode": 200, "body": "Hello, World!"}
//
// ctx and event are passed as structs (attribute access: ctx.executionId,
// event.method); the handler returns a dict describing the HTTP response.
// Fallible host calls follow the Lua two-value convention via tuple unpacking:
//
//	resp, err = http.get("https://example.com")
package starlarkrt

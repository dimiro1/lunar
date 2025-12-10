package runner

import (
	"github.com/dimiro1/lunar/internal/events"
	lua "github.com/yuin/gopher-lua"
)

// httpEventToLuaTable converts an HTTPEvent to a Lua table
func httpEventToLuaTable(L *lua.LState, event events.HTTPEvent) *lua.LTable {
	tbl := L.NewTable()

	L.SetField(tbl, "method", lua.LString(event.Method))
	L.SetField(tbl, "path", lua.LString(event.Path))
	L.SetField(tbl, "relativePath", lua.LString(event.RelativePath))
	L.SetField(tbl, "body", lua.LString(event.Body))

	// Convert headers to Lua table
	headersTbl := L.NewTable()
	for k, v := range event.Headers {
		L.SetField(headersTbl, k, lua.LString(v))
	}
	L.SetField(tbl, "headers", headersTbl)

	// Convert query params to Lua table
	queryTbl := L.NewTable()
	for k, v := range event.Query {
		L.SetField(queryTbl, k, lua.LString(v))
	}
	L.SetField(tbl, "query", queryTbl)

	return tbl
}

// contextToLuaTable converts an ExecutionContext to a Lua table
func contextToLuaTable(L *lua.LState, ctx *events.ExecutionContext) *lua.LTable {
	tbl := L.NewTable()

	L.SetField(tbl, "executionId", lua.LString(ctx.ExecutionID))
	L.SetField(tbl, "functionId", lua.LString(ctx.FunctionID))
	L.SetField(tbl, "startedAt", lua.LNumber(ctx.StartedAt))

	if ctx.RequestID != "" {
		L.SetField(tbl, "requestId", lua.LString(ctx.RequestID))
	}

	if ctx.FunctionName != "" {
		L.SetField(tbl, "functionName", lua.LString(ctx.FunctionName))
	}

	if ctx.Version != "" {
		L.SetField(tbl, "version", lua.LString(ctx.Version))
	}

	if ctx.BaseURL != "" {
		L.SetField(tbl, "baseUrl", lua.LString(ctx.BaseURL))
	}

	return tbl
}

// luaTableToHTTPResponse converts a Lua table to an HTTPResponse
func luaTableToHTTPResponse(_ *lua.LState, tbl *lua.LTable) events.HTTPResponse {
	response := events.HTTPResponse{
		StatusCode: 200, // Default
		Headers:    make(map[string]string),
	}

	// Get statusCode
	if statusCode := tbl.RawGetString("statusCode"); statusCode != lua.LNil {
		response.StatusCode = int(lua.LVAsNumber(statusCode))
	}

	// Get body
	if body := tbl.RawGetString("body"); body != lua.LNil {
		response.Body = lua.LVAsString(body)
	}

	// Get headers
	if headers := tbl.RawGetString("headers"); headers != lua.LNil {
		if headersTbl, ok := headers.(*lua.LTable); ok {
			headersTbl.ForEach(func(k, v lua.LValue) {
				response.Headers[lua.LVAsString(k)] = lua.LVAsString(v)
			})
		}
	}

	// Get isBase64Encoded
	if isBase64 := tbl.RawGetString("isBase64Encoded"); isBase64 != lua.LNil {
		response.IsBase64Encoded = lua.LVAsBool(isBase64)
	}

	return response
}

package runner

import (
	"github.com/dimiro1/lunar/internal/events"
	"github.com/dimiro1/lunar/internal/runtime/router"
	lua "github.com/yuin/gopher-lua"
)

// registerRouter registers the router module with path matching and building functions.
// This is a thin wrapper around the stdlib/router package.
func registerRouter(L *lua.LState, ctx *events.ExecutionContext) {
	routerModule := L.NewTable()

	L.SetField(routerModule, "match", L.NewFunction(routerMatch))
	L.SetField(routerModule, "params", L.NewFunction(routerParams))
	L.SetField(routerModule, "path", L.NewFunction(makeRouterPath(ctx.FunctionID)))
	L.SetField(routerModule, "url", L.NewFunction(makeRouterURL(ctx.FunctionID, ctx.BaseURL)))

	L.SetGlobal("router", routerModule)
}

// makeRouterPath creates a function that builds paths for the current function
func makeRouterPath(functionID string) lua.LGFunction {
	return func(L *lua.LState) int {
		pattern := L.CheckString(1)
		params := extractStringParams(L, 2)
		L.Push(lua.LString(router.FunctionPath(functionID, pattern, params)))
		return 1
	}
}

// makeRouterURL creates a function that builds URLs for the current function
func makeRouterURL(functionID, baseURL string) lua.LGFunction {
	return func(L *lua.LState) int {
		pattern := L.CheckString(1)
		params := extractStringParams(L, 2)
		L.Push(lua.LString(router.FunctionURL(baseURL, functionID, pattern, params)))
		return 1
	}
}

// extractStringParams extracts string parameters from the Lua stack
func extractStringParams(L *lua.LState, argIndex int) map[string]string {
	if L.GetTop() < argIndex || L.Get(argIndex).Type() != lua.LTTable {
		return nil
	}
	return luaTableToStringMap(L, L.CheckTable(argIndex))
}

// luaTableToStringMap converts a Lua table to a map of strings
func luaTableToStringMap(L *lua.LState, tbl *lua.LTable) map[string]string {
	result := make(map[string]string)
	tbl.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			result[string(key)] = lua.LVAsString(v)
		}
	})
	return result
}

// routerMatch checks if a path matches a pattern
// Usage: local matched = router.match(path, pattern)
func routerMatch(L *lua.LState) int {
	path := L.CheckString(1)
	pattern := L.CheckString(2)

	result := router.Match(path, pattern)
	L.Push(lua.LBool(result.Matched))
	return 1
}

// routerParams extracts parameters from a path using a pattern
// Usage: local params = router.params(path, pattern)
func routerParams(L *lua.LState) int {
	path := L.CheckString(1)
	pattern := L.CheckString(2)

	result := router.Match(path, pattern)

	paramsTable := L.NewTable()
	if result.Matched {
		for key, value := range result.Params {
			L.SetField(paramsTable, key, lua.LString(value))
		}
	}

	L.Push(paramsTable)
	return 1
}

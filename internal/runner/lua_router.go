package runner

import (
	"strings"

	"github.com/dimiro1/lunar/internal/events"
	lua "github.com/yuin/gopher-lua"
)

func registerRouter(L *lua.LState, ctx *events.ExecutionContext) {
	routerModule := L.NewTable()

	L.SetField(routerModule, "match", L.NewFunction(routerMatch))
	L.SetField(routerModule, "params", L.NewFunction(routerParams))
	L.SetField(routerModule, "path", L.NewFunction(makeRouterPath(ctx.FunctionID)))
	L.SetField(routerModule, "url", L.NewFunction(makeRouterURL(ctx.FunctionID, ctx.BaseURL)))

	L.SetGlobal("router", routerModule)
}

func makeRouterPath(functionID string) lua.LGFunction {
	return func(L *lua.LState) int {
		pattern := L.CheckString(1)
		var params map[string]string
		if L.GetTop() >= 2 && L.Get(2).Type() == lua.LTTable {
			params = luaTableToStringMap(L, L.CheckTable(2))
		}
		fullPath := "/fn/" + functionID + buildPath(pattern, params)
		L.Push(lua.LString(fullPath))
		return 1
	}
}

func makeRouterURL(functionID, baseURL string) lua.LGFunction {
	return func(L *lua.LState) int {
		pattern := L.CheckString(1)
		var params map[string]string
		if L.GetTop() >= 2 && L.Get(2).Type() == lua.LTTable {
			params = luaTableToStringMap(L, L.CheckTable(2))
		}
		fullURL := strings.TrimSuffix(baseURL, "/") + "/fn/" + functionID + buildPath(pattern, params)
		L.Push(lua.LString(fullURL))
		return 1
	}
}

func buildPath(pattern string, params map[string]string) string {
	if len(params) == 0 {
		return pattern
	}
	result := pattern
	for key, value := range params {
		result = strings.ReplaceAll(result, ":"+key, value)
	}
	return result
}

func luaTableToStringMap(L *lua.LState, tbl *lua.LTable) map[string]string {
	result := make(map[string]string)
	tbl.ForEach(func(k, v lua.LValue) {
		if key, ok := k.(lua.LString); ok {
			result[string(key)] = lua.LVAsString(v)
		}
	})
	return result
}

func routerMatch(L *lua.LState) int {
	path := L.CheckString(1)
	pattern := L.CheckString(2)

	matched, _ := matchPath(path, pattern)
	L.Push(lua.LBool(matched))
	return 1
}

func routerParams(L *lua.LState) int {
	path := L.CheckString(1)
	pattern := L.CheckString(2)

	matched, params := matchPath(path, pattern)

	paramsTable := L.NewTable()
	if matched {
		for key, value := range params {
			L.SetField(paramsTable, key, lua.LString(value))
		}
	}

	L.Push(paramsTable)
	return 1
}

// matchPath matches a path against a pattern. Syntax: :name for params, * for wildcard.
func matchPath(path, pattern string) (bool, map[string]string) {
	params := make(map[string]string)

	path = strings.TrimSuffix(path, "/")
	pattern = strings.TrimSuffix(pattern, "/")
	if path == "" {
		path = "/"
	}
	if pattern == "" {
		pattern = "/"
	}

	pathSegments := splitPath(path)
	patternSegments := splitPath(pattern)
	hasWildcard := len(patternSegments) > 0 && patternSegments[len(patternSegments)-1] == "*"

	if hasWildcard {
		patternSegments = patternSegments[:len(patternSegments)-1]
		if len(pathSegments) <= len(patternSegments) {
			return false, nil
		}
	} else if len(pathSegments) != len(patternSegments) {
		return false, nil
	}

	for i, patternSeg := range patternSegments {
		pathSeg := pathSegments[i]
		if strings.HasPrefix(patternSeg, ":") {
			params[patternSeg[1:]] = pathSeg
		} else if pathSeg != patternSeg {
			return false, nil
		}
	}

	return true, params
}

func splitPath(path string) []string {
	parts := strings.Split(path, "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}
	return segments
}

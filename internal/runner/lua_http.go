package runner

import (
	internalhttp "github.com/dimiro1/lunar/internal/services/http"
	lua "github.com/yuin/gopher-lua"
)

// registerHTTP creates the global 'http' table with HTTP client functions
func registerHTTP(L *lua.LState, httpClient internalhttp.Client) {
	httpTable := L.NewTable()

	// http.get(url, options)
	L.SetField(httpTable, "get", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		options := L.OptTable(2, L.NewTable())

		req := internalhttp.Request{
			URL:     url,
			Headers: luaTableToHeaders(options.RawGetString("headers")),
			Query:   luaTableToQuery(options.RawGetString("query")),
		}

		resp, err := httpClient.Get(req)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(httpResponseToLuaTable(L, resp))
		L.Push(lua.LNil)
		return 2
	}))

	// http.post(url, options)
	L.SetField(httpTable, "post", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		options := L.OptTable(2, L.NewTable())

		req := internalhttp.Request{
			URL:     url,
			Headers: luaTableToHeaders(options.RawGetString("headers")),
			Query:   luaTableToQuery(options.RawGetString("query")),
			Body:    lua.LVAsString(options.RawGetString("body")),
		}

		resp, err := httpClient.Post(req)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(httpResponseToLuaTable(L, resp))
		L.Push(lua.LNil)
		return 2
	}))

	// http.put(url, options)
	L.SetField(httpTable, "put", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		options := L.OptTable(2, L.NewTable())

		req := internalhttp.Request{
			URL:     url,
			Headers: luaTableToHeaders(options.RawGetString("headers")),
			Query:   luaTableToQuery(options.RawGetString("query")),
			Body:    lua.LVAsString(options.RawGetString("body")),
		}

		resp, err := httpClient.Put(req)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(httpResponseToLuaTable(L, resp))
		L.Push(lua.LNil)
		return 2
	}))

	// http.delete(url, options)
	L.SetField(httpTable, "delete", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		options := L.OptTable(2, L.NewTable())

		req := internalhttp.Request{
			URL:     url,
			Headers: luaTableToHeaders(options.RawGetString("headers")),
			Query:   luaTableToQuery(options.RawGetString("query")),
		}

		resp, err := httpClient.Delete(req)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(httpResponseToLuaTable(L, resp))
		L.Push(lua.LNil)
		return 2
	}))

	L.SetGlobal("http", httpTable)
}

// luaTableToHeaders converts a Lua table to HTTP headers map
func luaTableToHeaders(lv lua.LValue) internalhttp.Headers {
	headers := make(internalhttp.Headers)
	if tbl, ok := lv.(*lua.LTable); ok {
		tbl.ForEach(func(k, v lua.LValue) {
			headers[lua.LVAsString(k)] = lua.LVAsString(v)
		})
	}
	return headers
}

// luaTableToQuery converts a Lua table to HTTP query params map
func luaTableToQuery(lv lua.LValue) internalhttp.Query {
	query := make(internalhttp.Query)
	if tbl, ok := lv.(*lua.LTable); ok {
		tbl.ForEach(func(k, v lua.LValue) {
			query[lua.LVAsString(k)] = lua.LVAsString(v)
		})
	}
	return query
}

// httpResponseToLuaTable converts an HTTP response to a Lua table
func httpResponseToLuaTable(L *lua.LState, resp internalhttp.Response) *lua.LTable {
	tbl := L.NewTable()
	L.SetField(tbl, "statusCode", lua.LNumber(resp.StatusCode))
	L.SetField(tbl, "body", lua.LString(resp.Body))

	// Convert headers to Lua table
	headersTbl := L.NewTable()
	for k, v := range resp.Headers {
		L.SetField(headersTbl, k, lua.LString(v))
	}
	L.SetField(tbl, "headers", headersTbl)

	return tbl
}

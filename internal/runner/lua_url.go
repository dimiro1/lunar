package runner

import (
	stdliburl "github.com/dimiro1/lunar/internal/runtime/url"
	lua "github.com/yuin/gopher-lua"
)

// registerURL registers the url module with URL parsing and encoding functions.
// This is a thin wrapper around the stdlib/url package.
func registerURL(L *lua.LState) {
	urlModule := L.NewTable()

	L.SetField(urlModule, "parse", L.NewFunction(urlParse))
	L.SetField(urlModule, "encode", L.NewFunction(urlEncode))
	L.SetField(urlModule, "decode", L.NewFunction(urlDecode))

	L.SetGlobal("url", urlModule)
}

// urlParse parses a URL string into components
// Usage: local parsed, err = url.parse(urlStr)
// Returns: { scheme, host, path, query, fragment, username, password }
func urlParse(L *lua.LState) int {
	urlStr := L.CheckString(1)

	parsedURL, err := stdliburl.Parse(urlStr)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert ParsedURL to Lua table
	result := L.NewTable()
	L.SetField(result, "scheme", lua.LString(parsedURL.Scheme))
	L.SetField(result, "host", lua.LString(parsedURL.Host))
	L.SetField(result, "path", lua.LString(parsedURL.Path))
	L.SetField(result, "fragment", lua.LString(parsedURL.Fragment))

	// Convert query parameters to Lua table
	queryTable := L.NewTable()
	for key, values := range parsedURL.Query {
		if len(values) == 1 {
			L.SetField(queryTable, key, lua.LString(values[0]))
		} else {
			// Multiple values - create an array
			arrayTable := L.NewTable()
			for i, v := range values {
				arrayTable.RawSetInt(i+1, lua.LString(v))
			}
			L.SetField(queryTable, key, arrayTable)
		}
	}
	L.SetField(result, "query", queryTable)

	// Add username and password if present
	if parsedURL.Username != "" {
		L.SetField(result, "username", lua.LString(parsedURL.Username))
	}
	if parsedURL.Password != "" {
		L.SetField(result, "password", lua.LString(parsedURL.Password))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}

// urlEncode URL-encodes a string
// Usage: local encoded = url.encode(str)
func urlEncode(L *lua.LState) int {
	L.Push(lua.LString(stdliburl.Encode(L.CheckString(1))))
	return 1
}

// urlDecode URL-decodes a string
// Usage: local decoded, err = url.decode(str)
func urlDecode(L *lua.LState) int {
	decoded, err := stdliburl.Decode(L.CheckString(1))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(decoded))
	L.Push(lua.LNil)
	return 2
}

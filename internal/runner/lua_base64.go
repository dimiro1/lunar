package runner

import (
	stdlibbase64 "github.com/dimiro1/lunar/internal/runtime/base64"
	lua "github.com/yuin/gopher-lua"
)

// registerBase64 registers the base64 module with encode/decode functions.
// This is a thin wrapper around the stdlib/base64 package.
func registerBase64(L *lua.LState) {
	base64Module := L.NewTable()

	L.SetField(base64Module, "encode", L.NewFunction(base64Encode))
	L.SetField(base64Module, "decode", L.NewFunction(base64Decode))

	L.SetGlobal("base64", base64Module)
}

// base64Encode encodes a string to base64
// Usage: local encoded = base64.encode(str)
func base64Encode(L *lua.LState) int {
	L.Push(lua.LString(stdlibbase64.Encode(L.CheckString(1))))
	return 1
}

// base64Decode decodes a base64 string
// Usage: local decoded, err = base64.decode(str)
func base64Decode(L *lua.LState) int {
	decoded, err := stdlibbase64.Decode(L.CheckString(1))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(decoded))
	L.Push(lua.LNil)
	return 2
}

package runner

import (
	"encoding/base64"

	lua "github.com/yuin/gopher-lua"
)

// registerBase64 registers the base64 module with encode/decode functions
func registerBase64(L *lua.LState) {
	base64Module := L.NewTable()

	// Register base64.encode function
	L.SetField(base64Module, "encode", L.NewFunction(base64Encode))

	// Register base64.decode function
	L.SetField(base64Module, "decode", L.NewFunction(base64Decode))

	// Set the base64 module as a global
	L.SetGlobal("base64", base64Module)
}

// base64Encode encodes a string to base64
// Usage: local encoded = base64.encode(str)
func base64Encode(L *lua.LState) int {
	str := L.CheckString(1)
	encoded := base64.StdEncoding.EncodeToString([]byte(str))
	L.Push(lua.LString(encoded))
	return 1
}

// base64Decode decodes a base64 string
// Usage: local decoded, err = base64.decode(str)
func base64Decode(L *lua.LState) int {
	str := L.CheckString(1)

	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(decoded))
	L.Push(lua.LNil)
	return 2
}

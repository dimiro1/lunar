package runner

import (
	"github.com/dimiro1/lunar/internal/runtime/random"
	lua "github.com/yuin/gopher-lua"
)

// registerRandom registers the random module with random generation functions.
// This is a thin wrapper around the stdlib/random package.
func registerRandom(L *lua.LState) {
	randomModule := L.NewTable()

	L.SetField(randomModule, "int", L.NewFunction(randomInt))
	L.SetField(randomModule, "float", L.NewFunction(randomFloat))
	L.SetField(randomModule, "string", L.NewFunction(randomString))
	L.SetField(randomModule, "bytes", L.NewFunction(randomBytes))
	L.SetField(randomModule, "hex", L.NewFunction(randomHex))
	L.SetField(randomModule, "id", L.NewFunction(randomID))

	L.SetGlobal("random", randomModule)
}

// randomInt generates a random integer between min and max (inclusive)
// Usage: local num = random.int(min, max)
func randomInt(L *lua.LState) int {
	minValue := L.CheckInt(1)
	maxValue := L.CheckInt(2)

	result, err := random.Int(minValue, maxValue)
	if err != nil {
		L.ArgError(1, err.Error())
		return 0
	}

	L.Push(lua.LNumber(result))
	return 1
}

// randomFloat generates a random float between 0.0 and 1.0
// Usage: local num = random.float()
func randomFloat(L *lua.LState) int {
	L.Push(lua.LNumber(random.Float()))
	return 1
}

// randomString generates a random alphanumeric string of specified length
// Usage: local str = random.string(length)
func randomString(L *lua.LState) int {
	length := L.CheckInt(1)

	result, err := random.String(length)
	if err != nil {
		L.ArgError(1, err.Error())
		return 0
	}

	L.Push(lua.LString(result))
	return 1
}

// randomBytes generates random bytes and returns them as base64-encoded string
// Usage: local bytes = random.bytes(length)
func randomBytes(L *lua.LState) int {
	length := L.CheckInt(1)

	result, err := random.Bytes(length)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(result))
	return 1
}

// randomHex generates random bytes and returns them as hex-encoded string
// Usage: local hexStr = random.hex(length)
func randomHex(L *lua.LState) int {
	length := L.CheckInt(1)

	result, err := random.Hex(length)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(result))
	return 1
}

// randomID generates a globally unique sortable ID using xid
// Usage: local id = random.id()
// Returns a 20-character string (smaller than UUID, sortable by creation time)
func randomID(L *lua.LState) int {
	L.Push(lua.LString(random.ID()))
	return 1
}

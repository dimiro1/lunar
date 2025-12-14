package runner

import (
	"github.com/dimiro1/lunar/internal/services/env"
	lua "github.com/yuin/gopher-lua"
)

// registerEnv creates the global 'env' table with environment variable functions
func registerEnv(L *lua.LState, envStore env.Store, functionID string) {
	envTable := L.NewTable()

	// env.get(key)
	L.SetField(envTable, "get", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value, err := envStore.Get(functionID, key)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(value))
		return 1
	}))

	// env.set(key, value)
	L.SetField(envTable, "set", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)
		err := envStore.Set(functionID, key, value)
		if err != nil {
			L.Push(lua.LFalse)
			return 1
		}
		L.Push(lua.LTrue)
		return 1
	}))

	// env.delete(key)
	L.SetField(envTable, "delete", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		err := envStore.Delete(functionID, key)
		if err != nil {
			L.Push(lua.LFalse)
			return 1
		}
		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("env", envTable)
}

package runner

import (
	"github.com/dimiro1/lunar/internal/services/kv"
	lua "github.com/yuin/gopher-lua"
)

// registerKV creates the global 'kv' table with key-value storage functions
func registerKV(L *lua.LState, kvStore kv.Store, functionID string) {
	kvTable := L.NewTable()

	// kv.get(key)
	L.SetField(kvTable, "get", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value, err := kvStore.Get(functionID, key)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(value))
		return 1
	}))

	// kv.set(key, value)
	L.SetField(kvTable, "set", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)
		err := kvStore.Set(functionID, key, value)
		if err != nil {
			L.Push(lua.LFalse)
			return 1
		}
		L.Push(lua.LTrue)
		return 1
	}))

	// kv.delete(key)
	L.SetField(kvTable, "delete", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		err := kvStore.Delete(functionID, key)
		if err != nil {
			L.Push(lua.LFalse)
			return 1
		}
		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("kv", kvTable)
}

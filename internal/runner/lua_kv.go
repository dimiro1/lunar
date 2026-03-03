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

	// kv.listKeys()
	L.SetField(kvTable, "listKeys", L.NewFunction(func(L *lua.LState) int {
		keys, err := kvStore.ListKeys(functionID)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		luaKeys := L.NewTable()
		index := 0
		for _, key := range keys {
			index++
			L.SetTable(luaKeys, lua.LNumber(index), lua.LString(key))
		}
		L.Push(luaKeys)

		return 1
	}))

	// kv.all()
	L.SetField(kvTable, "all", L.NewFunction(func(L *lua.LState) int {
		all, err := kvStore.All(functionID)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		luaAll := L.NewTable()
		for key, value := range all {
			L.SetTable(luaAll, lua.LString(key), lua.LString(value))
		}
		L.Push(luaAll)
		return 1
	}))

	// kv.getGlobal(key)
	L.SetField(kvTable, "getGlobal", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value, err := kvStore.GetGlobal(key)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(value))
		return 1
	}))

	// kv.setGlobal(key, value)
	L.SetField(kvTable, "setGlobal", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)
		err := kvStore.SetGlobal(key, value)
		if err != nil {
			L.Push(lua.LFalse)
			return 1
		}
		L.Push(lua.LTrue)
		return 1
	}))

	// kv.deleteGlobal(key)
	L.SetField(kvTable, "deleteGlobal", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		err := kvStore.DeleteGlobal(key)
		if err != nil {
			L.Push(lua.LFalse)
			return 1
		}
		L.Push(lua.LTrue)
		return 1
	}))

	// kv.listGlobalKeys()
	L.SetField(kvTable, "listGlobalKeys", L.NewFunction(func(L *lua.LState) int {
		keys, err := kvStore.ListGlobalKeys()
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		luaKeys := L.NewTable()
		index := 0
		for _, key := range keys {
			index++
			L.SetTable(luaKeys, lua.LNumber(index), lua.LString(key))
		}
		L.Push(luaKeys)

		return 1
	}))

	// kv.allGlobal()
	L.SetField(kvTable, "allGlobal", L.NewFunction(func(L *lua.LState) int {
		all, err := kvStore.AllGlobal()
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		luaAll := L.NewTable()
		for key, value := range all {
			L.SetTable(luaAll, lua.LString(key), lua.LString(value))
		}
		L.Push(luaAll)
		return 1
	}))

	L.SetGlobal("kv", kvTable)
}

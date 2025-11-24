package runner

import (
	"encoding/json"

	lua "github.com/yuin/gopher-lua"
)

// registerJSON registers the json module with encode/decode functions
func registerJSON(L *lua.LState) {
	jsonModule := L.NewTable()

	// Register json.encode function
	L.SetField(jsonModule, "encode", L.NewFunction(jsonEncode))

	// Register json.decode function
	L.SetField(jsonModule, "decode", L.NewFunction(jsonDecode))

	// Set the json module as a global
	L.SetGlobal("json", jsonModule)
}

// jsonEncode converts a Lua value to a JSON string
// Usage: local str = json.encode(data)
func jsonEncode(L *lua.LState) int {
	value := L.CheckAny(1)

	// Convert Lua value to Go value
	goValue := luaValueToGo(L, value)

	// Marshal to JSON
	jsonBytes, err := json.Marshal(goValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(jsonBytes))
	return 1
}

// jsonDecode converts a JSON string to a Lua value
// Usage: local data = json.decode(str)
func jsonDecode(L *lua.LState) int {
	jsonStr := L.CheckString(1)

	// Unmarshal JSON to Go value
	var goValue any
	if err := json.Unmarshal([]byte(jsonStr), &goValue); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Convert Go value to Lua value
	luaValue := goValueToLua(L, goValue)
	L.Push(luaValue)
	return 1
}

// luaValueToGo converts a Lua value to a Go value
func luaValueToGo(L *lua.LState, lv lua.LValue) any {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		// Check if this is an array or object
		maxN := 0
		isArray := true
		v.ForEach(func(key, _ lua.LValue) {
			if numKey, ok := key.(lua.LNumber); ok {
				n := int(numKey)
				if n > maxN {
					maxN = n
				}
			} else {
				isArray = false
			}
		})

		if isArray && maxN > 0 {
			// Convert to slice
			arr := make([]any, maxN)
			for i := 1; i <= maxN; i++ {
				arr[i-1] = luaValueToGo(L, L.GetTable(v, lua.LNumber(i)))
			}
			return arr
		}

		// Convert to map
		m := make(map[string]any)
		v.ForEach(func(key, value lua.LValue) {
			if str, ok := key.(lua.LString); ok {
				m[string(str)] = luaValueToGo(L, value)
			}
		})
		return m
	default:
		return nil
	}
}

// goValueToLua converts a Go value to a Lua value
func goValueToLua(L *lua.LState, v any) lua.LValue {
	switch val := v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(val)
	case float64:
		return lua.LNumber(val)
	case string:
		return lua.LString(val)
	case []any:
		tbl := L.NewTable()
		for i, item := range val {
			tbl.RawSetInt(i+1, goValueToLua(L, item))
		}
		return tbl
	case map[string]any:
		tbl := L.NewTable()
		for k, v := range val {
			L.SetField(tbl, k, goValueToLua(L, v))
		}
		return tbl
	default:
		return lua.LNil
	}
}

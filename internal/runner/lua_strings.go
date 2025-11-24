package runner

import (
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// registerStrings registers the strings module with string manipulation functions
func registerStrings(L *lua.LState) {
	stringsModule := L.NewTable()

	// Register string functions
	L.SetField(stringsModule, "trim", L.NewFunction(stringsTrim))
	L.SetField(stringsModule, "trimLeft", L.NewFunction(stringsTrimLeft))
	L.SetField(stringsModule, "trimRight", L.NewFunction(stringsTrimRight))
	L.SetField(stringsModule, "split", L.NewFunction(stringsSplit))
	L.SetField(stringsModule, "join", L.NewFunction(stringsJoin))
	L.SetField(stringsModule, "hasPrefix", L.NewFunction(stringsHasPrefix))
	L.SetField(stringsModule, "hasSuffix", L.NewFunction(stringsHasSuffix))
	L.SetField(stringsModule, "replace", L.NewFunction(stringsReplace))
	L.SetField(stringsModule, "toLower", L.NewFunction(stringsToLower))
	L.SetField(stringsModule, "toUpper", L.NewFunction(stringsToUpper))
	L.SetField(stringsModule, "contains", L.NewFunction(stringsContains))
	L.SetField(stringsModule, "repeat", L.NewFunction(stringsRepeat))

	// Set the strings module as a global
	L.SetGlobal("strings", stringsModule)
}

// stringsTrim removes leading and trailing whitespace
// Usage: local result = strings.trim(str)
func stringsTrim(L *lua.LState) int {
	str := L.CheckString(1)
	result := strings.TrimSpace(str)
	L.Push(lua.LString(result))
	return 1
}

// stringsTrimLeft removes leading whitespace
// Usage: local result = strings.trimLeft(str)
func stringsTrimLeft(L *lua.LState) int {
	str := L.CheckString(1)
	result := strings.TrimLeft(str, " \t\n\r")
	L.Push(lua.LString(result))
	return 1
}

// stringsTrimRight removes trailing whitespace
// Usage: local result = strings.trimRight(str)
func stringsTrimRight(L *lua.LState) int {
	str := L.CheckString(1)
	result := strings.TrimRight(str, " \t\n\r")
	L.Push(lua.LString(result))
	return 1
}

// stringsSplit splits a string by a separator
// Usage: local parts = strings.split(str, sep)
func stringsSplit(L *lua.LState) int {
	str := L.CheckString(1)
	sep := L.CheckString(2)

	parts := strings.Split(str, sep)

	// Create Lua array
	result := L.NewTable()
	for i, part := range parts {
		result.RawSetInt(i+1, lua.LString(part))
	}

	L.Push(result)
	return 1
}

// stringsJoin joins an array of strings with a separator
// Usage: local result = strings.join(array, sep)
func stringsJoin(L *lua.LState) int {
	array := L.CheckTable(1)
	sep := L.CheckString(2)

	// Convert Lua table to string slice
	var parts []string
	array.ForEach(func(_, v lua.LValue) {
		parts = append(parts, lua.LVAsString(v))
	})

	result := strings.Join(parts, sep)
	L.Push(lua.LString(result))
	return 1
}

// stringsHasPrefix checks if string has prefix
// Usage: local result = strings.hasPrefix(str, prefix)
func stringsHasPrefix(L *lua.LState) int {
	str := L.CheckString(1)
	prefix := L.CheckString(2)
	result := strings.HasPrefix(str, prefix)
	L.Push(lua.LBool(result))
	return 1
}

// stringsHasSuffix checks if string has suffix
// Usage: local result = strings.hasSuffix(str, suffix)
func stringsHasSuffix(L *lua.LState) int {
	str := L.CheckString(1)
	suffix := L.CheckString(2)
	result := strings.HasSuffix(str, suffix)
	L.Push(lua.LBool(result))
	return 1
}

// stringsReplace replaces occurrences of old with new
// Usage: local result = strings.replace(str, old, new, n)
// n is optional: -1 means replace all (default), 1 means replace first, etc.
func stringsReplace(L *lua.LState) int {
	str := L.CheckString(1)
	old := L.CheckString(2)
	replacement := L.CheckString(3)
	n := L.OptInt(4, -1) // default to replace all

	result := strings.Replace(str, old, replacement, n)
	L.Push(lua.LString(result))
	return 1
}

// stringsToLower converts string to lowercase
// Usage: local result = strings.toLower(str)
func stringsToLower(L *lua.LState) int {
	str := L.CheckString(1)
	result := strings.ToLower(str)
	L.Push(lua.LString(result))
	return 1
}

// stringsToUpper converts string to uppercase
// Usage: local result = strings.toUpper(str)
func stringsToUpper(L *lua.LState) int {
	str := L.CheckString(1)
	result := strings.ToUpper(str)
	L.Push(lua.LString(result))
	return 1
}

// stringsContains checks if string contains substring
// Usage: local result = strings.contains(str, substr)
func stringsContains(L *lua.LState) int {
	str := L.CheckString(1)
	substr := L.CheckString(2)
	result := strings.Contains(str, substr)
	L.Push(lua.LBool(result))
	return 1
}

// stringsRepeat repeats a string n times
// Usage: local result = strings.repeat(str, n)
func stringsRepeat(L *lua.LState) int {
	str := L.CheckString(1)
	count := L.CheckInt(2)
	result := strings.Repeat(str, count)
	L.Push(lua.LString(result))
	return 1
}

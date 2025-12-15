package runner

import (
	stdlibtime "github.com/dimiro1/lunar/internal/runtime/time"
	lua "github.com/yuin/gopher-lua"
)

// registerTime registers the time module with time-related functions.
// This is a thin wrapper around the stdlib/time package.
func registerTime(L *lua.LState) {
	timeModule := L.NewTable()

	L.SetField(timeModule, "now", L.NewFunction(timeNow))
	L.SetField(timeModule, "format", L.NewFunction(timeFormat))
	L.SetField(timeModule, "parse", L.NewFunction(timeParse))
	L.SetField(timeModule, "sleep", L.NewFunction(timeSleep(L)))

	L.SetGlobal("time", timeModule)
}

// timeNow returns the current Unix timestamp in seconds
// Usage: local timestamp = time.now()
func timeNow(L *lua.LState) int {
	L.Push(lua.LNumber(stdlibtime.Now()))
	return 1
}

// timeFormat formats a Unix timestamp to a string
// Uses Go's time format layout (e.g., "2006-01-02 15:04:05")
// Usage: local formatted = time.format(timestamp, layout)
func timeFormat(L *lua.LState) int {
	timestamp := L.CheckNumber(1)
	layout := L.CheckString(2)

	formatted := stdlibtime.Format(int64(timestamp), layout)
	L.Push(lua.LString(formatted))
	return 1
}

// timeParse parses a time string according to a layout
// Returns Unix timestamp or nil + error
// Usage: local timestamp, err = time.parse(timeStr, layout)
func timeParse(L *lua.LState) int {
	timeStr := L.CheckString(1)
	layout := L.CheckString(2)

	timestamp, err := stdlibtime.Parse(timeStr, layout)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LNumber(timestamp))
	L.Push(lua.LNil)
	return 2
}

// timeSleep returns a function that sleeps for the specified number of milliseconds
// Note: This will block the Lua execution
// Usage: time.sleep(1000)  -- sleep for 1 second
func timeSleep(L *lua.LState) lua.LGFunction {
	return func(L *lua.LState) int {
		milliseconds := L.CheckNumber(1)
		stdlibtime.Sleep(L.Context(), int64(milliseconds))
		return 0
	}
}

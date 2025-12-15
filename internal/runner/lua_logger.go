package runner

import (
	"github.com/dimiro1/lunar/internal/services/logger"
	lua "github.com/yuin/gopher-lua"
)

// registerLogger creates the global 'log' table with logging functions
func registerLogger(L *lua.LState, log logger.Logger, executionID string) {
	logTable := L.NewTable()

	// log.info(message)
	L.SetField(logTable, "info", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		log.Info(executionID, message)
		return 0
	}))

	// log.debug(message)
	L.SetField(logTable, "debug", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		log.Debug(executionID, message)
		return 0
	}))

	// log.warn(message)
	L.SetField(logTable, "warn", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		log.Warn(executionID, message)
		return 0
	}))

	// log.error(message)
	L.SetField(logTable, "error", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		log.Error(executionID, message)
		return 0
	}))

	L.SetGlobal("log", logTable)
}

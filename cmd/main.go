package main

import "github.com/dimiro1/lunar/internal/logging"

// main builds the dependency graph with uber/fx and runs it.
//
// fx.App.Run installs SIGINT/SIGTERM handlers, starts all OnStart hooks in
// dependency order, blocks until a signal (or an internal shutdown) arrives,
// then runs OnStop hooks in reverse order. The concrete wiring lives in app.go.
func main() {
	logging.Setup()
	newApp().Run()
}

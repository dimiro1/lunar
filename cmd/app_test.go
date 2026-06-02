package main

import (
	"testing"

	"go.uber.org/fx"
)

// TestAppGraph validates the complete fx dependency graph without building or
// running the application. fx.ValidateApp performs a dig dry-run: it checks that
// every provider's dependencies are satisfiable and that there are no cycles,
// but it does not invoke any constructor — so no database is opened and no
// filesystem side effects occur.
//
// This guards against wiring regressions: removing a module, adding a provider
// that needs an unprovided type, or introducing a dependency cycle will fail CI
// here instead of at process startup.
func TestAppGraph(t *testing.T) {
	if err := fx.ValidateApp(appOptions()); err != nil {
		t.Fatalf("fx dependency graph is invalid: %v", err)
	}
}

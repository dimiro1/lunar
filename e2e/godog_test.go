// Package e2e contains the end-to-end test suite for Lunar, written as
// executable specifications with Cucumber/Gherkin via godog.
//
// The tests are true end-to-end: every scenario drives the real dashboard in a
// headless Chrome browser (via chromedp) against an in-process Lunar server.
// The feature files read like a product spec — they talk about functions,
// versions, and the dashboard, never about selectors or HTTP. The Go step
// definitions translate that intent into browser interactions. The only thing
// done outside the browser is invoking a deployed function, which is a public
// HTTP endpoint a real client would call directly.
//
// Run the whole suite with:
//
//	mise run test-e2e      # or: go test ./e2e/...
//
// A Chrome/Chromium must be available on the machine.
package e2e

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

// TestMain runs the godog suite as part of `go test`. The feature files live in
// features/ and the step bindings are registered in InitializeScenario.
func TestMain(m *testing.M) {
	// Silence the server's request/migration logging so the godog report (the
	// thing a reader cares about) isn't buried under per-asset HTTP logs.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Default to the whole features/ directory; GODOG_FEATURE narrows the run to
	// a single file or scenario during development.
	paths := []string{"features"}
	if p := os.Getenv("GODOG_FEATURE"); p != "" {
		paths = []string{p}
	}

	opts := godog.Options{
		Format:   "pretty",
		Paths:    paths,
		Output:   colors.Colored(os.Stdout),
		Strict:   true,
		TestingT: nil,
		// Browser scenarios share a single Chrome process and each server is
		// per-scenario, so run scenarios sequentially for determinism.
		Concurrency: 1,
	}

	status := godog.TestSuite{
		Name:                 "lunar-e2e",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              &opts,
	}.Run()

	os.Exit(status)
}

// InitializeTestSuite wires suite-level lifecycle hooks. The shared headless
// Chrome process is torn down after every scenario has run.
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.AfterSuite(func() {
		shutdownAllocator()
	})
}

// InitializeScenario is invoked once per scenario. It creates a fresh world,
// starts a clean server before the scenario, tears everything down after, and
// registers every step binding against that world.
func InitializeScenario(sc *godog.ScenarioContext) {
	w := &world{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		return ctx, w.setup()
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w.teardown()
		return ctx, nil
	})

	w.registerSteps(sc)
}

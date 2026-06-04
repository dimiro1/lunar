# 0011. Layered testing strategy without a JS test runner

- Status: Accepted
- Date: 2026-06-02

## Context

Lunar spans a Go backend, a generated Go CLI (`lunar-cli`), and a buildless
browser frontend. Each layer needs testing, but the project's constraints shape
*how*:

- **No Node.js / no bundler** (see [ADR-0004](0004-no-nodejs-vendored-js.md)),
  so a Node-based frontend test runner (Jest, Vitest) is off the table.
- The frontend is plain ES modules served as-is, so tests should run them the way
  the browser does, with no transpile step.
- Confidence requires more than units: we want to know the real HTTP API works
  and that the dashboard actually drives that API in a browser.

We need a coherent set of test layers, each runnable through a single `mise`
task (see [ADR-0005](0005-mise-for-toolchain-and-tasks.md)) and CI-friendly.

## Decision

We will test at four layers, each with a dedicated task:

1. **Go unit tests** (`mise run test`) — standard `go test` across the server and
   CLI packages, excluding `e2e`. The fast inner loop.
2. **Frontend tests** with **Jasmine** (`mise run test-frontend`) — specs under
   `frontend/test/spec/` (per-component, plus `i18n`, `routes`, `utils`) run in a
   real browser via `SpecRunner.html`. A tiny Go `cmd/testserver` serves the
   `frontend/` directory on `:8888` and opens the runner; no Node runner is
   involved. Jasmine itself is vendored (see ADR-0004).
3. **CLI integration tests** (`mise run test-cli-integration`) — built-tagged
   (`-tags integration`) tests in `lunar-cli/integration` that exercise the
   generated client against a real running server.
4. **End-to-end browser tests** (`mise run test-e2e`) — **Cucumber/Gherkin**
   feature files under `e2e/features/`, run with
   [godog](https://github.com/cucumber/godog). Each scenario boots the real API
   on an `httptest.Server` against a migrated SQLite store and drives the
   dashboard in **headless Chrome via `chromedp`**. The features are written at
   the product level (no selectors, no implementation detail); the Go step
   definitions in `e2e/*_test.go` translate that intent into browser
   interactions. Function invocation — the one thing a real client does outside
   the browser — is exercised over plain HTTP against the public `/fn` endpoint.
   `mise run test-all` runs units + e2e together.

The unifying principle: wherever a runtime is needed, prefer **Go** as the test
harness (testserver, chromedp + godog, httptest) rather than introducing a
second language's tooling. godog keeps the e2e specs in Cucumber while the
harness stays pure Go.

## Consequences

- Every layer runs through one `mise` task, so contributors and CI invoke tests
  the same way, and the suite needs only the `mise`-managed toolchain (Go, Deno,
  a Chrome for e2e) — no Node.
- Frontend specs run the exact ES modules the browser ships, so there's no
  transpile-vs-runtime skew.
- e2e tests assemble the real `fx`-wired API and a real (migrated) SQLite
  database, giving high-fidelity coverage of the API-plus-dashboard path.
- Trade-offs:
  - Jasmine specs are browser-driven, so they're not part of the headless
    `go test ./...` run and are easy to forget in pure-CLI CI; running them
    requires a browser context. This is the cost of staying Node-free.
  - e2e tests depend on a Chrome/Chromium being available and use timing
    (`chromedp.Sleep`) in places, which can be slower and flakier than unit
    tests — hence they live behind their own task and the `e2e` package is
    excluded from the default `test` run.

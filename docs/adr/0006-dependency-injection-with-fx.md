# 0006. Dependency injection with uber/fx

- Status: Accepted
- Date: 2026-06-02

## Context

Lunar is composed of many cooperating subsystems — the Lua runner, the engine,
the API layer, housekeeping, and a set of built-in services (KV, logger, env,
HTTP, AI, email). These have a non-trivial dependency graph and a real lifecycle:
the database must open before services that use it, the HTTP server must start
after handlers are registered, and everything must shut down in the right order.

Wiring this by hand in `cmd` means a large, brittle constructor-call sequence
where ordering is implicit and adding a dependency ripples through the call
chain. We wanted construction and lifecycle to be declared locally by each
subsystem rather than centralised and manually ordered.

## Decision

We will use [`uber/fx`](https://github.com/uber-go/fx) as the dependency
injection and lifecycle framework. Each subsystem exposes an `fx.Module` (e.g.
`internal/api/module.go`, `internal/runner/module.go`, the `internal/services/*`
modules) that provides its own constructors and registers any
`fx.Lifecycle` start/stop hooks. `cmd/app.go` assembles the application by
composing these modules in `fx.New`.

Configuration is provided into the graph as the `config.Config` value (see
[ADR-0002](0002-configuration-via-environment-variables.md)), so modules depend
on settings by type rather than reaching for globals.

## Consequences

- Each subsystem owns its construction and lifecycle locally; `cmd` just lists
  the modules to include.
- Start/stop ordering is derived from the dependency graph instead of maintained
  by hand, which removes a common class of shutdown bugs.
- Adding a dependency is "ask for it in a constructor signature" rather than
  threading it through intermediate calls.
- Trade-off: `fx` introduces runtime (rather than compile-time) wiring, a graph
  that's resolved via reflection, and a learning curve for contributors new to
  it. For an application of this size with real lifecycle needs we judge the
  structure worth the indirection. This decision was implemented in commit
  `fc1c3e8` ("build the dependency graph with uber/fx").

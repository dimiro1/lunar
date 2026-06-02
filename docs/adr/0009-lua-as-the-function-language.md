# 0009. Lua as the function language

- Status: Accepted
- Date: 2026-06-02

## Context

Lunar is a Function-as-a-Service platform: users write functions that the server
executes in response to HTTP triggers. The central design question is what
language those functions are written in and how they run. Options ranged from
executing arbitrary binaries/containers (heavy, hard to sandbox, slow cold
starts) to embedding a scripting language in-process.

For a lightweight, single-binary, self-hosted tool we wanted function execution
to be in-process, fast to start, easy to sandbox (functions get capabilities only
through APIs we expose — HTTP client, KV, env, logging, AI, email), and bounded
by an execution timeout. We also wanted a small, approachable language that
doesn't require users to manage dependencies.

## Decision

We will use **Lua** as the function authoring language, executed in-process via
[`yuin/gopher-lua`](https://github.com/yuin/gopher-lua), a pure-Go Lua VM.

Each invocation runs in its own Lua state with only the capabilities Lunar
injects as built-in APIs, and is bounded by the configurable
`EXECUTION_TIMEOUT` (see
[ADR-0002](0002-configuration-via-environment-variables.md)). The function
subsystems are wired as `fx` modules (see
[ADR-0006](0006-dependency-injection-with-fx.md)).

We commit to backward compatibility for the Lua APIs while the project is in
beta, as stated in the README.

## Consequences

- Functions start instantly (no container/process spin-up) and run in the same
  binary, keeping the platform lightweight.
- A pure-Go VM preserves the CGo-free, single-binary build (see
  [ADR-0007](0007-single-binary-embedded-frontend.md) and
  [ADR-0008](0008-sqlite-as-the-datastore.md)).
- Sandboxing is capability-based: a function can only do what our injected APIs
  allow, which keeps the security model explicit.
- The Lua API surface is a long-lived contract; changes must preserve backward
  compatibility, which constrains how built-ins evolve.
- Trade-off: users write Lua specifically rather than bringing an arbitrary
  language or existing libraries. For Lunar's "simple serverless functions" niche
  this is a feature, not a limitation; broader language support would be a
  separate, larger decision.

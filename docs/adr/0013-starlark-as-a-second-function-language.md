# 0013. Starlark as a second function language

- Status: Accepted
- Date: 2026-06-03

## Context

[ADR-0009](0009-lua-as-the-function-language.md) chose Lua, executed in-process
via `yuin/gopher-lua`, as the function authoring language, and noted that
"broader language support would be a separate, larger decision." This is that
decision.

The execution layer was already designed for more than one language: the engine
depends on an `engine.Runtime` interface, not on Lua, and the host capabilities
(HTTP client, KV, env, logging, AI, email, plus the pure utilities) live in
language-agnostic `internal/services` and `internal/runtime` packages. Only the
thin binding layer in `internal/runner` was Lua-specific.

We wanted to offer a second language that (a) reuses that same host surface, (b)
preserves the in-process, single-binary, capability-sandboxed model, and (c)
gives users an alternative idiom. Lua's permissive sandbox (the base stdlib must
be actively restricted) and its unfamiliarity to some users were the main
motivations to look further.

## Decision

We will add **Starlark** as a second function language, executed in-process via
[`google/starlark-go`](https://github.com/google/starlark-go), alongside Lua.

- A new `internal/starlarkrt` package implements `engine.Runtime`, mirroring
  `internal/runner`: the same collaborators, the same `handler(ctx, event)`
  contract, and the same module set (`log`, `kv`, `env`, `http`, `json`,
  `base64`, `crypto`, `time`, `url`, `strings`, `random`, `router`, `ai`,
  `email`). `ctx` and `event` are passed as structs (attribute access); the
  handler returns a dict describing the HTTP response. Fallible host calls keep
  Lua's two-value convention through tuple unpacking (`resp, err = http.get(...)`).
- Language is **chosen once, at function creation**, and is sticky thereafter.
  It is stored per version (`function_versions.language` column, migration
  `000011`, default `'lua'` so existing rows are unchanged); when a later version
  is created (an edit/deploy) without an explicit language, the store carries the
  function's most recent version's language forward. It is set via
  `CreateFunctionInput.language` (GraphQL `Language` enum) and the
  `lunar functions create --language` flag; `updateFunction` does not accept a
  language.
- The engine selects the runtime by the executing version's language. Runtimes
  are contributed to an fx value group (`group:"runtimes"`), each tagged with its
  language, so adding a third language touches neither the engine nor existing
  runtimes. An empty language defaults to Lua.

## Consequences

- Both languages share one host surface and timeout/sandbox model; a new
  capability added to `internal/services` is exposed to both with two small
  bindings.
- Starlark tightens the sandbox: it is deterministic and side-effect-free by
  default (no filesystem, network, or clocks except through our APIs) and
  supports bounded execution, so the capability model is enforced by the language
  rather than by stripping a stdlib.
- `starlark-go` is pure Go, preserving the CGo-free single binary (see
  [ADR-0007](0007-single-binary-embedded-frontend.md)).
- The language is fixed at creation and inherited by every later version, so a
  function's behavior is predictable: editing code never silently changes the
  runtime. Switching an existing function's language is intentionally not
  supported — create a new function instead.
- Cost: a second binding layer and a second API doc to keep in sync, and a second
  long-lived API contract. Starlark is intentionally not full Python (no `while`,
  no recursion by default, no classes, limited stdlib); the authoring guide is
  explicit about this so users do not expect Python semantics.
- We commit to the same beta backward-compatibility stance for the Starlark APIs
  as for Lua.

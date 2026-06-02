# 0002. Configuration via environment variables

- Status: Accepted
- Date: 2026-06-02

## Context

Lunar ships as a single self-hosted binary that people run on their own
machines, in Docker, and on platforms like Railway. These environments differ in
how they inject settings, but every one of them can set environment variables.
We need a configuration mechanism that:

- works identically across local, Docker, and PaaS deployments;
- requires no config file to exist for a first run (good defaults);
- keeps secrets such as the API key out of the source tree;
- is straightforward to test without mutating global process state.

The realistic alternatives were a config file format (YAML/TOML), command-line
flags, or environment variables. Config files add a parsing layer and a "where
does it live" question for a single-binary tool. Flags are awkward to thread
through container platforms and don't compose well with secret managers.

## Decision

We will load all runtime configuration from the process environment, following
the [12-factor](https://12factor.net/config) approach, and bind it to a typed
`config.Config` struct using [`caarlos0/env`](https://github.com/caarlos0/env)
struct tags (`env:"..."`, `envDefault:"..."`).

Configuration lives in its own `internal/config` package (not in `cmd`) so that
the per-feature `fx` modules can depend on it directly — see
[ADR-0006](0006-dependency-injection-with-fx.md).

Concerns that a struct tag can't express are handled in code right after parsing:
a custom parser for `EXECUTION_TIMEOUT` (an integer count of seconds), a
computed default for `BASE_URL` (`http://localhost:<PORT>`), creation of the data
directory, and an API-key fallback chain of env var → on-disk file → freshly
generated key.

`config.parse` accepts an explicit environment map so loading can be unit-tested
without touching `os.Environ`.

## Consequences

- The same binary configures itself the same way everywhere; deployment docs are
  just a list of variables with defaults.
- A fresh run works with zero configuration — sensible defaults plus a
  self-generated, persisted API key.
- Secrets are supplied at runtime and never committed.
- Standardising on `caarlos0/env` keeps loading declarative; the few exceptions
  are localised and documented in the package doc comment.
- Trade-off: deeply nested or list-of-object configuration is clumsy as flat
  environment variables. This is acceptable given Lunar's small, flat config
  surface; if that changes we will revisit with a new ADR rather than bolt on a
  file format ad hoc.

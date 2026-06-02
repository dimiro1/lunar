# 0005. mise for toolchain and task management

- Status: Accepted
- Date: 2026-06-02

## Context

The project depends on several pinned tools — Go, golangci-lint, Deno (see
[ADR-0004](0004-no-nodejs-vendored-js.md)), GoReleaser, and Air — and on a
collection of repeatable commands: build server/CLI, run tests at several levels,
lint, format, vendor JS, run dev mode, tag releases. Previously this lived in a
`Makefile` plus prose instructions for installing tools, which drifts: a
contributor's locally-installed Go or linter rarely matches CI, and "how do I run
X" lives outside version control.

We want one declarative file that both **pins the toolchain** (so every machine
and CI run uses identical versions) and **defines the tasks**, replacing the
"install these tools yourself + Makefile" split.

## Decision

We will use [`mise`](https://mise.jdx.dev/) as the single entry point for both
the toolchain and project tasks, configured in `mise.toml`.

- `[tools]` pins exact versions of Go, golangci-lint, Deno, GoReleaser, and Air;
  `mise install` provisions them and `mise up` bumps them.
- `[env]` holds shared variables, including the pinned frontend dependency
  versions consumed by the `vendor-js` task.
- `[tasks.*]` defines every project command (`build`, `test`, `test-e2e`,
  `lint`, `fmt-frontend`, `run`, `dev`, `vendor-js`, `tag`, …) with `depends`,
  `sources`/`outputs` for incremental runs, and `usage` for arguments.

CI and contributors invoke work through `mise run <task>` rather than ad-hoc
commands or a Makefile.

## Consequences

- Local and CI environments use byte-identical tool versions; "works on my
  machine" toolchain drift largely disappears.
- Onboarding is `mise install` followed by `mise run <task>`; the task list is
  self-documenting via each task's `description`.
- Tasks and tool versions live in one version-controlled file, reviewed like any
  other change.
- `sources`/`outputs` give make-style incremental builds without a Makefile.
- Trade-off: contributors must install `mise` first, and we depend on a
  relatively young tool. We judge the reproducibility win worth it, and tasks
  remain plain shell that could be lifted out if needed.

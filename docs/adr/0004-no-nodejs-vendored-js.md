# 0004. No Node.js: vendored JS, Deno for tooling

- Status: Accepted
- Date: 2026-06-02

## Context

Lunar is a Go project that happens to have a rich browser frontend. The default
path for that frontend would drag in the Node.js world: a `package.json`, an
`node_modules/` tree of transitive dependencies, a lockfile, and a bundler. For a
small self-hosted tool that wants to stay auditable and ship as one Go binary,
that ecosystem brings real costs — supply-chain surface area, churn, a second
language runtime in every contributor's environment and in CI, and a build step
between the source and what runs.

At the same time we still want *some* JS tooling: a formatter for the frontend
code, and a way to run the browser test suite.

## Decision

We will not use Node.js, npm, or a bundler anywhere in the project.

Browser dependencies (Mithril, Monaco, highlight.js, Jasmine) are **vendored**
into `frontend/vendor/` by the `vendor-js` `mise` task, which downloads pinned
versions straight from CDNs/registries (`unpkg`, `cdnjs`, the npm registry
tarball for Monaco) using `curl`/`tar`. Versions are pinned as variables in
`mise.toml`, so vendoring is reproducible and upgrades are a deliberate diff.

For tooling that genuinely needs a JS runtime we use [Deno](https://deno.com/)
(itself installed via `mise`): `deno fmt` formats the frontend, excluding the
vendored tree. The frontend test suite runs through a small Go `testserver`
binary that serves Jasmine in a browser — no Node test runner involved.

## Consequences

- There is no `node_modules/`, no `package.json`, and no lockfile to manage or
  audit; the vendored files in git *are* the dependency manifest.
- Every dependency is a concrete, reviewable file at a pinned version. Upgrades
  show up as explicit diffs from re-running `vendor-js`.
- Contributors need only the tools `mise` installs (Go, Deno, etc.); there's no
  separate Node version to match.
- The vendored files are committed, which adds some weight to the repository. We
  accept this in exchange for reproducibility and a buildless frontend.
- Trade-off: we forgo npm's convenience and the framework ecosystems that assume
  a bundler. This is the deliberate counterpart to
  [ADR-0003](0003-frontend-with-mithril.md) and
  [ADR-0007](0007-single-binary-embedded-frontend.md).

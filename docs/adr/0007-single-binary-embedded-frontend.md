# 0007. Single binary with embedded frontend

- Status: Accepted
- Date: 2026-06-02

## Context

Lunar's value proposition includes being lightweight and self-hosted: "single
binary, no external dependencies". A FaaS dashboard, though, is made of static
assets — HTML, CSS, JS, the Monaco editor, vendored libraries. If those assets
ship separately from the binary, operators must deploy and path-configure a web
root, versions can skew between binary and assets, and "just run the binary"
stops being true.

We need the compiled server and its frontend to be one indivisible artifact.

## Decision

We will embed the entire frontend into the Go binary using `go:embed`. The
`frontend` package embeds `css`, `js`, `vendor`, `index.html`, and `llms.txt`
into an `embed.FS` and exposes a `Handler()` that serves them via
`http.FileServer`.

This is the constraint that drives the buildless, vendored frontend: because the
embedded files must be the literal files we ship, there can be no bundler output
step (see [ADR-0003](0003-frontend-with-mithril.md) and
[ADR-0004](0004-no-nodejs-vendored-js.md)). Combined with pure-Go SQLite (see
[ADR-0008](0008-sqlite-as-the-datastore.md)), the result is a CGo-free,
dependency-free single binary, released via GoReleaser.

## Consequences

- Deployment is copying one binary; there is no asset directory to manage and no
  binary/frontend version skew.
- The frontend is served straight from memory with no filesystem layout
  assumptions.
- Builds are reproducible and the release artifact is self-contained.
- Trade-off: changing a frontend file requires recompiling to embed it for a
  production build (dev mode via Air rebuilds automatically, so the day-to-day
  loop is unaffected). The binary also carries the weight of all assets,
  including Monaco — acceptable for a self-hosted dashboard.

# 0003. Frontend with Mithril.js

- Status: Accepted
- Date: 2026-06-02

## Context

Lunar's dashboard is a genuine single-page application: a code editor (Monaco),
routing across functions/versions/executions/logs, i18n, a command palette, and
many reusable components. We need a client-side framework, but the project has
two strong constraints that rule out the mainstream React/Vue/Svelte path:

- **No Node.js toolchain** (see [ADR-0004](0004-no-nodejs-vendored-js.md)). That
  removes JSX/TSX, bundlers, and the npm-based ecosystem those frameworks assume.
- **Ship inside a single Go binary** (see
  [ADR-0007](0007-single-binary-embedded-frontend.md)). The frontend has to be a
  set of static files we can `go:embed`, with no build artifact pipeline.

We need a framework that is small, works as a single `<script>` with no build
step, has a built-in router and XHR layer, and renders via plain function calls
rather than a compiler-dependent template syntax.

## Decision

We will build the dashboard with [Mithril.js](https://mithril.js.org/), vendored
as a single minified file and loaded via a `<script>` tag in `index.html`.

Views are authored as plain ES modules using Mithril's hyperscript (`m(...)`)
API, organised under `frontend/js` into `components/`, `views/`, `routes.js`,
`api.js`, and an `i18n/` layer. No JSX, no transpilation: the `.js` files we
write are the `.js` files the browser runs.

## Consequences

- The frontend has zero build step. Editing a `.js` file and reloading is the
  whole dev loop; the embedded files are exactly what we authored.
- Mithril is tiny and batteries-included (routing + `m.request` for XHR), so we
  avoid pulling a constellation of micro-dependencies to fill gaps.
- Hyperscript instead of JSX means component code is a little more verbose and
  there's no template-level type checking. We accept this; it's the cost of
  having no compiler.
- We pin the Mithril version in `mise.toml` and vendor it via the `vendor-js`
  task, so upgrades are explicit and reproducible.
- Trade-off: we step outside the dominant React ecosystem, so off-the-shelf
  component libraries don't apply. In return we keep the whole frontend
  inspectable, buildless, and embeddable.

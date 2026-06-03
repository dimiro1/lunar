# 0012. GraphQL for the management API

- Status: Accepted
- Date: 2026-06-03

## Context

Lunar's management API (functions, versions, executions, tokens) was a
hand-written REST surface under `/api/*`, described by a hand-maintained
OpenAPI document at `internal/api/docs/openapi.yaml`. Nothing connected the two:
the spec was a *document* kept in step with the handlers by discipline, not by
the compiler, so the two could — and did — drift. The CLI was generated *from*
that spec (oapi-codegen for the client, a custom `lunar-cli/tools/gen` generator
for the Cobra commands), inheriting any drift the spec carried.

Two concrete problems pushed us off this design:

- **Overfetching.** `GET /api/functions` returned every function's full active
  version — including the entire Lua source — plus its env and KV maps, just to
  render a list of names and statuses.
- **No enforced contract.** A field added to a handler but not the spec (or vice
  versa) compiled and shipped. A latent example surfaced during the migration:
  the KV editor's request shape had drifted from what the REST handler accepted.

## Decision

We will serve the entire management API over **GraphQL** using
[`gqlgen`](https://github.com/99designs/gqlgen), with the schema in
`internal/graph/schema/*.graphqls` as the single source of truth. gqlgen
generates the resolver interfaces, so the Go compiler refuses to build until
every schema field has an implementation — drift becomes a compile error rather
than a review responsibility. `gqlgen.yml` binds GraphQL types directly to the
existing `internal/store` structs, so there are no duplicate DTOs.

GraphQL is mounted at `POST /graphql` (behind the same auth middleware as the
old REST routes) with a public GraphiQL playground at `GET /graphql`, following
the per-subsystem `fx.Module` pattern from
[ADR-0006](0006-dependency-injection-with-fx.md). The frontend (`frontend/js/api.js`)
and the CLI (via `hasura/go-graphql-client`) both consume it; validation logic
shared by the resolvers lives in `internal/validation`.

Two endpoints deliberately **stay REST**, because they do not fit a query
language:

- `/fn/{id}` — public function invocation with arbitrary verbs, paths, request
  bodies, and a verbatim response passthrough.
- `/api/auth/*` — login/logout (HttpOnly cookies) and the CLI device-authorization
  flow, which runs *before* a caller is authenticated.

We removed the REST `/api/*` management handlers, the hand-written
`openapi.yaml`, the Swagger `/docs` UI, and the oapi-codegen + `tools/gen` CLI
generators.

## Consequences

- The schema can no longer drift from the code: adding a field without a
  resolver fails to compile, and removing one leaves dead code the compiler
  flags.
- The overfetch is gone — list views select only the fields they render, and
  env/KV/source are lazy field resolvers fetched only when asked for.
- Clients collapse multi-call detail views into a single query, and the CLI is
  now hand-written Go (no codegen step) validated against the live schema by
  introspection.
- We trade Swagger's familiarity and REST's HTTP-status semantics for GraphQL's
  `200 + errors` model; clients translate the `errors` array, and the auth
  middleware still returns a real HTTP 401 ahead of the resolver.
- Field resolvers can be N+1 if env/KV are ever selected across a list; at
  SQLite scale this is fine, and dataloaders remain an option if it ever isn't.
- Timestamps are still carried as `Int`; introducing a dedicated
  `Timestamp`/`Int64` scalar is deferred follow-up work.

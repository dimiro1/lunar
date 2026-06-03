// Package api provides the HTTP API server for the lunar platform.
//
// The management API (functions, versions, executions, tokens) is served over
// GraphQL at /graphql; its schema in internal/graph is the source of truth. The
// REST surface this package still owns is intentionally small:
//
//   - /graphql - GraphQL management API (POST) + GraphiQL playground (GET)
//   - /api/auth/* - login/logout and the CLI device-authorization flow
//   - /fn/{function_id} - public runtime function execution (passthrough)
//
// The REST endpoints are documented in docs/rest-endpoints.md; the GraphQL
// surface is described by its schema in internal/graph/schema.
package api

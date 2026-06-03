# REST endpoints

Lunar's management API (functions, versions, executions, tokens) is served over
**GraphQL** at `/graphql` — see [ADR-0012](adr/0012-graphql-management-api.md)
and the GraphiQL playground at `GET /graphql` for that surface.

A few endpoints stay REST because they don't fit a query language. They are
intentionally few, and — unlike the GraphQL schema, which the compiler keeps in
sync — they are documented here by hand. This file is that reference.

All request and response bodies are JSON unless noted.

## Cookie authentication (dashboard)

Used by the web dashboard. The cookie carries the admin API key.

### `POST /api/auth/login`

No auth required.

Request:

```json
{ "apiKey": "<admin API key>" }
```

Responses:

- `200` — `{ "success": true }`, plus a `Set-Cookie: auth_token=...` (HttpOnly,
  SameSite=Strict, `Secure` when served over HTTPS, 1-day expiry).
- `400` — `{ "success": false, "error": "Invalid request body" }`
- `401` — `{ "success": false, "error": "Invalid API key" }`

### `POST /api/auth/logout`

No auth required. Clears the `auth_token` cookie.

- `200` — `{ "success": true }`

## Device authorization (CLI login)

An OAuth-style device flow so the CLI can obtain an API token without the user
pasting one. The CLI calls `device-request`, the user approves in the dashboard
(which calls `device-approve`), and the CLI polls `device-token` until a token
is issued. See `lunar-cli login`.

### `POST /api/auth/device-request`

No auth required. Starts a flow.

- `200`:

```json
{
  "device_code": "<opaque>",
  "user_code": "ABCD2345",
  "approval_url": "<base URL>/#!/device-approve/<device_code>",
  "expires_in": 300,
  "interval": 5
}
```

`expires_in` and `interval` are seconds; the device code lives for 5 minutes.

### `GET /api/auth/device-token?code=<device_code>`

No auth required. Polled by the CLI every `interval` seconds.

- `200` — `{ "status": "pending" }` while waiting,
  `{ "status": "denied" }` if rejected, or
  `{ "status": "approved", "token": "<API token>" }` once approved.
- `400` — missing `code`.
- `404` — unknown or expired `code`.

### `GET /api/auth/device-approve?code=<device_code>`

**Auth required** (dashboard cookie/bearer). Returns the pending request so the
SPA can render the approval screen.

- `200`:

```json
{
  "device_code": "<opaque>",
  "user_code": "ABCD2345",
  "status": "pending",
  "expires_at": 1780000000
}
```

- `400` — missing `code`; `404` — unknown or expired `code`.

### `POST /api/auth/device-approve`

**Auth required.** The dashboard approves or denies a pending request. On
approval a new API token named `CLI (<user_code>)` is created and handed to the
polling CLI.

Request:

```json
{ "device_code": "<opaque>", "action": "allow" }
```

`action` is `"allow"` or `"deny"`.

- `200` — `{ "success": true }`
- `400` — missing `device_code` or invalid `action`.
- `404` — unknown or expired `device_code`.

## Function execution

### `<METHOD> /fn/{function_id}` and `/fn/{function_id}/{path...}`

**No auth required** — functions are public. This is a passthrough: any HTTP
method, path suffix, headers, query, and body are delivered to the function as
its event, and the function's HTTP response (status, headers, body) is relayed
back verbatim.

Request headers Lunar interprets:

- `X-Trigger: cron` records the execution's trigger as `cron` (default is
  `http`).

Response headers Lunar adds:

- `X-Function-Id`, `X-Function-Version-Id`, `X-Execution-Id`
- `X-Execution-Duration-Ms`

Status codes:

- `2xx`/whatever the function returns on success (defaults to `200` and
  `Content-Type: application/json` if the function doesn't set them).
- `404` — function not found.
- `403` — function disabled.
- `500` — no active version, or the function errored during execution.

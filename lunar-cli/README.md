# lunar-cli

Command-line client for [Lunar](https://github.com/dimiro1/lunar). This README is for contributors working on the CLI internals. If you just want to install and use the CLI, start with the [root README](../README.md#cli).

The CLI is a thin client over the server's **GraphQL API** (`/graphql`). The GraphQL schema is the single source of truth — the server won't compile until every field has a resolver, and the CLI's queries are validated against the live schema by introspection. There is no code-generation step: commands are hand-written Go, which keeps each file small and self-contained.

## Prerequisites

- Go 1.26 or newer

## Binary Name

The released executable is named `lunar-cli`, and the Cobra help output uses the same command name.

## How it works

Each command builds a GraphQL operation as a string, runs it through a small set
of helpers, and prints the result. Two endpoints stay REST and bypass GraphQL:
the device-auth login flow (`cmd/auth.go`) and direct function invocation
(`cmd/invoke.go`, the pass-through `/fn/*` endpoint).

**GraphQL client (`cmd/graphql.go`)**

A thin wrapper around [`hasura/go-graphql-client`](https://github.com/hasura/go-graphql-client).
`mustGraphQLClient()` targets `<server>/graphql` and injects the bearer token via
a custom `http.RoundTripper`. `execRaw` runs an operation and returns the decoded
top-level `data` object. Higher-level helpers shape that into the output the
renderer (and existing scripts) expect:

| Helper | Use | Output shape |
|--------|-----|--------------|
| `gqlObject` | single resource (errors on `null`, the GraphQL 404) | the object |
| `gqlConnection` | paginated list (`{nodes, pageInfo}`) | `{<listKey>: [...], "pagination": {...}}` |
| `gqlList` | plain list | `{<listKey>: [...]}` |
| `gqlSuccess` | boolean mutation (delete/revoke) | `{"success": <bool>}` |

**Command files (`cmd/*.go`)**

Commands are grouped by domain, one file each. Each file declares a `const`
GraphQL selection (aliased from the schema's camelCase to the snake_case the
output renderer expects, e.g. `created_at: createdAt`) and a Cobra command per
operation:

| File | Commands |
|------|----------|
| `cmd/functions.go` | `functions list/get/create/update/delete/env/kv/next-run` |
| `cmd/versions.go` | `versions list/get/activate/delete/diff` |
| `cmd/executions.go` | `executions list/get/logs/ai-requests/email-requests` |
| `cmd/tokens.go` | `tokens list/revoke` |
| `cmd/auth.go` | `login` / `logout` (REST device flow) |
| `cmd/invoke.go` | `invoke` (REST `/fn/*` pass-through) |
| `cmd/llms.go` | `llms` (fetches `/llms.txt`) |

## Directory structure

```
lunar-cli/
├── main.go                 Entry point
├── go.mod                  Module: github.com/dimiro1/lunar/lunar-cli
│
├── cmd/
│   ├── root.go             Root command, global flags (--server, --token), output helpers
│   ├── graphql.go          GraphQL client + gqlObject/gqlConnection/gqlList/gqlSuccess helpers
│   ├── functions.go        lunar-cli functions ...
│   ├── versions.go         lunar-cli versions ...
│   ├── executions.go       lunar-cli executions ...
│   ├── tokens.go           lunar-cli tokens ...
│   ├── auth.go             lunar-cli login / logout   (REST device flow)
│   ├── invoke.go           lunar-cli invoke           (REST /fn/* pass-through)
│   ├── llms.go             lunar-cli llms
│   └── output.go           Pretty / JSON rendering
│
├── config/
│   └── config.go           Read/write ~/.config/lunar/config.yaml
│
└── skills/                 Bundled AI agent skill definitions
```

## Adding a command

1. Pick (or create) the domain file in `cmd/`, e.g. `cmd/functions.go`.
2. Declare the Cobra command and register it in an `init()`:
   ```go
   func init() {
       functionsCmd.AddCommand(myCmd) // or rootCmd.AddCommand for a top-level command
   }
   ```
3. In the `RunE`, build the GraphQL operation and call the matching helper:
   ```go
   func runMyCommand(cmd *cobra.Command, args []string) error {
       query := `query ($id: ID!) { function(id: $id) {` + functionFields + `} }`
       return gqlObject(cmd.Context(), query, map[string]any{"id": args[0]}, "function")
   }
   ```
   Use field aliases (`snake_case: camelCase`) in the selection so the output matches the rest of the CLI. After adding the field/operation to the server's GraphQL schema, run `go build ./...` — introspection and the schema's resolver requirement keep the CLI and server honest.

## AI Agent Skills

The CLI ships with [Claude Code](https://claude.ai/code) skill definitions that teach your AI coding agent how to use the Lunar CLI and write Lua functions. Once installed, the agent will know the full CLI and Lua API without you having to explain anything.

Install the CLI first. If you are working from source in this directory, build it with `go build -o lunar-cli .` and use `./lunar-cli`.

```bash
lunar-cli skills list             # show available skills
lunar-cli skills show lunar-cli   # CLI command reference
lunar-cli skills show lunar-lua   # Lua function authoring guide
```

To install them, ask your agent:

> "Install the Lunar skills from the `lunar-cli skills` CLI command."

## Building

```bash
go build -o lunar-cli .
```

## Installing locally

```bash
go install github.com/dimiro1/lunar/lunar-cli@latest
```

If you are working from a local source checkout, prefer `go build -o lunar-cli .`.

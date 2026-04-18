# lunar-cli

Command-line client for [Lunar](https://github.com/dimiro1/lunar). This README is for contributors working on the CLI internals and code generation. If you just want to install and use the CLI, start with the [root README](../README.md#cli).

The majority of commands are **auto-generated** from the OpenAPI spec at `../internal/api/docs/openapi.yaml`, so the CLI always stays in sync with the API.

## Prerequisites

- Go 1.26 or newer

## Binary Name

The released executable is named `lunar-cli`, and the Cobra help output uses the same command name.

## How it works

Code generation runs in two layers:

```
internal/api/docs/openapi.yaml   (source of truth)
        │
        ├─▶ oapi-codegen ──────▶ client/client.gen.go
        │                        (typed HTTP client + all schema types)
        │
        └─▶ tools/gen ─────────▶ cmd/functions.gen.go
                                  cmd/versions.gen.go
                                  cmd/executions.gen.go
                                  cmd/tokens.gen.go
                                  (Cobra subcommands)
```

**Layer 1 — HTTP client (`oapi-codegen`)**

`client/generate.go` runs [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) against the spec to produce `client/client.gen.go`. This file contains all schema types (`Function`, `Execution`, etc.) and a `ClientWithResponses` struct with one typed method per API operation (e.g. `ListFunctionsWithResponse`, `CreateFunctionWithResponse`).

**Layer 2 — Cobra commands (`tools/gen`)**

`generate.go` runs the custom generator at `tools/gen/main.go`. It parses the spec and, for each API tag that maps to a CLI group, emits a `*.gen.go` file containing Cobra commands wired to the generated HTTP client.

Tags → files:

| OpenAPI tag | Generated file |
|-------------|----------------|
| Functions | `cmd/functions.gen.go` |
| Versions | `cmd/versions.gen.go` |
| Executions | `cmd/executions.gen.go` |
| API Tokens | `cmd/tokens.gen.go` |

Tags **not** generated (implemented manually):

| OpenAPI tag | Manual file | Reason |
|-------------|-------------|--------|
| Authentication | `cmd/auth.go` | Device auth flow needs interactive browser handling |
| Device Authorization | `cmd/auth.go` | Same as above |
| Runtime | `cmd/invoke.go` | Pass-through HTTP call, not a typed API request |

## Directory structure

```
cli/
├── main.go                 Entry point
├── generate.go             go:generate directive for the Cobra generator
├── go.mod                  Module: github.com/dimiro1/lunar/cli
│
├── cmd/
│   ├── root.go             Root command, global flags (--server, --token), mustClient()
│   ├── auth.go             lunar-cli login / logout  (manual)
│   ├── invoke.go           lunar-cli invoke          (manual)
│   ├── functions.gen.go    lunar-cli functions ...   (generated)
│   ├── versions.gen.go     lunar-cli versions ...    (generated)
│   ├── executions.gen.go   lunar-cli executions ...  (generated)
│   └── tokens.gen.go       lunar-cli tokens ...      (generated)
│
├── client/
│   ├── generate.go         go:generate directive for oapi-codegen
│   ├── oapi-codegen.yaml   oapi-codegen configuration
│   └── client.gen.go       Generated HTTP client (do not edit)
│
├── config/
│   └── config.go           Read/write ~/.config/lunar/config.yaml
│
└── tools/
    ├── tools.go            Build-constraint import keeping oapi-codegen in go.sum
    └── gen/
        └── main.go         Generator: openapi.yaml → cmd/*.gen.go
```

## Regenerating after an API change

Whenever `../internal/api/docs/openapi.yaml` is updated, run from this directory:

```bash
go generate ./...
```

This runs both generators in order:

1. `client/generate.go` → regenerates `client/client.gen.go` via `oapi-codegen`
2. `generate.go` → regenerates `cmd/*.gen.go` via `tools/gen`

Then verify the result compiles:

```bash
go build ./...
```

## How the generator works (`tools/gen/main.go`)

The generator is a standalone Go program invoked via `go run ./tools/gen`. It:

1. Parses `openapi.yaml` into minimal Go structs (paths, operations, schemas, parameters).
2. Groups operations by OpenAPI tag using `tagConfigs`.
3. For each tag group, iterates operations and derives:
   - **Command name** — from the `operationId` (strip the tag noun suffix, e.g. `listFunctions` → `list`). Override map handles special cases (`updateEnvVars` → `env`, `getVersionDiff` → `diff`, etc.).
   - **Path params** → positional `cobra.ExactArgs` arguments.
   - **Query params** → `--flag` flags bound to a `client.XxxParams` struct.
   - **Body fields** → `--flag` flags; required fields get `MarkFlagRequired`; optional fields check `cmd.Flags().Changed()` before setting the pointer field. Enum fields are cast to the appropriate `client.XxxType`.
   - **Map body fields** (e.g. `env_vars`, `kv`) → `--flag KEY=VALUE` repeatable flags parsed with `strings.SplitN`.
   - **Code fields** — any `string` field named `code` also accepts `"-"` to read from stdin.
4. Renders the Go source using `fmt.Fprintf` into a `bytes.Buffer`.
5. Formats the result with `go/format` and writes the file.

To add support for a new tag, add an entry to `tagConfigs` in `tools/gen/main.go`:

```go
var tagConfigs = map[string]tagConfig{
    // ...existing tags...
    "My Tag": {fileBase: "mytag", varName: "mytag", commandUse: "mytag", stripSuffix: []string{"mytag", "mytags"}},
}
```

To override a generated command name for a specific operation, add it to `commandNameOverrides`:

```go
var commandNameOverrides = map[string]string{
    // ...existing overrides...
    "myOperationId": "my-command",
}
```

## Adding a manual command

If an operation is too complex to generate (interactive prompts, streaming, etc.):

1. Create `cmd/mycommand.go` in the `cmd` package.
2. Register the command in `init()`:
   ```go
   func init() {
       rootCmd.AddCommand(myCmd) // top-level
       // or: someGroupCmd.AddCommand(myCmd)
   }
   ```
3. Use `mustClient()` to get an authenticated HTTP client and `printJSON(resp.Body)` to print the response.

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
go install github.com/dimiro1/lunar/cli@latest
```

If you are working from a local source checkout, prefer `go build -o lunar-cli .`.

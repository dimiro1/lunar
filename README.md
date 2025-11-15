# FaaS-Go (temporary name)

A lightweight, self-hosted Function-as-a-Service platform written in Go with Lua scripting.

## Features

* **Simple Lua Functions** - Write serverless functions in Lua
* **HTTP Triggers** - Execute functions via HTTP requests
* **Built-in APIs** - HTTP client, KV store, environment variables, logging, and more
* **Version Control** - Track and manage function versions
* **Execution History** - Monitor function executions and logs
* **Web Dashboard** - Manage functions through a clean web interface
* **Lightweight** - Single binary, no external dependencies

## Quick Start

### Building from Source

```bash
git clone https://github.com/dimiro1/faas-go.git
cd faas-go
make build
```

### Running

```bash
./faas-go
```

The API will be available at `http://localhost:3000` and the dashboard at `http://localhost:3000/`.

## Writing Functions

Functions are written in Lua and must export a `handler` function:

```lua
function handler(ctx, event)
  -- ctx contains execution context (executionId, functionId, etc.)
  -- event contains HTTP request data (method, path, query, body, headers)
  
  log.info("Function started")
  
  return {
    statusCode = 200,
    headers = { ["Content-Type"] = "application/json" },
    body = json.encode({ message = "Hello, World!" })
  }
end
```

### Available APIs

* **log** - Logging utilities (info, debug, warn, error)
* **kv** - Key-value storage (get, set, delete)
* **env** - Environment variables (get)
* **http** - HTTP client (get, post, put, delete)
* **json** - JSON encoding/decoding
* **crypto** - Cryptographic functions (md5, sha256, hmac, uuid)
* **time** - Time utilities (now, format, sleep)
* **url** - URL utilities (parse, encode, decode)
* **strings** - String manipulation
* **random** - Random generators
* **base64** - Base64 encoding/decoding

### Example: Counter Function

```lua
function handler(ctx, event)
  -- Get current count from KV store
  local count = kv.get("counter") or "0"
  local newCount = tonumber(count) + 1
  
  -- Save updated count
  kv.set("counter", tostring(newCount))
  
  log.info("Counter incremented to: " .. newCount)
  
  return {
    statusCode = 200,
    headers = { ["Content-Type"] = "application/json" },
    body = json.encode({ count = newCount })
  }
end
```

### Calling Functions

```bash
curl -X GET http://localhost:3000/fn/{function-id}
curl -X POST http://localhost:3000/fn/{function-id} -d '{"key":"value"}'
curl -X GET http://localhost:3000/fn/{function-id}?name=John
```

## Configuration

FaaS-Go can be configured via environment variables:

```bash
PORT=3000                 # HTTP server port (default: 3000)
DATA_DIR=./data           # Data directory for SQLite database (default: ./data)
EXECUTION_TIMEOUT=300     # Function execution timeout in seconds (default: 300)
```

## Architecture

* **Backend** - Go with standard library HTTP server, SQLite database
* **Frontend** - Mithril.js SPA with Monaco Editor
* **Runtime** - GopherLua for Lua script execution
* **Storage** - SQLite for functions, versions, executions, KV store, and environment variables

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Author

Claudemiro Alves Feitosa Neto

## License

MIT License - see [LICENSE](LICENSE) file for details.

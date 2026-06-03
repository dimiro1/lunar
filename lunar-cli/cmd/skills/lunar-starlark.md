---
name: lunar-starlark
description: Starlark function authoring guide for the Lunar FaaS platform. Use when writing, reviewing, or debugging Lunar Starlark functions. Covers the handler signature, all stdlib modules (log, kv, env, http, json, base64, crypto, time, url, strings, ai, email, random, router), and common patterns.
argument-hint: [module or pattern]
---

# Lunar Starlark Function Guide

Every function is a `.star` file that defines a `handler` function. The runtime
calls it on each HTTP request. Starlark is a deterministic, sandboxed dialect of
Python ([google/starlark-go](https://github.com/google/starlark-go)).

> Starlark is **not** full Python. There is no `while` loop (use `for` over a
> range), no recursion, no classes, no exceptions (`try`/`except`), and no Python
> standard library. Use the Lunar modules below for I/O and helpers.

> For the full, always-up-to-date API reference run `lunar llms` or fetch
> `/llms.txt` from your Lunar server. A function's language is chosen when it is
> created and stays fixed for its lifetime.

## Handler Signature

```python
def handler(ctx, event):
    return {
        "statusCode": 200,                       # optional, defaults to 200
        "body": "response text",
        "headers": {"Content-Type": "text/plain"},
        "isBase64Encoded": False,                # optional
    }
```

The handler returns a **dict**. `ctx` and `event` are **structs** (attribute
access). Functions that can fail return a `(value, error)` tuple — unpack both:

```python
resp, err = http.get("https://example.com")
if err != None:
    return {"statusCode": 502, "body": err}
```

### ctx — execution metadata (attribute access)

| Field | Type | Description |
|-------|------|-------------|
| `ctx.executionId` | string | Unique ID for this execution |
| `ctx.functionId` | string | Function identifier |
| `ctx.functionName` | string | Function name |
| `ctx.version` | string | Function version |
| `ctx.requestId` | string | HTTP request ID |
| `ctx.startedAt` | int | Unix timestamp (seconds) |
| `ctx.baseUrl` | string | Base URL of the deployment |

### event — incoming HTTP request (attribute access)

| Field | Type | Description |
|-------|------|-------------|
| `event.method` | string | HTTP method (`GET`, `POST`, etc.) |
| `event.path` | string | Full path including `/fn/{id}` prefix |
| `event.relativePath` | string | Path without `/fn/{id}` prefix |
| `event.body` | string | Request body as string |
| `event.headers` | dict | Request headers — `event.headers["Name"]` |
| `event.query` | dict | Query parameters — `event.query["key"]` |

## Standard Library

### log — structured logging

```python
log.info("message")
log.debug("message")
log.warn("message")
log.error("message")
```

### kv — persistent key-value storage

Scoped to the function by default. The global store is shared across functions.

```python
# Function-scoped
val = kv.get("key")              # str | None
kv.set("key", "value")           # bool
kv.delete("key")                 # bool
kv.listKeys()                    # list[str] | None
kv.all()                         # dict | None

# Global store
kv.getGlobal("key")              # str | None
kv.setGlobal("key", "value")     # bool
kv.deleteGlobal("key")           # bool
kv.listGlobalKeys()              # list[str] | None
kv.allGlobal()                   # dict | None
```

### env — environment variables

Scoped to the function. Set via `lunar functions env <id> --env KEY=VALUE`.

```python
val = env.get("MY_KEY")          # str | None
env.set("MY_KEY", "value")       # bool
env.delete("MY_KEY")             # bool
```

### http — outbound HTTP

Each call returns `(response, error)`.

```python
res, err = http.get(url, options)
res, err = http.post(url, options)
res, err = http.put(url, options)
res, err = http.delete(url, options)
```

Options dict (all optional):
```python
{
    "headers": {"Authorization": "Bearer token"},
    "query":   {"param": "value"},
    "body":    "request body",   # POST/PUT only
}
```

Response dict: `res["statusCode"]` (int), `res["body"]` (str), `res["headers"]` (dict).

### json — encode / decode

Both return `(value, error)`:

```python
str, err = json.encode(value)
val, err = json.decode(str)
```

### base64 — encode / decode

```python
encoded = base64.encode("hello")        # str
decoded, err = base64.decode(encoded)    # (str, error)
```

### crypto — hashing, HMAC, UUID

```python
# Hashes (hex string)
crypto.md5(str); crypto.sha1(str); crypto.sha256(str); crypto.sha512(str)
# HMAC (hex string)
crypto.hmac_sha1(message, key)
crypto.hmac_sha256(message, key)
crypto.hmac_sha512(message, key)
# UUID v4
id = crypto.uuid()
```

### time — timestamps and formatting

Uses Go's reference time layout: `2006-01-02 15:04:05`.

```python
time.now()                       # Unix seconds (int)
time.format(timestamp, layout)   # str
time.parse(value, layout)        # (int, error)
time.sleep(milliseconds)
```

### url — parsing and encoding

```python
parsed, err = url.parse("https://example.com/api?k=v")
# parsed["scheme"], ["host"], ["path"], ["fragment"], ["query"], ["username"], ["password"]
encoded = url.encode("hello world")      # "hello+world"
decoded, err = url.decode("hello+world")
```

### strings — utilities

```python
strings.trim(s); strings.trimLeft(s); strings.trimRight(s)
strings.split(s, sep)            # list[str]
strings.join(list, sep)          # str
strings.hasPrefix(s, prefix)     # bool
strings.hasSuffix(s, suffix)     # bool
strings.contains(s, substr)      # bool
strings.replace(s, old, new)     # all occurrences (pass n to limit)
strings.toLower(s); strings.toUpper(s)
strings.repeat(s, count)
```

### ai — LLM chat completions

Requires `ANTHROPIC_API_KEY` or `OPENAI_API_KEY` set via `lunar functions env`.
Returns `(response, error)`.

```python
res, err = ai.chat({
    "provider": "anthropic",             # "openai" or "anthropic"
    "model": "claude-haiku-4-5-20251001",
    "messages": [
        {"role": "system", "content": "You are helpful"},
        {"role": "user", "content": user_message},
    ],
    "max_tokens": 1024,                  # optional, default 1024
    "temperature": 0.7,                  # optional
    "endpoint": "https://custom.api",    # optional override
})
# res["content"], res["model"], res["usage"]["input_tokens"], res["usage"]["output_tokens"]
```

### email — send via Resend

Requires `RESEND_API_KEY` set via `lunar functions env`. Returns `(result, error)`.

```python
res, err = email.send({
    "from": "sender@yourdomain.com",     # required
    "to": "user@example.com",            # str or list
    "subject": "Hello!",                 # required
    "html": "<p>Body</p>",               # required if no text
    "text": "Plain text",                # required if no html
    "cc": "cc@example.com",              # optional
    "bcc": ["bcc@example.com"],          # optional
    "reply_to": "reply@example.com",     # optional
    "headers": {"X-Custom": "v"},        # optional
    "tags": [{"name": "n", "value": "v"}],  # optional
    "scheduled_at": time.now() + 3600,   # optional (Unix ts or ISO 8601)
})
# res["id"]  (Resend email ID)
```

### random — secure random generation

```python
random.int(min, max)             # integer in [min, max] inclusive
random.float()                   # float in [0.0, 1.0)
random.string(length)            # alphanumeric string
random.bytes(length)             # (str, error) base64-encoded random bytes
random.hex(length)               # (str, error) hex-encoded random bytes
random.id()                      # globally unique sortable 20-char xid
```

### router — path matching and URL building

```python
router.match("/users/42", "/users/:id")          # True
router.params("/users/42", "/users/:id")          # {"id": "42"}
router.match("/files/a/b/c", "/files/*")           # True (wildcard, end only)
router.path("/users/:id", {"id": "42"})            # "/fn/{id}/users/42"
router.url("/users/:id", {"id": "42"})             # "http://host/fn/{id}/users/42"
```

## Common Patterns

### Parse JSON body

```python
data, err = json.decode(event.body)
if err != None:
    body, _ = json.encode({"error": "Invalid JSON: " + err})
    return {"statusCode": 400, "body": body}
```

### JSON response helper

```python
def json_response(status, data):
    body, _ = json.encode(data)
    return {
        "statusCode": status,
        "headers": {"Content-Type": "application/json"},
        "body": body,
    }
```

### Simple REST router

```python
def handler(ctx, event):
    path = event.relativePath
    if event.method == "GET" and router.match(path, "/users/:id"):
        params = router.params(path, "/users/:id")
        return json_response(200, {"id": params["id"]})
    return json_response(404, {"error": "not found"})
```

## Creating a Starlark function

The language is selected at creation and cannot be changed afterward; later
`update`s reuse the function's language.

```sh
lunar functions create --name my-fn --language starlark --code - < handler.star
lunar functions update <id> --code - < handler.star   # stays Starlark
```

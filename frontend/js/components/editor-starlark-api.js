/**
 * @fileoverview Starlark editor language config: API documentation,
 * hover/completion snippets, and the Monaco language mapping (Starlark reuses
 * Monaco's "python" grammar). Consumed by editor-completions.js.
 *
 * @typedef {import('./editor-lua-api.js').APIDocEntry} APIDocEntry
 * @typedef {import('./editor-lua-api.js').EditorSnippet} EditorSnippet
 * @typedef {import('./editor-lua-api.js').EditorLanguageConfig} EditorLanguageConfig
 */

/** @type {Object.<string, APIDocEntry>} */
const docs = {
  "ctx.executionId": {
    signature: "ctx.executionId: str",
    snippet: "ctx.executionId",
    description: "Unique identifier for this execution",
  },
  "ctx.functionId": {
    signature: "ctx.functionId: str",
    snippet: "ctx.functionId",
    description: "Function identifier",
  },
  "ctx.functionName": {
    signature: "ctx.functionName: str",
    snippet: "ctx.functionName",
    description: "Function name",
  },
  "ctx.version": {
    signature: "ctx.version: str",
    snippet: "ctx.version",
    description: "Function version",
  },
  "ctx.requestId": {
    signature: "ctx.requestId: str",
    snippet: "ctx.requestId",
    description: "HTTP request identifier",
  },
  "ctx.startedAt": {
    signature: "ctx.startedAt: int",
    snippet: "ctx.startedAt",
    description: "Execution start timestamp (Unix seconds)",
  },
  "ctx.baseUrl": {
    signature: "ctx.baseUrl: str",
    snippet: "ctx.baseUrl",
    description: "Base URL of the server deployment",
  },
  "event.method": {
    signature: "event.method: str",
    snippet: "event.method",
    description: "HTTP method (GET, POST, PUT, DELETE, etc.)",
  },
  "event.path": {
    signature: "event.path: str",
    snippet: "event.path",
    description: "Request path",
  },
  "event.body": {
    signature: "event.body: str",
    snippet: "event.body",
    description: "Request body as string",
  },
  "event.headers": {
    signature: "event.headers: dict",
    snippet: 'event.headers["${1:Header-Name}"]',
    description: 'Request headers dict, e.g. event.headers["Content-Type"]',
  },
  "event.query": {
    signature: "event.query: dict",
    snippet: 'event.query["${1:param}"]',
    description: 'Query parameters dict, e.g. event.query["id"]',
  },
  "event.relativePath": {
    signature: "event.relativePath: str",
    snippet: "event.relativePath",
    description:
      "Request path without /fn/{function_id} prefix (e.g., /api/users)",
  },
  "log.info": {
    signature: "log.info(message: str)",
    snippet: 'log.info("${1:message}")',
    description: "Log an informational message",
  },
  "log.debug": {
    signature: "log.debug(message: str)",
    snippet: 'log.debug("${1:message}")',
    description: "Log a debug message",
  },
  "log.warn": {
    signature: "log.warn(message: str)",
    snippet: 'log.warn("${1:message}")',
    description: "Log a warning message",
  },
  "log.error": {
    signature: "log.error(message: str)",
    snippet: 'log.error("${1:message}")',
    description: "Log an error message",
  },
  "kv.get": {
    signature: "kv.get(key: str) -> str | None",
    snippet: 'kv.get("${1:key}")',
    description:
      "Get a value from the key-value store. Returns None if key does not exist.",
  },
  "kv.set": {
    signature: "kv.set(key: str, value: str) -> bool",
    snippet: 'kv.set("${1:key}", "${2:value}")',
    description: "Set a key-value pair in the store",
  },
  "kv.delete": {
    signature: "kv.delete(key: str) -> bool",
    snippet: 'kv.delete("${1:key}")',
    description: "Delete a key from the store",
  },
  "kv.listKeys": {
    signature: "kv.listKeys() -> list",
    snippet: "kv.listKeys()",
    description: "List all keys in the key-value store",
  },
  "kv.all": {
    signature: "kv.all() -> dict",
    snippet: "kv.all()",
    description: "Get all function-scoped key-value pairs from the store",
  },
  "kv.getGlobal": {
    signature: "kv.getGlobal(key: str) -> str | None",
    snippet: 'kv.getGlobal("${1:key}")',
    description:
      "Get a value from the global key-value store. Returns None if key does not exist.",
  },
  "kv.setGlobal": {
    signature: "kv.setGlobal(key: str, value: str) -> bool",
    snippet: 'kv.setGlobal("${1:key}", "${2:value}")',
    description: "Set a key-value pair in the global store",
  },
  "kv.deleteGlobal": {
    signature: "kv.deleteGlobal(key: str) -> bool",
    snippet: 'kv.deleteGlobal("${1:key}")',
    description: "Delete a key from the global store",
  },
  "kv.listGlobalKeys": {
    signature: "kv.listGlobalKeys() -> list",
    snippet: "kv.listGlobalKeys()",
    description: "List all keys in the global key-value store",
  },
  "kv.allGlobal": {
    signature: "kv.allGlobal() -> dict",
    snippet: "kv.allGlobal()",
    description: "Get all global key-value pairs from the store",
  },
  "env.get": {
    signature: "env.get(key: str) -> str | None",
    snippet: 'env.get("${1:key}")',
    description: "Get an environment variable. Returns None if not set.",
  },
  "http.get": {
    signature: "http.get(url: str, options: dict = {}) -> (dict, error)",
    snippet: 'http.get("${1:url}")',
    description:
      "Make a GET request. Returns a (response, error) tuple; response has statusCode, body, headers.",
  },
  "http.post": {
    signature: "http.post(url: str, options: dict = {}) -> (dict, error)",
    snippet: 'http.post("${1:url}", {"body": "${2:body}"})',
    description: "Make a POST request. Returns a (response, error) tuple.",
  },
  "http.put": {
    signature: "http.put(url: str, options: dict = {}) -> (dict, error)",
    snippet: 'http.put("${1:url}", {"body": "${2:body}"})',
    description: "Make a PUT request. Returns a (response, error) tuple.",
  },
  "http.delete": {
    signature: "http.delete(url: str, options: dict = {}) -> (dict, error)",
    snippet: 'http.delete("${1:url}")',
    description: "Make a DELETE request. Returns a (response, error) tuple.",
  },
  "json.encode": {
    signature: "json.encode(value) -> (str, error)",
    snippet: "json.encode(${1:value})",
    description: "Encode a value (dict/list/...) to a JSON string",
  },
  "json.decode": {
    signature: "json.decode(s: str) -> (value, error)",
    snippet: "json.decode(${1:event.body})",
    description: "Decode a JSON string to a Starlark value",
  },
  "base64.encode": {
    signature: "base64.encode(s: str) -> str",
    snippet: 'base64.encode("${1:string}")',
    description: "Encode a string to base64",
  },
  "base64.decode": {
    signature: "base64.decode(s: str) -> (str, error)",
    snippet: 'base64.decode("${1:base64String}")',
    description: "Decode a base64 string",
  },
  "crypto.md5": {
    signature: "crypto.md5(s: str) -> str",
    snippet: 'crypto.md5("${1:string}")',
    description: "Computes MD5 hash and returns hex-encoded result",
  },
  "crypto.sha1": {
    signature: "crypto.sha1(s: str) -> str",
    snippet: 'crypto.sha1("${1:string}")',
    description: "Computes SHA1 hash and returns hex-encoded result",
  },
  "crypto.sha256": {
    signature: "crypto.sha256(s: str) -> str",
    snippet: 'crypto.sha256("${1:string}")',
    description: "Computes SHA256 hash and returns hex-encoded result",
  },
  "crypto.sha512": {
    signature: "crypto.sha512(s: str) -> str",
    snippet: 'crypto.sha512("${1:string}")',
    description: "Computes SHA512 hash and returns hex-encoded result",
  },
  "crypto.hmac_sha1": {
    signature: "crypto.hmac_sha1(message: str, key: str) -> str",
    snippet: 'crypto.hmac_sha1("${1:message}", "${2:key}")',
    description: "Computes HMAC-SHA1 and returns hex-encoded result",
  },
  "crypto.hmac_sha256": {
    signature: "crypto.hmac_sha256(message: str, key: str) -> str",
    snippet: 'crypto.hmac_sha256("${1:message}", "${2:key}")',
    description: "Computes HMAC-SHA256 and returns hex-encoded result",
  },
  "crypto.hmac_sha512": {
    signature: "crypto.hmac_sha512(message: str, key: str) -> str",
    snippet: 'crypto.hmac_sha512("${1:message}", "${2:key}")',
    description: "Computes HMAC-SHA512 and returns hex-encoded result",
  },
  "crypto.uuid": {
    signature: "crypto.uuid() -> str",
    snippet: "crypto.uuid()",
    description: "Generates a new UUID v4 (36 characters)",
  },
  "time.now": {
    signature: "time.now() -> int",
    snippet: "time.now()",
    description: "Returns current Unix timestamp in seconds",
  },
  "time.format": {
    signature: "time.format(timestamp: int, layout: str) -> str",
    snippet: 'time.format(${1:timestamp}, "${2:2006-01-02 15:04:05}")',
    description:
      'Formats Unix timestamp to string using Go time layout (e.g., "2006-01-02 15:04:05")',
  },
  "time.parse": {
    signature: "time.parse(value: str, layout: str) -> (int, error)",
    snippet: 'time.parse("${1:value}", "${2:2006-01-02 15:04:05}")',
    description: "Parses time string using layout",
  },
  "time.sleep": {
    signature: "time.sleep(milliseconds: int)",
    snippet: "time.sleep(${1:milliseconds})",
    description: "Sleeps for specified milliseconds",
  },
  "url.parse": {
    signature: "url.parse(url: str) -> (dict, error)",
    snippet: 'url.parse("${1:url}")',
    description:
      "Parses URL into a dict with scheme, host, path, query, fragment",
  },
  "url.encode": {
    signature: "url.encode(s: str) -> str",
    snippet: 'url.encode("${1:string}")',
    description: "URL-encodes a string",
  },
  "url.decode": {
    signature: "url.decode(s: str) -> (str, error)",
    snippet: 'url.decode("${1:encodedString}")',
    description: "URL-decodes a string",
  },
  "strings.trim": {
    signature: "strings.trim(s: str) -> str",
    snippet: 'strings.trim("${1:string}")',
    description: "Removes leading and trailing whitespace",
  },
  "strings.trimLeft": {
    signature: "strings.trimLeft(s: str) -> str",
    snippet: 'strings.trimLeft("${1:string}")',
    description: "Removes leading whitespace",
  },
  "strings.trimRight": {
    signature: "strings.trimRight(s: str) -> str",
    snippet: 'strings.trimRight("${1:string}")',
    description: "Removes trailing whitespace",
  },
  "strings.split": {
    signature: "strings.split(s: str, sep: str) -> list",
    snippet: 'strings.split("${1:string}", "${2:separator}")',
    description: "Splits string by separator; returns a list",
  },
  "strings.join": {
    signature: "strings.join(parts: list, sep: str) -> str",
    snippet: 'strings.join(${1:parts}, "${2:separator}")',
    description: "Joins list elements with separator",
  },
  "strings.hasPrefix": {
    signature: "strings.hasPrefix(s: str, prefix: str) -> bool",
    snippet: 'strings.hasPrefix("${1:string}", "${2:prefix}")',
    description: "Returns True if string starts with prefix",
  },
  "strings.hasSuffix": {
    signature: "strings.hasSuffix(s: str, suffix: str) -> bool",
    snippet: 'strings.hasSuffix("${1:string}", "${2:suffix}")',
    description: "Returns True if string ends with suffix",
  },
  "strings.replace": {
    signature:
      "strings.replace(s: str, old: str, new: str, n: int = -1) -> str",
    snippet: 'strings.replace("${1:string}", "${2:old}", "${3:new}")',
    description: "Replaces occurrences; n=-1 for all (default), 1 for first",
  },
  "strings.toLower": {
    signature: "strings.toLower(s: str) -> str",
    snippet: 'strings.toLower("${1:string}")',
    description: "Converts string to lowercase",
  },
  "strings.toUpper": {
    signature: "strings.toUpper(s: str) -> str",
    snippet: 'strings.toUpper("${1:string}")',
    description: "Converts string to uppercase",
  },
  "strings.contains": {
    signature: "strings.contains(s: str, substr: str) -> bool",
    snippet: 'strings.contains("${1:string}", "${2:substring}")',
    description: "Returns True if string contains substring",
  },
  "strings.repeat": {
    signature: "strings.repeat(s: str, n: int) -> str",
    snippet: 'strings.repeat("${1:string}", ${2:n})',
    description: "Repeats string n times",
  },
  "random.int": {
    signature: "random.int(min: int, max: int) -> int",
    snippet: "random.int(${1:min}, ${2:max})",
    description: "Generates random integer between min and max (inclusive)",
  },
  "random.float": {
    signature: "random.float() -> float",
    snippet: "random.float()",
    description: "Generates random float between 0.0 and 1.0",
  },
  "random.string": {
    signature: "random.string(length: int) -> str",
    snippet: "random.string(${1:length})",
    description: "Generates random alphanumeric string",
  },
  "random.bytes": {
    signature: "random.bytes(length: int) -> (str, error)",
    snippet: "random.bytes(${1:length})",
    description: "Generates random bytes and returns base64-encoded string",
  },
  "random.hex": {
    signature: "random.hex(length: int) -> (str, error)",
    snippet: "random.hex(${1:length})",
    description: "Generates random bytes and returns hex-encoded string",
  },
  "random.id": {
    signature: "random.id() -> str",
    snippet: "random.id()",
    description: "Generates globally unique sortable ID (20-character string)",
  },
  "ai.chat": {
    signature: "ai.chat(options: dict) -> (dict, error)",
    snippet: `ai.chat({
    "provider": "\${1:openai}",
    "model": "\${2:gpt-4o-mini}",
    "messages": [
        {"role": "user", "content": "\${3:Hello}"}
    ]
})`,
    description:
      "Send chat completion request to AI provider (openai or anthropic). Returns a (response, error) tuple; response has content, model, usage.",
  },
  "email.send": {
    signature: "email.send(options: dict) -> (dict, error)",
    snippet: `email.send({
    "from": "\${1:sender@example.com}",
    "to": "\${2:recipient@example.com}",
    "subject": "\${3:Subject}",
    "text": "\${4:Message body}"
})`,
    description:
      "Send email via Resend. Requires RESEND_API_KEY env var. Returns a (response, error) tuple; response has id.",
  },
  "router.match": {
    signature: "router.match(path: str, pattern: str) -> bool",
    snippet: 'router.match(${1:event.relativePath}, "${2:/users/:id}")',
    description:
      "Test if path matches pattern. Use :name for parameters, * for wildcard.",
  },
  "router.params": {
    signature: "router.params(path: str, pattern: str) -> dict",
    snippet: 'router.params(${1:event.relativePath}, "${2:/users/:id}")',
    description:
      'Extract parameters from path. Returns a dict keyed by param name (e.g., {"id": "42"}).',
  },
  "router.path": {
    signature: "router.path(pattern: str, params: dict = {}) -> str",
    snippet: 'router.path("${1:/users/:id}", {"${2:id}": "${3:value}"})',
    description:
      "Build full path with /fn/{functionId} prefix. Supports parameter substitution.",
  },
  "router.url": {
    signature: "router.url(pattern: str, params: dict = {}) -> str",
    snippet: 'router.url("${1:/users/:id}", {"${2:id}": "${3:value}"})',
    description:
      "Build full URL including base URL and /fn/{functionId} prefix. Supports parameter substitution.",
  },
};

/** @type {EditorSnippet[]} */
const snippets = [
  {
    label: "handler",
    documentation: "HTTP handler function template",
    insertText: [
      "def handler(ctx, event):",
      '    log.info("Function started")',
      "",
      "    return {",
      '        "statusCode": 200,',
      '        "headers": {"Content-Type": "application/json"},',
      '        "body": \'{"message": "Hello"}\',',
      "    }",
    ].join("\n"),
  },
  {
    label: "example-counter",
    documentation: "Simple counter using KV store",
    insertText: [
      "def handler(ctx, event):",
      "    # Get current count from KV store",
      '    count = kv.get("counter") or "0"',
      "    new_count = int(count) + 1",
      "",
      "    # Save updated count",
      '    kv.set("counter", str(new_count))',
      '    log.info("Counter incremented to: " + str(new_count))',
      "",
      '    body, _ = json.encode({"count": new_count})',
      "    return {",
      '        "statusCode": 200,',
      '        "headers": {"Content-Type": "application/json"},',
      '        "body": body,',
      "    }",
    ].join("\n"),
  },
  {
    label: "example-hello",
    documentation: "Hello world with query parameters",
    insertText: [
      "def handler(ctx, event):",
      "    # Read query parameters from event",
      '    name = event.query.get("name", "World")',
      "",
      '    log.info("Greeting: " + name)',
      "",
      '    body, _ = json.encode({"message": "Hello, " + name + "!"})',
      "    return {",
      '        "statusCode": 200,',
      '        "headers": {"Content-Type": "application/json"},',
      '        "body": body,',
      "    }",
    ].join("\n"),
  },
  {
    label: "example-health",
    documentation: "Health check example",
    insertText: [
      "def handler(ctx, event):",
      "    # Check if a website is up",
      '    resp, err = http.get("https://www.google.com/")',
      "",
      '    if err == None and resp["statusCode"] == 200:',
      '        log.info("Site is up")',
      '        body, _ = json.encode({"status": "UP"})',
      "        return {",
      '            "statusCode": 200,',
      '            "headers": {"Content-Type": "application/json"},',
      '            "body": body,',
      "        }",
      "",
      '    log.error("Site is down")',
      '    body, _ = json.encode({"status": "DOWN"})',
      "    return {",
      '        "statusCode": 502,',
      '        "headers": {"Content-Type": "application/json"},',
      '        "body": body,',
      "    }",
    ].join("\n"),
  },
];

/** @type {EditorLanguageConfig} */
export const starlarkEditorApi = {
  language: "starlark",
  monacoLanguage: "python",
  hoverFence: "python",
  docs,
  snippets,
};

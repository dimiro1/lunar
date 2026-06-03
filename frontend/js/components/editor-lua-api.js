/**
 * @fileoverview Lua editor language config: API documentation, hover/completion
 * snippets, and the Monaco language mapping. Consumed by editor-completions.js.
 */

/**
 * @typedef {Object} APIDocEntry
 * @property {string} signature - Function signature for hover display
 * @property {string} snippet - Code snippet for autocompletion
 * @property {string} description - Description of the function
 */

/**
 * @typedef {Object} EditorSnippet
 * @property {string} label - Completion label
 * @property {string} insertText - Snippet body (Monaco snippet syntax)
 * @property {string} documentation - Short description
 */

/**
 * @typedef {Object} EditorLanguageConfig
 * @property {string} language - Lunar function language ("lua", "starlark")
 * @property {string} monacoLanguage - Monaco language id used for the editor
 * @property {string} hoverFence - Fenced-code language used in hover popups
 * @property {Object.<string, APIDocEntry>} docs - Per-symbol API documentation
 * @property {EditorSnippet[]} snippets - Standalone snippets (handler, examples)
 */

/** @type {Object.<string, APIDocEntry>} */
const docs = {
  "ctx.executionId": {
    signature: "ctx.executionId: string",
    snippet: "ctx.executionId",
    description: "Unique identifier for this execution",
  },
  "ctx.functionId": {
    signature: "ctx.functionId: string",
    snippet: "ctx.functionId",
    description: "Function identifier",
  },
  "ctx.functionName": {
    signature: "ctx.functionName: string",
    snippet: "ctx.functionName",
    description: "Function name",
  },
  "ctx.version": {
    signature: "ctx.version: string",
    snippet: "ctx.version",
    description: "Function version",
  },
  "ctx.requestId": {
    signature: "ctx.requestId: string",
    snippet: "ctx.requestId",
    description: "HTTP request identifier",
  },
  "ctx.startedAt": {
    signature: "ctx.startedAt: number",
    snippet: "ctx.startedAt",
    description: "Execution start timestamp (Unix seconds)",
  },
  "ctx.baseUrl": {
    signature: "ctx.baseUrl: string",
    snippet: "ctx.baseUrl",
    description: "Base URL of the server deployment",
  },
  "event.method": {
    signature: "event.method: string",
    snippet: "event.method",
    description: "HTTP method (GET, POST, PUT, DELETE, etc.)",
  },
  "event.path": {
    signature: "event.path: string",
    snippet: "event.path",
    description: "Request path",
  },
  "event.body": {
    signature: "event.body: string",
    snippet: "event.body",
    description: "Request body as string",
  },
  "event.headers": {
    signature: "event.headers: table",
    snippet: "event.headers",
    description: "Request headers (table with header name as key)",
  },
  "event.query": {
    signature: "event.query: table",
    snippet: "event.query",
    description: "Query parameters (table with param name as key)",
  },
  "event.relativePath": {
    signature: "event.relativePath: string",
    snippet: "event.relativePath",
    description:
      "Request path without /fn/{function_id} prefix (e.g., /api/users)",
  },
  "log.info": {
    signature: "log.info(message: string)",
    snippet: 'log.info("${1:message}")',
    description: "Log an informational message",
  },
  "log.debug": {
    signature: "log.debug(message: string)",
    snippet: 'log.debug("${1:message}")',
    description: "Log a debug message",
  },
  "log.warn": {
    signature: "log.warn(message: string)",
    snippet: 'log.warn("${1:message}")',
    description: "Log a warning message",
  },
  "log.error": {
    signature: "log.error(message: string)",
    snippet: 'log.error("${1:message}")',
    description: "Log an error message",
  },
  "kv.get": {
    signature: "kv.get(key: string): string | nil",
    snippet: 'kv.get("${1:key}")',
    description:
      "Get a value from the key-value store. Returns nil if key does not exist.",
  },
  "kv.set": {
    signature: "kv.set(key: string, value: string)",
    snippet: 'kv.set("${1:key}", "${2:value}")',
    description: "Set a key-value pair in the store",
  },
  "kv.delete": {
    signature: "kv.delete(key: string)",
    snippet: 'kv.delete("${1:key}")',
    description: "Delete a key from the store",
  },
  "kv.listKeys": {
    signature: "kv.listKeys(): table",
    snippet: "kv.listKeys()",
    description: "List all keys in the key-value store",
  },
  "kv.all": {
    signature: "kv.all(): table",
    snippet: "kv.all()",
    description: "Get all function-scoped key-value pairs from the store",
  },
  "kv.getGlobal": {
    signature: "kv.getGlobal(key: string): string | nil",
    snippet: 'kv.getGlobal("${1:key}")',
    description:
      "Get a value from the global key-value store. Returns nil if key does not exist.",
  },
  "kv.setGlobal": {
    signature: "kv.setGlobal(key: string, value: string)",
    snippet: 'kv.setGlobal("${1:key}", "${2:value}")',
    description: "Set a key-value pair in the global store",
  },
  "kv.deleteGlobal": {
    signature: "kv.deleteGlobal(key: string)",
    snippet: 'kv.deleteGlobal("${1:key}")',
    description: "Delete a key from the global store",
  },
  "kv.listGlobalKeys": {
    signature: "kv.listGlobalKeys(): table",
    snippet: "kv.listGlobalKeys()",
    description: "List all keys in the global key-value store",
  },
  "kv.allGlobal": {
    signature: "kv.allGlobal(): table",
    snippet: "kv.allGlobal()",
    description: "Get all global key-value pairs from the store",
  },
  "env.get": {
    signature: "env.get(key: string): string | nil",
    snippet: 'env.get("${1:key}")',
    description: "Get an environment variable. Returns nil if not set.",
  },
  "http.get": {
    signature: "http.get(url: string): {status, body, headers}",
    snippet: 'http.get("${1:url}")',
    description:
      "Make a GET request. Returns table with status, body, and headers.",
  },
  "http.post": {
    signature: "http.post(url: string, body: string): {status, body, headers}",
    snippet: 'http.post("${1:url}", "${2:body}")',
    description: "Make a POST request.",
  },
  "http.put": {
    signature: "http.put(url: string, body: string): {status, body, headers}",
    snippet: 'http.put("${1:url}", "${2:body}")',
    description: "Make a PUT request.",
  },
  "http.delete": {
    signature: "http.delete(url: string): {status, body, headers}",
    snippet: 'http.delete("${1:url}")',
    description: "Make a DELETE request.",
  },
  "json.encode": {
    signature: "json.encode(table: table): string",
    snippet: "json.encode(${1:table})",
    description: "Encode a Lua table to JSON string",
  },
  "json.decode": {
    signature: "json.decode(jsonString: string): table",
    snippet: 'json.decode("${1:jsonString}")',
    description: "Decode a JSON string to Lua table",
  },
  "base64.encode": {
    signature: "base64.encode(str: string): string",
    snippet: 'base64.encode("${1:string}")',
    description: "Encode a string to base64",
  },
  "base64.decode": {
    signature: "base64.decode(base64Str: string): string",
    snippet: 'base64.decode("${1:base64String}")',
    description: "Decode a base64 string",
  },
  "crypto.md5": {
    signature: "crypto.md5(str: string): string",
    snippet: 'crypto.md5("${1:string}")',
    description: "Computes MD5 hash and returns hex-encoded result",
  },
  "crypto.sha1": {
    signature: "crypto.sha1(str: string): string",
    snippet: 'crypto.sha1("${1:string}")',
    description: "Computes SHA1 hash and returns hex-encoded result",
  },
  "crypto.sha256": {
    signature: "crypto.sha256(str: string): string",
    snippet: 'crypto.sha256("${1:string}")',
    description: "Computes SHA256 hash and returns hex-encoded result",
  },
  "crypto.sha512": {
    signature: "crypto.sha512(str: string): string",
    snippet: 'crypto.sha512("${1:string}")',
    description: "Computes SHA512 hash and returns hex-encoded result",
  },
  "crypto.hmac_sha1": {
    signature: "crypto.hmac_sha1(message: string, key: string): string",
    snippet: 'crypto.hmac_sha1("${1:message}", "${2:key}")',
    description: "Computes HMAC-SHA1 and returns hex-encoded result",
  },
  "crypto.hmac_sha256": {
    signature: "crypto.hmac_sha256(message: string, key: string): string",
    snippet: 'crypto.hmac_sha256("${1:message}", "${2:key}")',
    description: "Computes HMAC-SHA256 and returns hex-encoded result",
  },
  "crypto.hmac_sha512": {
    signature: "crypto.hmac_sha512(message: string, key: string): string",
    snippet: 'crypto.hmac_sha512("${1:message}", "${2:key}")',
    description: "Computes HMAC-SHA512 and returns hex-encoded result",
  },
  "crypto.uuid": {
    signature: "crypto.uuid(): string",
    snippet: "crypto.uuid()",
    description: "Generates a new UUID v4 (36 characters)",
  },
  "time.now": {
    signature: "time.now(): number",
    snippet: "time.now()",
    description: "Returns current Unix timestamp in seconds",
  },
  "time.format": {
    signature: "time.format(timestamp: number, layout: string): string",
    snippet: 'time.format(${1:timestamp}, "${2:2006-01-02 15:04:05}")',
    description:
      'Formats Unix timestamp to string using Go time layout (e.g., "2006-01-02 15:04:05")',
  },
  "time.parse": {
    signature:
      "time.parse(timeStr: string, layout: string): number | nil, error | nil",
    snippet: 'time.parse("${1:timeStr}", "${2:2006-01-02 15:04:05}")',
    description: "Parses time string using layout",
  },
  "time.sleep": {
    signature: "time.sleep(milliseconds: number)",
    snippet: "time.sleep(${1:milliseconds})",
    description: "Sleeps for specified milliseconds",
  },
  "url.parse": {
    signature: "url.parse(urlStr: string): table | nil, error | nil",
    snippet: 'url.parse("${1:url}")',
    description:
      "Parses URL into components table with scheme, host, path, query, fragment",
  },
  "url.encode": {
    signature: "url.encode(str: string): string",
    snippet: 'url.encode("${1:string}")',
    description: "URL-encodes a string",
  },
  "url.decode": {
    signature: "url.decode(encodedStr: string): string | nil, error | nil",
    snippet: 'url.decode("${1:encodedString}")',
    description: "URL-decodes a string",
  },
  "strings.trim": {
    signature: "strings.trim(str: string): string",
    snippet: 'strings.trim("${1:string}")',
    description: "Removes leading and trailing whitespace",
  },
  "strings.trimLeft": {
    signature: "strings.trimLeft(str: string): string",
    snippet: 'strings.trimLeft("${1:string}")',
    description: "Removes leading whitespace",
  },
  "strings.trimRight": {
    signature: "strings.trimRight(str: string): string",
    snippet: 'strings.trimRight("${1:string}")',
    description: "Removes trailing whitespace",
  },
  "strings.split": {
    signature: "strings.split(str: string, sep: string): table",
    snippet: 'strings.split("${1:string}", "${2:separator}")',
    description: "Splits string by separator; returns array table",
  },
  "strings.join": {
    signature: "strings.join(array: table, sep: string): string",
    snippet: 'strings.join(${1:array}, "${2:separator}")',
    description: "Joins array elements with separator",
  },
  "strings.hasPrefix": {
    signature: "strings.hasPrefix(str: string, prefix: string): boolean",
    snippet: 'strings.hasPrefix("${1:string}", "${2:prefix}")',
    description: "Returns true if string starts with prefix",
  },
  "strings.hasSuffix": {
    signature: "strings.hasSuffix(str: string, suffix: string): boolean",
    snippet: 'strings.hasSuffix("${1:string}", "${2:suffix}")',
    description: "Returns true if string ends with suffix",
  },
  "strings.replace": {
    signature:
      "strings.replace(str: string, old: string, new: string, n?: number): string",
    snippet: 'strings.replace("${1:string}", "${2:old}", "${3:new}", ${4:-1})',
    description: "Replaces occurrences; n=-1 for all, 1 for first",
  },
  "strings.toLower": {
    signature: "strings.toLower(str: string): string",
    snippet: 'strings.toLower("${1:string}")',
    description: "Converts string to lowercase",
  },
  "strings.toUpper": {
    signature: "strings.toUpper(str: string): string",
    snippet: 'strings.toUpper("${1:string}")',
    description: "Converts string to uppercase",
  },
  "strings.contains": {
    signature: "strings.contains(str: string, substr: string): boolean",
    snippet: 'strings.contains("${1:string}", "${2:substring}")',
    description: "Returns true if string contains substring",
  },
  "strings.repeat": {
    signature: "strings.repeat(str: string, n: number): string",
    snippet: 'strings.repeat("${1:string}", ${2:n})',
    description: "Repeats string n times",
  },
  "random.int": {
    signature: "random.int(min: number, max: number): number",
    snippet: "random.int(${1:min}, ${2:max})",
    description: "Generates random integer between min and max (inclusive)",
  },
  "random.float": {
    signature: "random.float(): number",
    snippet: "random.float()",
    description: "Generates random float between 0.0 and 1.0",
  },
  "random.string": {
    signature: "random.string(length: number): string",
    snippet: "random.string(${1:length})",
    description: "Generates random alphanumeric string",
  },
  "random.bytes": {
    signature: "random.bytes(length: number): string | nil, error | nil",
    snippet: "random.bytes(${1:length})",
    description: "Generates random bytes and returns base64-encoded string",
  },
  "random.hex": {
    signature: "random.hex(length: number): string | nil, error | nil",
    snippet: "random.hex(${1:length})",
    description: "Generates random bytes and returns hex-encoded string",
  },
  "random.id": {
    signature: "random.id(): string",
    snippet: "random.id()",
    description: "Generates globally unique sortable ID (20-character string)",
  },
  "ai.chat": {
    signature: "ai.chat(options: table): table | nil, error | nil",
    snippet: `ai.chat({
\tprovider = "\${1:openai}",
\tmodel = "\${2:gpt-4o-mini}",
\tmessages = {
\t\t{role = "user", content = "\${3:Hello}"}
\t}
})`,
    description:
      "Send chat completion request to AI provider (openai or anthropic). Returns {content, model, usage}.",
  },
  "email.send": {
    signature: "email.send(options: table): table | nil, error | nil",
    snippet: `email.send({
\tfrom = "\${1:sender@example.com}",
\tto = "\${2:recipient@example.com}",
\tsubject = "\${3:Subject}",
\ttext = "\${4:Message body}"
})`,
    description:
      "Send email via Resend. Requires RESEND_API_KEY env var. scheduled_at accepts Unix timestamp or ISO 8601 string. Returns {id}.",
  },
  "router.match": {
    signature: "router.match(path: string, pattern: string): boolean",
    snippet: 'router.match(${1:path}, "${2:/users/:id}")',
    description:
      "Test if path matches pattern. Use :name for parameters, * for wildcard.",
  },
  "router.params": {
    signature: "router.params(path: string, pattern: string): table",
    snippet: 'router.params(${1:path}, "${2:/users/:id}")',
    description:
      'Extract parameters from path. Returns table with param names as keys (e.g., {id = "42"}).',
  },
  "router.path": {
    signature: "router.path(pattern: string, params?: table): string",
    snippet: 'router.path("${1:/users/:id}", {${2:id} = "${3:value}"})',
    description:
      "Build full path with /fn/{functionId} prefix. Supports parameter substitution.",
  },
  "router.url": {
    signature: "router.url(pattern: string, params?: table): string",
    snippet: 'router.url("${1:/users/:id}", {${2:id} = "${3:value}"})',
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
      "function handler(ctx, event)",
      '\tlog.info("Function started")',
      "\t",
      "\treturn {",
      "\t\tstatusCode = 200,",
      '\t\theaders = { ["Content-Type"] = "application/json" },',
      '\t\tbody = \'{"message": "Hello"}\'',
      "\t}",
      "end",
    ].join("\n"),
  },
  {
    label: "example-counter",
    documentation: "Simple counter using KV store",
    insertText: [
      "function handler(ctx, event)",
      "\t-- Get current count from KV store",
      '\tlocal count = kv.get("counter") or "0"',
      "\tlocal newCount = tonumber(count) + 1",
      "\t",
      "\t-- Save updated count",
      '\tkv.set("counter", tostring(newCount))',
      "\t",
      '\tlog.info("Counter incremented to: " .. newCount)',
      "\t",
      "\treturn {",
      "\t\tstatusCode = 200,",
      '\t\theaders = { ["Content-Type"] = "application/json" },',
      "\t\tbody = json.encode({ count = newCount })",
      "\t}",
      "end",
    ].join("\n"),
  },
  {
    label: "example-hello",
    documentation: "Hello world with query parameters",
    insertText: [
      "function handler(ctx, event)",
      "\t-- Parse query parameters from event",
      '\tlocal name = "World"',
      "\tif event.query and event.query.name then",
      "\t\tname = event.query.name",
      "\tend",
      "\t",
      '\tlog.info("Greeting: " .. name)',
      "\t",
      "\treturn {",
      "\t\tstatusCode = 200,",
      '\t\theaders = { ["Content-Type"] = "application/json" },',
      '\t\tbody = json.encode({ message = "Hello, " .. name .. "!" })',
      "\t}",
      "end",
    ].join("\n"),
  },
  {
    label: "example-health",
    documentation: "Health check example",
    insertText: [
      "function handler(ctx, event)",
      "\t-- Check if a website is up",
      '\tlocal url = "https://www.google.com/"',
      "\t",
      "\tlocal response = http.get(url)",
      "\t",
      "\tif response.statusCode == 200 then",
      '\t\tlog.info("Site is up")',
      "\t\t",
      "\t\treturn {",
      "\t\t\tstatusCode = 200,",
      '\t\t\theaders = { ["Content-Type"] = "application/json" },',
      '\t\t\tbody = json.encode({ status = "UP" })',
      "\t\t}",
      "\telse",
      '\t\tlog.error("Site is down: " .. response.statusCode)',
      "\t\t",
      "\t\treturn {",
      "\t\t\tstatusCode = 502,",
      '\t\t\theaders = { ["Content-Type"] = "application/json" },',
      '\t\t\tbody = json.encode({ status = "DOWN" })',
      "\t\t}",
      "\tend",
      "end",
    ].join("\n"),
  },
];

/** @type {EditorLanguageConfig} */
export const luaEditorApi = {
  language: "lua",
  monacoLanguage: "lua",
  hoverFence: "lua",
  docs,
  snippets,
};

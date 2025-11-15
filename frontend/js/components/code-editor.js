const API_DOCS = {
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
};

const registerLuaCompletions = () => {
  if (!window.monaco || window.__luaCompletionsRegistered) return;
  window.__luaCompletionsRegistered = true;

  // Register hover provider
  monaco.languages.registerHoverProvider("lua", {
    provideHover: (model, position) => {
      const word = model.getWordAtPosition(position);
      if (!word) return null;

      const line = model.getLineContent(position.lineNumber);
      const beforeWord = line.substring(0, word.startColumn - 1);
      const match = beforeWord.match(/(\w+)\.$/);

      if (match) {
        const module = match[1];
        const fullName = `${module}.${word.word}`;
        const doc = API_DOCS[fullName];

        if (doc) {
          return {
            contents: [
              { value: `**${fullName}**` },
              { value: `\`\`\`lua\n${doc.signature}\n\`\`\`` },
              { value: doc.description },
            ],
          };
        }
      }

      return null;
    },
  });

  // Register completion provider
  monaco.languages.registerCompletionItemProvider("lua", {
    provideCompletionItems: (model, position) => {
      // Generate suggestions from API_DOCS
      const suggestions = Object.entries(API_DOCS).map(([name, doc]) => ({
        label: name,
        kind: monaco.languages.CompletionItemKind.Method,
        insertText: doc.snippet,
        insertTextRules:
          monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
        documentation: doc.description,
      }));

      // Add handler template
      suggestions.push({
        label: "handler",
        kind: monaco.languages.CompletionItemKind.Snippet,
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
        insertTextRules:
          monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
        documentation: "HTTP handler function template",
      });

      return { suggestions };
    },
  });
};

export const CodeEditor = {
  view: (vnode) => {
    const {
      id = "code-editor",
      value = "",
      onChange = null,
      readOnly = false,
      language = "lua",
      theme = "vs-dark",
      lineNumbers = true,
      minimap = false,
      height = "400px",
    } = vnode.attrs;

    return m("div", {
      id: id,
      style: `height: ${height}; border: 1px solid #444;`,
      oncreate: (divVnode) => {
        const container = divVnode.dom;
        if (container && window.monaco) {
          registerLuaCompletions();

          const editor = monaco.editor.create(container, {
            value: value || "",
            language: language,
            theme: theme,
            readOnly: readOnly,
            automaticLayout: true,
            minimap: {
              enabled: minimap,
            },
            lineNumbers: lineNumbers ? "on" : "off",
            scrollBeyondLastLine: false,
            fontSize: 14,
            tabSize: 2,
            suggestOnTriggerCharacters: true,
            quickSuggestions: true,
          });

          if (onChange) {
            editor.onDidChangeModelContent(() => {
              onChange(editor.getValue());
            });
          }

          vnode.state.editor = editor;
        }
      },
      onupdate: (divVnode) => {
        if (vnode.state.editor && value !== vnode.state.editor.getValue()) {
          const position = vnode.state.editor.getPosition();
          vnode.state.editor.setValue(value || "");
          if (position) {
            vnode.state.editor.setPosition(position);
          }
        }
      },
      onremove: () => {
        if (vnode.state.editor) {
          vnode.state.editor.dispose();
          vnode.state.editor = null;
        }
      },
    });
  },
};

import { Icons } from '../icons.js';

// Helper to create syntax-highlighted code block
const codeBlock = (code, language = 'lua') => {
  return m('pre', {
    style: 'border-radius: 6px; overflow-x: auto; margin: 0;'
  }, [
    m('code', {
      class: `language-${language}`,
      oncreate: (vnode) => {
        if (window.hljs) {
          window.hljs.highlightElement(vnode.dom);
        }
      },
      onupdate: (vnode) => {
        if (window.hljs) {
          window.hljs.highlightElement(vnode.dom);
        }
      }
    }, code.trim())
  ]);
};

// Function documentation component
export const FunctionDocs = {
  expanded: false,

  view: () => {
    return m('.card.mb-24', { style: 'border: 1px solid #404040;' }, [
      m('.card-header', {
        style: 'cursor: pointer;',
        onclick: () => {
          FunctionDocs.expanded = !FunctionDocs.expanded;
        }
      }, [
        m('.card-title', [
          Icons.arrowPath(),
          '  Function API Documentation',
        ]),
        m('.card-subtitle', FunctionDocs.expanded ? 'Click to collapse' : 'Click to expand'),
      ]),

      FunctionDocs.expanded && m('.modal-body', { style: 'max-height: 500px; overflow-y: auto;' }, [
        // Function Structure
        m('h3', { style: 'color: #ff6b2c; margin-bottom: 12px;' }, 'Function Structure'),
        codeBlock(`function handler(ctx, event)
  -- Your code here
  return {
    statusCode = 200,
    headers = { ["Content-Type"] = "application/json" },
    body = '{"message": "Hello"}'
  }
end`),

        // Event Object
        m('h3', { style: 'color: #ff6b2c; margin: 24px 0 12px;' }, 'Event Object'),
        m('p', { style: 'margin-bottom: 12px; color: #a3a3a3;' }, 'HTTP Request properties:'),
        m('ul', { style: 'margin-left: 20px; margin-bottom: 12px; color: #e5e5e5;' }, [
          m('li', m('code', { style: 'color: #86efac;' }, 'event.method'), ' - HTTP method (GET, POST, PUT, DELETE)'),
          m('li', m('code', { style: 'color: #86efac;' }, 'event.path'), ' - Request path'),
          m('li', m('code', { style: 'color: #86efac;' }, 'event.headers'), ' - Request headers table'),
          m('li', m('code', { style: 'color: #86efac;' }, 'event.body'), ' - Request body string'),
          m('li', m('code', { style: 'color: #86efac;' }, 'event.query'), ' - Query parameters table'),
        ]),

        // Available APIs
        m('h3', { style: 'color: #ff6b2c; margin: 24px 0 12px;' }, 'Available APIs'),

        // Log API
        m('h4', { style: 'color: #ffffff; margin: 16px 0 8px;' }, 'log - Logging'),
        codeBlock(`log.info("Info message")
log.debug("Debug message")
log.warn("Warning message")
log.error("Error message")`),

        // JSON API
        m('h4', { style: 'color: #ffffff; margin: 16px 0 8px;' }, 'json - JSON encoding/decoding'),
        codeBlock(`local data = json.decode(event.body)
local json_str = json.encode({ key = "value" })`),

        // KV API
        m('h4', { style: 'color: #ffffff; margin: 16px 0 8px;' }, 'kv - Key-Value Store'),
        codeBlock(`kv.set("counter", "42")
local value = kv.get("counter")
kv.delete("old-key")`),

        // HTTP API
        m('h4', { style: 'color: #ffffff; margin: 16px 0 8px;' }, 'http - HTTP Client'),
        codeBlock(`local response = http.get("https://api.example.com/data", {
  headers = { ["Authorization"] = "Bearer token" }
})

http.post("https://api.example.com/data", {
  body = '{"test": "data"}',
  headers = { ["Content-Type"] = "application/json" }
})`),

        // Blob API
        m('h4', { style: 'color: #ffffff; margin: 16px 0 8px;' }, 'blob - Blob Storage'),
        codeBlock(`-- Store blob (base64 encoded)
blob.put("files/data.txt", "SGVsbG8=", "text/plain")

-- Retrieve blob
local data = blob.get("files/data.txt")

-- Get metadata
local meta = blob.metadata("files/data.txt")

-- List blobs with prefix
local files = blob.list("files/")

-- Delete blob
blob.delete("files/data.txt")`),

        // Email API
        m('h4', { style: 'color: #ffffff; margin: 16px 0 8px;' }, 'email - Email Service'),
        codeBlock(`email.send({
  from = "sender@example.com",
  fromName = "John Sender",
  to = "recipient@example.com",
  subject = "Test Email",
  text = "Plain text body",
  html = "<p>HTML body</p>",
  cc = {"cc@example.com"},
  replyTo = "reply@example.com"
})`),

        // LLM API
        m('h4', { style: 'color: #ffffff; margin: 16px 0 8px;' }, 'llm - Large Language Models'),
        codeBlock(`local response = llm.complete({
  model = "gpt-4",
  messages = {
    { role = "system", content = "You are helpful" },
    { role = "user", content = "What is Rust?" }
  },
  temperature = 0.7,
  maxTokens = 100
})

log.info(response.message.content)`),

        // Environment Variables
        m('h4', { style: 'color: #ffffff; margin: 16px 0 8px;' }, 'env - Environment Variables'),
        codeBlock(`local api_key = env.get("API_KEY")
local database_url = env.get("DATABASE_URL")`),

        // Complete Example
        m('h3', { style: 'color: #ff6b2c; margin: 24px 0 12px;' }, 'Complete Example'),
        codeBlock(`function handler(ctx, event)
  log.info("Processing: " .. event.method)

  -- Validate API key
  local apiKey = event.headers["X-API-Key"]
  if not apiKey then
    return { statusCode = 401, body = "Unauthorized" }
  end

  if event.method == "POST" then
    -- Parse JSON body
    local data = json.decode(event.body)

    -- Store in KV
    kv.set("user:" .. data.id, json.encode(data))

    -- Call external API
    http.post("https://webhook.example.com", {
      body = event.body,
      headers = { ["Content-Type"] = "application/json" }
    })

    return {
      statusCode = 201,
      headers = { ["Content-Type"] = "application/json" },
      body = json.encode({ success = true })
    }
  end

  return { statusCode = 200, body = "OK" }
end`),
      ]),
    ]);
  }
};

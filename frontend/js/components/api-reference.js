/**
 * @fileoverview API Reference component for displaying Lua API documentation.
 */

/**
 * @typedef {Object} DocItemDef
 * @property {string} name - Item name/signature
 * @property {string} type - Type (string, number, table, function, module)
 * @property {string} description - Item description
 */

/**
 * @typedef {Object} DocGroup
 * @property {string} name - Group name
 * @property {DocItemDef[]} items - Items in the group
 */

/**
 * @typedef {Object} APISection
 * @property {string} id - Unique section identifier
 * @property {string} name - Section display name
 * @property {string} [description] - Section description
 * @property {DocItemDef[]} [items] - Direct items (if no groups)
 * @property {DocGroup[]} [groups] - Grouped items
 */

/**
 * API Reference component with tabbed sections.
 * Displays documentation for Lua API functions and modules.
 * @type {Object}
 */
export const APIReference = {
  /**
   * Renders the API reference component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {APISection[]} [vnode.attrs.sections=[]] - API sections to display
   * @param {string} [vnode.attrs.activeSection] - Currently active section ID
   * @param {(sectionId: string) => void} [vnode.attrs.onSectionChange] - Section change callback
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { sections = [], activeSection, onSectionChange } = vnode.attrs;
    const active = activeSection || sections[0]?.id;

    return m(".api-reference", [
      // Tab headers
      m(
        ".api-reference__tabs",
        sections.map((section) =>
          m(
            "button.api-reference__tab",
            {
              key: section.id,
              class: section.id === active ? "api-reference__tab--active" : "",
              onclick: () => onSectionChange && onSectionChange(section.id),
            },
            section.name,
          )
        ),
      ),
      // Content
      m(
        ".api-reference__content",
        sections
          .filter((section) => section.id === active)
          .map((section) =>
            m(".api-reference__panel", { key: section.id }, [
              section.description &&
              m("p.api-reference__description", section.description),
              section.items &&
              section.items.map((item) => m(DocItem, { key: item.name, item })),
              section.groups &&
              section.groups.map((group, i) =>
                m(".api-reference__group", { key: group.name }, [
                  m(
                    "h4.api-reference__group-header",
                    {
                      class: i === 0
                        ? "api-reference__group-header--first"
                        : "",
                    },
                    group.name,
                  ),
                  group.items.map((item) =>
                    m(DocItem, { key: item.name, item })
                  ),
                ])
              ),
            ])
          ),
      ),
      // Footer
      m(".api-reference__footer", [
        m(
          "a.api-reference__footer-link",
          {
            href: "/llms.txt",
            target: "_blank",
            rel: "noopener noreferrer",
          },
          "LLM Documentation",
        ),
      ]),
    ]);
  },
};

/**
 * Documentation item component.
 * Displays a single API item with name, type, and description.
 * @type {Object}
 * @private
 */
const DocItem = {
  /**
   * Renders the documentation item.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {DocItemDef} vnode.attrs.item - The documentation item
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { item } = vnode.attrs;
    const typeClass = getTypeClass(item.type);

    return m(".api-doc-item", [
      m(".api-doc-item__header", [
        m("span.api-doc-item__name", item.name),
        m("span.api-doc-item__type", { class: typeClass }, item.type),
      ]),
      m("p.api-doc-item__description", item.description),
    ]);
  },
};

/**
 * Gets the CSS class for a type badge.
 * @param {string} type - The type name
 * @returns {string} CSS class name
 */
function getTypeClass(type) {
  switch (type) {
    case "string":
      return "api-doc-item__type--string";
    case "number":
      return "api-doc-item__type--number";
    case "table":
      return "api-doc-item__type--table";
    case "function":
      return "api-doc-item__type--function";
    case "module":
      return "api-doc-item__type--module";
    default:
      return "api-doc-item__type--default";
  }
}

/**
 * Default API sections for Lua functions.
 * Contains documentation for handler inputs, I/O, data transformation, and utilities.
 * @type {APISection[]}
 */
export const LuaAPISections = [
  {
    id: "ai",
    name: "AI",
    description: "AI provider integrations",
    groups: [
      {
        name: "Chat (ai)",
        items: [
          {
            name: "ai.chat(options)",
            type: "function",
            description: "Chat completion with OpenAI or Anthropic",
          },
        ],
      },
    ],
  },
  {
    id: "email",
    name: "Email",
    description: "Email sending via Resend",
    groups: [
      {
        name: "Send (email)",
        items: [
          {
            name: "email.send(options)",
            type: "function",
            description: "Send email via Resend API",
          },
        ],
      },
    ],
  },
  {
    id: "handler",
    name: "Handler",
    description: "Handler function inputs",
    groups: [
      {
        name: "Context (ctx)",
        items: [
          {
            name: "ctx.executionId",
            type: "string",
            description: "Unique execution identifier",
          },
          {
            name: "ctx.functionId",
            type: "string",
            description: "Function identifier",
          },
          {
            name: "ctx.functionName",
            type: "string",
            description: "Function name",
          },
          {
            name: "ctx.version",
            type: "string",
            description: "Function version",
          },
          {
            name: "ctx.requestId",
            type: "string",
            description: "HTTP request identifier",
          },
          {
            name: "ctx.startedAt",
            type: "number",
            description: "Start timestamp (Unix)",
          },
          {
            name: "ctx.baseUrl",
            type: "string",
            description: "Server base URL",
          },
        ],
      },
      {
        name: "Event (event)",
        items: [
          {
            name: "event.method",
            type: "string",
            description: "HTTP method (GET, POST, etc.)",
          },
          { name: "event.path", type: "string", description: "Request path" },
          {
            name: "event.body",
            type: "string",
            description: "Request body as string",
          },
          {
            name: "event.headers",
            type: "table",
            description: "Request headers table",
          },
          {
            name: "event.query",
            type: "table",
            description: "Query parameters table",
          },
        ],
      },
    ],
  },
  {
    id: "io",
    name: "I/O",
    description: "Input/output operations",
    groups: [
      {
        name: "Logging (log)",
        items: [
          {
            name: "log.info(msg)",
            type: "function",
            description: "Log info message",
          },
          {
            name: "log.debug(msg)",
            type: "function",
            description: "Log debug message",
          },
          {
            name: "log.warn(msg)",
            type: "function",
            description: "Log warning message",
          },
          {
            name: "log.error(msg)",
            type: "function",
            description: "Log error message",
          },
        ],
      },
      {
        name: "Key-Value Store (kv)",
        items: [
          {
            name: "kv.get(key)",
            type: "function",
            description: "Get value from store",
          },
          {
            name: "kv.set(key, value)",
            type: "function",
            description: "Set key-value pair",
          },
          {
            name: "kv.delete(key)",
            type: "function",
            description: "Delete key from store",
          },
        ],
      },
      {
        name: "Environment (env)",
        items: [
          {
            name: "env.get(key)",
            type: "function",
            description: "Get environment variable",
          },
        ],
      },
      {
        name: "HTTP Client (http)",
        items: [
          {
            name: "http.get(url)",
            type: "function",
            description: "GET request",
          },
          {
            name: "http.post(url, body)",
            type: "function",
            description: "POST request",
          },
          {
            name: "http.put(url, body)",
            type: "function",
            description: "PUT request",
          },
          {
            name: "http.delete(url)",
            type: "function",
            description: "DELETE request",
          },
        ],
      },
    ],
  },
  {
    id: "data",
    name: "Data",
    description: "Data transformation",
    groups: [
      {
        name: "JSON (json)",
        items: [
          {
            name: "json.encode(table)",
            type: "function",
            description: "Encode table to JSON",
          },
          {
            name: "json.decode(str)",
            type: "function",
            description: "Decode JSON to table",
          },
        ],
      },
      {
        name: "Base64 (base64)",
        items: [
          {
            name: "base64.encode(str)",
            type: "function",
            description: "Encode to base64",
          },
          {
            name: "base64.decode(str)",
            type: "function",
            description: "Decode from base64",
          },
        ],
      },
      {
        name: "Crypto (crypto)",
        items: [
          {
            name: "crypto.md5(str)",
            type: "function",
            description: "MD5 hash (hex)",
          },
          {
            name: "crypto.sha256(str)",
            type: "function",
            description: "SHA256 hash (hex)",
          },
          {
            name: "crypto.hmac_sha256(msg, key)",
            type: "function",
            description: "HMAC-SHA256 (hex)",
          },
          {
            name: "crypto.uuid()",
            type: "function",
            description: "Generate UUID v4",
          },
        ],
      },
    ],
  },
  {
    id: "utils",
    name: "Utils",
    description: "Utility functions",
    groups: [
      {
        name: "Time (time)",
        items: [
          {
            name: "time.now()",
            type: "function",
            description: "Current Unix timestamp",
          },
          {
            name: "time.format(ts, layout)",
            type: "function",
            description: "Format timestamp",
          },
          {
            name: "time.parse(str, layout)",
            type: "function",
            description: "Parse time string",
          },
          {
            name: "time.sleep(ms)",
            type: "function",
            description: "Sleep milliseconds",
          },
        ],
      },
      {
        name: "Strings (strings)",
        items: [
          {
            name: "strings.trim(str)",
            type: "function",
            description: "Trim whitespace",
          },
          {
            name: "strings.split(str, sep)",
            type: "function",
            description: "Split by separator",
          },
          {
            name: "strings.join(arr, sep)",
            type: "function",
            description: "Join with separator",
          },
          {
            name: "strings.contains(str, sub)",
            type: "function",
            description: "Contains substring",
          },
          {
            name: "strings.replace(str, old, new)",
            type: "function",
            description: "Replace in string",
          },
        ],
      },
      {
        name: "Random (random)",
        items: [
          {
            name: "random.int(min, max)",
            type: "function",
            description: "Random integer",
          },
          {
            name: "random.float()",
            type: "function",
            description: "Random float 0.0-1.0",
          },
          {
            name: "random.string(len)",
            type: "function",
            description: "Random alphanumeric",
          },
          {
            name: "random.id()",
            type: "function",
            description: "Unique sortable ID",
          },
        ],
      },
    ],
  },
];

/**
 * @fileoverview API Reference component for displaying Lua API documentation.
 */

import { t } from "../i18n/index.js";

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
          t("apiReference.llmDocumentation"),
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
        m(
          "span.api-doc-item__type",
          { class: typeClass },
          getLocalizedType(item.type),
        ),
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
 * Gets the localized display name for a type.
 * @param {string} type - The type name
 * @returns {string} Localized type name
 */
function getLocalizedType(type) {
  return t(`luaApi.types.${type}`);
}

/**
 * Gets localized API sections for Lua functions.
 * Contains documentation for handler inputs, I/O, data transformation, and utilities.
 * @returns {APISection[]} Localized API sections
 */
export function getLuaAPISections() {
  return [
    {
      id: "ai",
      name: t("luaApi.ai.name"),
      description: t("luaApi.ai.description"),
      groups: [
        {
          name: t("luaApi.ai.groups.chat"),
          items: [
            {
              name: "ai.chat(options)",
              type: "function",
              description: t("luaApi.ai.items.chat"),
            },
          ],
        },
      ],
    },
    {
      id: "email",
      name: t("luaApi.email.name"),
      description: t("luaApi.email.description"),
      groups: [
        {
          name: t("luaApi.email.groups.send"),
          items: [
            {
              name: "email.send(options)",
              type: "function",
              description: t("luaApi.email.items.send"),
            },
          ],
        },
      ],
    },
    {
      id: "handler",
      name: t("luaApi.handler.name"),
      description: t("luaApi.handler.description"),
      groups: [
        {
          name: t("luaApi.handler.groups.context"),
          items: [
            {
              name: "ctx.executionId",
              type: "string",
              description: t("luaApi.handler.items.executionId"),
            },
            {
              name: "ctx.functionId",
              type: "string",
              description: t("luaApi.handler.items.functionId"),
            },
            {
              name: "ctx.functionName",
              type: "string",
              description: t("luaApi.handler.items.functionName"),
            },
            {
              name: "ctx.version",
              type: "string",
              description: t("luaApi.handler.items.version"),
            },
            {
              name: "ctx.requestId",
              type: "string",
              description: t("luaApi.handler.items.requestId"),
            },
            {
              name: "ctx.startedAt",
              type: "number",
              description: t("luaApi.handler.items.startedAt"),
            },
            {
              name: "ctx.baseUrl",
              type: "string",
              description: t("luaApi.handler.items.baseUrl"),
            },
          ],
        },
        {
          name: t("luaApi.handler.groups.event"),
          items: [
            {
              name: "event.method",
              type: "string",
              description: t("luaApi.handler.items.method"),
            },
            {
              name: "event.path",
              type: "string",
              description: t("luaApi.handler.items.path"),
            },
            {
              name: "event.body",
              type: "string",
              description: t("luaApi.handler.items.body"),
            },
            {
              name: "event.headers",
              type: "table",
              description: t("luaApi.handler.items.headers"),
            },
            {
              name: "event.query",
              type: "table",
              description: t("luaApi.handler.items.query"),
            },
          ],
        },
      ],
    },
    {
      id: "io",
      name: t("luaApi.io.name"),
      description: t("luaApi.io.description"),
      groups: [
        {
          name: t("luaApi.io.groups.logging"),
          items: [
            {
              name: "log.info(msg)",
              type: "function",
              description: t("luaApi.io.items.logInfo"),
            },
            {
              name: "log.debug(msg)",
              type: "function",
              description: t("luaApi.io.items.logDebug"),
            },
            {
              name: "log.warn(msg)",
              type: "function",
              description: t("luaApi.io.items.logWarn"),
            },
            {
              name: "log.error(msg)",
              type: "function",
              description: t("luaApi.io.items.logError"),
            },
          ],
        },
        {
          name: t("luaApi.io.groups.kv"),
          items: [
            {
              name: "kv.get(key)",
              type: "function",
              description: t("luaApi.io.items.kvGet"),
            },
            {
              name: "kv.set(key, value)",
              type: "function",
              description: t("luaApi.io.items.kvSet"),
            },
            {
              name: "kv.delete(key)",
              type: "function",
              description: t("luaApi.io.items.kvDelete"),
            },
          ],
        },
        {
          name: t("luaApi.io.groups.env"),
          items: [
            {
              name: "env.get(key)",
              type: "function",
              description: t("luaApi.io.items.envGet"),
            },
          ],
        },
        {
          name: t("luaApi.io.groups.http"),
          items: [
            {
              name: "http.get(url)",
              type: "function",
              description: t("luaApi.io.items.httpGet"),
            },
            {
              name: "http.post(url, body)",
              type: "function",
              description: t("luaApi.io.items.httpPost"),
            },
            {
              name: "http.put(url, body)",
              type: "function",
              description: t("luaApi.io.items.httpPut"),
            },
            {
              name: "http.delete(url)",
              type: "function",
              description: t("luaApi.io.items.httpDelete"),
            },
          ],
        },
      ],
    },
    {
      id: "data",
      name: t("luaApi.data.name"),
      description: t("luaApi.data.description"),
      groups: [
        {
          name: t("luaApi.data.groups.json"),
          items: [
            {
              name: "json.encode(table)",
              type: "function",
              description: t("luaApi.data.items.jsonEncode"),
            },
            {
              name: "json.decode(str)",
              type: "function",
              description: t("luaApi.data.items.jsonDecode"),
            },
          ],
        },
        {
          name: t("luaApi.data.groups.base64"),
          items: [
            {
              name: "base64.encode(str)",
              type: "function",
              description: t("luaApi.data.items.base64Encode"),
            },
            {
              name: "base64.decode(str)",
              type: "function",
              description: t("luaApi.data.items.base64Decode"),
            },
          ],
        },
        {
          name: t("luaApi.data.groups.crypto"),
          items: [
            {
              name: "crypto.md5(str)",
              type: "function",
              description: t("luaApi.data.items.md5"),
            },
            {
              name: "crypto.sha256(str)",
              type: "function",
              description: t("luaApi.data.items.sha256"),
            },
            {
              name: "crypto.hmac_sha256(msg, key)",
              type: "function",
              description: t("luaApi.data.items.hmacSha256"),
            },
            {
              name: "crypto.uuid()",
              type: "function",
              description: t("luaApi.data.items.uuid"),
            },
          ],
        },
      ],
    },
    {
      id: "utils",
      name: t("luaApi.utils.name"),
      description: t("luaApi.utils.description"),
      groups: [
        {
          name: t("luaApi.utils.groups.time"),
          items: [
            {
              name: "time.now()",
              type: "function",
              description: t("luaApi.utils.items.timeNow"),
            },
            {
              name: "time.format(ts, layout)",
              type: "function",
              description: t("luaApi.utils.items.timeFormat"),
            },
            {
              name: "time.parse(str, layout)",
              type: "function",
              description: t("luaApi.utils.items.timeParse"),
            },
            {
              name: "time.sleep(ms)",
              type: "function",
              description: t("luaApi.utils.items.timeSleep"),
            },
          ],
        },
        {
          name: t("luaApi.utils.groups.strings"),
          items: [
            {
              name: "strings.trim(str)",
              type: "function",
              description: t("luaApi.utils.items.trim"),
            },
            {
              name: "strings.split(str, sep)",
              type: "function",
              description: t("luaApi.utils.items.split"),
            },
            {
              name: "strings.join(arr, sep)",
              type: "function",
              description: t("luaApi.utils.items.join"),
            },
            {
              name: "strings.contains(str, sub)",
              type: "function",
              description: t("luaApi.utils.items.contains"),
            },
            {
              name: "strings.replace(str, old, new)",
              type: "function",
              description: t("luaApi.utils.items.replace"),
            },
          ],
        },
        {
          name: t("luaApi.utils.groups.random"),
          items: [
            {
              name: "random.int(min, max)",
              type: "function",
              description: t("luaApi.utils.items.randomInt"),
            },
            {
              name: "random.float()",
              type: "function",
              description: t("luaApi.utils.items.randomFloat"),
            },
            {
              name: "random.string(len)",
              type: "function",
              description: t("luaApi.utils.items.randomString"),
            },
            {
              name: "random.id()",
              type: "function",
              description: t("luaApi.utils.items.randomId"),
            },
          ],
        },
      ],
    },
  ];
}

/**
 * Legacy export for backwards compatibility.
 * @deprecated Use getLuaAPISections() instead
 * @type {APISection[]}
 */
export const LuaAPISections = getLuaAPISections();

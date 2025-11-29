/**
 * @fileoverview AI Request viewer component with table structure and expandable rows.
 */

import { icons } from "../icons.js";
import { Badge, BadgeSize, BadgeVariant } from "./badge.js";
import { CodeViewer } from "./code-viewer.js";
import { formatUnixTimestamp } from "../utils.js";
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from "./table.js";

import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').AIRequest} AIRequest
 */

/**
 * Maximum characters to display in JSON before truncating.
 * @type {number}
 */
const MAX_JSON_DISPLAY_LENGTH = 5000;

/**
 * AI Request viewer component for displaying AI API requests in a table with expandable details.
 * @type {Object}
 */
export const AIRequestViewer = {
  /**
   * Track which rows are expanded.
   * @type {Set<string>}
   */
  expandedRows: new Set(),

  /**
   * Toggles expansion state for a row.
   * @param {string} id - Request ID
   */
  toggleRow(id) {
    if (this.expandedRows.has(id)) {
      this.expandedRows.delete(id);
    } else {
      this.expandedRows.add(id);
    }
  },

  /**
   * Formats JSON string for display with optional truncation.
   * @param {string} jsonStr - JSON string
   * @param {boolean} [truncate=true] - Whether to truncate long content
   * @returns {{formatted: string, truncated: boolean}} Formatted JSON and truncation status
   */
  formatJSON(jsonStr, truncate = true) {
    if (!jsonStr) {
      return { formatted: "", truncated: false };
    }

    // If it's already an object, stringify it
    if (typeof jsonStr === "object") {
      const formatted = JSON.stringify(jsonStr, null, 2);
      if (truncate && formatted.length > MAX_JSON_DISPLAY_LENGTH) {
        return {
          formatted: formatted.substring(0, MAX_JSON_DISPLAY_LENGTH) +
            "\n\n" +
            t("aiRequestViewer.truncated"),
          truncated: true,
        };
      }
      return { formatted, truncated: false };
    }

    try {
      const parsed = JSON.parse(jsonStr);
      const formatted = JSON.stringify(parsed, null, 2);
      if (truncate && formatted.length > MAX_JSON_DISPLAY_LENGTH) {
        return {
          formatted: formatted.substring(0, MAX_JSON_DISPLAY_LENGTH) +
            "\n\n" +
            t("aiRequestViewer.truncated"),
          truncated: true,
        };
      }
      return { formatted, truncated: false };
    } catch {
      return { formatted: String(jsonStr), truncated: false };
    }
  },

  /**
   * Renders the AI request viewer component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {AIRequest[]} [vnode.attrs.requests=[]] - Array of AI requests
   * @param {string} [vnode.attrs.maxHeight='400px'] - Maximum height
   * @param {boolean} [vnode.attrs.noBorder=false] - Remove border styling
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      requests = [],
      maxHeight = "400px",
      noBorder = false,
    } = vnode.attrs;

    if (requests.length === 0) {
      return m(Table, [
        m(TableBody, [
          m(TableEmpty, {
            colspan: 7,
            icon: "network",
            message: t("aiRequestViewer.noRequests"),
          }),
        ]),
      ]);
    }

    return m(
      ".ai-request-viewer",
      {
        class: noBorder ? "ai-request-viewer--no-border" : "",
        style: maxHeight ? `max-height: ${maxHeight}; overflow-y: auto` : "",
      },
      [
        m(Table, { style: "table-layout: fixed;" }, [
          m(TableHeader, [
            m(TableRow, [
              m(TableHead, { style: "width: 2rem;" }, ""),
              m(
                TableHead,
                { style: "width: 10%;" },
                t("aiRequestViewer.provider"),
              ),
              m(
                TableHead,
                { style: "width: 25%;" },
                t("aiRequestViewer.model"),
              ),
              m(
                TableHead,
                { style: "width: 10%;" },
                t("aiRequestViewer.status"),
              ),
              m(
                TableHead,
                { style: "width: 20%;" },
                t("aiRequestViewer.tokens"),
              ),
              m(
                TableHead,
                { style: "width: 15%;" },
                t("aiRequestViewer.duration"),
              ),
              m(TableHead, { style: "width: 20%;" }, t("aiRequestViewer.time")),
            ]),
          ]),
          m(
            TableBody,
            requests.map((req) => this.renderRequestRows(req)),
          ),
        ]),
      ],
    );
  },

  /**
   * Renders the rows for a single AI request (main row + optional expanded row).
   * @param {AIRequest} req - The AI request
   * @returns {Object} Mithril fragment with keyed children
   */
  renderRequestRows(req) {
    const isExpanded = this.expandedRows.has(req.id);

    const rows = [
      // Main row
      m(
        TableRow,
        {
          key: req.id,
          class: "ai-request-viewer__row" +
            (isExpanded ? " ai-request-viewer__row--expanded" : ""),
          onclick: () => this.toggleRow(req.id),
          style: "cursor: pointer;",
        },
        [
          // Chevron
          m(TableCell, { style: "width: 2rem; padding-right: 0;" }, [
            m(
              ".ai-request-viewer__chevron",
              {
                class: isExpanded ? "ai-request-viewer__chevron--expanded" : "",
              },
              m.trust(icons.chevronRight()),
            ),
          ]),

          // Provider
          m(TableCell, [
            m(
              Badge,
              {
                variant: req.provider === "openai"
                  ? BadgeVariant.DEFAULT
                  : BadgeVariant.SECONDARY,
                size: BadgeSize.SM,
              },
              req.provider.toUpperCase(),
            ),
          ]),

          // Model
          m(TableCell, { mono: true }, req.model),

          // Status
          m(TableCell, [
            m(
              Badge,
              {
                variant: req.status === "success"
                  ? BadgeVariant.SUCCESS
                  : BadgeVariant.DESTRUCTIVE,
                size: BadgeSize.SM,
              },
              t(`common.status.${req.status}`),
            ),
          ]),

          // Tokens
          m(
            TableCell,
            { mono: true },
            req.input_tokens !== null &&
              req.input_tokens !== undefined &&
              req.output_tokens !== null &&
              req.output_tokens !== undefined
              ? [
                m("span", req.input_tokens),
                m(
                  "span.ai-request-viewer__token-label",
                  ` ${t("aiRequestViewer.in")} `,
                ),
                m("span", req.output_tokens),
                m(
                  "span.ai-request-viewer__token-label",
                  ` ${t("aiRequestViewer.out")}`,
                ),
              ]
              : "-",
          ),

          // Duration
          m(TableCell, { mono: true }, `${req.duration_ms}ms`),

          // Time
          m(TableCell, formatUnixTimestamp(req.created_at, "time")),
        ],
      ),
    ];

    // Add expanded content row if expanded
    if (isExpanded) {
      const panels = [
        this.renderJSONPanel(t("aiRequestViewer.request"), req.request_json),
      ];
      if (req.response_json) {
        panels.push(
          this.renderJSONPanel(
            t("aiRequestViewer.response"),
            req.response_json,
          ),
        );
      }

      rows.push(
        m("tr.ai-request-viewer__expanded-row", { key: req.id + "-expanded" }, [
          m("td", { colspan: 7 }, [
            m(".ai-request-viewer__content", [
              // Error message if present
              req.error_message
                ? m(".ai-request-viewer__error", [
                  m("strong", t("aiRequestViewer.error") + ": "),
                  req.error_message,
                ])
                : null,

              // Endpoint
              m(".ai-request-viewer__endpoint", [
                m("strong", t("aiRequestViewer.endpoint") + ": "),
                m("code", req.endpoint),
              ]),

              // Request/Response panels
              m(
                ".ai-request-viewer__panels",
                {
                  class: panels.length === 1
                    ? "ai-request-viewer__panels--single"
                    : "",
                },
                panels,
              ),
            ]),
          ]),
        ]),
      );
    }

    return m.fragment({ key: req.id }, rows);
  },

  /**
   * Renders a JSON panel using CodeViewer.
   * @param {string} title - Panel title
   * @param {string} jsonStr - JSON string
   * @returns {Object} Mithril vnode
   */
  renderJSONPanel(title, jsonStr) {
    const { formatted } = this.formatJSON(jsonStr);

    return m(CodeViewer, {
      code: formatted,
      language: "json",
      title,
      maxHeight: "200px",
      padded: true,
    });
  },
};

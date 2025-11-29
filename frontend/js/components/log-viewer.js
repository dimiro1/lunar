/**
 * @fileoverview Log viewer component for displaying execution logs.
 */

import { t } from "../i18n/index.js";

/**
 * @typedef {Object} LogEntry
 * @property {string} [id] - Unique log entry ID
 * @property {string} [timestamp] - Formatted timestamp
 * @property {string} [level] - Log level (INFO, WARN, ERROR, DEBUG)
 * @property {string} message - Log message content
 */

/**
 * Log viewer component for displaying execution logs with color-coded levels.
 * @type {Object}
 */
export const LogViewer = {
  /**
   * Renders the log viewer component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {LogEntry[]} [vnode.attrs.logs=[]] - Array of log entries to display
   * @param {string} [vnode.attrs.maxHeight='300px'] - Maximum height with overflow scroll
   * @param {boolean} [vnode.attrs.noBorder=false] - Remove border styling
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { logs = [], maxHeight = "300px", noBorder = false } = vnode.attrs;

    return m(
      ".log-viewer",
      {
        class: noBorder ? "log-viewer--no-border" : "",
        style: maxHeight ? `max-height: ${maxHeight}` : "",
      },
      [
        logs.length === 0
          ? m(".log-viewer__empty", t("logViewer.noLogs"))
          : logs.map((log, i) =>
            m(
              ".log-viewer__entry",
              {
                key: log.id || i,
                class: i === logs.length - 1 ? "log-viewer__entry--last" : "",
              },
              [
                log.timestamp &&
                m("span.log-viewer__timestamp", log.timestamp),
                m(
                  "span.log-viewer__level",
                  {
                    class: `log-viewer__level--${
                      (log.level || "info").toLowerCase()
                    }`,
                  },
                  (log.level || "INFO").toUpperCase(),
                ),
                m("span.log-viewer__message", log.message),
              ],
            )
          ),
      ],
    );
  },
};

/**
 * @fileoverview Diff viewer component for displaying code differences.
 */

import { icons } from "../icons.js";
import { t } from "../i18n/index.js";

/**
 * Enum for diff line types.
 * @readonly
 * @enum {string}
 */
export const LineType = {
  /** Line was added in the new version */
  ADDED: "added",
  /** Line was removed from the old version */
  REMOVED: "removed",
  /** Line is unchanged between versions */
  UNCHANGED: "unchanged",
};

/**
 * @typedef {Object} DiffLine
 * @property {string} type - Line type (added, removed, unchanged)
 * @property {string} content - Line content
 * @property {number} oldLine - Line number in an old version (0 if added)
 * @property {number} newLine - Line number in a new version (0 if removed)
 */

/**
 * Version labels component showing diff statistics.
 * Displays additions and deletions count.
 * @type {Object}
 */
export const VersionLabels = {
  /**
   * Renders the version labels component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} vnode.attrs.oldLabel - Label for old version
   * @param {string} vnode.attrs.newLabel - Label for new version
   * @param {string} [vnode.attrs.oldMeta] - Metadata for old version
   * @param {string} [vnode.attrs.newMeta] - Metadata for new version
   * @param {number} vnode.attrs.additions - Number of added lines
   * @param {number} vnode.attrs.deletions - Number of deleted lines
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { additions, deletions } = vnode.attrs;

    return m(".diff-version-labels", [
      m("span.diff-stats-summary", [
        additions > 0 &&
        m("span.diff-stats-added", [
          m.trust(icons.plusSmall()),
          ` ${additions} ${
            t(additions !== 1 ? "diff.additions" : "diff.addition")
          } `,
        ]),
        deletions > 0 &&
        m("span.diff-stats-removed", [
          m.trust(icons.minusSmall()),
          ` ${deletions} ${
            t(deletions !== 1 ? "diff.deletions" : "diff.deletion")
          }`,
        ]),
      ]),
    ]);
  },
};

/**
 * Diff viewer component for displaying code differences.
 * Renders a table with old/new line numbers and change indicators.
 * @type {Object}
 */
export const DiffViewer = {
  /**
   * Renders the diff viewer component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {DiffLine[]} [vnode.attrs.lines=[]] - Array of diff lines to display
   * @param {string} [vnode.attrs.maxHeight] - Maximum height with overflow scroll
   * @param {boolean} [vnode.attrs.noBorder] - Remove border styling
   * @param {string} [vnode.attrs.language] - Language for syntax highlighting
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { lines = [], maxHeight, noBorder, language } = vnode.attrs;

    return m(
      ".diff-container",
      {
        class: noBorder ? "diff-container--no-border" : "",
        role: "region",
        "aria-label": t("diff.codeDiff"),
      },
      [
        m(
          ".diff-scroll",
          { style: maxHeight ? `max-height: ${maxHeight}` : "" },
          [
            m("table.diff-table", [
              m(
                "tbody",
                lines.map((line) => m(DiffLine, { line, language })),
              ),
            ]),
          ],
        ),
      ],
    );
  },
};

/**
 * Single diff line component.
 * Renders a table row with line numbers, type indicator, and content.
 * @type {Object}
 * @private
 */
const DiffLine = {
  /**
   * Renders a single diff line.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {DiffLine} vnode.attrs.line - The diff line data
   * @param {string} [vnode.attrs.language] - Language for syntax highlighting
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { line } = vnode.attrs;
    const lineClass = `diff-line--${line.type}`;

    return m("tr", { class: lineClass }, [
      m("td.diff-line-number", line.oldLine > 0 ? line.oldLine : ""),
      m("td.diff-line-number", line.newLine > 0 ? line.newLine : ""),
      m("td.diff-line-type", {
        class: `diff-type--${line.type}`,
        "aria-label": getTypeAriaLabel(line.type),
      }),
      m("td.diff-content", line.content || " "),
    ]);
  },
};

/**
 * Gets the ARIA label for a line type.
 * @param {string} type - Line type (added, removed, unchanged)
 * @returns {string} Human-readable description for screen readers
 */
function getTypeAriaLabel(type) {
  switch (type) {
    case LineType.ADDED:
      return t("diff.lineAdded");
    case LineType.REMOVED:
      return t("diff.lineRemoved");
    default:
      return t("diff.unchangedLine");
  }
}

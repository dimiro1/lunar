/**
 * @fileoverview Version diff view for comparing two function versions.
 */

import { icons } from "../icons.js";
import { API } from "../api.js";
import { BackButton } from "../components/button.js";
import { routes } from "../routes.js";
import { Card, CardContent, CardHeader } from "../components/card.js";
import {
  Badge,
  BadgeSize,
  BadgeVariant,
  IDBadge,
  StatusBadge,
} from "../components/badge.js";
import {
  DiffViewer,
  LineType,
  VersionLabels,
} from "../components/diff-viewer.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').LunarFunction} LunarFunction
 * @typedef {import('../types.js').DiffResponse} DiffResponse
 */

/**
 * Version diff view component.
 * Displays a side-by-side or unified diff between two function versions.
 * @type {Object}
 */
export const VersionDiff = {
  /**
   * Currently loaded function.
   * @type {LunarFunction|null}
   */
  func: null,

  /**
   * Diff data from the API.
   * @type {DiffResponse|null}
   */
  diffData: null,

  /**
   * Whether the view is loading.
   * @type {boolean}
   */
  loading: true,

  /**
   * Initializes the view and loads diff data.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Mithril vnode attributes
   * @param {string} vnode.attrs.id - Function ID
   * @param {number} vnode.attrs.v1 - First version number
   * @param {number} vnode.attrs.v2 - Second version number
   */
  oninit: (vnode) => {
    VersionDiff.loadData(vnode.attrs.id, vnode.attrs.v1, vnode.attrs.v2);
  },

  /**
   * Loads function and diff data.
   * @param {string} functionId - Function ID
   * @param {number} v1 - First version number
   * @param {number} v2 - Second version number
   * @returns {Promise<void>}
   */
  loadData: async (functionId, v1, v2) => {
    VersionDiff.loading = true;
    try {
      const [func, diffData] = await Promise.all([
        API.functions.get(functionId),
        API.versions.diff(functionId, v1, v2),
      ]);
      VersionDiff.func = func;
      VersionDiff.diffData = diffData;
    } catch (e) {
      console.error("Failed to load diff:", e);
    } finally {
      VersionDiff.loading = false;
      m.redraw();
    }
  },

  /**
   * Renders the version diff view.
   * @returns {Object} Mithril vnode
   */
  view: () => {
    if (VersionDiff.loading) {
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", t("versionDiff.loadingDiff")),
      ]);
    }

    if (!VersionDiff.func || !VersionDiff.diffData) {
      return m(
        ".fade-in",
        m(Card, m(CardContent, t("versionDiff.diffNotFound"))),
      );
    }

    const func = VersionDiff.func;

    return m(".fade-in", [
      // Header
      m(".function-details-header", [
        m(".function-details-left", [
          m(BackButton, { href: routes.functionVersions(func.id) }),
          m(".function-details-divider"),
          m(".function-details-info", [
            m("h1.function-details-title", [
              func.name,
              m(IDBadge, { id: func.id }),
              m(
                Badge,
                {
                  variant: BadgeVariant.OUTLINE,
                  size: BadgeSize.SM,
                  mono: true,
                },
                `v${VersionDiff.diffData.old_version} → v${VersionDiff.diffData.new_version}`,
              ),
            ]),
            m(
              "p.function-details-description",
              func.description || t("common.noDescription"),
            ),
          ]),
        ]),
        m(".function-details-actions", [
          m(StatusBadge, { enabled: !func.disabled, glow: true }),
        ]),
      ]),

      m(Card, [
        m(
          CardHeader,
          {
            title: t("versionDiff.codeChanges"),
          },
          [
            m(VersionLabels, {
              oldLabel: `v${VersionDiff.diffData.old_version}`,
              newLabel: `v${VersionDiff.diffData.new_version}`,
              additions: VersionDiff.diffData.diff.filter(
                (l) => l.line_type === "added",
              ).length,
              deletions: VersionDiff.diffData.diff.filter(
                (l) => l.line_type === "removed",
              ).length,
            }),
          ],
        ),
        m(CardContent, { noPadding: true }, [
          m(DiffViewer, {
            lines: VersionDiff.diffData.diff.map((line) => ({
              type: line.line_type === "added"
                ? LineType.ADDED
                : line.line_type === "removed"
                ? LineType.REMOVED
                : LineType.UNCHANGED,
              content: line.content,
              oldLine: line.old_line || 0,
              newLine: line.new_line || 0,
            })),
            maxHeight: "600px",
            noBorder: true,
          }),
        ]),
      ]),
    ]);
  },
};

/**
 * @fileoverview Function versions view with version history and comparison.
 */

import { icons } from "../icons.js";
import { API } from "../api.js";
import { Toast } from "../components/toast.js";
import { Pagination } from "../components/pagination.js";
import { formatUnixTimestamp, getFunctionTabs } from "../utils.js";
import { paths, routes } from "../routes.js";
import {
  BackButton,
  Button,
  ButtonSize,
  ButtonVariant,
} from "../components/button.js";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
} from "../components/card.js";
import {
  Badge,
  BadgeSize,
  BadgeVariant,
  IDBadge,
  StatusBadge,
} from "../components/badge.js";
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/table.js";
import { TabContent, Tabs } from "../components/tabs.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').LunarFunction} LunarFunction
 * @typedef {import('../types.js').FunctionVersion} FunctionVersion
 */

/**
 * Function versions view component.
 * Displays version history with the ability to activate versions and compare diffs.
 * @type {Object}
 */
export const FunctionVersions = {
  /**
   * Currently loaded function.
   * @type {LunarFunction|null}
   */
  func: null,

  /**
   * Array of loaded versions.
   * @type {FunctionVersion[]}
   */
  versions: [],

  /**
   * Whether the view is loading.
   * @type {boolean}
   */
  loading: true,

  /**
   * Number of versions per page.
   * @type {number}
   */
  versionsLimit: 20,

  /**
   * Current pagination offset.
   * @type {number}
   */
  versionsOffset: 0,

  /**
   * Total number of versions.
   * @type {number}
   */
  versionsTotal: 0,

  /**
   * Array of selected version numbers for comparison (max 2).
   * @type {number[]}
   */
  selectedVersions: [],

  /**
   * Initializes the view and loads data.
   * @param {Object} vnode - Mithril vnode
   */
  oninit: (vnode) => {
    FunctionVersions.selectedVersions = [];
    FunctionVersions.loadData(vnode.attrs.id);
  },

  /**
   * Loads function and versions data.
   * @param {string} id - Function ID
   * @returns {Promise<void>}
   */
  loadData: async (id) => {
    FunctionVersions.loading = true;
    try {
      const [func, versions] = await Promise.all([
        API.functions.get(id),
        API.versions.list(
          id,
          FunctionVersions.versionsLimit,
          FunctionVersions.versionsOffset,
        ),
      ]);
      FunctionVersions.func = func;
      FunctionVersions.versions = versions.versions || [];
      FunctionVersions.versionsTotal = versions.pagination?.total || 0;
    } catch (e) {
      console.error("Failed to load function:", e);
    } finally {
      FunctionVersions.loading = false;
      m.redraw();
    }
  },

  /**
   * Reloads versions with current pagination.
   * @returns {Promise<void>}
   */
  loadVersions: async () => {
    try {
      const versions = await API.versions.list(
        FunctionVersions.func.id,
        FunctionVersions.versionsLimit,
        FunctionVersions.versionsOffset,
      );
      FunctionVersions.versions = versions.versions || [];
      FunctionVersions.versionsTotal = versions.pagination?.total || 0;
      m.redraw();
    } catch (e) {
      console.error("Failed to load versions:", e);
    }
  },

  /**
   * Handles page change from pagination.
   * @param {number} newOffset - New pagination offset
   */
  handlePageChange: (newOffset) => {
    FunctionVersions.versionsOffset = newOffset;
    FunctionVersions.loadVersions();
  },

  /**
   * Handles limit change from pagination.
   * @param {number} newLimit - New items per page limit
   */
  handleLimitChange: (newLimit) => {
    FunctionVersions.versionsLimit = newLimit;
    FunctionVersions.versionsOffset = 0;
    FunctionVersions.loadVersions();
  },

  /**
   * Toggles version selection for comparison.
   * Maintains max 2 selections using FIFO.
   * @param {number} version - Version number to toggle
   */
  toggleVersionSelection: (version) => {
    const idx = FunctionVersions.selectedVersions.indexOf(version);
    if (idx === -1) {
      if (FunctionVersions.selectedVersions.length < 2) {
        FunctionVersions.selectedVersions.push(version);
      } else {
        FunctionVersions.selectedVersions.shift();
        FunctionVersions.selectedVersions.push(version);
      }
    } else {
      FunctionVersions.selectedVersions.splice(idx, 1);
    }
  },

  /**
   * Activates a specific version.
   * @param {number} version - Version number to activate
   * @returns {Promise<void>}
   */
  activateVersion: async (version) => {
    if (!confirm(t("versionsPage.activateConfirm", { version }))) return;
    try {
      await API.versions.activate(FunctionVersions.func.id, version);
      Toast.show(t("versionsPage.versionActivated", { version }), "success");
      await FunctionVersions.loadData(FunctionVersions.func.id);
    } catch (e) {
      Toast.show(t("versionsPage.failedToActivate"), "error");
    }
  },

  /**
   * Navigates to the version diff view for selected versions.
   */
  compareVersions: () => {
    if (FunctionVersions.selectedVersions.length !== 2) return;
    const [v1, v2] = FunctionVersions.selectedVersions.sort((a, b) => a - b);
    m.route.set(paths.functionDiff(FunctionVersions.func.id, v1, v2));
  },

  /**
   * Renders the function versions view.
   * @param {Object} _vnode - Mithril vnode
   * @returns {Object} Mithril vnode
   */
  view: (_vnode) => {
    if (FunctionVersions.loading) {
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", t("functions.loadingFunction")),
      ]);
    }

    if (!FunctionVersions.func) {
      return m(
        ".fade-in",
        m(Card, m(CardContent, t("common.functionNotFound"))),
      );
    }

    const func = FunctionVersions.func;

    return m(".fade-in", [
      // Header
      m(".function-details-header", [
        m(".function-details-left", [
          m(BackButton, { href: routes.functions() }),
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
                `v${func.active_version.version}`,
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

      // Tabs
      m(Tabs, {
        tabs: getFunctionTabs(func.id),
        activeTab: "versions",
      }),

      // Content
      m(TabContent, [
        m(".versions-tab-container", [
          // Version List
          m(Card, { style: "margin-bottom: 1.5rem" }, [
            m(CardHeader, {
              title: t("versions.title"),
              subtitle: t("versionsPage.versionsCount", {
                count: FunctionVersions.versionsTotal,
              }),
            }),
            FunctionVersions.versions.length === 0
              ? m(CardContent, [
                m(TableEmpty, {
                  icon: "inbox",
                  message: t("versions.emptyState"),
                }),
              ])
              : [
                m(Table, [
                  m(TableHeader, [
                    m(TableRow, [
                      m(TableHead, { style: "width: 40px" }, ""),
                      m(TableHead, t("versions.columns.version")),
                      m(TableHead, t("versions.columns.createdAt")),
                      m(TableHead, t("versions.columns.actions")),
                    ]),
                  ]),
                  m(
                    TableBody,
                    FunctionVersions.versions.map((ver) =>
                      m(
                        TableRow,
                        {
                          key: ver.version,
                          class: FunctionVersions.selectedVersions.includes(
                              ver.version,
                            )
                            ? "table__row--selected"
                            : "",
                        },
                        [
                          m(TableCell, [
                            m("input[type=checkbox]", {
                             ["aria-label"]: t("form.checkBox"),
                              checked: FunctionVersions.selectedVersions
                                .includes(
                                  ver.version,
                                ),
                              onchange: () =>
                                FunctionVersions.toggleVersionSelection(
                                  ver.version,
                                ),
                            }),
                          ]),
                          m(TableCell, [
                            m(
                              "span",
                              {
                                style:
                                  "font-family: var(--font-mono); margin-right: 0.5rem;",
                              },
                              `v${ver.version}`,
                            ),
                            ver.version === func.active_version.version &&
                            m(
                              Badge,
                              {
                                variant: BadgeVariant.SUCCESS,
                                size: BadgeSize.SM,
                              },
                              t("versionsPage.active"),
                            ),
                          ]),
                          m(TableCell, formatUnixTimestamp(ver.created_at)),
                          m(TableCell, { align: "right" }, [
                            ver.version !== func.active_version.version &&
                            m(
                              Button,
                              {
                                variant: ButtonVariant.OUTLINE,
                                size: ButtonSize.SM,
                                onclick: (e) => {
                                  e.stopPropagation();
                                  FunctionVersions.activateVersion(
                                    ver.version,
                                  );
                                },
                              },
                              t("versionsPage.activate"),
                            ),
                          ]),
                        ],
                      )
                    ),
                  ),
                ]),
                m(Pagination, {
                  total: FunctionVersions.versionsTotal,
                  limit: FunctionVersions.versionsLimit,
                  offset: FunctionVersions.versionsOffset,
                  onPageChange: FunctionVersions.handlePageChange,
                  onLimitChange: FunctionVersions.handleLimitChange,
                }),
              ],
            m(CardFooter, [
              m(
                Button,
                {
                  variant: ButtonVariant.PRIMARY,
                  onclick: FunctionVersions.compareVersions,
                  disabled: FunctionVersions.selectedVersions.length !== 2,
                },
                FunctionVersions.selectedVersions.length === 2
                  ? t("versionsPage.compareVersions", {
                    v1: FunctionVersions.selectedVersions[0],
                    v2: FunctionVersions.selectedVersions[1],
                  })
                  : t("versionsPage.selectToCompare"),
              ),
            ]),
          ]),
        ]),
      ]),
    ]);
  },
};

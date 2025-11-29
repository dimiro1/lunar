/**
 * @fileoverview Function executions view with execution history.
 */

import { icons } from "../icons.js";
import { API } from "../api.js";
import { Pagination } from "../components/pagination.js";
import { formatUnixTimestamp, getFunctionTabs } from "../utils.js";
import { paths, routes } from "../routes.js";
import { BackButton } from "../components/button.js";
import { Card, CardContent, CardHeader } from "../components/card.js";
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
 * @typedef {import('../types.js').LunarFunction} lunarFunction
 * @typedef {import('../types.js').Execution} Execution
 */

/**
 * Function executions view component.
 * Displays paginated execution history for a function.
 * @type {Object}
 */
export const FunctionExecutions = {
  /**
   * Currently loaded function.
   * @type {lunarFunction|null}
   */
  func: null,

  /**
   * Array of loaded executions.
   * @type {Execution[]}
   */
  executions: [],

  /**
   * Whether the view is loading.
   * @type {boolean}
   */
  loading: true,

  /**
   * Number of executions per page.
   * @type {number}
   */
  executionsLimit: 20,

  /**
   * Current pagination offset.
   * @type {number}
   */
  executionsOffset: 0,

  /**
   * Total number of executions.
   * @type {number}
   */
  executionsTotal: 0,

  /**
   * Initializes the view and loads data.
   * @param {Object} vnode - Mithril vnode
   */
  oninit: (vnode) => {
    FunctionExecutions.loadData(vnode.attrs.id);
  },

  /**
   * Loads function and executions data.
   * @param {string} id - Function ID
   * @returns {Promise<void>}
   */
  loadData: async (id) => {
    FunctionExecutions.loading = true;
    try {
      const [func, executions] = await Promise.all([
        API.functions.get(id),
        API.executions.list(
          id,
          FunctionExecutions.executionsLimit,
          FunctionExecutions.executionsOffset,
        ),
      ]);
      FunctionExecutions.func = func;
      FunctionExecutions.executions = executions.executions || [];
      FunctionExecutions.executionsTotal = executions.pagination?.total || 0;
    } catch (e) {
      console.error("Failed to load function:", e);
    } finally {
      FunctionExecutions.loading = false;
      m.redraw();
    }
  },

  /**
   * Reloads executions with current pagination.
   * @returns {Promise<void>}
   */
  loadExecutions: async () => {
    try {
      const executions = await API.executions.list(
        FunctionExecutions.func.id,
        FunctionExecutions.executionsLimit,
        FunctionExecutions.executionsOffset,
      );
      FunctionExecutions.executions = executions.executions || [];
      FunctionExecutions.executionsTotal = executions.pagination?.total || 0;
      m.redraw();
    } catch (e) {
      console.error("Failed to load executions:", e);
    }
  },

  /**
   * Handles page change from pagination.
   * @param {number} newOffset - New pagination offset
   */
  handlePageChange: (newOffset) => {
    FunctionExecutions.executionsOffset = newOffset;
    FunctionExecutions.loadExecutions();
  },

  /**
   * Handles limit change from pagination.
   * @param {number} newLimit - New items per page limit
   */
  handleLimitChange: (newLimit) => {
    FunctionExecutions.executionsLimit = newLimit;
    FunctionExecutions.executionsOffset = 0;
    FunctionExecutions.loadExecutions();
  },

  /**
   * Renders the function executions view.
   * @param {Object} _vnode - Mithril vnode
   * @returns {Object} Mithril vnode
   */
  view: (_vnode) => {
    if (FunctionExecutions.loading) {
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", t("functions.loadingFunction")),
      ]);
    }

    if (!FunctionExecutions.func) {
      return m(
        ".fade-in",
        m(Card, m(CardContent, t("common.functionNotFound"))),
      );
    }

    const func = FunctionExecutions.func;

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
        activeTab: "executions",
      }),

      // Content
      m(TabContent, [
        m(".executions-tab-container", [
          m(Card, [
            m(CardHeader, {
              title: t("executions.title"),
              subtitle: t("executions.totalCount", {
                count: FunctionExecutions.executionsTotal,
              }),
            }),
            FunctionExecutions.executions.length === 0
              ? m(CardContent, [
                m(TableEmpty, {
                  icon: "inbox",
                  message: t("executions.emptyState"),
                }),
              ])
              : [
                m(Table, [
                  m(TableHeader, [
                    m(TableRow, [
                      m(TableHead, t("executions.columns.id")),
                      m(TableHead, t("executions.columns.status")),
                      m(TableHead, t("executions.columns.duration")),
                      m(TableHead, t("executions.columns.time")),
                    ]),
                  ]),
                  m(
                    TableBody,
                    FunctionExecutions.executions.map((exec) =>
                      m(
                        TableRow,
                        {
                          key: exec.id,
                          onclick: () => m.route.set(paths.execution(exec.id)),
                        },
                        [
                          m(TableCell, m(IDBadge, { id: exec.id })),
                          m(
                            TableCell,
                            m(
                              Badge,
                              {
                                variant: exec.status === "success"
                                  ? BadgeVariant.SUCCESS
                                  : BadgeVariant.DESTRUCTIVE,
                                size: BadgeSize.SM,
                              },
                              t(`common.status.${exec.status}`),
                            ),
                          ),
                          m(
                            TableCell,
                            { mono: true },
                            exec.duration_ms
                              ? `${exec.duration_ms}ms`
                              : t("common.na"),
                          ),
                          m(TableCell, formatUnixTimestamp(exec.created_at)),
                        ],
                      )
                    ),
                  ),
                ]),
                m(Pagination, {
                  total: FunctionExecutions.executionsTotal,
                  limit: FunctionExecutions.executionsLimit,
                  offset: FunctionExecutions.executionsOffset,
                  onPageChange: FunctionExecutions.handlePageChange,
                  onLimitChange: FunctionExecutions.handleLimitChange,
                }),
              ],
          ]),
        ]),
      ]),
    ]);
  },
};

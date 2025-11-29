/**
 * @fileoverview Functions list view - displays all functions with pagination.
 */

import { icons } from "../icons.js";
import { API } from "../api.js";
import { t } from "../i18n/index.js";
import { Pagination } from "../components/pagination.js";
import { Button, ButtonVariant } from "../components/button.js";
import { Card, CardContent, CardHeader } from "../components/card.js";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/table.js";
import {
  Badge,
  BadgeSize,
  BadgeVariant,
  StatusBadge,
} from "../components/badge.js";

/**
 * @typedef {import('../types.js').LunarFunction} LunarFunction
 */

/**
 * Functions list a view component.
 * Displays a paginated table of all functions.
 * @type {Object}
 */
export const FunctionsList = {
  /**
   * Array of loaded functions.
   * @type {LunarFunction[]}
   */
  functions: [],

  /**
   * Whether the view is loading.
   * @type {boolean}
   */
  loading: true,

  /**
   * Number of functions per page.
   * @type {number}
   */
  limit: 20,

  /**
   * Current pagination offset.
   * @type {number}
   */
  offset: 0,

  /**
   * Total number of functions.
   * @type {number}
   */
  total: 0,

  /**
   * Initializes the view and loads functions.
   */
  oninit: () => {
    FunctionsList.loadFunctions();
  },

  /**
   * Loads functions from the API.
   * @returns {Promise<void>}
   */
  loadFunctions: async () => {
    FunctionsList.loading = true;
    try {
      const response = await API.functions.list(
        FunctionsList.limit,
        FunctionsList.offset,
      );
      FunctionsList.functions = response.functions || [];
      FunctionsList.total = response.pagination?.total || 0;
    } catch (e) {
      console.error("Failed to load functions:", e);
    } finally {
      FunctionsList.loading = false;
      m.redraw();
    }
  },

  /**
   * Handles page change from pagination.
   * @param {number} newOffset - New pagination offset
   */
  handlePageChange: (newOffset) => {
    FunctionsList.offset = newOffset;
    FunctionsList.loadFunctions();
  },

  /**
   * Handles limit change from pagination.
   * @param {number} newLimit - New items per page limit
   */
  handleLimitChange: (newLimit) => {
    FunctionsList.limit = newLimit;
    FunctionsList.offset = 0;
    FunctionsList.loadFunctions();
  },

  /**
   * Renders the functions list view.
   * @returns {Object} Mithril vnode
   */
  view: () => {
    if (FunctionsList.loading) {
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", t("functions.loadingFunctions")),
      ]);
    }

    return m(".fade-in", [
      m(".page-header", [
        m(".page-header__title", [
          m("div", [
            m("h1", t("functions.title")),
            m(".page-header__subtitle", t("functions.subtitle")),
          ]),
          m(
            Button,
            {
              variant: ButtonVariant.PRIMARY,
              href: "#!/functions/new",
              icon: "plus",
            },
            t("functions.newFunction"),
          ),
        ]),
      ]),

      m(Card, [
        m(CardHeader, {
          title: t("functions.allFunctions"),
          subtitle: t("functions.totalCount", { count: FunctionsList.total }),
        }),

        FunctionsList.functions.length === 0
          ? m(CardContent, [
            m(".table__empty", [
              m(".table__empty-icon", m.trust(icons.inbox())),
              m(
                "p.table__empty-message",
                t("functions.emptyState"),
              ),
            ]),
          ])
          : [
            m(Table, [
              m(TableHeader, [
                m(TableRow, [
                  m(TableHead, t("functions.columns.name")),
                  m(TableHead, t("functions.columns.description")),
                  m(TableHead, t("functions.columns.status")),
                  m(TableHead, t("functions.columns.version")),
                ]),
              ]),
              m(
                TableBody,
                FunctionsList.functions.map((func) =>
                  m(
                    TableRow,
                    {
                      key: func.id,
                      onclick: () => m.route.set(`/functions/${func.id}`),
                    },
                    [
                      m(TableCell, { mono: true }, func.name),
                      m(
                        TableCell,
                        func.description ||
                          m("span.text-muted", t("common.noDescription")),
                      ),
                      m(
                        TableCell,
                        m(StatusBadge, { enabled: !func.disabled }),
                      ),
                      m(
                        TableCell,
                        m(
                          Badge,
                          {
                            variant: BadgeVariant.SUCCESS,
                            size: BadgeSize.SM,
                            mono: true,
                          },
                          `v${func.active_version.version}`,
                        ),
                      ),
                    ],
                  )
                ),
              ),
            ]),
            m(Pagination, {
              total: FunctionsList.total,
              limit: FunctionsList.limit,
              offset: FunctionsList.offset,
              onPageChange: FunctionsList.handlePageChange,
              onLimitChange: FunctionsList.handleLimitChange,
            }),
          ],
      ]),
    ]);
  },
};

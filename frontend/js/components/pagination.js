/**
 * @fileoverview Pagination components for navigating through paginated data.
 */

import { t } from "../i18n/index.js";

/**
 * Full pagination component with info text and per-page selector.
 * @type {Object}
 */
export const Pagination = {
  /**
   * Renders the pagination component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {number} vnode.attrs.total - Total number of items
   * @param {number} vnode.attrs.limit - Items per page
   * @param {number} vnode.attrs.offset - Current offset (0-indexed)
   * @param {(newOffset: number) => void} vnode.attrs.onPageChange - Callback when page changes
   * @param {(newLimit: number) => void} vnode.attrs.onLimitChange - Callback when limit changes
   * @param {boolean} [vnode.attrs.showPerPage=true] - Show per-page selector
   * @returns {Object|null} Mithril vnode or null if no items
   */
  view: function (vnode) {
    const {
      total,
      limit,
      offset,
      onPageChange,
      onLimitChange,
      showPerPage = true,
    } = vnode.attrs;

    const currentPage = Math.floor(offset / limit) + 1;
    const totalPages = Math.ceil(total / limit);
    const start = offset + 1;
    const end = Math.min(offset + limit, total);

    if (total === 0) {
      return null;
    }

    const hasPrev = currentPage > 1;
    const hasNext = currentPage < totalPages;

    /** @type {number[]} */
    const perPageOptions = [10, 20, 50];

    return m(
      "nav.pagination",
      { role: "navigation", "aria-label": "Pagination" },
      [
        m(".pagination__info", [
          t("pagination.showing") + " ",
          m("span.pagination__highlight", start),
          " " + t("pagination.to") + " ",
          m("span.pagination__highlight", end),
          " " + t("pagination.of") + " ",
          m("span.pagination__highlight", total),
          " " + t("pagination.results"),
        ]),

        m(".pagination__controls", [
          showPerPage &&
          m(
            "select.pagination__select",
            {
              "aria-label": "Results per page",
              value: limit,
              onchange: (e) => onLimitChange(parseInt(e.target.value)),
            },
            perPageOptions.map((opt) =>
              m(
                "option",
                { value: opt, selected: opt === limit },
                t("pagination.perPage", { count: opt }),
              )
            ),
          ),

          m(".pagination__buttons", [
            m(
              "button.pagination__btn.pagination__btn--prev",
              {
                disabled: !hasPrev,
                onclick: () => hasPrev && onPageChange(offset - limit),
                "aria-label": "Go to previous page",
              },
              t("pagination.previous"),
            ),
            m(
              "button.pagination__btn",
              {
                disabled: !hasNext,
                onclick: () => hasNext && onPageChange(offset + limit),
                "aria-label": "Go to next page",
              },
              t("pagination.next"),
            ),
          ]),
        ]),
      ],
    );
  },
};

/**
 * Simple pagination without info text - just prev/next buttons.
 * @type {Object}
 */
export const SimplePagination = {
  /**
   * Renders simple prev/next pagination.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {boolean} vnode.attrs.hasPrev - Whether previous page exists
   * @param {boolean} vnode.attrs.hasNext - Whether next page exists
   * @param {() => void} vnode.attrs.onPrev - Callback for previous button
   * @param {() => void} vnode.attrs.onNext - Callback for next button
   * @returns {Object} Mithril vnode
   */
  view: function (vnode) {
    const { hasPrev, hasNext, onPrev, onNext } = vnode.attrs;

    return m(".pagination__controls", [
      m(".pagination__buttons", [
        m(
          "button.pagination__btn.pagination__btn--prev",
          {
            disabled: !hasPrev,
            onclick: () => hasPrev && onPrev(),
            "aria-label": "Go to previous page",
          },
          t("pagination.previous"),
        ),
        m(
          "button.pagination__btn",
          {
            disabled: !hasNext,
            onclick: () => hasNext && onNext(),
            "aria-label": "Go to next page",
          },
          t("pagination.next"),
        ),
      ]),
    ]);
  },
};

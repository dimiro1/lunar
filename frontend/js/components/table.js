/**
 * @fileoverview Table components for displaying tabular data.
 */

import { icons } from "../icons.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').IconName} IconName
 */

/**
 * Table wrapper component with responsive scrolling.
 * @type {Object}
 */
export const Table = {
  /**
   * Renders the table wrapper.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {boolean} [vnode.attrs.hoverable=true] - Enable row hover effects
   * @param {boolean} [vnode.attrs.striped=false] - Enable striped rows
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render (TableHeader, TableBody)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      hoverable = true,
      striped = false,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    return m(".table-wrapper", { class: className }, [
      m(".table-responsive", [
        m(
          "table.table",
          {
            "data-table-hoverable": hoverable || undefined,
            "data-table-striped": striped || undefined,
            ...attrs,
          },
          vnode.children,
        ),
      ]),
    ]);
  },
};

/**
 * Table Header component (thead).
 * @type {Object}
 */
export const TableHeader = {
  /**
   * Renders the table header.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render (TableRow)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { class: className = "", ...attrs } = vnode.attrs;
    return m(
      "thead.table__header",
      {
        class: className,
        ...attrs,
      },
      vnode.children,
    );
  },
};

/**
 * Table Body component (tbody).
 * @type {Object}
 */
export const TableBody = {
  /**
   * Renders the table body.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render (TableRow)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { class: className = "", ...attrs } = vnode.attrs;
    return m(
      "tbody.table__body",
      {
        class: className,
        ...attrs,
      },
      vnode.children,
    );
  },
};

/**
 * Table Row component (tr).
 * @type {Object}
 */
export const TableRow = {
  /**
   * Renders the table row.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {boolean} [vnode.attrs.selected=false] - Whether row is selected
   * @param {() => void} [vnode.attrs.onclick] - Click handler
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render (TableHead, TableCell)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      selected = false,
      onclick,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const classes = [
      "table__row",
      selected && "table__row--selected",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return m(
      "tr",
      {
        class: classes,
        onclick,
        "aria-selected": selected || undefined,
        ...attrs,
      },
      vnode.children,
    );
  },
};

/**
 * Table Head cell component (th).
 * @type {Object}
 */
export const TableHead = {
  /**
   * Renders the table head cell.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.width] - Column width (e.g., "200 px", "20%")
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { width, class: className = "", ...attrs } = vnode.attrs;

    return m(
      "th.table__head",
      {
        class: className,
        scope: "col",
        style: width ? { width } : undefined,
        ...attrs,
      },
      vnode.children,
    );
  },
};

/**
 * Table Cell component (td).
 * @type {Object}
 */
export const TableCell = {
  /**
   * Renders the table cell.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {boolean} [vnode.attrs.mono=false] - Use monospace font
   * @param {('left'|'center'|'right')} [vnode.attrs.align] - Text alignment
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      mono = false,
      align,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const classes = [
      "table__cell",
      mono && "table__cell--mono",
      align === "center" && "table__cell--center",
      align === "right" && "table__cell--right",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return m(
      "td",
      {
        class: classes,
        ...attrs,
      },
      vnode.children,
    );
  },
};

/**
 * Table Empty state component - displayed when the table has no data.
 * @type {Object}
 */
export const TableEmpty = {
  /**
   * Renders the empty table state.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {number} [vnode.attrs.colspan=1] - Number of columns to span
   * @param {string} [vnode.attrs.message='No data available'] - Message to display
   * @param {IconName} [vnode.attrs.icon='inbox'] - Icon to display
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      colspan = 1,
      message = t("table.noData"),
      icon = "inbox",
      class: className = "",
    } = vnode.attrs;

    return m("tr", [
      m(
        "td.table__empty",
        {
          colspan,
          class: className,
        },
        [
          icon && m(".table__empty-icon", m.trust(icons[icon]())),
          m("p.table__empty-message", message),
        ],
      ),
    ]);
  },
};

/**
 * @typedef {Object} ColumnDefinition
 * @property {string} name - Column header text
 * @property {string} [width] - Column width
 */

/**
 * Helper component to create a header row from column definitions.
 * @type {Object}
 */
export const TableHeaderRow = {
  /**
   * Renders a header row from column definitions.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {(string|ColumnDefinition)[]} [vnode.attrs.columns=[]] - Column definitions
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { columns = [] } = vnode.attrs;

    return m(
      TableRow,
      {},
      columns.map((col) => {
        const name = typeof col === "string" ? col : col.name;
        const width = typeof col === "object" ? col.width : undefined;
        return m(TableHead, { width }, name);
      }),
    );
  },
};

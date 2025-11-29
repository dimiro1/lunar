/**
 * @fileoverview Navigation bar components.
 */

import { icons } from "../icons.js";
import { Kbd } from "./kbd.js";
import { LanguageSelector } from "./language-selector.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').IconName} IconName
 */

/**
 * @typedef {Object} BreadcrumbItem
 * @property {string} label - Breadcrumb text
 * @property {string} [href] - Link URL (if not active)
 * @property {boolean} [active] - Whether this is the current page
 */

/**
 * Navbar container component.
 * @type {Object}
 */
export const Navbar = {
  /**
   * Renders the navbar container.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { class: className = "", ...attrs } = vnode.attrs;

    return m(
      "header.navbar",
      {
        class: className,
        ...attrs,
      },
      vnode.children,
    );
  },
};

/**
 * Navbar Section component for grouping navbar items.
 * @type {Object}
 */
export const NavbarSection = {
  /**
   * Renders the navbar section.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { class: className = "", ...attrs } = vnode.attrs;

    return m(
      "div.navbar__section",
      {
        class: className,
        ...attrs,
      },
      vnode.children,
    );
  },
};

/**
 * Navbar Brand component - logo and name.
 * @type {Object}
 */
export const NavbarBrand = {
  /**
   * Renders the navbar brand.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.name='Dashboard'] - Brand name
   * @param {string} [vnode.attrs.href='#!/'] - Brand link URL
   * @param {IconName} [vnode.attrs.icon='bolt'] - Brand icon name
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      name = "Dashboard",
      href = "#!/",
      icon = "bolt",
      class: className = "",
      ...attrs
    } = vnode.attrs;

    return m(
      "a.navbar__brand",
      {
        href,
        class: className,
        ...attrs,
      },
      [
        m(".navbar__brand-icon", m.trust(icons[icon]())),
        m("span.navbar__brand-name", name),
      ],
    );
  },
};

/**
 * Navbar Breadcrumb component.
 * @type {Object}
 */
export const NavbarBreadcrumb = {
  /**
   * Renders the navbar breadcrumb.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {BreadcrumbItem[]} [vnode.attrs.items=[]] - Breadcrumb items
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { items = [], class: className = "" } = vnode.attrs;

    return m(
      "nav.navbar__breadcrumb",
      {
        class: className,
        "aria-label": "Breadcrumb",
      },
      items
        .map((item, i) => [
          i > 0 &&
          m(
            "span.navbar__breadcrumb-separator",
            { "aria-hidden": "true" },
            "/",
          ),
          item.active
            ? m(
              "span.navbar__breadcrumb-current",
              { "aria-current": "page" },
              item.label,
            )
            : m("a.navbar__breadcrumb-link", { href: item.href }, item.label),
        ])
        .flat(),
    );
  },
};

/**
 * Navbar Search button component.
 * @type {Object}
 */
export const NavbarSearch = {
  /**
   * Renders the navbar search button.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.placeholder='Search'] - Placeholder text
   * @param {string} [vnode.attrs.shortcut='⌘K'] - Keyboard shortcut to display
   * @param {() => void} [vnode.attrs.onclick] - Click handler
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      placeholder = "Search",
      shortcut = "⌘K",
      onclick,
      class: className = "",
    } = vnode.attrs;

    return m(
      "button.navbar__search",
      {
        class: className,
        onclick,
      },
      [m("span", placeholder), shortcut && m(Kbd, { small: true }, shortcut)],
    );
  },
};

/**
 * Navbar Divider component - vertical separator.
 * @type {Object}
 */
export const NavbarDivider = {
  /**
   * Renders the navbar divider.
   * @returns {Object} Mithril vnode
   */
  view() {
    return m(".navbar__divider", { "aria-hidden": "true" });
  },
};

/**
 * Navbar Action component - button or link in navbar.
 * @type {Object}
 */
export const NavbarAction = {
  /**
   * Renders the navbar action.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.label] - Button/link text
   * @param {string} [vnode.attrs.href] - If provided, renders as anchor
   * @param {IconName} [vnode.attrs.icon] - Icon name
   * @param {() => void} [vnode.attrs.onclick] - Click handler (for button)
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      label,
      href,
      icon,
      onclick,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const content = [icon && m.trust(icons[icon]()), label];

    if (href) {
      return m(
        "a.navbar__action",
        {
          href,
          class: className,
          ...attrs,
        },
        content,
      );
    }

    return m(
      "button.navbar__action",
      {
        onclick,
        class: className,
        ...attrs,
      },
      content,
    );
  },
};

/**
 * Standard Header component - convenience wrapper for common navbar layout.
 * @type {Object}
 */
export const Header = {
  /**
   * Renders the standard header.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.breadcrumb] - Current page breadcrumb text
   * @param {() => void} [vnode.attrs.onLogout] - Logout button handler
   * @param {() => void} [vnode.attrs.onSearch] - Search button handler
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { breadcrumb, onLogout, onSearch } = vnode.attrs;

    return m(Navbar, [
      m(NavbarSection, [
        m(NavbarBrand, { name: t("nav.dashboard"), href: "#!/" }),
        breadcrumb &&
        m(
          "span.navbar__breadcrumb-separator",
          { "aria-hidden": "true" },
          "/",
        ),
        breadcrumb &&
        m(NavbarBreadcrumb, {
          items: [{ label: breadcrumb, active: true }],
        }),
      ]),
      m(NavbarSection, [
        m(NavbarSearch, {
          placeholder: t("nav.search"),
          shortcut: "⌘K",
          onclick: onSearch,
        }),
        m(NavbarDivider),
        m(LanguageSelector),
        m(NavbarDivider),
        m(NavbarAction, {
          label: t("nav.logout"),
          onclick: onLogout,
        }),
      ]),
    ]);
  },
};

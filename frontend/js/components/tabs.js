/**
 * @fileoverview Tab navigation components.
 */

import { icons } from "../icons.js";

/**
 * @typedef {import('../types.js').IconName} IconName
 */

/**
 * @typedef {Object} TabDefinition
 * @property {string} id - Unique tab identifier
 * @property {string} [label] - Tab display label (alias: name)
 * @property {string} [name] - Tab display name (alias: label)
 * @property {string} [href] - Link URL for navigation tabs
 * @property {boolean} [active] - Whether tab is active (for href-based tabs)
 * @property {boolean} [disabled] - Whether tab is disabled
 * @property {IconName} [icon] - Icon name to show before label
 * @property {string|number} [badge] - Badge text to show after label
 */

/**
 * Tabs component for navigation or content switching.
 * @type {Object}
 */
export const Tabs = {
  /**
   * Renders the tab component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {TabDefinition[]} [vnode.attrs.tabs=[]] - Array of tab definitions
   * @param {string} [vnode.attrs.activeTab] - Currently active tab ID (for click-based tabs)
   * @param {(tabId: string) => void} [vnode.attrs.onTabChange] - Callback when tab is clicked
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      tabs = [],
      activeTab,
      onTabChange,
      class: className = "",
    } = vnode.attrs;

    return m(
      "div.tabs",
      {
        class: className,
        role: "tablist",
      },
      tabs.map((tab) => {
        const tabId = tab.id;
        const tabLabel = tab.label || tab.name;
        const isActive = activeTab ? activeTab === tabId : tab.active;
        const isDisabled = tab.disabled;

        return m(
          "a.tabs__item",
          {
            href: tab.href || "#",
            class: [
              isActive && "tabs__item--active",
              isDisabled && "tabs__item--disabled",
            ]
              .filter(Boolean)
              .join(" "),
            role: "tab",
            "aria-selected": isActive ? "true" : "false",
            "aria-label": `tab-${tabLabel}`,
            "aria-disabled": isDisabled || undefined,
            tabindex: isDisabled ? -1 : undefined,
            "data-tab": "true",
            "data-tab-active": isActive || undefined,
            "data-tab-disabled": isDisabled || undefined,
            onclick: (e) => {
              if (isDisabled) {
                e.preventDefault();
                return;
              }
              if (onTabChange) {
                e.preventDefault();
                onTabChange(tabId);
              }
            },
          },
          [
            tab.icon && m("span.tabs__icon", m.trust(icons[tab.icon]())),
            tabLabel,
            tab.badge &&
            m(
              "span.tabs__badge",
              {
                class: isActive ? "tabs__badge--active" : "",
              },
              tab.badge,
            ),
          ],
        );
      }),
    );
  },
};

/**
 * Tab Content wrapper component for tab panels.
 * @type {Object}
 */
export const TabContent = {
  /**
   * Renders the tab content panel.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.id] - Tab panel ID (links to tab via aria-controls)
   * @param {boolean} [vnode.attrs.active=false] - Whether this panel is visible
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { id, active = false, class: className = "" } = vnode.attrs;

    // If no id provided, just render children (container mode)
    if (!id) {
      return m("div.tab-content", { class: className }, vnode.children);
    }

    return m(
      "div.tab-content",
      {
        id: `tab-${id}`,
        class: [className, active ? "tab-content--active" : ""]
          .filter(Boolean)
          .join(" "),
        role: "tabpanel",
        "aria-labelledby": `tab-${id}`,
        hidden: !active || undefined,
      },
      vnode.children,
    );
  },
};

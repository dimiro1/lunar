/**
 * @fileoverview Command palette component for quick navigation and search.
 */

import { icons } from "../icons.js";
import { API } from "../api.js";
import { paths } from "../routes.js";
import { i18n, localeNames, t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').LunarFunction} lunarFunction
 */

/**
 * @typedef {('nav'|'function'|'action'|'language'|'custom')} CommandItemType
 */

/**
 * @typedef {Object} CommandItem
 * @property {CommandItemType} type - Type of command item
 * @property {string} label - Display label
 * @property {string} [description] - Optional description
 * @property {string} [path] - Navigation path
 * @property {string} icon - Icon name from icons module
 * @property {string} [id] - Function ID (for function/action types)
 * @property {boolean} [disabled] - Whether the function is disabled
 * @property {string} [locale] - Locale code (for language type)
 * @property {string} [source] - Source identifier for custom items
 * @property {Function} [onSelect] - Custom action callback (for custom type)
 */

/**
 * Command palette component for quick navigation and function search.
 * Activated with Cmd+K (Mac) or Ctrl+K (Windows/Linux).
 * @type {Object}
 */
export const CommandPalette = {
  /**
   * Whether the palette is currently open.
   * @type {boolean}
   */
  isOpen: false,

  /**
   * Current search query.
   * @type {string}
   */
  query: "",

  /**
   * Filtered search results.
   * @type {CommandItem[]}
   */
  results: [],

  /**
   * Currently selected item index.
   * @type {number}
   */
  selectedIndex: 0,

  /**
   * All loaded functions.
   * @type {LunarFunction[]}
   */
  functions: [],

  /**
   * Whether functions are being loaded.
   * @type {boolean}
   */
  loading: false,

  /**
   * Registered custom command items from other components.
   * @type {CommandItem[]}
   */
  customItems: [],

  /**
   * Registers custom command items.
   * @param {string} source - Unique identifier for the source (e.g., 'function-code')
   * @param {CommandItem[]} items - Items to register
   */
  registerItems: (source, items) => {
    // Remove any existing items from this source first
    CommandPalette.customItems = CommandPalette.customItems.filter(
      (item) => item.source !== source,
    );
    // Add new items with source tag
    items.forEach((item) => {
      CommandPalette.customItems.push({ ...item, source });
    });
  },

  /**
   * Unregisters all custom command items from a source.
   * @param {string} source - Source identifier to remove
   */
  unregisterItems: (source) => {
    CommandPalette.customItems = CommandPalette.customItems.filter(
      (item) => item.source !== source,
    );
  },

  /**
   * Opens the command palette.
   */
  open: () => {
    CommandPalette.isOpen = true;
    CommandPalette.query = "";
    CommandPalette.selectedIndex = 0;
    CommandPalette.loadFunctions();
    m.redraw();
    // Focus input after render
    setTimeout(() => {
      const input = document.querySelector(".command-palette__input");
      if (input) input.focus();
    }, 10);
  },

  /**
   * Closes the command palette.
   */
  close: () => {
    CommandPalette.isOpen = false;
    CommandPalette.query = "";
    CommandPalette.results = [];
    m.redraw();
  },

  /**
   * Loads all functions from the API.
   * @returns {Promise<void>}
   */
  loadFunctions: async () => {
    CommandPalette.loading = true;
    try {
      const response = await API.functions.list(100, 0);
      CommandPalette.functions = response.functions || [];
      CommandPalette.updateResults();
    } catch (e) {
      console.error("Failed to load functions:", e);
      CommandPalette.functions = [];
    } finally {
      CommandPalette.loading = false;
      m.redraw();
    }
  },

  /**
   * Updates the filtered results based on the current search query.
   * Combines navigation items and function-related actions.
   */
  updateResults: () => {
    const q = CommandPalette.query.toLowerCase().trim();

    // Navigation items
    const navItems = [
      {
        type: "nav",
        label: t("functions.title"),
        description: t("commandPalette.actions.viewFunctions"),
        path: paths.functions(),
        icon: "bolt",
      },
      {
        type: "nav",
        label: t("functions.newFunction"),
        description: t("commandPalette.actions.createFunction"),
        path: paths.functionCreate(),
        icon: "plus",
      },
    ];

    // Function items with actions for each function
    const functionItems = [];
    CommandPalette.functions.forEach((func) => {
      // Main function entry - goes to code
      functionItems.push({
        type: "function",
        label: func.name,
        description: t("commandPalette.actions.goToCode"),
        path: paths.functionCode(func.id),
        icon: "code",
        id: func.id,
        disabled: func.disabled,
      });
      // Additional actions for each function
      functionItems.push({
        type: "action",
        label: `${func.name} → ${t("tabs.versions")}`,
        description: t("commandPalette.actions.viewVersions"),
        path: paths.functionVersions(func.id),
        icon: "listBullet",
        id: func.id,
        disabled: func.disabled,
      });
      functionItems.push({
        type: "action",
        label: `${func.name} → ${t("tabs.executions")}`,
        description: t("commandPalette.actions.viewExecutions"),
        path: paths.functionExecutions(func.id),
        icon: "chartBar",
        id: func.id,
        disabled: func.disabled,
      });
      functionItems.push({
        type: "action",
        label: `${func.name} → ${t("tabs.settings")}`,
        description: t("commandPalette.actions.configureFunction"),
        path: paths.functionSettings(func.id),
        icon: "cog",
        id: func.id,
        disabled: func.disabled,
      });
      functionItems.push({
        type: "action",
        label: `${func.name} → ${t("tabs.test")}`,
        description: t("commandPalette.actions.testFunction"),
        path: paths.functionTest(func.id),
        icon: "beaker",
        id: func.id,
        disabled: func.disabled,
      });
    });

    // Language items
    const currentLocale = i18n.getLocale();
    const languageItems = i18n.getAvailableLocales().map((locale) => ({
      type: "language",
      label: `${t("commandPalette.actions.switchLanguage")} → ${
        localeNames[locale]
      }`,
      description: locale === currentLocale
        ? t("commandPalette.currentLanguage")
        : "",
      icon: "globe",
      locale: locale,
    }));

    // Combine and filter (custom items first for visibility)
    const allItems = [
      ...CommandPalette.customItems,
      ...navItems,
      ...languageItems,
      ...functionItems,
    ];

    if (q) {
      CommandPalette.results = allItems.filter(
        (item) =>
          item.label.toLowerCase().includes(q) ||
          (item.description && item.description.toLowerCase().includes(q)),
      );
    } else {
      CommandPalette.results = allItems;
    }

    // Reset selection if out of bounds
    if (CommandPalette.selectedIndex >= CommandPalette.results.length) {
      CommandPalette.selectedIndex = Math.max(
        0,
        CommandPalette.results.length - 1,
      );
    }
  },

  /**
   * Scrolls the selected item into view.
   */
  scrollToSelected: () => {
    setTimeout(() => {
      const selected = document.querySelector(
        ".command-palette__item--selected",
      );
      if (selected) {
        selected.scrollIntoView({ block: "nearest" });
      }
    }, 0);
  },

  /**
   * Handles keyboard navigation and selection.
   * @param {KeyboardEvent} e - Keyboard event
   */
  handleKeyDown: (e) => {
    if (e.key === "ArrowDown" || (e.ctrlKey && e.key === "n")) {
      e.preventDefault();
      CommandPalette.selectedIndex = Math.min(
        CommandPalette.selectedIndex + 1,
        CommandPalette.results.length - 1,
      );
      m.redraw();
      CommandPalette.scrollToSelected();
    } else if (e.key === "ArrowUp" || (e.ctrlKey && e.key === "p")) {
      e.preventDefault();
      CommandPalette.selectedIndex = Math.max(
        CommandPalette.selectedIndex - 1,
        0,
      );
      m.redraw();
      CommandPalette.scrollToSelected();
    } else if (e.key === "Enter") {
      e.preventDefault();
      const selected = CommandPalette.results[CommandPalette.selectedIndex];
      if (selected) {
        CommandPalette.selectItem(selected);
      }
    } else if (e.key === "Escape") {
      CommandPalette.close();
    }
  },

  /**
   * Selects an item and navigates to its path or performs its action.
   * @param {CommandItem} item - The item to select
   */
  selectItem: (item) => {
    CommandPalette.close();
    if (item.type === "custom" && item.onSelect) {
      item.onSelect();
    } else if (item.type === "language") {
      i18n.setLocale(item.locale);
    } else {
      m.route.set(item.path);
    }
  },

  /**
   * Initializes the global keyboard listener for Cmd+K/Ctrl+K.
   */
  init: () => {
    document.addEventListener("keydown", (e) => {
      // Cmd+K or Ctrl+K
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        if (CommandPalette.isOpen) {
          CommandPalette.close();
        } else {
          CommandPalette.open();
        }
      }
    });
  },

  /**
   * Renders the command palette overlay and content.
   * @returns {Object|null} Mithril vnode or null if closed
   */
  view: () => {
    if (!CommandPalette.isOpen) return null;

    return m(
      ".command-palette-overlay",
      {
        onclick: (e) => {
          if (e.target.classList.contains("command-palette-overlay")) {
            CommandPalette.close();
          }
        },
      },
      [
        m(".command-palette", [
          m(".command-palette__header", [
            m(".command-palette__input-wrapper", [
              m("span.command-palette__search-icon", m.trust(icons.search())),
              m("input.command-palette__input", {
                type: "text",
                placeholder: t("commandPalette.searchPlaceholder"),
                value: CommandPalette.query,
                oninput: (e) => {
                  CommandPalette.query = e.target.value;
                  CommandPalette.selectedIndex = 0;
                  CommandPalette.updateResults();
                },
                onkeydown: CommandPalette.handleKeyDown,
              }),
            ]),
          ]),
          m(".command-palette__results", [
            CommandPalette.loading
              ? m(".command-palette__loading", [
                m.trust(icons.spinner()),
                " " + t("commandPalette.loading"),
              ])
              : CommandPalette.results.length === 0
              ? m(".command-palette__empty", t("commandPalette.noResults"))
              : CommandPalette.results.map((item, index) =>
                m(
                  ".command-palette__item",
                  {
                    class: [
                      index === CommandPalette.selectedIndex
                        ? "command-palette__item--selected"
                        : "",
                      item.disabled ? "command-palette__item--disabled" : "",
                    ]
                      .filter(Boolean)
                      .join(" "),
                    onclick: () => CommandPalette.selectItem(item),
                    onmouseenter: () => {
                      CommandPalette.selectedIndex = index;
                      m.redraw();
                    },
                  },
                  [
                    m(
                      "span.command-palette__item-icon",
                      m.trust(icons[item.icon]()),
                    ),
                    m(".command-palette__item-content", [
                      m("span.command-palette__item-label", item.label),
                      item.description &&
                      m(
                        "span.command-palette__item-description",
                        item.description,
                      ),
                    ]),
                    item.type === "function" &&
                    item.disabled &&
                    m("span.command-palette__item-badge", t("common.disabled")),
                  ],
                )
              ),
          ]),
          m(".command-palette__footer", [
            m("span.command-palette__hint", [
              m("kbd", "↑↓"),
              " " + t("commandPalette.toNavigate"),
            ]),
            m("span.command-palette__hint", [
              m("kbd", "↵"),
              " " + t("commandPalette.toSelect"),
            ]),
            m("span.command-palette__hint", [
              m("kbd", "esc"),
              " " + t("commandPalette.toClose"),
            ]),
          ]),
        ]),
      ],
    );
  },
};

// Initialize on load
CommandPalette.init();

/**
 * @fileoverview Language selector component for the navigation bar.
 */

import { icons } from "../icons.js";
import { i18n, localeNames } from "../i18n/index.js";

/**
 * Language selector dropdown component.
 * Allows users to switch between available languages.
 * @type {Object}
 */
export const LanguageSelector = {
  /**
   * Whether the dropdown is currently open.
   * @type {boolean}
   */
  isOpen: false,

  /**
   * Toggles the dropdown open/closed state.
   */
  toggle() {
    LanguageSelector.isOpen = !LanguageSelector.isOpen;
  },

  /**
   * Closes the dropdown.
   */
  close() {
    LanguageSelector.isOpen = false;
  },

  /**
   * Selects a language and closes the dropdown.
   * @param {string} locale - The locale code to select
   */
  selectLanguage(locale) {
    i18n.setLocale(locale);
    LanguageSelector.close();
  },

  /**
   * Handles click outside to close dropdown.
   * @param {Event} e - Click event
   */
  handleClickOutside(e) {
    if (!e.target.closest(".language-selector")) {
      LanguageSelector.close();
      m.redraw();
    }
  },

  /**
   * Sets up click outside listener when component is created.
   */
  oncreate() {
    document.addEventListener("click", LanguageSelector.handleClickOutside);
  },

  /**
   * Removes click outside listener when component is removed.
   */
  onremove() {
    document.removeEventListener("click", LanguageSelector.handleClickOutside);
  },

  /**
   * Renders the language selector component.
   * @returns {Object} Mithril vnode
   */
  view() {
    const currentLocale = i18n.getLocale();
    const availableLocales = i18n.getAvailableLocales();

    return m(".language-selector", [
      m(
        "button.language-selector__toggle",
        {
          onclick: (e) => {
            e.stopPropagation();
            LanguageSelector.toggle();
          },
          "aria-expanded": LanguageSelector.isOpen,
          "aria-haspopup": "listbox",
        },
        [
          m("span.language-selector__icon", m.trust(icons.globe())),
          m("span.language-selector__label", localeNames[currentLocale]),
          m("span.language-selector__chevron", m.trust(icons.chevronDown())),
        ],
      ),

      LanguageSelector.isOpen &&
      m(
        ".language-selector__dropdown",
        {
          role: "listbox",
          "aria-label": "Select language",
        },
        availableLocales.map((locale) =>
          m(
            "button.language-selector__option",
            {
              key: locale,
              class: locale === currentLocale
                ? "language-selector__option--active"
                : "",
              role: "option",
              "aria-selected": locale === currentLocale,
              onclick: () => LanguageSelector.selectLanguage(locale),
            },
            [
              m("span", localeNames[locale]),
              locale === currentLocale &&
              m(
                "span.language-selector__check",
                m.trust(icons.check()),
              ),
            ],
          )
        ),
      ),
    ]);
  },
};

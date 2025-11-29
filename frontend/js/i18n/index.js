/**
 * @fileoverview Core i18n module for localization support.
 */

import en from "./locales/en.js";
import ptBR from "./locales/pt-BR.js";

/**
 * Available locales with their translation dictionaries.
 */
const locales = {
  en: en,
  "pt-BR": ptBR,
};

/**
 * Locale display names for the language selector.
 */
export const localeNames = {
  en: "English",
  "pt-BR": "Português (Brasil)",
};

const DEFAULT_LOCALE = "en";
const STORAGE_KEY = "lunar-locale";

/**
 * i18n module - manages translations and current locale.
 */
export const i18n = {
  /**
   * Current active locale code.
   * @type {string}
   */
  locale: DEFAULT_LOCALE,

  /**
   * Initializes i18n from localStorage or browser defaults.
   */
  init() {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored && locales[stored]) {
      this.locale = stored;
    } else {
      // Auto-detect from browser language
      const browserLang = navigator.language;
      if (locales[browserLang]) {
        this.locale = browserLang;
      } else if (browserLang.startsWith("pt")) {
        this.locale = "pt-BR";
      }
    }
  },

  /**
   * Gets the current locale.
   * @returns {string} Current locale code
   */
  getLocale() {
    return this.locale;
  },

  /**
   * Sets the current locale and persists to localStorage.
   * @param {string} locale - Locale code to set
   */
  setLocale(locale) {
    if (locales[locale]) {
      this.locale = locale;
      localStorage.setItem(STORAGE_KEY, locale);
      m.redraw();
    }
  },

  /**
   * Gets all available locale codes.
   * @returns {string[]} Array of locale codes
   */
  getAvailableLocales() {
    return Object.keys(locales);
  },

  /**
   * Translates a key with optional interpolation.
   * @param {string} key - Translation key (dot notation supported)
   * @param {Object} [params] - Parameters for interpolation
   * @returns {string} Translated string
   *
   * @example
   * t('nav.logout') // "Logout"
   * t('functions.totalCount', { count: 100 })
   * // "100 functions total"
   */
  t(key, params = {}) {
    const translations = locales[this.locale] || locales[DEFAULT_LOCALE];

    // Support dot notation: 'nav.logout' -> translations.nav.logout
    let value = key.split(".").reduce((obj, k) => obj?.[k], translations);

    // Fallback to default locale if key not found
    if (value === undefined) {
      value = key
        .split(".")
        .reduce((obj, k) => obj?.[k], locales[DEFAULT_LOCALE]);
    }

    // Fallback to key itself if still not found
    if (value === undefined) {
      console.warn(`i18n: Missing translation for key "${key}"`);
      return key;
    }

    // Interpolation: replace {{param}} with values
    if (typeof value === "string" && Object.keys(params).length > 0) {
      return value.replace(/\{\{(\w+)\}\}/g, (match, param) => {
        return params[param] !== undefined ? params[param] : match;
      });
    }

    return value;
  },
};

/**
 * Shorthand translation function.
 * @param {string} key - Translation key
 * @param {Object} [params] - Parameters for interpolation
 * @returns {string} Translated string
 */
export const t = (key, params) => i18n.t(key, params);

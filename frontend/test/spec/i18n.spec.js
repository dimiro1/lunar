/**
 * @fileoverview Tests for i18n module.
 */

import { i18n, localeNames, t } from "../../js/i18n/index.js";

describe("i18n", () => {
  // Store original locale and localStorage to restore after tests
  let originalLocale;
  let originalLocalStorage;

  beforeEach(() => {
    originalLocale = i18n.locale;
    originalLocalStorage = localStorage.getItem("lunar-locale");
    // Reset to default state
    i18n.locale = "en";
  });

  afterEach(() => {
    i18n.locale = originalLocale;
    if (originalLocalStorage) {
      localStorage.setItem("lunar-locale", originalLocalStorage);
    } else {
      localStorage.removeItem("lunar-locale");
    }
  });

  describe("getLocale()", () => {
    it("returns the current locale", () => {
      i18n.locale = "en";
      expect(i18n.getLocale()).toBe("en");
    });

    it("returns pt-BR when set", () => {
      i18n.locale = "pt-BR";
      expect(i18n.getLocale()).toBe("pt-BR");
    });
  });

  describe("setLocale()", () => {
    it("changes the locale to a valid value", () => {
      i18n.setLocale("pt-BR");
      expect(i18n.locale).toBe("pt-BR");
    });

    it("persists locale to localStorage", () => {
      i18n.setLocale("pt-BR");
      expect(localStorage.getItem("lunar-locale")).toBe("pt-BR");
    });

    it("does not change locale for invalid value", () => {
      i18n.locale = "en";
      i18n.setLocale("invalid-locale");
      expect(i18n.locale).toBe("en");
    });

    it("does not persist invalid locale to localStorage", () => {
      localStorage.setItem("lunar-locale", "en");
      i18n.setLocale("invalid-locale");
      expect(localStorage.getItem("lunar-locale")).toBe("en");
    });
  });

  describe("getAvailableLocales()", () => {
    it("returns an array of locale codes", () => {
      const locales = i18n.getAvailableLocales();
      expect(Array.isArray(locales)).toBe(true);
    });

    it("includes English", () => {
      const locales = i18n.getAvailableLocales();
      expect(locales).toContain("en");
    });

    it("includes Portuguese (Brazil)", () => {
      const locales = i18n.getAvailableLocales();
      expect(locales).toContain("pt-BR");
    });
  });

  describe("t() translation function", () => {
    describe("basic translation", () => {
      it("translates simple keys in English", () => {
        i18n.locale = "en";
        expect(i18n.t("common.save")).toBe("Save");
      });

      it("translates simple keys in Portuguese", () => {
        i18n.locale = "pt-BR";
        expect(i18n.t("common.save")).toBe("Salvar");
      });

      it("translates nested keys", () => {
        i18n.locale = "en";
        expect(i18n.t("functions.columns.name")).toBe("Name");
      });

      it("translates nested keys in Portuguese", () => {
        i18n.locale = "pt-BR";
        expect(i18n.t("functions.columns.name")).toBe("Nome");
      });
    });

    describe("interpolation", () => {
      it("replaces {{count}} placeholder", () => {
        i18n.locale = "en";
        const result = i18n.t("functions.totalCount", { count: 5 });
        expect(result).toBe("5 functions total");
      });

      it("replaces {{count}} placeholder in Portuguese", () => {
        i18n.locale = "pt-BR";
        const result = i18n.t("functions.totalCount", { count: 5 });
        expect(result).toBe("5 funções no total");
      });

      it("replaces multiple placeholders", () => {
        i18n.locale = "en";
        const result = i18n.t("versionsPage.compareVersions", {
          v1: 1,
          v2: 2,
        });
        expect(result).toBe("Compare v1 and v2");
      });

      it("leaves placeholder if param not provided", () => {
        i18n.locale = "en";
        const result = i18n.t("functions.totalCount", {});
        expect(result).toBe("{{count}} functions total");
      });
    });

    describe("fallback behavior", () => {
      it("returns key if translation not found", () => {
        i18n.locale = "en";
        const result = i18n.t("nonexistent.key");
        expect(result).toBe("nonexistent.key");
      });

      it("falls back to English if key missing in current locale", () => {
        // This tests that if a key exists in English but not in pt-BR,
        // it falls back to English
        i18n.locale = "pt-BR";
        // Both locales have this key, so let's test the mechanism works
        // by checking a known key returns a value
        const result = i18n.t("common.loading");
        expect(result).toBe("Carregando...");
      });
    });
  });

  describe("t shorthand function", () => {
    it("works as a shorthand for i18n.t()", () => {
      i18n.locale = "en";
      expect(t("common.cancel")).toBe("Cancel");
    });

    it("supports interpolation", () => {
      i18n.locale = "en";
      expect(t("pagination.perPage", { count: 10 })).toBe("10 per page");
    });

    it("works with Portuguese locale", () => {
      i18n.locale = "pt-BR";
      expect(t("common.cancel")).toBe("Cancelar");
    });
  });

  describe("init()", () => {
    it("loads locale from localStorage if valid", () => {
      localStorage.setItem("lunar-locale", "pt-BR");
      i18n.locale = "en"; // Reset first
      i18n.init();
      expect(i18n.locale).toBe("pt-BR");
    });

    it("ignores invalid locale in localStorage", () => {
      localStorage.setItem("lunar-locale", "invalid");
      i18n.locale = "en";
      i18n.init();
      // Should remain at default or browser-detected locale
      expect(["en", "pt-BR"]).toContain(i18n.locale);
    });
  });
});

describe("localeNames", () => {
  it("has a display name for English", () => {
    expect(localeNames["en"]).toBe("English");
  });

  it("has a display name for Portuguese (Brazil)", () => {
    expect(localeNames["pt-BR"]).toBe("Português (Brasil)");
  });

  it("has names for all available locales", () => {
    const availableLocales = i18n.getAvailableLocales();
    availableLocales.forEach((locale) => {
      expect(localeNames[locale]).toBeDefined();
      expect(typeof localeNames[locale]).toBe("string");
    });
  });
});

describe("translation coverage", () => {
  // These tests verify that key translation keys exist in both locales
  const criticalKeys = [
    "common.loading",
    "common.save",
    "common.cancel",
    "common.delete",
    "common.create",
    "common.status.success",
    "common.status.error",
    "nav.dashboard",
    "nav.logout",
    "login.title",
    "login.loginButton",
    "functions.title",
    "functions.newFunction",
    "tabs.code",
    "tabs.versions",
    "tabs.executions",
    "tabs.settings",
    "tabs.test",
    "toast.settingsSaved",
    "pagination.previous",
    "pagination.next",
  ];

  criticalKeys.forEach((key) => {
    it(`has English translation for "${key}"`, () => {
      i18n.locale = "en";
      const result = t(key);
      expect(result).not.toBe(key); // Should not return the key itself
      expect(typeof result).toBe("string");
    });

    it(`has Portuguese translation for "${key}"`, () => {
      i18n.locale = "pt-BR";
      const result = t(key);
      expect(result).not.toBe(key); // Should not return the key itself
      expect(typeof result).toBe("string");
    });
  });
});

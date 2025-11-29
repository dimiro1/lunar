/**
 * @fileoverview Tests for LanguageSelector component.
 */

import { LanguageSelector } from "../../../js/components/language-selector.js";
import { i18n, localeNames } from "../../../js/i18n/index.js";

describe("LanguageSelector", () => {
  // Store original locale to restore after tests
  let originalLocale;

  beforeEach(() => {
    originalLocale = i18n.getLocale();
    LanguageSelector.isOpen = false;
  });

  afterEach(() => {
    i18n.setLocale(originalLocale);
    LanguageSelector.isOpen = false;
  });

  describe("toggle()", () => {
    it("opens dropdown when closed", () => {
      LanguageSelector.isOpen = false;
      LanguageSelector.toggle();
      expect(LanguageSelector.isOpen).toBe(true);
    });

    it("closes dropdown when open", () => {
      LanguageSelector.isOpen = true;
      LanguageSelector.toggle();
      expect(LanguageSelector.isOpen).toBe(false);
    });
  });

  describe("close()", () => {
    it("closes the dropdown", () => {
      LanguageSelector.isOpen = true;
      LanguageSelector.close();
      expect(LanguageSelector.isOpen).toBe(false);
    });
  });

  describe("selectLanguage()", () => {
    it("changes the locale", () => {
      i18n.setLocale("en");
      LanguageSelector.selectLanguage("pt-BR");
      expect(i18n.getLocale()).toBe("pt-BR");
    });

    it("closes the dropdown after selection", () => {
      LanguageSelector.isOpen = true;
      LanguageSelector.selectLanguage("en");
      expect(LanguageSelector.isOpen).toBe(false);
    });
  });

  describe("view()", () => {
    it("renders the language selector container", () => {
      const result = LanguageSelector.view();
      expect(result.tag).toBe("div");
      expect(result).toHaveClass("language-selector");
    });

    it("renders the toggle button", () => {
      const result = LanguageSelector.view();
      const toggleButton = result.children[0];

      expect(toggleButton.tag).toBe("button");
      expect(toggleButton).toHaveClass("language-selector__toggle");
    });

    it("displays current locale name on toggle button", () => {
      i18n.setLocale("en");
      const result = LanguageSelector.view();
      const toggleButton = result.children[0];

      // Find the label span
      const labelSpan = toggleButton.children.find(
        (child) =>
          child && getVnodeClass(child).includes("language-selector__label"),
      );
      expect(labelSpan).toBeTruthy();
      // Children can be a string directly or an array with text vnode
      const textContent = Array.isArray(labelSpan.children)
        ? labelSpan.children[0]?.children || labelSpan.children[0]
        : labelSpan.children;
      expect(textContent).toBe(localeNames["en"]);
    });

    it("displays Portuguese locale name when pt-BR is selected", () => {
      i18n.setLocale("pt-BR");
      const result = LanguageSelector.view();
      const toggleButton = result.children[0];

      const labelSpan = toggleButton.children.find(
        (child) =>
          child && getVnodeClass(child).includes("language-selector__label"),
      );
      // Children can be a string directly or an array with text vnode
      const textContent = Array.isArray(labelSpan.children)
        ? labelSpan.children[0]?.children || labelSpan.children[0]
        : labelSpan.children;
      expect(textContent).toBe(localeNames["pt-BR"]);
    });

    it("has correct aria attributes on toggle button", () => {
      LanguageSelector.isOpen = false;
      const result = LanguageSelector.view();
      const toggleButton = result.children[0];

      expect(toggleButton.attrs["aria-expanded"]).toBe(false);
      expect(toggleButton.attrs["aria-haspopup"]).toBe("listbox");
    });

    it("sets aria-expanded to true when dropdown is open", () => {
      LanguageSelector.isOpen = true;
      const result = LanguageSelector.view();
      const toggleButton = result.children[0];

      expect(toggleButton.attrs["aria-expanded"]).toBe(true);
    });

    it("does not render dropdown when closed", () => {
      LanguageSelector.isOpen = false;
      const result = LanguageSelector.view();

      // Second child should be false (dropdown not rendered)
      expect(result.children[1]).toBeFalsy();
    });

    it("renders dropdown when open", () => {
      LanguageSelector.isOpen = true;
      const result = LanguageSelector.view();
      const dropdown = result.children[1];

      expect(dropdown).toBeTruthy();
      expect(dropdown).toHaveClass("language-selector__dropdown");
      expect(dropdown.attrs.role).toBe("listbox");
    });

    it("renders an option for each available locale", () => {
      LanguageSelector.isOpen = true;
      const result = LanguageSelector.view();
      const dropdown = result.children[1];
      const availableLocales = i18n.getAvailableLocales();

      expect(dropdown.children.length).toBe(availableLocales.length);
    });

    it("marks current locale option as active", () => {
      i18n.setLocale("en");
      LanguageSelector.isOpen = true;
      const result = LanguageSelector.view();
      const dropdown = result.children[1];

      const activeOption = dropdown.children.find(
        (child) =>
          child && child.attrs && child.attrs["aria-selected"] === true,
      );
      expect(activeOption).toBeTruthy();
      expect(activeOption).toHaveClass("language-selector__option--active");
    });

    it("renders check icon for selected locale", () => {
      i18n.setLocale("en");
      LanguageSelector.isOpen = true;
      const result = LanguageSelector.view();
      const dropdown = result.children[1];

      const activeOption = dropdown.children.find(
        (child) =>
          child && child.attrs && child.attrs["aria-selected"] === true,
      );

      // The active option should have a check icon
      const hasCheckIcon = activeOption.children.some(
        (child) =>
          child && getVnodeClass(child).includes("language-selector__check"),
      );
      expect(hasCheckIcon).toBe(true);
    });

    it("does not render check icon for non-selected locales", () => {
      i18n.setLocale("en");
      LanguageSelector.isOpen = true;
      const result = LanguageSelector.view();
      const dropdown = result.children[1];

      const inactiveOption = dropdown.children.find(
        (child) =>
          child && child.attrs && child.attrs["aria-selected"] === false,
      );

      // The inactive option should not have a check icon (last child is false)
      const hasCheckIcon = inactiveOption.children.some(
        (child) =>
          child && getVnodeClass(child).includes("language-selector__check"),
      );
      expect(hasCheckIcon).toBe(false);
    });

    it("option buttons have correct role attribute", () => {
      LanguageSelector.isOpen = true;
      const result = LanguageSelector.view();
      const dropdown = result.children[1];

      dropdown.children.forEach((option) => {
        expect(option.attrs.role).toBe("option");
      });
    });

    it("renders globe icon in toggle button", () => {
      const result = LanguageSelector.view();
      const toggleButton = result.children[0];

      const iconSpan = toggleButton.children.find(
        (child) =>
          child && getVnodeClass(child).includes("language-selector__icon"),
      );
      expect(iconSpan).toBeTruthy();
    });

    it("renders chevron icon in toggle button", () => {
      const result = LanguageSelector.view();
      const toggleButton = result.children[0];

      const chevronSpan = toggleButton.children.find(
        (child) =>
          child && getVnodeClass(child).includes("language-selector__chevron"),
      );
      expect(chevronSpan).toBeTruthy();
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
});

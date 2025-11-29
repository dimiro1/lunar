/**
 * @fileoverview Tests for Form components - focused on critical functionality.
 */

import {
  CopyInput,
  FormCheckbox,
  FormInput,
  FormSelect,
  PasswordInput,
} from "../../../js/components/form.js";
import { t } from "../../../js/i18n/index.js";

/**
 * Helper to find a vnode in children by predicate
 */
function findChild(children, predicate) {
  if (!children) return null;
  const arr = Array.isArray(children) ? children : [children];
  for (const child of arr) {
    if (!child) continue;
    if (predicate(child)) return child;
    if (child.children) {
      const found = findChild(child.children, predicate);
      if (found) return found;
    }
  }
  return null;
}

describe("FormInput", () => {
  it("uses specified type", () => {
    const vnode = { attrs: { type: "email" }, children: [] };
    const result = FormInput.view(vnode);
    expect(result.attrs.type).toBe("email");
  });

  it("wraps with icon wrapper when icon is provided", () => {
    const vnode = { attrs: { icon: "search" }, children: [] };
    const result = FormInput.view(vnode);
    expect(result.tag).toBe("div");
  });
});

describe("FormSelect", () => {
  it("renders options from array", () => {
    const vnode = { attrs: { options: ["One", "Two", "Three"] }, children: [] };
    const result = FormSelect.view(vnode);
    expect(result.children.length).toBe(3);
  });

  it("marks selected option", () => {
    const vnode = {
      attrs: { options: ["A", "B", "C"], selected: "B" },
      children: [],
    };
    const result = FormSelect.view(vnode);
    expect(result.children[1].attrs.selected).toBe(true);
  });
});

describe("FormCheckbox", () => {
  it("sets checked attribute", () => {
    const vnode = { attrs: { checked: true }, children: [] };
    const result = FormCheckbox.view(vnode);
    const input = result.children.find((c) => c && c.tag === "input");
    expect(input.attrs.checked).toBe(true);
  });
});

describe("PasswordInput", () => {
  function createVnode(attrs = {}) {
    const vnode = { attrs, state: {}, children: [] };
    PasswordInput.oninit(vnode);
    return vnode;
  }

  it("initializes with password hidden", () => {
    const vnode = createVnode();
    const result = PasswordInput.view(vnode);
    const input = findChild(result.children, (c) => c && c.tag === "input");
    expect(input.attrs.type).toBe("password");
  });

  it("shows password when visibility toggled", () => {
    const vnode = createVnode();
    vnode.state.visible = true;
    const result = PasswordInput.view(vnode);
    const input = findChild(result.children, (c) => c && c.tag === "input");
    expect(input.attrs.type).toBe("text");
  });

  it("clicking toggle button toggles visibility", () => {
    const vnode = createVnode();
    expect(vnode.state.visible).toBe(false);

    const result = PasswordInput.view(vnode);
    const button = findChild(
      result.children,
      (c) =>
        c &&
        c.tag === "button" &&
        getVnodeClass(c).includes("form-password-toggle"),
    );

    button.attrs.onclick();
    expect(vnode.state.visible).toBe(true);

    button.attrs.onclick();
    expect(vnode.state.visible).toBe(false);
  });
});

describe("CopyInput", () => {
  function createVnode(attrs = {}) {
    const vnode = { attrs, state: {}, children: [] };
    CopyInput.oninit(vnode);
    return vnode;
  }

  it('shows "Copied!" feedback after copy', () => {
    const vnode = createVnode({ value: "test" });
    vnode.state.copied = true;
    const result = CopyInput.view(vnode);

    const button = findChild(
      result.children,
      (c) =>
        c &&
        c.tag === "button" &&
        getVnodeClass(c).includes("form-copy-button"),
    );
    expect(button.attrs.title).toBe(t("form.copied"));
  });

  it("clicking copy button calls clipboard API", async () => {
    const writeTextSpy = jasmine
      .createSpy("writeText")
      .and.returnValue(Promise.resolve());

    if (!navigator.clipboard) {
      Object.defineProperty(navigator, "clipboard", {
        value: { writeText: writeTextSpy },
        configurable: true,
      });
    } else {
      spyOn(navigator.clipboard, "writeText").and.returnValue(
        Promise.resolve(),
      );
    }

    const vnode = createVnode({ value: "copy-this-value" });
    const result = CopyInput.view(vnode);

    const button = findChild(
      result.children,
      (c) =>
        c &&
        c.tag === "button" &&
        getVnodeClass(c).includes("form-copy-button"),
    );

    await button.attrs.onclick();

    if (navigator.clipboard.writeText.calls) {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        "copy-this-value",
      );
    } else {
      expect(writeTextSpy).toHaveBeenCalledWith("copy-this-value");
    }
  });
});

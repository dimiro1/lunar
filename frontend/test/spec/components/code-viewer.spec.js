/**
 * @fileoverview Tests for CodeViewer component - focused on critical functionality.
 */

import { CodeViewer } from "../../../js/components/code-viewer.js";

describe("CodeViewer", () => {
  it("renders header when showHeader and language provided", () => {
    const vnode = {
      attrs: { showHeader: true, language: "lua" },
      children: [],
    };
    const result = CodeViewer.view(vnode);

    const header = result.children.find(
      (c) =>
        c && c.tag === "div" &&
        getVnodeClass(c).includes("code-viewer__header"),
    );
    expect(header).toBeTruthy();
  });

  it("does not render header when showHeader is false", () => {
    const vnode = {
      attrs: { showHeader: false, language: "lua" },
      children: [],
    };
    const result = CodeViewer.view(vnode);

    const header = result.children.find(
      (c) =>
        c && c.tag === "div" &&
        getVnodeClass(c).includes("code-viewer__header"),
    );
    expect(header).toBeFalsy();
  });

  it("applies language class to code element", () => {
    const vnode = { attrs: { language: "javascript" }, children: [] };
    const result = CodeViewer.view(vnode);

    const content = result.children.find(
      (c) =>
        c && c.tag === "div" &&
        getVnodeClass(c).includes("code-viewer__content"),
    );
    const pre = content.children.find((c) => c && c.tag === "pre");
    const code = pre.children.find((c) => c && c.tag === "code");
    expect(getVnodeClass(code)).toContain("language-javascript");
  });

  it("applies wrap class when wrap is true", () => {
    const vnode = { attrs: { wrap: true }, children: [] };
    const result = CodeViewer.view(vnode);

    const content = result.children.find(
      (c) =>
        c && c.tag === "div" &&
        getVnodeClass(c).includes("code-viewer__content"),
    );
    const pre = content.children.find((c) => c && c.tag === "pre");
    expect(getVnodeClass(pre)).toContain("code-viewer__pre--wrap");
  });

  it("does not apply wrap class when wrap is false", () => {
    const vnode = { attrs: { wrap: false }, children: [] };
    const result = CodeViewer.view(vnode);

    const content = result.children.find(
      (c) =>
        c && c.tag === "div" &&
        getVnodeClass(c).includes("code-viewer__content"),
    );
    const pre = content.children.find((c) => c && c.tag === "pre");
    expect(getVnodeClass(pre)).not.toContain("code-viewer__pre--wrap");
  });

  it("applies both padded and wrap classes when both are true", () => {
    const vnode = { attrs: { padded: true, wrap: true }, children: [] };
    const result = CodeViewer.view(vnode);

    const content = result.children.find(
      (c) =>
        c && c.tag === "div" &&
        getVnodeClass(c).includes("code-viewer__content"),
    );
    const pre = content.children.find((c) => c && c.tag === "pre");
    expect(getVnodeClass(pre)).toContain("code-viewer__pre--padded");
    expect(getVnodeClass(pre)).toContain("code-viewer__pre--wrap");
  });
});

/**
 * @fileoverview Tests for HtmlPreview component.
 */

import {
  HtmlPreview,
  HtmlPreviewMode,
} from "../../../js/components/html-preview.js";
import { Card, CardContent, CardHeader } from "../../../js/components/card.js";
import { CodeViewer } from "../../../js/components/code-viewer.js";
import { CommandPalette } from "../../../js/components/command-palette.js";

describe("HtmlPreviewMode", () => {
  it("exports CODE mode", () => {
    expect(HtmlPreviewMode.CODE).toBe("code");
  });

  it("exports PREVIEW mode", () => {
    expect(HtmlPreviewMode.PREVIEW).toBe("preview");
  });
});

describe("HtmlPreview", () => {
  function createVnode(attrs = {}) {
    return {
      attrs: { html: "<h1>Hello</h1>", ...attrs },
      state: {},
    };
  }

  describe("oninit()", () => {
    it("initializes mode to preview by default", () => {
      const vnode = createVnode();
      HtmlPreview.oninit(vnode);

      expect(vnode.state.mode).toBe(HtmlPreviewMode.PREVIEW);
    });

    it("initializes mode to code when defaultMode is code", () => {
      const vnode = createVnode({ defaultMode: HtmlPreviewMode.CODE });
      HtmlPreview.oninit(vnode);

      expect(vnode.state.mode).toBe(HtmlPreviewMode.CODE);
    });

    it("initializes mode to preview when defaultMode is preview", () => {
      const vnode = createVnode({ defaultMode: HtmlPreviewMode.PREVIEW });
      HtmlPreview.oninit(vnode);

      expect(vnode.state.mode).toBe(HtmlPreviewMode.PREVIEW);
    });
  });

  describe("view()", () => {
    it("renders Card component", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      expect(result.tag).toBe(Card);
    });

    it("renders CardHeader with title", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const header = result.children.find((c) => c && c.tag === CardHeader);
      expect(header).toBeTruthy();
      expect(header.attrs.title).toBeTruthy();
    });

    it("renders CardContent", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      expect(content).toBeTruthy();
      expect(content.attrs.noPadding).toBe(true);
    });

    it("applies custom style to Card", () => {
      const vnode = createVnode({ style: "margin-bottom: 1rem" });
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      expect(result.attrs.style).toBe("margin-bottom: 1rem");
    });
  });

  describe("preview mode", () => {
    it("renders iframe in preview mode", () => {
      const vnode = createVnode({ html: "<p>Test</p>" });
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const iframe = content.children.find((c) => c && c.tag === "iframe");
      expect(iframe).toBeTruthy();
    });

    it("sets iframe srcdoc to html content", () => {
      const html = "<div>My HTML</div>";
      const vnode = createVnode({ html });
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const iframe = content.children.find((c) => c && c.tag === "iframe");
      expect(iframe.attrs.srcdoc).toBe(html);
    });

    it("sets iframe sandbox attribute", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const iframe = content.children.find((c) => c && c.tag === "iframe");
      expect(iframe.attrs.sandbox).toBe("allow-same-origin");
    });

    it("applies maxHeight to iframe style", () => {
      const vnode = createVnode({ maxHeight: "500px" });
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const iframe = content.children.find((c) => c && c.tag === "iframe");
      expect(iframe.attrs.style).toContain("500px");
    });

    it("uses default maxHeight of 400px", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const iframe = content.children.find((c) => c && c.tag === "iframe");
      expect(iframe.attrs.style).toContain("400px");
    });
  });

  describe("code mode", () => {
    it("renders CodeViewer in code mode", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.CODE;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const codeViewer = content.children.find(
        (c) => c && c.tag === CodeViewer,
      );
      expect(codeViewer).toBeTruthy();
    });

    it("passes html as code to CodeViewer", () => {
      const html = "<section>Content</section>";
      const vnode = createVnode({ html });
      vnode.state.mode = HtmlPreviewMode.CODE;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const codeViewer = content.children.find(
        (c) => c && c.tag === CodeViewer,
      );
      expect(codeViewer.attrs.code).toBe(html);
    });

    it("sets language to html in CodeViewer", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.CODE;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const codeViewer = content.children.find(
        (c) => c && c.tag === CodeViewer,
      );
      expect(codeViewer.attrs.language).toBe("html");
    });

    it("passes maxHeight to CodeViewer", () => {
      const vnode = createVnode({ maxHeight: "300px" });
      vnode.state.mode = HtmlPreviewMode.CODE;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const codeViewer = content.children.find(
        (c) => c && c.tag === CodeViewer,
      );
      expect(codeViewer.attrs.maxHeight).toBe("300px");
    });

    it("sets noBorder to true on CodeViewer", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.CODE;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const codeViewer = content.children.find(
        (c) => c && c.tag === CodeViewer,
      );
      expect(codeViewer.attrs.noBorder).toBe(true);
    });

    it("sets padded to true on CodeViewer", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.CODE;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const codeViewer = content.children.find(
        (c) => c && c.tag === CodeViewer,
      );
      expect(codeViewer.attrs.padded).toBe(true);
    });
  });

  describe("toggle button", () => {
    it("renders toggle button in header actions", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const header = result.children.find((c) => c && c.tag === CardHeader);
      expect(header.attrs.actions).toBeTruthy();
      expect(header.attrs.actions.length).toBeGreaterThan(0);
    });

    it("toggle button has html-preview__toggle class", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const header = result.children.find((c) => c && c.tag === CardHeader);
      const toggleBtn = header.attrs.actions.find(
        (a) => a && a.tag === "button",
      );
      expect(toggleBtn).toBeTruthy();
      expect(getVnodeClass(toggleBtn)).toContain("html-preview__toggle");
    });

    it("clicking toggle button switches from preview to code mode", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const header = result.children.find((c) => c && c.tag === CardHeader);
      const toggleBtn = header.attrs.actions.find(
        (a) => a && a.tag === "button",
      );

      toggleBtn.attrs.onclick();
      expect(vnode.state.mode).toBe(HtmlPreviewMode.CODE);
    });

    it("clicking toggle button switches from code to preview mode", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.CODE;
      const result = HtmlPreview.view(vnode);

      const header = result.children.find((c) => c && c.tag === CardHeader);
      const toggleBtn = header.attrs.actions.find(
        (a) => a && a.tag === "button",
      );

      toggleBtn.attrs.onclick();
      expect(vnode.state.mode).toBe(HtmlPreviewMode.PREVIEW);
    });
  });

  describe("custom title", () => {
    it("uses custom title when provided", () => {
      const vnode = createVnode({ title: "Custom Title" });
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const header = result.children.find((c) => c && c.tag === CardHeader);
      expect(header.attrs.title).toBe("Custom Title");
    });

    it("uses custom title in both modes", () => {
      const vnode = createVnode({ title: "My HTML" });

      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      let result = HtmlPreview.view(vnode);
      let header = result.children.find((c) => c && c.tag === CardHeader);
      expect(header.attrs.title).toBe("My HTML");

      vnode.state.mode = HtmlPreviewMode.CODE;
      result = HtmlPreview.view(vnode);
      header = result.children.find((c) => c && c.tag === CardHeader);
      expect(header.attrs.title).toBe("My HTML");
    });
  });

  describe("empty html", () => {
    it("handles empty html string", () => {
      const vnode = createVnode({ html: "" });
      vnode.state.mode = HtmlPreviewMode.PREVIEW;

      expect(() => HtmlPreview.view(vnode)).not.toThrow();
    });

    it("renders iframe with empty srcdoc when html is empty", () => {
      const vnode = createVnode({ html: "" });
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const iframe = content.children.find((c) => c && c.tag === "iframe");
      expect(iframe.attrs.srcdoc).toBe("");
    });

    it("renders CodeViewer with empty code when html is empty", () => {
      const vnode = createVnode({ html: "" });
      vnode.state.mode = HtmlPreviewMode.CODE;
      const result = HtmlPreview.view(vnode);

      const content = result.children.find((c) => c && c.tag === CardContent);
      const codeViewer = content.children.find(
        (c) => c && c.tag === CodeViewer,
      );
      expect(codeViewer.attrs.code).toBe("");
    });
  });

  describe("command palette integration", () => {
    let registerSpy;
    let unregisterSpy;

    beforeEach(() => {
      registerSpy = spyOn(CommandPalette, "registerItems");
      unregisterSpy = spyOn(CommandPalette, "unregisterItems");
    });

    it("generates unique instance ID on oninit", () => {
      const vnode1 = createVnode();
      const vnode2 = createVnode();
      HtmlPreview.oninit(vnode1);
      HtmlPreview.oninit(vnode2);

      expect(vnode1.state.instanceId).toBeDefined();
      expect(vnode2.state.instanceId).toBeDefined();
      expect(vnode1.state.instanceId).not.toBe(vnode2.state.instanceId);
    });

    it("registers palette commands on oninit", () => {
      const vnode = createVnode();
      HtmlPreview.oninit(vnode);

      expect(registerSpy).toHaveBeenCalled();
      const [source, items] = registerSpy.calls.mostRecent().args;
      expect(source).toBe(vnode.state.instanceId);
      expect(items.length).toBe(1);
    });

    it("registers 'switch to code' command in preview mode", () => {
      const vnode = createVnode({ defaultMode: HtmlPreviewMode.PREVIEW });
      HtmlPreview.oninit(vnode);

      const [, items] = registerSpy.calls.mostRecent().args;
      expect(items[0].icon).toBe("code");
      expect(items[0].type).toBe("custom");
      expect(typeof items[0].onSelect).toBe("function");
    });

    it("registers 'switch to preview' command in code mode", () => {
      const vnode = createVnode({ defaultMode: HtmlPreviewMode.CODE });
      HtmlPreview.oninit(vnode);

      const [, items] = registerSpy.calls.mostRecent().args;
      expect(items[0].icon).toBe("eye");
      expect(items[0].type).toBe("custom");
      expect(typeof items[0].onSelect).toBe("function");
    });

    it("unregisters palette commands on onremove", () => {
      const vnode = createVnode();
      HtmlPreview.oninit(vnode);
      HtmlPreview.onremove(vnode);

      expect(unregisterSpy).toHaveBeenCalledWith(vnode.state.instanceId);
    });

    it("re-registers commands when mode is toggled via command palette", () => {
      const vnode = createVnode({ defaultMode: HtmlPreviewMode.PREVIEW });
      HtmlPreview.oninit(vnode);

      // Get the onSelect callback
      const [, items] = registerSpy.calls.mostRecent().args;
      const onSelect = items[0].onSelect;

      // Simulate command palette selection
      registerSpy.calls.reset();
      onSelect();

      expect(vnode.state.mode).toBe(HtmlPreviewMode.CODE);
      expect(registerSpy).toHaveBeenCalled();
    });

    it("re-registers commands when mode is toggled via button", () => {
      const vnode = createVnode();
      vnode.state.mode = HtmlPreviewMode.PREVIEW;
      vnode.state.instanceId = "test-instance";
      const result = HtmlPreview.view(vnode);

      // Find and click the toggle button
      const header = result.children.find((c) => c && c.tag === CardHeader);
      const toggleBtn = header.attrs.actions.find(
        (a) => a && a.tag === "button",
      );

      registerSpy.calls.reset();
      toggleBtn.attrs.onclick();

      expect(vnode.state.mode).toBe(HtmlPreviewMode.CODE);
      expect(registerSpy).toHaveBeenCalled();
    });
  });
});

/**
 * @fileoverview HTML preview component with toggle between code view and rendered preview.
 */

import { icons } from "../icons.js";
import { CodeViewer } from "./code-viewer.js";
import { Card, CardContent, CardHeader } from "./card.js";
import { t } from "../i18n/index.js";
import { CommandPalette } from "./command-palette.js";

/**
 * Counter for generating unique component instance IDs.
 * @type {number}
 */
let instanceCounter = 0;

/**
 * View modes for the HTML preview component.
 * @enum {string}
 */
export const HtmlPreviewMode = {
  CODE: "code",
  PREVIEW: "preview",
};

/**
 * HTML preview component that can toggle between code view and iframe preview.
 * Uses Card components for consistent styling.
 * @type {Object}
 */
export const HtmlPreview = {
  /**
   * Renders the HTML preview component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} vnode.attrs.html - HTML content to display
   * @param {string} [vnode.attrs.title] - Card header title
   * @param {string} [vnode.attrs.maxHeight='400px'] - Maximum height for the content area
   * @param {string} [vnode.attrs.defaultMode='preview'] - Initial view mode ('code' or 'preview')
   * @param {string} [vnode.attrs.style] - Additional inline styles for the card
   * @returns {Object} Mithril vnode
   */
  oninit(vnode) {
    const { defaultMode = HtmlPreviewMode.PREVIEW } = vnode.attrs;
    vnode.state.mode = defaultMode;
    vnode.state.instanceId = `html-preview-${++instanceCounter}`;
    HtmlPreview.registerPaletteCommands(vnode);
  },

  onremove(vnode) {
    CommandPalette.unregisterItems(vnode.state.instanceId);
  },

  /**
   * Registers commands with the command palette.
   * @param {Object} vnode - Mithril vnode
   */
  registerPaletteCommands(vnode) {
    const { mode, instanceId } = vnode.state;
    const isCodeMode = mode === HtmlPreviewMode.CODE;
    const items = [];

    if (isCodeMode) {
      items.push({
        type: "custom",
        label: t("execution.switchToPreview"),
        description: t("execution.switchToPreviewDesc"),
        icon: "eye",
        onSelect: () => {
          vnode.state.mode = HtmlPreviewMode.PREVIEW;
          HtmlPreview.registerPaletteCommands(vnode);
          m.redraw();
        },
      });
    } else {
      items.push({
        type: "custom",
        label: t("execution.switchToCode"),
        description: t("execution.switchToCodeDesc"),
        icon: "code",
        onSelect: () => {
          vnode.state.mode = HtmlPreviewMode.CODE;
          HtmlPreview.registerPaletteCommands(vnode);
          m.redraw();
        },
      });
    }

    CommandPalette.registerItems(instanceId, items);
  },

  view(vnode) {
    const {
      html = "",
      title,
      maxHeight = "400px",
      style = "",
    } = vnode.attrs;

    const { mode } = vnode.state;
    const isCodeMode = mode === HtmlPreviewMode.CODE;

    const toggleMode = () => {
      vnode.state.mode = isCodeMode
        ? HtmlPreviewMode.PREVIEW
        : HtmlPreviewMode.CODE;
      HtmlPreview.registerPaletteCommands(vnode);
    };

    // Toggle button for header actions
    const toggleButton = m(
      "button.html-preview__toggle",
      {
        onclick: toggleMode,
        title: isCodeMode
          ? t("execution.showPreview")
          : t("execution.showCode"),
      },
      m.trust(isCodeMode ? icons.eye() : icons.code()),
    );

    return m(Card, { style }, [
      m(CardHeader, {
        title: title ||
          (isCodeMode
            ? t("execution.responseBody")
            : t("execution.htmlPreview")),
        actions: [toggleButton],
      }),
      m(CardContent, { noPadding: true }, [
        isCodeMode
          ? m(CodeViewer, {
            code: html,
            language: "html",
            maxHeight,
            noBorder: true,
            padded: true,
          })
          : m("iframe.html-preview__iframe", {
            srcdoc: html,
            style:
              `width: 100%; height: ${maxHeight}; border: none; background: white;`,
            sandbox: "allow-same-origin",
          }),
      ]),
    ]);
  },
};

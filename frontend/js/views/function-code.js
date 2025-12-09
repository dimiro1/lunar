/**
 * @fileoverview Function code editor view with Monaco editor and API reference.
 */

import { icons } from "../icons.js";
import { API } from "../api.js";
import { Toast } from "../components/toast.js";
import { CodeEditor } from "../components/code-editor.js";
import {
  BackButton,
  Button,
  ButtonSize,
  ButtonVariant,
} from "../components/button.js";
import { Card, CardContent, MaximizableCard } from "../components/card.js";
import {
  Badge,
  BadgeSize,
  BadgeVariant,
  IDBadge,
  StatusBadge,
} from "../components/badge.js";
import { TabContent, Tabs } from "../components/tabs.js";
import { getFunctionTabs } from "../utils.js";
import { routes } from "../routes.js";
import {
  APIReference,
  getLuaAPISections,
} from "../components/api-reference.js";
import { t } from "../i18n/index.js";
import { CommandPalette } from "../components/command-palette.js";

/**
 * @typedef {import('../types.js').LunarFunction} lunarFunction
 */

/**
 * Function code editor view component.
 * Provides a Monaco editor for editing function code with an API reference sidebar.
 * @type {Object}
 */
export const FunctionCode = {
  /**
   * Currently loaded function.
   * @type {LunarFunction|null}
   */
  func: null,

  /**
   * Whether the view is loading.
   * @type {boolean}
   */
  loading: true,

  /**
   * Currently active API reference section.
   * @type {string}
   */
  activeApiSection: "handler",

  /**
   * Edited code (null if unchanged from a saved version).
   * @type {string|null}
   */
  editedCode: null,

  /**
   * Whether the code editor is maximized.
   * @type {boolean}
   */
  isCodeMaximized: false,

  /**
   * Initializes the view and loads the function.
   * @param {Object} vnode - Mithril vnode
   */
  oninit: (vnode) => {
    FunctionCode.editedCode = null;
    FunctionCode.isCodeMaximized = false;
    FunctionCode.loadFunction(vnode.attrs.id);
    FunctionCode.registerPaletteCommands();
  },

  /**
   * Cleans up when the view is removed.
   */
  onremove: () => {
    CommandPalette.unregisterItems("function-code");
  },

  /**
   * Registers commands with the command palette.
   */
  registerPaletteCommands: () => {
    const items = [];

    if (FunctionCode.isCodeMaximized) {
      items.push({
        type: "custom",
        label: t("code.restoreEditor"),
        description: t("code.restoreEditorDesc"),
        icon: "arrowsPointingIn",
        onSelect: () => {
          FunctionCode.isCodeMaximized = false;
          FunctionCode.registerPaletteCommands();
          m.redraw();
        },
      });
    } else {
      items.push({
        type: "custom",
        label: t("code.maximizeEditor"),
        description: t("code.maximizeEditorDesc"),
        icon: "arrowsPointingOut",
        onSelect: () => {
          FunctionCode.isCodeMaximized = true;
          FunctionCode.registerPaletteCommands();
          m.redraw();
        },
      });
    }

    CommandPalette.registerItems("function-code", items);
  },

  /**
   * Loads a function by ID from the API.
   * @param {string} id - Function ID
   * @returns {Promise<void>}
   */
  loadFunction: async (id) => {
    FunctionCode.loading = true;
    try {
      FunctionCode.func = await API.functions.get(id);
    } catch (e) {
      console.error("Failed to load function:", e);
    } finally {
      FunctionCode.loading = false;
      m.redraw();
    }
  },

  /**
   * Saves the edited code to the API, creating a new version.
   * @returns {Promise<void>}
   */
  saveCode: async () => {
    if (FunctionCode.editedCode === null) return;

    try {
      await API.functions.update(FunctionCode.func.id, {
        code: FunctionCode.editedCode,
      });
      Toast.show(t("code.codeSaved"), "success");
      FunctionCode.editedCode = null;
      await FunctionCode.loadFunction(FunctionCode.func.id);
    } catch (e) {
      Toast.show(t("code.failedToSave") + ": " + e.message, "error");
    }
  },

  /**
   * Renders the function code editor view.
   * @param {Object} _vnode - Mithril vnode
   * @returns {Object} Mithril vnode
   */
  view: (_vnode) => {
    if (FunctionCode.loading) {
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", t("functions.loadingFunction")),
      ]);
    }

    if (!FunctionCode.func) {
      return m(
        ".fade-in",
        m(Card, m(CardContent, t("common.functionNotFound"))),
      );
    }

    const func = FunctionCode.func;

    return m(".fade-in", [
      // Header
      m(".function-details-header", [
        m(".function-details-left", [
          m(BackButton, { href: routes.functions() }),
          m(".function-details-divider"),
          m(".function-details-info", [
            m("h1.function-details-title", [
              func.name,
              m(IDBadge, { id: func.id }),
              m(
                Badge,
                {
                  variant: BadgeVariant.OUTLINE,
                  size: BadgeSize.SM,
                  mono: true,
                },
                `v${func.active_version.version}`,
              ),
            ]),
            m(
              "p.function-details-description",
              func.description || t("common.noDescription"),
            ),
          ]),
        ]),
        m(".function-details-actions", [
          m(StatusBadge, { enabled: !func.disabled, glow: true }),
          m(
            Button,
            {
              variant: ButtonVariant.PRIMARY,
              size: ButtonSize.SM,
              onclick: FunctionCode.saveCode,
              disabled: FunctionCode.editedCode === null,
            },
            t("common.saveChanges"),
          ),
        ]),
      ]),

      // Tabs
      m(Tabs, {
        tabs: getFunctionTabs(func.id),
        activeTab: "code",
      }),

      // Content
      m(TabContent, [
        m(".code-tab-container", [
          m(
            MaximizableCard,
            {
              title: "main.lua",
              icon: "code",
              class: "code-card",
              headerActions: [m("span.code-editor-lang", "lua")],
              isMaximized: FunctionCode.isCodeMaximized,
              onToggleMaximize: (val) => {
                FunctionCode.isCodeMaximized = val;
                FunctionCode.registerPaletteCommands();
                m.redraw();
              },
            },
            m(CodeEditor, {
              id: "code-viewer",
              height: "calc(100vh - 340px)",
              value: FunctionCode.editedCode !== null
                ? FunctionCode.editedCode
                : func.active_version.code,
              onChange: (value) => {
                if (value !== func.active_version.code) {
                  FunctionCode.editedCode = value;
                } else {
                  FunctionCode.editedCode = null;
                }
                m.redraw();
              },
            }),
          ),
          m(".api-reference-sidebar", [
            m(APIReference, {
              sections: getLuaAPISections(),
              activeSection: FunctionCode.activeApiSection,
              onSectionChange: (id) => {
                FunctionCode.activeApiSection = id;
              },
            }),
          ]),
        ]),
      ]),
    ]);
  },
};

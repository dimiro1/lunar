/**
 * @fileoverview Function store view for key-value management.
 */

import { icons } from "../icons.js";
import { API } from "../api.js";
import { t } from "../i18n/index.js";
import { Toast } from "../components/toast.js";
import { BackButton, Button, ButtonVariant } from "../components/button.js";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
} from "../components/card.js";
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
  FormHelp,
} from "../components/form.js";
import { DataEditor } from "../components/data-editor.js";
/**
 * @typedef {import('../types.js').LunarFunction} LunarFunction
 */

/**
 * @typedef {Object} StoreItem
 * @property {string} key - KV data key
 * @property {string} value - KV data value
 * @property {('original'|'added'|'modified'|'removed')} state - Edit state
 * @property {string} [originalKey] - Original key before editing
 */

/**
 * Function store view component.
 * Manages function key-value store including function-scoped and global KV data.
 * @type {Object}
 */
export const FunctionData = {
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
   * Array of function-scoped kv entries with edit state.
   * @type {StoreItem[]}
   */
  scopedStoreItems: [],

  /**
   * Array of global kv entries with edit state.
   * @type {StoreItem[]}
   */
  globalStoreItems: [],
  
  /**
   * Initializes the view and loads the function.
   * @param {Object} vnode - Mithril vnode
   */
  oninit: (vnode) => {
      FunctionData.scopedStoreItems = [];
      FunctionData.globalStoreItems = [];
      FunctionData.scopedDataErrors = {};
      FunctionData.globalDataErrors = {};
      FunctionData.loadFunction(vnode.attrs.id);
  },

  /**
   * Loads a function by ID and initializes env vars.
   * @param {string} id - Function ID
   * @returns {Promise<void>}
   */
  loadFunction: async (id) => {
    FunctionData.loading = true;
    try {
      const func = await API.functions.get(id);
      FunctionData.func = func;
      FunctionData.scopedStoreItems = Object.entries(
        FunctionData.func.scoped_data || {},
      ).map(([key, value]) => ({
        key,
        value,
        state: "original",
        originalKey: key,
      }));
      FunctionData.globalStoreItems = Object.entries(
        FunctionData.func.global_data || {},
      ).map(([key, value]) => ({
        key,
        value,
        state: "original",
        originalKey: key,
      }));
    } catch (e) {
      console.error("Failed to load function:", e);
    } finally {
      FunctionData.loading = false;
      m.redraw();
    }
  },

  /**
   * Checks if there are unsaved kv changes.
   * @returns {boolean} True if there are changes
   */
  hasScopedStoreChanges: () => {
    return (
      FunctionData.scopedStoreItems.some((v) => v.state !== "original")
    );
  },

  /**
   * Checks if there are unsaved kv changes.
   * @returns {boolean} True if there are changes
   */
  hasGlobalStoreChanges: () => {
    return (
      FunctionData.globalStoreItems.some((v) => v.state !== "original")
    );
 },
  
  /**
   * Saves function-scoped or global key-value changes to the API.
   * @returns {Promise<void>}
   */
  saveStoreChanges: async (isGlobal) => {
    FunctionData.scopedDataErrors = {};
    FunctionData.globalDataErrors = {};
      
    try {
        const entries = [];
        let checkedEntries = isGlobal ? FunctionData.globalStoreItems : FunctionData.scopedStoreItems;

        checkedEntries.forEach((entry) => {
            if (entry.state !== "removed") {
                const key = entry.key || "";
                const value = entry.value || "";
                if (key || value) {
                    entries.push({ key: key, value: value });
                }
            }
        });

      await API.functions.updateKvStore(FunctionData.func.id, { global: isGlobal, kvEntries: entries });

      Toast.show(t("toast.kvSaved"), "success");
      await FunctionData.loadFunction(FunctionData.func.id);
    } catch (e) {
        if (isGlobal) {
            FunctionData.globalDataErrors.general = e.message;
        } else {
            FunctionData.scopedDataErrors.general = e.message;
        }
        m.redraw();
    }
  },
  
  /**
   * Checks if there are unsaved status changes.
   * @returns {boolean} True if there are changes
   */
  hasStatusChanges: () => {
    return FunctionData.editedDisabled !== null;
  },

  /**
   * Renders the function settings view.
   * @param {Object} _vnode - Mithril vnode
   * @returns {Object} Mithril vnode
   */
  view: (_vnode) => {
    if (FunctionData.loading) {
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", t("functions.loadingFunction")),
      ]);
    }

    if (!FunctionData.func) {
      return m(
        ".fade-in",
        m(Card, m(CardContent, t("common.functionNotFound"))),
      );
    }

    const func = FunctionData.func;

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
        ]),
      ]),

      // Tabs
      m(Tabs, {
        tabs: getFunctionTabs(func.id),
        activeTab: "data",
      }),

      // Content
      m(TabContent, [
        m(".kv-tab-container", [
          // Function-scoped key-value store
          m(Card, { style: "margin-bottom: 1.5rem" }, [
            m(CardHeader, {
              title: t("kvStore.functionScoped"),
              subtitle: t("kvStore.variablesCount", {
                count: FunctionData.scopedStoreItems.filter((v) =>
                  v.state !== "removed"
                ).length,
              }),
            }),
            m(CardContent, [
              FunctionData.scopedDataErrors.general &&
              m(FormHelp, {
                error: true,
                text: FunctionData.scopedDataErrors.general,
                style: "margin-bottom: 1rem",
              }),

              m(DataEditor, {
                dataVars: FunctionData.scopedStoreItems,
                onAdd: () => {
                  FunctionData.scopedStoreItems.push({
                    key: "",
                    value: "",
                    state: "added",
                  });
                  delete FunctionData.scopedDataErrors.general;
                },
                onToggleRemove: (i) => {
                  const dataVar = FunctionData.scopedStoreItems[i];
                  if (dataVar.state === "removed") {
                    dataVar.state = "original";
                  } else if (dataVar.state === "added") {
                    FunctionData.scopedStoreItems.splice(i, 1);
                  } else {
                    dataVar.state = "removed";
                  }
                  delete FunctionData.scopedDataErrors.general;
                },
                onChange: (i, key, value) => {
                  FunctionData.scopedStoreItems[i].key = key;
                  FunctionData.scopedStoreItems[i].value = value;
                  if (FunctionData.scopedStoreItems[i].state === "original") {
                    FunctionData.scopedStoreItems[i].state = "modified";
                  }
                  delete FunctionData.scopedDataErrors.general;
                },
              }),
            ]),
            m(CardFooter, [
              m(
                Button,
                {
                  variant: ButtonVariant.PRIMARY,
                  onclick: () => FunctionData.saveStoreChanges(false),
                  disabled: !FunctionData.hasScopedStoreChanges(),
                },
                t("common.saveChanges"),
              ),
            ]),
          ]),
          // Global key-value store
          m(Card, { style: "margin-bottom: 1.5rem" }, [
            m(CardHeader, {
              title: t("kvStore.global"),
              subtitle: t("kvStore.variablesCount", {
                count: FunctionData.globalStoreItems.filter((v) =>
                  v.state !== "removed"
                ).length,
              }),
            }),
            m(CardContent, [
              FunctionData.globalDataErrors.general &&
              m(FormHelp, {
                error: true,
                text: FunctionData.globalDataErrors.general,
                style: "margin-bottom: 1rem",
              }),
              m(DataEditor, {
                dataVars: FunctionData.globalStoreItems,
                onAdd: () => {
                  FunctionData.globalStoreItems.push({
                    key: "",
                    value: "",
                    state: "added",
                  });
                  delete FunctionData.globalDataErrors.general;
                },
                onToggleRemove: (i) => {
                  const dataVar = FunctionData.globalStoreItems[i];
                  if (dataVar.state === "removed") {
                    dataVar.state = "original";
                  } else if (dataVar.state === "added") {
                    FunctionData.globalStoreItems.splice(i, 1);
                  } else {
                    dataVar.state = "removed";
                  }
                  delete FunctionData.globalDataErrors.general;
                },
                onChange: (i, key, value) => {
                  FunctionData.globalStoreItems[i].key = key;
                  FunctionData.globalStoreItems[i].value = value;
                  if (FunctionData.globalStoreItems[i].state === "original") {
                    FunctionData.globalStoreItems[i].state = "modified";
                  }
                  delete FunctionData.globalDataErrors.general;
                },
              }),
            ]),
            m(CardFooter, [
              m(
                Button,
                {
                  variant: ButtonVariant.PRIMARY,
                  onclick: () => FunctionData.saveStoreChanges(true),
                  disabled: !FunctionData.hasGlobalStoreChanges(),
                },
                t("common.saveChanges"),
              ),
            ]),
          ])
        ]),
      ]),
    ]);
  },
};

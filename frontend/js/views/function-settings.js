/**
 * @fileoverview Function settings view for configuration management.
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
  MethodBadges,
  StatusBadge,
} from "../components/badge.js";
import { TabContent, Tabs } from "../components/tabs.js";
import { getFunctionTabs } from "../utils.js";
import { paths, routes } from "../routes.js";
import {
  CopyInput,
  FormCheckbox,
  FormGroup,
  FormHelp,
  FormInput,
  FormLabel,
  FormTextarea,
} from "../components/form.js";
import { EnvEditor } from "../components/env-editor.js";

/**
 * @typedef {import('../types.js').LunarFunction} LunarFunction
 */

/**
 * @typedef {Object} EnvVar
 * @property {string} key - Environment variable key
 * @property {string} value - Environment variable value
 * @property {('original'|'added'|'modified'|'removed')} state - Edit state
 * @property {string} [originalKey] - Original key before editing
 */

/**
 * Function settings view component.
 * Manages function configuration including name, description, env vars, and status.
 * @type {Object}
 */
export const FunctionSettings = {
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
   * Edited name (null if unchanged).
   * @type {string|null}
   */
  editedName: null,

  /**
   * Edited description (null if unchanged).
   * @type {string|null}
   */
  editedDescription: null,

  /**
   * Edited disabled state (null if unchanged).
   * @type {boolean|null}
   */
  editedDisabled: null,

  /**
   * Edited retention days (null if unchanged).
   * @type {number|null}
   */
  editedRetentionDays: null,

  /**
   * Array of environment variables with edit state.
   * @type {EnvVar[]}
   */
  envVars: [],

  /**
   * Environment variable errors.
   * @type {Object.<string, string>}
   */
  envErrors: {},

  /**
   * Initializes the view and loads the function.
   * @param {Object} vnode - Mithril vnode
   */
  oninit: (vnode) => {
    FunctionSettings.editedName = null;
    FunctionSettings.editedDescription = null;
    FunctionSettings.editedDisabled = null;
    FunctionSettings.editedRetentionDays = null;
    FunctionSettings.envVars = [];
    FunctionSettings.envErrors = {};
    FunctionSettings.loadFunction(vnode.attrs.id);
  },

  /**
   * Loads a function by ID and initializes env vars.
   * @param {string} id - Function ID
   * @returns {Promise<void>}
   */
  loadFunction: async (id) => {
    FunctionSettings.loading = true;
    try {
      FunctionSettings.func = await API.functions.get(id);
      FunctionSettings.editedName = null;
      FunctionSettings.editedDescription = null;
      FunctionSettings.editedDisabled = null;
      FunctionSettings.editedRetentionDays = null;
      FunctionSettings.envVars = Object.entries(
        FunctionSettings.func.env_vars || {},
      ).map(([key, value]) => ({
        key,
        value,
        state: "original",
        originalKey: key,
      }));
      FunctionSettings.envErrors = {};
    } catch (e) {
      console.error("Failed to load function:", e);
    } finally {
      FunctionSettings.loading = false;
      m.redraw();
    }
  },

  /**
   * Checks if there are unsaved environment variable changes.
   * @returns {boolean} True if there are changes
   */
  hasEnvChanges: () => {
    return (
      FunctionSettings.envVars.some((v) => v.state !== "original") ||
      FunctionSettings.envVars.some((v) => v.state === "modified")
    );
  },

  /**
   * Saves environment variables to the API.
   * @returns {Promise<void>}
   */
  saveEnvVars: async () => {
    FunctionSettings.envErrors = {};

    try {
      const env_vars = {};
      FunctionSettings.envVars.forEach((envVar) => {
        if (envVar.state !== "removed") {
          const key = envVar.key || "";
          const value = envVar.value || "";
          if (key || value) {
            env_vars[key] = value;
          }
        }
      });

      await API.functions.updateEnv(FunctionSettings.func.id, env_vars);
      Toast.show(t("toast.envVarsUpdated"), "success");
      await FunctionSettings.loadFunction(FunctionSettings.func.id);
    } catch (e) {
      FunctionSettings.envErrors.general = e.message;
      m.redraw();
    }
  },

  /**
   * Checks if there are unsaved general settings changes.
   * @returns {boolean} True if there are changes
   */
  hasGeneralChanges: () => {
    return (
      FunctionSettings.editedName !== null ||
      FunctionSettings.editedDescription !== null ||
      FunctionSettings.editedRetentionDays !== null
    );
  },

  /**
   * Saves general settings (name, description, retention) to the API.
   * @returns {Promise<void>}
   */
  saveGeneralSettings: async () => {
    if (!FunctionSettings.hasGeneralChanges()) return;

    try {
      const updates = {};
      if (FunctionSettings.editedName !== null) {
        updates.name = FunctionSettings.editedName;
      }
      if (FunctionSettings.editedDescription !== null) {
        updates.description = FunctionSettings.editedDescription;
      }
      if (FunctionSettings.editedRetentionDays !== null) {
        updates.retention_days = FunctionSettings.editedRetentionDays || 7;
      }

      await API.functions.update(FunctionSettings.func.id, updates);
      Toast.show(t("toast.settingsSaved"), "success");
      await FunctionSettings.loadFunction(FunctionSettings.func.id);
    } catch (e) {
      Toast.show(t("toast.failedToSave") + ": " + e.message, "error");
    }
  },

  /**
   * Deletes the function after confirmation.
   * @returns {Promise<void>}
   */
  deleteFunction: async () => {
    if (
      !confirm(
        t("settings.deleteConfirm", { name: FunctionSettings.func.name }),
      )
    ) {
      return;
    }

    try {
      await API.functions.delete(FunctionSettings.func.id);
      Toast.show(t("toast.functionDeleted"), "success");
      m.route.set(paths.functions());
    } catch (e) {
      Toast.show(t("toast.failedToDelete") + ": " + e.message, "error");
    }
  },

  /**
   * Checks if there are unsaved status changes.
   * @returns {boolean} True if there are changes
   */
  hasStatusChanges: () => {
    return FunctionSettings.editedDisabled !== null;
  },

  /**
   * Saves status settings (enabled/disabled) to the API.
   * @returns {Promise<void>}
   */
  saveStatusSettings: async () => {
    if (!FunctionSettings.hasStatusChanges()) return;

    try {
      await API.functions.update(FunctionSettings.func.id, {
        disabled: FunctionSettings.editedDisabled,
      });
      const toastKey = FunctionSettings.editedDisabled
        ? "toast.functionDisabled"
        : "toast.functionEnabled";
      Toast.show(t(toastKey), "success");
      await FunctionSettings.loadFunction(FunctionSettings.func.id);
    } catch (e) {
      Toast.show(t("toast.failedToUpdate") + ": " + e.message, "error");
    }
  },

  /**
   * Renders the function settings view.
   * @param {Object} _vnode - Mithril vnode
   * @returns {Object} Mithril vnode
   */
  view: (_vnode) => {
    if (FunctionSettings.loading) {
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", t("functions.loadingFunction")),
      ]);
    }

    if (!FunctionSettings.func) {
      return m(
        ".fade-in",
        m(Card, m(CardContent, t("common.functionNotFound"))),
      );
    }

    const func = FunctionSettings.func;

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
        activeTab: "settings",
      }),

      // Content
      m(TabContent, [
        m(".settings-tab-container", [
          // General Settings
          m(Card, { style: "margin-bottom: 1.5rem" }, [
            m(CardHeader, { title: t("settings.generalConfig") }),
            m(CardContent, [
              m(FormGroup, [
                m(FormLabel, { text: t("settings.functionName") }),
                m(FormInput, {
                  value: FunctionSettings.editedName !== null
                    ? FunctionSettings.editedName
                    : func.name,
                  mono: true,
                  "aria-label": t("settings.functionName"),
                  oninput: (e) => {
                    const value = e.target.value;
                    if (value !== func.name) {
                      FunctionSettings.editedName = value;
                    } else {
                      FunctionSettings.editedName = null;
                    }
                  },
                }),
              ]),
              m(FormGroup, [
                m(FormLabel, { text: t("settings.description") }),
                m(FormTextarea, {
                  "aria-label": t("settings.description"),
                  value: FunctionSettings.editedDescription !== null
                    ? FunctionSettings.editedDescription
                    : func.description || "",
                  rows: 2,
                  oninput: (e) => {
                    const value = e.target.value;
                    if (value !== (func.description || "")) {
                      FunctionSettings.editedDescription = value;
                    } else {
                      FunctionSettings.editedDescription = null;
                    }
                  },
                }),
              ]),
              m(FormGroup, [
                m(FormLabel, { text: t("settings.logRetention") }),
                m(
                  "select.form-select",
                  {
                    id: "logRetention",
                    "aria-label": t("settings.logRetention"),
                    value: FunctionSettings.editedRetentionDays !== null
                      ? FunctionSettings.editedRetentionDays
                      : func.retention_days || 7,
                    onchange: (e) => {
                      const value = parseInt(e.target.value, 10);
                      if (value !== (func.retention_days || 7)) {
                        FunctionSettings.editedRetentionDays = value;
                      } else {
                        FunctionSettings.editedRetentionDays = null;
                      }
                    },
                  },
                  [
                    m("option", { value: 7 }, t("settings.retention.days7")),
                    m("option", { value: 15 }, t("settings.retention.days15")),
                    m("option", { value: 30 }, t("settings.retention.days30")),
                    m("option", { value: 365 }, t("settings.retention.year1")),
                  ],
                ),
                m(FormHelp, {
                  text: t("settings.retentionHelp"),
                }),
              ]),
            ]),
            m(CardFooter, [
              m(
                Button,
                {
                  variant: ButtonVariant.PRIMARY,
                  onclick: FunctionSettings.saveGeneralSettings,
                  disabled: !FunctionSettings.hasGeneralChanges(),
                },
                t("common.saveChanges"),
              ),
            ]),
          ]),

          // Environment Variables
          m(Card, { style: "margin-bottom: 1.5rem" }, [
            m(CardHeader, {
              title: t("settings.envVars"),
              subtitle: t("settings.variablesCount", {
                count: FunctionSettings.envVars.filter((v) =>
                  v.state !== "removed"
                ).length,
              }),
            }),
            m(CardContent, [
              FunctionSettings.envErrors.general &&
              m(FormHelp, {
                error: true,
                text: FunctionSettings.envErrors.general,
                style: "margin-bottom: 1rem",
              }),

              m(EnvEditor, {
                envVars: FunctionSettings.envVars,
                onAdd: () => {
                  FunctionSettings.envVars.push({
                    key: "",
                    value: "",
                    state: "added",
                  });
                  delete FunctionSettings.envErrors.general;
                },
                onToggleRemove: (i) => {
                  const envVar = FunctionSettings.envVars[i];
                  if (envVar.state === "removed") {
                    envVar.state = "original";
                  } else if (envVar.state === "added") {
                    FunctionSettings.envVars.splice(i, 1);
                  } else {
                    envVar.state = "removed";
                  }
                  delete FunctionSettings.envErrors.general;
                },
                onChange: (i, key, value) => {
                  FunctionSettings.envVars[i].key = key;
                  FunctionSettings.envVars[i].value = value;
                  if (FunctionSettings.envVars[i].state === "original") {
                    FunctionSettings.envVars[i].state = "modified";
                  }
                  delete FunctionSettings.envErrors.general;
                },
              }),
            ]),
            m(CardFooter, [
              m(
                Button,
                {
                  variant: ButtonVariant.PRIMARY,
                  onclick: FunctionSettings.saveEnvVars,
                  disabled: !FunctionSettings.hasEnvChanges(),
                },
                t("common.saveChanges"),
              ),
            ]),
          ]),

          // Network & Triggers
          m(Card, { style: "margin-bottom: 1.5rem" }, [
            m(CardHeader, { title: t("settings.network") }),
            m(CardContent, [
              m(FormGroup, [
                m(FormLabel, { text: t("settings.invocationUrl") }),
                m(CopyInput, {
                  value: `${window.location.origin}/fn/${func.id}`,
                  mono: true,
                  "aria-label": t("settings.invocationUrl"),
                }),
              ]),
              m(FormGroup, [
                m(FormLabel, { text: t("settings.supportedMethods") }),
                m(MethodBadges, {
                  methods: ["GET", "POST", "PUT", "PATCH", "DELETE"],
                }),
              ]),
            ]),
          ]),

          // Function Status
          m(Card, { variant: "warning", style: "margin-bottom: 1.5rem" }, [
            m(CardHeader, { title: t("settings.functionStatus") }),
            m(CardContent, [
              m(FormCheckbox, {
                id: "enable-function",
                label: t("settings.enableFunction"),
                description: t("settings.disableWarning"),
                checked: FunctionSettings.editedDisabled !== null
                  ? !FunctionSettings.editedDisabled
                  : !func.disabled,
                onchange: () => {
                  const newValue = FunctionSettings.editedDisabled !== null
                    ? !FunctionSettings.editedDisabled
                    : !func.disabled;
                  if (newValue === func.disabled) {
                    FunctionSettings.editedDisabled = null;
                  } else {
                    FunctionSettings.editedDisabled = newValue;
                  }
                },
              }),
            ]),
            m(CardFooter, [
              m(
                Button,
                {
                  variant: ButtonVariant.PRIMARY,
                  onclick: FunctionSettings.saveStatusSettings,
                  disabled: !FunctionSettings.hasStatusChanges(),
                },
                t("common.saveChanges"),
              ),
            ]),
          ]),

          // Danger Zone
          m(Card, { variant: "danger" }, [
            m(CardHeader, { title: t("settings.dangerZone") }),
            m(CardContent, [
              m(".danger-zone-item", [
                m(".danger-zone-info", [
                  m("p.danger-zone-title", t("settings.deleteFunction")),
                  m(
                    "p.danger-zone-description",
                    t("settings.deleteWarning"),
                  ),
                ]),
                m(
                  Button,
                  {
                    variant: ButtonVariant.DESTRUCTIVE,
                    onclick: FunctionSettings.deleteFunction,
                  },
                  t("common.delete"),
                ),
              ]),
            ]),
          ]),
        ]),
      ]),
    ]);
  },
};

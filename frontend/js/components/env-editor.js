/**
 * @fileoverview Environment variable editor component for managing key-value pairs.
 */

import { Button, ButtonSize, ButtonVariant } from "./button.js";
import { FormInput, PasswordInput } from "./form.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {('original'|'added'|'removed')} EnvVarState
 */

/**
 * @typedef {Object} EnvVar
 * @property {string} key - Environment variable key
 * @property {string} value - Environment variable value
 * @property {EnvVarState} [state] - State of the variable (original, added, removed)
 * @property {string} [originalKey] - Original key name (for tracking changes)
 */

/**
 * Environment variable editor component.
 * Allows adding, editing, and removing environment variables with visual state tracking.
 * @type {Object}
 */
export const EnvEditor = {
  /**
   * Renders the environment editor component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {EnvVar[]} [vnode.attrs.envVars=[]] - Array of environment variables
   * @param {function} vnode.attrs.onAdd - Callback when adding a new variable
   * @param {function(number): void} vnode.attrs.onToggleRemove - Callback to toggle remove state
   * @param {function(number, string, string): void} vnode.attrs.onChange - Callback when value changes (index, key, value)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { envVars = [], onAdd, onToggleRemove, onChange } = vnode.attrs;

    return m(".env-editor", [
      m(
        ".env-editor__rows",
        envVars.length === 0
          ? m(
            ".env-editor__empty",
            t("envVars.noVariables"),
          )
          : envVars.map((envVar, i) =>
            m(EnvRow, {
              key: envVar.originalKey || i,
              envVar,
              onToggleRemove: () => onToggleRemove(i),
              onChange: (key, value) => onChange(i, key, value),
            })
          ),
      ),
      m(".env-editor__actions", [
        m(
          Button,
          {
            variant: ButtonVariant.SECONDARY,
            size: ButtonSize.SM,
            icon: "plus",
            onclick: onAdd,
          },
          t("envVars.addVariable"),
        ),
      ]),
    ]);
  },
};

/**
 * Single environment variable row component.
 * @type {Object}
 * @private
 */
const EnvRow = {
  /**
   * Renders a single environment variable row.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {EnvVar} vnode.attrs.envVar - The environment variable
   * @param {function} vnode.attrs.onToggleRemove - Callback to toggle removal
   * @param {function(string, string): void} vnode.attrs.onChange - Callback when value changes (key, value)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { envVar, onToggleRemove, onChange } = vnode.attrs;
    const state = envVar.state || "original";
    const isRemoved = state === "removed";

    return m(
      ".env-editor__row",
      {
        "data-state": state,
        class: state === "removed"
          ? "env-editor__row--removed"
          : state === "added"
          ? "env-editor__row--added"
          : "",
      },
      [
        m(".env-editor__inputs", [
          m(".env-editor__key", [
            m(FormInput, {
              value: envVar.key,
              placeholder: t("envVars.keyPlaceholder"),
              mono: true,
              disabled: isRemoved,
              oninput: (e) => onChange(e.target.value, envVar.value),
            }),
          ]),
          m(".env-editor__value", [
            m(PasswordInput, {
              value: envVar.value,
              placeholder: t("envVars.valuePlaceholder"),
              mono: true,
              disabled: isRemoved,
              oninput: (e) => onChange(envVar.key, e.target.value),
            }),
          ]),
        ]),
        m(Button, {
          variant: ButtonVariant.GHOST,
          size: ButtonSize.ICON,
          icon: isRemoved ? "undo" : "trash",
          title: isRemoved ? t("envVars.restore") : t("envVars.remove"),
          onclick: onToggleRemove,
        }),
      ],
    );
  },
};

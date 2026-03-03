/**
 * @fileoverview Data editor component for managing key-value pairs.
 */

import { Button, ButtonSize, ButtonVariant } from "./button.js";
import { FormInput } from "./form.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {('original'|'added'|'removed')} DataVarState
 */

/**
 * @typedef {Object} DataVar
 * @property {string} key - Data variable key
 * @property {string} value - Data variable value
 * @property {DataVarState} [state] - State of the variable (original, added, removed)
 * @property {string} [originalKey] - Original key name (for tracking changes)
 */

/**
 * Data editor component.
 * Allows adding, editing, and removing data variables with visual state tracking.
 * @type {Object}
 */
export const DataEditor = {
  /**
   * Renders the data editor component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {DataVar[]} [vnode.attrs.dataVars=[]] - Array of data variables
   * @param {function} vnode.attrs.onAdd - Callback when adding a new variable
   * @param {function(number): void} vnode.attrs.onToggleRemove - Callback to toggle remove state
   * @param {function(number, string, string): void} vnode.attrs.onChange - Callback when value changes (index, key, value)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { dataVars = [], onAdd, onToggleRemove, onChange } = vnode.attrs;

    return m(".env-editor", [
      m(
        ".env-editor__rows",
        dataVars.length === 0
          ? m(
            ".env-editor__empty",
            t("kvStore.noEntries"),
          )
          : dataVars.map((dataVar, i) =>
            m(DataRow, {
              key: dataVar.originalKey || i,
              dataVar,
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
          t("kvStore.addEntry"),
        ),
      ]),
    ]);
  },
};

/**
 * Single data variable row component.
 * @type {Object}
 * @private
 */
const DataRow = {
  /**
   * Renders a single data variable row.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {DataVar} vnode.attrs.dataVar - The data variable
   * @param {function} vnode.attrs.onToggleRemove - Callback to toggle removal
   * @param {function(string, string): void} vnode.attrs.onChange - Callback when value changes (key, value)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { dataVar, onToggleRemove, onChange } = vnode.attrs;
    const state = dataVar.state || "original";
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
              value: dataVar.key,
              placeholder: t("envVars.keyPlaceholder"),
              mono: true,
              disabled: isRemoved,
              oninput: (e) => onChange(e.target.value, dataVar.value),
            }),
          ]),
          m(".env-editor__value", [
            m(FormInput, {
              value: dataVar.value,
              placeholder: t("envVars.valuePlaceholder"),
              mono: true,
              disabled: isRemoved,
              oninput: (e) => onChange(dataVar.key, e.target.value),
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
